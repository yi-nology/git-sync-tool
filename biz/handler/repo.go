package handler

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/service/stats"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// @Summary Register a new repository
// @Description Register a new git repository by path. If the repository does not exist in the database, it will be added.
// @Tags Repositories
// @Accept json
// @Produce json
// @Param request body api.RegisterRepoReq true "Repo info"
// @Success 200 {object} response.Response{data=api.RepoDTO}
// @Failure 400 {object} response.Response "Bad Request - Invalid input or path"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/repos [post]
func RegisterRepo(ctx context.Context, c *app.RequestContext) {
	var req api.RegisterRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Validate path
	gitSvc := git.NewGitService()
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

	repo := po.Repo{
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
	if err := db.NewRepoDAO().Create(&repo); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	audit.AuditSvc.Log(c, "CREATE", "repo:"+repo.Key, map[string]string{"name": repo.Name, "path": repo.Path})

	// Trigger async stats sync
	go func() {
		// Sync default branch or all? For now, try master and main
		// Or better: get HEAD branch
		head, err := gitSvc.GetHeadBranch(repo.Path)
		if err == nil && head != "" {
			stats.StatsSvc.SyncRepoStats(repo.ID, repo.Path, head)
		}
	}()

	response.Success(c, api.NewRepoDTO(repo))
}

// @Summary Scan a local repository
// @Description Scan a local directory to check if it's a valid git repository and retrieve its configuration.
// @Tags Repositories
// @Accept json
// @Produce json
// @Param request body api.ScanRepoReq true "Scan info"
// @Success 200 {object} response.Response{data=domain.GitRepoConfig}
// @Failure 400 {object} response.Response "Bad Request - Invalid path"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/repos/scan [post]
func ScanRepo(ctx context.Context, c *app.RequestContext) {
	var req api.ScanRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	gitSvc := git.NewGitService()
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

// @Summary Clone a remote repository
// @Description Clone a remote git repository to a local path asynchronously. Returns a task ID to track progress.
// @Tags Repositories
// @Accept json
// @Produce json
// @Param request body api.CloneRepoReq true "Clone info"
// @Success 200 {object} response.Response{data=map[string]string} "Returns task_id"
// @Failure 400 {object} response.Response "Bad Request - Directory exists or invalid input"
// @Router /api/repos/clone [post]
func CloneRepo(ctx context.Context, c *app.RequestContext) {
	var req api.CloneRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	gitSvc := git.NewGitService()
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
	git.GlobalTaskManager.AddTask(taskID)

	// Async Clone
	go func() {
		progressChan := make(chan string)

		go func() {
			for msg := range progressChan {
				git.GlobalTaskManager.AppendLog(taskID, msg)
			}
		}()

		err := gitSvc.CloneWithProgress(req.RemoteURL, req.LocalPath, req.AuthType, req.AuthKey, req.AuthSecret, progressChan)
		close(progressChan)

		if err != nil {
			git.GlobalTaskManager.UpdateStatus(taskID, "failed", err.Error())
			return
		}

		git.GlobalTaskManager.UpdateStatus(taskID, "success", "")

		// Register after success
		name := filepath.Base(req.LocalPath)
		repo := po.Repo{
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
		db.NewRepoDAO().Create(&repo)

		// Trigger async stats sync
		go func() {
			head, err := gitSvc.GetHeadBranch(repo.Path)
			if err == nil && head != "" {
				stats.StatsSvc.SyncRepoStats(repo.ID, repo.Path, head)
			}
		}()
	}()

	response.Success(c, map[string]string{"task_id": taskID})
}

// @Summary Get clone task status
// @Description Get the status and logs of a background clone task.
// @Tags Repositories
// @Param id path string true "Task ID"
// @Produce json
// @Success 200 {object} response.Response{data=git.Task}
// @Failure 404 {object} response.Response "Task not found"
// @Router /api/tasks/{id} [get]
func GetCloneTask(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	task, ok := git.GlobalTaskManager.GetTask(id)
	if !ok {
		response.NotFound(c, "task not found")
		return
	}
	response.Success(c, task)
}

// @Summary Test remote connection
// @Description Test if a remote git URL is accessible.
// @Tags System
// @Param request body api.TestConnectionReq true "Connection info"
// @Produce json
// @Success 200 {object} response.Response{data=map[string]string} "Status success or failed"
// @Failure 400 {object} response.Response "Bad Request"
// @Router /api/git/test-connection [post]
func TestConnection(ctx context.Context, c *app.RequestContext) {
	var req api.TestConnectionReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	gitSvc := git.NewGitService()
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
// @Description Get a list of all registered repositories in the system.
// @Tags Repositories
// @Produce json
// @Success 200 {object} response.Response{data=[]api.RepoDTO}
// @Router /api/repos [get]
func ListRepos(ctx context.Context, c *app.RequestContext) {
	repos, err := db.NewRepoDAO().FindAll()
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	var dtos []api.RepoDTO
	for _, r := range repos {
		dtos = append(dtos, api.NewRepoDTO(r))
	}
	response.Success(c, dtos)
}

// @Summary Get a repository by Key
// @Description Get details of a specific repository by its unique key.
// @Tags Repositories
// @Param key path string true "Repo Key"
// @Produce json
// @Success 200 {object} response.Response{data=api.RepoDTO}
// @Failure 404 {object} response.Response "Repo not found"
// @Router /api/repos/{key} [get]
func GetRepo(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")
	repo, err := db.NewRepoDAO().FindByKey(key)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}
	response.Success(c, api.NewRepoDTO(*repo))
}

// @Summary Update a repository
// @Tags Repositories
// @Accept json
// @Produce json
// @Param key path string true "Repo Key"
// @Param request body api.RegisterRepoReq true "Repo info"
// @Success 200 {object} response.Response{data=api.RepoDTO}
// @Router /api/repos/{key} [put]
func UpdateRepo(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")

	var req api.RegisterRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repoDAO := db.NewRepoDAO()
	repo, err := repoDAO.FindByKey(key)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	// Validate path if changed
	if req.Path != repo.Path {
		gitSvc := git.NewGitService()
		if !gitSvc.IsGitRepo(req.Path) {
			response.BadRequest(c, "path is not a valid git repository")
			return
		}
	}

	// Sync Remotes if provided
	if len(req.Remotes) > 0 {
		gitSvc := git.NewGitService()
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

	if err := repoDAO.Save(repo); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	audit.AuditSvc.Log(c, "UPDATE", "repo:"+repo.Key, map[string]string{"name": repo.Name})
	response.Success(c, api.NewRepoDTO(*repo))
}

// @Summary Delete a repository
// @Description Delete a repository from the system. This does not delete the files from disk, only the registration.
// @Tags Repositories
// @Param key path string true "Repo Key"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response "Repo not found"
// @Failure 400 {object} response.Response "Cannot delete if used in sync tasks"
// @Router /api/repos/{key} [delete]
func DeleteRepo(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")

	repoDAO := db.NewRepoDAO()
	repo, err := repoDAO.FindByKey(key)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	// Check if used in SyncTask
	count, _ := db.NewSyncTaskDAO().CountByRepoKey(repo.Key)
	if count > 0 {
		response.BadRequest(c, "cannot delete repo used in sync tasks")
		return
	}

	if err := repoDAO.Delete(repo); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	audit.AuditSvc.Log(c, "DELETE", "repo:"+repo.Key, nil)
	response.Success(c, map[string]string{"message": "deleted"})
}
