<template>
  <div id="app">
    <!-- 未登录状态：显示登录/注册页面 -->
    <router-view v-if="!authStore.isAuthenticated" />
    
    <!-- 已登录状态：显示主应用布局 -->
    <el-container v-else style="height: 100vh">
      <!-- 侧边栏 -->
      <el-aside width="200px" style="background-color: #304156">
        <div class="logo-section">
          <h3>邮箱管理系统</h3>
        </div>
        
        <el-menu
          :default-active="$route.path"
          router
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
          class="sidebar-menu"
        >
          <el-menu-item index="/">
            <el-icon><House /></el-icon>
            <span>仪表板</span>
          </el-menu-item>
          <el-menu-item index="/email-accounts">
            <el-icon><Message /></el-icon>
            <span>邮箱账户</span>
          </el-menu-item>
          <el-menu-item index="/platforms">
            <el-icon><Platform /></el-icon> <!-- Assuming Platform icon exists or use a generic one -->
            <span>平台管理</span>
          </el-menu-item>
          <el-menu-item index="/platform-registrations">
            <el-icon><Link /></el-icon>
            <span>平台注册</span>
          </el-menu-item>
          <el-menu-item index="/service-subscriptions">
            <el-icon><Tickets /></el-icon> <!-- Assuming Tickets icon or similar for subscriptions -->
            <span>服务订阅</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主体内容 -->
      <el-container>
        <!-- 头部 -->
        <el-header class="main-header">
          <div class="header-left">
            <!-- <h2>{{ getPageTitle() }}</h2> --> <!-- Replaced by SearchBar or keep if needed alongside -->
            <SearchBar />
          </div>
          
          <div class="header-right">
            <!-- 用户信息下拉菜单 -->
            <el-dropdown @command="handleUserCommand" class="user-dropdown">
              <span class="user-info">
                <el-avatar :size="32" class="user-avatar">
                  {{ authStore.userName.charAt(0).toUpperCase() }}
                </el-avatar>
                <span class="username">{{ authStore.userName }}</span>
                <el-icon class="dropdown-icon"><ArrowDown /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">
                    <el-icon><User /></el-icon>
                    个人信息
                  </el-dropdown-item>
                  <el-dropdown-item command="settings">
                    <el-icon><Setting /></el-icon>
                    系统设置
                  </el-dropdown-item>
                  <el-dropdown-item divided command="logout">
                    <el-icon><SwitchButton /></el-icon>
                    退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>

        <!-- 内容区域 -->
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script>
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessageBox, ElMessage } from 'element-plus'
import SearchBar from '@/components/ui/SearchBar.vue' // Import SearchBar
import {
  House,
  Message,
  // Setting, // Replaced by Platform and Tickets/Service
  Link,
  User,
  ArrowDown,
  SwitchButton,
  Platform, // Added for Platform Management
  Tickets, // Added for Service Subscriptions
  Setting // Keep Setting for user dropdown
} from '@element-plus/icons-vue'

export default {
  name: 'App',
  components: {
    House,
    Message,
    // Setting, // This was for sidebar, keep the import for dropdown
    Link,
    User,
    ArrowDown,
    SwitchButton,
    Platform,
    Tickets,
    SearchBar, // Register SearchBar
    Setting // Register Setting for user dropdown
  },
  setup() {
    const route = useRoute()
    const router = useRouter()
    const authStore = useAuthStore()
    
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
      handleUserCommand
    }
  }
}
</script>

<style>
/* 定义全局颜色变量，以便在各组件中统一使用 */
:root {
  --app-bg-color-light: #f4f6f9; /* 与 Dashboard 一致的背景色 */
  --app-text-color-primary: #2c3e50;
  --app-primary-color: #4A90E2; /* 主题蓝 */
  --app-sidebar-bg: #304156;
  --app-sidebar-text: #bfcbd9;
  --app-sidebar-active-text: #ffffff;
  --app-sidebar-hover-bg: #434a50;
  --app-header-bg: #ffffff;
  --app-header-border: #e6e6e6;
  --app-header-shadow: rgba(0,21,41,.08);
}

#app {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; /* 与 Dashboard 一致的字体 */
  height: 100vh;
  overflow: hidden;
  color: var(--app-text-color-primary);
}

