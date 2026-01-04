package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// @Summary Create a new tag
// @Description Create a tag on a specific branch or commit, optionally push to remote
// @Tags Tag
// @Accept json
// @Produce json
// @Param id path string true "Repo Key"
// @Param body body api.CreateTagReq true "Tag Info"
// @Success 200 {object} response.Response
// @Router /api/repos/{id}/tags [post]
func CreateTag(ctx context.Context, c *app.RequestContext) {
	// The route definition uses :key, so we must use "key" to retrieve it.
	// In register.go: repos.POST("/:key/tags", handler.CreateTag)
	repoKey := c.Param("key")
	var req api.CreateTagReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	// 1. Create Tag
	svc := git.NewGitService()
	// Get global author info if available, or rely on service default
	authorName, authorEmail, _ := svc.GetGlobalGitUser()

	// Auto-increment version logic
	if req.TagName == "auto" {
		latest, err := svc.GetLatestVersion(repo.Path)
		if err != nil || latest == "" {
			req.TagName = "v0.1.0" // Default start
		} else {
			// Parse latest (e.g. v1.0.1 or v1.0.1-5-g...)
			// If it has commits suffix, strip it?
			// Actually `GetLatestVersion` (describe --abbrev=0) returns the tag name itself like v1.0.1.
			// Simple increment logic: try to parse semver
			// This is a naive implementation string manipulation
			// v1.0.1 -> v1.0.2
			req.TagName = incrementVersion(latest)
		}
	}

	err = svc.CreateTag(repo.Path, req.TagName, req.Ref, req.Message, authorName, authorEmail)
	if err != nil {
		response.InternalServerError(c, "failed to create tag: "+err.Error())
		return
	}

	// 2. Push if requested
	if req.PushRemote != "" {
		// Need auth info
		authType := "none"
		authKey := ""
		authSecret := ""

		// Check if it's using DB auth
		if repo.ConfigSource == "database" {
			// Find remote config in DB? Current DB schema might store remotes in a separate table or JSON?
			// The current implementation seems to rely on `repo.Path` and `.git/config` mostly,
			// except for Clone/Sync tasks which store auth in `sync_tasks` table.
			// However, `GitRemote` auth might be stored in `git_remotes` table if it exists?
			// Checking `biz/dal/db` for `RemoteDAO`...
			// If not, we might rely on the fact that `PushTag` service method detects SSH keys or we pass nothing for now.
			// Ideally we should reuse the auth logic from `PushBranch`.

			// For now, let's assume local SSH keys or no auth for HTTP (unless cached).
			// Improving: check if we have stored auth for this remote.
		}

		// Simple push attempt
		err = svc.PushTag(repo.Path, req.PushRemote, req.TagName, authType, authKey, authSecret)
		if err != nil {
			// Tag created but push failed
			response.Success(c, map[string]string{
				"status": "created_local_only",
				"error":  "tag created but push failed: " + err.Error(),
			})
			return
		}
	}

	response.Success(c, nil)
}

// @Summary List tags
// @Description List all tags in the repository
// @Tags Tag
// @Param id path string true "Repo Key"
// @Produce json
// @Success 200 {object} response.Response{data=[]string}
// @Router /api/repos/{id}/tags [get]
func ListTags(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Param("key")
	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := git.NewGitService()
	tags, err := svc.GetTags(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, tags)
}

func incrementVersion(v string) string {
	// Remove v prefix
	hasV := false
	if len(v) > 0 && v[0] == 'v' {
		hasV = true
		v = v[1:]
	}

	// Split .
	parts := strings.Split(v, ".")
	if len(parts) > 0 {
		lastIdx := len(parts) - 1
		lastNum, err := strconv.Atoi(parts[lastIdx])
		if err == nil {
			parts[lastIdx] = strconv.Itoa(lastNum + 1)
		} else {
			// If last part is not number, append .1
			parts = append(parts, "1")
		}
	} else {
		return "v0.0.1"
	}

	res := strings.Join(parts, ".")
	if hasV {
		return "v" + res
	}
	return res
}
