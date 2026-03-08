package spec

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"

	"github.com/yi-nology/git-manage-service/biz/service/audit"
	lintSvc "github.com/yi-nology/git-manage-service/biz/service/lint"
	specService "github.com/yi-nology/git-manage-service/biz/service/spec"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func GetSpecTree(ctx context.Context, c *app.RequestContext) {
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

	tree, err := buildSpecTree(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, tree)
}

func buildSpecTree(repoPath string) ([]api.SpecFile, error) {
	// 收集所有 .spec 文件和目录
	nodes := make(map[string]*api.SpecFile)

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过 .git 目录
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		relPath, _ := filepath.Rel(repoPath, path)

		// 只收集 .spec 文件和目录
		if strings.HasSuffix(info.Name(), ".spec") || info.IsDir() {
			nodes[relPath] = &api.SpecFile{
				Name:    info.Name(),
				Path:    relPath,
				IsDir:   info.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime(),
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 使用 childrenMap 存储每个路径的子节点（指针）
	childrenMap := make(map[string][]*api.SpecFile)

	// 构建父子关系
	for path, node := range nodes {
		if path == "." {
			continue
		}

		parentPath := filepath.Dir(path)
		if parentPath == "" {
			parentPath = "."
		}

		childrenMap[parentPath] = append(childrenMap[parentPath], node)
	}

	// 递归构建树（使用指针）
	var buildTree func(path string) *api.SpecFile
	buildTree = func(path string) *api.SpecFile {
		node := nodes[path]
		if node == nil {
			return nil
		}

		// 获取子节点
		children := childrenMap[path]
		if len(children) > 0 {
			node.Children = make([]api.SpecFile, 0, len(children))
			for _, child := range children {
				// 递归构建子树
				buildTree(child.Path)
				node.Children = append(node.Children, *child)
			}
		}

		return node
	}

	// 从根节点开始构建
	root := buildTree(".")
	if root == nil {
		return []api.SpecFile{}, nil
	}

	// 过滤：只保留包含 .spec 文件的目录
	filterTree(root)

	if len(root.Children) > 0 {
		return root.Children, nil
	}

	return []api.SpecFile{}, nil
}

// createDirChain 创建目录链
func createDirChain(pathMap map[string]*api.SpecFile, path string, repoPath string) *api.SpecFile {
	if path == "." || path == "" {
		return pathMap["."]
	}

	// 检查是否已存在
	if dir, exists := pathMap[path]; exists {
		return dir
	}

	// 创建当前目录
	info, err := os.Stat(filepath.Join(repoPath, path))
	if err != nil {
		return nil
	}

	dir := &api.SpecFile{
		Name:    filepath.Base(path),
		Path:    path,
		IsDir:   true,
		ModTime: info.ModTime(),
	}
	pathMap[path] = dir

	// 递归创建父目录
	parentPath := filepath.Dir(path)
	if parentPath == "" {
		parentPath = "."
	}

	parent := createDirChain(pathMap, parentPath, repoPath)
	if parent != nil {
		parent.Children = append(parent.Children, *dir)
	}

	return dir
}

// filterTree 过滤树，只保留包含 .spec 文件的目录
func filterTree(node *api.SpecFile) bool {
	if !node.IsDir {
		// 文件：如果是 .spec 文件则保留
		return strings.HasSuffix(node.Name, ".spec")
	}

	// 目录：检查子节点
	var hasSpecFile bool
	var filteredChildren []api.SpecFile

	for i := range node.Children {
		if filterTree(&node.Children[i]) {
			filteredChildren = append(filteredChildren, node.Children[i])
			hasSpecFile = true
		}
	}

	node.Children = filteredChildren
	return hasSpecFile
}

func GetSpecContentByPath(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	path := c.Param("path")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}
	if path == "" {
		path = c.Query("path")
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()
	content, err := svc.GetSpecContent(repo.Path, path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, api.FileContent{
		Path:    path,
		Content: content,
	})
}

func SaveSpecContentByPath(ctx context.Context, c *app.RequestContext) {
	path := c.Param("path")
	if path == "" {
		response.BadRequest(c, "path is required")
		return
	}

	var req api.SaveSpecContentReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Path == "" {
		req.Path = path
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()

	err = svc.SaveSpecContent(repo.Path, req.Path, req.Content, req.Message)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	if req.AutoCommit && req.Message != "" {
		cmd := exec.Command("git", "add", req.Path)
		cmd.Dir = repo.Path
		if output, err := cmd.CombinedOutput(); err != nil {
			response.InternalServerError(c, "Failed to stage: "+string(output))
			return
		}

		cmd = exec.Command("git", "commit", "-m", req.Message)
		cmd.Dir = repo.Path
		if output, err := cmd.CombinedOutput(); err != nil {
			if !strings.Contains(string(output), "nothing to commit") {
				response.InternalServerError(c, "Failed to commit: "+string(output))
				return
			}
		}
	}

	audit.AuditSvc.Log(c, "SAVE_SPEC", "repo:"+repo.Key, map[string]string{
		"path":    req.Path,
		"message": req.Message,
	})

	response.Success(c, map[string]string{
		"message": "spec saved successfully",
		"path":    req.Path,
	})
}

func LintSpec(ctx context.Context, c *app.RequestContext) {
	var req api.LintRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Content == "" {
		response.BadRequest(c, "content is required")
		return
	}

	lintService := lintSvc.NewLintService()
	result, err := lintService.Lint(req.Content, req.Rules)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func GetLintRules(ctx context.Context, c *app.RequestContext) {
	rules, err := db.NewLintRuleDAO().FindAll()
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	var dtos []api.LintRule
	for _, r := range rules {
		dtos = append(dtos, api.LintRule{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Category:    r.Category,
			Severity:    r.Severity,
			Pattern:     r.Pattern,
			Enabled:     r.Enabled,
			Priority:    r.Priority,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		})
	}

	if dtos == nil {
		dtos = []api.LintRule{}
	}

	response.Success(c, dtos)
}

func UpdateLintRule(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "rule id is required")
		return
	}

	var req api.UpdateLintRuleReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dao := db.NewLintRuleDAO()
	rule, err := dao.FindByID(id)
	if err != nil {
		response.NotFound(c, "rule not found")
		return
	}

	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.Description != "" {
		rule.Description = req.Description
	}
	if req.Category != "" {
		rule.Category = req.Category
	}
	if req.Severity != "" {
		rule.Severity = req.Severity
	}
	if req.Pattern != "" {
		rule.Pattern = req.Pattern
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}

	if err := dao.Save(rule); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, api.LintRule{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Category:    rule.Category,
		Severity:    rule.Severity,
		Pattern:     rule.Pattern,
		Enabled:     rule.Enabled,
		Priority:    rule.Priority,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	})
}

func CreateLintRule(ctx context.Context, c *app.RequestContext) {
	var req api.CreateLintRuleReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.ID == "" {
		response.BadRequest(c, "id is required")
		return
	}
	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}

	dao := db.NewLintRuleDAO()
	exists, _ := dao.ExistsByID(req.ID)
	if exists {
		response.BadRequest(c, "rule with this id already exists")
		return
	}

	rule := &po.LintRule{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Severity:    req.Severity,
		Pattern:     req.Pattern,
		Enabled:     req.Enabled,
		Priority:    req.Priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if rule.Category == "" {
		rule.Category = "custom"
	}
	if rule.Severity == "" {
		rule.Severity = "warning"
	}

	if err := dao.Create(rule); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, api.LintRule{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Category:    rule.Category,
		Severity:    rule.Severity,
		Pattern:     rule.Pattern,
		Enabled:     rule.Enabled,
		Priority:    rule.Priority,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	})
}

