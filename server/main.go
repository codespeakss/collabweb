package main

import (
    "net/http"
)

func main() {
    // 设备列表
    http.HandleFunc("/api/devices", devicesHandler)

    // 认证相关
    http.HandleFunc("/api/auth/send-code", sendCodeHandler)
    http.HandleFunc("/api/auth/login", loginHandler)
    http.HandleFunc("/api/auth/register", registerHandler)
    http.HandleFunc("/api/auth/qr-ticket", qrTicketHandler)

    // 工作流
    http.HandleFunc("/api/workflow", workflowHandler)

    _ = http.ListenAndServe(":8080", nil)
}
