package provider

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/provider"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
)

func List(ctx context.Context, c *app.RequestContext) {
	dao := db.NewProviderConfigDAO()
	configs, err := dao.FindAll()
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to fetch providers: "+err.Error())
		return
	}
	credDAO := db.NewCredentialDAO()
	result := make([]api.ProviderConfigDTO, 0, len(configs))
	for _, cfg := range configs {
		dto := toProviderConfigDTO(&cfg)
		if cfg.CredentialID > 0 {
			if cred, err := credDAO.FindByID(cfg.CredentialID); err == nil {
				dto.CredentialName = cred.Name
			}
		}
		result = append(result, dto)
	}
	pkgresponse.Success(c, result)
}

func Get(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		pkgresponse.BadRequest(c, "Invalid ID")
		return
	}
	dao := db.NewProviderConfigDAO()
	cfg, err := dao.FindByID(id)
	if err != nil {
		pkgresponse.NotFound(c, "Provider config not found")
		return
	}
	dto := toProviderConfigDTO(cfg)
	if cfg.CredentialID > 0 {
		credDAO := db.NewCredentialDAO()
		if cred, err := credDAO.FindByID(cfg.CredentialID); err == nil {
			dto.CredentialName = cred.Name
		}
	}
	pkgresponse.Success(c, dto)
}

func Create(ctx context.Context, c *app.RequestContext) {
	var req api.CreateProviderConfigReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.Name == "" || req.Platform == "" {
		pkgresponse.BadRequest(c, "name and platform are required")
		return
	}
	if req.Platform != "gitlab" && req.Platform != "github" && req.Platform != "gitea" {
		pkgresponse.BadRequest(c, "platform must be gitlab, github or gitea")
		return
	}
	if req.CredentialID == 0 {
		pkgresponse.BadRequest(c, "credential_id is required")
		return
	}
	credDAO := db.NewCredentialDAO()
	if _, err := credDAO.FindByID(req.CredentialID); err != nil {
		pkgresponse.BadRequest(c, "credential not found")
		return
	}
	dao := db.NewProviderConfigDAO()
	cfg := &po.ProviderConfig{
		Name: req.Name, Platform: req.Platform, BaseURL: req.BaseURL,
		CredentialID: req.CredentialID, WebhookSecret: req.WebhookSecret,
	}
	if err := dao.Create(cfg); err != nil {
		pkgresponse.InternalServerError(c, "Failed to create provider config: "+err.Error())
		return
	}
	pkgresponse.Success(c, toProviderConfigDTO(cfg))
}

func Update(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		pkgresponse.BadRequest(c, "Invalid ID")
		return
	}
	var req api.UpdateProviderConfigReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	dao := db.NewProviderConfigDAO()
	cfg, err := dao.FindByID(id)
	if err != nil {
		pkgresponse.NotFound(c, "Provider config not found")
		return
	}
	if req.Name != "" {
		cfg.Name = req.Name
	}
	if req.BaseURL != "" {
		cfg.BaseURL = req.BaseURL
	}
	if req.CredentialID > 0 {
		credDAO := db.NewCredentialDAO()
		if _, err := credDAO.FindByID(req.CredentialID); err != nil {
			pkgresponse.BadRequest(c, "credential not found")
			return
		}
		cfg.CredentialID = req.CredentialID
	}
	if req.WebhookSecret != "" {
		cfg.WebhookSecret = req.WebhookSecret
	}
	if err := dao.Save(cfg); err != nil {
		pkgresponse.InternalServerError(c, "Failed to update provider config: "+err.Error())
		return
	}
	provider.GetManager().Invalidate(id)
	pkgresponse.Success(c, toProviderConfigDTO(cfg))
}

func Delete(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		pkgresponse.BadRequest(c, "Invalid ID")
		return
	}
	dao := db.NewProviderConfigDAO()
	if _, err := dao.FindByID(id); err != nil {
		pkgresponse.NotFound(c, "Provider config not found")
		return
	}
	if err := dao.Delete(id); err != nil {
		pkgresponse.InternalServerError(c, "Failed to delete provider config: "+err.Error())
		return
	}
	provider.GetManager().Invalidate(id)
	pkgresponse.Success(c, map[string]string{"message": "Provider config deleted"})
}

func Test(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		pkgresponse.BadRequest(c, "Invalid ID")
		return
	}
	p, err := provider.GetManager().GetProvider(id)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to get provider: "+err.Error())
		return
	}
	result, err := p.TestConnection(ctx)
	if err != nil {
		pkgresponse.InternalServerError(c, "Test connection failed: "+err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func toProviderConfigDTO(cfg *po.ProviderConfig) api.ProviderConfigDTO {
	return api.ProviderConfigDTO{
		ID: cfg.ID, Name: cfg.Name, Platform: cfg.Platform,
		BaseURL: cfg.BaseURL, CredentialID: cfg.CredentialID,
		HasWebhookSecret: cfg.WebhookSecret != "",
		WebhookEndpoint:  cfg.WebhookEndpoint,
		CreatedAt:        cfg.CreatedAt, UpdatedAt: cfg.UpdatedAt,
	}
}

func parseID(c *app.RequestContext) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	return uint(id), err
}
