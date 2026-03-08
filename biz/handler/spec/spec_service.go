package spec

import (
	"context"
	"os/exec"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
	"github.com/yi-nology/git-manage-service/biz/service/spec"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// ListSpecFiles 列出仓库中的 spec 文件
// @router /api/v1/spec/list [GET]
func ListSpecFiles(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	specSvc := spec.NewSpecService()
	files, err := specSvc.ListSpecFiles(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, files)
}

// GetSpecContent 获取 spec 文件内容
// @router /api/v1/spec/content [GET]
func GetSpecContent(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	path := c.Query("path")
	if repoKey == "" || path == "" {
		response.BadRequest(c, "repo_key and path are required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	specSvc := spec.NewSpecService()
	content, err := specSvc.GetSpecContent(repo.Path, path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{
		"content": content,
		"path":    path,
	})
}

// SaveSpecContent 保存 spec 文件
// @router /api/v1/spec/save [POST]
func SaveSpecContent(ctx context.Context, c *app.RequestContext) {
	var req api.SaveSpecReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	specSvc := spec.NewSpecService()

	// 先验证
	validationResult := specSvc.ValidateSpec(req.Content)
	if !validationResult.Valid && len(validationResult.Issues) > 0 {
		// 如果有错误级别的 issue，阻止保存
		for _, issue := range validationResult.Issues {
			if issue.Severity == "error" {
				response.BadRequest(c, "Spec validation failed: "+issue.Message)
				return
			}
		}
	}

	// 保存文件
	err = specSvc.SaveSpecContent(repo.Path, req.Path, req.Content, req.CommitMessage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	// 如果有 commit message，自动提交到 git
	if req.CommitMessage != "" {
		// 直接使用 git 命令
		cmd := exec.Command("git", "add", req.Path)
		cmd.Dir = repo.Path
		if output, err := cmd.CombinedOutput(); err != nil {
			response.InternalServerError(c, "Failed to stage: "+string(output))
			return
		}

		cmd = exec.Command("git", "commit", "-m", req.CommitMessage)
		cmd.Dir = repo.Path
		if output, err := cmd.CombinedOutput(); err != nil {
			// 如果没有改动，不算错误
			if !strings.Contains(string(output), "nothing to commit") {
				response.InternalServerError(c, "Failed to commit: "+string(output))
				return
			}
		}
	}

	audit.AuditSvc.Log(c, "SAVE_SPEC", "repo:"+repo.Key, map[string]string{
		"path":           req.Path,
		"commit_message": req.CommitMessage,
	})

	response.Success(c, map[string]interface{}{
		"message":          "spec saved successfully",
		"validation_result": validationResult,
	})
}

// ValidateSpec 验证 spec 文件
// @router /api/v1/spec/validate [POST]
func ValidateSpec(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Content string `json:"content" form:"content"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	specSvc := spec.NewSpecService()
	result := specSvc.ValidateSpec(req.Content)

	response.Success(c, result)
}

// GetSpecRules 获取 spec 规则列表
// @router /api/v1/spec/rules [GET]
func GetSpecRules(ctx context.Context, c *app.RequestContext) {
	specSvc := spec.NewSpecService()
	rules := specSvc.GetBuiltinRules()

	response.Success(c, rules)
}

// CreateSpecFile 创建新的 spec 文件
// @router /api/v1/spec/create [POST]
func CreateSpecFile(ctx context.Context, c *app.RequestContext) {
	var req api.CreateSpecFileReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	specSvc := spec.NewSpecService()
	path, err := specSvc.CreateSpecFile(repo.Path, req.Path, req.Name)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	audit.AuditSvc.Log(c, "CREATE_SPEC", "repo:"+repo.Key, map[string]string{
		"path": path,
	})

	response.Success(c, map[string]string{
		"path": path,
	})
}

// DeleteSpecFile 删除 spec 文件
// @router /api/v1/spec/delete [POST]
func DeleteSpecFile(ctx context.Context, c *app.RequestContext) {
	var req api.DeleteSpecFileReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	specSvc := spec.NewSpecService()
	err = specSvc.DeleteSpecFile(repo.Path, req.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	// 如果有 commit message，自动提交
	if req.CommitMessage != "" {
		cmd := exec.Command("git", "add", req.Path)
		cmd.Dir = repo.Path
		if output, err := cmd.CombinedOutput(); err != nil {
			response.InternalServerError(c, "Failed to stage: "+string(output))
			return
		}

		cmd = exec.Command("git", "commit", "-m", req.CommitMessage)
		cmd.Dir = repo.Path
		if output, err := cmd.CombinedOutput(); err != nil {
			if !strings.Contains(string(output), "nothing to commit") {
				response.InternalServerError(c, "Failed to commit: "+string(output))
				return
			}
		}
	}

	audit.AuditSvc.Log(c, "DELETE_SPEC", "repo:"+repo.Key, map[string]string{
		"path": req.Path,
	})

	response.Success(c, map[string]string{
		"message": "spec deleted",
	})
}
