package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
	"github.com/yi-nology/git-manage-service/biz/pkg/response"
	"github.com/yi-nology/git-manage-service/biz/service"
)

// @Summary List branches for a repository
// @Description List branches for statistics (simplified list).
// @Tags Stats
// @Param repo_key query string true "Repo Key"
// @Produce json
// @Success 200 {object} response.Response{data=[]string}
// @Failure 404 {object} response.Response "Repo not found"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/stats/branches [get]
func ListBranches(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")

	var repo model.Repo
	if err := dal.DB.Where("key = ?", repoKey).First(&repo).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	branches, err := service.NewGitService().GetBranches(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, branches)
}

// @Summary Get commit history for a branch
// @Tags Stats
// @Param repo_key query string true "Repo Key"
// @Param branch query string true "Branch Name"
// @Param since query string false "Since (YYYY-MM-DD)"
// @Param until query string false "Until (YYYY-MM-DD)"
// @Produce json
// @Success 200 {object} response.Response{data=[]model.Commit}
// @Router /api/stats/commits [get]
func ListCommits(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	branch := c.Query("branch")
	since := c.Query("since")
	until := c.Query("until")

	var repo model.Repo
	if err := dal.DB.Where("key = ?", repoKey).First(&repo).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	raw, err := service.NewGitService().GetCommits(repo.Path, branch, since, until)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	commits := service.StatsSvc.ParseCommits(raw)
	response.Success(c, commits)
}

// @Summary Get code statistics for a branch
// @Description Analyze code statistics (author contributions, file types, etc.) for a branch.
// @Tags Stats
// @Param repo_key query string true "Repo Key"
// @Param branch query string true "Branch Name"
// @Param since query string false "Since (YYYY-MM-DD)"
// @Param until query string false "Until (YYYY-MM-DD)"
// @Param author query string false "Filter by Author Name or Email"
// @Produce json
// @Success 200 {object} response.Response{data=model.StatsResponse}
// @Failure 404 {object} response.Response "Repo not found"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/stats/analyze [get]
func GetStats(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	branch := c.Query("branch")
	since := c.Query("since")
	until := c.Query("until")

	var repo model.Repo
	if err := dal.DB.Where("key = ?", repoKey).First(&repo).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	// This might take a while, consider async or cache
	// For now, we run it synchronously
	stats, err := service.StatsSvc.CalculateStats(repo.Path, branch, since, until)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	// Filter by author if requested
	author := c.Query("author")
	if author != "" {
		filtered := []*model.AuthorStat{}
		for _, a := range stats.Authors {
			if a.Name == author || a.Email == author {
				filtered = append(filtered, a)
			}
		}
		stats.Authors = filtered
		// Recalculate total lines based on filter
		total := 0
		for _, a := range filtered {
			total += a.TotalLines
		}
		stats.TotalLines = total
	}

	response.Success(c, stats)
}

// @Summary Export statistics as CSV
// @Tags Stats
// @Param repo_key query string true "Repo Key"
// @Param branch query string true "Branch Name"
// @Param since query string false "Since (YYYY-MM-DD)"
// @Param until query string false "Until (YYYY-MM-DD)"
// @Produce text/csv
// @Router /api/stats/export/csv [get]
func ExportStatsCSV(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	branch := c.Query("branch")
	since := c.Query("since")
	until := c.Query("until")

	var repo model.Repo
	if err := dal.DB.Where("key = ?", repoKey).First(&repo).Error; err != nil {
		c.JSON(consts.StatusNotFound, map[string]string{"error": "repo not found"})
		return
	}

	stats, err := service.StatsSvc.CalculateStats(repo.Path, branch, since, until)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=stats-%s-%s.csv", repo.Name, time.Now().Format("20060102")))

	// Write CSV Header
	c.Write([]byte("Author,Email,Total Effective Lines,Top Language\n"))

	for _, author := range stats.Authors {
		topLang := ""
		maxLines := 0
		for lang, count := range author.FileTypes {
			if count > maxLines {
				maxLines = count
				topLang = lang
			}
		}
		line := fmt.Sprintf("%s,%s,%d,%s\n", author.Name, author.Email, author.TotalLines, topLang)
		c.Write([]byte(line))
	}
}
