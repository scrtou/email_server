// 内容脚本 - 检测和提取表单信息

class FormDetector {
  constructor() {
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

    // 监听表单提交
    this.listenForFormSubmissions();
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
    const extractedData = this.extractSubmissionData(formData);
    
    if (extractedData.email_address || extractedData.login_username) {
      // 显示保存提示
      this.showSavePrompt(extractedData);
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

    // 5秒后自动消失
    setTimeout(() => {
      if (promptDiv.parentNode) {
        promptDiv.remove();
      }
    }, 5000);
  }

  saveToServer(data) {
    chrome.runtime.sendMessage({
      action: 'saveRegistration',
      data: data
    }, (response) => {
      if (response.success) {
        this.showNotification('账号信息已保存到服务器', 'success');
      } else {
        this.showNotification('保存失败: ' + response.error, 'error');
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
const formDetector = new FormDetector();
