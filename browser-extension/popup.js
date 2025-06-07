// 弹窗脚本

class PopupManager {
  constructor() {
    this.currentTab = 'login';
    this.isLoggedIn = false;
    this.accounts = [];
    this.filteredAccounts = [];
    this.currentAccount = null;
    this.passwordVisible = false;
    this.generatorInitialized = false;
    this.currentGeneratorTab = 'password';
    this.init();
  }

  init() {
    // 首先设置初始显示状态（默认显示登录页面）
    this.setInitialState();

    this.setupNavigation();
    this.setupEventListeners();
    this.checkLoginStatus();
    this.loadCurrentTabData();
  }

  setInitialState() {
    // 确保初始状态为登录页面
    console.log('🔄 设置初始状态：显示登录页面');

    // 显示登录页面
    const loginHeader = document.getElementById('login-header');
    const loginPage = document.getElementById('login-page');
    if (loginHeader) loginHeader.style.display = 'block';
    if (loginPage) loginPage.style.display = 'block';

    // 隐藏主应用界面
    const mainHeader = document.getElementById('main-header');
    const mainContent = document.getElementById('main-content');
    const mainNav = document.getElementById('main-nav');

    if (mainHeader) mainHeader.style.display = 'none';
    if (mainContent) mainContent.style.display = 'none';
    if (mainNav) mainNav.style.display = 'none';
  }

  setupNavigation() {
    const navItems = document.querySelectorAll('.nav-item');
    const tabContents = document.querySelectorAll('.tab-content');

    navItems.forEach(navItem => {
      navItem.addEventListener('click', () => {
        const tabName = navItem.dataset.tab;

        // 更新导航状态
        navItems.forEach(item => item.classList.remove('active'));
        tabContents.forEach(tc => tc.classList.remove('active'));

        navItem.classList.add('active');

        // 根据不同的标签页显示不同内容
        if (tabName === 'vault') {
          this.updateHeaderTitle('密码库');
          if (this.isLoggedIn) {
            document.getElementById('vault-tab').classList.add('active');
          } else {
            document.getElementById('login-tab').classList.add('active');
          }
        } else if (tabName === 'generator') {
          this.updateHeaderTitle('生成器');
          console.log('🔧 切换到生成器标签页');
          const generatorTab = document.getElementById('generator-tab');
          console.log('🔧 生成器标签页元素:', generatorTab);
          generatorTab.classList.add('active');
          console.log('🔧 生成器标签页类名:', generatorTab.className);
          this.showPasswordGenerator();
        } else if (tabName === 'send') {
          this.updateHeaderTitle('Send');
          // TODO: 实现Send功能
          this.showSendFeature();
        } else if (tabName === 'settings') {
          chrome.runtime.openOptionsPage();
          return;
        }

        this.currentTab = tabName;
        this.loadCurrentTabData();
      });
    });
  }

  setupEventListeners() {
    // 安全地添加事件监听器，包含错误处理
    this.safeAddEventListener('login-form', 'submit', (e) => {
      e.preventDefault();
      this.handleLogin();
    });

    this.safeAddEventListener('manual-form', 'submit', (e) => {
      e.preventDefault();
      this.handleManualAdd();
    });

    // 设置按钮 - 特别处理
    this.safeAddEventListener('settings-btn', 'click', (e) => {
      e.preventDefault(); // 阻止链接的默认行为
      console.log('🔧 设置按钮被点击，Chrome API状态:', typeof chrome, chrome?.runtime?.openOptionsPage);

      if (chrome && chrome.runtime && chrome.runtime.openOptionsPage) {
        chrome.runtime.openOptionsPage();
      } else {
        console.error('❌ Chrome API不可用');
        alert('Chrome API不可用，请在扩展环境中使用');
      }
    });

    this.safeAddEventListener('add-btn', 'click', () => {
      this.showAddForm();
    });

    this.safeAddEventListener('search-input', 'input', (e) => {
      this.handleSearch(e.target.value);
    });

    // 详情页面事件监听器
    this.safeAddEventListener('detail-back-btn', 'click', () => {
      this.showVaultTab();
    });

    this.safeAddEventListener('detail-edit-btn', 'click', () => {
      this.editCurrentAccount();
    });

    this.safeAddEventListener('toggle-password-btn', 'click', () => {
      this.togglePasswordVisibility();
    });

    this.safeAddEventListener('detail-delete-btn', 'click', () => {
      this.deleteCurrentAccount();
    });

    // 编辑页面事件监听器
    this.safeAddEventListener('edit-back-btn', 'click', () => {
      this.showAccountDetail(this.currentAccount.id);
    });

    this.safeAddEventListener('edit-save-btn', 'click', () => {
      this.saveAccountEdit();
    });

    this.safeAddEventListener('toggle-edit-password-btn', 'click', () => {
      this.toggleEditPasswordVisibility();
    });
  }

  // 安全地添加事件监听器的辅助方法
  safeAddEventListener(elementId, eventType, handler) {
    const element = document.getElementById(elementId);
    if (element) {
      element.addEventListener(eventType, handler);
      console.log(`✅ 已绑定事件监听器: ${elementId} -> ${eventType}`);
    } else {
      console.warn(`⚠️ 元素不存在: ${elementId}`);
    }
  }

  showPasswordGenerator() {
    console.log('🔧 显示密码生成器');

    // 显示生成器标签页
    const generatorTab = document.getElementById('generator-tab');
    console.log('🔧 showPasswordGenerator - 生成器标签页元素:', generatorTab);
    if (generatorTab) {
      console.log('🔧 showPasswordGenerator - 添加active类之前:', generatorTab.className);
      generatorTab.classList.add('active');
      console.log('🔧 showPasswordGenerator - 添加active类之后:', generatorTab.className);

      // 检查其他标签页的状态
      const vaultTab = document.getElementById('vault-tab');
      console.log('🔧 showPasswordGenerator - 密码库标签页状态:', vaultTab?.className);

      // 初始化生成器
      if (!this.generatorInitialized) {
        this.initPasswordGenerator();
        this.generatorInitialized = true;
      }

      // 更新按钮状态
      this.updateLengthButtons();
      this.updateWordsButtons();
      this.updateUsernameLengthButtons();

      // 生成初始密码
      this.generatePassword();
    } else {
      console.error('❌ 找不到生成器标签页元素');
    }
  }

