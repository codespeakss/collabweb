# Golang 服务端部署说明

## 构建

```bash
cd server
bash build.sh
```

## 运行

```bash
./server
```

服务默认监听 8080 端口，接口示例：

- GET http://localhost:8080/api/devices

返回模拟设备数据 JSON。

## 依赖
- Go 1.18+

如需修改端口或数据，请编辑 `main.go`。
