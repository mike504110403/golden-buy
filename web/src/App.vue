<template>
  <div id="app" class="min-h-screen bg-gray-50">
    <!-- 頂部導航欄 -->
    <header class="bg-white shadow-sm border-b">
      <div class="container mx-auto px-4">
        <div class="flex items-center justify-between h-16">
          <!-- Logo -->
          <div class="flex items-center space-x-2">
            <span class="text-3xl">🥇</span>
            <h1 class="text-xl font-bold text-gray-900">Golden Buy</h1>
            <span class="text-sm text-gray-500">貴金屬交易平台</span>
          </div>

          <!-- 導航 -->
          <nav class="hidden md:flex space-x-6">
            <router-link 
              to="/" 
              class="text-gray-700 hover:text-gold-600 transition-colors"
              active-class="text-gold-600 font-medium"
            >
              儀表板
            </router-link>
            <router-link 
              to="/orders" 
              class="text-gray-700 hover:text-gold-600 transition-colors"
              active-class="text-gold-600 font-medium"
            >
              訂單管理
            </router-link>
          </nav>

          <!-- 用戶資訊 -->
          <div class="flex items-center space-x-4">
            <!-- WebSocket 連接狀態 -->
            <div class="flex items-center space-x-2">
              <span 
                :class="[
                  'w-2 h-2 rounded-full',
                  priceStore.wsConnected ? 'bg-green-500' : 'bg-red-500'
                ]"
              ></span>
              <span class="text-sm text-gray-600">
                {{ priceStore.wsConnected ? '已連接' : '未連接' }}
              </span>
            </div>

            <!-- 用戶餘額 -->
            <div v-if="userStore.user" class="text-sm">
              <span class="text-gray-600">餘額：</span>
              <span class="font-medium text-gray-900">
                ${{ formatCurrency(userStore.user.balance) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- 主內容區 -->
    <main class="container mx-auto px-4 py-6">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { usePriceStore } from './stores/price'
import { useUserStore } from './stores/user'
import { formatCurrency } from './utils/format'

const priceStore = usePriceStore()
const userStore = useUserStore()

onMounted(async () => {
  // 初始化數據
  await Promise.all([
    priceStore.fetchCurrentPrices(),
    userStore.fetchUserInfo()
  ])
  
  // 連接 WebSocket
  await priceStore.connectWebSocket()
})

onUnmounted(() => {
  // 斷開 WebSocket
  priceStore.disconnectWebSocket()
})
</script>
