import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import router from '@/router'

// 根据环境判断API基础URL
const getBaseURL = () => {
  if (process.env.NODE_ENV === 'development') {
    return 'http://localhost:5555/api/v1'
  }
  return 'http://localhost:5555/api/v1'
}

const API_BASE_URL = getBaseURL()

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 - 添加认证token
api.interceptors.request.use(
  config => {
    const authStore = useAuthStore()
    
    // 添加认证token
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    
    console.log('API请求:', config.method?.toUpperCase(), config.url)
    return config
  },
  error => {
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理认证错误
api.interceptors.response.use(
  response => {
    console.log('API响应:', response.config.url, response.data)
    
    if (response.data.code === 200) {
      return response.data.data
    } else {
      throw new Error(response.data.message || '请求失败')
    }
  },
  error => {
    console.error('API请求错误:', error)
    
    if (error.response) {
      const status = error.response.status
      const message = error.response.data?.message || `请求失败 (${status})`
      
      // 处理认证错误
      if (status === 401) {
        const authStore = useAuthStore()
        authStore.logout()
        router.push('/login')
        ElMessage.error('登录已过期，请重新登录')
        return Promise.reject(error)
      }
      
      throw new Error(message)
    } else if (error.request) {
      throw new Error('网络连接失败，请检查服务器是否运行')
    } else {
      throw new Error(error.message || '请求配置错误')
    }
  }
)

// 认证相关API
export const authAPI = {
  login: (data) => api.post('/auth/login', data),
  register: (data) => api.post('/auth/register', data),
  logout: () => api.post('/user/logout'),
  getProfile: () => api.get('/user/profile'),
  updateProfile: (data) => api.put('/user/profile', data),
  changePassword: (data) => api.post('/user/change-password', data),
  refreshToken: () => api.post('/auth/refresh')
}

// 原有的API方法
export const emailAPI = {
  getEmails: (params = {}) => api.get('/emails', { params }),
  createEmail: (data) => api.post('/emails', data),
  updateEmail: (id, data) => api.put(`/emails/${id}`, data),
  deleteEmail: (id) => api.delete(`/emails/${id}`),
  getEmailById: (id) => api.get(`/emails/${id}`),
  getEmailServices: (id) => api.get(`/emails/${id}/services`)
}

export const serviceAPI = {
  getServices: (params = {}) => api.get('/services', { params }),
  createService: (data) => api.post('/services', data),
  updateService: (id, data) => api.put(`/services/${id}`, data),
  deleteService: (id) => api.delete(`/services/${id}`),
  getServiceById: (id) => api.get(`/services/${id}`),
  getServiceEmails: (id) => api.get(`/services/${id}/emails`)
}

export const emailServiceAPI = {
  getAllEmailServices: (params = {}) => api.get('/email-services', { params }),
  createEmailService: (data) => api.post('/email-services', data),
  updateEmailService: (id, data) => api.put(`/email-services/${id}`, data),
  deleteEmailService: (id) => api.delete(`/email-services/${id}`)
}

export const dashboardAPI = {
  getDashboard: () => api.get('/dashboard')
}

export default api
