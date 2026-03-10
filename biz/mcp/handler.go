package mcp

import (
	"encoding/json"
)

func (s *MCPServer) HandleRequest(request []byte) ([]byte, error) {
	var toolReq ToolRequest
	if err := json.Unmarshal(request, &toolReq); err != nil {
		return s.errorResponse("Invalid request format")
	}

	tool, exists := s.tools[toolReq.Tool]
	if !exists {
		return s.errorResponse("Tool not found")
	}

	switch tool.Name {
	case "git_clone":
		return s.handleGitClone(toolReq.Parameters)
	case "git_fetch":
		return s.handleGitFetch(toolReq.Parameters)
	case "git_push":
		return s.handleGitPush(toolReq.Parameters)
	case "git_checkout":
		return s.handleGitCheckout(toolReq.Parameters)
	case "git_branches":
		return s.handleGitBranches(toolReq.Parameters)
	case "git_add":
		return s.handleGitAdd(toolReq.Parameters)
	case "git_commit":
		return s.handleGitCommit(toolReq.Parameters)
	case "git_status":
		return s.handleGitStatus(toolReq.Parameters)
	case "git_log":
		return s.handleGitLog(toolReq.Parameters)
	case "git_auth":
		return s.handleGitAuth(toolReq.Parameters)
	case "notification_send":
		return s.handleNotificationSend(toolReq.Parameters)
	case "notification_channels":
		return s.handleNotificationChannels(toolReq.Parameters)
	case "sync_task":
		return s.handleSyncTask(toolReq.Parameters)
	case "sync_run":
		return s.handleSyncRun(toolReq.Parameters)
	default:
		return s.errorResponse("Tool not implemented")
	}
}
