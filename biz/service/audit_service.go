package service

import (
	"encoding/json"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
)

type AuditService struct{}

var AuditSvc *AuditService

func InitAuditService() {
	AuditSvc = &AuditService{}
}

// Log records an audit log entry
func (s *AuditService) Log(c *app.RequestContext, action, target string, details interface{}) {
	// Try to get IP and UA from context if available
	ip := ""
	ua := ""
	if c != nil {
		ip = c.ClientIP()
		ua = string(c.UserAgent())
	}

	detailsJSON, _ := json.Marshal(details)

	logEntry := model.AuditLog{
		Action:    action,
		Target:    target,
		Operator:  "system", // TODO: Replace with actual user when auth is implemented
		Details:   string(detailsJSON),
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	}

	// Run in background to not block main flow? 
	// Or sync to ensure audit? Usually async is better for performance unless strict audit required.
	// For now, sync is safer to ensure recording.
	go func() {
		dal.DB.Create(&logEntry)
	}()
}
