<template>
  <div id="app">
    <!-- 未登录状态：显示登录/注册页面 -->
    <router-view v-if="!authStore.isAuthenticated" />
    
    <!-- 已登录状态：显示主应用布局 -->
    <el-container v-else style="height: 100vh">
      <!-- 侧边栏 -->
      <el-aside
        width="160px"
        style="background-color: #304156"
        role="navigation"
        aria-label="主导航菜单"
        class="sidebar-container"
      >
        <div class="logo-section">
          <h3>账号管理系统</h3>
        </div>

        <div class="sidebar-content">
          <el-menu
            :default-active="$route.path"
            router
            background-color="#304156"
            text-color="#bfcbd9"
            active-text-color="#409EFF"
            class="sidebar-menu"
            role="menubar"
            aria-label="主导航"
          >
            <el-menu-item
              index="/"
              role="menuitem"
              aria-label="仪表板页面"
            >
              <el-icon aria-hidden="true"><House /></el-icon>
              <span>仪表板</span>
            </el-menu-item>
            <el-menu-item
              index="/email-accounts"
              role="menuitem"
              aria-label="邮箱账户管理页面"
            >
              <el-icon aria-hidden="true"><Message /></el-icon>
              <span>邮箱账户</span>
            </el-menu-item>
            <el-menu-item
              index="/platforms"
              role="menuitem"
              aria-label="平台管理页面"
            >
              <el-icon aria-hidden="true"><Platform /></el-icon>
              <span>平台管理</span>
            </el-menu-item>
            <el-menu-item
              index="/platform-registrations"
              role="menuitem"
              aria-label="平台注册管理页面"
            >
              <el-icon aria-hidden="true"><Link /></el-icon>
              <span>平台注册</span>
            </el-menu-item>
            <el-menu-item
              index="/service-subscriptions"
              role="menuitem"
              aria-label="服务订阅管理页面"
            >
              <el-icon aria-hidden="true"><Tickets /></el-icon>
              <span>服务订阅</span>
            </el-menu-item>
          </el-menu>
        </div>

        <!-- 底部用户信息区域 -->
        <div class="sidebar-footer">
          <div class="user-info-section">
            <el-dropdown @command="handleUserCommand" trigger="click" placement="top-start">
              <div class="user-avatar-section" style="cursor: pointer;">
                <el-avatar
                  :size="36"
                  class="user-avatar"
                  :alt="`${authStore.userName} 的头像`"
                >
                  {{ authStore.userName.charAt(0).toUpperCase() }}
                </el-avatar>
                <div class="user-details">
                  <div class="username">{{ authStore.userName }}</div>
                  <div class="user-role">管理员</div>
                </div>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">
                    <el-icon><User /></el-icon>个人信息
                  </el-dropdown-item>
                  <el-dropdown-item command="settings">
                    <el-icon><Setting /></el-icon>系统设置
                  </el-dropdown-item>
                  <el-dropdown-item command="logout" divided>
                    <el-icon><SwitchButton /></el-icon>退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </el-aside>

      <!-- 主体内容 -->
      <el-main
        class="main-content"
        role="main"
        aria-label="主要内容区域"
      >
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<script>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessageBox, ElMessage } from 'element-plus'
import {
  House,
  Message,
  Link,
  User,
  SwitchButton,
  Platform,
  Tickets,
  Setting
} from '@element-plus/icons-vue'

export default {
  name: 'App',
  components: {
    House,
    Message,
    Link,
    User,
    SwitchButton,
    Platform,
    Tickets,
    Setting
  },
  setup() {
    const route = useRoute()
    const router = useRouter()
    const authStore = useAuthStore()
    const dropdownVisible = ref(false) // 用于控制下拉菜单的可见性，如果需要手动控制

    // 页面标题映射
    const pageTitles = {
      '/': '系统概览',
      '/email-accounts': '邮箱账户管理',
      '/platforms': '平台管理',
      '/platform-registrations': '平台注册信息管理',
      '/service-subscriptions': '服务订阅管理',
      '/profile': '个人信息'
    }
    
    const getPageTitle = () => {
      return pageTitles[route.path] || '邮箱服务管理系统'
    }
    
    const handleUserCommand = async (command) => {
      switch (command) {
        case 'profile':
          router.push('/profile')
          break
        case 'settings':
          ElMessage.info('系统设置功能开发中...')
          break
        case 'logout':
          try {
            await ElMessageBox.confirm(
              '确定要退出登录吗？',
              '确认退出',
              {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
                type: 'warning'
              }
            )
            await authStore.logout()
            router.push('/login')
          } catch (error) {
            // 用户取消操作
          }
          break
      }
    }
    
    // 初始化时检查认证状态
    onMounted(async () => {
      if (authStore.token && !authStore.user) {
        await authStore.fetchUserProfile()
      }
    })
    
    return {
      authStore,
      getPageTitle,
      handleUserCommand,
      dropdownVisible
    }
  }
}
</script>

