package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
	"github.com/yi-nology/git-manage-service/biz/pkg/response"
	"github.com/yi-nology/git-manage-service/biz/service"
)

// @Summary Compare two branches
// @Description Compare two branches and return diff statistics and file list.
// @Tags Merge
// @Param key path string true "Repo Key"
// @Param base query string true "Base Branch"
// @Param target query string true "Target Branch"
// @Success 200 {object} response.Response{data=map[string]interface{}} "Map with 'stat' and 'files'"
// @Failure 400 {object} response.Response "Bad Request - Missing params"
// @Failure 404 {object} response.Response "Repo not found"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/repos/{key}/compare [get]
func CompareBranches(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")
	base := c.Query("base")
	target := c.Query("target")

	if base == "" || target == "" {
		response.BadRequest(c, "base and target branches are required")
		return
	}

	var repo model.Repo
	if err := dal.DB.Where("key = ?", key).First(&repo).Error; err != nil {
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
// @Description Get the raw diff content between two branches, optionally for a specific file.
// @Tags Merge
// @Param key path string true "Repo Key"
// @Param base query string true "Base Branch"
// @Param target query string true "Target Branch"
// @Param file query string false "Specific File"
// @Success 200 {object} response.Response{data=map[string]string} "Map with 'diff'"
// @Failure 404 {object} response.Response "Repo not found"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/repos/{key}/diff [get]
func GetDiffContent(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")
	base := c.Query("base")
	target := c.Query("target")
	file := c.Query("file")

	var repo model.Repo
	if err := dal.DB.Where("key = ?", key).First(&repo).Error; err != nil {
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
// @Description Perform a dry run of a merge to check for conflicts without modifying the repository.
// @Tags Merge
// @Param key path string true "Repo Key"
// @Param base query string true "Base Branch"
// @Param target query string true "Target Branch"
// @Success 200 {object} response.Response "Success status, does not mean no conflict, check data"
// @Failure 404 {object} response.Response "Repo not found"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/repos/{key}/merge/check [get]
func MergeCheck(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")
	base := c.Query("base")     // Source (feature)
	target := c.Query("target") // Destination (main)

	var repo model.Repo
	if err := dal.DB.Where("key = ?", key).First(&repo).Error; err != nil {
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
// @Param key path string true "Repo Key"
// @Param request body MergeReq true "Merge Info"
// @Success 200 {object} response.Response
// @Router /api/repos/{key}/merge [post]
func ExecuteMerge(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")

	var req MergeReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var repo model.Repo
	if err := dal.DB.Where("key = ?", key).First(&repo).Error; err != nil {
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
		reportURL := fmt.Sprintf("/merge_report.html?repo_key=%s&source=%s&target=%s&merge_id=%s", repo.Key, req.Source, req.Target, mergeID)

		// Log conflict
		service.AuditSvc.Log(c, "MERGE_CONFLICT", "repo:"+repo.Key, map[string]interface{}{
			"source":    req.Source,
			"target":    req.Target,
			"conflicts": check.Conflicts,
			"merge_id":  mergeID,
		})

		c.JSON(200, response.Response{
			Code:    409, // Conflict
			Message: "Merge conflict detected",
			Data: map[string]interface{}{
				"conflicts":  check.Conflicts,
				"report_url": reportURL,
				"merge_id":   mergeID,
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
// @Param key path string true "Repo Key"
// @Param base query string true "Base"
// @Param target query string true "Target"
// @Success 200 {file} octet-stream
// @Router /api/repos/{key}/patch [get]
func GetPatch(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")
	base := c.Query("base")
	target := c.Query("target")

	var repo model.Repo
	if err := dal.DB.Where("key = ?", key).First(&repo).Error; err != nil {
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
