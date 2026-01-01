package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
	"github.com/yi-nology/git-manage-service/biz/pkg/response"
	"github.com/yi-nology/git-manage-service/biz/service"
)

// @Summary Trigger a sync task via Webhook
// @Tags Webhook
// @Param token query string true "Webhook Token"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /api/webhooks/trigger [post]
func HandleWebhookTrigger(ctx context.Context, c *app.RequestContext) {
	token := c.Query("token")
	if token == "" {
		response.BadRequest(c, "missing token")
		return
	}

	var task model.SyncTask
	if err := dal.DB.Preload("SourceRepo").Preload("TargetRepo").Where("webhook_token = ?", token).First(&task).Error; err != nil {
		response.NotFound(c, "invalid token or task not found")
		return
	}

	if !task.Enabled {
		response.Error(c, consts.StatusForbidden, 403, "task is disabled")
		return
	}

	// Run Async
	go func() {
		svc := service.NewSyncService()
		svc.ExecuteSync(&task)
	}()

	response.Success(c, map[string]string{
		"status":   "triggered",
		"task_key": task.Key,
		"message":  "Sync task triggered successfully",
	})
}
