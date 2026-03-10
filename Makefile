.PHONY: build build-http build-rpc build-all build-full build-frontend build-frontend-integrate run run-http run-rpc run-frontend preview-frontend clean clean-frontend gen kitex-gen hz-gen test lint fmt help desktop desktop-darwin desktop-windows desktop-linux desktop-all

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 编译参数
# 重要: SQLite 驱动需要 CGO, 因此必须设置 CGO_ENABLED=1
export CGO_ENABLED=1
LDFLAGS := -X 'github.com/yi-nology/git-manage-service/pkg/appinfo.Version=$(VERSION)' \
           -X 'github.com/yi-nology/git-manage-service/pkg/appinfo.BuildTime=$(BUILD_TIME)' \
           -X 'github.com/yi-nology/git-manage-service/pkg/appinfo.GitCommit=$(GIT_COMMIT)'

# 默认目标
all: build-full

# 一键构建完整服务（前端+后端）
build-full:
	@echo "========================================"
	@echo "Building Full-Stack Service..."
	@echo "========================================"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo ""
	@echo "[1/2] Building Frontend..."
	@if [ ! -d "frontend/node_modules" ]; then \
		echo "Installing frontend dependencies..."; \
		cd frontend && npm install; \
	fi
	@cd frontend && npm run build
	@echo "Copying frontend assets to public directory..."
	@rm -rf public
	@cp -r frontend/dist public
	@echo "✓ Frontend build complete"
	@echo ""
	@echo "[2/2] Building Backend..."
	@mkdir -p output
	@go build -ldflags "$(LDFLAGS)" -o output/git-manage-service ./cmd/server
	@echo "✓ Backend build complete"
	@echo ""
	@echo "========================================"
	@echo "✓ Full-Stack Build Complete!"
	@echo "========================================"
	@echo ""
	@echo "Run the service:"
	@echo "  ./output/git-manage-service --mode=all"
	@echo ""
	@echo "Or use:"
	@echo "  make run"
	@echo ""

# 构建目标（注入版本信息）
build:
	@echo "Building git-manage-service..."
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@mkdir -p output
	go build -ldflags "$(LDFLAGS)" -o output/git-manage-service ./cmd/server

build-http:
	@echo "Building HTTP-only service..."
	@mkdir -p output
	go build -ldflags "$(LDFLAGS)" -o output/git-manage-service-http ./cmd/server

build-rpc:
	@echo "Building RPC-only service..."
	@mkdir -p output
	go build -ldflags "$(LDFLAGS)" -o output/git-manage-service-rpc ./cmd/server

# 多平台构建
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p output
	@for OS in linux darwin windows; do \
		for ARCH in amd64 arm64; do \
			EXT=""; \
			if [ "$$OS" = "windows" ]; then EXT=".exe"; fi; \
			echo "Building $$OS/$$ARCH..."; \
			GOOS=$$OS GOARCH=$$ARCH go build -ldflags "$(LDFLAGS)" \
				-o output/git-manage-service-$$OS-$$ARCH$$EXT ./cmd/server; \
			done; \
		done
	@echo "Build complete. Binaries in output/"

# 前端构建
build-frontend:
	@echo "Building frontend..."
	@if [ ! -d "frontend/node_modules" ]; then \
		echo "Installing frontend dependencies..."; \
		cd frontend && npm install; \
	fi
	@cd frontend && npm run build
	@echo "Frontend build complete. Output in frontend/dist/"

# 前端构建并集成到后端
build-frontend-integrate:
	@echo "Building and integrating frontend..."
	@if [ ! -d "frontend/node_modules" ]; then \
		echo "Installing frontend dependencies..."; \
		cd frontend && npm install; \
	fi
	@cd frontend && npm run build
	@echo "Copying frontend assets to public directory..."
	@rm -rf public
	@cp -r frontend/dist public
	@echo "Frontend integrated successfully. Backend can now serve frontend from ./public/"

# 前端开发服务器
run-frontend:
	@echo "Starting frontend dev server..."
	@cd frontend && npm run dev

# 前端预览
preview-frontend:
	@echo "Starting frontend preview server..."
	@cd frontend && npm run preview

# 运行目标
run:
	go run ./cmd/server --mode=all

run-http:
	go run ./cmd/server --mode=http

run-rpc:
	go run ./cmd/server --mode=rpc

# 代码生成
gen:
	@chmod +x script/gen.sh
	./script/gen.sh

kitex-gen:
	cd biz && kitex -module github.com/yi-nology/git-manage-service/biz \
		-service git_service -I ../idl ../idl/git.proto

