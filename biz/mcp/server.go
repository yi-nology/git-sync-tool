package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/service/notification"
	syncservice "github.com/yi-nology/git-manage-service/biz/service/sync"
)

type MCPServer struct {
	tools               map[string]ToolDefinition
	gitService          *git.GitService
	notificationService *notification.NotificationService
	syncService         *syncservice.SyncService
	auditHandler        *auditHandler
	statsHandler        *statsHandler
	storageHandler      *storageHandler
	listener            net.Listener
	isRunning           bool
	mu                  sync.Mutex
}

func NewMCPServer() *MCPServer {
	return &MCPServer{
		tools:               make(map[string]ToolDefinition),
		gitService:          git.NewGitService(),
		notificationService: notification.NewNotificationService(),
		syncService:         syncservice.NewSyncService(),
		auditHandler:        newAuditHandler(),
		statsHandler:        newStatsHandler(),
		storageHandler:      newStorageHandler(),
		isRunning:           false,
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

func (s *MCPServer) Start() error {
	if err := s.LoadTools(); err != nil {
		return fmt.Errorf("failed to load tools: %v", err)
	}

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}

	s.mu.Lock()
	s.listener = listener
	s.isRunning = true
	s.mu.Unlock()

	log.Println("MCP server started on port 9000")

	for {
		s.mu.Lock()
		running := s.isRunning
		s.mu.Unlock()

		if !running {
			break
		}

		// 接受连接
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handleConnection(conn)
	}

	return nil
}

func (s *MCPServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	s.isRunning = false

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %v", err)
		}
		log.Println("MCP server listener closed")
	}

	log.Println("MCP server stopped")
	return nil
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
