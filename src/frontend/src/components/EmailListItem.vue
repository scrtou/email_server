<template>
  <el-card shadow="hover" class="email-list-item" @click="goToEmail">
    <div class="email-header">
      <span :class="['from', { 'is-read': email.isRead }]">{{ fromText }}</span>
      <span class="date">{{ formattedDate }}</span>
    </div>
    <div :class="['subject', { 'is-read': email.isRead }]">{{ email.subject }}</div>
    <div class="snippet">{{ email.snippet }}</div>
  </el-card>
</template>

<script>
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { format } from 'date-fns';

export default {
  name: 'EmailListItem',
  props: {
    email: {
      type: Object,
      required: true,
    },
  },
  setup(props) {
    const router = useRouter();

    const fromText = computed(() => {
      if (!props.email.from || props.email.from.length === 0) {
        return 'Unknown Sender';
      }
      return props.email.from.map(f => f.name || f.address).join(', ');
    });

    const formattedDate = computed(() => {
      if (!props.email.date) return '';
      return format(new Date(props.email.date), 'yyyy-MM-dd HH:mm');
    });

    const goToEmail = () => {
      // Use messageId for routing as it's the unique identifier from IMAP
      router.push({ name: 'EmailDetail', params: { id: props.email.messageId } });
    };

    return {
      fromText,
      formattedDate,
      goToEmail,
    };
  },
};
</script>

<style scoped>
.email-list-item {
  margin-bottom: 10px;
  cursor: pointer;
}
.email-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 5px;
}
.from {
  font-weight: bold;
}
.subject {
  font-weight: bold;
  margin-bottom: 5px;
}
.is-read {
  font-weight: normal;
  color: #606266;
}
.date {
  color: #909399;
  font-size: 12px;
}
.snippet {
  color: #606266;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>