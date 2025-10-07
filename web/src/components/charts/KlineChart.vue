<template>
  <div class="kline-chart-container">
    <div class="chart-header">
      <div class="chart-title">
        <h3 class="text-lg font-semibold text-gray-900">{{ metalInfo.nameCn }} K線圖</h3>
        <p class="text-sm text-gray-500">{{ metalInfo.name }} - {{ interval }}</p>
      </div>
      <div class="chart-controls">
        <select v-model="selectedInterval" @change="onIntervalChange" class="px-3 py-1 border rounded text-sm">
          <option value="1m">1分鐘</option>
          <option value="5m">5分鐘</option>
          <option value="15m">15分鐘</option>
          <option value="30m">30分鐘</option>
          <option value="1h">1小時</option>
          <option value="4h">4小時</option>
          <option value="1d">1天</option>
        </select>
      </div>
    </div>
    
    <div ref="chartContainer" class="chart-container"></div>
    
    <div v-if="loading" class="chart-loading">
      <div class="flex items-center justify-center p-4">
        <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-500"></div>
        <span class="ml-2 text-sm text-gray-500">載入 {{ selectedInterval }} 數據中...</span>
      </div>
    </div>
    
    <div v-if="error" class="chart-error">
      <div class="text-red-500 text-sm p-2 bg-red-50 rounded">
        <div class="font-semibold">⚠️ 圖表載入失敗</div>
        <div>{{ error }}</div>
        <div class="text-xs mt-1 text-gray-600">
          請嘗試其他時間間隔或稍後再試
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { createChart, ColorType } from 'lightweight-charts'
import type { IChartApi, ISeriesApi, CandlestickData } from 'lightweight-charts'
import type { MetalSymbol } from '../../types'
import { getMetalInfo } from '../../utils/constants'
import { priceApi } from '../../api'

const props = defineProps<{
  symbol: MetalSymbol
}>()

const metalInfo = getMetalInfo(props.symbol)
const selectedInterval = ref('1m')
const interval = ref('1分鐘')
const loading = ref(false)
const error = ref('')
let refreshTimer: number | null = null

// Chart 相關
const chartContainer = ref<HTMLDivElement>()
let chart: IChartApi | null = null
let candlestickSeries: ISeriesApi<'Candlestick'> | null = null

// 初始化圖表
const initChart = () => {
  if (!chartContainer.value) return
  
  chart = createChart(chartContainer.value, {
    width: chartContainer.value.clientWidth,
    height: 400,
    layout: {
      background: { type: ColorType.Solid, color: '#ffffff' },
      textColor: '#333333',
    },
    grid: {
      vertLines: { color: '#f0f0f0' },
      horzLines: { color: '#f0f0f0' },
    },
    crosshair: {
      mode: 1,
    },
    rightPriceScale: {
      borderColor: '#cccccc',
    },
    timeScale: {
      borderColor: '#cccccc',
      timeVisible: true,
      secondsVisible: false,
    },
  })

  candlestickSeries = (chart as any).addCandlestickSeries({
    upColor: '#26a69a',
    downColor: '#ef5350',
    borderDownColor: '#ef5350',
    borderUpColor: '#26a69a',
    wickDownColor: '#ef5350',
    wickUpColor: '#26a69a',
  } as any)
}

