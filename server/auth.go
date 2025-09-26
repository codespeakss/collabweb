package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

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
		"token":    token,
		"user":     map[string]string{"account": req.Account},
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
