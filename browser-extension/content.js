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
    console.log('🚀 FormDetector 初始化');
    this.checkExtensionStatus();
    this.startFormDetection();
    this.listenForMessages();
  }

  // 检查扩展状态
  checkExtensionStatus() {
    if (!this.isExtensionContextValid()) {
      console.warn('⚠️ 扩展上下文无效，某些功能可能不可用');
      // 延迟显示通知，避免在页面加载时立即显示
      setTimeout(() => {
        this.showNotification('扩展需要重新加载，请刷新页面', 'error');
      }, 2000);
    } else {
      console.log('✅ 扩展上下文有效');
    }
  }

  listenForMessages() {
    chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
      if (request.action === 'startFormDetection') {
        this.startFormDetection();
        sendResponse({ success: true });
      } else if (request.action === 'triggerAutoFill') {
        // 手动触发自动填充（可以通过popup或快捷键触发）
        this.triggerManualAutoFill();
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

      // 如果是登录表单，为账号输入框添加聚焦监听器
      if (formData.isLoginForm) {
        this.attachAutoFillListeners(form, formData);
      }
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

  // 为账号输入框添加聚焦监听器
  attachAutoFillListeners(form, formData) {
    console.log('🎯 为登录表单添加自动填充监听器');

    // 为邮箱字段添加聚焦监听
    if (formData.emailField) {
      this.addFocusListener(formData.emailField, form, formData, 'email');
    }

    // 为用户名字段添加聚焦监听
    if (formData.usernameField) {
      this.addFocusListener(formData.usernameField, form, formData, 'username');
    }
  }

  // 添加聚焦监听器
  addFocusListener(inputField, form, formData, fieldType) {
    console.log(`🔍 为${fieldType}字段添加聚焦监听器:`, inputField);

    // 防止重复添加监听器
    if (inputField.hasAttribute('data-autofill-listener')) {
      return;
    }
    inputField.setAttribute('data-autofill-listener', 'true');

    // 聚焦事件监听器
    const focusHandler = () => {
      console.log(`👆 用户聚焦到${fieldType}字段，检查自动填充`);
      this.checkAutoFillOnFocus(form, formData, inputField);
    };

    // 点击事件监听器（有些情况下focus事件可能不触发）
    const clickHandler = () => {
      console.log(`🖱️ 用户点击${fieldType}字段，检查自动填充`);
      // 延迟一点执行，确保焦点已经设置
      setTimeout(() => {
        this.checkAutoFillOnFocus(form, formData, inputField);
      }, 50);
    };

    inputField.addEventListener('focus', focusHandler);
    inputField.addEventListener('click', clickHandler);

    // 存储事件处理器引用，以便后续清理
    inputField._autoFillHandlers = {
      focus: focusHandler,
      click: clickHandler
    };
  }

  // 当用户聚焦到输入框时检查自动填充
  checkAutoFillOnFocus(form, formData, targetField) {
    console.log('🔍 用户聚焦输入框，检查自动填充:', { domain: this.getPlatformName() });

    // 检查是否已经有内容（避免覆盖用户已输入的内容）
    if (targetField.value && targetField.value.trim() !== '') {
      console.log('📝 输入框已有内容，跳过自动填充');
      return;
    }

    // 获取当前域名匹配的注册信息
    this.safeSendMessage({
      action: 'getRegistrationsByDomain',
      domain: this.getPlatformName()
    }, (response) => {
      if (!response) {
        console.log('❌ 无法获取注册信息，可能是扩展上下文失效');
        return;
      }

      console.log('📡 获取域名匹配注册信息响应:', response);

      if (response && response.success && response.data && response.data.length > 0) {
        console.log('✅ 找到匹配的注册信息，数量:', response.data.length);

        if (response.data.length === 1) {
          // 只有一个匹配的账号，直接填充
          console.log('🚀 单个账号，直接自动填充');
          this.performAutoFill(form, formData, response.data[0]);
        } else {
          // 多个匹配的账号，显示选择界面
          console.log('📋 多个账号，显示选择器');
          this.showAccountSelector(form, formData, response.data, targetField);
        }
      } else {
        console.log('ℹ️ 未找到匹配的注册信息');
      }
    });
  }

  // 检查扩展上下文是否有效
  isExtensionContextValid() {
    try {
      // 尝试访问chrome.runtime，如果失败说明上下文无效
      return !!(chrome && chrome.runtime && chrome.runtime.id);
    } catch (error) {
      console.error('❌ 扩展上下文检查失败:', error);
      return false;
    }
  }

  // 安全的消息发送方法
  safeSendMessage(message, callback) {
    if (!this.isExtensionContextValid()) {
      console.log('❌ 扩展上下文无效，无法发送消息:', message.action);
      if (callback) callback(null);
      return false;
    }

    try {
      chrome.runtime.sendMessage(message, (response) => {
        // 检查是否有运行时错误
        if (chrome.runtime.lastError) {
          console.error('❌ Chrome运行时错误:', chrome.runtime.lastError.message);
          if (callback) callback(null);
          return;
        }

        if (callback) callback(response);
      });
      return true;
    } catch (error) {
      console.error('❌ 发送消息时出错:', error);
      if (callback) callback(null);
      return false;
    }
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

    // 添加详细的字段识别日志
    console.log('🔍 表单字段识别详情:', {
      emailField: formData.emailField ? {
        tagName: formData.emailField.tagName,
        type: formData.emailField.type,
        name: formData.emailField.name,
        id: formData.emailField.id,
        value: formData.emailField.value,
        placeholder: formData.emailField.placeholder
      } : null,
      passwordField: formData.passwordField ? {
        tagName: formData.passwordField.tagName,
        type: formData.passwordField.type,
        name: formData.passwordField.name,
        id: formData.passwordField.id,
        value: formData.passwordField.value ? '***' : 'empty',
        valueLength: formData.passwordField.value ? formData.passwordField.value.length : 0,
        placeholder: formData.passwordField.placeholder
      } : null,
      usernameField: formData.usernameField ? {
        tagName: formData.usernameField.tagName,
        type: formData.usernameField.type,
        name: formData.usernameField.name,
        id: formData.usernameField.id,
        value: formData.usernameField.value,
        placeholder: formData.usernameField.placeholder
      } : null
    });

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
        // 查找是否存在完全匹配的注册信息
        // 必须平台名称相同，且邮箱或用户名完全匹配（不能为空）
        const existingRegistration = response.data.find(reg => {
          const platformMatch = reg.platform_name === data.platform_name;

          // 邮箱匹配：两者都有邮箱且相同
          const emailMatch = data.email_address && reg.email_address &&
                            reg.email_address === data.email_address;

          // 用户名匹配：两者都有用户名且相同
          const usernameMatch = data.login_username && reg.login_username &&
                               reg.login_username === data.login_username;

          return platformMatch && (emailMatch || usernameMatch);
        });

        console.log('🔍 手动模式：查找结果:', {
          totalRegistrations: response.data.length,
          searchPlatform: data.platform_name,
          searchEmail: data.email_address,
          searchUsername: data.login_username,
          foundExisting: !!existingRegistration,
          existingId: existingRegistration?.id,
          existingEmail: existingRegistration?.email_address,
          existingUsername: existingRegistration?.login_username,
          matchDetails: existingRegistration ? {
            platformMatch: existingRegistration.platform_name === data.platform_name,
            emailMatch: data.email_address && existingRegistration.email_address &&
                       existingRegistration.email_address === data.email_address,
            usernameMatch: data.login_username && existingRegistration.login_username &&
                          existingRegistration.login_username === data.login_username
          } : null
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

        console.log('🔐 手动模式：密码比较详情:', {
          existingPassword: existingPassword ? `${existingPassword.substring(0, 5)}...${existingPassword.slice(-3)}` : 'empty',
          newPassword: newPassword ? `${newPassword.substring(0, 5)}...${newPassword.slice(-3)}` : 'empty',
          existingPasswordLength: existingPassword ? existingPassword.length : 0,
          newPasswordLength: newPassword ? newPassword.length : 0,
          passwordsMatch: existingPassword === newPassword,
          exactMatch: existingPassword === newPassword,
          // 添加更详细的比较信息
          existingPasswordFull: existingPassword, // 临时显示完整密码用于调试
          newPasswordFull: newPassword // 临时显示完整密码用于调试
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

        console.log('🔐 自动模式：密码比较详情:', {
          existingPassword: existingPassword ? `${existingPassword.substring(0, 5)}...${existingPassword.slice(-3)}` : 'empty',
          newPassword: newPassword ? `${newPassword.substring(0, 5)}...${newPassword.slice(-3)}` : 'empty',
          existingPasswordLength: existingPassword ? existingPassword.length : 0,
          newPasswordLength: newPassword ? newPassword.length : 0,
          passwordsMatch: existingPassword === newPassword,
          exactMatch: existingPassword === newPassword,
          // 添加更详细的比较信息
          existingPasswordFull: existingPassword, // 临时显示完整密码用于调试
          newPasswordFull: newPassword, // 临时显示完整密码用于调试
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

  // 检查是否需要自动填充（保留用于手动触发）
  checkAutoFill(form, formData) {
    console.log('🔍 手动检查自动填充:', { domain: this.getPlatformName() });

    // 找到目标输入框
    const targetField = formData.emailField || formData.usernameField;

    // 获取当前域名匹配的注册信息
    chrome.runtime.sendMessage({
      action: 'getRegistrationsByDomain',
      domain: this.getPlatformName()
    }, (response) => {
      console.log('📡 获取域名匹配注册信息响应:', response);

      if (response && response.success && response.data && response.data.length > 0) {
        console.log('✅ 找到匹配的注册信息，数量:', response.data.length);

        if (response.data.length === 1) {
          // 只有一个匹配的账号，直接填充
          this.performAutoFill(form, formData, response.data[0]);
        } else {
          // 多个匹配的账号，显示选择界面
          this.showAccountSelector(form, formData, response.data, targetField);
        }
      } else {
        console.log('ℹ️ 未找到匹配的注册信息');
      }
    });
  }

  // 手动触发自动填充（通过popup或快捷键）
  triggerManualAutoFill() {
    console.log('🔧 手动触发自动填充');

    // 查找当前页面的登录表单
    const forms = document.querySelectorAll('form');
    let loginForm = null;
    let loginFormData = null;

    for (const form of forms) {
      const formData = this.extractFormData(form);
      if (formData.isLoginForm) {
        loginForm = form;
        loginFormData = formData;
        break;
      }
    }

    if (loginForm && loginFormData) {
      console.log('✅ 找到登录表单，开始自动填充');
      this.checkAutoFill(loginForm, loginFormData);
    } else {
      console.log('❌ 未找到登录表单');
      this.showNotification('未找到登录表单', 'error');
    }
  }

  // 关闭已存在的选择器
  closeExistingSelector() {
    const existingDropdown = document.getElementById('email-server-account-dropdown');
    if (existingDropdown) {
      existingDropdown.remove();
      console.log('🗑️ 关闭已存在的下拉选择器');
    }

    const existingModal = document.getElementById('email-server-account-modal');
    if (existingModal) {
      existingModal.remove();
      console.log('🗑️ 关闭已存在的模态选择器');
    }
  }

  // 执行自动填充
  performAutoFill(form, formData, accountData) {
    console.log('🚀 执行自动填充:', {
      platform: accountData.platform_name,
      email: accountData.email_address,
      username: accountData.login_username
    });

    // 填充邮箱字段
    if (formData.emailField && accountData.email_address) {
      formData.emailField.value = accountData.email_address;
      formData.emailField.dispatchEvent(new Event('input', { bubbles: true }));
      formData.emailField.dispatchEvent(new Event('change', { bubbles: true }));
    }

    // 填充用户名字段
    if (formData.usernameField && accountData.login_username) {
      formData.usernameField.value = accountData.login_username;
      formData.usernameField.dispatchEvent(new Event('input', { bubbles: true }));
      formData.usernameField.dispatchEvent(new Event('change', { bubbles: true }));
    }

    // 获取并填充密码
    if (formData.passwordField && accountData.id) {
      this.safeSendMessage({
        action: 'getRegistrationPassword',
        id: accountData.id
      }, (passwordResponse) => {
        if (!passwordResponse) {
          console.log('❌ 无法获取密码，可能是扩展上下文失效');
          return;
        }

        if (passwordResponse && passwordResponse.success && passwordResponse.data && passwordResponse.data.password) {
          formData.passwordField.value = passwordResponse.data.password;
          formData.passwordField.dispatchEvent(new Event('input', { bubbles: true }));
          formData.passwordField.dispatchEvent(new Event('change', { bubbles: true }));

          console.log('✅ 自动填充完成');
          this.showNotification('已自动填充登录信息', 'success');
        }
      });
    }
  }

  // 显示账号选择器（下拉式选择）
  showAccountSelector(form, formData, accounts, targetField = null) {
    console.log('📋 显示账号选择器，账号数量:', accounts.length);

    // 如果没有传入targetField，则自动查找
    if (!targetField) {
      targetField = formData.emailField || formData.usernameField;
    }

    if (!targetField) {
      console.log('❌ 未找到目标输入框，使用模态对话框');
      this.showModalAccountSelector(form, formData, accounts);
      return;
    }

    // 关闭任何已存在的选择器
    this.closeExistingSelector();

    // 获取输入框的位置信息
    const rect = targetField.getBoundingClientRect();
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
    const scrollLeft = window.pageXOffset || document.documentElement.scrollLeft;

    // 创建下拉选择器
    const selectorDiv = document.createElement('div');
    selectorDiv.id = 'email-server-account-dropdown';
    selectorDiv.style.cssText = `
      position: absolute;
      top: ${rect.bottom + scrollTop + 2}px;
      left: ${rect.left + scrollLeft}px;
      width: ${Math.max(rect.width, 300)}px;
      background: #fff;
      border: 1px solid #ddd;
      border-radius: 8px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.15);
      z-index: 10003;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      font-size: 14px;
      max-height: 300px;
      overflow-y: auto;
    `;

    // 创建账号选项
    let accountsHtml = accounts.map((account, index) => {
      const displayName = account.email_address || account.login_username || '未知账号';
      const platformIcon = this.getPlatformIcon(account.platform_name);

      return `
        <div class="account-option" data-index="${index}" style="
          display: flex;
          align-items: center;
          padding: 12px 16px;
          border-bottom: 1px solid #f0f0f0;
          cursor: pointer;
          transition: background-color 0.2s;
        " onmouseover="this.style.backgroundColor='#f8f9fa'" onmouseout="this.style.backgroundColor='white'">
          <div style="
            width: 32px;
            height: 32px;
            border-radius: 6px;
            background: #4285f4;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-right: 12px;
            font-size: 16px;
            color: white;
            font-weight: bold;
          ">
            ${platformIcon}
          </div>
          <div style="flex: 1;">
            <div style="font-weight: 500; color: #202124; margin-bottom: 2px;">
              ${account.platform_name}
            </div>
            <div style="font-size: 13px; color: #5f6368;">
              ${displayName}
            </div>
          </div>
          <div style="
            width: 20px;
            height: 20px;
            border-radius: 4px;
            border: 2px solid #4285f4;
            display: flex;
            align-items: center;
            justify-content: center;
          ">
            <div style="
              width: 8px;
              height: 8px;
              background: #4285f4;
              border-radius: 2px;
            "></div>
          </div>
        </div>
      `;
    }).join('');

    // 添加"新增登录"选项
    accountsHtml += `
      <div class="add-new-option" style="
        display: flex;
        align-items: center;
        padding: 12px 16px;
        cursor: pointer;
        transition: background-color 0.2s;
        border-top: 1px solid #e8eaed;
      " onmouseover="this.style.backgroundColor='#f8f9fa'" onmouseout="this.style.backgroundColor='white'">
        <div style="
          width: 32px;
          height: 32px;
          border-radius: 6px;
          border: 2px dashed #4285f4;
          display: flex;
          align-items: center;
          justify-content: center;
          margin-right: 12px;
          font-size: 18px;
          color: #4285f4;
        ">
          +
        </div>
        <div style="flex: 1;">
          <div style="font-weight: 500; color: #4285f4;">
            新增登录
          </div>
        </div>
      </div>
    `;

    selectorDiv.innerHTML = accountsHtml;

    // 添加到页面
    document.body.appendChild(selectorDiv);

    // 绑定事件
    selectorDiv.querySelectorAll('.account-option').forEach((option, index) => {
      option.addEventListener('click', () => {
        this.performAutoFill(form, formData, accounts[index]);
        selectorDiv.remove();
      });
    });

    // 绑定"新增登录"事件
    const addNewOption = selectorDiv.querySelector('.add-new-option');
    if (addNewOption) {
      addNewOption.addEventListener('click', () => {
        selectorDiv.remove();
        console.log('🆕 用户选择新增登录');
        // 这里可以触发新增账号的流程
      });
    }

    // 点击外部关闭
    const closeHandler = (event) => {
      if (!selectorDiv.contains(event.target) && !targetField.contains(event.target)) {
        selectorDiv.remove();
        document.removeEventListener('click', closeHandler);
      }
    };

    // 延迟添加点击监听，避免立即触发
    setTimeout(() => {
      document.addEventListener('click', closeHandler);
    }, 100);

    // ESC键关闭
    const escHandler = (event) => {
      if (event.key === 'Escape') {
        selectorDiv.remove();
        document.removeEventListener('keydown', escHandler);
      }
    };
    document.addEventListener('keydown', escHandler);
  }

  // 获取平台图标
  getPlatformIcon(platformName) {
    const platform = platformName.toLowerCase();

    // 常见平台的图标映射
    const iconMap = {
      'google': 'G',
      'gmail': 'G',
      'github': 'GH',
      'facebook': 'F',
      'twitter': 'T',
      'linkedin': 'in',
      'microsoft': 'M',
      'apple': '',
      'amazon': 'A',
      'netflix': 'N',
      'spotify': 'S',
      'instagram': 'IG',
      'youtube': 'YT',
      'localhost': '🏠',
      '127.0.0.1': '🏠'
    };

    // 查找匹配的图标
    for (const [key, icon] of Object.entries(iconMap)) {
      if (platform.includes(key)) {
        return icon;
      }
    }

    // 默认使用平台名称的首字母
    return platformName.charAt(0).toUpperCase();
  }

  // 备用的模态对话框选择器（当无法定位输入框时使用）
  showModalAccountSelector(form, formData, accounts) {
    console.log('📋 显示模态账号选择器');

    // 创建选择器界面
    const selectorDiv = document.createElement('div');
    selectorDiv.id = 'email-server-account-modal';
    selectorDiv.style.cssText = `
      position: fixed;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      background: #fff;
      border-radius: 12px;
      padding: 24px;
      box-shadow: 0 8px 32px rgba(0,0,0,0.2);
      z-index: 10003;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      font-size: 14px;
      max-width: 400px;
      min-width: 320px;
    `;

    let accountsHtml = accounts.map((account, index) => `
      <div class="account-option" data-index="${index}" style="
        display: flex;
        align-items: center;
        padding: 12px;
        border-radius: 8px;
        margin-bottom: 8px;
        cursor: pointer;
        transition: background-color 0.2s;
        border: 1px solid #e8eaed;
      " onmouseover="this.style.backgroundColor='#f8f9fa'" onmouseout="this.style.backgroundColor='white'">
        <div style="
          width: 32px;
          height: 32px;
          border-radius: 6px;
          background: #4285f4;
          display: flex;
          align-items: center;
          justify-content: center;
          margin-right: 12px;
          font-size: 16px;
          color: white;
          font-weight: bold;
        ">
          ${this.getPlatformIcon(account.platform_name)}
        </div>
        <div style="flex: 1;">
          <div style="font-weight: 500; color: #202124; margin-bottom: 2px;">
            ${account.email_address || account.login_username || '未知账号'}
          </div>
          <div style="font-size: 13px; color: #5f6368;">
            ${account.platform_name}
          </div>
        </div>
      </div>
    `).join('');

    selectorDiv.innerHTML = `
      <div style="margin-bottom: 20px; font-weight: 600; color: #202124; text-align: center; font-size: 16px;">
        选择要填充的账号
      </div>
      <div style="margin-bottom: 20px; max-height: 300px; overflow-y: auto;">
        ${accountsHtml}
      </div>
      <div style="text-align: center;">
        <button id="cancel-modal-selector" style="
          padding: 10px 20px;
          background: #f8f9fa;
          color: #5f6368;
          border: 1px solid #dadce0;
          border-radius: 6px;
          cursor: pointer;
          font-size: 14px;
        ">
          取消
        </button>
      </div>
    `;

    // 添加背景遮罩
    const overlay = document.createElement('div');
    overlay.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background: rgba(0,0,0,0.4);
      z-index: 10002;
    `;

    document.body.appendChild(overlay);
    document.body.appendChild(selectorDiv);

    // 绑定事件
    selectorDiv.querySelectorAll('.account-option').forEach((option, index) => {
      option.addEventListener('click', () => {
        this.performAutoFill(form, formData, accounts[index]);
        overlay.remove();
        selectorDiv.remove();
      });
    });

    document.getElementById('cancel-modal-selector').addEventListener('click', () => {
      overlay.remove();
      selectorDiv.remove();
    });

    // 点击遮罩关闭
    overlay.addEventListener('click', () => {
      overlay.remove();
      selectorDiv.remove();
    });
  }

  showNotification(message, type) {
    // 设置默认类型
    if (!type) type = 'success';

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
