package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Client 一个 WebSocket 连接
type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte   // 写缓冲，避免并发写
	Done   chan struct{} // 连接关闭信号
}

// Hub 管理所有连接（userID → 连接列表，支持一个用户多设备）
type Hub struct {
	clients map[int64][]*Client
	mu      sync.RWMutex
}

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{clients: make(map[int64][]*Client)}
}

// DefaultHub 全局 Hub 实例
var DefaultHub = NewHub()

// Register 注册连接
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c.UserID] = append(h.clients[c.UserID], c)
}

// Unregister 注销连接
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	list := h.clients[c.UserID]
	for i, item := range list {
		if item == c {
			h.clients[c.UserID] = append(list[:i], list[i+1:]...)
			break
		}
	}
	if len(h.clients[c.UserID]) == 0 {
		delete(h.clients, c.UserID)
	}
}

// Push 推送给某用户的所有连接（非阻塞，连接已关闭则跳过）
func (h *Hub) Push(userID int64, data []byte) {
	h.mu.RLock()
	list := h.clients[userID]
	h.mu.RUnlock()

	for _, c := range list {
		select {
		case c.Send <- data:
		case <-c.Done: // 连接已关闭，跳过
		}
	}
}
