package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/mike/golden-buy/platform/internal/model"
)

// Hub WebSocket 連接管理器
type Hub struct {
	// 已註冊的客戶端
	clients map[*Client]bool

	// 客戶端訂閱的商品
	subscriptions map[string]map[*Client]bool // symbol -> clients

	// 來自客戶端的消息
	broadcast chan []byte

	// 註冊請求
	register chan *Client

	// 註銷請求
	unregister chan *Client

	// 訂閱請求
	subscribe chan *Subscription

	// 取消訂閱請求
	unsubscribe chan *Subscription

	// 保護 clients 和 subscriptions 的互斥鎖
	mu sync.RWMutex
}

// Subscription 訂閱請求
type Subscription struct {
	Client *Client
	Symbol string
}

// Message WebSocket 消息格式
type Message struct {
	Type    string      `json:"type"` // "subscribe", "unsubscribe", "price_update", "error"
	Symbol  string      `json:"symbol,omitempty"`
	Symbols []string    `json:"symbols,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PriceUpdate 價格更新消息
type PriceUpdate struct {
	Symbol        string  `json:"symbol"`
	Price         float64 `json:"price"`
	ChangePercent float64 `json:"change_percent"`
	Timestamp     int64   `json:"timestamp"`
}

// NewHub 創建新的 Hub
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		subscriptions: make(map[string]map[*Client]bool),
		broadcast:     make(chan []byte, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		subscribe:     make(chan *Subscription),
		unsubscribe:   make(chan *Subscription),
	}
}

// Run 啟動 Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("🔌 New WebSocket client connected (total: %d)", len(h.clients))

			// 發送歡迎消息
			welcomeMsg := Message{
				Type: "connected",
				Data: map[string]interface{}{
					"message": "Connected to Golden Buy Platform",
					"symbols": []string{"GOLD", "SILVER", "PLATINUM", "PALLADIUM"},
				},
			}
			if data, err := json.Marshal(welcomeMsg); err == nil {
				client.send <- data
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// 從所有訂閱中移除
				for symbol, clients := range h.subscriptions {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.subscriptions, symbol)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("🔌 WebSocket client disconnected (total: %d)", len(h.clients))

		case sub := <-h.subscribe:
			h.mu.Lock()
			if h.subscriptions[sub.Symbol] == nil {
				h.subscriptions[sub.Symbol] = make(map[*Client]bool)
			}
			h.subscriptions[sub.Symbol][sub.Client] = true
			h.mu.Unlock()
			log.Printf("📊 Client subscribed to %s", sub.Symbol)

			// 發送訂閱確認
			confirmMsg := Message{
				Type:   "subscribed",
				Symbol: sub.Symbol,
			}
			if data, err := json.Marshal(confirmMsg); err == nil {
				sub.Client.send <- data
			}

		case sub := <-h.unsubscribe:
			h.mu.Lock()
			if clients, ok := h.subscriptions[sub.Symbol]; ok {
				delete(clients, sub.Client)
				if len(clients) == 0 {
					delete(h.subscriptions, sub.Symbol)
				}
			}
			h.mu.Unlock()
			log.Printf("📊 Client unsubscribed from %s", sub.Symbol)

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastPrice 廣播價格更新到訂閱的客戶端
func (h *Hub) BroadcastPrice(price *model.Price) {
	h.mu.RLock()
	clients, ok := h.subscriptions[price.Symbol]
	h.mu.RUnlock()

	log.Printf("🔍 BroadcastPrice: symbol=%s, hasClients=%v, clientCount=%d", 
		price.Symbol, ok, len(clients))

	if !ok || len(clients) == 0 {
		log.Printf("⚠️ No clients subscribed to %s", price.Symbol)
		return
	}

	// 構建價格更新消息
	priceUpdate := PriceUpdate{
		Symbol:        price.Symbol,
		Price:         price.Price,
		ChangePercent: price.ChangePercent,
		Timestamp:     price.Timestamp,
	}

	msg := Message{
		Type:   "price_update",
		Symbol: price.Symbol,
		Data:   priceUpdate,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ Failed to marshal price update: %v", err)
		return
	}

	// 發送給所有訂閱該商品的客戶端
	h.mu.RLock()
	for client := range clients {
		select {
		case client.send <- data:
		default:
			// 客戶端發送緩衝區已滿，跳過
		}
	}
	h.mu.RUnlock()
}

// GetClientCount 獲取當前連接的客戶端數量
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetSubscriptionCount 獲取當前訂閱數量
func (h *Hub) GetSubscriptionCount() map[string]int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	counts := make(map[string]int)
	for symbol, clients := range h.subscriptions {
		counts[symbol] = len(clients)
	}

	return counts
}