/* 侧边栏样式 */
.el-aside {
  background-color: var(--app-sidebar-bg) !important; /* 确保覆盖 */
}
.logo-section {
  padding: 20px;
  text-align: center;
  border-bottom: 1px solid var(--app-sidebar-hover-bg); /* 使用变量 */
}

.logo-section h3 {
  color: var(--app-sidebar-text);
  margin: 0;
  font-size: 16px;
  font-weight: 600; /* 略微加粗 */
}

.sidebar-menu {
  border: none;
}
.sidebar-menu .el-menu-item {
  border-radius: 0;
  margin: 0;
  color: var(--app-sidebar-text); /* 确保文字颜色应用 */
}
.sidebar-menu .el-menu-item .el-icon {
  color: var(--app-sidebar-text); /* 图标颜色 */
}

.sidebar-menu .el-menu-item:hover {
  background-color: var(--app-sidebar-hover-bg) !important;
}
.sidebar-menu .el-menu-item.is-active {
  background-color: var(--app-primary-color) !important;
  color: var(--app-sidebar-active-text) !important;
}
.sidebar-menu .el-menu-item.is-active .el-icon {
  color: var(--app-sidebar-active-text) !important; /* 激活状态图标颜色 */
}


/* 头部样式 */
.main-header {
  background-color: var(--app-header-bg);
  border-bottom: 1px solid var(--app-header-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 1px 4px var(--app-header-shadow);
}

.header-left {
  flex-grow: 1;
  display: flex;
  align-items: center;
}

.header-left h2 {
  margin: 0;
  margin-right: 20px;
  line-height: 60px; /* el-header 默认高度 */
  color: var(--app-text-color-primary);
  font-size: 18px;
  font-weight: 500;
}

.header-right {
  display: flex;
  align-items: center;
}

/* 用户下拉菜单样式 */
.user-dropdown {
  cursor: pointer;
}

.user-info {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background-color: #f5f7fa; /* 保持一个轻微的悬停效果 */
}

.user-avatar {
  margin-right: 8px;
  background-color: var(--app-primary-color);
  color: white;
  font-weight: bold;
}

.username {
  margin-right: 4px;
  color: var(--app-text-color-primary); /* 使用全局文字颜色 */
  font-size: 14px;
}

.dropdown-icon {
  color: #C0C4CC; /* Element Plus 默认图标颜色 */
  font-size: 12px;
  transition: transform 0.3s;
}

.user-dropdown:hover .dropdown-icon {
  transform: rotate(180deg);
}

/* 主内容区域样式 */
.main-content {
  background-color: var(--app-bg-color-light); /* 使用变量 */
  overflow-y: auto;
  padding: 20px; /* 与 Dashboard 的 padding 保持一致或按需调整 */
}

/* 下拉菜单项样式 */
.el-dropdown-menu__item {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  font-size: 14px; /* 统一字体大小 */
}

.el-dropdown-menu__item .el-icon {
  margin-right: 8px;
  color: #909399; /* Element Plus 默认图标颜色 */
}

.el-dropdown-menu__item:hover {
  background-color: #ecf5ff; /* Element Plus 风格的悬停 */
  color: var(--app-primary-color);
}
.el-dropdown-menu__item:hover .el-icon {
  color: var(--app-primary-color);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .el-aside {
    width: 64px !important; /* Element Plus 默认折叠宽度 */
  }
  
  .logo-section {
    padding: 10px;
  }
  
  .logo-section h3 {
    display: none;
  }
  
  .sidebar-menu .el-menu-item span {
    display: none;
  }
  .sidebar-menu .el-menu-item .el-icon {
    margin-left: 0; /* 调整图标位置 */
  }
  
  .header-left h2 {
    font-size: 16px;
  }
  
  .username {
    display: none;
  }
  .main-content {
    padding: 16px; /* 移动端减小内边距 */
  }
}

/* 滚动条样式 (保持不变或按需调整) */
.main-content::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
.main-content::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}
.main-content::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}
.main-content::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 全局加载状态样式 (如果需要一个非常通用的加载样式) */
/* 注意：Dashboard.vue 中有其局部的 loading-container 样式，会优先应用 */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh; /* 或具体父容器高度 */
  background-color: var(--app-bg-color-light); /* 使用统一背景色 */
}
.loading-container .el-loading-text {
  color: var(--app-primary-color) !important;
}
</style>