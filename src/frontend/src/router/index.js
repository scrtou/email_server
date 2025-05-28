import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// 页面组件
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import Dashboard from '../views/Dashboard.vue'
import EmailList from '../views/EmailList.vue'
import ServiceList from '../views/ServiceList.vue'
import EmailServiceList from '../views/EmailServiceList.vue'
import Profile from '../views/Profile.vue'
// 删除这行：import Layout from '../components/Layout.vue'

const routes = [
  // 认证相关路由（不需要布局）
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresGuest: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: Register,
    meta: { requiresGuest: true }
  },
  
  // 主应用路由 - 直接使用组件，不嵌套Layout
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard,
    meta: { requiresAuth: true }
  },
  {
    path: '/emails',
    name: 'EmailList',
    component: EmailList,
    meta: { requiresAuth: true }
  },
  {
    path: '/services',
    name: 'ServiceList',
    component: ServiceList,
    meta: { requiresAuth: true }
  },
  {
    path: '/email-services',
    name: 'EmailServiceList',
    component: EmailServiceList,
    meta: { requiresAuth: true }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: Profile,
    meta: { requiresAuth: true }
  },
  
  // 404页面
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  // 检查是否需要认证
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
    return
  }
  
  // 检查是否需要游客状态（已登录用户不能访问登录页面）
  if (to.meta.requiresGuest && authStore.isAuthenticated) {
    next('/')
    return
  }
  
  next()
})

export default router
