// 认证调试工具 - 在popup的开发者工具Console中执行

console.log('🔐 开始认证调试...');

window.authDebug = {
  
  // 1. 检查当前认证状态
  async checkAuthStatus() {
    console.log('📋 检查认证状态...');
    
    return new Promise((resolve) => {
      chrome.storage.sync.get(['serverURL', 'token', 'username'], (result) => {
        console.log('存储中的认证信息:', {
          serverURL: result.serverURL,
          hasToken: !!result.token,
          tokenLength: result.token?.length,
          tokenPreview: result.token ? result.token.substring(0, 20) + '...' : 'null',
          username: result.username
        });
        resolve(result);
      });
    });
  },
  
  // 2. 测试登录
  async testLogin(username, password) {
    if (!username || !password) {
      console.error('❌ 请提供用户名和密码');
      return;
    }
    
    console.log(`🔐 测试登录: ${username}`);
    
    return new Promise((resolve) => {
      chrome.runtime.sendMessage({
        action: 'login',
        username,
        password
      }, (response) => {
        console.log('登录响应:', response);
        if (response.success) {
          console.log('✅ 登录成功');
          // 立即检查认证状态
          setTimeout(() => this.checkAuthStatus(), 500);
        } else {
          console.error('❌ 登录失败:', response.error);
        }
        resolve(response);
      });
    });
  },
  
  // 3. 测试获取数据
  async testGetData() {
    console.log('📊 测试获取平台注册数据...');
    
    return new Promise((resolve) => {
      chrome.runtime.sendMessage({
        action: 'getRegistrations'
      }, (response) => {
        console.log('获取数据响应:', response);
        if (response.success) {
          console.log('✅ 数据获取成功，项目数量:', response.data?.data?.length || 0);
        } else {
          console.error('❌ 数据获取失败:', response.error);
        }
        resolve(response);
      });
    });
  },
  
  // 4. 手动设置token
  async setToken(token) {
    if (!token) {
      console.error('❌ 请提供token');
      return;
    }
    
    console.log('🔑 手动设置token...');
    
    return new Promise((resolve) => {
      chrome.storage.sync.get(['serverURL', 'username'], (current) => {
        const newConfig = {
          ...current,
          token: token
        };
        
        chrome.storage.sync.set(newConfig, () => {
          console.log('✅ Token已保存到存储');
          
          // 通知background.js
          chrome.runtime.sendMessage({
            action: 'saveConfig',
            config: newConfig
          }, (response) => {
            console.log('Background更新响应:', response);
            resolve(response);
          });
        });
      });
    });
  },
  
  // 5. 清除认证信息
  async clearAuth() {
    console.log('🗑️ 清除认证信息...');
    
    return new Promise((resolve) => {
      chrome.storage.sync.get(['serverURL', 'username'], (current) => {
        const newConfig = {
          ...current,
          token: ''
        };
        
        chrome.storage.sync.set(newConfig, () => {
          console.log('✅ 认证信息已清除');
          
          // 通知background.js
          chrome.runtime.sendMessage({
            action: 'saveConfig',
            config: newConfig
          }, (response) => {
            console.log('Background更新响应:', response);
            resolve(response);
          });
        });
      });
    });
  },
  
  // 6. 完整的认证流程测试
  async fullAuthTest(username, password) {
    console.log('🧪 开始完整认证流程测试...');
    
    try {
      // 1. 检查初始状态
      console.log('1️⃣ 检查初始状态...');
      await this.checkAuthStatus();
      
      // 2. 清除旧的认证信息
      console.log('2️⃣ 清除旧认证信息...');
      await this.clearAuth();
      
      // 3. 执行登录
      console.log('3️⃣ 执行登录...');
      const loginResult = await this.testLogin(username, password);
      
      if (!loginResult.success) {
        console.error('❌ 登录失败，停止测试');
        return { success: false, step: 'login', error: loginResult.error };
      }
      
      // 4. 等待一下让配置生效
      console.log('4️⃣ 等待配置生效...');
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // 5. 检查认证状态
      console.log('5️⃣ 检查登录后状态...');
      await this.checkAuthStatus();
      
      // 6. 测试数据获取
      console.log('6️⃣ 测试数据获取...');
      const dataResult = await this.testGetData();
      
      if (dataResult.success) {
        console.log('✅ 完整认证流程测试成功！');
        return { success: true, data: dataResult.data };
      } else {
        console.error('❌ 数据获取失败');
        return { success: false, step: 'getData', error: dataResult.error };
      }
      
    } catch (error) {
      console.error('❌ 认证流程测试出错:', error);
      return { success: false, step: 'unknown', error: error.message };
    }
  },
  
  // 7. 检查服务器连接
  async testServerConnection() {
    console.log('🌐 测试服务器连接...');
    
    const config = await this.checkAuthStatus();
    
    if (!config.serverURL) {
      console.error('❌ 服务器地址未配置');
      return { success: false, error: '服务器地址未配置' };
    }
    
    try {
      const response = await fetch(`${config.serverURL}/api/v1/health`);
      
      if (response.ok) {
        const data = await response.json();
        console.log('✅ 服务器连接正常:', data);
        return { success: true, data };
      } else {
        console.error('❌ 服务器响应错误:', response.status, response.statusText);
        return { success: false, error: `HTTP ${response.status}` };
      }
    } catch (error) {
      console.error('❌ 服务器连接失败:', error);
      return { success: false, error: error.message };
    }
  }
};

// 自动执行初始检查
console.log('🚀 自动执行认证状态检查...');
authDebug.checkAuthStatus().then(config => {
  if (!config.serverURL) {
    console.log('⚠️ 服务器地址未配置');
    console.log('💡 请先配置服务器地址: authDebug.setServerURL("https://accountback.azhen.de")');
  } else if (!config.token) {
    console.log('⚠️ 未登录状态');
    console.log('💡 执行登录测试: authDebug.testLogin("用户名", "密码")');
  } else {
    console.log('✅ 已有认证信息，测试数据获取...');
    authDebug.testGetData();
  }
});

console.log('🎯 认证调试工具已加载！');
console.log('📖 可用命令:');
console.log('  authDebug.checkAuthStatus() - 检查认证状态');
console.log('  authDebug.testLogin(username, password) - 测试登录');
console.log('  authDebug.testGetData() - 测试获取数据');
console.log('  authDebug.setToken(token) - 手动设置token');
console.log('  authDebug.clearAuth() - 清除认证信息');
console.log('  authDebug.fullAuthTest(username, password) - 完整认证流程测试');
console.log('  authDebug.testServerConnection() - 测试服务器连接');
