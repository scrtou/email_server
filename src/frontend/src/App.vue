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
            <h2>{{ getPageTitle() }}</h2>
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
import {
  House,
  Message,
  // Setting, // Replaced by Platform and Tickets/Service
  Link,
  User,
  ArrowDown,
  SwitchButton,
  Platform, // Added for Platform Management
  Tickets // Added for Service Subscriptions
} from '@element-plus/icons-vue'

export default {
  name: 'App',
  components: {
    House,
    Message,
    // Setting,
    Link,
    User,
    ArrowDown,
    SwitchButton,
    Platform,
    Tickets
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
/* 样式保持不变 */
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  height: 100vh;
  overflow: hidden;
}

/* 侧边栏样式 */
.logo-section {
  padding: 20px;
  text-align: center;
  border-bottom: 1px solid #434a50;
}

.logo-section h3 {
  color: #bfcbd9;
  margin: 0;
  font-size: 16px;
  font-weight: normal;
}

.sidebar-menu {
  border: none;
}

.sidebar-menu .el-menu-item {
  border-radius: 0;
  margin: 0;
}

.sidebar-menu .el-menu-item:hover {
  background-color: #434a50 !important;
}

.sidebar-menu .el-menu-item.is-active {
  background-color: #409EFF !important;
  color: #fff !important;
}

/* 头部样式 */
.main-header {
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 1px 4px rgba(0,21,41,.08);
}

.header-left h2 {
  margin: 0;
  line-height: 60px;
  color: #303133;
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
  transition: all 0.3s;
}

.user-info:hover {
  background-color: #f5f7fa;
}

.user-avatar {
  margin-right: 8px;
  background-color: #409EFF;
  color: white;
  font-weight: bold;
}

.username {
  margin-right: 4px;
  color: #606266;
  font-size: 14px;
}

.dropdown-icon {
  color: #C0C4CC;
  font-size: 12px;
  transition: transform 0.3s;
}

.user-dropdown:hover .dropdown-icon {
  transform: rotate(180deg);
}

/* 主内容区域样式 */
.main-content {
  background-color: #f0f2f5;
  overflow-y: auto;
  padding: 20px;
}

/* 下拉菜单项样式 */
.el-dropdown-menu__item {
  display: flex;
  align-items: center;
  padding: 8px 16px;
}

.el-dropdown-menu__item .el-icon {
  margin-right: 8px;
  color: #909399;
}

.el-dropdown-menu__item:hover .el-icon {
  color: #409EFF;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .el-aside {
    width: 64px !important;
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
  
  .header-left h2 {
    font-size: 16px;
  }
  
  .username {
    display: none;
  }
}

/* 滚动条样式 */
.main-content::-webkit-scrollbar {
  width: 6px;
}

.main-content::-webkit-scrollbar-track {
  background: #f1f1f1;
}

.main-content::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.main-content::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 加载状态样式 */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f0f2f5;
}
</style>