#!/bin/bash
# 清理无用构建产物，打包当前 server 目录源码，并通过 scp 上传到远端
# 默认上传到 67_root:/root/，可通过 REMOTE 环境变量覆盖
# 用法：
#   sh ./pack_and_ship.sh
#   REMOTE=user@host:/path sh ./pack_and_ship.sh

set -euo pipefail

# 进入脚本所在目录（server/）
cd "$(dirname "$0")"

APP_NAME=server
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
DIST_DIR=dist
ARCHIVE_NAME="${APP_NAME}-src-${TIMESTAMP}.tar.gz"
ARCHIVE_PATH="${DIST_DIR}/${ARCHIVE_NAME}"

# 远端目标（可通过 REMOTE 覆盖）
REMOTE_DEFAULT="67_root:/root/"
REMOTE_TARGET="${REMOTE:-$REMOTE_DEFAULT}"

# 1) 清理本目录构建产物
if [ -f ./clean.sh ]; then
  echo "[1/3] Cleaning build artifacts via clean.sh ..."
  # 用 sh 执行以避免可执行权限问题
  sh ./clean.sh || true
else
  echo "[1/3] No clean.sh found, skipping explicit clean step."
fi

# 确保 dist 目录存在（在清理之后再创建，避免被清理脚本删除）
mkdir -p "$DIST_DIR"

# 2) 打包源码（排除二进制与产物目录）
#    按需排除：dist/、server、server.exe、.git、*.tar.gz
#    保留：*.go、go.mod、go.sum、*.sh、*.md 等项目文件

echo "[2/3] Creating source archive: ${ARCHIVE_PATH}"
# 先删除旧的同名文件（若存在）
rm -f "$ARCHIVE_PATH" || true

tar \
  --exclude "./${DIST_DIR}" \
  --exclude "./server" \
  --exclude "./server.exe" \
  --exclude "./.git" \
  --exclude "*.tar.gz" \
  -czf "$ARCHIVE_PATH" \
  .

echo "Archive created: $ARCHIVE_PATH"

# 3) 通过 scp 上传
if ! command -v scp >/dev/null 2>&1; then
  echo "[3/3] scp command not found. Please install OpenSSH client or ensure scp is available." >&2
  exit 1
fi

echo "[3/3] Uploading to ${REMOTE_TARGET}"
scp "$ARCHIVE_PATH" "$REMOTE_TARGET"

echo "Done. Uploaded ${ARCHIVE_NAME} to ${REMOTE_TARGET}"
