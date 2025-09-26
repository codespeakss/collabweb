package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "strings"
    "time"
)

type Device struct {
    ID         string `json:"id"` // d开头的12字节字符串
    Name       string `json:"name"`
    Type       string `json:"type"`
    LastOnline int64  `json:"lastOnline"` // 最近在线时间戳
}

// --- Workflow DAG types ---
type WorkflowNode struct {
    ID     string  `json:"id"`
    Name   string  `json:"name"`
    X      float64 `json:"x"`
    Y      float64 `json:"y"`
    Status string  `json:"status"`
    Desc   string  `json:"desc"`
}

type WorkflowEdge struct {
    From  string `json:"from"`
    To    string `json:"to"`
    Type  string `json:"type,omitempty"`
    Label string `json:"label,omitempty"`
}

type WorkflowResponse struct {
    Nodes []WorkflowNode `json:"nodes"`
    Edges []WorkflowEdge `json:"edges"`
}

func mockDevices() []Device {
    devices := make([]Device, 300)
    types := []string{"Sensor", "Actuator", "Gateway", "Camera"}
    for i := 0; i < 300; i++ {
        devices[i] = Device{
            ID:         "d" + fmt.Sprintf("%012d", i+1),
            Name:       fmt.Sprintf("设备%03d", i+1),
            Type:       types[i%len(types)],
            LastOnline: 1695638400 + int64(i*60),
        }
    }
    return devices
}

func devicesHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    devices := mockDevices()
    // 分页参数
    page := 1
    pageSize := 20
    q := r.URL.Query()
    if p := q.Get("page"); p != "" {
        if v, err := strconv.Atoi(p); err == nil && v > 0 {
            page = v
        }
    }
    if ps := q.Get("pageSize"); ps != "" {
        if v, err := strconv.Atoi(ps); err == nil && v > 0 {
            pageSize = v
        }
    }
    start := (page - 1) * pageSize
    end := start + pageSize
    if start > len(devices) {
        start = len(devices)
    }
    if end > len(devices) {
        end = len(devices)
    }
    paged := devices[start:end]
    // 返回分页数据和总数
    resp := map[string]interface{}{
        "devices":  paged,
        "total":    len(devices),
        "page":     page,
        "pageSize": pageSize,
    }
    json.NewEncoder(w).Encode(resp)
}

// --- Auth mocks ---
type sendCodeReq struct {
    Account string `json:"account"`
    Channel string `json:"channel"` // email | sms
}

type loginReq struct {
    Account  string `json:"account"`
    Password string `json:"password"`
    Code     string `json:"code"`
    MFA      string `json:"mfa"`
    Remember bool   `json:"remember"`
}

