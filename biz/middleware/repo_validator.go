package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/constants"
	"github.com/yi-nology/git-manage-service/pkg/logger"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// ValidateRepo 验证 repo_key 参数并注入 repo 到 Context
// 使用方式: router.GET("/path", middleware.ValidateRepo(), handler)
func ValidateRepo() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		repoKey := c.Query("repo_key")
		if repoKey == "" {
			logger.Warn("repo_key is required", logrus.Fields{
				"path":   string(c.Path()),
				"method": string(c.Method()),
			})
			response.BadRequest(c, "repo_key is required")
			c.Abort()
			return
		}

		repo, err := db.NewRepoDAO().FindByKey(repoKey)
		if err != nil {
			logger.Warn("repository not found", logrus.Fields{
				"repo_key": repoKey,
				"error":    err.Error(),
			})
			response.NotFound(c, "repository not found")
			c.Abort()
			return
		}

		// 将 repo 注入到 Context
		c.Set(constants.ContextKeyRepo, repo)
		c.Next(ctx)
	}
}

// ValidateRepoFromPath 从路径参数验证 repo
func ValidateRepoFromPath(paramName string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		repoKey := c.Param(paramName)
		if repoKey == "" {
			response.BadRequest(c, paramName+" is required")
			c.Abort()
			return
		}

		repo, err := db.NewRepoDAO().FindByKey(repoKey)
		if err != nil {
			response.NotFound(c, "repository not found")
			c.Abort()
			return
		}

		c.Set(constants.ContextKeyRepo, repo)
		c.Next(ctx)
	}
}

// GetRepo 从 Context 获取已验证的仓库
func GetRepo(c *app.RequestContext) *po.Repo {
	if repo, exists := c.Get(constants.ContextKeyRepo); exists {
		if r, ok := repo.(*po.Repo); ok {
			return r
		}
	}
	return nil
}

// MustGetRepo 从 Context 获取仓库，不存在则 panic
func MustGetRepo(c *app.RequestContext) *po.Repo {
	repo := GetRepo(c)
	if repo == nil {
		panic("repo not found in context")
	}
	return repo
}

// OptionalRepo 可选的 repo 验证，如果提供了 repo_key 则验证并注入
func OptionalRepo() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		repoKey := c.Query("repo_key")
		if repoKey == "" {
			c.Next(ctx)
			return
		}

		repo, err := db.NewRepoDAO().FindByKey(repoKey)
		if err != nil {
			response.NotFound(c, "repository not found")
			c.Abort()
			return
		}

		c.Set(constants.ContextKeyRepo, repo)
		c.Next(ctx)
	}
}
