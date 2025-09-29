package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// 存储独立访客的 map
var uniqueVisitors = sync.Map{}

// TrackRequest 用于接收前端请求
type TrackRequest struct {
	VisitorID string `json:"visitorId"`
	Timestamp string `json:"timestamp"`
}

func main() {
	// 路由：根路径返回 HTML 页面
	http.HandleFunc("/", serveHTML)
	// 路由：API 记录 UV
	http.HandleFunc("/api/track-uv", trackUVHandler)

	// 启动 HTTP 服务
	port := ":8080"
	log.Printf("Server running on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}

// 返回 HTML 页面
func serveHTML(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>UV 统计示例</title>
    <script>
        function getUniqueVisitorId() {
            let visitorId = localStorage.getItem('visitorId');
            if (!visitorId) {
                visitorId = 'visitor_' + Date.now() + Math.random().toString(36).substr(2, 9);
                localStorage.setItem('visitorId', visitorId);
            }
            return visitorId;
        }

        function trackVisitor() {
            const visitorId = getUniqueVisitorId();
            fetch('/api/track-uv', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    visitorId: visitorId,
                    timestamp: new Date().toISOString(),
                }),
            }).then(response => response.json())
              .then(data => {
                  console.log('UV 数据已发送');
                  console.log('当前独立访客数:', data.uniqueVisitors);
                  document.getElementById('uvCount').innerText = '当前独立访客数：' + data.uniqueVisitors;
              }).catch(error => {
                  console.error('UV 数据发送失败', error);
              });
        }

        window.onload = function() {
            trackVisitor();
        };
    </script>
</head>
<body>
    <h1>欢迎来到我的网站</h1>
    <p id="uvCount">加载中...</p>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// 处理访客数据
func trackUVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req TrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	uniqueVisitors.Store(req.VisitorID, true)

	visitorCount := getUniqueVisitorCount()
	response := map[string]interface{}{
		"message":        "UV 数据已记录",
		"uniqueVisitors": visitorCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 获取当前独立访客的数量
func getUniqueVisitorCount() int {
	count := 0
	uniqueVisitors.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

