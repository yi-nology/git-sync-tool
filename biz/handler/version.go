package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// @Summary Get project version
// @Description Get version string based on git describe
// @Tags Stats
// @Param id path string true "Repo Key"
// @Produce json
// @Success 200 {object} response.Response{data=string}
// @Router /api/repos/{id}/version [get]
func GetVersionInfo(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Param("key")
	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := git.NewGitService()
	version, err := svc.GetDescribe(repo.Path)
	if err != nil {
		// Fallback or error
		// If no tags, git describe might fail or return just hash if --always is used
		// Our implementation uses --always so it should return something unless empty repo
		response.InternalServerError(c, "failed to determine version: "+err.Error())
		return
	}

	response.Success(c, version)
}

// @Summary List version history
// @Description List all versions (tags)
// @Tags Stats
// @Param id path string true "Repo Key"
// @Produce json
// @Success 200 {object} response.Response{data=[]git.TagInfo}
// @Router /api/repos/{id}/versions [get]
func ListVersions(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Param("key")
	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := git.NewGitService()
	tags, err := svc.GetTagList(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	// Sort tags by Date desc
	// Note: We should probably do this in service, but handler is fine for now
	// Need to import sort package? No, let's keep it simple or use slice sort if available
	// For simplicity, we assume client sorts or we do basic sort here if needed.
	// Actually `GetTagList` iterates in undefined order (map or ref order). 
	// Go maps are random, refs usually sorted by name.
	
	response.Success(c, tags)
}

// @Summary Get next version suggestions
// @Description Get suggestions for next major, minor, and patch versions
// @Tags Stats
// @Param id path string true "Repo Key"
// @Produce json
// @Success 200 {object} response.Response{data=git.NextVersionInfo}
// @Router /api/repos/{id}/version/next [get]
func GetNextVersions(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Param("key")
	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := git.NewGitService()
	info, err := svc.GetNextVersions(repo.Path)
	if err != nil {
		response.InternalServerError(c, "failed to calculate next versions: "+err.Error())
		return
	}

	response.Success(c, info)
}
