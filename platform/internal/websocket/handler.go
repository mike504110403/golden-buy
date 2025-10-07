package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允許所有來源（生產環境應該限制）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler WebSocket 處理器
type Handler struct {
	hub *Hub
}

// NewHandler 創建新的 WebSocket 處理器
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

// ServeWS 處理 WebSocket 連接
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		hub:  h.hub,
		conn: conn,
		send: make(chan []byte, 256),
	}

	client.hub.register <- client

	// 在新的 goroutine 中處理讀寫
	go client.writePump()
	go client.readPump()
}

// GetHub 獲取 Hub 實例
func (h *Handler) GetHub() *Hub {
	return h.hub
}
