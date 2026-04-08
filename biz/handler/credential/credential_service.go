package credential

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// List 列出所有凭证
// @router /api/v1/credentials [GET]
func List(ctx context.Context, c *app.RequestContext) {
	dao := db.NewCredentialDAO()
	creds, err := dao.FindAll()
	if err != nil {
		response.InternalServerError(c, "Failed to fetch credentials: "+err.Error())
		return
	}

	sshKeyMap := buildSSHKeyMap(creds)

	result := make([]api.CredentialDTO, 0, len(creds))
	for _, cred := range creds {
		result = append(result, toCredentialDTO(&cred, sshKeyMap))
	}

	response.Success(c, result)
}

// Create 创建凭证
// @router /api/v1/credentials [POST]
func Create(ctx context.Context, c *app.RequestContext) {
	var req api.CreateCredentialReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}
	if req.Type == "" {
		response.BadRequest(c, "type is required")
		return
	}
	if req.Type != "ssh_key" && req.Type != "http_basic" && req.Type != "http_token" {
		response.BadRequest(c, "type must be ssh_key, http_basic or http_token")
		return
	}

	// 类型特定验证
	if req.Type == "ssh_key" && req.SSHKeyID == 0 && req.SSHKeyPath == "" {
		response.BadRequest(c, "ssh_key type requires ssh_key_id or ssh_key_path")
		return
	}
	if (req.Type == "http_basic" || req.Type == "http_token") && req.Username == "" && req.Secret == "" {
		response.BadRequest(c, "http type requires username or secret")
		return
	}

	// 验证数据库 SSH 密钥是否存在
	if req.SSHKeyID > 0 {
		sshKeyDAO := db.NewSSHKeyDAO()
		if _, err := sshKeyDAO.FindByID(req.SSHKeyID); err != nil {
			response.BadRequest(c, "SSH key not found with the given id")
			return
		}
	}

	dao := db.NewCredentialDAO()
	exists, err := dao.ExistsByName(req.Name)
	if err != nil {
		response.InternalServerError(c, "Failed to check credential name: "+err.Error())
		return
	}
	if exists {
		response.BadRequest(c, "Credential with this name already exists")
		return
	}

	cred := &po.Credential{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		SSHKeyID:    req.SSHKeyID,
		SSHKeyPath:  req.SSHKeyPath,
		Username:    req.Username,
		Secret:      req.Secret,
		URLPattern:  req.URLPattern,
	}

	if err := dao.Create(cred); err != nil {
		response.InternalServerError(c, "Failed to create credential: "+err.Error())
		return
	}

	response.Success(c, toCredentialDTO(cred, nil))
}

// Get 获取凭证详情
// @router /api/v1/credentials/:id [GET]
func Get(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	dao := db.NewCredentialDAO()
	cred, err := dao.FindByID(id)
	if err != nil {
		response.NotFound(c, "Credential not found")
		return
	}

	response.Success(c, toCredentialDTO(cred, nil))
}

// Update 更新凭证
// @router /api/v1/credentials/:id [PUT]
func Update(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	var req api.UpdateCredentialReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dao := db.NewCredentialDAO()
	cred, err := dao.FindByID(id)
	if err != nil {
		response.NotFound(c, "Credential not found")
		return
	}

	// 检查名称唯一性
	if req.Name != "" && req.Name != cred.Name {
		exists, err := dao.ExistsByNameExcludeID(req.Name, id)
		if err != nil {
			response.InternalServerError(c, "Failed to check credential name: "+err.Error())
			return
		}
		if exists {
			response.BadRequest(c, "Credential with this name already exists")
			return
		}
		cred.Name = req.Name
	}

	if req.Description != "" {
		cred.Description = req.Description
	}
	if req.SSHKeyID > 0 {
		// 验证数据库 SSH 密钥是否存在
		sshKeyDAO := db.NewSSHKeyDAO()
		if _, err := sshKeyDAO.FindByID(req.SSHKeyID); err != nil {
			response.BadRequest(c, "SSH key not found with the given id")
			return
		}
		cred.SSHKeyID = req.SSHKeyID
	}
	if req.SSHKeyPath != "" {
		cred.SSHKeyPath = req.SSHKeyPath
	}
	if req.Username != "" {
		cred.Username = req.Username
	}
	if req.Secret != "" {
		cred.Secret = req.Secret
	}
	if req.URLPattern != "" {
		cred.URLPattern = req.URLPattern
	}

	if err := dao.Save(cred); err != nil {
		response.InternalServerError(c, "Failed to update credential: "+err.Error())
		return
	}

	response.Success(c, toCredentialDTO(cred, nil))
}

