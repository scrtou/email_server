// 后台脚本 - 处理API通信和数据存储

class EmailServerAPI {
  constructor() {
    this.baseURL = '';
    this.token = '';
    this.initialized = false;
    this.init();
  }

  async init() {
    try {
      const config = await this.getStoredConfig();
      this.baseURL = config.serverURL || '';
      this.token = config.token || '';
      this.initialized = true;
      console.log('🔧 EmailServerAPI初始化完成:', { baseURL: this.baseURL, hasToken: !!this.token });
    } catch (error) {
      console.error('❌ EmailServerAPI初始化失败:', error);
    }
  }

  async ensureInitialized() {
    if (!this.initialized) {
      console.log('⏳ 等待API初始化...');
      await this.init();
    }
  }

  async getStoredConfig() {
    return new Promise((resolve) => {
      chrome.storage.sync.get(['serverURL', 'token', 'username', 'password'], (result) => {
        console.log('📦 从存储中读取配置:', result);
        // 设置默认服务器地址
        if (!result.serverURL) {
          result.serverURL = 'https://accountback.azhen.de';
        }
        resolve(result);
      });
    });
  }

  async saveConfig(config) {
    console.log('💾 保存配置到存储:', config);
    return new Promise((resolve) => {
      chrome.storage.sync.set(config, () => {
        console.log('✅ 配置保存完成');
        resolve();
      });
    });
  }

