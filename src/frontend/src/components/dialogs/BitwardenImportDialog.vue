<template>
  <el-dialog
    :model-value="visible"
    title="导入 Bitwarden CSV"
    width="600px"
    :before-close="handleClose"
    @update:model-value="$emit('update:visible', $event)"
    class="bitwarden-import-dialog"
    append-to-body
    draggable
    align-center
    :close-on-click-modal="false"
  >
    <div class="import-content">
      <!-- 文件上传区域 -->
      <div class="upload-section">
        <el-upload
          ref="uploadRef"
          class="upload-dragger"
          drag
          :auto-upload="false"
          :show-file-list="false"
          accept=".csv"
          :on-change="handleFileChange"
          :before-upload="() => false"
        >
          <div class="upload-content">
            <el-icon class="upload-icon" size="48">
              <Upload />
            </el-icon>
            <div class="upload-text">
              <p class="upload-title">选择 Bitwarden CSV 文件</p>
              <p class="upload-hint">点击或拖拽文件到此区域上传</p>
            </div>
          </div>
        </el-upload>
        
        <!-- 已选择的文件信息 -->
        <div v-if="selectedFile" class="file-info">
          <el-tag type="success" size="large" closable @close="clearFile">
            <el-icon><Document /></el-icon>
            {{ selectedFile.name }}
            <span class="file-size">({{ formatFileSize(selectedFile.size) }})</span>
          </el-tag>
        </div>
      </div>

      <!-- 导入选项 -->
      <div class="import-options">
        <el-card shadow="never" class="options-card">
          <template #header>
            <span class="options-title">导入选项</span>
          </template>
          <el-checkbox v-model="importPasswords" size="large">
            <span class="checkbox-label">导入密码</span>
            <span class="checkbox-hint">（请谨慎操作，密码将以加密形式存储）</span>
          </el-checkbox>
        </el-card>
      </div>

      <!-- 状态信息 -->
      <div v-if="statusMessage" class="status-section">
        <el-alert
          :title="statusMessage"
          :type="messageType === 'success' ? 'success' : messageType === 'error' ? 'error' : 'info'"
          :closable="false"
          show-icon
        >
          <template v-if="importedCount !== null" #default>
            <p>{{ statusMessage }}</p>
            <p class="import-count">成功导入 <strong>{{ importedCount }}</strong> 条记录</p>
          </template>
        </el-alert>
      </div>

      <!-- 进度条 -->
      <div v-if="isLoading || uploadProgress > 0" class="progress-section">
        <el-progress
          :percentage="uploadProgress"
          :status="uploadProgress === 100 ? 'success' : undefined"
          :stroke-width="8"
        />
        <p class="progress-text">{{ progressText }}</p>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose" :disabled="isLoading">
          {{ messageType === 'success' ? '关闭' : '取消' }}
        </el-button>
        <el-button
          v-if="messageType !== 'success'"
          type="primary"
          @click="handleImport"
          :loading="isLoading"
          :disabled="!selectedFile || isLoading"
        >
          {{ isLoading ? '正在导入...' : '开始导入' }}
        </el-button>
        <el-button
          v-else
          type="success"
          @click="handleClose"
        >
          完成
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue';
import { Upload, Document } from '@element-plus/icons-vue';
import api from '@/utils/api';

// Props
// eslint-disable-next-line no-undef
defineProps({
  visible: {
    type: Boolean,
    default: false,
  },
});

// Emits
// eslint-disable-next-line no-undef
const emit = defineEmits(['update:visible', 'import-success']);

// Refs
const uploadRef = ref(null);
const selectedFile = ref(null);
const importPasswords = ref(false); // 默认不勾选导入密码
const isLoading = ref(false);
const statusMessage = ref('');
const messageType = ref(''); // 'success', 'error', 'info'
const importedCount = ref(null);
const uploadProgress = ref(0);
const progressText = ref('');

// Methods
const handleFileChange = (file) => {
  selectedFile.value = file.raw;
  statusMessage.value = '';
  messageType.value = '';
  importedCount.value = null;
  uploadProgress.value = 0;
};

const clearFile = () => {
  selectedFile.value = null;
  statusMessage.value = '';
  messageType.value = '';
  importedCount.value = null;
  uploadProgress.value = 0;
  if (uploadRef.value) {
    uploadRef.value.clearFiles();
  }
};

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const handleClose = () => {
  if (!isLoading.value) {
    emit('update:visible', false);
    // 重置状态
    setTimeout(() => {
      clearFile();
      statusMessage.value = '';
      messageType.value = '';
      importedCount.value = null;
      uploadProgress.value = 0;
      progressText.value = '';
      importPasswords.value = false; // 重置为默认不勾选
    }, 300);
  }
};

