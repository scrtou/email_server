<template>
  <el-dialog
    :model-value="visible"
    :title="title"
    width="500px"
    :before-close="handleClose"
    @update:modelValue="$emit('update:visible', $event)"
    class="modal-dialog"
    append-to-body
    draggable
    align-center
  >
    <div class="modal-content">
      <slot></slot>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <slot name="footer">
          <el-button @click="handleCancel">{{ cancelButtonText }}</el-button>
          <el-button type="primary" @click="handleConfirm">{{ confirmButtonText }}</el-button>
        </slot>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ElDialog, ElButton } from 'element-plus';

// eslint-disable-next-line no-undef
defineProps({
  visible: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: '提示',
  },
  confirmButtonText: {
    type: String,
    default: '确认',
  },
  cancelButtonText: {
    type: String,
    default: '取消',
  },
});

// eslint-disable-next-line no-undef
const emit = defineEmits(['update:visible', 'close', 'confirm', 'cancel']);

const handleClose = (done) => {
  emit('update:visible', false);
  emit('close');
  if (done) {
    done();
  }
};

const handleConfirm = () => {
  emit('confirm');
  // emit('update:visible', false); // Optionally close on confirm, parent should decide
};

const handleCancel = () => {
  emit('cancel');
  emit('update:visible', false); // Always close on cancel
};
</script>

<style scoped>
.modal-dialog :deep(.el-dialog__header) {
  margin-right: 0; /* Reset Element Plus default */
  border-bottom: 1px solid #ebeef5;
  padding: 15px 20px;
}

.modal-dialog :deep(.el-dialog__title) {
  font-size: 1.1rem;
  font-weight: 600;
}

.modal-dialog :deep(.el-dialog__body) {
  padding: 20px 30px; /* 增加左右内边距以确保内容居中 */
  max-height: 60vh;
  overflow-y: auto;
}

.modal-dialog :deep(.el-dialog__footer) {
  border-top: 1px solid #ebeef5;
  padding: 10px 20px;
  text-align: right;
}

.dialog-footer .el-button {
  min-width: 80px;
}
</style>