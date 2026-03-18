package audit

import (
	"encoding/json"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type AuditService struct {
	auditDAO *db.AuditLogDAO
}

var AuditSvc *AuditService

func InitAuditService() {
	AuditSvc = &AuditService{
		auditDAO: db.NewAuditLogDAO(),
	}
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

	logEntry := po.AuditLog{
		Action:    action,
		Target:    target,
		Operator:  "system", // TODO: Replace with actual user when auth is implemented
		Details:   string(detailsJSON),
		IPAddress: ip,
		UserAgent: ua,
	}

	// Run in background to not block main flow?
	// Or sync to ensure audit? Usually async is better for performance unless strict audit required.
	// For now, sync is safer to ensure recording.
	go func() {
		if err := s.auditDAO.Create(&logEntry); err != nil {
			// Log the error but don't block the main flow
			// TODO: Add proper logging here
			_ = err // 暂时使用下划线忽略错误，避免空分支
		}
	}()
}
