// API Response 類型
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

// 貴金屬符號
export type MetalSymbol = 'GOLD' | 'SILVER' | 'PLATINUM' | 'PALLADIUM'

// 價格資料
export interface Price {
  symbol: MetalSymbol
  price: number
  change_percent: number
  timestamp: number
  updated_at: string
}

// 價格列表
export type PriceMap = Record<MetalSymbol, Price>

// K 線資料
export interface Kline {
  timestamp: number
  open: number
  high: number
  low: number
  close: number
  volume: number
}

// K 線查詢參數
export interface KlineQuery {
  symbol: MetalSymbol
  interval: '1m' | '5m' | '15m' | '30m' | '1h' | '4h' | '1d'
  start?: number
  end?: number
  limit?: number
}

// K 線回應
export interface KlineResponse {
  symbol: MetalSymbol
  interval: string
  count: number
  klines: Kline[]
}

// 用戶資料
export interface User {
  id: string
  username: string
  email: string
  balance: number
  role: string
}

// WebSocket 消息類型
export type WSMessageType = 'connected' | 'subscribed' | 'unsubscribed' | 'price_update' | 'error' | 'pong'

// WebSocket 消息
export interface WSMessage {
  type: WSMessageType
  symbol?: MetalSymbol
  symbols?: MetalSymbol[]
  data?: any
  error?: string
}

// 訂閱消息
export interface WSSubscribeMessage {
  type: 'subscribe' | 'unsubscribe'
  symbol?: MetalSymbol
  symbols?: MetalSymbol[]
}

// 價格更新消息
export interface WSPriceUpdate {
  symbol: MetalSymbol
  price: number
  change_percent: number
  timestamp: number
}

// 貴金屬資訊
export interface MetalInfo {
  symbol: MetalSymbol
  name: string
  nameCn: string
  icon: string
  color: string
}

