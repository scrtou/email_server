<!-- InboxView.vue -->
<template>
    <div class="inbox-view" @scroll="handleScroll" ref="inboxContainer">
      <div class="inbox-header">
        <h1>Inbox</h1>
        <div class="header-controls">
          <el-button
            type="primary"
            :icon="RefreshIcon"
            @click="handleRefresh"
            :loading="inboxStore.isLoading"
            size="default"
          >
            刷新
          </el-button>
          <el-select
            v-model="inboxStore.selectedFolder"
            placeholder="选择文件夹"
            @change="handleFolderChange"
            style="width: 150px;"
          >
            <el-option label="收件箱" value="inbox" />
            <el-option label="垃圾邮件" value="junkemail" />
            <el-option label="已发送" value="sentitems" />
            <el-option label="草稿箱" value="drafts" />
            <el-option label="已删除" value="deleteditems" />
          </el-select>
          <el-select
            v-model="inboxStore.selectedAccountId"
            placeholder="Select an Account"
            @change="handleAccountChange"
            style="width: 250px;"
            filterable
          >
            <el-option
              v-for="account in emailAccountStore.emailAccounts"
              :key="account.id"
              :label="account.email_address"
              :value="account.id"
            />
          </el-select>
        </div>
      </div>
      <!-- Global error display, always shows if there's an error -->
      <el-alert
        v-if="inboxStore.error"
        :title="inboxStore.error"
        type="error"
        show-icon
        closable
        @close="inboxStore.error = null"
        style="margin-bottom: 20px;"
      />
      


      <div v-if="inboxStore.isLoading && !inboxStore.emails.length" v-loading="true" class="loading-spinner"></div>

      <div v-else-if="inboxStore.emails.length > 0" class="email-list">
        <!-- ★★★★★ KEY BINDING FIX IS HERE ★★★★★ -->
        <!-- 使用sortedEmails getter来显示按日期逆序排列的邮件 -->
        <email-list-item
          v-for="email in inboxStore.sortedEmails"
          :key="email.messageId"
          :email="email"
        />
      </div>

      <!-- Show empty state only if there's no loading and no error -->
      <el-empty v-else-if="!inboxStore.isLoading && !inboxStore.error" description="Select an account to view emails."></el-empty>

      <div v-if="inboxStore.isLoading && inboxStore.emails.length > 0" class="loading-more">
        Loading more...
      </div>
      
      <div v-if="!inboxStore.hasMore && inboxStore.emails.length > 0" class="no-more-emails">
        No more emails.
      </div>
    </div>
</template>

<script>
// The <script> and <style> sections remain unchanged.
import { onMounted, onUnmounted, ref } from 'vue';
import { useInboxStore } from '@/stores/inbox';
import { useEmailAccountStore } from '@/stores/emailAccount';
import EmailListItem from '@/components/EmailListItem.vue';
import { Refresh as RefreshIcon } from '@element-plus/icons-vue';


export default {
  name: 'InboxView',
  components: {
    EmailListItem,
  },
  setup() {
    const inboxStore = useInboxStore();
    const emailAccountStore = useEmailAccountStore();
    const inboxContainer = ref(null);

    const handleAccountChange = async (accountId) => {
      console.log('🔄 handleAccountChange called with accountId:', accountId);
      console.log('🔄 inboxStore:', inboxStore);
      console.log('🔄 inboxStore.selectAccount:', inboxStore.selectAccount);
      try {
        await inboxStore.selectAccount(accountId);
        console.log('🔄 selectAccount call completed');
      } catch (error) {
        console.error('🔄 Error calling selectAccount:', error);
      }
    };

    const handleRefresh = async () => {
      console.log('🔄 handleRefresh called');
      if (!inboxStore.selectedAccountId) {
        console.log('❌ No account selected for refresh');
        return;
      }

      try {
        // 重置邮件列表并重新获取
        await inboxStore.selectAccount(inboxStore.selectedAccountId);
        console.log('✅ Refresh completed');
      } catch (error) {
        console.error('❌ Error during refresh:', error);
      }
    };

    const handleFolderChange = async (folderName) => {
      console.log('📁 handleFolderChange called with folderName:', folderName);
      if (!inboxStore.selectedAccountId) {
        console.log('❌ No account selected for folder change');
        return;
      }

      try {
        await inboxStore.selectFolder(folderName);
        console.log('✅ Folder change completed');
      } catch (error) {
        console.error('❌ Error during folder change:', error);
      }
    };

    const handleScroll = () => {
      const container = inboxContainer.value;
      if (container) {
        const { scrollTop, scrollHeight, clientHeight } = container;
        if (scrollHeight - scrollTop - clientHeight < 100) {
          inboxStore.fetchEmails(true);
        }
      }
    };

   // InboxView.vue -> onMounted

// ... in setup() ...

onMounted(async () => { // ★ 标记为 async
  console.log('🚀 InboxView onMounted started');

  // 清空旧数据
  inboxStore.emails = [];
  inboxStore.error = null;
  emailAccountStore.clearAccounts();

  // 1. 异步获取账户列表
  console.log('📋 Fetching email accounts...');
  await emailAccountStore.fetchEmailAccounts(1, 10000);
  console.log('📋 Email accounts fetched:', emailAccountStore.emailAccounts);

  // 2. 检查是否有账户
  if (emailAccountStore.emailAccounts.length > 0) {
    console.log('✅ Found email accounts, selecting first one...');
    // 3. 如果当前没有选中的账户，或者选中的账户不在新列表里，就默认选中第一个
    const currentAccountExists = emailAccountStore.emailAccounts.some(acc => acc.id === inboxStore.selectedAccountId);
    if (!inboxStore.selectedAccountId || !currentAccountExists) {
        console.log('🎯 Selecting first account:', emailAccountStore.emailAccounts[0].id);
        // ★ 主动触发账户选择流程
        await handleAccountChange(emailAccountStore.emailAccounts[0].id);
    } else {
        console.log('🎯 Using existing selected account:', inboxStore.selectedAccountId);
        // 如果已选中的账户有效，则为其获取邮件
        await handleAccountChange(inboxStore.selectedAccountId);
    }
  } else {
    console.log('❌ No email accounts found');
  }
});

    onUnmounted(() => {
      const container = inboxContainer.value;
      if (container) {
        container.removeEventListener('scroll', handleScroll);
      }
    });

    return {
      inboxStore,
      emailAccountStore,
      inboxContainer,
      handleScroll,
      handleAccountChange,
      handleRefresh,
      handleFolderChange,
      RefreshIcon,
    };
  },
};
</script>

<style scoped>
.inbox-view {
  padding: 20px;
  height: 100%;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
.inbox-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
}
.header-controls {
  display: flex;
  align-items: center;
  gap: 15px;
}
.email-list {
  flex-grow: 1;
  overflow-y: auto;
}
.loading-spinner {
  height: 200px;
}
.email-list {
  margin-top: 20px;
}
.loading-more, .no-more-emails {
  text-align: center;
  padding: 20px;
  color: #909399;
}
</style>