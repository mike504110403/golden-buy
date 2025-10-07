import type { MetalSymbol, WSMessage, WSSubscribeMessage, WSPriceUpdate } from '../types'

type MessageHandler = (message: WSMessage) => void
type PriceUpdateHandler = (data: WSPriceUpdate) => void

export class WebSocketService {
  private ws: WebSocket | null = null
  private url: string
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 3000
  private heartbeatInterval: number | null = null
  private messageHandlers: Set<MessageHandler> = new Set()
  private priceUpdateHandlers: Set<PriceUpdateHandler> = new Set()
  private isManualClose = false

  constructor(url?: string) {
    this.url = url || import.meta.env.VITE_WS_BASE_URL + '/ws/prices' || 'ws://localhost:8080/ws/prices'
  }

  // 連接 WebSocket
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.isManualClose = false
        this.ws = new WebSocket(this.url)

        this.ws.onopen = () => {
          console.log('✅ WebSocket 已連接')
          this.reconnectAttempts = 0
          this.startHeartbeat()
          resolve()
        }

    this.ws.onmessage = (event) => {
      try {
        console.log('📨 收到 WebSocket 消息:', event.data)
        
        // 處理多個 JSON 對象的情況（用換行符分隔）
        const data = event.data.trim()
        if (data.includes('\n')) {
          // 多個 JSON 對象
          const lines = data.split('\n').filter((line: string) => line.trim())
          for (const line of lines) {
            try {
              const message: WSMessage = JSON.parse(line)
              this.handleMessage(message)
            } catch (lineError: any) {
              console.error('解析單行 WebSocket 消息失敗:', lineError, 'Line:', line)
            }
          }
        } else {
          // 單個 JSON 對象
          const message: WSMessage = JSON.parse(data)
          this.handleMessage(message)
        }
      } catch (error) {
        console.error('解析 WebSocket 消息失敗:', error, 'Data:', event.data)
      }
    }

        this.ws.onerror = (error) => {
          console.error('❌ WebSocket 錯誤:', error)
          reject(error)
        }

        this.ws.onclose = () => {
          console.log('🔌 WebSocket 已斷開')
          this.stopHeartbeat()
          
          if (!this.isManualClose && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnect()
          }
        }
      } catch (error) {
        reject(error)
      }
    })
  }

  // 斷開連接
  disconnect() {
    this.isManualClose = true
    this.stopHeartbeat()
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  // 重新連接
  private reconnect() {
    this.reconnectAttempts++
    console.log(`🔄 重新連接中... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
    
    setTimeout(() => {
      this.connect().catch((error) => {
        console.error('重新連接失敗:', error)
      })
    }, this.reconnectDelay)
  }

  // 訂閱商品
  subscribe(symbols: MetalSymbol | MetalSymbol[]) {
    const message: WSSubscribeMessage = {
      type: 'subscribe',
      symbols: Array.isArray(symbols) ? symbols : [symbols]
    }
    this.send(message)
  }

  // 取消訂閱
  unsubscribe(symbols: MetalSymbol | MetalSymbol[]) {
    const message: WSSubscribeMessage = {
      type: 'unsubscribe',
      symbols: Array.isArray(symbols) ? symbols : [symbols]
    }
    this.send(message)
  }

  // 發送消息
  private send(message: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket 未連接，無法發送消息')
    }
  }

  // 處理消息
  private handleMessage(message: WSMessage) {
    console.log('🔄 處理 WebSocket 消息:', message)
    
    // 觸發通用消息處理器
    this.messageHandlers.forEach(handler => handler(message))

    // 特殊處理價格更新
    if (message.type === 'price_update' && message.data) {
      console.log('💰 價格更新消息:', message.data)
      this.priceUpdateHandlers.forEach(handler => handler(message.data))
    }
  }

  // 註冊消息處理器
  onMessage(handler: MessageHandler) {
    this.messageHandlers.add(handler)
    return () => this.messageHandlers.delete(handler)
  }

  // 註冊價格更新處理器
  onPriceUpdate(handler: PriceUpdateHandler) {
    this.priceUpdateHandlers.add(handler)
    return () => this.priceUpdateHandlers.delete(handler)
  }

  // 心跳檢測
  private startHeartbeat() {
    this.heartbeatInterval = window.setInterval(() => {
      this.send({ type: 'ping' })
    }, 30000) // 每 30 秒發送一次心跳
  }

  private stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval)
      this.heartbeatInterval = null
    }
  }

  // 獲取連接狀態
  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  // 獲取連接狀態描述
  get connectionState(): string {
    if (!this.ws) return 'CLOSED'
    switch (this.ws.readyState) {
      case WebSocket.CONNECTING: return 'CONNECTING'
      case WebSocket.OPEN: return 'OPEN'
      case WebSocket.CLOSING: return 'CLOSING'
      case WebSocket.CLOSED: return 'CLOSED'
      default: return 'UNKNOWN'
    }
  }
}

// 單例
export const wsService = new WebSocketService()

// 調試用：輸出 WebSocket URL
console.log('🔗 WebSocket URL:', wsService['url'])

export default wsService

