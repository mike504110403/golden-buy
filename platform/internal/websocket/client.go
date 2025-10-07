package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 寫入等待時間
	writeWait = 10 * time.Second

	// Pong 等待時間
	pongWait = 60 * time.Second

	// Ping 發送間隔
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

// Client WebSocket 客戶端
type Client struct {
	hub *Hub

	// WebSocket 連接
	conn *websocket.Conn

	// 發送緩衝區
	send chan []byte
}

// readPump 從 WebSocket 連接讀取消息
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error: %v", err)
			}
			break
		}

		// 處理客戶端消息
		c.handleMessage(message)
	}
}

// writePump 向 WebSocket 連接寫入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 關閉了發送通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 將排隊的消息一起寫入
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 處理客戶端消息
func (c *Client) handleMessage(message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("❌ Failed to unmarshal message: %v", err)
		c.sendError("Invalid message format")
		return
	}

	switch msg.Type {
	case "subscribe":
		// 訂閱單個或多個商品
		if msg.Symbol != "" {
			c.hub.subscribe <- &Subscription{
				Client: c,
				Symbol: msg.Symbol,
			}
		}

		if len(msg.Symbols) > 0 {
			for _, symbol := range msg.Symbols {
				c.hub.subscribe <- &Subscription{
					Client: c,
					Symbol: symbol,
				}
			}
		}

	case "unsubscribe":
		// 取消訂閱
		if msg.Symbol != "" {
			c.hub.unsubscribe <- &Subscription{
				Client: c,
				Symbol: msg.Symbol,
			}
		}

		if len(msg.Symbols) > 0 {
			for _, symbol := range msg.Symbols {
				c.hub.unsubscribe <- &Subscription{
					Client: c,
					Symbol: symbol,
				}
			}
		}

	case "ping":
		// 回應 pong
		c.sendPong()

	default:
		log.Printf("⚠️  Unknown message type: %s", msg.Type)
		c.sendError("Unknown message type")
	}
}

// sendError 發送錯誤消息
func (c *Client) sendError(errMsg string) {
	msg := Message{
		Type:  "error",
		Error: errMsg,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case c.send <- data:
	default:
		// 發送緩衝區已滿
	}
}

// sendPong 發送 pong 消息
func (c *Client) sendPong() {
	msg := Message{
		Type: "pong",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case c.send <- data:
	default:
		// 發送緩衝區已滿
	}
}

