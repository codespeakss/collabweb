package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// TimeResponse 定义返回的 JSON 结构
type TimeResponse struct {
	ISO8601   string `json:"iso8601"`    // 标准 ISO8601 格式带时区
	Timestamp int64  `json:"timestamp"`  // Unix 时间戳（秒）
	Readable  string `json:"readable"`   // 可读格式（本地化）
	Zone      string `json:"zone"`       // 时区名称
}

// timeHandler 处理 /time 请求
func timeHandler(w http.ResponseWriter, r *http.Request) {
	// 尝试加载中国大陆时区（Asia/Shanghai）
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如果加载失败则使用系统本地时区作为回退
		loc = time.Local
	}

	now := time.Now().In(loc)

	resp := TimeResponse{
		ISO8601:   now.Format(time.RFC3339),
		Timestamp: now.Unix(),
		Readable:  now.Format("2006-01-02 15:04:05"),
		Zone:      now.Location().String(),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// 可选：允许跨域简单测试（如果只是本地用可以移除）
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// 写入错误时记录日志（不向客户端返回详细错误）
		log.Printf("failed to encode response: %v", err)
	}
}

func main() {
	http.HandleFunc("/time", timeHandler)

	addr := ":80"
	log.Printf("Starting local time server at %s (endpoint: /time)", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

