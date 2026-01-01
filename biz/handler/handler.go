package handler

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
	"github.com/yi-nology/git-manage-service/biz/pkg/response"
	"github.com/yi-nology/git-manage-service/biz/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type RegisterRepoReq struct {
	Name         string                    `json:"name"`
	Path         string                    `json:"path"`
	RemoteURL    string                    `json:"remote_url"`
	AuthType     string                    `json:"auth_type"`
	AuthKey      string                    `json:"auth_key"`
	AuthSecret   string                    `json:"auth_secret"`
	ConfigSource string                    `json:"config_source"`
	Remotes      []model.GitRemote         `json:"remotes"`      // Optional list of remotes to sync
	RemoteAuths  map[string]model.AuthInfo `json:"remote_auths"` // Optional auth per remote
}

// @Summary Register a new repository
// @Tags Repositories
// @Accept json
// @Produce json
// @Param request body RegisterRepoReq true "Repo info"
// @Success 200 {object} response.Response{data=model.Repo}
// @Router /api/repos [post]
func RegisterRepo(ctx context.Context, c *app.RequestContext) {
	var req RegisterRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Validate path
	gitSvc := service.NewGitService()
	if !gitSvc.IsGitRepo(req.Path) {
		response.BadRequest(c, "path is not a valid git repository")
		return
	}

	// Sync Remotes if provided
	if len(req.Remotes) > 0 {
		// Get existing remotes
		existingConfig, err := gitSvc.GetRepoConfig(req.Path)
		if err == nil {
			// Remove remotes not in request
			for _, existing := range existingConfig.Remotes {
				found := false
				for _, r := range req.Remotes {
					if r.Name == existing.Name {
						found = true
						break
					}
				}
				if !found {
					gitSvc.RemoveRemote(req.Path, existing.Name)
				}
			}

			// Add or Update remotes
			for _, r := range req.Remotes {
				// Remove first to ensure update (simple way)
				gitSvc.RemoveRemote(req.Path, r.Name)
				if err := gitSvc.AddRemote(req.Path, r.Name, r.FetchURL, r.IsMirror); err != nil {
					// log error but continue?
				}
				// Handle PushURL if different
				if r.PushURL != "" && r.PushURL != r.FetchURL {
					gitSvc.SetRemotePushURL(req.Path, r.Name, r.PushURL)
				}
			}
		}
	}

	repo := model.Repo{
		Key:          uuid.New().String(),
		Name:         req.Name,
		Path:         req.Path,
		RemoteURL:    req.RemoteURL,
		AuthType:     req.AuthType,
		AuthKey:      req.AuthKey,
		AuthSecret:   req.AuthSecret,
		ConfigSource: req.ConfigSource,
		RemoteAuths:  req.RemoteAuths,
	}
	if repo.ConfigSource == "" {
		repo.ConfigSource = "local"
	}
	if err := dal.DB.Create(&repo).Error; err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	service.AuditSvc.Log(c, "CREATE", "repo:"+repo.Key, map[string]string{"name": repo.Name, "path": repo.Path})
	response.Success(c, repo)
}

type ScanRepoReq struct {
	Path string `json:"path"`
}

// @Summary Scan a local repository
// @Tags Repositories
// @Accept json
// @Produce json
// @Param request body ScanRepoReq true "Scan info"
// @Success 200 {object} response.Response{data=model.GitRepoConfig}
// @Router /api/repos/scan [post]
func ScanRepo(ctx context.Context, c *app.RequestContext) {
	var req ScanRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	gitSvc := service.NewGitService()
	if !gitSvc.IsGitRepo(req.Path) {
		response.BadRequest(c, "path is not a valid git repository")
		return
	}

	config, err := gitSvc.GetRepoConfig(req.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, config)
}

type CloneRepoReq struct {
	RemoteURL    string `json:"remote_url"`
	LocalPath    string `json:"local_path"`
	AuthType     string `json:"auth_type"`
	AuthKey      string `json:"auth_key"`
	AuthSecret   string `json:"auth_secret"`
	ConfigSource string `json:"config_source"`
}

