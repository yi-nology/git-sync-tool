package webhook_event

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/webhookevent"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
)

func List(ctx context.Context, c *app.RequestContext) {
	var req api.ListWebhookEventsReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	events, total, err := webhookevent.List(req.EventType, req.Source, req.Status, req.Page, req.PageSize)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to list webhook events: "+err.Error())
		return
	}
	pkgresponse.Success(c, map[string]interface{}{
		"items": events,
		"total": total,
	})
}

func Retry(ctx context.Context, c *app.RequestContext) {
	var req struct {
		EventID uint `json:"event_id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.EventID == 0 {
		pkgresponse.BadRequest(c, "event_id is required")
		return
	}
	if err := webhookevent.Retry(req.EventID); err != nil {
		pkgresponse.InternalServerError(c, "Failed to retry event: "+err.Error())
		return
	}
	pkgresponse.Success(c, map[string]string{"message": "Event retried"})
}
