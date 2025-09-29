package main

import (
    "net/http"
)

func main() {
    // API v1 - 设备资源
    http.HandleFunc("/api/v1/devices", devicesCollectionHandler)     // GET list, POST create
    http.HandleFunc("/api/v1/devices/", deviceResourceHandler)      // GET/PUT/DELETE by id

    // API v1 - 认证资源
    http.HandleFunc("/api/v1/auth/sessions", authSessionsHandler)   // POST login, DELETE logout
    http.HandleFunc("/api/v1/auth/users", authUsersHandler)        // POST register
    http.HandleFunc("/api/v1/auth/codes", authCodesHandler)        // POST send verification code
    http.HandleFunc("/api/v1/auth/qr-tickets", authQRTicketsHandler) // POST generate QR ticket

    // API v1 - 工作流资源
    http.HandleFunc("/api/v1/workflows", workflowsCollectionHandler) // GET list, POST create
    http.HandleFunc("/api/v1/workflows/", workflowResourceHandler)   // GET/PUT/DELETE by id

    _ = http.ListenAndServe(":8080", nil)
}
