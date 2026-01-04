#!/bin/bash

APP_NAME="git-manage-service"
# Based on Makefile build-all output
APP_BIN="./output/git-manage-service"
PID_FILE="run.pid"
LOG_FILE="app.log"

build() {
    echo "Building $APP_NAME..."
    # make build-all
    CGO_ENABLED=1 go build -o output/git-manage-service cmd/all/main.go
    if [ $? -ne 0 ]; then
        echo "Build failed!"
        exit 1
    fi
    echo "Build success!"
}

start() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat $PID_FILE)
        if ps -p $PID > /dev/null; then
            echo "$APP_NAME is already running (PID: $PID)"
            return
        else
            echo "PID file exists but process is not running. Removing stale PID file."
            rm $PID_FILE
        fi
    fi

    if [ ! -f "$APP_BIN" ]; then
        build
    fi

    echo "Starting $APP_NAME..."
    nohup $APP_BIN > $LOG_FILE 2>&1 &
    echo $! > $PID_FILE
    echo "$APP_NAME started (PID: $(cat $PID_FILE))"
}

stop() {
    if [ ! -f "$PID_FILE" ]; then
        echo "$APP_NAME is not running (PID file not found)"
        return
    fi

    PID=$(cat $PID_FILE)
    if ps -p $PID > /dev/null; then
        echo "Stopping $APP_NAME (PID: $PID)..."
        kill $PID
        # Wait for process to exit
        for i in {1..10}; do
            if ! ps -p $PID > /dev/null; then
                break
            fi
            sleep 1
        done
        if ps -p $PID > /dev/null; then
             echo "Process did not stop, forcing kill..."
             kill -9 $PID
        fi
    else
        echo "Process $PID is not running."
    fi
    rm $PID_FILE
    echo "$APP_NAME stopped"
}

restart() {
    stop
    sleep 1
    start
}

status() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat $PID_FILE)
        if ps -p $PID > /dev/null; then
            echo "$APP_NAME is running (PID: $PID)"
        else
            echo "$APP_NAME is not running (stale PID file)"
        fi
    else
        echo "$APP_NAME is not running"
    fi
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
    *)
        echo "Usage: $0 {start|stop|restart|build|status}"
        exit 1
        ;;
esac
