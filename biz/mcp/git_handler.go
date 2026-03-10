package mcp

import (
	"encoding/json"
	"fmt"
)

func (s *MCPServer) handleGitClone(params json.RawMessage) ([]byte, error) {
	var cloneParams struct {
		RemoteURL  string `json:"remote_url"`
		LocalPath  string `json:"local_path"`
		AuthType   string `json:"auth_type"`
		AuthKey    string `json:"auth_key"`
		AuthSecret string `json:"auth_secret"`
	}

	if err := json.Unmarshal(params, &cloneParams); err != nil {
		return s.errorResponse("Invalid parameters")
	}

	err := s.gitService.Clone(
		cloneParams.RemoteURL,
		cloneParams.LocalPath,
		cloneParams.AuthType,
		cloneParams.AuthKey,
		cloneParams.AuthSecret,
	)

	if err != nil {
		return s.errorResponse(fmt.Sprintf("Clone failed: %v", err))
	}

	return s.successResponse("Repository cloned successfully")
}

func (s *MCPServer) handleGitFetch(params json.RawMessage) ([]byte, error) {
	var fetchParams struct {
		Path   string `json:"path"`
		Remote string `json:"remote"`
	}

	if err := json.Unmarshal(params, &fetchParams); err != nil {
		return s.errorResponse("Invalid parameters")
	}

	if fetchParams.Remote == "" {
		fetchParams.Remote = "origin"
	}

	err := s.gitService.Fetch(fetchParams.Path, fetchParams.Remote, nil)
	if err != nil {
		return s.errorResponse(fmt.Sprintf("Fetch failed: %v", err))
	}

	return s.successResponse("Fetch completed successfully")
}

func (s *MCPServer) handleGitPush(params json.RawMessage) ([]byte, error) {
	var pushParams struct {
		Path         string   `json:"path"`
		TargetRemote string   `json:"target_remote"`
		SourceHash   string   `json:"source_hash"`
		TargetBranch string   `json:"target_branch"`
		Options      []string `json:"options"`
	}

	if err := json.Unmarshal(params, &pushParams); err != nil {
		return s.errorResponse("Invalid parameters")
	}

	if pushParams.TargetRemote == "" {
		pushParams.TargetRemote = "origin"
	}

	err := s.gitService.Push(
		pushParams.Path,
		pushParams.TargetRemote,
		pushParams.SourceHash,
		pushParams.TargetBranch,
		pushParams.Options,
		nil,
	)

	if err != nil {
		return s.errorResponse(fmt.Sprintf("Push failed: %v", err))
	}

	return s.successResponse("Push completed successfully")
}

func (s *MCPServer) handleGitCheckout(params json.RawMessage) ([]byte, error) {
	var checkoutParams struct {
		Path   string `json:"path"`
		Branch string `json:"branch"`
	}

	if err := json.Unmarshal(params, &checkoutParams); err != nil {
		return s.errorResponse("Invalid parameters")
	}

	err := s.gitService.CheckoutBranch(checkoutParams.Path, checkoutParams.Branch)
	if err != nil {
		return s.errorResponse(fmt.Sprintf("Checkout failed: %v", err))
	}

	return s.successResponse("Branch checked out successfully")
}

func (s *MCPServer) handleGitBranches(params json.RawMessage) ([]byte, error) {
	var branchesParams struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(params, &branchesParams); err != nil {
		return s.errorResponse("Invalid parameters")
	}

	branches, err := s.gitService.GetBranches(branchesParams.Path)
	if err != nil {
		return s.errorResponse(fmt.Sprintf("Failed to get branches: %v", err))
	}

	responseData := struct {
		Branches []string `json:"branches"`
	}{
		Branches: branches,
	}

	data, err := json.Marshal(responseData)
	if err != nil {
		return s.errorResponse("Failed to marshal response")
	}

	resp := ToolResponse{
		Success: true,
		Message: "Branches retrieved successfully",
		Data:    data,
	}

	content, _ := json.Marshal(resp)
	return content, nil
}

