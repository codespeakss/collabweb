package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// TimeResponse 定义返回的 JSON 结构
type TimeResponse struct {
	ISO8601   string `json:"iso8601"`   // 标准 ISO8601 格式带时区
	Timestamp int64  `json:"timestamp"` // Unix 时间戳（秒）
	Readable  string `json:"readable"`  // 可读格式（本地化）
	Zone      string `json:"zone"`      // 时区名称
}

// timeHandler 处理 /time 请求
func timeHandler(w http.ResponseWriter, r *http.Request) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func main() {
	http.HandleFunc("/time", timeHandler)

	addr := ":443" // HTTPS 默认端口
	log.Printf("Starting local HTTPS server at %s (endpoint: /time)", addr)

	// 使用自签名证书启动 HTTPS 服务
	// 需要在当前目录放置 server.crt 和 server.key 文件
	if err := http.ListenAndServeTLS(addr, "server.crt", "server.key", nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

