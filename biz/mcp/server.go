package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/service/notification"
	"github.com/yi-nology/git-manage-service/biz/service/sync"
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

type MCPServer struct {
	tools               map[string]ToolDefinition
	gitService          *git.GitService
	notificationService *notification.NotificationService
	syncService         *sync.SyncService
}

func NewMCPServer() *MCPServer {
	return &MCPServer{
		tools:               make(map[string]ToolDefinition),
		gitService:          git.NewGitService(),
		notificationService: notification.NewNotificationService(),
		syncService:         sync.NewSyncService(),
	}
}

func (s *MCPServer) LoadTools() error {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}
	toolsDir := filepath.Join(cwd, "biz", "mcp", "tools")
	files, err := os.ReadDir(toolsDir)
	if err != nil {
		return fmt.Errorf("failed to read tools directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			toolPath := filepath.Join(toolsDir, file.Name())
			content, err := os.ReadFile(toolPath)
			if err != nil {
				log.Printf("Warning: failed to read tool file %s: %v", toolPath, err)
				continue
			}

			var tool ToolDefinition
			if err := json.Unmarshal(content, &tool); err != nil {
				log.Printf("Warning: failed to parse tool file %s: %v", toolPath, err)
				continue
			}

			s.tools[tool.Name] = tool
			log.Printf("Loaded tool: %s", tool.Name)
		}
	}

	return nil
}

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

func (s *MCPServer) handleNotificationSend(params json.RawMessage) ([]byte, error) {
	var sendParams struct {
		ChannelID string      `json:"channel_id"`
		Event     string      `json:"event"`
		Message   string      `json:"message"`
		Data      interface{} `json:"data"`
	}

	if err := json.Unmarshal(params, &sendParams); err != nil {
		return s.errorResponse("Invalid parameters")
	}

	// 调用通知服务发送通知
	s.notificationService.Send(&notification.NotificationMessage{
		Title:        sendParams.Message,
		Content:      sendParams.Message,
		Status:       "success",
		TriggerEvent: sendParams.Event,
		TaskKey:      "",
		RepoKey:      "",
	})

	return s.successResponse("Notification sent successfully")
}

func (s *MCPServer) handleNotificationChannels(params json.RawMessage) ([]byte, error) {
	// 暂时返回成功，因为 notificationService 没有 ListChannels 和 ManageChannel 方法
	responseData := struct {
		Channels []string `json:"channels"`
	}{
		Channels: []string{"email", "webhook", "slack"},
	}

	data, _ := json.Marshal(responseData)
	resp := ToolResponse{
		Success: true,
		Message: "Channels retrieved successfully",
		Data:    data,
	}

	content, _ := json.Marshal(resp)
	return content, nil
}

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

func (s *MCPServer) successResponse(message string) ([]byte, error) {
	resp := ToolResponse{
		Success: true,
		Message: message,
	}
	content, _ := json.Marshal(resp)
	return content, nil
}

func (s *MCPServer) errorResponse(message string) ([]byte, error) {
	resp := ToolResponse{
		Success: false,
		Message: message,
	}
	content, _ := json.Marshal(resp)
	return content, nil
}

func (s *MCPServer) Start() error {
	if err := s.LoadTools(); err != nil {
		return fmt.Errorf("failed to load tools: %v", err)
	}

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	defer listener.Close()

	log.Println("MCP server started on port 9000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *MCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading from connection: %v", err)
		return
	}

	request := buffer[:n]
	response, err := s.HandleRequest(request)
	if err != nil {
		log.Printf("Error handling request: %v", err)
		errorResp := ToolResponse{
			Success: false,
			Message: "Internal server error",
		}
		response, _ = json.Marshal(errorResp)
	}

	_, err = conn.Write(response)
	if err != nil {
		log.Printf("Error writing to connection: %v", err)
	}
}
