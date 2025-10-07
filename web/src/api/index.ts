import axios from 'axios'
import type { AxiosInstance } from 'axios'
import type { ApiResponse, Price, PriceMap, KlineQuery, KlineResponse, User } from '../types'

// 創建 Axios 實例
const api: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 請求攔截器
api.interceptors.request.use(
  (config) => {
    // 可以在這裡添加 token
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 響應攔截器
api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    console.error('API Error:', error)
    return Promise.reject(error)
  }
)

// API 方法
export const priceApi = {
  // 獲取所有當前價格
  getCurrentPrices: () => {
    return api.get<any, ApiResponse<PriceMap>>('/api/prices/current')
  },

  // 獲取單個商品當前價格
  getCurrentPrice: (symbol: string) => {
    return api.get<any, ApiResponse<Price>>(`/api/prices/current?symbol=${symbol}`)
  },

  // 獲取 K 線資料
  getKlines: (params: KlineQuery) => {
    return api.get<any, ApiResponse<KlineResponse>>('/api/prices/history', { params })
  }
}

export const userApi = {
  // 獲取用戶資訊
  getUserInfo: () => {
    return api.get<any, ApiResponse<User>>('/api/user/info')
  }
}

// 健康檢查
export const healthCheck = () => {
  return api.get<any, ApiResponse>('/health')
}

export default api

