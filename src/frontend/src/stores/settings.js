import { defineStore } from 'pinia';

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    // 页面独立分页设置
    pageSettings: {
      // 邮箱账户管理页面
      emailAccounts: {
        pageSize: 8,
        pageSizeOptions: [5,10, 15, 30, 50]
      },
      // 平台管理页面
      platforms: {
        pageSize: 8,
        pageSizeOptions: [5,10, 15, 30, 50]
      },
      // 平台注册管理页面
      platformRegistrations: {
        pageSize: 8,
        pageSizeOptions: [5,10, 15, 30, 50]
      },
      // 服务订阅管理页面
      serviceSubscriptions: {
        pageSize: 8,
        pageSizeOptions: [5,10, 15, 30, 50]
      }
    },
    // 全局默认设置（作为后备）
    globalDefaults: {
      pageSize: 10,
      pageSizeOptions: [5, 10, 15, 20, 30, 50]
    }
  }),

  getters: {
    // 获取指定页面的分页大小
    getPageSize: (state) => (pageName) => {
      return state.pageSettings[pageName]?.pageSize || state.globalDefaults.pageSize;
    },

    // 获取指定页面的分页选项
    getPageSizeOptions: (state) => (pageName) => {
      return state.pageSettings[pageName]?.pageSizeOptions || state.globalDefaults.pageSizeOptions;
    },

    // 兼容旧版本的getter（逐步废弃）
    getDefaultPageSize: (state) => state.globalDefaults.pageSize,
  },

  actions: {
    // 设置指定页面的分页大小
    setPageSize(pageName, size) {
      if (!this.pageSettings[pageName]) {
        this.pageSettings[pageName] = {
          pageSize: size,
          pageSizeOptions: this.globalDefaults.pageSizeOptions
        };
      } else {
        this.pageSettings[pageName].pageSize = size;
      }

      // 保存到 localStorage
      localStorage.setItem(`pageSize_${pageName}`, size.toString());
    },

    // 从 localStorage 加载所有页面设置
    loadSettings() {
      // 加载各个页面的设置
      Object.keys(this.pageSettings).forEach(pageName => {
        const savedPageSize = localStorage.getItem(`pageSize_${pageName}`);
        if (savedPageSize) {
          const size = parseInt(savedPageSize, 10);
          const options = this.pageSettings[pageName].pageSizeOptions;
          if (options.includes(size)) {
            this.pageSettings[pageName].pageSize = size;
          }
        }
      });

      // 兼容旧版本的全局设置
      const oldSavedPageSize = localStorage.getItem('defaultPageSize');
      if (oldSavedPageSize) {
        const size = parseInt(oldSavedPageSize, 10);
        if (this.globalDefaults.pageSizeOptions.includes(size)) {
          this.globalDefaults.pageSize = size;
        }
      }
    },

    // 重置指定页面的设置
    resetPageSettings(pageName) {
      if (this.pageSettings[pageName]) {
        // 重置为默认值
        const defaultSettings = {
          emailAccounts: { pageSize: 10 },
          platforms: { pageSize: 15 },
          platformRegistrations: { pageSize: 8 },
          serviceSubscriptions: { pageSize: 12 }
        };

        this.pageSettings[pageName].pageSize = defaultSettings[pageName]?.pageSize || this.globalDefaults.pageSize;
        localStorage.removeItem(`pageSize_${pageName}`);
      }
    },

    // 重置所有设置
    resetAllSettings() {
      Object.keys(this.pageSettings).forEach(pageName => {
        this.resetPageSettings(pageName);
      });
      this.globalDefaults.pageSize = 10;
      localStorage.removeItem('defaultPageSize');
    }
  }
});
