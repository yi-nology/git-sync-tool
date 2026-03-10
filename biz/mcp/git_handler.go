package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/yi-nology/git-manage-service/biz/service/git"
)

type gitHandler struct {
	service *git.GitService
}

func newGitHandler() *gitHandler {
	return &gitHandler{
		service: git.NewGitService(),
	}
}

func (h *gitHandler) handleGitClone(params json.RawMessage) ([]byte, error) {
	var cloneParams struct {
		RemoteURL  string `json:"remote_url"`
		LocalPath  string `json:"local_path"`
		AuthType   string `json:"auth_type"`
		AuthKey    string `json:"auth_key"`
		AuthSecret string `json:"auth_secret"`
	}

	if err := json.Unmarshal(params, &cloneParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	err := h.service.Clone(
		cloneParams.RemoteURL,
		cloneParams.LocalPath,
		cloneParams.AuthType,
		cloneParams.AuthKey,
		cloneParams.AuthSecret,
	)

	if err != nil {
		resp := ToolResponse{
			Success: false,
			Message: fmt.Sprintf("Clone failed: %v", err),
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	resp := ToolResponse{
		Success: true,
		Message: "Repository cloned successfully",
	}
	content, _ := json.Marshal(resp)
	return content, nil
}

func (h *gitHandler) handleGitFetch(params json.RawMessage) ([]byte, error) {
	var fetchParams struct {
		Path   string `json:"path"`
		Remote string `json:"remote"`
	}

	if err := json.Unmarshal(params, &fetchParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	if fetchParams.Remote == "" {
		fetchParams.Remote = "origin"
	}

	err := h.service.Fetch(fetchParams.Path, fetchParams.Remote, nil)
	if err != nil {
		resp := ToolResponse{
			Success: false,
			Message: fmt.Sprintf("Fetch failed: %v", err),
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	resp := ToolResponse{
		Success: true,
		Message: "Fetch completed successfully",
	}
	content, _ := json.Marshal(resp)
	return content, nil
}

func (h *gitHandler) handleGitPush(params json.RawMessage) ([]byte, error) {
	var pushParams struct {
		Path         string   `json:"path"`
		TargetRemote string   `json:"target_remote"`
		SourceHash   string   `json:"source_hash"`
		TargetBranch string   `json:"target_branch"`
		Options      []string `json:"options"`
	}

	if err := json.Unmarshal(params, &pushParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	if pushParams.TargetRemote == "" {
		pushParams.TargetRemote = "origin"
	}

	err := h.service.Push(
		pushParams.Path,
		pushParams.TargetRemote,
		pushParams.SourceHash,
		pushParams.TargetBranch,
		pushParams.Options,
		nil,
	)

	if err != nil {
		resp := ToolResponse{
			Success: false,
			Message: fmt.Sprintf("Push failed: %v", err),
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	resp := ToolResponse{
		Success: true,
		Message: "Push completed successfully",
	}
	content, _ := json.Marshal(resp)
	return content, nil
}

func (h *gitHandler) handleGitCheckout(params json.RawMessage) ([]byte, error) {
	var checkoutParams struct {
		Path   string `json:"path"`
		Branch string `json:"branch"`
	}

	if err := json.Unmarshal(params, &checkoutParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	err := h.service.CheckoutBranch(checkoutParams.Path, checkoutParams.Branch)
	if err != nil {
		resp := ToolResponse{
			Success: false,
			Message: fmt.Sprintf("Checkout failed: %v", err),
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	resp := ToolResponse{
		Success: true,
		Message: "Branch checked out successfully",
	}
	content, _ := json.Marshal(resp)
	return content, nil
}

func (h *gitHandler) handleGitBranches(params json.RawMessage) ([]byte, error) {
	var branchesParams struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(params, &branchesParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	branches, err := h.service.GetBranches(branchesParams.Path)
	if err != nil {
		resp := ToolResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get branches: %v", err),
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	responseData := struct {
		Branches []string `json:"branches"`
	}{
		Branches: branches,
	}

	data, err := json.Marshal(responseData)
	if err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Failed to marshal response",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	resp := ToolResponse{
		Success: true,
		Message: "Branches retrieved successfully",
		Data:    data,
	}

	content, _ := json.Marshal(resp)
	return content, nil
}

func (h *gitHandler) handleGitAdd(params json.RawMessage) ([]byte, error) {
	var addParams struct {
		Path  string   `json:"path"`
		Files []string `json:"files"`
	}

	if err := json.Unmarshal(params, &addParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	var err error
	if len(addParams.Files) > 0 {
		err = h.service.AddFiles(addParams.Path, addParams.Files)
	} else {
		err = h.service.AddAll(addParams.Path)
	}

	if err != nil {
		resp := ToolResponse{
			Success: false,
			Message: fmt.Sprintf("Add failed: %v", err),
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	resp := ToolResponse{
		Success: true,
		Message: "Files added successfully",
	}
	content, _ := json.Marshal(resp)
	return content, nil
}

func (h *gitHandler) handleGitCommit(params json.RawMessage) ([]byte, error) {
	var commitParams struct {
		Path        string `json:"path"`
		Message     string `json:"message"`
		AuthorName  string `json:"author_name"`
		AuthorEmail string `json:"author_email"`
	}

	if err := json.Unmarshal(params, &commitParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	err := h.service.Commit(
		commitParams.Path,
		commitParams.Message,
		commitParams.AuthorName,
		commitParams.AuthorEmail,
	)

	if err != nil {
		resp := ToolResponse{
			Success: false,
			Message: fmt.Sprintf("Commit failed: %v", err),
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	resp := ToolResponse{
		Success: true,
		Message: "Changes committed successfully",
	}
	content, _ := json.Marshal(resp)
	return content, nil
}

func (h *gitHandler) handleGitStatus(params json.RawMessage) ([]byte, error) {
	var statusParams struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(params, &statusParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	status, err := h.service.GetStatus(statusParams.Path)
	if err != nil {
		resp := ToolResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get status: %v", err),
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	responseData := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	data, err := json.Marshal(responseData)
	if err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Failed to marshal response",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	resp := ToolResponse{
		Success: true,
		Message: "Status retrieved successfully",
		Data:    data,
	}

	content, _ := json.Marshal(resp)
	return content, nil
}

func (h *gitHandler) handleGitLog(params json.RawMessage) ([]byte, error) {
	var logParams struct {
		Path   string `json:"path"`
		Branch string `json:"branch"`
		Since  string `json:"since"`
		Until  string `json:"until"`
	}

	if err := json.Unmarshal(params, &logParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	if logParams.Branch == "" {
		logParams.Branch = "HEAD"
	}

	logs, err := h.service.GetCommits(logParams.Path, logParams.Branch, logParams.Since, logParams.Until)
	if err != nil {
		resp := ToolResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get logs: %v", err),
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	responseData := struct {
		Logs string `json:"logs"`
	}{
		Logs: logs,
	}

	data, err := json.Marshal(responseData)
	if err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Failed to marshal response",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	resp := ToolResponse{
		Success: true,
		Message: "Logs retrieved successfully",
		Data:    data,
	}

	content, _ := json.Marshal(resp)
	return content, nil
}

func (h *gitHandler) handleGitAuth(params json.RawMessage) ([]byte, error) {
	var authParams struct {
		AuthType   string `json:"auth_type"`
		AuthKey    string `json:"auth_key"`
		AuthSecret string `json:"auth_secret"`
	}

	if err := json.Unmarshal(params, &authParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	if authParams.AuthType != "ssh" && authParams.AuthType != "https" {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid authentication type",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	resp := ToolResponse{
		Success: true,
		Message: "Authentication setup successful",
	}
	content, _ := json.Marshal(resp)
	return content, nil
}
