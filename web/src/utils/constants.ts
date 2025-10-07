import type { MetalInfo, MetalSymbol } from '../types'

// 貴金屬資訊
export const METAL_INFO: Record<MetalSymbol, MetalInfo> = {
  GOLD: {
    symbol: 'GOLD',
    name: 'Gold',
    nameCn: '黃金',
    icon: '🥇',
    color: '#FFD700'
  },
  SILVER: {
    symbol: 'SILVER',
    name: 'Silver',
    nameCn: '白銀',
    icon: '🥈',
    color: '#C0C0C0'
  },
  PLATINUM: {
    symbol: 'PLATINUM',
    name: 'Platinum',
    nameCn: '鉑金',
    icon: '⚪',
    color: '#E5E4E2'
  },
  PALLADIUM: {
    symbol: 'PALLADIUM',
    name: 'Palladium',
    nameCn: '鈀金',
    icon: '⚫',
    color: '#C9C0BB'
  }
}

// K 線間隔選項
export const KLINE_INTERVALS = [
  { label: '1 分鐘', value: '1m' },
  { label: '5 分鐘', value: '5m' },
  { label: '15 分鐘', value: '15m' },
  { label: '30 分鐘', value: '30m' },
  { label: '1 小時', value: '1h' },
  { label: '4 小時', value: '4h' },
  { label: '1 天', value: '1d' }
]

// 獲取貴金屬資訊
export const getMetalInfo = (symbol: MetalSymbol): MetalInfo => {
  return METAL_INFO[symbol]
}

// 所有貴金屬符號
export const ALL_METALS: MetalSymbol[] = ['GOLD', 'SILVER', 'PLATINUM', 'PALLADIUM']
export const METAL_SYMBOLS = ALL_METALS