  async login(username, password) {
    await this.ensureInitialized();

    console.log('🔐 尝试登录:', { username, baseURL: this.baseURL });

    if (!this.baseURL) {
      console.error('❌ 服务器地址未配置');
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
        console.log('🔐 登录响应数据:', data);

        // 检查token是否存在
        if (data.data && data.data.token) {
          this.token = data.data.token;
          console.log('✅ Token已设置:', this.token.substring(0, 10) + '...');

          // 获取当前配置，只更新token，保留其他配置
          const currentConfig = await this.getStoredConfig();
          await this.saveConfig({ ...currentConfig, token: this.token });
          console.log('💾 Token已保存到存储');

          return { success: true, data };
        } else {
          console.error('❌ 登录响应中没有token:', data);
          return { success: false, error: '登录响应格式错误：缺少token' };
        }
      } else {
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const error = await response.json();
          errorMessage = error.message || error.error || errorMessage;
          console.error('❌ 登录失败:', error);
        } catch (parseError) {
          console.error('❌ 解析登录错误响应失败:', parseError);
        }
        return { success: false, error: errorMessage };
      }
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  async checkPlatformRegistrationConflict(registrationData) {
    await this.ensureInitialized();

    if (!this.baseURL) {
      return { hasConflict: false, error: '请先在设置中配置服务器地址' };
    }

    try {
      // 使用一个新的API端点来只检查冲突，不实际保存
      const response = await fetch(`${this.baseURL}/api/v1/platform-registrations/check-conflict`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.token}`
        },
        body: JSON.stringify(registrationData)
      });

      if (response.ok) {
        const data = await response.json();
        return { hasConflict: false, data };
      } else if (response.status === 409) {
        // 检测到冲突
        const error = await response.json();
        return {
          hasConflict: true,
          conflictData: error.data,
          message: error.message
        };
      } else {
        const error = await response.json();
        return { hasConflict: false, error: error.message };
      }
    } catch (error) {
      return { hasConflict: false, error: error.message };
    }
  }

  async createPlatformRegistration(registrationData) {
    await this.ensureInitialized();

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
      } else if (response.status === 409) {
        // 处理冲突情况
        const error = await response.json();
        return {
          success: false,
          error: error.message,
          conflict: true,
          conflictData: error.data
        };
      } else {
        const error = await response.json();
        return { success: false, error: error.message };
      }
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  async getPlatformRegistrations() {
    await this.ensureInitialized();

    console.log('🔍 获取平台注册信息:', { baseURL: this.baseURL, hasToken: !!this.token, tokenLength: this.token?.length });

    if (!this.baseURL) {
      return { success: false, error: '请先在设置中配置服务器地址' };
    }

    if (!this.token) {
      console.error('❌ Token为空，需要重新登录');
      return { success: false, error: '认证信息已过期，请重新登录' };
    }

    try {
      const headers = {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      };

      console.log('📡 发送请求:', {
        url: `${this.baseURL}/api/v1/platform-registrations?pageSize=0`,
        headers: { ...headers, Authorization: `Bearer ${this.token.substring(0, 10)}...` }
      });

      const response = await fetch(`${this.baseURL}/api/v1/platform-registrations?pageSize=0`, {
        method: 'GET',
        headers
      });

      console.log('📨 响应状态:', response.status, response.statusText);

      if (response.ok) {
        const responseData = await response.json();
        console.log('✅ 获取数据成功:', responseData);
        return { success: true, data: responseData.data };
      } else {
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const error = await response.json();
          errorMessage = error.message || error.error || errorMessage;
          console.error('❌ 服务器错误:', error);
        } catch (parseError) {
          console.error('❌ 解析错误响应失败:', parseError);
        }
        return { success: false, error: errorMessage };
      }
    } catch (error) {
      console.error('❌ 网络请求失败:', error);
      return { success: false, error: `网络错误: ${error.message}` };
    }
  }

  async getPlatformRegistrationById(id) {
    await this.ensureInitialized();

    console.log('🔍 获取平台注册详情:', { id, baseURL: this.baseURL, hasToken: !!this.token });

    if (!this.baseURL) {
      return { success: false, error: '请先在设置中配置服务器地址' };
    }

    if (!this.token) {
      console.error('❌ Token为空，需要重新登录');
      return { success: false, error: '认证信息已过期，请重新登录' };
    }

    try {
      const headers = {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      };

      console.log('📡 发送请求:', {
        url: `${this.baseURL}/api/v1/platform-registrations/${id}`,
        headers: { ...headers, Authorization: `Bearer ${this.token.substring(0, 10)}...` }
      });

      const response = await fetch(`${this.baseURL}/api/v1/platform-registrations/${id}`, {
        method: 'GET',
        headers
      });

      console.log('📨 响应状态:', response.status, response.statusText);

      if (response.ok) {
        const responseData = await response.json();
        console.log('✅ 获取详情成功:', responseData);
        // 提取实际的数据部分
        return { success: true, data: responseData.data };
      } else {
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const error = await response.json();
          errorMessage = error.message || error.error || errorMessage;
          console.error('❌ 服务器错误:', error);
        } catch (parseError) {
          console.error('❌ 解析错误响应失败:', parseError);
        }
        return { success: false, error: errorMessage };
      }
    } catch (error) {
      console.error('❌ 网络请求失败:', error);
      return { success: false, error: `网络错误: ${error.message}` };
    }
  }

  async getPlatformRegistrationPassword(id) {
    await this.ensureInitialized();

    console.log('🔍 获取平台注册密码:', { id, baseURL: this.baseURL, hasToken: !!this.token });

    if (!this.baseURL) {
      return { success: false, error: '请先在设置中配置服务器地址' };
    }

    if (!this.token) {
      console.error('❌ Token为空，需要重新登录');
      return { success: false, error: '认证信息已过期，请重新登录' };
    }

    try {
      const headers = {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      };

      console.log('📡 发送请求:', {
        url: `${this.baseURL}/api/v1/platform-registrations/${id}/password`,
        headers: { ...headers, Authorization: `Bearer ${this.token.substring(0, 10)}...` }
      });

      const response = await fetch(`${this.baseURL}/api/v1/platform-registrations/${id}/password`, {
        method: 'GET',
        headers
      });

      console.log('📨 响应状态:', response.status, response.statusText);

      if (response.ok) {
        const responseData = await response.json();
        console.log('✅ 获取密码成功');
        return { success: true, data: responseData.data };
      } else {
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const error = await response.json();
          errorMessage = error.message || error.error || errorMessage;
          console.error('❌ 服务器错误:', error);
        } catch (parseError) {
          console.error('❌ 解析错误响应失败:', parseError);
        }
        return { success: false, error: errorMessage };
      }
    } catch (error) {
      console.error('❌ 网络请求失败:', error);
      return { success: false, error: `网络错误: ${error.message}` };
    }
  }

  async updatePlatformRegistration(id, data) {
    await this.ensureInitialized();

    console.log('📝 更新平台注册信息:', { id, data, baseURL: this.baseURL, hasToken: !!this.token });

    if (!this.baseURL) {
      return { success: false, error: '请先在设置中配置服务器地址' };
    }

    if (!this.token) {
      console.error('❌ Token为空，需要重新登录');
      return { success: false, error: '认证信息已过期，请重新登录' };
    }

    try {
      const headers = {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      };

      console.log('📡 发送请求:', {
        url: `${this.baseURL}/api/v1/platform-registrations/${id}`,
        method: 'PUT',
        headers: { ...headers, Authorization: `Bearer ${this.token.substring(0, 10)}...` },
        data
      });

      const response = await fetch(`${this.baseURL}/api/v1/platform-registrations/${id}`, {
        method: 'PUT',
        headers,
        body: JSON.stringify(data)
      });

      console.log('📨 响应状态:', response.status, response.statusText);

      if (response.ok) {
        const responseData = await response.json();
        console.log('✅ 更新成功:', responseData);
        return { success: true, data: responseData.data };
      } else {
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const error = await response.json();
          errorMessage = error.message || error.error || errorMessage;
          console.error('❌ 服务器错误:', error);
        } catch (parseError) {
          console.error('❌ 解析错误响应失败:', parseError);
        }
        return { success: false, error: errorMessage };
      }
    } catch (error) {
      console.error('❌ 网络请求失败:', error);
      return { success: false, error: `网络错误: ${error.message}` };
    }
  }

  async deletePlatformRegistration(id) {
    await this.ensureInitialized();

    console.log('🗑️ 删除平台注册信息:', { id, baseURL: this.baseURL, hasToken: !!this.token });

    if (!this.baseURL) {
      return { success: false, error: '请先在设置中配置服务器地址' };
    }

    if (!this.token) {
      console.error('❌ Token为空，需要重新登录');
      return { success: false, error: '认证信息已过期，请重新登录' };
    }

    try {
      const headers = {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      };

      console.log('📡 发送请求:', {
        url: `${this.baseURL}/api/v1/platform-registrations/${id}`,
        method: 'DELETE',
        headers: { ...headers, Authorization: `Bearer ${this.token.substring(0, 10)}...` }
      });

      const response = await fetch(`${this.baseURL}/api/v1/platform-registrations/${id}`, {
        method: 'DELETE',
        headers
      });

      console.log('📨 响应状态:', response.status, response.statusText);

      if (response.ok) {
        const responseData = await response.json();
        console.log('✅ 删除成功:', responseData);
        return { success: true, data: responseData.data };
      } else {
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const error = await response.json();
          errorMessage = error.message || error.error || errorMessage;
          console.error('❌ 服务器错误:', error);
        } catch (parseError) {
          console.error('❌ 解析错误响应失败:', parseError);
        }
        return { success: false, error: errorMessage };
      }
    } catch (error) {
      console.error('❌ 网络请求失败:', error);
      return { success: false, error: `网络错误: ${error.message}` };
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

      case 'checkRegistrationConflict':
        const conflictResult = await api.checkPlatformRegistrationConflict(request.data);
        sendResponse(conflictResult);
        break;

      case 'saveRegistration':
        const saveResult = await api.createPlatformRegistration(request.data);
        sendResponse(saveResult);
        break;

      case 'updateRegistrationPassword':
        // 使用完整的数据更新，如果提供了的话
        const updateData = request.data || { login_password: request.password };
        console.log('🔄 更新密码请求:', {
          id: request.id,
          hasData: !!request.data,
          updateData: { ...updateData, login_password: '***' }
        });
        const updatePasswordResult = await api.updatePlatformRegistration(request.id, updateData);
        sendResponse(updatePasswordResult);
        break;

      case 'getRegistrations':
        const getResult = await api.getPlatformRegistrations();
        sendResponse(getResult);
        break;

      case 'getRegistrationById':
        const getByIdResult = await api.getPlatformRegistrationById(request.id);
        sendResponse(getByIdResult);
        break;

      case 'getRegistrationPassword':
        const getPasswordResult = await api.getPlatformRegistrationPassword(request.id);
        sendResponse(getPasswordResult);
        break;

      case 'getAutoSaveSetting':
        // 获取自动保存设置
        chrome.storage.sync.get(['autoSave'], (result) => {
          sendResponse({ autoSave: result.autoSave || false });
        });
        return true; // 保持消息通道开放

      case 'updateRegistration':
        const updateResult = await api.updatePlatformRegistration(request.id, request.data);
        sendResponse(updateResult);
        break;

      case 'deleteRegistration':
        const deleteResult = await api.deletePlatformRegistration(request.id);
        sendResponse(deleteResult);
        break;

      case 'getConfig':
        const config = await api.getStoredConfig();
        sendResponse(config);
        break;

      case 'saveConfig':
        await api.saveConfig(request.config);
        // 立即更新API实例的配置
        api.baseURL = request.config.serverURL || '';
        if (request.config.token) {
          api.token = request.config.token;
        }
        api.initialized = true; // 标记为已初始化
        console.log('✅ 配置已更新:', { baseURL: api.baseURL, hasToken: !!api.token });
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
