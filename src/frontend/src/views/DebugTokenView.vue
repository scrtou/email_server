<template>
  <div class="debug-view">
    <h2>令牌状态调试</h2>
    
    <el-button @click="testBasicTokens" type="primary">测试基础令牌数据</el-button>
    <el-button @click="testTokenStatus" type="success">测试令牌状态API</el-button>
    
    <div v-if="debugData" class="debug-result">
      <h3>调试结果：</h3>
      <pre>{{ JSON.stringify(debugData, null, 2) }}</pre>
    </div>
    
    <div v-if="error" class="error-result">
      <h3>错误信息：</h3>
      <pre>{{ error }}</pre>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import api from '@/utils/api'
import { ElMessage } from 'element-plus'

export default {
  name: 'DebugTokenView',
  setup() {
    const debugData = ref(null)
    const error = ref(null)
    
    const testBasicTokens = async () => {
      try {
        error.value = null
        console.log('Testing basic tokens...')
        const response = await api.get('/tokens/test')
        debugData.value = response.data || response
        ElMessage.success('基础令牌数据获取成功')
      } catch (err) {
        console.error('Test basic tokens error:', err)
        error.value = err.response?.data || err.message
        ElMessage.error('基础令牌数据获取失败')
      }
    }
    
    const testTokenStatus = async () => {
      try {
        error.value = null
        console.log('Testing token status...')
        const response = await api.get('/tokens/status')
        debugData.value = response.data || response
        ElMessage.success('令牌状态获取成功')
      } catch (err) {
        console.error('Test token status error:', err)
        error.value = err.response?.data || err.message
        ElMessage.error('令牌状态获取失败')
      }
    }
    
    return {
      debugData,
      error,
      testBasicTokens,
      testTokenStatus
    }
  }
}
</script>

<style scoped>
.debug-view {
  padding: 20px;
}

.debug-result {
  margin-top: 20px;
  padding: 15px;
  background-color: #f5f5f5;
  border-radius: 8px;
}

.error-result {
  margin-top: 20px;
  padding: 15px;
  background-color: #fff2f0;
  border: 1px solid #ffccc7;
  border-radius: 8px;
}

pre {
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style>