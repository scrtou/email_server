<template>
  <div class="attachment-list" v-if="attachments && attachments.length > 0">
    <h3>Attachments</h3>
    <el-tag
      v-for="attachment in attachments"
      :key="attachment.filename"
      class="attachment-tag"
    >
      <el-icon><Paperclip /></el-icon>
      <span class="filename">{{ attachment.filename }}</span>
      <span class="size">({{ formatSize(attachment.size) }})</span>
    </el-tag>
  </div>
</template>

<script>
import { Paperclip } from '@element-plus/icons-vue';

export default {
  name: 'AttachmentList',
  components: {
    Paperclip,
  },
  props: {
    attachments: {
      type: Array,
      default: () => [],
    },
  },
  setup() {
    const formatSize = (bytes) => {
      if (bytes === 0) return '0 Bytes';
      const k = 1024;
      const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    return {
      formatSize,
    };
  },
};
</script>

<style scoped>
.attachment-list {
  margin-top: 20px;
  padding-top: 10px;
  border-top: 1px solid #ebeef5;
}
.attachment-tag {
  margin-right: 10px;
  margin-bottom: 10px;
}
.filename {
  margin-left: 5px;
}
.size {
  margin-left: 5px;
  color: #909399;
}
</style>