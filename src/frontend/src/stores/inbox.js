// stores/inbox.js

import { defineStore } from 'pinia';
import { getInboxEmails, getEmailDetail } from '@/utils/api';

export const useInboxStore = defineStore('inbox', {
  state: () => ({
    emails: [],
    isLoading: false,
    error: null,
    page: 1,
    pageSize: 20,
    totalEmails: 0,
    hasMore: true,
    selectedAccountId: null,
    selectedFolder: 'inbox', // 当前选择的文件夹
  }),
  getters: {
    // 返回按日期逆序排列的邮件（最新的在前面）
    sortedEmails: (state) => {
      return [...state.emails].sort((a, b) => {
        // 如果没有日期，放到最后
        if (!a.date && !b.date) return 0;
        if (!a.date) return 1;
        if (!b.date) return -1;

        // 按日期逆序排列（最新的在前面）
        return new Date(b.date) - new Date(a.date);
      });
    }
  },
  actions: {
   async selectAccount(accountId) {
     this.selectedAccountId = accountId;
     this.emails = [];
     this.page = 1;
     this.totalEmails = 0;
     this.hasMore = true;
     this.error = null;
     await this.fetchEmails();
   },

   async selectFolder(folderName) {
     this.selectedFolder = folderName;
     this.emails = [];
     this.page = 1;
     this.totalEmails = 0;
     this.hasMore = true;
     this.error = null;
     await this.fetchEmails();
   },

    async fetchEmails(loadMore = false) {
      console.log('📧 fetchEmails called with loadMore:', loadMore);
      console.log('📧 selectedAccountId:', this.selectedAccountId);

      if (!this.selectedAccountId) {
        this.error = "Please select an email account.";
        console.log('❌ No selectedAccountId, returning');
        return;
      }
      if (this.isLoading || (!this.hasMore && loadMore)) {
        console.log('❌ Already loading or no more emails, returning');
        return;
      }

      console.log('📧 Starting fetchEmails...');
      this.isLoading = true;
      this.error = null;

      if (!loadMore) {
        this.page = 1;
        this.emails = [];
        this.hasMore = true;
      }

     // stores/inbox.js -> fetchEmails

// ...
      try {
        const response = await getInboxEmails({
          page: this.page,
          pageSize: this.pageSize,
          account_id: this.selectedAccountId,
          folder: this.selectedFolder
        });

        // ★★★★★ CORE FIX IS HERE ★★★★★
        // The API interceptor already extracts response.data.data, so response is the actual data
        console.log('📧 Raw response:', response);
        console.log('📧 Response type:', typeof response);
        console.log('📧 Response keys:', Object.keys(response));

        // response is already the extracted data from the interceptor
        const responseData = response;
        console.log('📧 Processed responseData:', responseData);

        const newEmails = responseData.emails || [];
        const total = responseData.total || 0;

        console.log('📧 New emails:', newEmails);
        console.log('📧 Total:', total);

        if (newEmails.length > 0) {
          this.emails.push(...newEmails);
          console.log('📧 Updated emails array:', this.emails);
        }

        this.totalEmails = total;
        this.page += 1;
        
        // Update hasMore status based on the total count
        if (this.emails.length >= this.totalEmails) {
          this.hasMore = false;
        }


      } catch (error) {
        // More robust error handling
        if (error.response && error.response.data && error.response.data.error) {
          this.error = error.response.data.error;
        } else {
          this.error = error.message || 'An unknown error occurred while fetching emails.';
        }
        console.error('Error fetching emails:', error);
      } finally {
        this.isLoading = false;
      }
    },
    getEmailById(messageId) {
      // Corrected to use the right property from backend response
      return this.emails.find(email => email.messageId === messageId);
    },

    async fetchEmailDetail(messageId) {
      console.log('📧 fetchEmailDetail called with messageId:', messageId);

      if (!this.selectedAccountId) {
        throw new Error("Please select an email account.");
      }

      try {
        const response = await getEmailDetail(messageId, {
          account_id: this.selectedAccountId
        });

        console.log('📧 Email detail response:', response);

        // The API interceptor already extracts the data, so response is the actual email data
        const emailDetail = response;
        console.log('📧 Email detail data:', emailDetail);

        // Update the email in our local store if it exists
        const existingEmailIndex = this.emails.findIndex(email => email.messageId === messageId);
        if (existingEmailIndex !== -1) {
          this.emails[existingEmailIndex] = emailDetail;
          console.log('📧 Updated email in store at index:', existingEmailIndex);
        }

        return emailDetail;
      } catch (error) {
        console.error('Error fetching email detail:', error);
        throw error;
      }
    }
  },
});