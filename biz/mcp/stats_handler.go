package mcp

import (
	"encoding/json"
)

type statsHandler struct {}

func newStatsHandler() *statsHandler {
	return &statsHandler{}
}

func (h *statsHandler) handleStatsCode(params json.RawMessage) ([]byte, error) {
	var codeParams struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(params, &codeParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	// 这里应该调用统计服务的代码统计方法
	// 暂时返回成功
	responseData := struct {
		Lines      int    `json:"lines"`
		Files      int    `json:"files"`
		Languages  []string `json:"languages"`
	}{
		Lines:      1000,
		Files:      50,
		Languages:  []string{"Go", "JavaScript", "HTML"},
	}

	data, _ := json.Marshal(responseData)
	resp := ToolResponse{
		Success: true,
		Message: "Code statistics retrieved successfully",
		Data:    data,
	}
	content, _ := json.Marshal(resp)
	return content, nil
}

func (h *statsHandler) handleStatsLanguage(params json.RawMessage) ([]byte, error) {
	var langParams struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(params, &langParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	// 这里应该调用统计服务的语言统计方法
	// 暂时返回成功
	responseData := struct {
		Languages map[string]float64 `json:"languages"`
	}{
		Languages: map[string]float64{
			"Go": 60.0,
			"JavaScript": 30.0,
			"HTML": 10.0,
		},
	}

	data, _ := json.Marshal(responseData)
	resp := ToolResponse{
		Success: true,
		Message: "Language statistics retrieved successfully",
		Data:    data,
	}
	content, _ := json.Marshal(resp)
	return content, nil
}