  showSendFeature() {
    // TODO: 实现Send功能界面
    alert('Send功能开发中...');
  }

  showAddForm() {
    // 切换到手动添加标签页
    document.querySelectorAll('.nav-item').forEach(item => item.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(tc => tc.classList.remove('active'));

    document.getElementById('manual-tab').classList.add('active');
    this.currentTab = 'manual';
    this.fillCurrentSiteInfo();
  }

  handleSearch(query) {
    this.filteredAccounts = this.accounts.filter(account => {
      const searchText = query.toLowerCase();
      return (
        (account.platform_name || '').toLowerCase().includes(searchText) ||
        (account.email_address || '').toLowerCase().includes(searchText) ||
        (account.login_username || '').toLowerCase().includes(searchText)
      );
    });
    this.displayAccounts(this.filteredAccounts);
  }



  async checkLoginStatus() {
    console.log('🔍 检查登录状态...');

    try {
      const config = await this.sendMessage({ action: 'getConfig' });
      console.log('📋 配置信息:', config);

      // 检查服务器地址配置
      if (!config.serverURL) {
        console.warn('⚠️ 服务器地址未配置');
        this.updateStatus('请先在设置中配置服务器地址', 'disconnected');
      }

      if (config && config.token) {
        console.log('✅ 用户已登录，显示主应用');
        this.isLoggedIn = true;
        this.showMainApp();
      } else {
        console.log('❌ 用户未登录，显示登录页面');
        this.isLoggedIn = false;
        this.showLoginPage();
      }
    } catch (error) {
      console.error('❌ 检查登录状态失败:', error);
      // 出错时默认显示登录页面
      this.isLoggedIn = false;
      this.showLoginPage();
    }
  }

  updateStatus(message, type) {
    const statusEl = document.getElementById('status');
    statusEl.textContent = message;
    statusEl.className = `status-banner ${type}`;
    statusEl.style.display = 'block';

    // 3秒后隐藏状态栏
    setTimeout(() => {
      statusEl.style.display = 'none';
    }, 3000);
  }

  showLoginPage() {
    console.log('🔐 显示登录页面');

    // 隐藏主应用界面
    const mainHeader = document.getElementById('main-header');
    const mainContent = document.getElementById('main-content');
    const mainNav = document.getElementById('main-nav');

    if (mainHeader) mainHeader.style.display = 'none';
    if (mainContent) mainContent.style.display = 'none';
    if (mainNav) mainNav.style.display = 'none';

    // 显示登录页面
    const loginHeader = document.getElementById('login-header');
    const loginPage = document.getElementById('login-page');

    if (loginHeader) loginHeader.style.display = 'block';
    if (loginPage) loginPage.style.display = 'block';

    console.log('✅ 登录页面已显示');
  }

  showMainApp() {
    console.log('🏠 显示主应用界面');

    // 隐藏登录页面
    const loginHeader = document.getElementById('login-header');
    const loginPage = document.getElementById('login-page');

    if (loginHeader) loginHeader.style.display = 'none';
    if (loginPage) loginPage.style.display = 'none';

    // 显示主应用界面
    const mainHeader = document.getElementById('main-header');
    const mainContent = document.getElementById('main-content');
    const mainNav = document.getElementById('main-nav');

    if (mainHeader) mainHeader.style.display = 'flex';
    if (mainContent) mainContent.style.display = 'block';
    if (mainNav) mainNav.style.display = 'flex';

    // 默认显示密码库页面并加载数据
    this.showVaultTab();
    this.loadAccounts(); // 立即加载账号数据

    console.log('✅ 主应用界面已显示');
  }

  updateHeaderTitle(title) {
    const headerTitle = document.querySelector('#main-header h1');
    if (headerTitle) {
      headerTitle.textContent = title;
    }
  }

  showVaultTab() {
    document.querySelectorAll('.tab-content').forEach(tc => tc.classList.remove('active'));
    document.getElementById('vault-tab').classList.add('active');

    // 更新导航状态
    document.querySelectorAll('.nav-item').forEach(item => item.classList.remove('active'));
    document.querySelector('[data-tab="vault"]').classList.add('active');

    // 更新标题
    this.updateHeaderTitle('密码库');
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
      this.showMessage('login', '登录成功！', 'success');

      // 清空表单
      document.getElementById('login-form').reset();

      // 延迟切换到主应用
      setTimeout(() => {
        this.showMainApp();
        this.loadAccounts();
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

    // 验证密码字段
    if (data.login_password && data.login_password.trim() !== '') {
      // 检查密码长度
      if (data.login_password.trim().length < 6) {
        this.showMessage('manual', '密码长度不能少于6位', 'error');
        return;
      }
      // 检查密码长度上限
      if (data.login_password.trim().length > 128) {
        this.showMessage('manual', '密码长度不能超过128位', 'error');
        return;
      }
    }

    this.showMessage('manual', '保存中...', 'success');

    const result = await this.sendMessage({
      action: 'saveRegistration',
      data: data
    });

    if (result.success) {
      this.showMessage('manual', '账号保存成功！', 'success');
      document.getElementById('manual-form').reset();

      // 刷新账号列表
      this.loadAccounts();

      // 2秒后切换回密码库页面
      setTimeout(() => {
        this.showVaultTab();
      }, 2000);
    } else {
      this.showMessage('manual', '保存失败: ' + result.error, 'error');
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
    document.getElementById('accounts-list').innerHTML = '';
    document.getElementById('accounts-error').style.display = 'none';
    document.getElementById('empty-state').style.display = 'none';

    const result = await this.sendMessage({ action: 'getRegistrations' });

    document.getElementById('accounts-loading').style.display = 'none';

    if (result.success) {
      console.log('📋 获取到的账号列表数据:', result.data);
      this.accounts = result.data || [];
      this.filteredAccounts = [...this.accounts];
      this.displayAccounts(this.filteredAccounts);
      this.updateItemCount(this.accounts.length);
    } else {
      document.getElementById('accounts-error').textContent = '加载失败: ' + result.error;
      document.getElementById('accounts-error').style.display = 'block';
    }
  }

  updateItemCount(count) {
    document.getElementById('item-count').textContent = count;
  }

  displayAccounts(accounts) {
    const listEl = document.getElementById('accounts-list');
    const emptyStateEl = document.getElementById('empty-state');

    if (accounts.length === 0) {
      listEl.innerHTML = '';
      emptyStateEl.style.display = 'block';
    } else {
      emptyStateEl.style.display = 'none';
      listEl.innerHTML = accounts.map(account => this.createAccountItem(account)).join('');

      // 添加点击事件监听器
      listEl.querySelectorAll('.account-item').forEach((item, index) => {
        item.addEventListener('click', () => {
          this.handleAccountClick(accounts[index]);
        });
      });

      // 添加复制按钮事件监听器
      listEl.querySelectorAll('.copy-btn').forEach((btn, index) => {
        btn.addEventListener('click', (e) => {
          e.stopPropagation();
          this.copyAccountInfo(accounts[index]);
        });
      });

      // 添加更多选项按钮事件监听器
      listEl.querySelectorAll('.more-btn').forEach((btn, index) => {
        btn.addEventListener('click', (e) => {
          e.stopPropagation();
          this.showAccountOptions(accounts[index]);
        });
      });
    }
  }

  createAccountItem(account) {
    const platformName = account.platform_name || '未知平台';
    const iconLetter = platformName.charAt(0).toUpperCase();
    const isServer = platformName.match(/\d+\.\d+\.\d+\.\d+/);
    const iconClass = isServer ? 'server' : '';

    let details = '';
    if (account.email_address) {
      details = account.email_address;
    } else if (account.login_username) {
      details = account.login_username;
    } else {
      details = '无详细信息';
    }

    return `
      <div class="account-item" data-id="${account.id || ''}">
        <div class="account-icon ${iconClass}">${iconLetter}</div>
        <div class="account-info">
          <div class="account-name">${platformName}</div>
          <div class="account-details">${details}</div>
        </div>
        <div class="account-actions">
          <button class="action-btn copy-btn" title="复制">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
              <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
            </svg>
          </button>
          <button class="action-btn more-btn" title="更多选项">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 8c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z"/>
            </svg>
          </button>
        </div>
      </div>
    `;
  }

  async handleAccountClick(account) {
    console.log('点击账号:', account);
    this.currentAccount = account;
    await this.showAccountDetail(account.id);
  }

  async copyAccountInfo(account) {
    try {
      let textToCopy = '';
      if (account.login_password) {
        textToCopy = account.login_password;
      } else if (account.email_address) {
        textToCopy = account.email_address;
      } else if (account.login_username) {
        textToCopy = account.login_username;
      }

      if (textToCopy) {
        await navigator.clipboard.writeText(textToCopy);
        this.showToast('已复制到剪贴板');
      }
    } catch (error) {
      console.error('复制失败:', error);
      this.showToast('复制失败');
    }
  }

  showAccountOptions(account) {
    // TODO: 显示账号选项菜单（编辑、删除等）
    console.log('显示账号选项:', account);
  }

  showToast(message) {
    // 简单的提示消息
    const toast = document.createElement('div');
    toast.style.cssText = `
      position: fixed;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      background: rgba(0, 0, 0, 0.8);
      color: white;
      padding: 8px 16px;
      border-radius: 4px;
      font-size: 12px;
      z-index: 1000;
    `;
    toast.textContent = message;
    document.body.appendChild(toast);

    setTimeout(() => {
      document.body.removeChild(toast);
    }, 2000);
  }

  loadCurrentTabData() {
    console.log('📂 加载当前标签页数据:', this.currentTab, '登录状态:', this.isLoggedIn);

    switch (this.currentTab) {
      case 'vault':
        if (this.isLoggedIn) {
          console.log('🔄 加载密码库数据...');
          this.loadAccounts();
        }
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

        // 显示自动填充建议
        const suggestionEl = document.getElementById('auto-fill-suggestion');
        if (suggestionEl && this.isLoggedIn) {
          suggestionEl.style.display = 'block';
        }
      }
    } catch (error) {
      console.log('无法获取当前标签页信息:', error);
    }
  }

  showMessage(prefix, message, type) {
    const errorEl = document.getElementById(`${prefix}-error`);
    const successEl = document.getElementById(`${prefix}-success`);

    // 隐藏所有消息
    if (errorEl) errorEl.style.display = 'none';
    if (successEl) successEl.style.display = 'none';

    // 显示对应类型的消息
    if (type === 'error' && errorEl) {
      errorEl.textContent = message;
      errorEl.style.display = 'block';
    } else if (successEl) {
      successEl.textContent = message;
      successEl.style.display = 'block';
    }

    // 3秒后自动隐藏成功消息
    if (type === 'success' && successEl) {
      setTimeout(() => {
        successEl.style.display = 'none';
      }, 3000);
    }
  }

  async showAccountDetail(accountId) {
    console.log('🔍 显示账号详情:', accountId);

    // 切换到详情页面
    document.querySelectorAll('.tab-content').forEach(tc => tc.classList.remove('active'));
    document.getElementById('detail-tab').classList.add('active');
    this.currentTab = 'detail';

    // 显示加载状态
    document.getElementById('detail-loading').style.display = 'block';
    document.getElementById('detail-content').style.display = 'none';
    document.getElementById('detail-error').style.display = 'none';

    try {
      // 获取详细信息
      const result = await this.sendMessage({
        action: 'getRegistrationById',
        id: accountId
      });

      document.getElementById('detail-loading').style.display = 'none';

      if (result.success) {
        console.log('🔍 API返回的完整结果:', result);
        console.log('🔍 result.data的类型:', typeof result.data);
        console.log('🔍 result.data的内容:', JSON.stringify(result.data, null, 2));
        this.displayAccountDetail(result.data);
        document.getElementById('detail-content').style.display = 'block';
      } else {
        document.getElementById('detail-error').textContent = '加载详情失败: ' + result.error;
        document.getElementById('detail-error').style.display = 'block';
      }
    } catch (error) {
      console.error('获取账号详情失败:', error);
      document.getElementById('detail-loading').style.display = 'none';
      document.getElementById('detail-error').textContent = '获取详情时发生错误';
      document.getElementById('detail-error').style.display = 'block';
    }
  }

  displayAccountDetail(accountData) {
    console.log('📋 显示账号详情数据:', accountData);

    // 详细调试信息
    console.log('🔍 字段详细信息:');
    console.log('  - id:', accountData.id, '(类型:', typeof accountData.id, ')');
    console.log('  - platform_name:', accountData.platform_name, '(类型:', typeof accountData.platform_name, ')');
    console.log('  - platform_website_url:', accountData.platform_website_url, '(类型:', typeof accountData.platform_website_url, ')');
    console.log('  - email_account_id:', accountData.email_account_id, '(类型:', typeof accountData.email_account_id, ')');
    console.log('  - email_address:', accountData.email_address, '(类型:', typeof accountData.email_address, ')');
    console.log('  - login_username:', accountData.login_username, '(类型:', typeof accountData.login_username, ')');
    console.log('  - phone_number:', accountData.phone_number, '(类型:', typeof accountData.phone_number, ')');
    console.log('  - notes:', accountData.notes, '(类型:', typeof accountData.notes, ')');
    console.log('  - created_at:', accountData.created_at, '(类型:', typeof accountData.created_at, ')');

    const platformName = accountData.platform_name || '未知平台';
    const iconLetter = platformName.charAt(0).toUpperCase();
    const isServer = platformName.match(/\d+\.\d+\.\d+\.\d+/);

    // 更新页面标题
    document.getElementById('detail-title').textContent = platformName;

    // 更新平台信息
    const platformIcon = document.getElementById('detail-platform-icon');
    platformIcon.textContent = iconLetter;
    platformIcon.className = `platform-icon ${isServer ? 'server' : ''}`;

    document.getElementById('detail-platform-name').textContent = platformName;
    document.getElementById('detail-platform-url').textContent = accountData.platform_website_url || platformName;

    // 更新字段信息
    this.updateDetailField('detail-email', accountData.email_address);
    this.updateDetailField('detail-username', accountData.login_username);
    this.updateDetailField('detail-phone', accountData.phone_number);
    this.updateDetailField('detail-notes', accountData.notes);

    // 格式化创建时间
    const createdAt = accountData.created_at ?
      new Date(accountData.created_at).toLocaleString('zh-CN') :
      '未知';
    this.updateDetailField('detail-created', createdAt);

    // 重置密码显示状态
    this.passwordVisible = false;
    document.getElementById('detail-password').textContent = '••••••••';
    document.getElementById('detail-password').className = 'field-text password-hidden';

    // 设置复制按钮事件
    this.setupDetailCopyButtons(accountData);
  }

  updateDetailField(elementId, value) {
    const element = document.getElementById(elementId);
    console.log(`🔧 更新字段 ${elementId}:`, value, '(类型:', typeof value, ', 长度:', value?.length, ')');

    if (value && value.trim() !== '') {
      element.textContent = value;
      element.className = 'field-text';
      console.log(`✅ ${elementId} 设置为:`, value);
    } else {
      element.textContent = '未设置';
      element.className = 'field-text empty';
      console.log(`❌ ${elementId} 为空，显示"未设置"`);
    }
  }

  setupDetailCopyButtons(accountData) {
    // 移除之前的事件监听器
    document.querySelectorAll('.copy-field-btn').forEach(btn => {
      btn.replaceWith(btn.cloneNode(true));
    });

    // 重新添加事件监听器
    document.querySelectorAll('.copy-field-btn').forEach(btn => {
      btn.addEventListener('click', async (e) => {
        e.stopPropagation();
        const field = btn.dataset.field;
        await this.copyDetailField(field, accountData);
      });
    });
  }

  async copyDetailField(field, accountData) {
    let textToCopy = '';

    switch (field) {
      case 'email':
        textToCopy = accountData.email_address || '';
        break;
      case 'username':
        textToCopy = accountData.login_username || '';
        break;
      case 'password':
        if (this.passwordVisible) {
          textToCopy = document.getElementById('detail-password').textContent;
        } else {
          // 需要获取密码
          try {
            const result = await this.sendMessage({
              action: 'getRegistrationPassword',
              id: accountData.id
            });
            if (result.success) {
              textToCopy = result.data.password;
            } else {
              this.showToast('获取密码失败');
              return;
            }
          } catch (error) {
            this.showToast('获取密码失败');
            return;
          }
        }
        break;
      case 'phone':
        textToCopy = accountData.phone_number || '';
        break;
      case 'notes':
        textToCopy = accountData.notes || '';
        break;
    }

    if (textToCopy) {
      try {
        await navigator.clipboard.writeText(textToCopy);
        this.showToast('已复制到剪贴板');
      } catch (error) {
        console.error('复制失败:', error);
        this.showToast('复制失败');
      }
    } else {
      this.showToast('该字段为空');
    }
  }

  async togglePasswordVisibility() {
    if (this.passwordVisible) {
      // 隐藏密码
      document.getElementById('detail-password').textContent = '••••••••';
      document.getElementById('detail-password').className = 'field-text password-hidden';
      this.passwordVisible = false;
    } else {
      // 显示密码
      try {
        const result = await this.sendMessage({
          action: 'getRegistrationPassword',
          id: this.currentAccount.id
        });

        if (result.success) {
          document.getElementById('detail-password').textContent = result.data.password;
          document.getElementById('detail-password').className = 'field-text';
          this.passwordVisible = true;
        } else {
          this.showToast('获取密码失败: ' + result.error);
        }
      } catch (error) {
        console.error('获取密码失败:', error);
        this.showToast('获取密码失败');
      }
    }
  }

  editCurrentAccount() {
    console.log('✏️ 编辑账号:', this.currentAccount);
    this.showEditPage();
  }

  async showEditPage() {
    // 切换到编辑页面
    document.querySelectorAll('.tab-content').forEach(tc => tc.classList.remove('active'));
    document.getElementById('edit-tab').classList.add('active');
    this.currentTab = 'edit';

    // 更新页面标题
    document.getElementById('edit-title').textContent = `编辑 ${this.currentAccount.platform_name || '账号'}`;

    // 填充表单数据
    this.fillEditForm();

    // 隐藏消息
    document.getElementById('edit-error').style.display = 'none';
    document.getElementById('edit-success').style.display = 'none';
  }

  fillEditForm() {
    if (!this.currentAccount) return;

    // 填充基本信息
    document.getElementById('edit-platform-name').value = this.currentAccount.platform_name || '';
    document.getElementById('edit-email').value = this.currentAccount.email_address || '';
    document.getElementById('edit-username').value = this.currentAccount.login_username || '';
    document.getElementById('edit-phone').value = this.currentAccount.phone_number || '';
    document.getElementById('edit-notes').value = this.currentAccount.notes || '';

    // 密码字段留空（用户需要输入新密码才会更新）
    document.getElementById('edit-password').value = '';
  }

  toggleEditPasswordVisibility() {
    const passwordInput = document.getElementById('edit-password');
    if (passwordInput.type === 'password') {
      passwordInput.type = 'text';
    } else {
      passwordInput.type = 'password';
    }
  }

  async saveAccountEdit() {
    console.log('💾 保存账号编辑');

    const form = document.getElementById('edit-form');
    const formData = new FormData(form);
    const data = Object.fromEntries(formData.entries());

    // 移除不允许编辑的字段（邮箱地址和用户名）
    delete data.email_address;
    delete data.login_username;
    console.log('🔒 已移除不可编辑字段: email_address, login_username');

    // 验证密码字段
    if (data.login_password && data.login_password.trim() !== '') {
      // 检查密码长度
      if (data.login_password.trim().length < 6) {
        document.getElementById('edit-error').textContent = '密码长度不能少于6位';
        document.getElementById('edit-error').style.display = 'block';
        return;
      }
      // 检查密码长度上限
      if (data.login_password.trim().length > 128) {
        document.getElementById('edit-error').textContent = '密码长度不能超过128位';
        document.getElementById('edit-error').style.display = 'block';
        return;
      }
    } else {
      // 移除空的密码字段
      delete data.login_password;
    }

    // 显示保存状态
    document.getElementById('edit-loading').style.display = 'block';
    document.getElementById('edit-error').style.display = 'none';
    document.getElementById('edit-success').style.display = 'none';

    try {
      const result = await this.sendMessage({
        action: 'updateRegistration',
        id: this.currentAccount.id,
        data: data
      });

      document.getElementById('edit-loading').style.display = 'none';

      if (result.success) {
        document.getElementById('edit-success').textContent = '保存成功！';
        document.getElementById('edit-success').style.display = 'block';

        // 更新当前账号数据
        Object.assign(this.currentAccount, data);

        // 刷新账号列表
        this.loadAccounts();

        // 2秒后返回详情页面
        setTimeout(() => {
          this.showAccountDetail(this.currentAccount.id);
        }, 2000);
      } else {
        document.getElementById('edit-error').textContent = '保存失败: ' + result.error;
        document.getElementById('edit-error').style.display = 'block';
      }
    } catch (error) {
      console.error('保存编辑失败:', error);
      document.getElementById('edit-loading').style.display = 'none';
      document.getElementById('edit-error').textContent = '保存时发生错误';
      document.getElementById('edit-error').style.display = 'block';
    }
  }

  async deleteCurrentAccount() {
    if (!this.currentAccount) return;

    const platformName = this.currentAccount.platform_name || '此账号';
    if (!confirm(`确定要删除 ${platformName} 吗？\n\n此操作不可撤销！`)) {
      return;
    }

    console.log('🗑️ 删除账号:', this.currentAccount);

    try {
      const result = await this.sendMessage({
        action: 'deleteRegistration',
        id: this.currentAccount.id
      });

      if (result.success) {
        this.showToast('账号删除成功');

        // 刷新账号列表
        this.loadAccounts();

        // 返回密码库列表
        setTimeout(() => {
          this.showVaultTab();
        }, 1000);
      } else {
        this.showToast('删除失败: ' + result.error);
      }
    } catch (error) {
      console.error('删除账号失败:', error);
      this.showToast('删除时发生错误');
    }
  }

  sendMessage(message) {
    return new Promise((resolve) => {
      chrome.runtime.sendMessage(message, resolve);
    });
  }

  // 密码生成器相关方法
  initPasswordGenerator() {
    console.log('🔧 初始化密码生成器');

    // 加载保存的生成器设置
    this.loadGeneratorSettings();

    // 设置生成器标签页切换
    this.setupGeneratorTabs();

    // 设置密码生成器事件监听器
    this.setupPasswordGeneratorEvents();

    // 设置密码短语生成器事件监听器
    this.setupPassphraseGeneratorEvents();

    // 设置用户名生成器事件监听器
    this.setupUsernameGeneratorEvents();
  }

  setupGeneratorTabs() {
    const tabItems = document.querySelectorAll('.generator-tab-item');
    const contents = document.querySelectorAll('.generator-content');

    tabItems.forEach(tab => {
      tab.addEventListener('click', () => {
        const tabName = tab.dataset.generatorTab;

        // 更新标签页状态
        tabItems.forEach(t => t.classList.remove('active'));
        contents.forEach(c => c.classList.remove('active'));

        tab.classList.add('active');
        document.getElementById(`${tabName}-generator`).classList.add('active');

        this.currentGeneratorTab = tabName;

        // 根据当前标签页生成内容
        if (tabName === 'password') {
          this.generatePassword();
        } else if (tabName === 'passphrase') {
          this.generatePassphrase();
        } else if (tabName === 'username') {
          this.generateUsername();
        }
      });
    });
  }

  setupPasswordGeneratorEvents() {
    // 长度滑块
    const lengthSlider = document.getElementById('length-slider');
    const lengthValue = document.getElementById('length-value');

    if (lengthSlider && lengthValue) {
      lengthSlider.addEventListener('input', (e) => {
        lengthValue.textContent = e.target.value;
        this.updateLengthButtons();
        this.generatePassword();
        this.saveGeneratorSettings();
      });
    }

    // 长度增减按钮
    this.safeAddEventListener('length-decrease', 'click', () => {
      this.adjustLength(-1);
    });

    this.safeAddEventListener('length-increase', 'click', () => {
      this.adjustLength(1);
    });

    // 复选框
    const checkboxes = ['include-uppercase', 'include-lowercase', 'include-numbers', 'include-symbols', 'avoid-ambiguous'];
    checkboxes.forEach(id => {
      const checkbox = document.getElementById(id);
      if (checkbox) {
        checkbox.addEventListener('change', () => {
          this.generatePassword();
          this.saveGeneratorSettings();
        });
      }
    });

    // 最少个数输入
    const minInputs = ['min-numbers', 'min-symbols'];
    minInputs.forEach(id => {
      const input = document.getElementById(id);
      if (input) {
        input.addEventListener('input', () => {
          this.generatePassword();
          this.saveGeneratorSettings();
        });
      }
    });

    // 刷新和复制按钮
    this.safeAddEventListener('refresh-password', 'click', () => {
      this.generatePassword();
    });

    this.safeAddEventListener('copy-password', 'click', () => {
      this.copyToClipboard(document.getElementById('generated-password').textContent);
    });
  }

  setupPassphraseGeneratorEvents() {
    // 单词数量滑块
    const wordsSlider = document.getElementById('words-slider');
    const wordsValue = document.getElementById('words-value');

    if (wordsSlider && wordsValue) {
      wordsSlider.addEventListener('input', (e) => {
        wordsValue.textContent = e.target.value;
        this.updateWordsButtons();
        this.generatePassphrase();
        this.saveGeneratorSettings();
      });
    }

    // 单词数量增减按钮
    this.safeAddEventListener('words-decrease', 'click', () => {
      this.adjustWords(-1);
    });

    this.safeAddEventListener('words-increase', 'click', () => {
      this.adjustWords(1);
    });

    // 分隔符输入
    const separator = document.getElementById('separator');
    if (separator) {
      separator.addEventListener('input', () => {
        this.generatePassphrase();
        this.saveGeneratorSettings();
      });
    }

    // 首字母大写复选框
    const capitalizeWords = document.getElementById('capitalize-words');
    if (capitalizeWords) {
      capitalizeWords.addEventListener('change', () => {
        this.generatePassphrase();
        this.saveGeneratorSettings();
      });
    }

    // 刷新和复制按钮
    this.safeAddEventListener('refresh-passphrase', 'click', () => {
      this.generatePassphrase();
    });

    this.safeAddEventListener('copy-passphrase', 'click', () => {
      this.copyToClipboard(document.getElementById('generated-passphrase').textContent);
    });
  }

  setupUsernameGeneratorEvents() {
    // 长度滑块
    const lengthSlider = document.getElementById('username-length-slider');
    const lengthValue = document.getElementById('username-length-value');

    if (lengthSlider && lengthValue) {
      lengthSlider.addEventListener('input', (e) => {
        lengthValue.textContent = e.target.value;
        this.updateUsernameLengthButtons();
        this.generateUsername();
        this.saveGeneratorSettings();
      });
    }

    // 用户名长度增减按钮
    this.safeAddEventListener('username-length-decrease', 'click', () => {
      this.adjustUsernameLength(-1);
    });

    this.safeAddEventListener('username-length-increase', 'click', () => {
      this.adjustUsernameLength(1);
    });

    // 包含数字复选框
    const includeNumbers = document.getElementById('include-numbers-username');
    if (includeNumbers) {
      includeNumbers.addEventListener('change', () => {
        this.generateUsername();
        this.saveGeneratorSettings();
      });
    }

    // 刷新和复制按钮
    this.safeAddEventListener('refresh-username', 'click', () => {
      this.generateUsername();
    });

    this.safeAddEventListener('copy-username', 'click', () => {
      this.copyToClipboard(document.getElementById('generated-username').textContent);
    });
  }

  generatePassword() {
    const length = parseInt(document.getElementById('length-slider')?.value || '8');
    const includeUppercase = document.getElementById('include-uppercase')?.checked || false;
    const includeLowercase = document.getElementById('include-lowercase')?.checked || false;
    const includeNumbers = document.getElementById('include-numbers')?.checked || false;
    const includeSymbols = document.getElementById('include-symbols')?.checked || false;
    const avoidAmbiguous = document.getElementById('avoid-ambiguous')?.checked || false;
    const minNumbers = parseInt(document.getElementById('min-numbers')?.value || '1');
    const minSymbols = parseInt(document.getElementById('min-symbols')?.value || '1');

    let charset = '';

    if (includeUppercase) {
      charset += avoidAmbiguous ? 'ABCDEFGHJKMNPQRSTUVWXYZ' : 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    }
    if (includeLowercase) {
      charset += avoidAmbiguous ? 'abcdefghijkmnpqrstuvwxyz' : 'abcdefghijklmnopqrstuvwxyz';
    }
    if (includeNumbers) {
      charset += avoidAmbiguous ? '23456789' : '0123456789';
    }
    if (includeSymbols) {
      charset += '!@#$%^&*()_+-=[]{}|;:,.<>?';
    }

    if (!charset) {
      charset = 'abcdefghijklmnopqrstuvwxyz'; // 默认字符集
    }

    let password = '';
    let numberCount = 0;
    let symbolCount = 0;

    // 生成密码
    for (let i = 0; i < length; i++) {
      const randomIndex = Math.floor(Math.random() * charset.length);
      const char = charset[randomIndex];
      password += char;

      if ('0123456789'.includes(char)) numberCount++;
      if ('!@#$%^&*()_+-=[]{}|;:,.<>?'.includes(char)) symbolCount++;
    }

    // 确保满足最少个数要求
    if (includeNumbers && numberCount < minNumbers) {
      const numbersNeeded = minNumbers - numberCount;
      const numberChars = avoidAmbiguous ? '23456789' : '0123456789';
      for (let i = 0; i < numbersNeeded && i < password.length; i++) {
        const randomPos = Math.floor(Math.random() * password.length);
        const randomNumber = numberChars[Math.floor(Math.random() * numberChars.length)];
        password = password.substring(0, randomPos) + randomNumber + password.substring(randomPos + 1);
      }
    }

    if (includeSymbols && symbolCount < minSymbols) {
      const symbolsNeeded = minSymbols - symbolCount;
      const symbolChars = '!@#$%^&*()_+-=[]{}|;:,.<>?';
      for (let i = 0; i < symbolsNeeded && i < password.length; i++) {
        const randomPos = Math.floor(Math.random() * password.length);
        const randomSymbol = symbolChars[Math.floor(Math.random() * symbolChars.length)];
        password = password.substring(0, randomPos) + randomSymbol + password.substring(randomPos + 1);
      }
    }

    const passwordOutput = document.getElementById('generated-password');
    if (passwordOutput) {
      passwordOutput.textContent = password;
    }
  }

  generatePassphrase() {
    const wordCount = parseInt(document.getElementById('words-slider')?.value || '4');
    const separator = document.getElementById('separator')?.value || '-';
    const capitalizeWords = document.getElementById('capitalize-words')?.checked || false;

    // 常用单词列表
    const words = [
      'apple', 'banana', 'cherry', 'dragon', 'elephant', 'forest', 'guitar', 'house',
      'island', 'jungle', 'kitchen', 'lemon', 'mountain', 'ocean', 'piano', 'queen',
      'river', 'sunset', 'tiger', 'umbrella', 'village', 'window', 'yellow', 'zebra',
      'bridge', 'castle', 'dream', 'eagle', 'flower', 'garden', 'happy', 'ice',
      'journey', 'knight', 'light', 'magic', 'nature', 'orange', 'peace', 'quiet',
      'rainbow', 'star', 'tree', 'unique', 'victory', 'water', 'extra', 'youth',
      'adventure', 'beautiful', 'creative', 'delicious', 'energy', 'freedom', 'gentle',
      'harmony', 'inspire', 'joyful', 'kindness', 'lovely', 'mystery', 'natural'
    ];

    let selectedWords = [];
    for (let i = 0; i < wordCount; i++) {
      const randomIndex = Math.floor(Math.random() * words.length);
      let word = words[randomIndex];

      if (capitalizeWords) {
        word = word.charAt(0).toUpperCase() + word.slice(1);
      }

      selectedWords.push(word);
    }

    const passphrase = selectedWords.join(separator);

    const passphraseOutput = document.getElementById('generated-passphrase');
    if (passphraseOutput) {
      passphraseOutput.textContent = passphrase;
    }
  }

  generateUsername() {
    const length = parseInt(document.getElementById('username-length-slider')?.value || '6');
    const includeNumbers = document.getElementById('include-numbers-username')?.checked || false;

    const prefixes = ['user', 'guest', 'player', 'member', 'admin', 'super', 'pro', 'elite'];
    const adjectives = ['cool', 'fast', 'smart', 'brave', 'quick', 'strong', 'bright', 'sharp'];
    const nouns = ['cat', 'dog', 'bird', 'fish', 'lion', 'bear', 'wolf', 'fox'];

    let username = '';

    if (length <= 8) {
      // 短用户名：只使用前缀
      const prefix = prefixes[Math.floor(Math.random() * prefixes.length)];
      username = prefix;

      if (includeNumbers && username.length < length) {
        const numbersToAdd = Math.min(4, length - username.length);
        for (let i = 0; i < numbersToAdd; i++) {
          username += Math.floor(Math.random() * 10);
        }
      }
    } else {
      // 长用户名：使用组合
      const prefix = prefixes[Math.floor(Math.random() * prefixes.length)];
      const adjective = adjectives[Math.floor(Math.random() * adjectives.length)];
      const noun = nouns[Math.floor(Math.random() * nouns.length)];

      username = prefix + '_' + adjective + '_' + noun;

      if (includeNumbers) {
        const year = new Date().getFullYear();
        const randomNum = Math.floor(Math.random() * 1000);
        username += '_' + (Math.random() > 0.5 ? year : randomNum);
      }
    }

    // 调整长度
    if (username.length > length) {
      username = username.substring(0, length);
    } else if (username.length < length) {
      const chars = 'abcdefghijklmnopqrstuvwxyz';
      while (username.length < length) {
        username += chars[Math.floor(Math.random() * chars.length)];
      }
    }

    const usernameOutput = document.getElementById('generated-username');
    if (usernameOutput) {
      usernameOutput.textContent = username;
    }
  }

  copyToClipboard(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(() => {
        this.showToast('已复制到剪贴板');
      }).catch(err => {
        console.error('复制失败:', err);
        this.fallbackCopyToClipboard(text);
      });
    } else {
      this.fallbackCopyToClipboard(text);
    }
  }

  fallbackCopyToClipboard(text) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    try {
      document.execCommand('copy');
      this.showToast('已复制到剪贴板');
    } catch (err) {
      console.error('复制失败:', err);
      this.showToast('复制失败');
    }

    document.body.removeChild(textArea);
  }

  // 长度调整方法
  adjustLength(delta) {
    const slider = document.getElementById('length-slider');
    const valueEl = document.getElementById('length-value');
    if (slider && valueEl) {
      const currentValue = parseInt(slider.value);
      const newValue = Math.max(parseInt(slider.min), Math.min(parseInt(slider.max), currentValue + delta));
      slider.value = newValue;
      valueEl.textContent = newValue;
      this.updateLengthButtons();
      this.generatePassword();
      this.saveGeneratorSettings();
    }
  }

  updateLengthButtons() {
    const slider = document.getElementById('length-slider');
    const decreaseBtn = document.getElementById('length-decrease');
    const increaseBtn = document.getElementById('length-increase');

    if (slider && decreaseBtn && increaseBtn) {
      const currentValue = parseInt(slider.value);
      decreaseBtn.disabled = currentValue <= parseInt(slider.min);
      increaseBtn.disabled = currentValue >= parseInt(slider.max);
    }
  }

  // 单词数量调整方法
  adjustWords(delta) {
    const slider = document.getElementById('words-slider');
    const valueEl = document.getElementById('words-value');
    if (slider && valueEl) {
      const currentValue = parseInt(slider.value);
      const newValue = Math.max(parseInt(slider.min), Math.min(parseInt(slider.max), currentValue + delta));
      slider.value = newValue;
      valueEl.textContent = newValue;
      this.updateWordsButtons();
      this.generatePassphrase();
      this.saveGeneratorSettings();
    }
  }

  updateWordsButtons() {
    const slider = document.getElementById('words-slider');
    const decreaseBtn = document.getElementById('words-decrease');
    const increaseBtn = document.getElementById('words-increase');

    if (slider && decreaseBtn && increaseBtn) {
      const currentValue = parseInt(slider.value);
      decreaseBtn.disabled = currentValue <= parseInt(slider.min);
      increaseBtn.disabled = currentValue >= parseInt(slider.max);
    }
  }

  // 用户名长度调整方法
  adjustUsernameLength(delta) {
    const slider = document.getElementById('username-length-slider');
    const valueEl = document.getElementById('username-length-value');
    if (slider && valueEl) {
      const currentValue = parseInt(slider.value);
      const newValue = Math.max(parseInt(slider.min), Math.min(parseInt(slider.max), currentValue + delta));
      slider.value = newValue;
      valueEl.textContent = newValue;
      this.updateUsernameLengthButtons();
      this.generateUsername();
      this.saveGeneratorSettings();
    }
  }

  updateUsernameLengthButtons() {
    const slider = document.getElementById('username-length-slider');
    const decreaseBtn = document.getElementById('username-length-decrease');
    const increaseBtn = document.getElementById('username-length-increase');

    if (slider && decreaseBtn && increaseBtn) {
      const currentValue = parseInt(slider.value);
      decreaseBtn.disabled = currentValue <= parseInt(slider.min);
      increaseBtn.disabled = currentValue >= parseInt(slider.max);
    }
  }

  // 生成器设置保存和加载
  async loadGeneratorSettings() {
    try {
      const settings = await this.getGeneratorSettings();
      console.log('🔧 加载生成器设置:', settings);

      // 密码生成器设置
      if (settings.passwordLength !== undefined) {
        const lengthSlider = document.getElementById('length-slider');
        const lengthValue = document.getElementById('length-value');
        if (lengthSlider && lengthValue) {
          lengthSlider.value = settings.passwordLength;
          lengthValue.textContent = settings.passwordLength;
        }
      }

      if (settings.includeUppercase !== undefined) {
        const checkbox = document.getElementById('include-uppercase');
        if (checkbox) checkbox.checked = settings.includeUppercase;
      }

      if (settings.includeLowercase !== undefined) {
        const checkbox = document.getElementById('include-lowercase');
        if (checkbox) checkbox.checked = settings.includeLowercase;
      }

      if (settings.includeNumbers !== undefined) {
        const checkbox = document.getElementById('include-numbers');
        if (checkbox) checkbox.checked = settings.includeNumbers;
      }

      if (settings.includeSymbols !== undefined) {
        const checkbox = document.getElementById('include-symbols');
        if (checkbox) checkbox.checked = settings.includeSymbols;
      }

      if (settings.avoidAmbiguous !== undefined) {
        const checkbox = document.getElementById('avoid-ambiguous');
        if (checkbox) checkbox.checked = settings.avoidAmbiguous;
      }

      if (settings.minNumbers !== undefined) {
        const input = document.getElementById('min-numbers');
        if (input) input.value = settings.minNumbers;
      }

      if (settings.minSymbols !== undefined) {
        const input = document.getElementById('min-symbols');
        if (input) input.value = settings.minSymbols;
      }

      // 密码短语生成器设置
      if (settings.wordsCount !== undefined) {
        const wordsSlider = document.getElementById('words-slider');
        const wordsValue = document.getElementById('words-value');
        if (wordsSlider && wordsValue) {
          wordsSlider.value = settings.wordsCount;
          wordsValue.textContent = settings.wordsCount;
        }
      }

      if (settings.separator !== undefined) {
        const input = document.getElementById('separator');
        if (input) input.value = settings.separator;
      }

      if (settings.capitalizeWords !== undefined) {
        const checkbox = document.getElementById('capitalize-words');
        if (checkbox) checkbox.checked = settings.capitalizeWords;
      }

      // 用户名生成器设置
      if (settings.usernameLength !== undefined) {
        const lengthSlider = document.getElementById('username-length-slider');
        const lengthValue = document.getElementById('username-length-value');
        if (lengthSlider && lengthValue) {
          lengthSlider.value = settings.usernameLength;
          lengthValue.textContent = settings.usernameLength;
        }
      }

      if (settings.includeNumbersUsername !== undefined) {
        const checkbox = document.getElementById('include-numbers-username');
        if (checkbox) checkbox.checked = settings.includeNumbersUsername;
      }

    } catch (error) {
      console.error('❌ 加载生成器设置失败:', error);
    }
  }

  async saveGeneratorSettings() {
    try {
      const settings = {
        // 密码生成器设置
        passwordLength: parseInt(document.getElementById('length-slider')?.value || '8'),
        includeUppercase: document.getElementById('include-uppercase')?.checked || false,
        includeLowercase: document.getElementById('include-lowercase')?.checked || false,
        includeNumbers: document.getElementById('include-numbers')?.checked || false,
        includeSymbols: document.getElementById('include-symbols')?.checked || false,
        avoidAmbiguous: document.getElementById('avoid-ambiguous')?.checked || false,
        minNumbers: parseInt(document.getElementById('min-numbers')?.value || '1'),
        minSymbols: parseInt(document.getElementById('min-symbols')?.value || '1'),

        // 密码短语生成器设置
        wordsCount: parseInt(document.getElementById('words-slider')?.value || '4'),
        separator: document.getElementById('separator')?.value || '-',
        capitalizeWords: document.getElementById('capitalize-words')?.checked || false,

        // 用户名生成器设置
        usernameLength: parseInt(document.getElementById('username-length-slider')?.value || '6'),
        includeNumbersUsername: document.getElementById('include-numbers-username')?.checked || false
      };

      console.log('💾 保存生成器设置:', settings);
      await this.storeGeneratorSettings(settings);
    } catch (error) {
      console.error('❌ 保存生成器设置失败:', error);
    }
  }

  getGeneratorSettings() {
    return new Promise((resolve) => {
      if (chrome && chrome.storage && chrome.storage.local) {
        chrome.storage.local.get(['generatorSettings'], (result) => {
          resolve(result.generatorSettings || {});
        });
      } else {
        // 模拟环境下使用localStorage
        const settings = localStorage.getItem('generatorSettings');
        resolve(settings ? JSON.parse(settings) : {});
      }
    });
  }

  storeGeneratorSettings(settings) {
    return new Promise((resolve) => {
      if (chrome && chrome.storage && chrome.storage.local) {
        chrome.storage.local.set({ generatorSettings: settings }, resolve);
      } else {
        // 模拟环境下使用localStorage
        localStorage.setItem('generatorSettings', JSON.stringify(settings));
        resolve();
      }
    });
  }
}

// 初始化弹窗管理器
document.addEventListener('DOMContentLoaded', () => {
  new PopupManager();
});
