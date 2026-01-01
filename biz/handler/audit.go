package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
	"github.com/yi-nology/git-manage-service/biz/pkg/response"
)

// @Summary List audit logs
// @Tags Audit
// @Produce json
// @Success 200 {object} response.Response{data=[]model.AuditLog}
// @Router /api/audit/logs [get]
func ListAuditLogs(ctx context.Context, c *app.RequestContext) {
	var logs []model.AuditLog
	dal.DB.Order("created_at desc").Limit(100).Find(&logs)
	response.Success(c, logs)
}
