<template>
  <div class="card hover:shadow-xl transition-shadow duration-300">
    <!-- 標題 -->
    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center space-x-2">
        <span class="text-3xl">{{ metalInfo.icon }}</span>
        <div>
          <h3 class="font-bold text-gray-900">{{ metalInfo.nameCn }}</h3>
          <p class="text-sm text-gray-500">{{ metalInfo.name }}</p>
        </div>
      </div>
    </div>

    <!-- 價格 -->
    <div v-if="price" class="space-y-2">
      <div class="font-mono-number text-3xl font-bold text-gray-900">
        ${{ formatPrice(price.price) }}
      </div>
      
      <!-- 漲跌幅 -->
      <div 
        :class="['flex items-center space-x-1', getPriceChangeClass(price.change_percent)]"
      >
        <span class="text-xl">{{ getPriceChangeIcon(price.change_percent) }}</span>
        <span class="font-mono-number font-medium text-lg">
          {{ formatPercent(price.change_percent) }}
        </span>
      </div>

      <!-- 更新時間 -->
      <div class="text-xs text-gray-400">
        {{ formatRelativeTimeReactive(price.timestamp) }}
      </div>
      
      <!-- 調試信息 -->
      <div class="text-xs text-blue-500" v-if="price">
        調試: {{ price.symbol }} = ${{ price.price }} ({{ price.change_percent }}%)
      </div>
    </div>

    <!-- 載入中 -->
    <div v-else class="space-y-2">
      <div class="h-10 bg-gray-200 rounded animate-pulse"></div>
      <div class="h-6 bg-gray-200 rounded animate-pulse w-1/2"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import type { MetalSymbol } from '../../types'
import { usePriceStore } from '../../stores/price'
import { getMetalInfo } from '../../utils/constants'
import { 
  formatPrice, 
  formatPercent, 
  formatRelativeTimeReactive,
  getPriceChangeClass,
  getPriceChangeIcon 
} from '../../utils/format'

const props = defineProps<{
  symbol: MetalSymbol
}>()

const priceStore = usePriceStore()

const metalInfo = computed(() => getMetalInfo(props.symbol))
const price = computed(() => {
  const p = priceStore.getPrice(props.symbol)
  console.log('🎯 PriceCard price computed:', props.symbol, p)
  return p
})

// 自動計數的秒數
const currentTime = ref(Date.now())

onMounted(() => {
  // 每秒更新時間，觸發響應式更新
  const timer = setInterval(() => {
    currentTime.value = Date.now()
  }, 1000)
  
  onUnmounted(() => {
    clearInterval(timer)
  })
})

// 強制響應式更新
const forceUpdate = ref(0)
setInterval(() => {
  forceUpdate.value++
}, 1000)
</script>

