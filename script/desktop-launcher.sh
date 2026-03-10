#!/bin/bash

# Git Manage Service Desktop Launcher
# This script starts the backend service and opens it in the default browser

set -e

APP_NAME="Git Manage Service"
PORT=12345
PID_FILE="/tmp/git-manage-service.pid"

# 检查是否已经在运行
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if ps -p $PID > /dev/null 2>&1; then
        echo "$APP_NAME is already running (PID: $PID)"
        echo "Opening browser..."
        open "http://localhost:$PORT"
        exit 0
    else
        # 清理过期的 PID 文件
        rm -f "$PID_FILE"
    fi
fi

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# 启动后端服务
echo "Starting $APP_NAME..."
"$SCRIPT_DIR/git-manage-service" --mode=all &
SERVER_PID=$!

# 保存 PID
echo $SERVER_PID > "$PID_FILE"

# 等待服务启动
echo "Waiting for service to start..."
sleep 3

# 检查服务是否启动成功
if ! ps -p $SERVER_PID > /dev/null 2>&1; then
    echo "Failed to start service"
    rm -f "$PID_FILE"
    exit 1
fi

# 打开浏览器
echo "Opening browser..."
open "http://localhost:$PORT"

echo "$APP_NAME started successfully (PID: $SERVER_PID)"
echo "Press Ctrl+C to stop"

# 等待用户中断
trap "echo '\nStopping service...'; kill $SERVER_PID; rm -f '$PID_FILE'; exit 0" INT TERM

# 保持脚本运行
wait $SERVER_PID
