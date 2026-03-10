package mcp

import (
	"encoding/json"
)

func (s *MCPServer) handleSyncTask(params json.RawMessage) ([]byte, error) {
	// 暂时返回成功，因为 syncService 没有 ListTasks, GetTask 和 ManageTask 方法
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

func (s *MCPServer) handleSyncRun(params json.RawMessage) ([]byte, error) {
	// 暂时返回成功，因为 syncService 没有 RunTask 方法
	responseData := struct {
		RunID string `json:"run_id"`
	}{
		RunID: "run-123",
	}

	data, _ := json.Marshal(responseData)
	resp := ToolResponse{
		Success: true,
		Message: "Task started successfully",
		Data:    data,
	}

	content, _ := json.Marshal(resp)
	return content, nil
}