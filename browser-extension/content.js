// 内容脚本 - 检测和提取表单信息
console.log('🚀 Email Server扩展内容脚本已加载！版本: 2025-06-08-00:20', window.location.href);

class FormDetector {
  constructor() {
    console.log('🔧 FormDetector构造函数被调用');
    this.isDetecting = false;
    this.detectedForms = new Set();
    this.init();
  }

  init() {
    this.startFormDetection();
    this.listenForMessages();
  }

  listenForMessages() {
    chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
      if (request.action === 'startFormDetection') {
        this.startFormDetection();
        sendResponse({ success: true });
      }
    });
  }

  startFormDetection() {
    if (this.isDetecting) return;
    this.isDetecting = true;

    // 检测现有表单
    this.detectExistingForms();

    // 监听新表单的出现
    this.observeFormChanges();

    // 表单提交监听在 attachFormListener 方法中处理
  }

  detectExistingForms() {
    const forms = document.querySelectorAll('form');
    forms.forEach(form => this.analyzeForm(form));
  }

  observeFormChanges() {
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        mutation.addedNodes.forEach((node) => {
          if (node.nodeType === Node.ELEMENT_NODE) {
            if (node.tagName === 'FORM') {
              this.analyzeForm(node);
            } else {
              const forms = node.querySelectorAll && node.querySelectorAll('form');
              if (forms) {
                forms.forEach(form => this.analyzeForm(form));
              }
            }
          }
        });
      });
    });

    observer.observe(document.body, {
      childList: true,
      subtree: true
    });
  }

  analyzeForm(form) {
    if (this.detectedForms.has(form)) return;

    const formData = this.extractFormData(form);
    if (formData.isLoginForm || formData.isRegisterForm) {
      this.detectedForms.add(form);
      this.attachFormListener(form, formData);
    }
  }

  extractFormData(form) {
    const inputs = form.querySelectorAll('input');
    const formData = {
      isLoginForm: false,
      isRegisterForm: false,
      emailField: null,
      usernameField: null,
      passwordField: null,
      confirmPasswordField: null,
      fields: []
    };

    inputs.forEach(input => {
      const type = input.type.toLowerCase();
      const name = input.name.toLowerCase();
      const id = input.id.toLowerCase();
      const placeholder = (input.placeholder || '').toLowerCase();

      const fieldInfo = {
        element: input,
        type: type,
        name: name,
        id: id,
        placeholder: placeholder
      };

      formData.fields.push(fieldInfo);

      // 检测邮箱字段
      if (type === 'email' || 
          name.includes('email') || 
          id.includes('email') || 
          placeholder.includes('email') ||
          placeholder.includes('邮箱')) {
        formData.emailField = input;
      }

      // 检测用户名字段
      if (name.includes('username') || 
          name.includes('user') ||
          id.includes('username') || 
          id.includes('user') ||
          placeholder.includes('username') ||
          placeholder.includes('用户名')) {
        formData.usernameField = input;
      }

      // 检测密码字段
      if (type === 'password') {
        if (!formData.passwordField) {
          formData.passwordField = input;
        } else if (name.includes('confirm') || 
                   id.includes('confirm') ||
                   placeholder.includes('confirm') ||
                   placeholder.includes('确认')) {
          formData.confirmPasswordField = input;
        }
      }
    });

    // 判断表单类型
    const formText = form.textContent.toLowerCase();
    const hasLogin = formText.includes('login') || formText.includes('登录') || formText.includes('sign in');
    const hasRegister = formText.includes('register') || formText.includes('注册') || formText.includes('sign up');

    if (formData.confirmPasswordField || hasRegister) {
      formData.isRegisterForm = true;
    } else if (formData.passwordField && (hasLogin || formData.usernameField || formData.emailField)) {
      formData.isLoginForm = true;
    }

    return formData;
  }

  attachFormListener(form, formData) {
    form.addEventListener('submit', (event) => {
      this.handleFormSubmission(event, formData);
    });
  }

  handleFormSubmission(event, formData) {
    console.log('🎯 表单提交被检测到！新版本代码正在运行');
    const extractedData = this.extractSubmissionData(formData);

    console.log('📋 提取的数据:', extractedData);

    if (extractedData.email_address || extractedData.login_username) {
      console.log('✅ 检测到有效数据，检查自动保存设置...');

      // 检查自动保存设置
      chrome.runtime.sendMessage({
        action: 'getAutoSaveSetting'
      }, (response) => {
        console.log('⚙️ 自动保存设置响应:', response);

        if (chrome.runtime.lastError) {
          console.error('❌ 获取设置时出错:', chrome.runtime.lastError);
          console.log('💬 出错时默认显示确认提示');
          this.showSavePrompt(extractedData);
          return;
        }

        if (response && response.autoSave) {
          console.log('🚀 自动保存已启用，直接保存');
          this.autoSaveToServer(extractedData);
        } else {
          console.log('💬 自动保存未启用，先检查是否需要提示');
          // 自动保存未启用，先检查是否真的需要保存（智能检测）
          this.checkIfNeedToPromptManual(extractedData);
        }
      });
    } else {
      console.log('❌ 未检测到有效的邮箱或用户名数据');
    }
  }

  extractSubmissionData(formData) {
    const data = {
      platform_name: this.getPlatformName(),
      email_address: '',
      login_username: '',
      login_password: '',
      notes: `自动检测于 ${new Date().toLocaleString()}`
    };

    if (formData.emailField && formData.emailField.value) {
      data.email_address = formData.emailField.value;
    }

    if (formData.usernameField && formData.usernameField.value) {
      data.login_username = formData.usernameField.value;
    }

    if (formData.passwordField && formData.passwordField.value) {
      data.login_password = formData.passwordField.value;
    }

    return data;
  }

  getPlatformName() {
    const hostname = window.location.hostname;
    // 移除www前缀和常见的子域名
    return hostname.replace(/^(www\.|m\.|mobile\.)/, '');
  }

  showSavePrompt(data) {
    // 创建保存提示框
    const promptDiv = document.createElement('div');
    promptDiv.id = 'email-server-save-prompt';
    promptDiv.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      background: #fff;
      border: 2px solid #007cba;
      border-radius: 8px;
      padding: 15px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.15);
      z-index: 10000;
      font-family: Arial, sans-serif;
      font-size: 14px;
      max-width: 300px;
    `;

    promptDiv.innerHTML = `
      <div style="margin-bottom: 10px; font-weight: bold; color: #007cba;">
        检测到账号信息
      </div>
      <div style="margin-bottom: 10px; font-size: 12px; color: #666;">
        平台: ${data.platform_name}<br>
        ${data.email_address ? `邮箱: ${data.email_address}<br>` : ''}
        ${data.login_username ? `用户名: ${data.login_username}<br>` : ''}
      </div>
      <div style="display: flex; gap: 10px;">
        <button id="save-to-server" style="flex: 1; padding: 8px; background: #007cba; color: white; border: none; border-radius: 4px; cursor: pointer;">
          保存到服务器
        </button>
        <button id="dismiss-prompt" style="flex: 1; padding: 8px; background: #ccc; color: #333; border: none; border-radius: 4px; cursor: pointer;">
          忽略
        </button>
      </div>
    `;

    document.body.appendChild(promptDiv);

    // 绑定按钮事件
    document.getElementById('save-to-server').addEventListener('click', () => {
      this.saveToServer(data);
      promptDiv.remove();
    });

    document.getElementById('dismiss-prompt').addEventListener('click', () => {
      promptDiv.remove();
    });
  }



  // 检查是否需要提示用户（手动模式）- 简化逻辑，直接显示保存提示
  checkIfNeedToPromptManual(data) {
    console.log('🔍 手动模式：检查是否需要提示用户保存');

    // 手动模式下的简化逻辑：
    // 1. 对于新账号：直接显示保存提示让用户选择
    // 2. 对于已存在账号：检查密码是否有变化，有变化才提示

    // 先检查是否存在相同的平台注册信息
    chrome.runtime.sendMessage({
      action: 'getRegistrations'
    }, (response) => {
      console.log('📡 手动模式：获取注册列表响应:', {
        success: response?.success,
        hasData: !!(response?.data),
        dataLength: response?.data?.length,
        error: response?.error
      });

      if (response && response.success && response.data && Array.isArray(response.data)) {
        // 查找是否存在相同平台的注册信息
        const existingRegistration = response.data.find(reg =>
          reg.platform_name === data.platform_name &&
          (reg.email_address === data.email_address || reg.login_username === data.login_username)
        );

        console.log('🔍 手动模式：查找结果:', {
          totalRegistrations: response.data.length,
          searchPlatform: data.platform_name,
          searchEmail: data.email_address,
          searchUsername: data.login_username,
          foundExisting: !!existingRegistration,
          existingId: existingRegistration?.id
        });

        if (existingRegistration) {
          console.log('⚠️ 手动模式：找到已存在的注册信息，检查密码是否有变化');
          // 模拟冲突数据结构
          const conflictData = {
            existing_id: existingRegistration.id
          };
          this.checkPasswordChangeAndPromptForManual(data, conflictData);
        } else {
          console.log('💬 手动模式：新账号，显示保存提示');
          this.showSavePrompt(data);
        }
      } else {
        // 无法获取注册列表，为安全起见显示保存提示
        console.log('❌ 手动模式：无法获取注册列表，显示保存提示。错误:', response?.error);
        this.showSavePrompt(data);
      }
    });
  }

  // 检查密码是否有变化，决定是否提示用户（手动模式）
  checkPasswordChangeAndPromptForManual(newData, conflictData) {
    console.log('🔍 手动模式：检查密码变化:', {
      existing_id: conflictData.existing_id,
      newPassword: newData.login_password ? '***' : 'empty'
    });

    // 获取现有注册信息的密码进行比较
    chrome.runtime.sendMessage({
      action: 'getRegistrationPassword',
      id: conflictData.existing_id
    }, (response) => {
      console.log('📡 手动模式：获取密码响应:', {
        success: response.success,
        hasPassword: !!(response.data && response.data.password),
        error: response.error
      });

      if (response.success) {
        const existingPassword = response.data ? response.data.password : '';
        const newPassword = newData.login_password;

        console.log('🔐 手动模式：密码比较:', {
          existingPasswordLength: existingPassword ? existingPassword.length : 0,
          newPasswordLength: newPassword ? newPassword.length : 0,
          passwordsMatch: existingPassword === newPassword
        });

        // 比较密码是否有变化
        const hasExistingPassword = existingPassword && existingPassword.trim() !== '';
        const hasNewPassword = newPassword && newPassword.trim() !== '';

        if (hasNewPassword && (!hasExistingPassword || existingPassword !== newPassword)) {
          // 密码有变化或首次设置密码，显示更新密码确认框
          console.log('⚠️ 手动模式：密码有变化，显示更新密码确认框');
          this.showUpdateConfirmation(newData, conflictData);
        } else {
          // 密码没有变化，静默处理，不打扰用户
          console.log('✅ 手动模式：密码未变化，不显示提示');
        }
      } else {
        // 无法获取现有密码，为安全起见，显示更新密码确认框
        console.log('❌ 手动模式：无法获取现有密码，显示更新密码确认框');
        this.showUpdateConfirmation(newData, conflictData);
      }
    });
  }

  // 自动保存方法 - 直接尝试保存，智能处理冲突
  autoSaveToServer(data) {
    console.log('🚀 开始自动保存:', {
      platform: data.platform_name,
      email: data.email_address,
      username: data.login_username,
      hasPassword: !!data.login_password
    });

    chrome.runtime.sendMessage({
      action: 'saveRegistration',
      data: data
    }, (response) => {
      console.log('📡 自动保存响应:', {
        success: response.success,
        conflict: response.conflict,
        error: response.error,
        conflictData: response.conflictData
      });

      if (response.success) {
        console.log('✅ 自动保存成功');
        this.showNotification('账号信息已自动保存', 'success');
      } else if (response.conflict && response.conflictData) {
        console.log('⚠️ 检测到冲突，开始检查密码变化');
        // 检查密码是否有变化，只有变化时才提示更新
        this.checkPasswordChangeAndPrompt(data, response.conflictData);
      } else {
        // 其他错误不显示通知，避免打扰用户
        console.log('❌ 自动保存失败:', response.error);
      }
    });
  }

  // 手动保存方法 - 用户主动选择保存，强制保存或更新
  saveToServer(data) {
    console.log('💾 用户主动选择保存到服务器');

    chrome.runtime.sendMessage({
      action: 'saveRegistration',
      data: data
    }, (response) => {
      if (response.success) {
        console.log('✅ 手动保存成功');
        this.showNotification('账号信息已保存到服务器', 'success');
      } else if (response.conflict && response.conflictData) {
        // 用户主动选择保存时，如果有冲突，直接更新密码，不再询问
        console.log('⚠️ 检测到冲突，用户主动保存，直接更新密码');
        this.updatePassword(response.conflictData.existing_id, data.login_password, data);
      } else {
        console.log('❌ 手动保存失败:', response.error);
        this.showNotification('保存失败: ' + response.error, 'error');
      }
    });
  }

  showUpdateConfirmation(data, conflictData) {
    // 创建更新确认对话框，与其他弹框保持一致的样式
    const confirmDiv = document.createElement('div');
    confirmDiv.id = 'email-server-update-confirm';
    confirmDiv.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      background: #fff;
      border: 2px solid #ffc107;
      border-radius: 8px;
      padding: 15px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.15);
      z-index: 10002;
      font-family: Arial, sans-serif;
      font-size: 14px;
      max-width: 300px;
    `;

    confirmDiv.innerHTML = `
      <div style="margin-bottom: 10px; font-weight: bold; color: #ffc107;">
        ⚠️ 检测到密码变化
      </div>
      <div style="margin-bottom: 10px; font-size: 12px; color: #666;">
        平台: ${data.platform_name}<br>
        ${data.email_address ? `邮箱: ${data.email_address}<br>` : ''}
        ${data.login_username ? `用户名: ${data.login_username}<br>` : ''}
        是否要更新密码？
      </div>
      <div style="display: flex; gap: 10px;">
        <button id="update-password-btn" style="flex: 1; padding: 8px; background: #ffc107; color: #333; border: none; border-radius: 4px; cursor: pointer; font-weight: bold;">
          更新密码
        </button>
        <button id="cancel-update-btn" style="flex: 1; padding: 8px; background: #ccc; color: #333; border: none; border-radius: 4px; cursor: pointer;">
          忽略
        </button>
      </div>
    `;

    document.body.appendChild(confirmDiv);

    // 绑定按钮事件
    document.getElementById('update-password-btn').addEventListener('click', () => {
      this.updatePassword(conflictData.existing_id, data.login_password, data);
      confirmDiv.remove();
    });

    document.getElementById('cancel-update-btn').addEventListener('click', () => {
      confirmDiv.remove();
    });
  }



  // 检查密码是否有变化，决定是否提示用户（自动模式）
  checkPasswordChangeAndPrompt(newData, conflictData) {
    console.log('🔍 开始检查密码变化:', {
      existing_id: conflictData.existing_id,
      newPassword: newData.login_password ? '***' : 'empty'
    });

    // 获取现有注册信息的密码进行比较
    chrome.runtime.sendMessage({
      action: 'getRegistrationPassword',
      id: conflictData.existing_id
    }, (response) => {
      console.log('📡 获取密码响应:', {
        success: response.success,
        hasPassword: !!(response.data && response.data.password),
        error: response.error,
        responseData: response.data
      });

      if (response.success) {
        const existingPassword = response.data ? response.data.password : '';
        const newPassword = newData.login_password;

        console.log('🔐 密码比较:', {
          existingPasswordLength: existingPassword ? existingPassword.length : 0,
          newPasswordLength: newPassword ? newPassword.length : 0,
          passwordsMatch: existingPassword === newPassword,
          responseData: response.data
        });

        // 比较密码是否有变化
        // 如果数据库中没有密码（空字符串或null），且新密码存在，认为是首次设置密码
        // 如果数据库中有密码，且新密码与现有密码不同，认为是密码变化
        const hasExistingPassword = existingPassword && existingPassword.trim() !== '';
        const hasNewPassword = newPassword && newPassword.trim() !== '';

        if (hasNewPassword && (!hasExistingPassword || existingPassword !== newPassword)) {
          // 密码有变化或首次设置密码，提示用户是否更新
          console.log('⚠️ 密码有变化或首次设置，显示更新提示');
          this.showUpdateConfirmation(newData, conflictData);
        } else {
          // 密码没有变化，不提示用户，静默处理
          console.log('✅ 密码未变化，跳过更新提示');
        }
      } else {
        // 无法获取现有密码，为安全起见，还是提示用户
        console.log('❌ 无法获取现有密码，显示更新提示。错误:', response.error);
        this.showUpdateConfirmation(newData, conflictData);
      }
    });
  }

  updatePassword(registrationId, newPassword, originalData = null) {
    console.log('🔄 开始更新密码:', {
      registrationId,
      hasNewPassword: !!newPassword,
      hasOriginalData: !!originalData
    });

    // 如果有原始数据，传递完整的更新信息
    const updateData = originalData ? {
      email_address: originalData.email_address,
      login_username: originalData.login_username,
      login_password: newPassword,
      notes: originalData.notes,
      phone_number: originalData.phone_number
    } : {
      login_password: newPassword
    };

    chrome.runtime.sendMessage({
      action: 'updateRegistrationPassword',
      id: registrationId,
      password: newPassword,
      data: updateData
    }, (response) => {
      console.log('📡 密码更新响应:', response);
      if (response.success) {
        this.showNotification('密码已成功更新', 'success');
      } else {
        this.showNotification('密码更新失败: ' + response.error, 'error');
      }
    });
  }

  showNotification(message, type) {
    const notification = document.createElement('div');
    notification.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      padding: 12px 20px;
      border-radius: 6px;
      color: white;
      font-family: Arial, sans-serif;
      font-size: 14px;
      z-index: 10001;
      ${type === 'success' ? 'background: #28a745;' : 'background: #dc3545;'}
    `;
    notification.textContent = message;
    document.body.appendChild(notification);

    setTimeout(() => {
      if (notification.parentNode) {
        notification.remove();
      }
    }, 3000);
  }
}

// 初始化表单检测器
console.log('🎯 开始初始化FormDetector...');
const formDetector = new FormDetector();
console.log('✅ FormDetector初始化完成:', formDetector);
