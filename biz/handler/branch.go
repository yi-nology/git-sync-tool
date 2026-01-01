package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
	"github.com/yi-nology/git-manage-service/biz/pkg/response"
	"github.com/yi-nology/git-manage-service/biz/service"
)

// @Summary List branches
// @Description List branches for a repository with pagination and filtering. Includes ahead/behind sync status if available.
// @Tags Branches
// @Param key path string true "Repo Key"
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Page size (default 100)"
// @Param keyword query string false "Search keyword (name or author)"
// @Success 200 {object} response.Response{data=map[string]interface{}} "Map containing 'total' and 'list' of BranchInfo"
// @Failure 404 {object} response.Response "Repo not found"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/repos/{key}/branches [get]
func ListRepoBranches(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")

	var repo model.Repo
	if err := dal.DB.Where("key = ?", key).First(&repo).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := service.NewGitService()
	branches, err := gitSvc.ListBranchesWithInfo(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	// Filter
	keyword := c.Query("keyword")
	if keyword != "" {
		var filtered []model.BranchInfo
		keyword = strings.ToLower(keyword)
		for _, b := range branches {
			if strings.Contains(strings.ToLower(b.Name), keyword) ||
				strings.Contains(strings.ToLower(b.Author), keyword) {
				filtered = append(filtered, b)
			}
		}
		branches = filtered
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize < 1 {
		pageSize = 100 // Default high for now
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(branches) {
		start = len(branches)
	}
	if end > len(branches) {
		end = len(branches)
	}

	paged := branches[start:end]

	// Enrich with description and sync status
	// We might need to fetch first to ensure status is up to date, but that's slow.
	// Let's assume background fetch or manual fetch.
	// Or we can do a quick fetch --all if page=1? No, too slow.

	for i := range paged {
		b := &paged[i]
		// Description
		desc, _ := gitSvc.GetBranchDescription(repo.Path, b.Name)
		_ = desc // Ignored for now as it's not in struct, or add it?

		// Sync Status
		if b.Upstream != "" {
			ahead, behind, _ := gitSvc.GetBranchSyncStatus(repo.Path, b.Name, b.Upstream)
			b.Ahead = ahead
			b.Behind = behind
		}
	}

	// Return result with total count
	response.Success(c, map[string]interface{}{
		"total": len(branches),
		"list":  paged,
	})
}

// @Summary Create a branch
// @Tags Branches
// @Param key path string true "Repo Key"
// @Param request body model.CreateBranchReq true "Create info"
// @Success 200 {object} response.Response
// @Router /api/repos/{key}/branches [post]
func CreateBranch(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")

	var req model.CreateBranchReq
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
	if err := gitSvc.CreateBranch(repo.Path, req.Name, req.BaseRef); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	service.AuditSvc.Log(c, "CREATE_BRANCH", "repo:"+repo.Key, map[string]string{
		"branch": req.Name,
		"base":   req.BaseRef,
	})
	response.Success(c, map[string]string{"message": "created"})
}

// @Summary Delete a branch
// @Tags Branches
// @Param key path string true "Repo Key"
// @Param name path string true "Branch Name"
// @Param force query bool false "Force delete"
// @Success 200 {object} response.Response
// @Router /api/repos/{key}/branches/{name} [delete]
func DeleteBranch(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")
	name := c.Param("name")
	force := c.Query("force") == "true"

	var repo model.Repo
	if err := dal.DB.Where("key = ?", key).First(&repo).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := service.NewGitService()
	if err := gitSvc.DeleteBranch(repo.Path, name, force); err != nil {
		// If error says "not fully merged", client should prompt to use force
		response.InternalServerError(c, err.Error())
		return
	}

	service.AuditSvc.Log(c, "DELETE_BRANCH", "repo:"+repo.Key, map[string]string{
		"branch": name,
		"force":  strconv.FormatBool(force),
	})
	response.Success(c, map[string]string{"message": "deleted"})
}

// @Summary Update a branch (Rename/Desc)
// @Description Rename a branch or update its description.
// @Tags Branches
// @Param key path string true "Repo Key"
// @Param name path string true "Current Branch Name"
// @Param request body model.UpdateBranchReq true "Update info"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "Bad Request"
// @Failure 404 {object} response.Response "Repo not found"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/repos/{key}/branches/{name} [put]
func UpdateBranch(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")
	currentName := c.Param("name")

	var req model.UpdateBranchReq
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

	// Rename
	if req.NewName != "" && req.NewName != currentName {
		if err := gitSvc.RenameBranch(repo.Path, currentName, req.NewName); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		// Update currentName for subsequent ops (like desc)
		currentName = req.NewName
	}

	// Description
	if req.Desc != "" {
		if err := gitSvc.SetBranchDescription(repo.Path, currentName, req.Desc); err != nil {
			// Log but maybe not fail?
		}
	}

	service.AuditSvc.Log(c, "UPDATE_BRANCH", "repo:"+repo.Key, map[string]string{
		"old_name": c.Param("name"), // Original name
		"new_name": req.NewName,
		"desc":     req.Desc,
	})
	response.Success(c, map[string]string{"message": "updated"})
}
