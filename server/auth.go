package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// 认证相关请求结构
type CreateSessionRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Code     string `json:"code"`
	MFA      string `json:"mfa"`
	Remember bool   `json:"remember"`
}

type CreateUserRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Code     string `json:"code"`
	MFA      string `json:"mfa"`
}

type CreateCodeRequest struct {
	Account string `json:"account"`
	Channel string `json:"channel"` // email | sms
}

// 认证相关响应结构
type SessionResponse struct {
	Token    string            `json:"token"`
	User     map[string]string `json:"user"`
	Remember bool              `json:"remember"`
}

type CodeResponse struct {
	Message string `json:"message"`
	CodeID  string `json:"codeId"`
}

type QRTicketResponse struct {
	Ticket    string `json:"ticket"`
	ExpiresAt int64  `json:"expiresAt"`
}

// POST /api/v1/auth/sessions (login), DELETE /api/v1/auth/sessions (logout)
func authSessionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createSession(w, r) // 登录
	case http.MethodDelete:
		deleteSession(w, r) // 登出
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

// POST /api/v1/auth/users (register)
func authUsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createUser(w, r) // 注册
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

// POST /api/v1/auth/codes (send verification code)
func authCodesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createCode(w, r) // 发送验证码
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

// POST /api/v1/auth/qr-tickets (generate QR ticket)
func authQRTicketsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createQRTicket(w, r) // 生成二维码票据
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
	}
}

// 创建会话（登录）
func createSession(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	var req CreateSessionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	// 验证必填字段
	if strings.TrimSpace(req.Account) == "" || strings.TrimSpace(req.Password) == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "账号或密码缺失"})
		return
	}

	// 仅支持邮箱登录
	if !strings.Contains(req.Account, "@") {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "仅支持邮箱登录"})
		return
	}

	// 密码长度验证
	if len(req.Password) < 6 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "密码太短(>=6)"})
		return
	}

	// 生成会话令牌
	token := fmt.Sprintf("session-%d", time.Now().Unix())
	
	response := SessionResponse{
		Token:    token,
		User:     map[string]string{"account": req.Account},
		Remember: req.Remember,
	}

	writeJSON(w, http.StatusCreated, response)
}

// 删除会话（登出）
func deleteSession(w http.ResponseWriter, r *http.Request) {
	// 在真实实现中，这里会验证 Authorization header 并删除对应的会话
	// 目前是 mock 实现，直接返回成功
	writeJSON(w, http.StatusNoContent, nil)
}

// 创建用户（注册）
func createUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	var req CreateUserRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	// 验证必填字段
	if strings.TrimSpace(req.Account) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Code) == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "请填写完整(账号/密码/验证码)"})
		return
	}

	// 仅支持邮箱注册
	if !strings.Contains(req.Account, "@") {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "仅支持邮箱注册"})
		return
	}

	// 密码长度验证
	if len(req.Password) < 6 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "密码太短(>=6)"})
		return
	}

	// 在真实实现中，这里会验证验证码并创建用户
	writeJSON(w, http.StatusCreated, map[string]string{"message": "注册成功", "userId": fmt.Sprintf("user-%d", time.Now().Unix())})
}

// 创建验证码（发送验证码）
func createCode(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	var req CreateCodeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	// 验证必填字段
	if strings.TrimSpace(req.Account) == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "账号不能为空"})
		return
	}

	// 仅支持邮箱验证码
	if req.Channel != "email" || !strings.Contains(req.Account, "@") {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "仅支持邮箱注册/验证码"})
		return
	}

	// 生成验证码 ID
	codeId := fmt.Sprintf("code-%d", time.Now().UnixNano())
	
	response := CodeResponse{
		Message: fmt.Sprintf("验证码已发送至 %s", req.Account),
		CodeID:  codeId,
	}

	writeJSON(w, http.StatusCreated, response)
}

// 创建二维码票据
func createQRTicket(w http.ResponseWriter, r *http.Request) {
	// 生成二维码票据
	ticket := fmt.Sprintf("qr-ticket-%d", time.Now().UnixNano())
	expiresAt := time.Now().Add(5 * time.Minute).Unix() // 5分钟过期
	
	response := QRTicketResponse{
		Ticket:    ticket,
		ExpiresAt: expiresAt,
	}

	writeJSON(w, http.StatusCreated, response)
}
