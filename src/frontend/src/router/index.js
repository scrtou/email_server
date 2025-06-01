import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// Layout Components
import AuthLayout from '../layouts/AuthLayout.vue'
import AppLayout from '../layouts/AppLayout.vue'

// Page Components
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import Dashboard from '../views/Dashboard.vue'
// import EmailList from '../views/EmailList.vue'; // Removed as it's not found and likely deprecated
import ServiceList from '../views/ServiceList.vue'
import EmailServiceList from '../views/EmailServiceList.vue'
import Profile from '../views/Profile.vue'
import EmailAccountListView from '../views/EmailAccountListView.vue'
import PlatformListView from '../views/PlatformListView.vue'
import PlatformRegistrationListView from '../views/PlatformRegistrationListView.vue'
import ServiceSubscriptionListView from '../views/ServiceSubscriptionListView.vue'
import SearchResultView from '../views/SearchResultView.vue' // Import SearchResultView
import ImportBitwardenView from '../views/ImportBitwardenView.vue' // Import Bitwarden Import View
// import Layout from '../components/Layout.vue' // This was commented out, ensure it's not needed
 
const routes = [
  // Authentication routes with AuthLayout
  {
    path: '/auth',
    component: AuthLayout,
    children: [
      {
        path: 'login', // Relative to /auth, so /auth/login
        name: 'Login',
        component: Login,
        meta: { requiresGuest: true }
      },
      {
        path: 'register', // Relative to /auth, so /auth/register
        name: 'Register',
        component: Register,
        meta: { requiresGuest: true }
      }
    ]
  },
  
  // Main application routes with AppLayout
  {
    path: '/',
    component: AppLayout,
    meta: { requiresAuth: true }, // Apply auth guard to the layout
    children: [
      {
        path: '', // Default child for '/', effectively the dashboard
        name: 'Dashboard',
        component: Dashboard,
        // meta: { requiresAuth: true } // Already covered by parent
      },
      // { // Removed EmailList route
      //   path: 'emails',
      //   name: 'EmailList',
      //   component: EmailList,
      //   // meta: { requiresAuth: true }
      // },
      {
        path: 'services',
        name: 'ServiceList',
        component: ServiceList,
        // meta: { requiresAuth: true }
      },
      {
        path: 'email-services',
        name: 'EmailServiceList',
        component: EmailServiceList,
        // meta: { requiresAuth: true }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: Profile,
        // meta: { requiresAuth: true }
      },
      // Email Account Management Routes
      {
        path: 'email-accounts',
        name: 'EmailAccountList',
        component: EmailAccountListView,
      },
      // { // Commented out as EmailAccount create is now a modal
      //   path: 'email-accounts/create',
      //   name: 'EmailAccountCreate',
      //   component: EmailAccountForm,
      //   props: true // Allows route.params to be passed as props
      // },
      // { // Commented out as EmailAccount edit is now a modal
      //   path: 'email-accounts/edit/:id',
      //   name: 'EmailAccountEdit',
      //   component: EmailAccountForm,
      //   props: true // Allows route.params.id to be passed as prop 'id'
      // },
      // Platform Management Routes
      {
        path: 'platforms',
        name: 'PlatformList',
        component: PlatformListView,
      },
      // { // Commented out as Platform create is now a modal
      //   path: 'platforms/create',
      //   name: 'PlatformCreate',
      //   component: PlatformForm,
      //   props: true
      // },
      // { // Commented out as Platform edit is now a modal
      //   path: 'platforms/edit/:id',
      //   name: 'PlatformEdit',
      //   component: PlatformForm,
      //   props: true
      // },
      // Platform Registration Management Routes
      {
        path: 'platform-registrations',
        name: 'PlatformRegistrationList',
        component: PlatformRegistrationListView,
      },
      // { // Commented out as PlatformRegistration create is now a modal
      //   path: 'platform-registrations/create',
      //   name: 'PlatformRegistrationCreate',
      //   component: PlatformRegistrationForm,
      //   props: true
      // },
      // { // Commented out as PlatformRegistration edit is now a modal
      //   path: 'platform-registrations/edit/:id',
      //   name: 'PlatformRegistrationEdit',
      //   component: PlatformRegistrationForm,
      //   props: true
      // },
      // Service Subscription Management Routes
      {
        path: 'service-subscriptions',
        name: 'ServiceSubscriptionList',
        component: ServiceSubscriptionListView,
      },
      // { // Commented out as ServiceSubscription create is now a modal
      //   path: 'service-subscriptions/create',
      //   name: 'ServiceSubscriptionCreate',
      //   component: ServiceSubscriptionForm,
      //   props: true
      // },
      // { // Commented out as ServiceSubscription edit is now a modal
      //   path: 'service-subscriptions/edit/:id',
      //   name: 'ServiceSubscriptionEdit',
      //   component: ServiceSubscriptionForm,
      //   props: true
      // },
      // Search Results Route
      {
        path: 'search-results',
        name: 'search-results', // Changed from 'SearchResults' to 'search-results' for consistency
        component: SearchResultView,
        // meta: { requiresAuth: true } // Already covered by parent AppLayout
      },
      // Bitwarden Import Route
      {
        path: 'import/bitwarden',
        name: 'ImportBitwarden',
        component: ImportBitwardenView,
        // meta: { requiresAuth: true } // Inherited from AppLayout
      }
      // Add other authenticated routes as children of AppLayout here
    ]
  },
  
  // Redirect /login and /register to /auth/login and /auth/register for old paths
  { path: '/login', redirect: '/auth/login' },
  { path: '/register', redirect: '/auth/register' },

  // 404 Page - redirect to dashboard or a dedicated 404 view
  {
    path: '/:pathMatch(.*)*',
    // name: 'NotFound', // Optional: for a dedicated 404 component
    // component: NotFoundView, // Optional: a dedicated 404 component
    redirect: '/' // Or redirect to a specific 404 page if AppLayout handles it
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