// 載入 K 線數據
const loadKlineData = async () => {
  if (!candlestickSeries) {
    console.log(`⚠️ candlestickSeries 未初始化，跳過載入 ${props.symbol} K線數據`)
    return
  }
  
  try {
    loading.value = true
    error.value = ''
    
    console.log(`📊 正在載入 ${props.symbol} K線數據...`)
    console.log(`📊 請求參數:`, {
      symbol: props.symbol,
      interval: selectedInterval.value,
      limit: 100
    })
    
    const response = await priceApi.getKlines({
      symbol: props.symbol,
      interval: selectedInterval.value as any,
      limit: 100
    })
    
    console.log(`📊 API 響應:`, response)
    
    if (response.success && response.data) {
      const klines = response.data.klines
      console.log(`📊 收到 ${klines.length} 筆 K線數據`)
      
      // 處理空數據或填充數據
      if (klines.length === 0) {
        console.warn(`⚠️ ${props.symbol} ${selectedInterval.value} 沒有數據，顯示空圖表`)
        // 不設置錯誤，讓圖表顯示空狀態
        candlestickSeries.setData([])
        return
      }
      
      // 先過濾和修復數據
      const validKlines = klines.filter(() => {
        // 保留所有數據，包括填充的空數據（價格為 0）
        return true
      })
      
      console.log(`📊 過濾後的有效 K線數據: ${validKlines.length} 筆`)
      
      // 檢查是否有數據（包括填充數據）
      if (validKlines.length === 0) {
        console.warn(`⚠️ ${props.symbol} ${selectedInterval.value} 沒有數據`)
        candlestickSeries.setData([])
        return
      }
      
      const chartData: CandlestickData[] = validKlines
        .map(kline => {
          // 確保時間戳是有效的
          const timestamp = kline.timestamp / 1000
          if (isNaN(timestamp) || timestamp <= 0) {
            console.warn(`⚠️ 無效的時間戳: ${kline.timestamp}`)
            return null
          }
          
          // 確保價格數據是有效的數字
          const open = Number(kline.open)
          const high = Number(kline.high)
          const low = Number(kline.low)
          const close = Number(kline.close)
          
          // 修復數據問題
          const validOpen = open === 0 ? close : open
          const validHigh = high === 0 ? Math.max(validOpen, close) : high
          const validLow = low === 0 ? Math.min(validOpen, close) : low
          const validClose = close === 0 ? validOpen : close
          
          // 確保 OHLC 數據的邏輯性
          const finalHigh = Math.max(validHigh, validOpen, validClose)
          const finalLow = Math.min(validLow, validOpen, validClose)
          
          return {
            time: timestamp as any,
            open: validOpen,
            high: finalHigh,
            low: finalLow,
            close: validClose,
          }
        })
        .filter((item): item is CandlestickData => item !== null) // 類型守衛過濾 null 值
      
      console.log(`📊 格式化後的圖表數據:`, chartData.slice(0, 3), '...')
      
      // 檢查圖表數據是否有效
      if (chartData.length === 0) {
        console.warn(`⚠️ ${props.symbol} ${selectedInterval.value} 格式化後沒有數據`)
        candlestickSeries.setData([])
        return
      }
      
      // 金融圖表數據驗證和修復
      const validatedData = validateAndFixChartData(chartData)
      
      if (validatedData.length === 0) {
        console.warn(`⚠️ ${props.symbol} ${selectedInterval.value} 數據驗證失敗，顯示空圖表`)
        candlestickSeries.setData([])
        return
      }
      
      // 檢查數據範圍是否合理
      const prices = validatedData.flatMap(d => [d.open, d.high, d.low, d.close]).filter(p => p > 0)
      
      if (prices.length > 0) {
        const minPrice = Math.min(...prices)
        const maxPrice = Math.max(...prices)
        const priceRange = maxPrice - minPrice
        
        console.log(`📊 價格範圍: ${minPrice.toFixed(2)} - ${maxPrice.toFixed(2)} (範圍: ${priceRange.toFixed(2)})`)
        
        if (priceRange > 10000) {
          console.warn(`⚠️ ${props.symbol} ${selectedInterval.value} 價格波動過大 (${priceRange.toFixed(2)})，可能數據異常`)
        }
      } else {
        console.log(`📊 ${props.symbol} ${selectedInterval.value} 只有填充數據（價格為 0）`)
      }
      
      candlestickSeries.setData(validatedData)
      console.log(`✅ 載入 ${props.symbol} K線數據完成:`, validatedData.length, '筆')
    } else {
      console.error('❌ API 響應失敗:', response)
      // 不設置錯誤，讓圖表顯示空狀態而不是錯誤提示
      candlestickSeries.setData([])
    }
  } catch (err) {
    console.error('❌ 載入 K 線數據失敗:', err)
    error.value = '載入 K 線數據失敗'
  } finally {
    loading.value = false
  }
}

// 處理時間間隔變更
const onIntervalChange = () => {
  const intervalMap: Record<string, string> = {
    '1m': '1分鐘',
    '5m': '5分鐘',
    '15m': '15分鐘',
    '30m': '30分鐘',
    '1h': '1小時',
    '4h': '4小時',
    '1d': '1天'
  }
  interval.value = intervalMap[selectedInterval.value] || '1分鐘'
  
  // 停止舊的自動刷新，重新載入數據，啟動新的自動刷新
  stopAutoRefresh()
  loadKlineData()
  startAutoRefresh()
}

