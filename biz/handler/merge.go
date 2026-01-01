package handler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
	"github.com/yi-nology/git-manage-service/biz/pkg/response"
	"github.com/yi-nology/git-manage-service/biz/service"
)

// @Summary Compare two branches
// @Tags Merge
// @Param id path int true "Repo ID"
// @Param base query string true "Base Branch"
// @Param target query string true "Target Branch"
// @Success 200 {object} response.Response
// @Router /api/repos/{id}/compare [get]
func CompareBranches(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	base := c.Query("base")
	target := c.Query("target")

	if base == "" || target == "" {
		response.BadRequest(c, "base and target branches are required")
		return
	}

	var repo model.Repo
	if err := dal.DB.First(&repo, id).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := service.NewGitService()
	
	stat, err := gitSvc.GetDiffStat(repo.Path, base, target)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	
	files, err := gitSvc.GetDiffFiles(repo.Path, base, target)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]interface{}{
		"stat":  stat,
		"files": files,
	})
}

// @Summary Get raw diff content
// @Tags Merge
// @Param id path int true "Repo ID"
// @Param base query string true "Base Branch"
// @Param target query string true "Target Branch"
// @Param file query string false "Specific File"
// @Success 200 {object} response.Response
// @Router /api/repos/{id}/diff [get]
func GetDiffContent(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	base := c.Query("base")
	target := c.Query("target")
	file := c.Query("file")

	var repo model.Repo
	if err := dal.DB.First(&repo, id).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := service.NewGitService()
	content, err := gitSvc.GetRawDiff(repo.Path, base, target, file)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, map[string]string{"diff": content})
}

// @Summary Dry run merge to check conflicts
// @Tags Merge
// @Param id path int true "Repo ID"
// @Param base query string true "Base Branch"
// @Param target query string true "Target Branch"
// @Success 200 {object} response.Response
// @Router /api/repos/{id}/merge/check [get]
func MergeCheck(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	base := c.Query("base")   // Source (feature)
	target := c.Query("target") // Destination (main)

	var repo model.Repo
	if err := dal.DB.First(&repo, id).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := service.NewGitService()
	result, err := gitSvc.MergeDryRun(repo.Path, base, target)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	
	response.Success(c, result)
}

type MergeReq struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Message  string `json:"message"`
	Strategy string `json:"strategy"` // Not implemented yet
}

// @Summary Execute merge
// @Tags Merge
// @Param id path int true "Repo ID"
// @Param request body MergeReq true "Merge Info"
// @Success 200 {object} response.Response
// @Router /api/repos/{id}/merge [post]
func ExecuteMerge(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	
	var req MergeReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var repo model.Repo
	if err := dal.DB.First(&repo, id).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := service.NewGitService()
	
	// Double check conflicts
	check, err := gitSvc.MergeDryRun(repo.Path, req.Source, req.Target)
	if err != nil {
		response.InternalServerError(c, "Pre-merge check failed: "+err.Error())
		return
	}
	if !check.Success {
		// Generate conflict report URL
		mergeID := uuid.New().String()
		// Use a static HTML page that takes params to show the report
		reportURL := fmt.Sprintf("/merge_report.html?repo_id=%d&source=%s&target=%s&merge_id=%s", id, req.Source, req.Target, mergeID)
		
		// Log conflict
		service.AuditSvc.Log(c, "MERGE_CONFLICT", "repo:"+repo.Key, map[string]interface{}{
			"source": req.Source,
			"target": req.Target,
			"conflicts": check.Conflicts,
			"merge_id": mergeID,
		})
		
		c.JSON(200, response.Response{
			Code: 409, // Conflict
			Message: "Merge conflict detected",
			Data: map[string]interface{}{
				"conflicts": check.Conflicts,
				"report_url": reportURL,
				"merge_id": mergeID,
			},
		})
		return
	}

	// Proceed with merge
	if err := gitSvc.Merge(repo.Path, req.Source, req.Target, req.Message); err != nil {
		response.InternalServerError(c, "Merge execution failed: "+err.Error())
		return
	}
	
	service.AuditSvc.Log(c, "MERGE_SUCCESS", "repo:"+repo.Key, map[string]string{
		"source": req.Source,
		"target": req.Target,
	})
	
	response.Success(c, map[string]string{"status": "merged"})
}

// @Summary Download patch
// @Tags Merge
// @Param id path int true "Repo ID"
// @Param base query string true "Base"
// @Param target query string true "Target"
// @Success 200 {file} octet-stream
// @Router /api/repos/{id}/patch [get]
func GetPatch(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	base := c.Query("base")
	target := c.Query("target")

	var repo model.Repo
	if err := dal.DB.First(&repo, id).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := service.NewGitService()
	patch, err := gitSvc.GetPatch(repo.Path, base, target)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-%s-%s.patch", repo.Name, base, time.Now().Format("20060102")))
	c.Write([]byte(patch))
}
