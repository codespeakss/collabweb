package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// 升级 HTTP 为 WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

// 处理 WebSocket 连接
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("升级 WebSocket 失败:", err)
		return
	}
	defer conn.Close()

	counter := 1
	for {
		message := fmt.Sprintf("当前消息编号: %d", counter)
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("发送消息失败:", err)
			break
		}
		counter++
		time.Sleep(1 * time.Second) // 每秒发送一次
	}
}

// 处理浏览器访问页面
func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>WebSocket 实时更新</title>
</head>
<body>
	<h1>WebSocket 实时消息</h1>
	<div id="messages"></div>

	<script>
		let ws = new WebSocket("ws://" + location.host + "/ws");
		let messagesDiv = document.getElementById("messages");

		ws.onmessage = function(event) {
			let p = document.createElement("p");
			p.textContent = event.data;
			messagesDiv.appendChild(p);
		};
	</script>
</body>
</html>`
	w.Write([]byte(html))
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/ws", wsHandler)

	fmt.Println("服务器启动，访问 http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