<style>
/* Enhanced Design System - Global CSS Custom Properties */
:root {
  /* === Color Palette === */
  /* Primary Colors */
  --color-primary-50: #eff6ff;
  --color-primary-100: #dbeafe;
  --color-primary-200: #bfdbfe;
  --color-primary-300: #93c5fd;
  --color-primary-400: #60a5fa;
  --color-primary-500: #3b82f6;
  --color-primary-600: #2563eb;
  --color-primary-700: #1d4ed8;
  --color-primary-800: #1e40af;
  --color-primary-900: #1e3a8a;

  /* Semantic Colors */
  --color-success-50: #ecfdf5;
  --color-success-500: #10b981;
  --color-success-600: #059669;
  --color-warning-50: #fffbeb;
  --color-warning-500: #f59e0b;
  --color-warning-600: #d97706;
  --color-error-50: #fef2f2;
  --color-error-500: #ef4444;
  --color-error-600: #dc2626;
  --color-info-50: #f0f9ff;
  --color-info-500: #06b6d4;
  --color-info-600: #0891b2;

  /* Neutral Colors */
  --color-gray-50: #f9fafb;
  --color-gray-100: #f3f4f6;
  --color-gray-200: #e5e7eb;
  --color-gray-300: #d1d5db;
  --color-gray-400: #9ca3af;
  --color-gray-500: #6b7280;
  --color-gray-600: #4b5563;
  --color-gray-700: #374151;
  --color-gray-800: #1f2937;
  --color-gray-900: #111827;

  /* Application Colors (Legacy compatibility) */
  --app-bg-color-light: var(--color-gray-50);
  --app-text-color-primary: var(--color-gray-800);
  --app-primary-color: var(--color-primary-600);
  --app-sidebar-bg: var(--color-gray-800);
  --app-sidebar-text: var(--color-gray-300);
  --app-sidebar-active-text: #ffffff;
  --app-sidebar-hover-bg: var(--color-gray-700);
  --app-header-bg: #ffffff;
  --app-header-border: var(--color-gray-200);
  --app-header-shadow: rgba(0, 0, 0, 0.05);

  /* === Typography === */
  --font-family-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  --font-family-mono: "SF Mono", Monaco, "Cascadia Code", "Roboto Mono", Consolas, "Courier New", monospace;

  /* Font Sizes */
  --text-xs: 0.75rem;    /* 12px */
  --text-sm: 0.875rem;   /* 14px */
  --text-base: 1rem;     /* 16px */
  --text-lg: 1.125rem;   /* 18px */
  --text-xl: 1.25rem;    /* 20px */
  --text-2xl: 1.5rem;    /* 24px */
  --text-3xl: 1.875rem;  /* 30px */
  --text-4xl: 2.25rem;   /* 36px */

  /* Font Weights */
  --font-light: 300;
  --font-normal: 400;
  --font-medium: 500;
  --font-semibold: 600;
  --font-bold: 700;

  /* Line Heights */
  --leading-tight: 1.25;
  --leading-normal: 1.5;
  --leading-relaxed: 1.625;

  /* === Spacing === */
  --space-1: 0.25rem;   /* 4px */
  --space-2: 0.5rem;    /* 8px */
  --space-3: 0.75rem;   /* 12px */
  --space-4: 1rem;      /* 16px */
  --space-5: 1.25rem;   /* 20px */
  --space-6: 1.5rem;    /* 24px */
  --space-8: 2rem;      /* 32px */
  --space-10: 2.5rem;   /* 40px */
  --space-12: 3rem;     /* 48px */
  --space-16: 4rem;     /* 64px */

  /* === Shadows === */
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-base: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  --shadow-xl: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);

  /* === Border Radius === */
  --radius-sm: 0.25rem;  /* 4px */
  --radius-base: 0.375rem; /* 6px */
  --radius-md: 0.5rem;   /* 8px */
  --radius-lg: 0.75rem;  /* 12px */
  --radius-xl: 1rem;     /* 16px */
  --radius-full: 9999px;

  /* === Transitions === */
  --transition-fast: 150ms ease-in-out;
  --transition-base: 250ms ease-in-out;
  --transition-slow: 350ms ease-in-out;

  /* === Z-Index === */
  --z-dropdown: 1000;
  --z-sticky: 1020;
  --z-fixed: 1030;
  --z-modal: 1040;
  --z-popover: 1050;
  --z-tooltip: 1060;
}

