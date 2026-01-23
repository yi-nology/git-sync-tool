.PHONY: build build-http build-rpc run run-http run-rpc clean gen kitex-gen hz-gen test lint fmt help

# 默认目标
all: build

# 构建目标
build:
	@echo "Building git-manage-service..."
	go build -o output/git-manage-service main.go

build-http:
	@echo "Building HTTP-only service..."
	go build -o output/git-manage-service-http main.go

build-rpc:
	@echo "Building RPC-only service..."
	go build -o output/git-manage-service-rpc main.go

# 运行目标
run:
	go run main.go --mode=all

run-http:
	go run main.go --mode=http

run-rpc:
	go run main.go --mode=rpc

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

# 帮助
help:
	@echo "Git Manage Service Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        - Build the service"
	@echo "  make run          - Run in 'all' mode (HTTP + RPC)"
	@echo "  make run-http     - Run HTTP server only"
	@echo "  make run-rpc      - Run RPC server only"
	@echo "  make gen          - Generate all code (Kitex + Hz)"
	@echo "  make kitex-gen    - Generate Kitex RPC code"
	@echo "  make hz-gen       - Generate Hz HTTP code"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run linter"
	@echo "  make fmt          - Format code"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make help         - Show this help"
