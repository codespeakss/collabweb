#!/bin/bash
# 构建 Go 服务端
cd "$(dirname "$0")"
go build -o server main.go