/* === Global Base Styles === */
* {
  box-sizing: border-box;
}

html {
  font-size: 16px;
  line-height: var(--leading-normal);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-rendering: optimizeLegibility;
}

body {
  margin: 0;
  padding: 0;
  font-family: var(--font-family-sans);
  color: var(--app-text-color-primary);
  background-color: var(--app-bg-color-light);
}

#app {
  font-family: var(--font-family-sans);
  height: 100vh;
  overflow: hidden;
  color: var(--app-text-color-primary);
}

/* === Utility Classes === */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}

.transition-all {
  transition: all var(--transition-base);
}

.transition-colors {
  transition: color var(--transition-base), background-color var(--transition-base), border-color var(--transition-base);
}

.transition-transform {
  transition: transform var(--transition-base);
}

/* === Enhanced Sidebar Styles === */
.sidebar-container {
  background: linear-gradient(180deg, var(--app-sidebar-bg) 0%, rgba(31, 41, 55, 0.95) 100%) !important;
  box-shadow: var(--shadow-lg);
  border-right: 1px solid var(--color-gray-700);
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.logo-section {
  padding: var(--space-6);
  text-align: center;
  border-bottom: 1px solid var(--color-gray-700);
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  flex-shrink: 0;
}

.logo-section h3 {
  color: var(--app-sidebar-active-text);
  margin: 0;
  font-size: var(--text-base);
  font-weight: var(--font-bold);
  letter-spacing: 0.025em;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
}

.sidebar-menu {
  border: none;
  background: transparent !important;
  padding: var(--space-2) 0;
}

.sidebar-menu .el-menu-item {
  border-radius: var(--radius-md);
  margin: var(--space-1) var(--space-3);
  color: var(--app-sidebar-text);
  font-weight: var(--font-medium);
  transition: all var(--transition-base);
  position: relative;
  overflow: hidden;
}

.sidebar-menu .el-menu-item::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(90deg, var(--color-primary-500), var(--color-primary-600));
  opacity: 0;
  transition: opacity var(--transition-base);
  z-index: -1;
}

.sidebar-menu .el-menu-item .el-icon {
  color: var(--app-sidebar-text);
  transition: all var(--transition-base);
  margin-right: var(--space-3);
}

.sidebar-menu .el-menu-item:hover {
  background-color: var(--app-sidebar-hover-bg) !important;
  color: var(--color-gray-100) !important;
  transform: translateX(4px);
  box-shadow: var(--shadow-md);
}

.sidebar-menu .el-menu-item:hover .el-icon {
  color: var(--color-gray-100) !important;
  transform: scale(1.1);
}

.sidebar-menu .el-menu-item.is-active {
  background: linear-gradient(90deg, var(--color-primary-600), var(--color-primary-500)) !important;
  color: var(--app-sidebar-active-text) !important;
  box-shadow: var(--shadow-lg);
  transform: translateX(4px);
}

.sidebar-menu .el-menu-item.is-active::before {
  opacity: 1;
}

.sidebar-menu .el-menu-item.is-active .el-icon {
  color: var(--app-sidebar-active-text) !important;
  transform: scale(1.1);
}

/* === Sidebar Footer Styles === */
.sidebar-footer {
  flex-shrink: 0;
  border-top: 1px solid var(--color-gray-700);
  background: rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
}

.user-info-section {
  padding: var(--space-4);
}

.user-avatar-section {
  display: flex;
  align-items: center;
  margin-bottom: var(--space-3);
}

.user-avatar-section .user-avatar {
  margin-right: var(--space-3);
  background: linear-gradient(135deg, var(--color-primary-500), var(--color-primary-600));
  color: white;
  font-weight: var(--font-semibold);
  box-shadow: var(--shadow-sm);
  border: 2px solid rgba(255, 255, 255, 0.2);
}

.user-details {
  flex: 1;
  min-width: 0;
}

