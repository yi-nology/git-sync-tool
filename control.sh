#!/bin/bash

APP_NAME="git-manage-service"
APP_BIN="./output/git-manage-service"
PID_FILE="run.pid"
LOG_FILE="app.log"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

build() {
    print_info "Building $APP_NAME..."
    
    # 使用 Makefile 构建（包含版本信息注入）
    make build-full
    
    if [ $? -ne 0 ]; then
        print_error "Build failed!"
        exit 1
    fi
    
    print_success "Build success!"
}

start() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat $PID_FILE)
        if ps -p $PID > /dev/null 2>&1; then
            print_warning "$APP_NAME is already running (PID: $PID)"
            return
        else
            print_info "PID file exists but process is not running. Removing stale PID file."
            rm $PID_FILE
        fi
    fi

    if [ ! -f "$APP_BIN" ]; then
        print_info "Binary not found, building..."
        build
    fi

    print_info "Starting $APP_NAME..."
    nohup $APP_BIN --mode=all > $LOG_FILE 2>&1 &
    echo $! > $PID_FILE
    
    # 等待启动
    sleep 2
    
    # 检查是否成功启动
    PID=$(cat $PID_FILE)
    if ps -p $PID > /dev/null 2>&1; then
        print_success "$APP_NAME started (PID: $PID)"
        print_info "Log file: $LOG_FILE"
        print_info "Check status: $0 status"
        print_info "View logs: tail -f $LOG_FILE"
    else
        print_error "$APP_NAME failed to start"
        print_info "Check log file: $LOG_FILE"
        rm $PID_FILE
        exit 1
    fi
}

stop() {
    if [ ! -f "$PID_FILE" ]; then
        print_warning "$APP_NAME is not running (PID file not found)"
        return
    fi

    PID=$(cat $PID_FILE)
    if ps -p $PID > /dev/null 2>&1; then
        print_info "Stopping $APP_NAME (PID: $PID)..."
        kill $PID
        
        # 等待进程退出
        for i in {1..10}; do
            if ! ps -p $PID > /dev/null 2>&1; then
                break
            fi
            sleep 1
        done
        
        # 如果还在运行，强制杀死
        if ps -p $PID > /dev/null 2>&1; then
            print_warning "Process did not stop gracefully, forcing kill..."
            kill -9 $PID
        fi
        
        print_success "$APP_NAME stopped"
    else
        print_warning "Process $PID is not running."
    fi
    
    rm $PID_FILE
}

restart() {
    print_info "Restarting $APP_NAME..."
    stop
    sleep 2
    start
}

status() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat $PID_FILE)
        if ps -p $PID > /dev/null 2>&1; then
            print_success "$APP_NAME is running (PID: $PID)"
            echo ""
            echo "Process info:"
            ps -p $PID -o pid,ppid,%cpu,%mem,vsz,rss,tty,stat,start,time,command
            echo ""
            echo "Listening ports:"
            lsof -Pan -p $PID -i 2>/dev/null | grep LISTEN || echo "  No listening ports found"
        else
            print_warning "$APP_NAME is not running (stale PID file)"
        fi
    else
        print_error "$APP_NAME is not running"
    fi
}

logs() {
    if [ ! -f "$LOG_FILE" ]; then
        print_error "Log file not found: $LOG_FILE"
        exit 1
    fi
    
    print_info "Showing logs from $LOG_FILE (Ctrl+C to exit)..."
    tail -f $LOG_FILE
}

version() {
    if [ ! -f "$APP_BIN" ]; then
        print_error "Binary not found, please build first: $0 build"
        exit 1
    fi
    
    $APP_BIN --version
}

help() {
    echo "Git Manage Service Control Script"
    echo ""
    echo "Usage: $0 {command}"
    echo ""
    echo "Commands:"
    echo "  start      Start the service"
    echo "  stop       Stop the service"
    echo "  restart    Restart the service"
    echo "  build      Build the service (with frontend)"
    echo "  status     Show service status"
    echo "  logs       View service logs (real-time)"
    echo "  version    Show version information"
    echo "  help       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build       # Build the service"
    echo "  $0 start       # Start the service"
    echo "  $0 status      # Check if running"
    echo "  $0 logs        # View logs"
    echo ""
}

case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    build)
        build
        ;;
    status)
        status
        ;;
    logs)
        logs
        ;;
    version)
        version
        ;;
    help|--help|-h)
        help
        ;;
    *)
        print_error "Unknown command: ${1:-}"
        echo ""
        help
        exit 1
        ;;
esac
