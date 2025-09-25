#!/bin/bash
# 构建并发布前端产物到 nginx 目录


# 1. 构建前端
npm install
npm run build

# 2.  nginx 目标目录
NGINX_DIR="67_root:/var/www/dev.bewantbe.com"

# 3. 拷贝构建产物到 nginx 目录
scp -r dist/* "$NGINX_DIR"/

echo "发布完成，前端已部署到 $NGINX_DIR"