// @Summary Clone a remote repository
// @Tags Repositories
// @Accept json
// @Produce json
// @Param request body CloneRepoReq true "Clone info"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /api/repos/clone [post]
func CloneRepo(ctx context.Context, c *app.RequestContext) {
	var req CloneRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	gitSvc := service.NewGitService()
	// Check if directory already exists
	if _, err := os.Stat(req.LocalPath); err == nil {
		// Exists, check if empty or is git repo
		if gitSvc.IsGitRepo(req.LocalPath) {
			response.BadRequest(c, "directory already contains a git repository")
			return
		}
	}

	// Create Task
	taskID := uuid.New().String()
	service.GlobalTaskManager.AddTask(taskID)

	// Async Clone
	go func() {
		progressChan := make(chan string)

		go func() {
			for msg := range progressChan {
				service.GlobalTaskManager.AppendLog(taskID, msg)
			}
		}()

		err := gitSvc.CloneWithProgress(req.RemoteURL, req.LocalPath, req.AuthType, req.AuthKey, req.AuthSecret, progressChan)
		close(progressChan)

		if err != nil {
			service.GlobalTaskManager.UpdateStatus(taskID, "failed", err.Error())
			return
		}

		service.GlobalTaskManager.UpdateStatus(taskID, "success", "")

		// Register after success
		name := filepath.Base(req.LocalPath)
		repo := model.Repo{
			Key:          uuid.New().String(),
			Name:         name,
			Path:         req.LocalPath,
			RemoteURL:    req.RemoteURL,
			AuthType:     req.AuthType,
			AuthKey:      req.AuthKey,
			AuthSecret:   req.AuthSecret,
			ConfigSource: req.ConfigSource,
		}
		if repo.ConfigSource == "" {
			repo.ConfigSource = "local"
		}
		dal.DB.Create(&repo)
	}()

	response.Success(c, map[string]string{"task_id": taskID})
}

// @Summary Get clone task status
// @Tags Repositories
// @Param id path string true "Task ID"
// @Produce json
// @Success 200 {object} response.Response{data=service.Task}
// @Router /api/tasks/{id} [get]
func GetCloneTask(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	task, ok := service.GlobalTaskManager.GetTask(id)
	if !ok {
		response.NotFound(c, "task not found")
		return
	}
	response.Success(c, task)
}

type TestConnectionReq struct {
	URL string `json:"url"`
}

// @Summary Test remote connection
// @Tags System
// @Param request body TestConnectionReq true "Connection info"
// @Produce json
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /api/git/test-connection [post]
func TestConnection(ctx context.Context, c *app.RequestContext) {
	var req TestConnectionReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	gitSvc := service.NewGitService()
	if err := gitSvc.TestRemoteConnection(req.URL); err != nil {
		// This is technically a success response but with failed status
		// Or we can return error? Let's keep existing logic but wrapped
		c.JSON(consts.StatusOK, response.Response{
			Code:    0,
			Message: "success",
			Data:    map[string]string{"status": "failed", "error": err.Error()},
		})
		return
	}

	response.Success(c, map[string]string{"status": "success"})
}

// @Summary List registered repositories
// @Tags Repositories
// @Produce json
// @Success 200 {object} response.Response{data=[]model.Repo}
// @Router /api/repos [get]
func ListRepos(ctx context.Context, c *app.RequestContext) {
	var repos []model.Repo
	dal.DB.Find(&repos)
	response.Success(c, repos)
}

// @Summary Update a repository
// @Tags Repositories
// @Accept json
// @Produce json
// @Param id path int true "Repo ID"
// @Param request body RegisterRepoReq true "Repo info"
// @Success 200 {object} response.Response{data=model.Repo}
// @Router /api/repos/{id} [put]
func UpdateRepo(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req RegisterRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var repo model.Repo
	if err := dal.DB.First(&repo, id).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	// Validate path if changed
	if req.Path != repo.Path {
		gitSvc := service.NewGitService()
		if !gitSvc.IsGitRepo(req.Path) {
			response.BadRequest(c, "path is not a valid git repository")
			return
		}
	}

	// Sync Remotes if provided
	if len(req.Remotes) > 0 {
		gitSvc := service.NewGitService()
		// Get existing remotes
		existingConfig, err := gitSvc.GetRepoConfig(req.Path)
		if err == nil {
			// Remove remotes not in request
			for _, existing := range existingConfig.Remotes {
				found := false
				for _, r := range req.Remotes {
					if r.Name == existing.Name {
						found = true
						break
					}
				}
				if !found {
					gitSvc.RemoveRemote(req.Path, existing.Name)
				}
			}

			// Add or Update remotes
			for _, r := range req.Remotes {
				// Remove first to ensure update (simple way)
				gitSvc.RemoveRemote(req.Path, r.Name)
				if err := gitSvc.AddRemote(req.Path, r.Name, r.FetchURL, r.IsMirror); err != nil {
					// log error but continue?
				}
				// Handle PushURL if different
				if r.PushURL != "" && r.PushURL != r.FetchURL {
					gitSvc.SetRemotePushURL(req.Path, r.Name, r.PushURL)
				}
			}
		}
	}

	repo.Name = req.Name
	repo.Path = req.Path
	repo.RemoteURL = req.RemoteURL
	repo.AuthType = req.AuthType
	repo.AuthKey = req.AuthKey
	repo.AuthSecret = req.AuthSecret
	repo.ConfigSource = req.ConfigSource
	repo.RemoteAuths = req.RemoteAuths
	if repo.ConfigSource == "" {
		repo.ConfigSource = "local"
	}

	if err := dal.DB.Save(&repo).Error; err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	service.AuditSvc.Log(c, "UPDATE", "repo:"+repo.Key, map[string]string{"name": repo.Name})
	response.Success(c, repo)
}

