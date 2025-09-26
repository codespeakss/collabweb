# Golang 服务端部署说明

## 构建

默认会构建以下目标：
- 当前平台（`$(go env GOOS)/$(go env GOARCH)`）
- Linux x86 版本：`linux/amd64` 与 `linux/386`

构建命令：
```bash
cd server
bash build.sh
# 产物输出在 server/dist/ 下，例如：
# dist/server-darwin-amd64
# dist/server-linux-amd64
# dist/server-linux-386
```
如需自定义构建目标（覆盖默认列表），可使用环境变量 `TARGETS_OVERRIDE`，以空格分隔多个 `<GOOS>/<GOARCH>`：
```bash
# 例如仅构建 Linux amd64 与 Windows amd64
cd server
TARGETS_OVERRIDE="linux/amd64 windows/amd64" bash build.sh
```
## 运行

构建完成后，请在 `dist/` 下选择对应平台的二进制运行。例如：

- 在 Linux x86_64 服务器上运行：
  ```bash
  cd server/dist
  ./server-linux-amd64
  ```
- 在 Linux 32 位（x86）环境上运行：
  ```bash
  cd server/dist
  ./server-linux-386
  ```
- 在本机（如 macOS）上运行当前平台产物：
  ```bash
  cd server/dist
  ./server-$(go env GOOS)-$(go env GOARCH)
  ```

服务默认监听 8080 端口，接口示例：

- GET http://localhost:8080/api/devices
- POST http://localhost:8080/api/auth/send-code
- POST http://localhost:8080/api/auth/login
- POST http://localhost:8080/api/auth/register
- GET  http://localhost:8080/api/auth/qr-ticket
- GET  http://localhost:8080/api/workflow

返回模拟设备数据 JSON。

## 依赖
- Go 1.18+

## 代码结构

服务端代码已按“功能模块”拆分：

- `main.go`：仅负责注册路由并启动 HTTP 服务。
- `common.go`：通用工具（例如 `writeJSON`）。
- `devices.go`：设备相关类型与 `devicesHandler`。
- `auth.go`：认证相关处理器（发送验证码、登录、注册、二维码 ticket）。
- `workflow.go`：工作流 DAG 的数据与 `workflowHandler`。

如需修改端口或新增路由，请编辑 `main.go`；
如需修改各功能的返回数据或逻辑，请编辑对应的功能文件。
