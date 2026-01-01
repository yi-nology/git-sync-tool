package handler

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
	"github.com/yi-nology/git-manage-service/biz/pkg/response"
	"github.com/yi-nology/git-manage-service/biz/service"
)

// @Summary Get repository status
// @Description Get the output of 'git status' for the repository.
// @Tags Workspace
// @Param key path string true "Repo Key"
// @Produce json
// @Success 200 {object} response.Response{data=map[string]string}
// @Failure 404 {object} response.Response "Repo not found"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/repos/{key}/status [get]
func GetRepoStatus(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")

	var repo model.Repo
	if err := dal.DB.Where("key = ?", key).First(&repo).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := service.NewGitService()
	status, err := gitSvc.GetStatus(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"status": status})
}

type SubmitChangesReq struct {
	Message string `json:"message"`
	Push    bool   `json:"push"`
}

// @Summary Submit changes (Add, Commit, Push)
// @Description Stage all changes, commit them with a message, and optionally push to remote.
// @Tags Workspace
// @Param key path string true "Repo Key"
// @Param request body SubmitChangesReq true "Submit info"
// @Produce json
// @Success 200 {object} response.Response{data=map[string]string}
// @Failure 400 {object} response.Response "Bad Request"
// @Failure 404 {object} response.Response "Repo not found"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/repos/{key}/submit [post]
func SubmitChanges(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")

	var req SubmitChangesReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Message == "" {
		response.BadRequest(c, "commit message is required")
		return
	}

	var repo model.Repo
	if err := dal.DB.Where("key = ?", key).First(&repo).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := service.NewGitService()

	// 1. Add .
	if err := gitSvc.AddAll(repo.Path); err != nil {
		response.InternalServerError(c, "Failed to stage files: "+err.Error())
		return
	}

	// 2. Commit
	// We might want to append status to message like the script, but usually simple message is fine.
	// The user script appended status snapshot. Let's replicate that if we want exact parity, 
	// but for UI it's better to just use what user typed. 
	// However, the prompt said "merge git status output and user input".
	// Let's do that to strictly follow "This feature on page".
	
	status, _ := gitSvc.GetStatus(repo.Path)
	fullMsg := fmt.Sprintf("%s\n\nGit Status Snapshot:\n%s", req.Message, status)

	if err := gitSvc.Commit(repo.Path, fullMsg); err != nil {
		// Rollback stage? git reset HEAD .
		_, _ = gitSvc.RunCommand(repo.Path, "reset", "HEAD", ".")
		response.InternalServerError(c, "Failed to commit: "+err.Error())
		return
	}

	msg := "Committed successfully"

	// 3. Push (Optional)
	if req.Push {
		if err := gitSvc.PushCurrent(repo.Path); err != nil {
			msg += ", but push failed: " + err.Error()
			// We don't fail the request because commit succeeded
			response.Success(c, map[string]string{"message": msg, "warning": "push_failed"})
			return
		}
		msg += " and pushed to remote"
	}

	service.AuditSvc.Log(c, "SUBMIT_CHANGES", "repo:"+repo.Key, map[string]interface{}{
		"message": req.Message,
		"push":    req.Push,
	})

	response.Success(c, map[string]string{"message": msg})
}
