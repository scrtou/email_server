// Vue.js 3 应用程序的主入口文件

import { createApp } from 'vue'
import ElementPlus from 'element-plus'// 引入 Element Plus 组件库
import { createPinia } from 'pinia'
import 'element-plus/dist/index.css'// 引入 Element Plus 样式
import '@/assets/styles/utilities.css'// 引入自定义工具样式
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
/*
App 组件的导入指向了应用程序的根组件，通常位于同级目录的 App.vue 文件中。
这个根组件是整个应用程序的顶层容器，所有其他组件都将作为它的子组件存在。
*/
import App from './App.vue'// 引入 App 组件
import router from './router'// 引入 router 路由
import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'

//Vue 3 使用 .use() 方法来注册插件和扩展功能
const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(ElementPlus)
/*
app.use(router) 将之前创建的路由器实例注册到应用程序中。这个注册过程会完成以下关键操作：
路由器会在应用程序中注册两个重要的全局组件：<router-link> 用于创建导航链接，<router-view> 用于渲染当前路由匹配的组件。
它还会在所有组件实例中注入 $router 和 $route 属性，使得组件可以访问路由器实例和当前路由信息。在组合式 API 中，可以通过 useRouter() 和 useRoute() 函数来访问这些功能。
*/
app.use(router)

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
  }
  
  // 初始化认证状态
  const authStore = useAuthStore()
  authStore.initAuth()

  // 初始化设置状态
  const settingsStore = useSettingsStore()
  settingsStore.loadSettings()
/*
最后这行代码将配置完成的 Vue 应用程序实例挂载到 DOM 中指定的元素上。'#app' 是一个 CSS 选择器，指向 HTML 文档中 id 为 "app" 的元素，通常这个元素位于 index.html 文件中：
挂载过程会触发以下重要步骤：
Vue 会将根组件 App 渲染成虚拟 DOM 树，然后将这个虚拟 DOM 转换为实际的 DOM 节点，并替换掉目标元素的内容。
挂载完成后，整个 Vue 应用程序开始运行，响应式系统被激活，路由器开始监听 URL 变化，所有的生命周期钩子按照正确的顺序执行。
*/


app.mount('#app')

