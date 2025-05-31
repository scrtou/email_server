import { defineStore } from 'pinia'
import { authAPI } from '@/utils/api'
import { ElMessage } from 'element-plus'
import { useNotificationStore } from './notification' // 导入 notification store

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    token: localStorage.getItem('token') || null,
    isLoading: false
  }),
  
  getters: {
    isAuthenticated: (state) => !!state.token,
    userName: (state) => state.user?.username || '用户',
    userRole: (state) => state.user?.role || 'user'
  },
  
  actions: {
    // 登录
    async login(credentials) {
      this.isLoading = true
      try {
        const response = await authAPI.login(credentials)
        
        this.token = response.token
        this.user = response.user
        
        // 保存token到localStorage
        localStorage.setItem('token', response.token)
        
        ElMessage.success('登录成功')
        return true
      } catch (error) {
        ElMessage.error(error.message || '登录失败')
        return false
      } finally {
        this.isLoading = false
      }
    },
    
    // 注册
    async register(userData) {
      this.isLoading = true
      try {
        const response = await authAPI.register(userData)
        
        this.token = response.token
        this.user = response.user
        
        localStorage.setItem('token', response.token)
        
        ElMessage.success('注册成功')
        return true
      } catch (error) {
        ElMessage.error(error.message || '注册失败')
        return false
      } finally {
        this.isLoading = false
      }
    },
    
    // 登出
    async logout() {
      try {
        if (this.token) {
          await authAPI.logout()
        }
      } catch (error) {
        console.error('登出请求失败:', error)
      } finally {
        this.user = null
        this.token = null
        localStorage.removeItem('token')

        // 重置 remindersLoaded 状态
        const notificationStore = useNotificationStore()
        notificationStore.resetRemindersLoaded()
        
        ElMessage.success('已退出登录')
      }
    },
    
    // 获取用户信息
    async fetchUserProfile() {
      if (!this.token) return false
      
      try {
        console.log('[AuthStore] Attempting to fetch user profile. Token:', this.token);
        this.user = await authAPI.getProfile()
        console.log('[AuthStore] Successfully fetched user profile:', this.user);
        return true
      } catch (error) {
        console.error('[AuthStore] Failed to fetch user profile. Error:', error, 'Token during failure:', this.token);
        console.log('[AuthStore] Logging out due to fetchUserProfile failure.');
        this.logout()
        return false
      }
    },
    
    // 更新用户信息
    async updateProfile(userData) {
      try {
        await authAPI.updateProfile(userData)
        await this.fetchUserProfile()
        ElMessage.success('个人信息更新成功')
        return true
      } catch (error) {
        ElMessage.error(error.message || '更新失败')
        return false
      }
    },
    
    // 修改密码
    async changePassword(passwordData) {
      try {
        await authAPI.changePassword(passwordData)
        ElMessage.success('密码修改成功')
        return true
      } catch (error) {
        ElMessage.error(error.message || '密码修改失败')
        return false
      }
    },
    
    // 初始化认证状态
    async initAuth() {
      if (this.token) {
        await this.fetchUserProfile()
      }
    }
  }
})
