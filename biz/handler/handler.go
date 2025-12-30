package handler

import (
	"context"
	"github.com/yi-nology/git-sync-tool/biz/dal"
	"github.com/yi-nology/git-sync-tool/biz/model"
	"github.com/yi-nology/git-sync-tool/biz/service"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type RegisterRepoReq struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// @Summary Register a new repository
// @Tags Repositories
// @Accept json
// @Produce json
// @Param request body RegisterRepoReq true "Repo info"
// @Success 200 {object} model.Repo
// @Router /api/repos [post]
func RegisterRepo(ctx context.Context, c *app.RequestContext) {
	var req RegisterRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Validate path
	gitSvc := service.NewGitService()
	if !gitSvc.IsGitRepo(req.Path) {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "path is not a valid git repository"})
		return
	}

	repo := model.Repo{
		Name: req.Name,
		Path: req.Path,
	}
	if err := dal.DB.Create(&repo).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, repo)
}

// @Summary List registered repositories
// @Tags Repositories
// @Produce json
// @Success 200 {array} model.Repo
// @Router /api/repos [get]
func ListRepos(ctx context.Context, c *app.RequestContext) {
	var repos []model.Repo
	dal.DB.Find(&repos)
	c.JSON(consts.StatusOK, repos)
}

// @Summary Create a sync task
// @Tags Tasks
// @Accept json
// @Produce json
// @Param request body model.SyncTask true "Task info"
// @Success 200 {object} model.SyncTask
// @Router /api/sync/tasks [post]
func CreateTask(ctx context.Context, c *app.RequestContext) {
	var req model.SyncTask
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := dal.DB.Create(&req).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	service.CronSvc.UpdateTask(req)
	c.JSON(consts.StatusOK, req)
}

// @Summary List sync tasks
// @Tags Tasks
// @Produce json
// @Success 200 {array} model.SyncTask
// @Router /api/sync/tasks [get]
func ListTasks(ctx context.Context, c *app.RequestContext) {
	var tasks []model.SyncTask
	dal.DB.Preload("SourceRepo").Preload("TargetRepo").Find(&tasks)
	c.JSON(consts.StatusOK, tasks)
}

// @Summary Get a sync task
// @Tags Tasks
// @Param id path int true "Task ID"
// @Produce json
// @Success 200 {object} model.SyncTask
// @Router /api/sync/tasks/{id} [get]
func GetTask(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var task model.SyncTask
	if err := dal.DB.Preload("SourceRepo").Preload("TargetRepo").First(&task, id).Error; err != nil {
		c.JSON(consts.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}
	c.JSON(consts.StatusOK, task)
}

// @Summary Update a sync task
// @Tags Tasks
// @Param id path int true "Task ID"
// @Param request body model.SyncTask true "Task info"
// @Produce json
// @Success 200 {object} model.SyncTask
// @Router /api/sync/tasks/{id} [put]
func UpdateTask(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req model.SyncTask
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	var task model.SyncTask
	if err := dal.DB.First(&task, id).Error; err != nil {
		c.JSON(consts.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	// Update fields
	task.SourceRepoID = req.SourceRepoID
	task.SourceRemote = req.SourceRemote
	task.SourceBranch = req.SourceBranch
	task.TargetRepoID = req.TargetRepoID
	task.TargetRemote = req.TargetRemote
	task.TargetBranch = req.TargetBranch
	task.PushOptions = req.PushOptions
	task.Cron = req.Cron
	task.Enabled = req.Enabled

	dal.DB.Save(&task)
	service.CronSvc.UpdateTask(task)

	c.JSON(consts.StatusOK, task)
}

type RunSyncReq struct {
	TaskID uint `json:"task_id"`
}

// @Summary Trigger a sync task manually
// @Tags Sync
// @Accept json
// @Produce json
// @Param request body RunSyncReq true "Task ID"
// @Success 200 {object} map[string]string
// @Router /api/sync/run [post]
func RunSync(ctx context.Context, c *app.RequestContext) {
	var req RunSyncReq
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	go func() {
		svc := service.NewSyncService()
		svc.RunTask(req.TaskID)
	}()

	c.JSON(consts.StatusOK, map[string]string{"status": "started"})
}

// @Summary Get sync execution history
// @Tags History
// @Produce json
// @Success 200 {array} model.SyncRun
// @Router /api/sync/history [get]
func ListHistory(ctx context.Context, c *app.RequestContext) {
	var runs []model.SyncRun
	dal.DB.Order("start_time desc").Limit(50).Preload("Task").Find(&runs)
	c.JSON(consts.StatusOK, runs)
}
