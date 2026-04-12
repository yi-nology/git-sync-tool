package webhookevent

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/provider"
)

func List(eventType, source, status string, page, pageSize int) ([]api.WebhookEventDTO, int, error) {
	dao := db.NewWebhookEventDAO()
	events, total, err := dao.List(eventType, source, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]api.WebhookEventDTO, 0, len(events))
	for _, e := range events {
		dtos = append(dtos, toEventDTO(&e))
	}
	return dtos, int(total), nil
}

func Retry(eventID uint) error {
	dao := db.NewWebhookEventDAO()
	var event po.WebhookEvent
	if err := db.DB.First(&event, eventID).Error; err != nil {
		return fmt.Errorf("event not found: %w", err)
	}
	event.Status = "received"
	event.ErrorMessage = ""
	return dao.Save(&event)
}

func ProcessIncomingEvent(event *provider.NormalizedEvent, providerCfgID uint) error {
	dao := db.NewWebhookEventDAO()
	eventID := event.ID

	_, err := dao.FindByEventID(eventID)
	if err == nil {
		return nil
	}

	var repoID uint
	if event.Repo != nil {
		repoDAO := db.NewRepoDAO()
		repos, rErr := repoDAO.FindAll()
		if rErr == nil {
			for _, r := range repos {
				if r.PlatformOwner+"/"+r.PlatformRepo == event.Repo.FullName {
					repoID = r.ID
					break
				}
			}
		}
	}

	var crID uint
	var platformCRNum int
	if event.CR != nil {
		platformCRNum = event.CR.Number
		if repoID > 0 {
			crDAO := db.NewChangeRequestDAO()
			if localCR, err := crDAO.FindByRepoAndNumber(repoID, event.CR.Number); err == nil {
				crID = localCR.ID
			}
		}
	}

	actorName := ""
	actorUsername := ""
	if event.Actor != nil {
		actorName = event.Actor.Name
		actorUsername = event.Actor.Username
	}

	payload := map[string]interface{}{
		"type":   event.Type,
		"source": string(event.Source),
	}
	if event.Branch != "" {
		payload["branch"] = event.Branch
	}
	if event.Tag != "" {
		payload["tag"] = event.Tag
	}

	whEvent := &po.WebhookEvent{
		EventID:          eventID,
		ProviderConfigID: providerCfgID,
		EventType:        event.Type,
		Source:           string(event.Source),
		RepoID:           repoID,
		CRID:             crID,
		PlatformCRNumber: platformCRNum,
		ActorName:        actorName,
		ActorUsername:    actorUsername,
		Payload:          payload,
		Status:           "received",
	}
	if err := dao.Create(whEvent); err != nil {
		return err
	}

	go applyRules(whEvent)

	return nil
}

func applyRules(event *po.WebhookEvent) {
	ruleDAO := db.NewWebhookRuleDAO()
	rules, err := ruleDAO.FindByProviderConfigID(event.ProviderConfigID)
	if err != nil {
		return
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		if !matchPattern(rule.EventTypePattern, event.EventType) {
			continue
		}
		if rule.RepoPattern != "" && rule.RepoPattern != "*" {
			repoDAO := db.NewRepoDAO()
			if repo, err := repoDAO.FindByID(event.RepoID); err == nil {
				fullName := repo.PlatformOwner + "/" + repo.PlatformRepo
				if !matchPattern(rule.RepoPattern, fullName) {
					continue
				}
			}
		}

		switch rule.Action {
		case "sync":
			triggerSync(rule.ActionConfig)
		case "notify":
			log.Printf("Webhook rule %s: notify action triggered for event %s", rule.Name, event.EventID)
		}
	}

	now := time.Now()
	event.Status = "processed"
	event.ProcessedAt = &now
	eventDAO := db.NewWebhookEventDAO()
	eventDAO.Save(event)
}

func triggerSync(config map[string]interface{}) {
	taskKey, _ := config["sync_task_key"].(string)
	if taskKey == "" {
		return
	}
	log.Printf("Webhook rule triggered sync for task: %s", taskKey)
}

func matchPattern(pattern, value string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	if strings.Contains(pattern, "*") {
		regex := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), "\\*", ".*") + "$"
		matched, _ := regexp.MatchString(regex, value)
		return matched
	}
	return pattern == value
}

func toEventDTO(e *po.WebhookEvent) api.WebhookEventDTO {
	return api.WebhookEventDTO{
		ID:               e.ID,
		EventID:          e.EventID,
		EventType:        e.EventType,
		Source:           e.Source,
		RepoID:           e.RepoID,
		CRID:             e.CRID,
		PlatformCRNumber: e.PlatformCRNumber,
		ActorName:        e.ActorName,
		ActorUsername:    e.ActorUsername,
		Status:           e.Status,
		ProcessedAt:      e.ProcessedAt,
		CreatedAt:        e.CreatedAt,
		ErrorMessage:     e.ErrorMessage,
	}
}
