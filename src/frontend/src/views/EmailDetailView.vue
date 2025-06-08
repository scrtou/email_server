<template>
  <div class="email-detail-view">
    <el-page-header @back="goBack" :content="email ? email.subject : 'Loading...'">
    </el-page-header>

    <el-card v-if="email" class="email-container">
      <div class="email-meta">
        <p><strong>From:</strong> {{ fromText }}</p>
        <p><strong>To:</strong> {{ toText }}</p>
        <p v-if="ccText"><strong>Cc:</strong> {{ ccText }}</p>
        <p><strong>Date:</strong> {{ formattedDate }}</p>
      </div>
      <hr />
      <div v-if="email.htmlBody" v-html="sanitizedHtml" class="email-body"></div>
      <div v-else class="email-body">
        <pre>{{ email.body }}</pre>
      </div>
      <attachment-list v-if="email.attachments && email.attachments.length > 0" :attachments="email.attachments" />
    </el-card>
    <div v-else v-loading="true" class="loading-spinner">
      <p>Loading email...</p>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useInboxStore } from '@/stores/inbox';
import AttachmentList from '@/components/AttachmentList.vue';
import DOMPurify from 'dompurify';
import { format } from 'date-fns';

export default {
  name: 'EmailDetailView',
  components: {
    AttachmentList,
  },
  setup() {
    const route = useRoute();
    const router = useRouter();
    const store = useInboxStore();
    const email = ref(null);

    const formatAddresses = (addresses) => {
      if (!addresses || addresses.length === 0) return '';
      return addresses.map(addr => addr.name ? `${addr.name} <${addr.address}>` : addr.address).join(', ');
    };

    const fromText = computed(() => formatAddresses(email.value?.from));
    const toText = computed(() => formatAddresses(email.value?.to));
    const ccText = computed(() => formatAddresses(email.value?.cc));

    const formattedDate = computed(() => {
      if (!email.value?.date) return '';
      return format(new Date(email.value.date), 'yyyy-MM-dd HH:mm:ss');
    });

    const sanitizedHtml = computed(() => {
      if (!email.value?.htmlBody) return '';
      return DOMPurify.sanitize(email.value.htmlBody);
    });

    const goBack = () => {
      router.push({ name: 'Inbox' });
    };



    onMounted(async () => {
      const emailId = route.params.id;

      try {
        // First try to get from local store
        let fetchedEmail = store.getEmailById(emailId);

        if (!fetchedEmail) {
          // If not found locally, try to fetch emails first
          if (!store.selectedAccountId) {
            // If no account is selected, we can't fetch emails
            console.error('No account selected for fetching emails');
            return;
          }
          await store.fetchEmails();
          fetchedEmail = store.getEmailById(emailId);
        }

        // If we still don't have the email or it doesn't have body content, fetch detail
        if (!fetchedEmail || (!fetchedEmail.body && !fetchedEmail.htmlBody)) {
          fetchedEmail = await store.fetchEmailDetail(emailId);
        }

        email.value = fetchedEmail;
      } catch (error) {
        console.error('Error loading email detail:', error);
        // You might want to show an error message to the user here
      }
    });

    return {
      email,
      fromText,
      toText,
      ccText,
      formattedDate,
      sanitizedHtml,
      goBack,
    };
  },
};
</script>

<style scoped>
.email-detail-view {
  padding: 20px;
}
.email-container {
  margin-top: 20px;
}
.email-meta {
  margin-bottom: 20px;
  color: #606266;
}
.email-meta p {
  margin: 5px 0;
}
.email-body {
  margin-top: 20px;
  line-height: 1.6;
}
.email-body pre {
  white-space: pre-wrap;
  word-wrap: break-word;
}
.loading-spinner {
  height: 300px;
}
</style>