// 響應式調整圖表大小
const resizeChart = () => {
  if (chart && chartContainer.value) {
    chart.applyOptions({
      width: chartContainer.value.clientWidth,
    })
  }
}

// 監聽窗口大小變化
let resizeObserver: ResizeObserver | null = null

onMounted(async () => {
  console.log(`🎯 KlineChart mounted for ${props.symbol}`)
  await nextTick()
  
  try {
    initChart()
    console.log(`📊 圖表初始化完成，開始載入數據...`)
    await loadKlineData()
    startAutoRefresh() // 啟動自動刷新
  } catch (error) {
    console.error(`❌ KlineChart 初始化失敗:`, error)
  }
  
  // 監聽容器大小變化
  if (chartContainer.value) {
    resizeObserver = new ResizeObserver(resizeChart)
    resizeObserver.observe(chartContainer.value)
  }
  
  // 監聽窗口大小變化
  window.addEventListener('resize', resizeChart)
})

onUnmounted(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
  }
  window.removeEventListener('resize', resizeChart)
  
  if (chart) {
    chart.remove()
    chart = null
  }
  
  stopAutoRefresh() // 停止自動刷新
})

// 監聽 symbol 變化
watch(() => props.symbol, () => {
  stopAutoRefresh() // 停止舊的自動刷新
  loadKlineData()
  startAutoRefresh() // 啟動新的自動刷新
})

// 自動刷新功能
const startAutoRefresh = () => {
  stopAutoRefresh() // 確保沒有重複的定時器
  
  // 根據時間間隔設置刷新頻率
  const refreshInterval = getRefreshInterval(selectedInterval.value)
  if (refreshInterval > 0) {
    refreshTimer = setInterval(async () => {
      console.log(`🔄 自動刷新 ${props.symbol} K線數據 (${selectedInterval.value})...`)
      await loadKlineData()
    }, refreshInterval) as unknown as number
    
    console.log(`⏰ 啟動 ${props.symbol} 自動刷新，間隔: ${refreshInterval}ms`)
  }
}

const stopAutoRefresh = () => {
  if (refreshTimer !== null) {
    clearInterval(refreshTimer)
    refreshTimer = null
    console.log(`🛑 停止 ${props.symbol} 自動刷新`)
  }
}

// 根據時間間隔獲取刷新間隔（毫秒）
const getRefreshInterval = (interval: string): number => {
  switch (interval) {
    case '1m': return 60 * 1000 // 1 分鐘刷新
    case '5m': return 5 * 60 * 1000 // 5 分鐘刷新
    case '15m': return 15 * 60 * 1000 // 15 分鐘刷新
    case '30m': return 30 * 60 * 1000 // 30 分鐘刷新
    case '1h': return 60 * 60 * 1000 // 1 小時刷新
    case '4h': return 4 * 60 * 60 * 1000 // 4 小時刷新
    case '1d': return 24 * 60 * 60 * 1000 // 1 天刷新
    default: return 60 * 1000 // 預設 1 分鐘刷新
  }
}

// 金融圖表數據驗證和修復
const validateAndFixChartData = (data: CandlestickData[]): CandlestickData[] => {
  if (!data || data.length === 0) {
    console.warn('📊 沒有數據需要驗證')
    return []
  }
  
  console.log(`📊 開始驗證 ${data.length} 筆數據...`)
  
  // 1. 去重：按時間戳去重，保留最完整的數據
  const deduplicatedData = deduplicateByTimestamp(data)
  console.log(`📊 去重後: ${deduplicatedData.length} 筆`)
  
  // 2. 數據修復：修復不完整的 OHLC 數據
  const fixedData = fixIncompleteOHLC(deduplicatedData)
  console.log(`📊 修復後: ${fixedData.length} 筆`)
  
  // 3. 時間序列驗證：確保時間序列連續性
  const validatedData = validateTimeSeries(fixedData)
  console.log(`📊 驗證後: ${validatedData.length} 筆`)
  
  return validatedData
}

