package handler

import (
	"context"
	"github.com/yi-nology/git-sync-tool/biz/service"
	"log"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

type WebhookRequest struct {
	TaskID uint `json:"task_id"`
}

// @Summary Trigger sync via Webhook
// @Tags Webhook
// @Param X-Hub-Signature-256 header string true "HMAC-SHA256 signature"
// @Param request body WebhookRequest true "Webhook payload"
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/webhooks/task-sync [post]
func HandleWebhookTrigger(ctx context.Context, c *app.RequestContext) {
	var req WebhookRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if req.TaskID == 0 {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "task_id is required"})
		return
	}

	log.Printf("Received valid webhook trigger for task %d", req.TaskID)

	// Trigger sync asynchronously
	go func(taskID uint) {
		svc := service.NewSyncService()
		if err := svc.RunTask(taskID); err != nil {
			log.Printf("Webhook triggered task %d failed: %v", taskID, err)
		} else {
			log.Printf("Webhook triggered task %d success", taskID)
		}
	}(req.TaskID)

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Sync triggered successfully",
		"task_id": req.TaskID,
	})
}
