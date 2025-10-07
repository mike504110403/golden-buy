import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { MetalSymbol, Price, PriceMap, Kline } from '../types'
import { priceApi } from '../api'
import { wsService } from '../api/websocket'

export const usePriceStore = defineStore('price', () => {
  // State
  const prices = ref<PriceMap>({} as PriceMap)
  const klines = ref<Record<MetalSymbol, Kline[]>>({} as Record<MetalSymbol, Kline[]>)
  const loading = ref(false)
  const wsConnected = ref(false)

  // Getters
  const getPrice = computed(() => (symbol: MetalSymbol) => {
    const price = prices.value[symbol]
    console.log('🔍 getPrice:', symbol, price)
    return price
  })

  const getAllPrices = computed(() => prices.value)

  const getKlines = computed(() => (symbol: MetalSymbol) => {
    return klines.value[symbol] || []
  })

  // Actions
  const fetchCurrentPrices = async () => {
    try {
      loading.value = true
      const response = await priceApi.getCurrentPrices()
      if (response.success && response.data) {
        prices.value = response.data
      }
    } catch (error) {
      console.error('獲取價格失敗:', error)
    } finally {
      loading.value = false
    }
  }

  const fetchKlines = async (symbol: MetalSymbol, interval: string = '1m', limit: number = 100) => {
    try {
      const response = await priceApi.getKlines({
        symbol,
        interval: interval as any,
        limit
      })
      if (response.success && response.data) {
        klines.value[symbol] = response.data.klines
      }
    } catch (error) {
      console.error('獲取 K 線失敗:', error)
    }
  }

  const updatePrice = (symbol: MetalSymbol, price: Price) => {
    prices.value[symbol] = price
  }

  const connectWebSocket = async () => {
    try {
      console.log('🔌 正在連接 WebSocket...')
      await wsService.connect()
      wsConnected.value = true
      console.log('✅ WebSocket 連接成功')

      // 訂閱價格更新
      wsService.onPriceUpdate((data) => {
        console.log('📊 收到價格更新:', data)
        // 直接更新價格，不管是否已存在
        prices.value[data.symbol] = {
          symbol: data.symbol,
          price: data.price,
          change_percent: data.change_percent,
          timestamp: data.timestamp,
          updated_at: new Date(data.timestamp).toISOString()
        }
        console.log('✅ 價格已更新:', data.symbol, '=', data.price)
      })

      // 訂閱所有商品
      console.log('📡 訂閱所有商品...')
      wsService.subscribe(['GOLD', 'SILVER', 'PLATINUM', 'PALLADIUM'])
    } catch (error) {
      console.error('❌ WebSocket 連接失敗:', error)
      wsConnected.value = false
    }
  }

  const disconnectWebSocket = () => {
    wsService.disconnect()
    wsConnected.value = false
  }

  return {
    // State
    prices,
    klines,
    loading,
    wsConnected,
    
    // Getters
    getPrice,
    getAllPrices,
    getKlines,
    
    // Actions
    fetchCurrentPrices,
    fetchKlines,
    updatePrice,
    connectWebSocket,
    disconnectWebSocket
  }
})

