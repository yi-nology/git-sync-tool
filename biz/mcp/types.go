package mcp

import (
	"encoding/json"
)

type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
	Returns     json.RawMessage `json:"returns"`
}

type ToolRequest struct {
	Tool       string          `json:"tool"`
	Parameters json.RawMessage `json:"parameters"`
}

type ToolResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}
