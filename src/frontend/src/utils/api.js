import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import router from '@/router'

// 根据环境判断API基础URL
const getBaseURL = () => {
  // Use VUE_APP_API_BASE_URL environment variable
  // Fallback to localhost for development if not set
  if (process.env.VUE_APP_API_BASE_URL) {
    return process.env.VUE_APP_API_BASE_URL;
  }
  // In development, with proxy, use a relative path that the proxy will catch.
  // The proxy is configured for '/api/v1'.
  if (process.env.NODE_ENV === 'development') {
    return '/api/v1'; // Proxy will forward this to target e.g. http://localhost:5555/api/v1
  }
  // Fallback for production if VUE_APP_API_BASE_URL is not set (should be configured in deployment)
  return '/api/v1'; // Example: relative path for production if served on same domain
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
    
    if (response.data.code === 200 || response.status === 201 || response.status === 200) { // Check for HTTP success too
      // If the backend response includes a meta object (typical for pagination)
      // return an object containing both data and meta.
      // Otherwise, just return the data.
      if (response.data.meta) {
        return { data: response.data.data, meta: response.data.meta };
      }
      // If 'reminders' exists directly in response.data (like from /users/me/reminders), return response.data
      if (response.data.reminders !== undefined) {
        return response.data;
      }
      // For non-paginated or create/update/delete successful responses that might only have data
      return response.data.data !== undefined ? response.data.data : {}; // Ensure data field exists or return empty obj
    } else {
      // Handle business logic errors (e.g., validation errors sent with a 200/2xx status but a non-200 business code)
      const message = response.data.message || '操作失败，未知错误';
      ElMessage.error(message);
      return Promise.reject(new Error(message));
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
  logout: () => api.post('/auth/logout'), // Updated path
  getProfile: () => api.get('/users/me'), // Updated path
  updateProfile: (data) => api.put('/users/me', data), // Updated path for consistency with proposal
  changePassword: (data) => api.post('/users/me/change-password', data), // Updated path for consistency
  refreshToken: () => api.post('/auth/refresh'),
  getReminders: () => api.get('/users/me/reminders'),
  markReminderAsRead: (id) => api.put(`/users/me/reminders/${id}/read`)
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
  getDashboard: () => api.get('/dashboard'), // This points to the old, now deprecated endpoint
  getSummary: () => api.get('/dashboard/summary') // New endpoint for dashboard summary
}

// EmailAccount API (New as per plan)
export const emailAccountAPI = {
  getAll: (params = {}) => api.get('/email-accounts', { params }),
  getById: (id) => api.get(`/email-accounts/${id}`),
  create: (data) => api.post('/email-accounts', data),
  update: (id, data) => api.put(`/email-accounts/${id}`, data),
  delete: (id) => api.delete(`/email-accounts/${id}`),
  getAssociatedPlatformRegistrations: (emailAccountId, params = {}) => api.get(`/email-accounts/${emailAccountId}/platform-registrations`, { params }),
  getUniqueProviders: () => api.get('/email-accounts/providers'), // Added for provider filter dropdown
};

// Platform API
export const platformAPI = {
  getAll: (params = {}) => api.get('/platforms', { params }),
  getById: (id) => api.get(`/platforms/${id}`),
  create: (data) => api.post('/platforms', data),
  update: (id, data) => api.put(`/platforms/${id}`, data),
  delete: (id) => api.delete(`/platforms/${id}`),
  getAssociatedEmailRegistrations: (platformId, params = {}) => api.get(`/platforms/${platformId}/email-registrations`, { params }),
};

// PlatformRegistration API
export const platformRegistrationAPI = {
  getAll: (params = {}) => api.get('/platform-registrations', { params }),
  getById: (id) => api.get(`/platform-registrations/${id}`),
  create: (data) => api.post('/platform-registrations', data), // For creating with IDs
  createByName: (data) => api.post('/platform-registrations/by-name', data), // For creating with names
  update: (id, data) => api.put(`/platform-registrations/${id}`, data),
  delete: (id) => api.delete(`/platform-registrations/${id}`),
  getAssociatedServiceSubscriptions: (registrationId, params = {}) => api.get(`/platform-registrations/${registrationId}/service-subscriptions`, { params }),
};

// ServiceSubscription API
export const serviceSubscriptionAPI = {
  getAll: (params = {}) => api.get('/service-subscriptions', { params }),
  getById: (id) => api.get(`/service-subscriptions/${id}`),
  create: (data) => api.post('/service-subscriptions', data),
  update: (id, data) => api.put(`/service-subscriptions/${id}`, data),
  delete: (id) => api.delete(`/service-subscriptions/${id}`),
};

// Search API
export const searchAPI = {
  search: (params = {}) => api.get('/search', { params }),
};
 
export default api
