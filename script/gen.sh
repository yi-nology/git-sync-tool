#!/bin/bash
# script/gen.sh - 代码生成脚本
# 生成 Kitex RPC 代码和 Hz HTTP 代码

set -e

PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$PROJECT_ROOT"

echo "=========================================="
echo "  Git Manage Service - Code Generator"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查工具是否安装
check_tool() {
    if ! command -v "$1" &> /dev/null; then
        error "$1 is not installed. Please install it first."
        return 1
    fi
    return 0
}

# 1. 检查依赖工具
info "Checking required tools..."
MISSING_TOOLS=()

if ! check_tool "protoc"; then
    MISSING_TOOLS+=("protoc")
fi

if ! check_tool "kitex"; then
    warn "kitex not found. Skipping Kitex code generation."
    warn "Install with: go install github.com/cloudwego/kitex/tool/cmd/kitex@latest"
    SKIP_KITEX=true
fi

if ! check_tool "hz"; then
    warn "hz not found. Skipping Hz code generation."
    warn "Install with: go install github.com/cloudwego/hertz/cmd/hz@latest"
    SKIP_HZ=true
fi

if [ ${#MISSING_TOOLS[@]} -ne 0 ]; then
    error "Missing required tools: ${MISSING_TOOLS[*]}"
    exit 1
fi

# 2. 生成 Kitex RPC 代码
if [ "$SKIP_KITEX" != "true" ]; then
    info "Generating Kitex RPC code..."
    cd biz
    kitex -module github.com/yi-nology/git-manage-service/biz \
          -service git_service \
          -I ../idl \
          ../idl/git.proto
    cd ..
    info "Kitex code generation completed"
else
    warn "Skipped Kitex code generation"
fi

# 3. 生成 Hz HTTP 代码（如果有 biz proto 文件）
if [ "$SKIP_HZ" != "true" ]; then
    if ls idl/biz/*.proto 1> /dev/null 2>&1; then
        info "Generating Hz HTTP code..."
        
        # 检查是否已初始化 Hz（通过检查 .hz 文件或 router/hz 目录）
        if [ ! -d "biz/router/hz" ]; then
            info "Initializing Hz project..."
            hz new -idl idl/biz/repo.proto \
                -I idl \
                -module github.com/yi-nology/git-manage-service \
                --handler_dir biz/handler/hz \
                --router_dir biz/router/hz \
                --model_dir biz/model/hz
        fi
        
        # 更新生成代码
        for proto in idl/biz/*.proto; do
            info "Processing $proto..."
            hz update -idl "$proto" \
                -I idl \
                --handler_dir biz/handler/hz \
                --model_dir biz/model/hz || warn "Failed to process $proto"
        done
        
        info "Hz code generation completed"
    else
        warn "No proto files found in idl/biz/. Skipping Hz code generation."
    fi
else
    warn "Skipped Hz code generation"
fi

# 4. 整理 Go 依赖
info "Tidying Go modules..."
go mod tidy

# 5. 格式化代码
info "Formatting code..."
go fmt ./...

echo ""
echo "=========================================="
info "Code generation completed successfully!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  1. Review generated code in biz/handler/hz/ and biz/router/hz/"
echo "  2. Implement business logic in generated handlers"
echo "  3. Run 'make build' to compile the project"
echo "  4. Run 'make run' to start the service"
