// 设置页面脚本

class OptionsManager {
  constructor() {
    this.init();
  }

  init() {
    this.loadSettings();
    this.setupEventListeners();
  }

  setupEventListeners() {
    // 保存设置表单
    document.getElementById('settings-form').addEventListener('submit', (e) => {
      e.preventDefault();
      this.saveSettings();
    });

    // 测试连接按钮
    document.getElementById('test-connection').addEventListener('click', () => {
      this.testConnection();
    });

    // 退出登录按钮
    document.getElementById('logout-btn').addEventListener('click', () => {
      this.logout();
    });

    // 高级设置变化监听
    const advancedInputs = document.querySelectorAll('.advanced-settings input');
    advancedInputs.forEach(input => {
      input.addEventListener('change', () => {
        this.saveAdvancedSettings();
      });
    });
  }

  async loadSettings() {
    try {
      const settings = await this.getStoredSettings();
      
      // 基本设置
      if (settings.serverURL) {
        document.getElementById('server-url').value = settings.serverURL;
      }
      if (settings.username) {
        document.getElementById('username').value = settings.username;
      }
      if (settings.password) {
        document.getElementById('password').value = settings.password;
      }

      // 高级设置
      document.getElementById('auto-detect').checked = settings.autoDetect !== false;
      document.getElementById('show-notifications').checked = settings.showNotifications !== false;
      document.getElementById('auto-save').checked = settings.autoSave === true;
      
      if (settings.excludedSites) {
        document.getElementById('excluded-sites').value = settings.excludedSites;
      }

    } catch (error) {
      this.showStatus('加载设置失败: ' + error.message, 'error');
    }
  }

  async saveSettings() {
    try {
      const formData = new FormData(document.getElementById('settings-form'));
      const settings = {
        serverURL: formData.get('serverURL'),
        username: formData.get('username'),
        password: formData.get('password')
      };

      // 验证服务器地址格式
      if (settings.serverURL) {
        try {
          new URL(settings.serverURL);
        } catch (error) {
          this.showStatus('服务器地址格式不正确', 'error');
          return;
        }
      }

      await this.storeSettings(settings);

      // 通知background.js配置已更新
      if (chrome && chrome.runtime) {
        chrome.runtime.sendMessage({
          action: 'saveConfig',
          config: settings
        }, (response) => {
          console.log('Background配置更新响应:', response);
        });
      }

      this.showStatus('设置已保存', 'success');

      // 如果提供了用户名和密码，尝试自动登录
      if (settings.username && settings.password) {
        this.autoLogin(settings);
      }

    } catch (error) {
      this.showStatus('保存设置失败: ' + error.message, 'error');
    }
  }

  async saveAdvancedSettings() {
    try {
      const currentSettings = await this.getStoredSettings();
      
      const advancedSettings = {
        ...currentSettings,
        autoDetect: document.getElementById('auto-detect').checked,
        showNotifications: document.getElementById('show-notifications').checked,
        autoSave: document.getElementById('auto-save').checked,
        excludedSites: document.getElementById('excluded-sites').value
      };

      await this.storeSettings(advancedSettings);
      this.showStatus('高级设置已保存', 'success');

    } catch (error) {
      this.showStatus('保存高级设置失败: ' + error.message, 'error');
    }
  }

  async testConnection() {
    const serverURL = document.getElementById('server-url').value;
    
    if (!serverURL) {
      this.showTestResult('请先输入服务器地址', 'error');
      return;
    }

    try {
      new URL(serverURL);
    } catch (error) {
      this.showTestResult('服务器地址格式不正确', 'error');
      return;
    }

    this.showTestResult('正在测试连接...', 'info');

    try {
      const response = await fetch(`${serverURL}/api/v1/health`, {
        method: 'GET',
        timeout: 5000
      });

      if (response.ok) {
        const data = await response.json();
        this.showTestResult(`连接成功！服务器状态: ${data.status}`, 'success');
      } else {
        this.showTestResult(`连接失败，HTTP状态码: ${response.status}`, 'error');
      }
    } catch (error) {
      if (error.name === 'TypeError' && error.message.includes('fetch')) {
        this.showTestResult('连接失败：无法访问服务器，请检查地址和网络连接', 'error');
      } else {
        this.showTestResult('连接失败: ' + error.message, 'error');
      }
    }
  }

  async autoLogin(settings) {
    try {
      const response = await fetch(`${settings.serverURL}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: settings.username,
          password: settings.password
        })
      });

      if (response.ok) {
        const data = await response.json();
        await this.storeSettings({
          ...settings,
          token: data.data.token
        });
        this.showStatus('设置已保存，自动登录成功', 'success');
      } else {
        const error = await response.json();
        this.showStatus('设置已保存，但自动登录失败: ' + error.message, 'error');
      }
    } catch (error) {
      this.showStatus('设置已保存，但自动登录失败: ' + error.message, 'error');
    }
  }

  showStatus(message, type) {
    const statusEl = document.getElementById('status');
    statusEl.textContent = message;
    statusEl.className = `status ${type}`;
    statusEl.style.display = 'block';

    // 3秒后自动隐藏成功消息
    if (type === 'success') {
      setTimeout(() => {
        statusEl.style.display = 'none';
      }, 3000);
    }
  }

  showTestResult(message, type) {
    const resultEl = document.getElementById('test-result');
    resultEl.textContent = message;
    resultEl.className = `test-result ${type}`;
    resultEl.style.display = 'block';

    // 5秒后自动隐藏
    setTimeout(() => {
      resultEl.style.display = 'none';
    }, 5000);
  }

  getStoredSettings() {
    return new Promise((resolve) => {
      chrome.storage.sync.get(null, (result) => {
        resolve(result);
      });
    });
  }

  storeSettings(settings) {
    return new Promise((resolve) => {
      chrome.storage.sync.set(settings, resolve);
    });
  }

  async logout() {
    if (confirm('确定要退出登录吗？这将清除所有登录信息。')) {
      try {
        // 清除token和密码
        const currentSettings = await this.getStoredSettings();
        const newSettings = {
          ...currentSettings,
          token: '',
          password: '' // 可选：是否也清除保存的密码
        };

        await this.storeSettings(newSettings);

        // 通知background.js清除token
        if (chrome && chrome.runtime) {
          chrome.runtime.sendMessage({
            action: 'saveConfig',
            config: newSettings
          }, (response) => {
            console.log('Background登出响应:', response);
          });
        }

        this.showStatus('已退出登录', 'success');

        // 清空密码字段显示
        document.getElementById('password').value = '';

      } catch (error) {
        this.showStatus('退出登录失败: ' + error.message, 'error');
      }
    }
  }
}

// 初始化设置管理器
document.addEventListener('DOMContentLoaded', () => {
  new OptionsManager();
});
