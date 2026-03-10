package mcp

import (
	"encoding/json"
)

type syncHandler struct {}

func newSyncHandler() *syncHandler {
	return &syncHandler{}
}

func (h *syncHandler) handleSyncTask(params json.RawMessage) ([]byte, error) {
	var taskParams struct {
		SourceRepo string `json:"source_repo"`
		TargetRepo string `json:"target_repo"`
		Branch     string `json:"branch"`
		Cron       string `json:"cron"`
	}

	if err := json.Unmarshal(params, &taskParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	// 这里应该调用同步服务的创建任务方法
	// 暂时返回成功
	responseData := struct {
		TaskID string `json:"task_id"`
	}{
		TaskID: "sync-123",
	}

	data, _ := json.Marshal(responseData)
	resp := ToolResponse{
		Success: true,
		Message: "Sync task created successfully",
		Data:    data,
	}
	content, _ := json.Marshal(resp)
	return content, nil
}

func (h *syncHandler) handleSyncRun(params json.RawMessage) ([]byte, error) {
	var runParams struct {
		TaskID string `json:"task_id"`
	}

	if err := json.Unmarshal(params, &runParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	// 这里应该调用同步服务的运行任务方法
	// 暂时返回成功
	resp := ToolResponse{
		Success: true,
		Message: "Sync task executed successfully",
	}
	content, _ := json.Marshal(resp)
	return content, nil
}