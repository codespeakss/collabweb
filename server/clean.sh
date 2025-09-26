#!/bin/bash
# 清理 server 目录的构建产物
set -euo pipefail

cd "$(dirname "$0")"

# 删除二进制产物
rm -f server server.exe || true

# 删除多平台构建产物目录
rm -rf dist || true

# 可选：清理 go 构建缓存（保留依赖模块缓存以避免重复下载）
(go clean -cache -testcache >/dev/null 2>&1) || true

echo "Cleaned server build artifacts."
