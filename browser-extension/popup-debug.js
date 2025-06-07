// 在popup的开发者工具Console中执行这个脚本
// 复制整个内容并粘贴到Console中

console.log('🔧 开始扩展调试...');

// 调试函数集合
window.debugExtension = {
  
  // 1. 检查当前配置
  async checkConfig() {
    console.log('📋 检查当前配置...');
    
    return new Promise((resolve) => {
      chrome.storage.sync.get(['serverURL', 'token', 'username', 'password'], (result) => {
        console.log('存储中的配置:', result);
        resolve(result);
      });
    });
  },
  
  // 2. 设置服务器地址
  async setServerURL(url = 'https://accountback.azhen.de') {
    console.log(`💾 设置服务器地址: ${url}`);
    
    return new Promise((resolve) => {
      chrome.storage.sync.set({ serverURL: url }, () => {
        console.log('✅ 服务器地址已保存到存储');
        
        // 通知background.js
        chrome.runtime.sendMessage({
          action: 'saveConfig',
          config: { serverURL: url }
        }, (response) => {
          console.log('Background响应:', response);
          resolve(response);
        });
      });
    });
  },
  
  // 3. 测试background通信
  async testBackground() {
    console.log('📨 测试Background通信...');
    
    return new Promise((resolve) => {
      chrome.runtime.sendMessage({ action: 'getConfig' }, (response) => {
        console.log('Background返回的配置:', response);
        resolve(response);
      });
    });
  },
  
  // 4. 测试登录
  async testLogin(username, password) {
    console.log(`🔐 测试登录: ${username}`);
    
    return new Promise((resolve) => {
      chrome.runtime.sendMessage({
        action: 'login',
        username,
        password
      }, (response) => {
        console.log('登录结果:', response);
        resolve(response);
      });
    });
  },
  
  // 5. 强制重新初始化
  async forceReload() {
    console.log('🔄 强制重新初始化...');
    
    return new Promise((resolve) => {
      chrome.runtime.sendMessage({ action: 'forceReload' }, (response) => {
        console.log('重新初始化结果:', response);
        resolve(response);
      });
    });
  },
  
  // 6. 完整诊断
  async fullDiagnosis() {
    console.log('🔍 开始完整诊断...');
    
    try {
      // 检查存储配置
      const storageConfig = await this.checkConfig();
      
      // 检查background配置
      const backgroundConfig = await this.testBackground();
      
      // 比较配置
      console.log('📊 配置对比:');
      console.log('存储配置:', storageConfig);
      console.log('Background配置:', backgroundConfig);
      
      // 检查是否一致
      const isConsistent = storageConfig.serverURL === backgroundConfig.serverURL;
      console.log(`配置一致性: ${isConsistent ? '✅ 一致' : '❌ 不一致'}`);
      
      if (!isConsistent) {
        console.log('🔧 尝试修复配置不一致问题...');
        await this.setServerURL(storageConfig.serverURL || 'https://accountback.azhen.de');
        await this.forceReload();
      }
      
      return {
        storageConfig,
        backgroundConfig,
        isConsistent
      };
      
    } catch (error) {
      console.error('❌ 诊断过程中出错:', error);
      return { error: error.message };
    }
  },
  
  // 7. 快速修复
  async quickFix() {
    console.log('⚡ 执行快速修复...');
    
    try {
      // 1. 设置服务器地址
      await this.setServerURL();
      
      // 2. 等待一下
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // 3. 强制重新加载
      await this.forceReload();
      
      // 4. 验证修复结果
      const result = await this.testBackground();
      
      if (result.serverURL) {
        console.log('✅ 快速修复成功！');
        return { success: true, config: result };
      } else {
        console.log('❌ 快速修复失败');
        return { success: false, error: '配置仍然为空' };
      }
      
    } catch (error) {
      console.error('❌ 快速修复过程中出错:', error);
      return { success: false, error: error.message };
    }
  }
};

// 自动执行诊断
console.log('🚀 自动执行诊断...');
debugExtension.fullDiagnosis().then(result => {
  console.log('📋 诊断完成:', result);
  
  if (!result.isConsistent || !result.backgroundConfig.serverURL) {
    console.log('🔧 检测到配置问题，建议执行快速修复');
    console.log('💡 执行: debugExtension.quickFix()');
  } else {
    console.log('✅ 配置正常，可以尝试登录');
    console.log('💡 执行: debugExtension.testLogin("用户名", "密码")');
  }
});

console.log('🎯 调试工具已加载！');
console.log('📖 可用命令:');
console.log('  debugExtension.checkConfig() - 检查配置');
console.log('  debugExtension.setServerURL() - 设置服务器地址');
console.log('  debugExtension.testBackground() - 测试Background通信');
console.log('  debugExtension.testLogin(username, password) - 测试登录');
console.log('  debugExtension.quickFix() - 快速修复');
console.log('  debugExtension.fullDiagnosis() - 完整诊断');
