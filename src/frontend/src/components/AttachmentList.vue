<template>
  <div class="attachment-list" v-if="attachments && attachments.length > 0">
    <div class="attachment-grid">
      <div
        v-for="attachment in attachments"
        :key="attachment.filename"
        class="attachment-item"
        @click="downloadAttachment(attachment)"
      >
        <div class="attachment-icon">
          <el-icon :size="24">
            <component :is="getFileIcon(attachment.filename)" />
          </el-icon>
        </div>
        <div class="attachment-info">
          <div class="attachment-name" :title="attachment.filename">
            {{ attachment.filename }}
          </div>
          <div class="attachment-meta">
            <span class="attachment-size">{{ formatSize(attachment.size) }}</span>
            <span class="attachment-type">{{ getFileType(attachment.filename) }}</span>
          </div>
        </div>
        <div class="attachment-actions">
          <el-button
            :icon="Download"
            size="small"
            type="primary"
            plain
            @click.stop="downloadAttachment(attachment)"
            title="下载附件"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import {
  Paperclip,
  Document,
  Picture,
  VideoPlay,
  Headset,
  Files,
  Download
} from '@element-plus/icons-vue';

export default {
  name: 'AttachmentList',
  components: {
    Paperclip,
    Document,
    Picture,
    VideoPlay,
    Headset,
    Files,
    Download,
  },
  props: {
    attachments: {
      type: Array,
      default: () => [],
    },
  },
  setup() {
    const formatSize = (bytes) => {
      if (!bytes || bytes === 0) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
    };

    const getFileType = (filename) => {
      if (!filename) return 'Unknown';
      const ext = filename.split('.').pop()?.toLowerCase();
      const typeMap = {
        // 图片
        jpg: 'JPG', jpeg: 'JPEG', png: 'PNG', gif: 'GIF', bmp: 'BMP', svg: 'SVG', webp: 'WebP',
        // 文档
        pdf: 'PDF', doc: 'DOC', docx: 'DOCX', xls: 'XLS', xlsx: 'XLSX', ppt: 'PPT', pptx: 'PPTX',
        txt: 'TXT', rtf: 'RTF', odt: 'ODT', ods: 'ODS', odp: 'ODP',
        // 视频
        mp4: 'MP4', avi: 'AVI', mov: 'MOV', wmv: 'WMV', flv: 'FLV', mkv: 'MKV', webm: 'WebM',
        // 音频
        mp3: 'MP3', wav: 'WAV', flac: 'FLAC', aac: 'AAC', ogg: 'OGG', wma: 'WMA',
        // 压缩包
        zip: 'ZIP', rar: 'RAR', '7z': '7Z', tar: 'TAR', gz: 'GZ',
        // 其他
        html: 'HTML', css: 'CSS', js: 'JS', json: 'JSON', xml: 'XML'
      };
      return typeMap[ext] || ext?.toUpperCase() || 'File';
    };

    const getFileIcon = (filename) => {
      if (!filename) return Files;
      const ext = filename.split('.').pop()?.toLowerCase();

      // 图片文件
      if (['jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg', 'webp'].includes(ext)) {
        return Picture;
      }
      // 视频文件
      if (['mp4', 'avi', 'mov', 'wmv', 'flv', 'mkv', 'webm'].includes(ext)) {
        return VideoPlay;
      }
      // 音频文件
      if (['mp3', 'wav', 'flac', 'aac', 'ogg', 'wma'].includes(ext)) {
        return Headset;
      }
      // 文档文件
      if (['pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'txt', 'rtf'].includes(ext)) {
        return Document;
      }
      // 默认文件图标
      return Files;
    };

    const downloadAttachment = (attachment) => {
      // TODO: 实现附件下载功能
      console.log('Download attachment:', attachment.filename);
      // 这里应该调用后端API来下载附件
    };

    return {
      formatSize,
      getFileType,
      getFileIcon,
      downloadAttachment,
      Download,
    };
  },
};
</script>

<style scoped>
.attachment-list {
  width: 100%;
}

.attachment-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 12px;
}

.attachment-item {
  display: flex;
  align-items: center;
  padding: 12px;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  gap: 12px;
}

.attachment-item:hover {
  background: #e3f2fd;
  border-color: #409eff;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.15);
}

.attachment-icon {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #409eff;
  color: white;
  border-radius: 8px;
}

.attachment-info {
  flex: 1;
  min-width: 0;
}

.attachment-name {
  font-weight: 600;
  color: #303133;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 4px;
}

.attachment-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #909399;
}

.attachment-size {
  font-weight: 500;
}

.attachment-type {
  background: #e1f5fe;
  color: #0277bd;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 500;
}

.attachment-actions {
  flex-shrink: 0;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.attachment-item:hover .attachment-actions {
  opacity: 1;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .attachment-grid {
    grid-template-columns: 1fr;
  }

  .attachment-actions {
    opacity: 1;
  }

  .attachment-item {
    padding: 10px;
  }

  .attachment-icon {
    width: 36px;
    height: 36px;
  }
}
</style>