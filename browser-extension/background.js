// 后台脚本 - 处理API通信和数据存储

class EmailServerAPI {
  constructor() {
    this.baseURL = '';
    this.token = '';
    this.init();
  }

  async init() {
    const config = await this.getStoredConfig();
    this.baseURL = config.serverURL || '';
    this.token = config.token || '';
  }

  async getStoredConfig() {
    return new Promise((resolve) => {
      chrome.storage.sync.get(['serverURL', 'token'], (result) => {
        resolve(result);
      });
    });
  }

  async saveConfig(config) {
    return new Promise((resolve) => {
      chrome.storage.sync.set(config, resolve);
    });
  }

  async login(username, password) {
    if (!this.baseURL) {
      return { success: false, error: '请先在设置中配置服务器地址' };
    }

    try {
      const response = await fetch(`${this.baseURL}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password })
      });

      if (response.ok) {
        const data = await response.json();
        this.token = data.data.access_token;
        await this.saveConfig({ token: this.token });
        return { success: true, data };
      } else {
        const error = await response.json();
        return { success: false, error: error.message };
      }
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  async createPlatformRegistration(registrationData) {
    if (!this.baseURL) {
      return { success: false, error: '请先在设置中配置服务器地址' };
    }

    try {
      const response = await fetch(`${this.baseURL}/api/v1/platform-registrations/by-name`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.token}`
        },
        body: JSON.stringify(registrationData)
      });

      if (response.ok) {
        const data = await response.json();
        return { success: true, data };
      } else {
        const error = await response.json();
        return { success: false, error: error.message };
      }
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  async getPlatformRegistrations() {
    if (!this.baseURL) {
      return { success: false, error: '请先在设置中配置服务器地址' };
    }

    try {
      const response = await fetch(`${this.baseURL}/api/v1/platform-registrations`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${this.token}`
        }
      });

      if (response.ok) {
        const data = await response.json();
        return { success: true, data };
      } else {
        const error = await response.json();
        return { success: false, error: error.message };
      }
    } catch (error) {
      return { success: false, error: error.message };
    }
  }
}

const api = new EmailServerAPI();

// 监听来自content script和popup的消息
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  (async () => {
    switch (request.action) {
      case 'login':
        const loginResult = await api.login(request.username, request.password);
        sendResponse(loginResult);
        break;

      case 'saveRegistration':
        const saveResult = await api.createPlatformRegistration(request.data);
        sendResponse(saveResult);
        break;

      case 'getRegistrations':
        const getResult = await api.getPlatformRegistrations();
        sendResponse(getResult);
        break;

      case 'getConfig':
        const config = await api.getStoredConfig();
        sendResponse(config);
        break;

      case 'saveConfig':
        await api.saveConfig(request.config);
        api.baseURL = request.config.serverURL || '';
        sendResponse({ success: true });
        break;

      default:
        sendResponse({ success: false, error: 'Unknown action' });
    }
  })();
  return true; // 保持消息通道开放以支持异步响应
});

// 监听标签页更新，检测登录页面
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (changeInfo.status === 'complete' && tab.url) {
    // 检测是否为登录相关页面
    const loginKeywords = ['login', 'signin', 'register', 'signup', 'auth'];
    const url = tab.url.toLowerCase();
    
    if (loginKeywords.some(keyword => url.includes(keyword))) {
      // 向content script发送消息，开始监听表单
      chrome.tabs.sendMessage(tabId, { action: 'startFormDetection' });
    }
  }
});