const handleImport = async () => {
  if (!selectedFile.value) {
    statusMessage.value = '请先选择一个 CSV 文件。';
    messageType.value = 'error';
    return;
  }

  isLoading.value = true;
  uploadProgress.value = 10;
  progressText.value = '正在上传文件...';
  statusMessage.value = '';
  messageType.value = '';
  importedCount.value = null;

  const formData = new FormData();
  formData.append('file', selectedFile.value);
  formData.append('importPasswords', importPasswords.value);

  try {
    uploadProgress.value = 30;
    progressText.value = '正在处理文件...';
    
    const response = await api.post('/import/bitwarden-csv', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      },
      timeout: 60000 // 设置60秒超时，适应大文件导入
    });

    uploadProgress.value = 60;
    progressText.value = '正在解析数据...';

    // 等待一小段时间显示解析进度
    await new Promise(resolve => setTimeout(resolve, 300));
    uploadProgress.value = 80;
    progressText.value = '正在保存数据...';

    console.log('Full response object:', response);

    // 检查响应状态
    if (response && typeof response.message !== 'undefined' && typeof response.savedCount !== 'undefined') {
      // 等待一小段时间显示保存进度
      await new Promise(resolve => setTimeout(resolve, 200));
      uploadProgress.value = 100;
      progressText.value = '导入完成！';

      statusMessage.value = response.message || '导入成功！';
      messageType.value = 'success';
      importedCount.value = response.savedCount;

      // 通知父组件导入成功
      emit('import-success', {
        count: response.savedCount || 0,
        message: response.message || '导入成功！'
      });
    } else {
      console.error('导入成功但响应结构不符:', response);
      statusMessage.value = '导入成功，但无法解析响应数据。';
      messageType.value = 'warning';
      importedCount.value = null;
      uploadProgress.value = 0;
    }

  } catch (error) {
    console.error('导入失败:', error);
    
    uploadProgress.value = 0;
    progressText.value = '';
    
    let errorMessage = '导入失败。';
    
    if (error.response && error.response.data) {
      if (typeof error.response.data === 'string') {
        errorMessage += ` 错误: ${error.response.data}`;
      } else if (error.response.data.error) {
        errorMessage += ` 错误: ${error.response.data.error}`;
      } else if (error.response.data.message) {
        errorMessage += ` 错误: ${error.response.data.message}`;
      } else {
        errorMessage += ` 后端返回了错误，但无法解析具体信息。状态码: ${error.response.status}`;
      }
    } else if (error.request) {
      errorMessage += ' 无法连接到服务器，请检查网络连接或后端服务状态。';
    } else if (error.message) {
      errorMessage += ` 错误: ${error.message}`;
    }
    
    if (errorMessage === '导入失败。') {
      errorMessage += ' 发生未知错误。';
    }

    statusMessage.value = errorMessage;
    messageType.value = 'error';
    importedCount.value = null;
  } finally {
    isLoading.value = false;
  }
};
</script>

<style scoped>
.bitwarden-import-dialog :deep(.el-dialog__header) {
  border-bottom: 1px solid #ebeef5;
  padding: 20px 24px 16px;
}

.bitwarden-import-dialog :deep(.el-dialog__title) {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.bitwarden-import-dialog :deep(.el-dialog__body) {
  padding: 24px;
  max-height: 70vh;
  overflow-y: auto;
}

.import-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* 上传区域样式 */
.upload-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.upload-dragger :deep(.el-upload-dragger) {
  width: 100%;
  height: 160px;
  border: 2px dashed #d9d9d9;
  border-radius: 8px;
  background: #fafafa;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.upload-dragger :deep(.el-upload-dragger:hover) {
  border-color: #409eff;
  background: #f0f9ff;
}

.upload-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  text-align: center;
}

.upload-icon {
  color: #c0c4cc;
  transition: color 0.3s ease;
}

.upload-dragger:hover .upload-icon {
  color: #409eff;
}

.upload-text {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.upload-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  margin: 0;
}

.upload-hint {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

/* 文件信息样式 */
.file-info {
  display: flex;
  justify-content: center;
}

.file-info .el-tag {
  padding: 8px 16px;
  font-size: 14px;
  border-radius: 6px;
}

.file-size {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}

/* 导入选项样式 */
.import-options {
  margin: 8px 0;
}

.options-card {
  border: 1px solid #ebeef5;
  border-radius: 8px;
}

.options-card :deep(.el-card__header) {
  padding: 16px 20px;
  background: #f8f9fa;
  border-bottom: 1px solid #ebeef5;
}

.options-card :deep(.el-card__body) {
  padding: 20px;
}

.options-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.checkbox-label {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-right: 8px;
}

.checkbox-hint {
  font-size: 12px;
  color: #909399;
}

/* 状态信息样式 */
.status-section {
  margin: 8px 0;
}

.import-count {
  margin-top: 8px;
  font-size: 14px;
}

/* 进度条样式 */
.progress-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 8px 0;
}

.progress-text {
  text-align: center;
  font-size: 14px;
  color: #606266;
  margin: 0;
}

/* 对话框底部样式 */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #ebeef5;
  background: #fafafa;
}

.dialog-footer .el-button {
  min-width: 80px;
  padding: 8px 20px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .bitwarden-import-dialog :deep(.el-dialog) {
    width: 90% !important;
    margin: 0 5%;
  }

  .upload-dragger :deep(.el-upload-dragger) {
    height: 120px;
  }

  .upload-title {
    font-size: 14px;
  }

  .upload-hint {
    font-size: 12px;
  }
}
</style>