func (s *MCPServer) handleGitAdd(params json.RawMessage) ([]byte, error) {
	var addParams struct {
		Path  string   `json:"path"`
		Files []string `json:"files"`
	}

	if err := json.Unmarshal(params, &addParams); err != nil {
		return s.errorResponse("Invalid parameters")
	}

	var err error
	if len(addParams.Files) > 0 {
		err = s.gitService.AddFiles(addParams.Path, addParams.Files)
	} else {
		err = s.gitService.AddAll(addParams.Path)
	}

	if err != nil {
		return s.errorResponse(fmt.Sprintf("Add failed: %v", err))
	}

	return s.successResponse("Files added successfully")
}

func (s *MCPServer) handleGitCommit(params json.RawMessage) ([]byte, error) {
	var commitParams struct {
		Path        string `json:"path"`
		Message     string `json:"message"`
		AuthorName  string `json:"author_name"`
		AuthorEmail string `json:"author_email"`
	}

	if err := json.Unmarshal(params, &commitParams); err != nil {
		return s.errorResponse("Invalid parameters")
	}

	err := s.gitService.Commit(
		commitParams.Path,
		commitParams.Message,
		commitParams.AuthorName,
		commitParams.AuthorEmail,
	)

	if err != nil {
		return s.errorResponse(fmt.Sprintf("Commit failed: %v", err))
	}

	return s.successResponse("Changes committed successfully")
}

func (s *MCPServer) handleGitStatus(params json.RawMessage) ([]byte, error) {
	var statusParams struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(params, &statusParams); err != nil {
		return s.errorResponse("Invalid parameters")
	}

	status, err := s.gitService.GetStatus(statusParams.Path)
	if err != nil {
		return s.errorResponse(fmt.Sprintf("Failed to get status: %v", err))
	}

	responseData := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	data, err := json.Marshal(responseData)
	if err != nil {
		return s.errorResponse("Failed to marshal response")
	}

	resp := ToolResponse{
		Success: true,
		Message: "Status retrieved successfully",
		Data:    data,
	}

	content, _ := json.Marshal(resp)
	return content, nil
}

func (s *MCPServer) handleGitLog(params json.RawMessage) ([]byte, error) {
	var logParams struct {
		Path   string `json:"path"`
		Branch string `json:"branch"`
		Since  string `json:"since"`
		Until  string `json:"until"`
	}

	if err := json.Unmarshal(params, &logParams); err != nil {
		return s.errorResponse("Invalid parameters")
	}

	if logParams.Branch == "" {
		logParams.Branch = "HEAD"
	}

	logs, err := s.gitService.GetCommits(logParams.Path, logParams.Branch, logParams.Since, logParams.Until)
	if err != nil {
		return s.errorResponse(fmt.Sprintf("Failed to get logs: %v", err))
	}

	responseData := struct {
		Logs string `json:"logs"`
	}{
		Logs: logs,
	}

	data, err := json.Marshal(responseData)
	if err != nil {
		return s.errorResponse("Failed to marshal response")
	}

	resp := ToolResponse{
		Success: true,
		Message: "Logs retrieved successfully",
		Data:    data,
	}

	content, _ := json.Marshal(resp)
	return content, nil
}

func (s *MCPServer) handleGitAuth(params json.RawMessage) ([]byte, error) {
	var authParams struct {
		AuthType   string `json:"auth_type"`
		AuthKey    string `json:"auth_key"`
		AuthSecret string `json:"auth_secret"`
	}

	if err := json.Unmarshal(params, &authParams); err != nil {
		return s.errorResponse("Invalid parameters")
	}

	// 这里只是验证认证信息格式，实际认证会在具体操作中使用
	if authParams.AuthType != "ssh" && authParams.AuthType != "https" {
		return s.errorResponse("Invalid authentication type")
	}

	return s.successResponse("Authentication setup successful")
}