hz-gen:
	@if ls idl/biz/*.proto 1> /dev/null 2>&1; then \
		for proto in idl/biz/*.proto; do \
			echo "Processing $$proto..."; \
			hz update -idl "$$proto" \
				--handler_dir biz/handler/hz \
				--router_dir biz/router/hz \
				--model_dir biz/model/hz; \
		done; \
	else \
		echo "No proto files found in idl/biz/"; \
	fi

# 测试和代码质量
test:
	go test -v ./...

lint:
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Skipping lint."; \
	fi

fmt:
	go fmt ./...

# 清理
clean:
	rm -rf output

clean-frontend:
	@echo "Cleaning frontend build artifacts..."
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	rm -rf public

# 帮助
help:
	@echo "Git Manage Service Makefile"
	@echo ""
	@echo "Quick Start:"
	@echo "  make build-full   - 🚀 Build complete service (frontend + backend) - RECOMMENDED"
	@echo "  make              - Same as 'make build-full'"
	@echo "  make run          - Run the built service"
	@echo ""
	@echo "Backend Build:"
	@echo "  make build        - Build the service (with version info)"
	@echo "  make build-all    - Build for multiple platforms (linux/darwin/windows, amd64/arm64)"
	@echo "  make build-http   - Build HTTP-only service"
	@echo "  make build-rpc    - Build RPC-only service"
	@echo ""
	@echo "Frontend Build:"
	@echo "  make build-frontend   - Build frontend (production)"
	@echo "  make build-frontend-integrate - Build frontend and copy to public/ for backend"
	@echo "  make run-frontend     - Start frontend dev server"
	@echo "  make preview-frontend - Preview frontend build"
	@echo ""
	@echo "Run Services:"
	@echo "  make run          - Run in 'all' mode (HTTP + RPC)"
	@echo "  make run-http     - Run HTTP server only"
	@echo "  make run-rpc      - Run RPC server only"
	@echo ""
	@echo "Code Generation:"
	@echo "  make gen          - Generate all code (Kitex + Hz)"
	@echo "  make kitex-gen    - Generate Kitex RPC code"
	@echo "  make hz-gen       - Generate Hz HTTP code"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run linter"
	@echo "  make fmt          - Format code"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean        - Clean backend build artifacts"
	@echo "  make clean-frontend - Clean frontend build artifacts"
	@echo ""
	@echo "Desktop Application (Wails):"
	@echo "  make desktop      - 🖥️  Build desktop app for current platform"
	@echo "  make desktop-dev  - 🔧 Run desktop app in development mode (debug)"
	@echo "  make desktop-darwin  - Build macOS app (universal)"
	@echo "  make desktop-windows - Build Windows exe + installer"
	@echo "  make desktop-linux   - Build Linux deb/rpm/AppImage"
	@echo "  make desktop-all     - Build for all platforms"
	@echo ""
	@echo "Other:"
	@echo "  make help         - Show this help"
	@echo ""
	@echo "Version Info:"
	@echo "  VERSION=$(VERSION)"
	@echo "  BUILD_TIME=$(BUILD_TIME)"
	@echo "  GIT_COMMIT=$(GIT_COMMIT)"

# ========================================
# Desktop Application Build (Wails)
# ========================================

desktop:
	@echo "======================================="
	@echo "Building Desktop Application..."
	@echo "======================================="
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo ""
	@echo "Checking Wails installation..."
	@if ! command -v wails &> /dev/null; then \
		echo "❌ Wails not found. Please run:"; \
		echo "   make setup-desktop"; \
		exit 1; \
	fi
	@echo "✓ Wails installed"
	@echo ""
	@echo "Building desktop application..."
	@wails build -clean
	@echo ""
	@echo "======================================="
	@echo "✓ Desktop Build Complete!"
	@echo "======================================="
	@echo ""
	@echo "Check the build/bin directory for output"
	@ls -lh build/bin/ 2>/dev/null || echo "Build output not found"

# 本地调试 Wails 应用
desktop-dev:
	@echo "======================================="
	@echo "Running Desktop Application in Development Mode..."
	@echo "======================================="
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo ""
	@echo "Checking Wails installation..."
	@if ! command -v wails &> /dev/null; then \
		echo "❌ Wails not found. Please run:"; \
		echo "   make setup-desktop"; \
		exit 1; \
	fi
	@echo "✓ Wails installed"
	@echo ""
	@echo "Starting development mode..."
	@wails dev
	@echo ""
	@echo "======================================="
	@echo "Development mode exited"
	@echo "======================================="

setup-desktop:
	@echo "Setting up desktop build environment..."
	@./script/setup-desktop.sh

desktop-darwin:
	@echo "Building macOS application..."
	@if ! command -v wails &> /dev/null; then \
		go install github.com/wailsapp/wails/v2/cmd/wails@latest; \
	fi
	@wails build -platform darwin/universal -clean
	@echo "✓ macOS build complete: build/bin/"

desktop-windows:
	@echo "Building Windows application..."
	@if ! command -v wails &> /dev/null; then \
		go install github.com/wailsapp/wails/v2/cmd/wails@latest; \
	fi
	@wails build -platform windows/amd64 -clean -nsis
	@echo "✓ Windows build complete: build/bin/"

desktop-linux:
	@echo "Building Linux applications..."
	@if ! command -v wails &> /dev/null; then \
		go install github.com/wailsapp/wails/v2/cmd/wails@latest; \
	fi
	@echo "Building DEB package..."
	@wails build -platform linux/amd64 -clean -deb
	@echo "Building RPM package..."
	@wails build -platform linux/amd64 -clean -rpm
	@echo "Building AppImage..."
	@wails build -platform linux/amd64 -clean -appimage
	@echo "✓ Linux builds complete: build/bin/"

desktop-all:
	@echo "========================================"
	@echo "Building All Desktop Platforms..."
	@echo "========================================"
	@$(MAKE) desktop-darwin
	@$(MAKE) desktop-windows
	@$(MAKE) desktop-linux
	@echo ""
	@echo "========================================"
	@echo "✓ All Desktop Builds Complete!"
	@echo "========================================"
	@echo ""
	@echo "Available builds:"
	@ls -lh build/bin/
