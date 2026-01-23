#!/bin/bash
# script/bootstrap.sh - 服务启动脚本
CURDIR=$(cd $(dirname $0); pwd)
RUNTIME_ROOT=${1:-$CURDIR}

export KITEX_RUNTIME_ROOT=$RUNTIME_ROOT
export KITEX_LOG_DIR="$RUNTIME_ROOT/log"

# 创建日志目录
mkdir -p "$KITEX_LOG_DIR"

BinaryName=git-manage-service
echo "Starting $BinaryName..."
exec $CURDIR/bin/${BinaryName}