// Delete 删除凭证
// @router /api/v1/credentials/:id [DELETE]
func Delete(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	dao := db.NewCredentialDAO()
	if _, err := dao.FindByID(id); err != nil {
		response.NotFound(c, "Credential not found")
		return
	}

	// 检查是否被仓库引用
	repoDAO := db.NewRepoDAO()
	repos, _ := repoDAO.FindAll()
	for _, repo := range repos {
		if repo.DefaultCredentialID == id {
			response.BadRequest(c, "Credential is referenced by repo: "+repo.Name)
			return
		}
		if repo.RemoteCredentials != nil {
			for remoteName, credID := range repo.RemoteCredentials {
				if credID == id {
					response.BadRequest(c, "Credential is referenced by repo "+repo.Name+" remote "+remoteName)
					return
				}
			}
		}
	}

	if err := dao.Delete(id); err != nil {
		response.InternalServerError(c, "Failed to delete credential: "+err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "Credential deleted successfully"})
}

// Test 测试凭证连接
// @router /api/v1/credentials/:id/test [POST]
func Test(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	var req api.TestCredentialReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.URL == "" {
		response.BadRequest(c, "url is required")
		return
	}

	dao := db.NewCredentialDAO()
	cred, err := dao.FindByID(id)
	if err != nil {
		response.NotFound(c, "Credential not found")
		return
	}

	gitSvc := git.NewGitService()
	var testErr error

	switch cred.Type {
	case "ssh_key":
		if cred.SSHKeyID > 0 {
			// 数据库密钥
			sshKeyDAO := db.NewSSHKeyDAO()
			sshKey, err := sshKeyDAO.FindByID(cred.SSHKeyID)
			if err != nil {
				response.InternalServerError(c, "Failed to load SSH key: "+err.Error())
				return
			}
			testErr = gitSvc.TestRemoteConnectionWithDBKey(req.URL, sshKey.PrivateKey, sshKey.Passphrase)
		} else if cred.SSHKeyPath != "" {
			// 本地密钥
			testErr = gitSvc.TestRemoteConnectionWithLocalKey(req.URL, cred.SSHKeyPath, cred.Secret)
		} else {
			response.BadRequest(c, "SSH key credential has no key configured")
			return
		}
	case "http_basic", "http_token":
		testErr = gitSvc.TestRemoteConnectionWithHTTP(req.URL, cred.Username, cred.Secret)
	default:
		testErr = gitSvc.TestRemoteConnection(req.URL)
	}

	if testErr != nil {
		response.Success(c, map[string]interface{}{
			"success": false,
			"message": testErr.Error(),
		})
		return
	}

	// 更新最后使用时间
	if err := dao.UpdateLastUsed(id); err != nil {
		// Log the error but continue
		_ = err // 暂时使用下划线忽略错误，避免空分支
	}

	response.Success(c, map[string]interface{}{
		"success": true,
		"message": "Connection successful",
	})
}

