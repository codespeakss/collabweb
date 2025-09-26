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

    // 工作流（兼容旧接口）
    http.HandleFunc("/api/workflow", workflowHandler)
    // 工作流列表与详情
    http.HandleFunc("/api/workflows", workflowsListHandler)   // GET list
    http.HandleFunc("/api/workflows/", workflowDetailHandler) // GET detail by id

    _ = http.ListenAndServe(":8080", nil)
}