// @Summary Delete a repository
// @Tags Repositories
// @Param id path int true "Repo ID"
// @Success 200 {object} response.Response
// @Router /api/repos/{id} [delete]
func DeleteRepo(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// Check if used in SyncTask
	var repo model.Repo
	if err := dal.DB.First(&repo, id).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	// Check if used in SyncTask
	var count int64
	dal.DB.Model(&model.SyncTask{}).Where("source_repo_key = ? OR target_repo_key = ?", repo.Key, repo.Key).Count(&count)
	if count > 0 {
		response.BadRequest(c, "cannot delete repo used in sync tasks")
		return
	}

	if err := dal.DB.Delete(&repo).Error; err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	service.AuditSvc.Log(c, "DELETE", "repo:"+repo.Key, nil)
	response.Success(c, map[string]string{"message": "deleted"})
}

// @Summary Create a sync task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param request body model.SyncTask true "Task info"
// @Success 200 {object} response.Response{data=model.SyncTask}
// @Router /api/sync/tasks [post]
func CreateTask(ctx context.Context, c *app.RequestContext) {
	var req model.SyncTask
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req.Key = uuid.New().String()
	// Should validate Repo existence

	if err := dal.DB.Create(&req).Error; err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	service.CronSvc.UpdateTask(req)
	service.AuditSvc.Log(c, "CREATE", "task:"+req.Key, req)
	response.Success(c, req)
}

// @Summary List sync tasks for a repo
// @Tags Tasks
// @Param repo_id query int false "Repo ID"
// @Param repo_key query string false "Repo Key"
// @Produce json
// @Success 200 {object} response.Response{data=[]model.SyncTask}
// @Router /api/sync/tasks [get]
func ListTasks(ctx context.Context, c *app.RequestContext) {
	repoIDStr := c.Query("repo_id")
	repoKey := c.Query("repo_key")
	var tasks []model.SyncTask

	db := dal.DB.Preload("SourceRepo").Preload("TargetRepo")

	if repoKey != "" {
		db = db.Where("source_repo_key = ? OR target_repo_key = ?", repoKey, repoKey)
	} else if repoIDStr != "" {
		repoID, _ := strconv.Atoi(repoIDStr)
		var repo model.Repo
		if err := dal.DB.First(&repo, repoID).Error; err == nil {
			db = db.Where("source_repo_key = ? OR target_repo_key = ?", repo.Key, repo.Key)
		}
	}

	db.Find(&tasks)
	response.Success(c, tasks)
}

// @Summary Get a sync task
// @Tags Tasks
// @Param id path int true "Task ID"
// @Produce json
// @Success 200 {object} response.Response{data=model.SyncTask}
// @Router /api/sync/tasks/{id} [get]
func GetTask(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var task model.SyncTask
	if err := dal.DB.Preload("SourceRepo").Preload("TargetRepo").First(&task, id).Error; err != nil {
		response.NotFound(c, "task not found")
		return
	}
	response.Success(c, task)
}

// @Summary Update a sync task
// @Tags Tasks
// @Param id path int true "Task ID"
// @Param request body model.SyncTask true "Task info"
// @Produce json
// @Success 200 {object} response.Response{data=model.SyncTask}
// @Router /api/sync/tasks/{id} [put]
func UpdateTask(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req model.SyncTask
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var task model.SyncTask
	if err := dal.DB.First(&task, id).Error; err != nil {
		response.NotFound(c, "task not found")
		return
	}

	// Update fields
	task.SourceRepoKey = req.SourceRepoKey
	task.SourceRemote = req.SourceRemote
	task.SourceBranch = req.SourceBranch
	task.TargetRepoKey = req.TargetRepoKey
	task.TargetRemote = req.TargetRemote
	task.TargetBranch = req.TargetBranch
	task.PushOptions = req.PushOptions
	task.Cron = req.Cron
	task.Enabled = req.Enabled

	// Reset webhook token if needed or requested?
	// For now keep existing or allow update if passed?
	// task.WebhookToken = req.WebhookToken

	dal.DB.Save(&task)
	service.CronSvc.UpdateTask(task)
	service.AuditSvc.Log(c, "UPDATE", "task:"+task.Key, task)

	response.Success(c, task)
}