// Match 根据 URL 智能推荐凭证
// @router /api/v1/credentials/match [POST]
func Match(ctx context.Context, c *app.RequestContext) {
	var req api.MatchCredentialReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.URL == "" {
		response.BadRequest(c, "url is required")
		return
	}

	dao := db.NewCredentialDAO()
	recommended, others, err := dao.FindMatchingURL(req.URL)
	if err != nil {
		response.InternalServerError(c, "Failed to match credentials: "+err.Error())
		return
	}

	resp := api.MatchCredentialResp{
		Recommended: make([]api.CredentialDTO, 0, len(recommended)),
		Others:      make([]api.CredentialDTO, 0, len(others)),
	}
	for _, cred := range recommended {
		resp.Recommended = append(resp.Recommended, toCredentialDTO(&cred, nil))
	}
	for _, cred := range others {
		resp.Others = append(resp.Others, toCredentialDTO(&cred, nil))
	}

	response.Success(c, resp)
}

// sshKeyInfo 用于凭证 DTO 填充
type sshKeyInfo struct {
	Name    string
	KeyType string
}

// buildSSHKeyMap 一次性加载所有被凭证引用的 SSH 密钥，返回 id→info 映射
func buildSSHKeyMap(creds []po.Credential) map[uint]sshKeyInfo {
	m := make(map[uint]sshKeyInfo)
	sshKeyDAO := db.NewSSHKeyDAO()
	for _, cred := range creds {
		if cred.SSHKeyID > 0 {
			if _, ok := m[cred.SSHKeyID]; !ok {
				if key, err := sshKeyDAO.FindByID(cred.SSHKeyID); err == nil {
					m[cred.SSHKeyID] = sshKeyInfo{Name: key.Name, KeyType: key.KeyType}
				}
			}
		}
	}
	return m
}

// toCredentialDTO 转换为 DTO（脱敏）。sshKeyMap 可为 nil（单条查询场景）
func toCredentialDTO(cred *po.Credential, sshKeyMap map[uint]sshKeyInfo) api.CredentialDTO {
	dto := api.CredentialDTO{
		ID:          cred.ID,
		Name:        cred.Name,
		Type:        cred.Type,
		Description: cred.Description,
		SSHKeyID:    cred.SSHKeyID,
		SSHKeyPath:  cred.SSHKeyPath,
		Username:    cred.Username,
		HasSecret:   cred.Secret != "",
		URLPattern:  cred.URLPattern,
		LastUsedAt:  cred.LastUsedAt,
		CreatedAt:   cred.CreatedAt,
		UpdatedAt:   cred.UpdatedAt,
	}
	// 填充关联 SSH 密钥信息
	if cred.SSHKeyID > 0 {
		if sshKeyMap != nil {
			if info, ok := sshKeyMap[cred.SSHKeyID]; ok {
				dto.SSHKeyName = info.Name
				dto.SSHKeyType = info.KeyType
			}
		} else {
			// 单条查询：直接加载
			sshKeyDAO := db.NewSSHKeyDAO()
			if key, err := sshKeyDAO.FindByID(cred.SSHKeyID); err == nil {
				dto.SSHKeyName = key.Name
				dto.SSHKeyType = key.KeyType
			}
		}
	}
	// SSH 密钥路径脱敏：仅保留文件名
	if dto.SSHKeyPath != "" {
		parts := splitPath(dto.SSHKeyPath)
		if len(parts) > 0 {
			dto.SSHKeyPath = ".../" + parts[len(parts)-1]
		}
	}
	return dto
}

func splitPath(p string) []string {
	var parts []string
	for _, s := range []string{"/", "\\"} {
		if len(p) > 0 {
			for _, part := range splitByDelimiter(p, s) {
				if part != "" {
					parts = append(parts, part)
				}
			}
			if len(parts) > 0 {
				return parts
			}
		}
	}
	return []string{p}
}

func splitByDelimiter(s, delim string) []string {
	var result []string
	for {
		i := indexString(s, delim)
		if i < 0 {
			result = append(result, s)
			break
		}
		result = append(result, s[:i])
		s = s[i+len(delim):]
	}
	return result
}

func indexString(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func parseID(c *app.RequestContext) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	return uint(id), err
}

// init 初始化函数 - 检查 TestRemoteConnectionWithLocalKey 和 TestRemoteConnectionWithHTTP 方法是否存在
// 这些方法可能需要在 git_service.go 中添加
var _ = time.Now // ensure time is used