.user-details .username {
  color: var(--app-sidebar-active-text);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  margin-bottom: var(--space-1);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-details .user-role {
  color: var(--app-sidebar-text);
  font-size: var(--text-xs);
  opacity: 0.8;
}

.user-actions {
  display: flex;
  justify-content: space-between;
  gap: var(--space-1);
}

.user-action-btn {
  color: var(--app-sidebar-text) !important;
  border: none !important;
  background: transparent !important;
  padding: var(--space-2) !important;
  border-radius: var(--radius-md) !important;
  transition: all var(--transition-base) !important;
  min-width: auto !important;
  height: auto !important;
}

.user-action-btn:hover {
  color: var(--app-sidebar-active-text) !important;
  background: rgba(255, 255, 255, 0.1) !important;
  transform: scale(1.1);
}

.user-action-btn.logout-btn:hover {
  color: var(--color-error-400) !important;
  background: rgba(239, 68, 68, 0.1) !important;
}




/* === Enhanced Main Content Area === */
.main-content {
  background: linear-gradient(135deg, var(--app-bg-color-light) 0%, rgba(249, 250, 251, 0.8) 100%);
  /* overflow-y: auto; */ /* Removed */
  display: flex; /* Added */
  flex-direction: column; /* Added */
  overflow: hidden; /* Added to prevent main content from scrolling */
  padding: 0; /* Keep padding on child if needed, or manage globally */
  position: relative;
}

.main-content::before {
  content: '';
  position: fixed;
  top: 0;
  left: 160px;
  right: 0;
  bottom: 0;
  background:
    radial-gradient(circle at 20% 20%, rgba(59, 130, 246, 0.05) 0%, transparent 50%),
    radial-gradient(circle at 80% 80%, rgba(16, 185, 129, 0.05) 0%, transparent 50%);
  pointer-events: none;
  z-index: -1;
}



/* === Enhanced Responsive Design === */
@media (max-width: 768px) {
  .sidebar-container {
    width: 64px !important;
    box-shadow: var(--shadow-xl);
  }

  .logo-section {
    padding: var(--space-3);
  }

  .logo-section h3 {
    display: none;
  }

  .sidebar-menu .el-menu-item {
    margin: var(--space-1) var(--space-2);
    justify-content: center;
  }

  .sidebar-menu .el-menu-item span {
    display: none;
  }

  .sidebar-menu .el-menu-item .el-icon {
    margin-right: 0;
    margin-left: 0;
  }

  .main-content::before {
    left: 64px;
  }

  /* 移动端隐藏用户详细信息，只显示头像和操作按钮 */
  .user-details {
    display: none;
  }

  .user-avatar-section {
    justify-content: center;
    margin-bottom: var(--space-2);
  }

  .user-actions {
    justify-content: center;
    gap: var(--space-2);
  }

  .user-action-btn {
    padding: var(--space-1) !important;
  }
}

/* === Enhanced Scrollbar Styles === */
.main-content::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.main-content::-webkit-scrollbar-track {
  background: var(--color-gray-100);
  border-radius: var(--radius-full);
}

.main-content::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, var(--color-gray-300), var(--color-gray-400));
  border-radius: var(--radius-full);
  border: 2px solid var(--color-gray-100);
}

.main-content::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(180deg, var(--color-gray-400), var(--color-gray-500));
}

.main-content::-webkit-scrollbar-corner {
  background: var(--color-gray-100);
}

/* === Enhanced Loading States === */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, var(--app-bg-color-light) 0%, rgba(249, 250, 251, 0.8) 100%);
  position: relative;
}

.loading-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background:
    radial-gradient(circle at 30% 30%, rgba(59, 130, 246, 0.1) 0%, transparent 50%),
    radial-gradient(circle at 70% 70%, rgba(16, 185, 129, 0.1) 0%, transparent 50%);
  animation: pulse 3s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 0.8; }
}

.loading-container .el-loading-text {
  color: var(--color-primary-600) !important;
  font-weight: var(--font-medium) !important;
  font-size: var(--text-base) !important;
}

/* === Accessibility Enhancements === */
.el-menu-item:focus-visible,
.user-dropdown:focus-visible,
.el-button:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
  border-radius: var(--radius-md);
}

/* High contrast mode support */
@media (prefers-contrast: high) {
  :root {
    --app-sidebar-bg: #000000;
    --app-sidebar-text: #ffffff;
    --app-text-color-primary: #000000;
    --color-primary-600: #0066cc;
  }
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
</style>