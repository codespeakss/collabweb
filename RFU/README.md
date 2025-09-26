# Recently Frequently Used (RFU)

以下是最近经常使用的命令汇总，便于复制粘贴：

## 部署与打包
```bash
sh server/pack_and_ship.sh; sh scripts/deploy-to-nginx.sh
```

## 本地快速运行（解压并运行服务端）
```bash
cd ~; mkdir -p server_src; tar -xzf server-src-*.tar.gz -C server_src; rm -rf server-*.tar.gz; cd server_src; go mod tidy; go run .
```
