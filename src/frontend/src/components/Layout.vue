<template>
    <div class="layout-container">
      <!-- 顶部导航 -->
      <el-header class="header">
        <div class="header-left">
          <h1>邮箱管理系统</h1>
        </div>
        
        <div class="header-right">
          <el-dropdown @command="handleUserCommand">
            <span class="user-dropdown">
              <el-icon><User /></el-icon>
              {{ authStore.userName }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>
                  个人信息
                </el-dropdown-item>
                <el-dropdown-item command="logout" divided>
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
  
      <el-container>
        <!-- 侧边栏 -->
        <el-aside width="200px" class="sidebar">
          <el-menu
            :default-active="$route.path"
            class="sidebar-menu"
            router
          >
            <el-menu-item index="/">
              <el-icon><Odometer /></el-icon>
              <span>概览</span>
            </el-menu-item>
            
            <el-menu-item index="/emails">
              <el-icon><Message /></el-icon>
              <span>邮箱管理</span>
            </el-menu-item>
            
            <el-menu-item index="/services">
              <el-icon><Grid /></el-icon>
              <span>服务管理</span>
            </el-menu-item>
            
            <el-menu-item index="/email-services">
              <el-icon><Connection /></el-icon>
              <span>关联管理</span>
            </el-menu-item>
          </el-menu>
        </el-aside>
  
        <!-- 主内容区域 -->
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </div>
  </template>
  
  <script>
  import { useRouter } from 'vue-router'
  import { useAuthStore } from '@/stores/auth'
  import { ElMessageBox } from 'element-plus'
  import {
    User,
    ArrowDown,
    SwitchButton,
    Odometer,
    Message,
    Grid,
    Connection
  } from '@element-plus/icons-vue'
  
  export default {
    name: 'MainLayout',
    components: {
      User,
      ArrowDown,
      SwitchButton,
      Odometer,
      Message,
      Grid,
      Connection
    },
    setup() {
      const router = useRouter()
      const authStore = useAuthStore()
      
      const handleUserCommand = async (command) => {
        switch (command) {
          case 'profile':
            router.push('/profile')
            break
          case 'logout':
            try {
              await ElMessageBox.confirm(
                '确定要退出登录吗？',
                '确认',
                {
                  confirmButtonText: '确定',
                  cancelButtonText: '取消',
                  type: 'warning'
                }
              )
              await authStore.logout()
              router.push('/login')
            } catch (error) {
              // 用户取消
            }
            break
        }
      }
      
      return {
        authStore,
        handleUserCommand
      }
    }
  }
  </script>
  
  <style scoped>
  .layout-container {
    height: 100vh;
    display: flex;
    flex-direction: column;
  }
  
  .header {
    background: #fff;
    border-bottom: 1px solid #e6e6e6;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 20px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    z-index: 1000;
  }
  
  .header-left h1 {
    margin: 0;
    color: #303133;
    font-size: 18px;
  }
  
  .header-right {
    display: flex;
    align-items: center;
  }
  
  .user-dropdown {
    display: flex;
    align-items: center;
    cursor: pointer;
    padding: 8px 12px;
    border-radius: 4px;
    transition: background-color 0.3s;
  }
  
  .user-dropdown:hover {
    background-color: #f5f7fa;
  }
  
  .user-dropdown .el-icon {
    margin: 0 4px;
  }
  
  .sidebar {
    background: #fff;
    border-right: 1px solid #e6e6e6;
    overflow: hidden;
  }
  
  .sidebar-menu {
    border: none;
    height: 100%;
  }
  
  .sidebar-menu .el-menu-item {
    height: 50px;
    line-height: 50px;
  }
  
  .main-content {
    background: #f5f7fa;
    overflow-y: auto;
  }
  
  :deep(.el-menu-item.is-active) {
    background-color: #ecf5ff !important;
    color: #409eff !important;
  }
  </style>