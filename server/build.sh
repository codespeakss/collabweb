#!/bin/bash
# 构建 Go 服务端（支持多平台交叉编译，包含 Linux x86 版本）
set -euo pipefail

cd "$(dirname "$0")"

APP_NAME=server
DIST_DIR=dist
mkdir -p "$DIST_DIR"

# 默认构建目标：当前平台 + Linux x86（amd64 与 386）
TARGETS=(
  "$(go env GOOS)/$(go env GOARCH)"
  "linux/amd64"
  "linux/386"
)

# 支持通过环境变量覆盖，例如：
# TARGETS_OVERRIDE="linux/amd64 linux/386 darwin/arm64" bash build.sh
if [[ -n "${TARGETS_OVERRIDE:-}" ]]; then
  IFS=' ' read -r -a TARGETS <<< "$TARGETS_OVERRIDE"
fi

for target in "${TARGETS[@]}"; do
  IFS='/' read -r GOOS GOARCH <<< "$target"
  OUT="$DIST_DIR/${APP_NAME}-${GOOS}-${GOARCH}"
  EXT=""
  if [[ "$GOOS" == "windows" ]]; then EXT=".exe"; fi
  echo "Building ${GOOS}/${GOARCH} -> ${OUT}${EXT}"
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
    go build -trimpath -ldflags "-s -w" -o "${OUT}${EXT}" .
done

echo "Done. Artifacts are in $DIST_DIR/"