func CommitSpec(ctx context.Context, c *app.RequestContext) {
	path := c.Param("path")
	if path == "" {
		response.BadRequest(c, "path is required")
		return
	}

	var req api.CommitSpecReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Message == "" {
		response.BadRequest(c, "message is required")
		return
	}

	if req.Path == "" {
		req.Path = path
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	if req.Content != "" {
		svc := specService.NewSpecService()
		if err := svc.SaveSpecContent(repo.Path, req.Path, req.Content, ""); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
	}

	cmd := exec.Command("git", "add", req.Path)
	cmd.Dir = repo.Path
	if output, err := cmd.CombinedOutput(); err != nil {
		response.InternalServerError(c, "Failed to stage: "+string(output))
		return
	}

	cmd = exec.Command("git", "commit", "-m", req.Message)
	cmd.Dir = repo.Path
	output, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(output), "nothing to commit") {
			response.InternalServerError(c, "Failed to commit: "+string(output))
			return
		}
	}

	audit.AuditSvc.Log(c, "COMMIT_SPEC", "repo:"+repo.Key, map[string]string{
		"path":    req.Path,
		"message": req.Message,
	})

	response.Success(c, map[string]string{
		"message": "committed successfully",
		"output":  string(output),
	})
}

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

	svc := specService.NewSpecService()
	files, err := svc.ListSpecFiles(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	// 确保返回空数组而不是 null
	if files == nil {
		files = []specService.SpecFileInfo{}
	}

	// 转换为 API DTO
	var dtos []api.SpecFileInfo
	for _, f := range files {
		dtos = append(dtos, api.SpecFileInfo{
			Name:    f.Name,
			Path:    f.Path,
			IsDir:   f.IsDir,
			Size:    f.Size,
			ModTime: f.ModTime,
		})
	}

	response.Success(c, dtos)
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

	svc := specService.NewSpecService()
	content, err := svc.GetSpecContent(repo.Path, path)
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

	svc := specService.NewSpecService()

	// 先验证
	validationResult := svc.ValidateSpec(req.Content)
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
	err = svc.SaveSpecContent(repo.Path, req.Path, req.Content, req.CommitMessage)
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
		"message":           "spec saved successfully",
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

	svc := specService.NewSpecService()
	result := svc.ValidateSpec(req.Content)

	response.Success(c, result)
}

// GetSpecRules 获取 spec 规则列表
// @router /api/v1/spec/rules [GET]
func GetSpecRules(ctx context.Context, c *app.RequestContext) {
	svc := specService.NewSpecService()
	rules := svc.GetBuiltinRules()

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

	svc := specService.NewSpecService()
	path, err := svc.CreateSpecFileWithContent(repo.Path, req.Path, req.Name, req.Content)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	audit.AuditSvc.Log(c, "CREATE_SPEC", "repo:"+repo.Key, map[string]string{
		"path": path,
	})

	response.Success(c, map[string]string{
		"path":    path,
		"message": "Spec 文件创建成功",
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

	svc := specService.NewSpecService()
	err = svc.DeleteSpecFile(repo.Path, req.Path)
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
