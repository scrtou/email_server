<template>
  <div class="import-bitwarden-view">
    <h1>导入 Bitwarden CSV</h1>
    <form @submit.prevent="handleImport">
      <div class="form-group">
        <label for="csvFile">选择 Bitwarden CSV 文件:</label>
        <input
          type="file"
          id="csvFile"
          ref="csvFile"
          accept=".csv"
          @change="handleFileChange"
          required
        />
      </div>
      <div class="form-group">
        <label>
          <input type="checkbox" v-model="importPasswords" />
          导入密码 (请谨慎操作)
        </label>
      </div>
      <button type="submit" :disabled="isLoading">
        {{ isLoading ? '正在导入...' : '导入' }}
      </button>
    </form>

    <div v-if="statusMessage" class="status-message" :class="messageType">
      <p>{{ statusMessage }}</p>
      <p v-if="importedCount !== null">成功导入 {{ importedCount }} 条记录。</p>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue';
import api from '@/utils/api'; // 假设 api 工具路径正确
 export default {
  name: 'ImportBitwardenView',
  setup() {
    const csvFile = ref(null);
    const selectedFile = ref(null);
    const importPasswords = ref(false);
    const isLoading = ref(false);
    const statusMessage = ref('');
    const messageType = ref(''); // 'success' or 'error'
    const importedCount = ref(null);

    const handleFileChange = (event) => {
      selectedFile.value = event.target.files[0];
      statusMessage.value = ''; // Clear previous messages on new file selection
      messageType.value = '';
      importedCount.value = null;
    };

    const handleImport = async () => {
      if (!selectedFile.value) {
        statusMessage.value = '请先选择一个 CSV 文件。';
        messageType.value = 'error';
        return;
      }

      isLoading.value = true;
      statusMessage.value = '正在上传和处理文件...';
      messageType.value = 'info'; // Or use a specific class for loading
      importedCount.value = null;

      const formData = new FormData();
      formData.append('file', selectedFile.value);
      formData.append('import_passwords', importPasswords.value);

      try {
        // Axios 拦截器会自动添加认证 token
        // 只需要确保 Content-Type 正确
        const response = await api.post('/import/bitwarden-csv', formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
            // Authorization header is handled by the interceptor
          }
        });

        // 添加这些日志输出
        console.log('Full response object (this is actual_payload from interceptor):', response);
        // console.log('Response data (this would be actual_payload.data, likely undefined):', response.data); // No longer needed with new logic

        // 由于 api.js 的响应拦截器已经提取了 response.data.data,
        // 这里的 'response' 变量直接就是后端返回的 data 对象内部的内容。
        if (response && typeof response.message !== 'undefined' && typeof response.savedCount !== 'undefined') {
          statusMessage.value = response.message || '导入成功！';
          messageType.value = 'success';
          importedCount.value = response.savedCount;
        } else {
          // 如果响应结构不符合预期
          console.error('导入成功但响应结构不符 (expected fields missing in actual_payload):', response);
          statusMessage.value = '导入成功，但无法解析响应数据。';
          messageType.value = 'warning';
          importedCount.value = null;
        }

      } catch (error) {
        console.error('导入失败:', error);
        let errorMessage = '导入失败。';

        // 更健壮地提取后端错误信息
        if (error.response && error.response.data) {
          // 检查后端是否直接返回了字符串错误
          if (typeof error.response.data === 'string') {
            errorMessage += ` 错误: ${error.response.data}`;
          }
          // 检查常见的错误结构 { error: "..." }
          else if (error.response.data.error) {
            errorMessage += ` 错误: ${error.response.data.error}`;
          }
          // 检查常见的错误结构 { message: "..." }
          else if (error.response.data.message) {
            errorMessage += ` 错误: ${error.response.data.message}`;
          }
          // 如果有响应数据但无法识别错误字段
          else {
            errorMessage += ` 后端返回了错误，但无法解析具体信息。状态码: ${error.response.status}`;
          }
        } else if (error.request) {
          // 请求已发出，但没有收到响应
          errorMessage += ' 无法连接到服务器，请检查网络连接或后端服务状态。';
        } else if (error.message) {
          // 设置请求时发生错误
           errorMessage += ` 错误: ${error.message}`;
        }
        
        // 如果以上都未能提取到具体错误，则显示通用错误
        if (errorMessage === '导入失败。') {
          errorMessage += ' 发生未知错误。';
        }

        statusMessage.value = errorMessage;
        messageType.value = 'error';
        importedCount.value = null;
      } finally {
        isLoading.value = false;
        // Reset file input for potential re-upload of the same file name
        if (csvFile.value) {
          csvFile.value.value = '';
        }
        selectedFile.value = null; // Clear selected file reference after attempt
      }
    };

    return {
      csvFile, // ref for the input element
      importPasswords,
      isLoading,
      statusMessage,
      messageType,
      importedCount,
      handleFileChange,
      handleImport,
    };
  },
};
</script>

<style scoped>
.import-bitwarden-view {
  max-width: 600px;
  margin: 2rem auto;
  padding: 2rem;
  border: 1px solid #ccc;
  border-radius: 8px;
  background-color: #f9f9f9;
}

h1 {
  text-align: center;
  margin-bottom: 1.5rem;
}

.form-group {
  margin-bottom: 1rem;
}

label {
  display: block;
  margin-bottom: 0.5rem;
}

input[type="file"] {
  display: block;
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
}

input[type="checkbox"] {
  margin-right: 0.5rem;
}

button {
  padding: 0.75rem 1.5rem;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s ease;
  width: 100%;
  font-size: 1rem;
}

button:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}

button:not(:disabled):hover {
  background-color: #0056b3;
}

.status-message {
  margin-top: 1.5rem;
  padding: 1rem;
  border-radius: 4px;
  border: 1px solid transparent;
}

.status-message.info {
  background-color: #e7f3fe;
  border-color: #d0eaff;
  color: #0c5460;
}

.status-message.success {
  background-color: #d4edda;
  border-color: #c3e6cb;
  color: #155724;
}

.status-message.error {
  background-color: #f8d7da;
  border-color: #f5c6cb;
  color: #721c24;
}
</style>