type registerReq struct {
    Account  string `json:"account"`
    Password string `json:"password"`
    Code     string `json:"code"`
    MFA      string `json:"mfa"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

func sendCodeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
        return
    }
    var req sendCodeReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Account == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
        return
    }
    // email-only
    if req.Channel != "email" || !strings.Contains(req.Account, "@") {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅支持邮箱注册/验证码"})
        return
    }
    // mock: always succeed
    writeJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("验证码已发送至 %s", req.Account)})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
        return
    }
    var req loginReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Account == "" || req.Password == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "账号或密码缺失"})
        return
    }
    // email-only for consistency on login as well
    if !strings.Contains(req.Account, "@") {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅支持邮箱登录"})
        return
    }
    // mock rule: password must be at least 6 chars
    if len(req.Password) < 6 {
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "密码太短(>=6)"})
        return
    }
    token := fmt.Sprintf("mock-token-%d", time.Now().Unix())
    writeJSON(w, http.StatusOK, map[string]interface{}{
        "token":   token,
        "user":    map[string]string{"account": req.Account},
        "remember": req.Remember,
    })
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
        return
    }
    var req registerReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Account == "" || req.Password == "" || req.Code == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请填写完整(账号/密码/验证码)"})
        return
    }
    if !strings.Contains(req.Account, "@") {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅支持邮箱注册"})
        return
    }
    if len(req.Password) < 6 {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "密码太短(>=6)"})
        return
    }
    writeJSON(w, http.StatusOK, map[string]string{"message": "注册成功"})
}

func qrTicketHandler(w http.ResponseWriter, r *http.Request) {
    // mock ticket
    t := fmt.Sprintf("qr-%d", time.Now().UnixNano())
    writeJSON(w, http.StatusOK, map[string]string{"ticket": t})
}

// mock workflow data (same layout as the current Vue demo)
func mockWorkflow() WorkflowResponse {
    nodes := []WorkflowNode{
        {ID: "A", Name: "任务A", X: 390, Y: 40, Status: "success", Desc: "数据准备"},
        {ID: "B", Name: "任务B", X: 250, Y: 120, Status: "running", Desc: "数据清洗"},
        {ID: "C", Name: "任务C", X: 530, Y: 120, Status: "success", Desc: "特征工程"},
        {ID: "D", Name: "任务D", X: 170, Y: 220, Status: "pending", Desc: "模型训练1"},
        {ID: "E", Name: "任务E", X: 310, Y: 220, Status: "pending", Desc: "模型训练2"},
        {ID: "F", Name: "任务F", X: 470, Y: 220, Status: "failed", Desc: "模型训练3"},
        {ID: "G", Name: "任务G", X: 610, Y: 220, Status: "pending", Desc: "模型训练4"},
        {ID: "H", Name: "任务H", X: 250, Y: 320, Status: "success", Desc: "评估1"},
        {ID: "I", Name: "任务I", X: 530, Y: 320, Status: "pending", Desc: "评估2"},
        {ID: "J", Name: "任务J", X: 390, Y: 400, Status: "pending", Desc: "结果汇总"},
        {ID: "K", Name: "任务K", X: 110, Y: 120, Status: "pending", Desc: "数据采样"},
        {ID: "L", Name: "任务L", X: 670, Y: 120, Status: "pending", Desc: "规则生成"},
        {ID: "M", Name: "任务M", X: 110, Y: 320, Status: "pending", Desc: "可视化报告"},
        {ID: "N", Name: "任务N", X: 670, Y: 320, Status: "pending", Desc: "在线部署"},
    }
    edges := []WorkflowEdge{
        {From: "A", To: "K", Type: "conditional", Label: "if sample"},
        {From: "A", To: "B"},
        {From: "A", To: "C"},
        {From: "B", To: "D"},
        {From: "B", To: "E"},
        {From: "C", To: "F"},
        {From: "C", To: "G"},
        {From: "K", To: "D", Type: "conditional", Label: "small set"},
        {From: "L", To: "G", Type: "conditional", Label: "rule add"},
        {From: "D", To: "H"},
        {From: "E", To: "H"},
        {From: "F", To: "I"},
        {From: "G", To: "I"},
        {From: "H", To: "J"},
        {From: "I", To: "J"},
        {From: "H", To: "M", Type: "conditional", Label: "report"},
        {From: "I", To: "N", Type: "conditional", Label: "deploy"},
        {From: "A", To: "L", Type: "conditional", Label: "if rules"},
    }
    return WorkflowResponse{Nodes: nodes, Edges: edges}
}

func workflowHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
        return
    }
    wf := mockWorkflow()
    writeJSON(w, http.StatusOK, wf)
}

func main() {
    http.HandleFunc("/api/devices", devicesHandler)
    http.HandleFunc("/api/auth/send-code", sendCodeHandler)
    http.HandleFunc("/api/auth/login", loginHandler)
    http.HandleFunc("/api/auth/register", registerHandler)
    http.HandleFunc("/api/auth/qr-ticket", qrTicketHandler)
    http.HandleFunc("/api/workflow", workflowHandler)
    _ = http.ListenAndServe(":8080", nil)
}
