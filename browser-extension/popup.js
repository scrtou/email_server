// 弹窗脚本

class PopupManager {
  constructor() {
    this.currentTab = 'login';
    this.isLoggedIn = false;
    this.init();
  }

  init() {
    this.setupTabSwitching();
    this.setupEventListeners();
    this.checkLoginStatus();
    this.loadCurrentTabData();
  }

  setupTabSwitching() {
    const tabs = document.querySelectorAll('.tab');
    const tabContents = document.querySelectorAll('.tab-content');

    tabs.forEach(tab => {
      tab.addEventListener('click', () => {
        const tabName = tab.dataset.tab;
        
        // 更新标签页状态
        tabs.forEach(t => t.classList.remove('active'));
        tabContents.forEach(tc => tc.classList.remove('active'));
        
        tab.classList.add('active');
        document.getElementById(`${tabName}-tab`).classList.add('active');
        
        this.currentTab = tabName;
        this.loadCurrentTabData();
      });
    });
  }

  setupEventListeners() {
    // 登录表单
    document.getElementById('login-form').addEventListener('submit', (e) => {
      e.preventDefault();
      this.handleLogin();
    });

    // 手动添加表单
    document.getElementById('manual-form').addEventListener('submit', (e) => {
      e.preventDefault();
      this.handleManualAdd();
    });

    // 设置按钮
    document.getElementById('settings-btn').addEventListener('click', () => {
      chrome.runtime.openOptionsPage();
    });

    // 刷新账号列表
    document.getElementById('refresh-accounts').addEventListener('click', () => {
      this.loadAccounts();
    });
  }

  async checkLoginStatus() {
    const config = await this.sendMessage({ action: 'getConfig' });
    
    if (config.token) {
      this.isLoggedIn = true;
      this.updateStatus('已连接到服务器', 'connected');
    } else {
      this.isLoggedIn = false;
      this.updateStatus('未连接到服务器', 'disconnected');
    }
  }

  updateStatus(message, type) {
    const statusEl = document.getElementById('status');
    statusEl.textContent = message;
    statusEl.className = `status ${type}`;
  }

  async handleLogin() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    
    this.showMessage('login', '登录中...', 'success');
    
    const result = await this.sendMessage({
      action: 'login',
      username,
      password
    });

    if (result.success) {
      this.isLoggedIn = true;
      this.updateStatus('已连接到服务器', 'connected');
      this.showMessage('login', '登录成功！', 'success');
      
      // 清空表单
      document.getElementById('login-form').reset();
      
      // 切换到账号列表标签页
      setTimeout(() => {
        document.querySelector('[data-tab="accounts"]').click();
      }, 1000);
    } else {
      this.showMessage('login', '登录失败: ' + result.error, 'error');
    }
  }

  async handleManualAdd() {
    if (!this.isLoggedIn) {
      this.showMessage('manual', '请先登录', 'error');
      return;
    }

    const formData = new FormData(document.getElementById('manual-form'));
    const data = Object.fromEntries(formData.entries());
    
    // 验证必填字段
    if (!data.platform_name) {
      this.showMessage('manual', '平台名称不能为空', 'error');
      return;
    }

    this.showMessage('manual', '添加中...', 'success');
    
    const result = await this.sendMessage({
      action: 'saveRegistration',
      data: data
    });

    if (result.success) {
      this.showMessage('manual', '账号添加成功！', 'success');
      document.getElementById('manual-form').reset();
    } else {
      this.showMessage('manual', '添加失败: ' + result.error, 'error');
    }
  }

  async loadAccounts() {
    if (!this.isLoggedIn) {
      document.getElementById('accounts-error').textContent = '请先登录';
      document.getElementById('accounts-error').style.display = 'block';
      document.getElementById('accounts-loading').style.display = 'none';
      return;
    }

    document.getElementById('accounts-loading').style.display = 'block';
    document.getElementById('accounts-list').style.display = 'none';
    document.getElementById('accounts-error').style.display = 'none';

    const result = await this.sendMessage({ action: 'getRegistrations' });

    document.getElementById('accounts-loading').style.display = 'none';

    if (result.success) {
      this.displayAccounts(result.data.data || []);
    } else {
      document.getElementById('accounts-error').textContent = '加载失败: ' + result.error;
      document.getElementById('accounts-error').style.display = 'block';
    }
  }

  displayAccounts(accounts) {
    const listEl = document.getElementById('accounts-list');
    
    if (accounts.length === 0) {
      listEl.innerHTML = '<div class="registration-item">暂无账号记录</div>';
    } else {
      listEl.innerHTML = accounts.map(account => `
        <div class="registration-item">
          <div class="platform">${account.platform_name || '未知平台'}</div>
          <div class="details">
            ${account.email_address ? `邮箱: ${account.email_address}` : ''}
            ${account.login_username ? `用户名: ${account.login_username}` : ''}
            ${account.created_at ? `创建时间: ${new Date(account.created_at).toLocaleString()}` : ''}
          </div>
        </div>
      `).join('');
    }
    
    listEl.style.display = 'block';
  }

  loadCurrentTabData() {
    switch (this.currentTab) {
      case 'accounts':
        this.loadAccounts();
        break;
      case 'manual':
        // 自动填充当前网站信息
        this.fillCurrentSiteInfo();
        break;
    }
  }

  async fillCurrentSiteInfo() {
    try {
      const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
      if (tab && tab.url) {
        const url = new URL(tab.url);
        const platformName = url.hostname.replace(/^(www\.|m\.|mobile\.)/, '');
        document.getElementById('manual-platform').value = platformName;
      }
    } catch (error) {
      console.log('无法获取当前标签页信息:', error);
    }
  }

  showMessage(prefix, message, type) {
    const errorEl = document.getElementById(`${prefix}-error`);
    const successEl = document.getElementById(`${prefix}-success`);
    
    // 隐藏所有消息
    errorEl.style.display = 'none';
    successEl.style.display = 'none';
    
    // 显示对应类型的消息
    if (type === 'error') {
      errorEl.textContent = message;
      errorEl.style.display = 'block';
    } else {
      successEl.textContent = message;
      successEl.style.display = 'block';
    }
    
    // 3秒后自动隐藏成功消息
    if (type === 'success') {
      setTimeout(() => {
        successEl.style.display = 'none';
      }, 3000);
    }
  }

  sendMessage(message) {
    return new Promise((resolve) => {
      chrome.runtime.sendMessage(message, resolve);
    });
  }
}

// 初始化弹窗管理器
document.addEventListener('DOMContentLoaded', () => {
  new PopupManager();
});
