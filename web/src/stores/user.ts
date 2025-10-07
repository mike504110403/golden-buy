import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '../types'
import { userApi } from '../api'

export const useUserStore = defineStore('user', () => {
  // State
  const user = ref<User | null>(null)
  const loading = ref(false)

  // Actions
  const fetchUserInfo = async () => {
    try {
      loading.value = true
      const response = await userApi.getUserInfo()
      if (response.success && response.data) {
        user.value = response.data
      }
    } catch (error) {
      console.error('獲取用戶資訊失敗:', error)
    } finally {
      loading.value = false
    }
  }

  const logout = () => {
    user.value = null
  }

  return {
    user,
    loading,
    fetchUserInfo,
    logout
  }
})

