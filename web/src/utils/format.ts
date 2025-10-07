import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'

// 啟用 UTC 和 timezone 插件
dayjs.extend(utc)
dayjs.extend(timezone)

// 設置預設時區為 UTC+8 (Asia/Taipei)
dayjs.tz.setDefault('Asia/Taipei')

// 格式化價格
export const formatPrice = (price: number, decimals: number = 2): string => {
  return price.toFixed(decimals)
}

// 格式化金額（帶千分位）
export const formatCurrency = (amount: number, decimals: number = 2): string => {
  return new Intl.NumberFormat('zh-TW', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  }).format(amount)
}

// 格式化百分比
export const formatPercent = (value: number, decimals: number = 2): string => {
  const sign = value >= 0 ? '+' : ''
  return `${sign}${value.toFixed(decimals)}%`
}

// 格式化時間（使用 UTC+8 時區）
export const formatTime = (timestamp: number, format: string = 'YYYY-MM-DD HH:mm:ss'): string => {
  return dayjs(timestamp).tz('Asia/Taipei').format(format)
}

// 格式化相對時間
export const formatRelativeTime = (timestamp: number): string => {
  const now = Date.now()
  const diff = now - timestamp
  
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (seconds < 60) return `${seconds} 秒前`
  if (minutes < 60) return `${minutes} 分鐘前`
  if (hours < 24) return `${hours} 小時前`
  return `${days} 天前`
}

// 自動計數的相對時間（用於響應式更新）
export const formatRelativeTimeReactive = (timestamp: number): string => {
  const now = Date.now()
  const diff = now - timestamp
  
  const seconds = Math.floor(diff / 1000)
  
  if (seconds < 60) return `${seconds} 秒前`
  return formatRelativeTime(timestamp)
}

// 獲取價格變化類名
export const getPriceChangeClass = (changePercent: number): string => {
  if (changePercent > 0) return 'text-up'
  if (changePercent < 0) return 'text-down'
  return 'text-gray-500'
}

// 獲取價格變化圖標
export const getPriceChangeIcon = (changePercent: number): string => {
  if (changePercent > 0) return '↑'
  if (changePercent < 0) return '↓'
  return '→'
}