// 按時間戳去重，保留最完整的數據
const deduplicateByTimestamp = (data: CandlestickData[]): CandlestickData[] => {
  const timestampMap = new Map<number, CandlestickData>()
  
  for (const item of data) {
    const timestamp = item.time as number
    const existing = timestampMap.get(timestamp)
    
    if (!existing) {
      timestampMap.set(timestamp, item)
    } else {
      // 比較數據完整性，保留更完整的數據
      const existingScore = getDataCompletenessScore(existing)
      const currentScore = getDataCompletenessScore(item)
      
      if (currentScore > existingScore) {
        timestampMap.set(timestamp, item)
      }
    }
  }
  
  return Array.from(timestampMap.values()).sort((a, b) => (a.time as number) - (b.time as number))
}

// 計算數據完整性分數
const getDataCompletenessScore = (data: CandlestickData): number => {
  let score = 0
  if (data.open > 0) score += 1
  if (data.high > 0) score += 1
  if (data.low > 0) score += 1
  if (data.close > 0) score += 1
  return score
}

// 修復不完整的 OHLC 數據
const fixIncompleteOHLC = (data: CandlestickData[]): CandlestickData[] => {
  return data.map(item => {
    const { open, high, low, close } = item
    
    // 如果所有價格都是 0，保持原樣（填充數據）
    if (open === 0 && high === 0 && low === 0 && close === 0) {
      return item
    }
    
    // 修復 close 為 0 的情況
    const validClose = close === 0 ? open : close
    
    // 修復 high 為 0 的情況
    const validHigh = high === 0 ? Math.max(open, validClose) : high
    
    // 修復 low 為 0 的情況
    const validLow = low === 0 ? Math.min(open, validClose) : low
    
    // 確保 OHLC 邏輯正確
    const finalHigh = Math.max(validHigh, open, validClose)
    const finalLow = Math.min(validLow, open, validClose)
    
    return {
      ...item,
      open,
      high: finalHigh,
      low: finalLow,
      close: validClose
    }
  })
}

// 驗證時間序列連續性
const validateTimeSeries = (data: CandlestickData[]): CandlestickData[] => {
  if (data.length === 0) return data
  
  // 按時間排序
  const sortedData = data.sort((a, b) => (a.time as number) - (b.time as number))
  
  // 檢查時間間隔是否合理
  const validatedData: CandlestickData[] = []
  
  for (let i = 0; i < sortedData.length; i++) {
    const current = sortedData[i]
    if (!current) continue
    
    const currentTime = current.time as number
    
    // 檢查時間戳是否合理（不能是未來時間）
    const now = Date.now() / 1000
    if (currentTime > now) {
      console.warn(`⚠️ 跳過未來時間戳: ${currentTime}`)
      continue
    }
    
    // 檢查與前一個數據的時間間隔
    if (i > 0) {
      const prev = sortedData[i - 1]
      if (prev) {
        const prevTime = prev.time as number
        const timeDiff = currentTime - prevTime
        
        // 如果時間間隔太小（小於 1 秒），跳過
        if (timeDiff < 1) {
          console.warn(`⚠️ 跳過時間間隔太小的數據: ${timeDiff}秒`)
          continue
        }
      }
    }
    
    validatedData.push(current)
  }
  
  return validatedData
}

// 獲取時間間隔的秒數（暫時未使用）
// const getIntervalSeconds = (interval: string): number => {
//   switch (interval) {
//     case '1m': return 60
//     case '5m': return 5 * 60
//     case '15m': return 15 * 60
//     case '30m': return 30 * 60
//     case '1h': return 60 * 60
//     case '4h': return 4 * 60 * 60
//     case '1d': return 24 * 60 * 60
//     default: return 60
//   }
// }
</script>

<style scoped>
.kline-chart-container {
  @apply bg-white rounded-lg shadow-sm border p-4;
}

.chart-header {
  @apply flex items-center justify-between mb-4;
}

.chart-title h3 {
  @apply text-lg font-semibold text-gray-900;
}

.chart-title p {
  @apply text-sm text-gray-500;
}

.chart-controls select {
  @apply px-3 py-1 border border-gray-300 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500;
}

.chart-container {
  @apply w-full h-96 relative;
}

.chart-loading {
  @apply absolute inset-0 flex flex-col items-center justify-center bg-white bg-opacity-75;
}

.chart-error {
  @apply absolute inset-0 flex items-center justify-center bg-red-50;
}
</style>