// @Summary Delete a sync task
// @Tags Tasks
// @Param id path int true "Task ID"
// @Success 200 {object} response.Response
// @Router /api/sync/tasks/{id} [delete]
func DeleteTask(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var task model.SyncTask
	if err := dal.DB.First(&task, id).Error; err != nil {
		response.NotFound(c, "task not found")
		return
	}

	dal.DB.Delete(&task)
	service.CronSvc.RemoveTask(task.ID)
	service.AuditSvc.Log(c, "DELETE", "task:"+task.Key, nil)

	response.Success(c, map[string]string{"message": "deleted"})
}

type RunSyncReq struct {
	TaskID uint `json:"task_id"`
}

// @Summary Trigger a sync task manually
// @Tags Sync
// @Accept json
// @Produce json
// @Param request body RunSyncReq true "Task ID"
// @Success 200 {object} response.Response
// @Router /api/sync/run [post]
func RunSync(ctx context.Context, c *app.RequestContext) {
	var req RunSyncReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	go func() {
		svc := service.NewSyncService()
		svc.RunTask(req.TaskID)
	}()

	service.AuditSvc.Log(c, "SYNC", "task_id:"+strconv.Itoa(int(req.TaskID)), nil)
	response.Success(c, map[string]string{"status": "started"})
}

type ExecuteSyncReq struct {
	RepoID       uint   `json:"repo_id"`
	SourceRemote string `json:"source_remote"` // "local", "origin", etc
	SourceBranch string `json:"source_branch"`
	TargetRemote string `json:"target_remote"`
	TargetBranch string `json:"target_branch"`
	PushOptions  string `json:"push_options"`
}

// @Summary Execute an ad-hoc sync
// @Tags Sync
// @Accept json
// @Produce json
// @Param request body ExecuteSyncReq true "Sync info"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /api/sync/execute [post]
func ExecuteSync(ctx context.Context, c *app.RequestContext) {
	var req ExecuteSyncReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var repo model.Repo
	if err := dal.DB.First(&repo, req.RepoID).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	// Construct a temporary task
	task := model.SyncTask{
		Key:           uuid.New().String(),
		SourceRepoKey: repo.Key,
		SourceRepo:    repo,
		SourceRemote:  req.SourceRemote,
		SourceBranch:  req.SourceBranch,
		TargetRepoKey: repo.Key, // Same repo for ad-hoc sync (usually)
		TargetRepo:    repo,
		TargetRemote:  req.TargetRemote,
		TargetBranch:  req.TargetBranch,
		PushOptions:   req.PushOptions,
	}

	go func() {
		svc := service.NewSyncService()
		svc.ExecuteSync(&task)
	}()

	service.AuditSvc.Log(c, "SYNC_ADHOC", "task:"+task.Key, task)
	response.Success(c, map[string]string{"status": "started", "task_key": task.Key})
}

// @Summary Get sync execution history
// @Tags History
// @Param repo_key query string false "Repo Key"
// @Produce json
// @Success 200 {object} response.Response{data=[]model.SyncRun}
// @Router /api/sync/history [get]
func ListHistory(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	db := dal.DB.Order("start_time desc").Limit(50).Preload("Task")

	if repoKey != "" {
		// Find tasks related to this repo
		var taskKeys []string
		dal.DB.Model(&model.SyncTask{}).
			Where("source_repo_key = ? OR target_repo_key = ?", repoKey, repoKey).
			Pluck("key", &taskKeys)

		if len(taskKeys) > 0 {
			db = db.Where("task_key IN ?", taskKeys)
		} else {
			// No tasks found, return empty history
			response.Success(c, []model.SyncRun{})
			return
		}
	}

	var runs []model.SyncRun
	db.Find(&runs)
	response.Success(c, runs)
}

// @Summary Delete a sync history record
// @Tags History
// @Param id path int true "History ID"
// @Success 200 {object} response.Response
// @Router /api/sync/history/{id} [delete]
func DeleteHistory(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	if err := dal.DB.Delete(&model.SyncRun{}, id).Error; err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, map[string]string{"message": "deleted"})
}
