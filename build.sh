#!/bin/bash
# 重要: SQLite 驱动需要 CGO, 因此必须设置 CGO_ENABLED=1
export CGO_ENABLED=1
RUN_NAME=hertz_service
mkdir -p output/bin
cp script/* output 2>/dev/null
chmod +x output/bootstrap.sh
go build -o output/bin/${RUN_NAME}