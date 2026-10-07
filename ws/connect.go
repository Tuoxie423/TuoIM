package ws

import (
	"log"
	"net/http"
	"time"

	"ginchat/cache"
	"ginchat/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: 生产环境限制来源
	},
}

// Connect 建立 WebSocket 连接（HTTP 升级）
func Connect(w http.ResponseWriter, r *http.Request) {
	// ① 从 URL query 解析 token（WebSocket 不能用 header）
	token := r.URL.Query().Get("token")
	claims, err := utils.NewJWT().ParseAccessToken(token)
	if err != nil {
		log.Printf("[WS] 鉴权失败: %v", err)
		return
	}

	// ② 升级连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] 升级失败: %v", err)
		return
	}
	log.Printf("[WS] 连接成功: userID=%d", claims.UserID)

	// ③ 创建 Client 并注册
	client := &Client{
		UserID: claims.UserID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Done:   make(chan struct{}),
	}
	DefaultHub.Register(client)
	cache.SetOnline(claims.UserID)

	// ④ 读写分离：writePump 单独 goroutine，readPump 独占读
	go client.writePump()

	// ⑤ 补推离线消息（上线时拉取）
	for _, data := range cache.GetOfflineMsg(claims.UserID) {
		client.Send <- []byte(data)
	}

	client.readPump()
}

// readPump 读循环（独占读；应用层心跳：收到任何消息续命，超时判死）
func (c *Client) readPump() {
	defer func() {
		log.Printf("[WS] 连接断开: userID=%d", c.UserID)
		close(c.Done)             // 通知 writePump 退出
		DefaultHub.Unregister(c)  // 从 Hub 移除
		cache.DelOnline(c.UserID) // 标记掉线
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // 读超时

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break // 超时/断开 → defer 清理
		}
		// 收到任何消息（含心跳）→ 续命
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}
}

// writePump 写循环（只写，监听 Done 退出）
func (c *Client) writePump() {
	for {
		select {
		case data := <-c.Send:
			c.Conn.WriteMessage(websocket.TextMessage, data)
		case <-c.Done:
			return
		}
	}
}
