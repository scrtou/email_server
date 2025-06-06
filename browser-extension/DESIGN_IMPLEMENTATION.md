# Email Server 浏览器插件 - 设计实现说明

## 1. 总体架构设计

### 1.1 架构概览

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Pages     │    │  Browser Ext    │    │  Email Server   │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │Login Forms  │◄┼────┼►│Content Script│ │    │ │   REST API  │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│                 │    │        │        │    │        ▲        │
│                 │    │        ▼        │    │        │        │
│                 │    │ ┌─────────────┐ │    │        │        │
│                 │    │ │Background   │◄┼────┼────────┘        │
│                 │    │ │Script       │ │    │                 │
│                 │    │ └─────────────┘ │    │                 │
│                 │    │        ▲        │    │                 │
│                 │    │        │        │    │                 │
│                 │    │ ┌─────────────┐ │    │                 │
│                 │    │ │Popup UI     │ │    │                 │
│                 │    │ └─────────────┘ │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 1.2 组件职责分工

| 组件 | 职责 | 通信方式 |
|------|------|----------|
| Content Script | 表单检测、信息提取、用户交互 | DOM 操作、消息传递 |
| Background Script | API 通信、数据存储、状态管理 | Chrome APIs、HTTP 请求 |
| Popup UI | 用户界面、配置管理、数据展示 | 消息传递、本地存储 |
| Options Page | 设置配置、连接测试、高级选项 | 消息传递、本地存储 |

## 2. 核心算法设计

### 2.1 表单检测算法

#### 2.1.1 检测策略
```javascript
// 多层次检测策略
const FormDetectionStrategy = {
  // 1. 静态检测：页面加载完成后扫描
  staticDetection: () => {
    document.querySelectorAll('form').forEach(analyzeForm);
  },
  
  // 2. 动态检测：监听 DOM 变化
  dynamicDetection: () => {
    new MutationObserver(handleDOMChanges).observe(document.body, {
      childList: true,
      subtree: true
    });
  },
  
  // 3. 事件检测：监听表单提交
  eventDetection: () => {
    document.addEventListener('submit', handleFormSubmit, true);
  }
};
```

#### 2.1.2 表单分析算法
```javascript
// 表单类型识别算法
function analyzeForm(form) {
  const analysis = {
    isLoginForm: false,
    isRegisterForm: false,
    confidence: 0,
    fields: extractFields(form)
  };
  
  // 基于字段类型判断
  const hasPassword = analysis.fields.some(f => f.type === 'password');
  const hasEmail = analysis.fields.some(f => f.type === 'email');
  const hasUsername = analysis.fields.some(f => f.isUsername);
  
  // 基于文本内容判断
  const formText = form.textContent.toLowerCase();
  const loginKeywords = ['login', 'sign in', '登录', '登入'];
  const registerKeywords = ['register', 'sign up', '注册', '注册'];
  
  // 综合判断逻辑
  if (hasPassword && (hasEmail || hasUsername)) {
    if (loginKeywords.some(k => formText.includes(k))) {
      analysis.isLoginForm = true;
      analysis.confidence = 0.8;
    } else if (registerKeywords.some(k => formText.includes(k))) {
      analysis.isRegisterForm = true;
      analysis.confidence = 0.7;
    }
  }
  
  return analysis;
}
```

### 2.2 字段识别算法

#### 2.2.1 字段类型识别
```javascript
function identifyFieldType(input) {
  const indicators = {
    email: {
      type: ['email'],
      name: ['email', 'mail', 'e-mail'],
      id: ['email', 'mail', 'e-mail'],
      placeholder: ['email', 'mail', '邮箱'],
      pattern: /email|mail/i
    },
    username: {
      name: ['username', 'user', 'login', 'account'],
      id: ['username', 'user', 'login', 'account'],
      placeholder: ['username', 'user', '用户名', '账号'],
      pattern: /user|login|account/i
    },
    password: {
      type: ['password'],
      name: ['password', 'pass', 'pwd'],
      id: ['password', 'pass', 'pwd'],
      placeholder: ['password', 'pass', '密码'],
      pattern: /password|pass|pwd/i
    }
  };
  
  // 权重计算
  let maxScore = 0;
  let detectedType = 'unknown';
  
  Object.entries(indicators).forEach(([type, rules]) => {
    let score = 0;
    
    if (rules.type && rules.type.includes(input.type)) score += 10;
    if (rules.name && rules.name.some(n => input.name.toLowerCase().includes(n))) score += 8;
    if (rules.id && rules.id.some(i => input.id.toLowerCase().includes(i))) score += 8;
    if (rules.placeholder && rules.placeholder.some(p => 
      input.placeholder.toLowerCase().includes(p))) score += 6;
    if (rules.pattern && rules.pattern.test(input.name + input.id + input.placeholder)) score += 4;
    
    if (score > maxScore) {
      maxScore = score;
      detectedType = type;
    }
  });
  
  return { type: detectedType, confidence: maxScore / 10 };
}
```

### 2.3 平台识别算法

```javascript
function identifyPlatform(url) {
  const hostname = new URL(url).hostname;
  
  // 移除常见前缀
  const cleanHostname = hostname.replace(/^(www\.|m\.|mobile\.|login\.|auth\.)/, '');
  
  // 特殊平台映射
  const platformMap = {
    'github.com': 'GitHub',
    'google.com': 'Google',
    'facebook.com': 'Facebook',
    'twitter.com': 'Twitter',
    'linkedin.com': 'LinkedIn'
  };
  
  return platformMap[cleanHostname] || cleanHostname;
}
```

## 3. 数据流设计

### 3.1 数据流向图

```
User Action → Content Script → Background Script → Email Server API
     ↓              ↓                ↓                    ↓
Form Submit → Extract Data → Process & Store → Save to Database
     ↓              ↓                ↓                    ↓
Show Prompt ← Format Data ← API Response ← Return Result
```

### 3.2 消息传递机制

#### 3.2.1 消息类型定义
```javascript
const MessageTypes = {
  // Content Script → Background Script
  SAVE_REGISTRATION: 'saveRegistration',
  GET_CONFIG: 'getConfig',
  
  // Popup → Background Script
  LOGIN: 'login',
  GET_REGISTRATIONS: 'getRegistrations',
  SAVE_CONFIG: 'saveConfig',
  
  // Background Script → Content Script
  START_DETECTION: 'startFormDetection',
  CONFIG_UPDATED: 'configUpdated'
};
```

#### 3.2.2 消息处理流程
```javascript
// Background Script 消息路由
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  const handlers = {
    [MessageTypes.SAVE_REGISTRATION]: handleSaveRegistration,
    [MessageTypes.LOGIN]: handleLogin,
    [MessageTypes.GET_REGISTRATIONS]: handleGetRegistrations,
    [MessageTypes.GET_CONFIG]: handleGetConfig,
    [MessageTypes.SAVE_CONFIG]: handleSaveConfig
  };
  
  const handler = handlers[request.action];
  if (handler) {
    handler(request, sender, sendResponse);
    return true; // 保持消息通道开放
  }
});
```

## 4. 存储设计

### 4.1 存储结构

```javascript
// Chrome Storage 数据结构
const StorageSchema = {
  // 服务器配置
  serverURL: 'string',           // Email Server 地址
  token: 'string',               // JWT Token
  
  // 用户配置
  username: 'string',            // 用户名（可选）
  password: 'string',            // 密码（可选，加密存储）
  
  // 插件设置
  autoDetect: 'boolean',         // 自动检测开关
  showNotifications: 'boolean',  // 显示通知开关
  autoSave: 'boolean',          // 自动保存开关
  excludedSites: 'string',      // 排除网站列表
  
  // 缓存数据
  lastSync: 'timestamp',        // 最后同步时间
  cachedRegistrations: 'array'  // 缓存的注册信息
};
```

### 4.2 数据安全

```javascript
// 敏感数据加密存储
class SecureStorage {
  static async setSecure(key, value) {
    const encrypted = await this.encrypt(value);
    return chrome.storage.sync.set({ [key]: encrypted });
  }
  
  static async getSecure(key) {
    const result = await chrome.storage.sync.get(key);
    return result[key] ? await this.decrypt(result[key]) : null;
  }
  
  static async encrypt(data) {
    // 使用 Web Crypto API 加密
    const key = await crypto.subtle.generateKey(
      { name: 'AES-GCM', length: 256 },
      false,
      ['encrypt', 'decrypt']
    );
    // ... 加密实现
  }
}
```

## 5. 用户界面设计

### 5.1 界面层次结构

```
Extension UI
├── Popup Interface
│   ├── Login Tab
│   ├── Account List Tab
│   └── Manual Add Tab
├── Options Page
│   ├── Server Settings
│   ├── Advanced Options
│   └── Connection Test
└── Content Overlays
    ├── Save Prompt
    └── Notification Toast
```

### 5.2 响应式设计

```css
/* 弹窗界面响应式设计 */
.popup-container {
  width: 350px;
  min-height: 400px;
  max-height: 600px;
}

@media (max-width: 400px) {
  .popup-container {
    width: 300px;
  }
  
  .form-group input {
    font-size: 16px; /* 防止移动端缩放 */
  }
}

/* 设置页面响应式设计 */
.options-container {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

@media (max-width: 768px) {
  .options-container {
    padding: 10px;
  }
}
```

### 5.3 用户体验设计

#### 5.3.1 加载状态
```javascript
// 统一的加载状态管理
class LoadingManager {
  static show(element, message = '加载中...') {
    element.innerHTML = `
      <div class="loading-spinner"></div>
      <div class="loading-text">${message}</div>
    `;
    element.classList.add('loading');
  }
  
  static hide(element, content) {
    element.classList.remove('loading');
    element.innerHTML = content;
  }
}
```

#### 5.3.2 错误处理
```javascript
// 统一的错误处理
class ErrorHandler {
  static show(element, error, retry = null) {
    element.innerHTML = `
      <div class="error-icon">⚠️</div>
      <div class="error-message">${error}</div>
      ${retry ? '<button class="retry-btn">重试</button>' : ''}
    `;
    element.classList.add('error');
  }
  
  static clear(element) {
    element.classList.remove('error');
  }
}
```

## 6. 性能优化设计

### 6.1 检测性能优化

```javascript
// 防抖处理，避免频繁检测
const debouncedDetection = debounce(() => {
  detectForms();
}, 300);

// 智能检测，只在必要时触发
function shouldDetectOnPage(url) {
  const excludedPatterns = [
    /\/api\//,
    /\.json$/,
    /\.xml$/,
    /\.pdf$/
  ];
  
  return !excludedPatterns.some(pattern => pattern.test(url));
}

// 缓存检测结果
const detectionCache = new Map();
function getCachedDetection(formSignature) {
  return detectionCache.get(formSignature);
}
```

### 6.2 内存管理

```javascript
// 清理机制
class MemoryManager {
  static cleanup() {
    // 清理过期的检测缓存
    const now = Date.now();
    for (const [key, value] of detectionCache.entries()) {
      if (now - value.timestamp > 5 * 60 * 1000) { // 5分钟过期
        detectionCache.delete(key);
      }
    }
    
    // 清理 DOM 引用
    this.cleanupDOMReferences();
  }
  
  static cleanupDOMReferences() {
    // 移除已失效的 DOM 元素引用
    detectedForms.forEach((form, index) => {
      if (!document.contains(form)) {
        detectedForms.splice(index, 1);
      }
    });
  }
}

// 定期清理
setInterval(() => {
  MemoryManager.cleanup();
}, 5 * 60 * 1000); // 每5分钟清理一次
```

## 7. 安全设计

### 7.1 权限最小化

```json
// manifest.json 权限设计
{
  "permissions": [
    "activeTab",      // 只访问当前活动标签页
    "storage",        // 本地存储
    "scripting"       // 脚本注入
  ],
  "host_permissions": [
    "http://localhost:*/*",  // 开发环境
    "https://localhost:*/*"  // 开发环境 HTTPS
  ]
}
```

### 7.2 数据验证

```javascript
// 输入验证
class InputValidator {
  static validateEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }
  
  static validateURL(url) {
    try {
      new URL(url);
      return true;
    } catch {
      return false;
    }
  }
  
  static sanitizeInput(input) {
    return input.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '');
  }
}
```

### 7.3 CSP 安全策略

```json
// Content Security Policy
{
  "content_security_policy": {
    "extension_pages": "script-src 'self'; object-src 'self'"
  }
}
```

## 8. 测试设计

### 8.1 单元测试

```javascript
// 表单检测测试
describe('FormDetector', () => {
  test('should detect login form', () => {
    const form = createMockLoginForm();
    const result = FormDetector.analyzeForm(form);
    expect(result.isLoginForm).toBe(true);
    expect(result.confidence).toBeGreaterThan(0.7);
  });
  
  test('should extract email field', () => {
    const form = createMockFormWithEmail();
    const fields = FormDetector.extractFields(form);
    const emailField = fields.find(f => f.type === 'email');
    expect(emailField).toBeDefined();
  });
});
```

### 8.2 集成测试

```javascript
// API 集成测试
describe('EmailServerAPI', () => {
  test('should login successfully', async () => {
    const api = new EmailServerAPI();
    const result = await api.login('testuser', 'testpass');
    expect(result.success).toBe(true);
    expect(result.data.access_token).toBeDefined();
  });
  
  test('should save registration', async () => {
    const registrationData = {
      platform_name: 'test.com',
      email_address: 'test@test.com',
      login_username: 'testuser',
      login_password: 'testpass'
    };
    
    const result = await api.createPlatformRegistration(registrationData);
    expect(result.success).toBe(true);
  });
});
```

## 9. 部署和维护

### 9.1 版本管理

```javascript
// 版本检查和更新
class VersionManager {
  static async checkForUpdates() {
    const currentVersion = chrome.runtime.getManifest().version;
    const latestVersion = await this.getLatestVersion();
    
    if (this.isNewerVersion(latestVersion, currentVersion)) {
      this.notifyUpdate(latestVersion);
    }
  }
  
  static compareVersions(v1, v2) {
    const parts1 = v1.split('.').map(Number);
    const parts2 = v2.split('.').map(Number);
    
    for (let i = 0; i < Math.max(parts1.length, parts2.length); i++) {
      const part1 = parts1[i] || 0;
      const part2 = parts2[i] || 0;
      
      if (part1 > part2) return 1;
      if (part1 < part2) return -1;
    }
    
    return 0;
  }
}
```

### 9.2 错误监控

```javascript
// 错误收集和报告
class ErrorReporter {
  static report(error, context) {
    const errorData = {
      message: error.message,
      stack: error.stack,
      context: context,
      timestamp: new Date().toISOString(),
      version: chrome.runtime.getManifest().version,
      userAgent: navigator.userAgent
    };
    
    // 发送到错误监控服务
    this.sendToMonitoring(errorData);
  }
  
  static sendToMonitoring(data) {
    // 实现错误数据发送逻辑
    console.error('Extension Error:', data);
  }
}

// 全局错误处理
window.addEventListener('error', (event) => {
  ErrorReporter.report(event.error, 'global');
});
```

## 10. 兼容性设计

### 10.1 浏览器兼容性

```javascript
// 浏览器特性检测
class BrowserCompatibility {
  static checkSupport() {
    const features = {
      manifestV3: chrome.runtime.getManifest().manifest_version === 3,
      serviceWorker: 'serviceWorker' in navigator,
      storageAPI: !!chrome.storage,
      scriptingAPI: !!chrome.scripting,
      webCrypto: !!window.crypto && !!window.crypto.subtle
    };

    return features;
  }

  static getPolyfills() {
    const polyfills = [];

    // Promise polyfill for older browsers
    if (!window.Promise) {
      polyfills.push('promise-polyfill');
    }

    // Fetch polyfill
    if (!window.fetch) {
      polyfills.push('whatwg-fetch');
    }

    return polyfills;
  }
}
```

### 10.2 网站兼容性

```javascript
// 网站特殊处理规则
const SiteCompatibilityRules = {
  'github.com': {
    formSelector: 'form[action*="session"]',
    usernameField: '#login_field',
    passwordField: '#password',
    submitDelay: 1000
  },

  'google.com': {
    formSelector: 'form[id*="gaia"]',
    usernameField: 'input[type="email"]',
    passwordField: 'input[type="password"]',
    multiStep: true
  },

  'facebook.com': {
    formSelector: '#login_form',
    usernameField: '#email',
    passwordField: '#pass',
    dynamicLoading: true
  }
};

// 应用特殊规则
function applyCompatibilityRules(hostname) {
  const rules = SiteCompatibilityRules[hostname];
  if (rules) {
    return new SiteHandler(rules);
  }
  return new DefaultSiteHandler();
}
```

## 11. 国际化设计

### 11.1 多语言支持

```javascript
// 国际化配置
const i18nConfig = {
  'zh-CN': {
    'extension_name': 'Email Server 账号管理器',
    'login_detected': '检测到登录信息',
    'save_to_server': '保存到服务器',
    'ignore': '忽略',
    'login_success': '登录成功',
    'save_failed': '保存失败'
  },

  'en-US': {
    'extension_name': 'Email Server Account Manager',
    'login_detected': 'Login information detected',
    'save_to_server': 'Save to server',
    'ignore': 'Ignore',
    'login_success': 'Login successful',
    'save_failed': 'Save failed'
  }
};

// 国际化函数
function i18n(key, params = {}) {
  const locale = chrome.i18n.getUILanguage() || 'en-US';
  const messages = i18nConfig[locale] || i18nConfig['en-US'];
  let message = messages[key] || key;

  // 参数替换
  Object.entries(params).forEach(([param, value]) => {
    message = message.replace(`{${param}}`, value);
  });

  return message;
}
```

### 11.2 本地化适配

```javascript
// 日期时间本地化
class LocaleFormatter {
  static formatDateTime(date) {
    const locale = chrome.i18n.getUILanguage();
    return new Intl.DateTimeFormat(locale, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    }).format(date);
  }

  static formatRelativeTime(date) {
    const locale = chrome.i18n.getUILanguage();
    const rtf = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' });

    const diff = Date.now() - date.getTime();
    const minutes = Math.floor(diff / (1000 * 60));
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days > 0) return rtf.format(-days, 'day');
    if (hours > 0) return rtf.format(-hours, 'hour');
    if (minutes > 0) return rtf.format(-minutes, 'minute');
    return rtf.format(0, 'second');
  }
}
```

## 12. 监控和分析

### 12.1 使用统计

```javascript
// 使用情况统计
class UsageAnalytics {
  static trackEvent(category, action, label = '', value = 0) {
    const event = {
      category,
      action,
      label,
      value,
      timestamp: Date.now(),
      version: chrome.runtime.getManifest().version
    };

    this.storeEvent(event);
  }

  static async storeEvent(event) {
    const events = await this.getStoredEvents();
    events.push(event);

    // 保持最近1000个事件
    if (events.length > 1000) {
      events.splice(0, events.length - 1000);
    }

    await chrome.storage.local.set({ analytics_events: events });
  }

  static trackFormDetection(success, formType) {
    this.trackEvent('form_detection', success ? 'success' : 'failure', formType);
  }

  static trackSaveAction(success, platform) {
    this.trackEvent('save_action', success ? 'success' : 'failure', platform);
  }
}
```

### 12.2 性能监控

```javascript
// 性能监控
class PerformanceMonitor {
  static startTimer(name) {
    performance.mark(`${name}_start`);
  }

  static endTimer(name) {
    performance.mark(`${name}_end`);
    performance.measure(name, `${name}_start`, `${name}_end`);

    const measure = performance.getEntriesByName(name)[0];
    if (measure.duration > 1000) { // 超过1秒的操作
      console.warn(`Slow operation: ${name} took ${measure.duration}ms`);
    }

    return measure.duration;
  }

  static monitorMemoryUsage() {
    if (performance.memory) {
      const memory = {
        used: performance.memory.usedJSHeapSize,
        total: performance.memory.totalJSHeapSize,
        limit: performance.memory.jsHeapSizeLimit
      };

      if (memory.used / memory.limit > 0.8) {
        console.warn('High memory usage detected:', memory);
      }

      return memory;
    }
  }
}
```

## 13. 扩展性设计

### 13.1 插件架构

```javascript
// 插件系统设计
class PluginManager {
  constructor() {
    this.plugins = new Map();
    this.hooks = new Map();
  }

  registerPlugin(name, plugin) {
    this.plugins.set(name, plugin);

    // 注册插件的钩子
    if (plugin.hooks) {
      Object.entries(plugin.hooks).forEach(([hookName, handler]) => {
        this.addHook(hookName, handler);
      });
    }
  }

  addHook(name, handler) {
    if (!this.hooks.has(name)) {
      this.hooks.set(name, []);
    }
    this.hooks.get(name).push(handler);
  }

  async executeHook(name, data) {
    const handlers = this.hooks.get(name) || [];
    let result = data;

    for (const handler of handlers) {
      result = await handler(result);
    }

    return result;
  }
}

// 示例插件：密码强度检查
const PasswordStrengthPlugin = {
  name: 'password-strength',
  hooks: {
    'before-save': async (data) => {
      if (data.login_password) {
        const strength = calculatePasswordStrength(data.login_password);
        data.password_strength = strength;
      }
      return data;
    }
  }
};
```

### 13.2 配置系统

```javascript
// 配置管理系统
class ConfigManager {
  constructor() {
    this.config = new Map();
    this.watchers = new Map();
  }

  async load() {
    const stored = await chrome.storage.sync.get('extension_config');
    if (stored.extension_config) {
      Object.entries(stored.extension_config).forEach(([key, value]) => {
        this.config.set(key, value);
      });
    }
  }

  async set(key, value) {
    const oldValue = this.config.get(key);
    this.config.set(key, value);

    // 保存到存储
    const configObj = Object.fromEntries(this.config);
    await chrome.storage.sync.set({ extension_config: configObj });

    // 通知观察者
    this.notifyWatchers(key, value, oldValue);
  }

  get(key, defaultValue = null) {
    return this.config.get(key) ?? defaultValue;
  }

  watch(key, callback) {
    if (!this.watchers.has(key)) {
      this.watchers.set(key, []);
    }
    this.watchers.get(key).push(callback);
  }

  notifyWatchers(key, newValue, oldValue) {
    const callbacks = this.watchers.get(key) || [];
    callbacks.forEach(callback => {
      try {
        callback(newValue, oldValue);
      } catch (error) {
        console.error('Config watcher error:', error);
      }
    });
  }
}
```

## 14. 开发工具和调试

### 14.1 调试工具

```javascript
// 开发模式调试工具
class DebugTools {
  static isDebugMode() {
    return chrome.runtime.getManifest().version.includes('dev') ||
           localStorage.getItem('debug_mode') === 'true';
  }

  static log(level, message, data = null) {
    if (!this.isDebugMode()) return;

    const timestamp = new Date().toISOString();
    const logEntry = {
      timestamp,
      level,
      message,
      data
    };

    console[level](`[${timestamp}] ${message}`, data);

    // 存储调试日志
    this.storeDebugLog(logEntry);
  }

  static async storeDebugLog(entry) {
    const logs = await this.getDebugLogs();
    logs.push(entry);

    // 保持最近500条日志
    if (logs.length > 500) {
      logs.splice(0, logs.length - 500);
    }

    await chrome.storage.local.set({ debug_logs: logs });
  }

  static async exportDebugLogs() {
    const logs = await this.getDebugLogs();
    const blob = new Blob([JSON.stringify(logs, null, 2)], {
      type: 'application/json'
    });

    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `debug-logs-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
  }
}
```

### 14.2 测试工具

```javascript
// 自动化测试工具
class TestRunner {
  constructor() {
    this.tests = [];
    this.results = [];
  }

  addTest(name, testFn) {
    this.tests.push({ name, testFn });
  }

  async runAll() {
    this.results = [];

    for (const test of this.tests) {
      try {
        const startTime = performance.now();
        await test.testFn();
        const duration = performance.now() - startTime;

        this.results.push({
          name: test.name,
          status: 'passed',
          duration
        });
      } catch (error) {
        this.results.push({
          name: test.name,
          status: 'failed',
          error: error.message,
          stack: error.stack
        });
      }
    }

    return this.results;
  }

  generateReport() {
    const passed = this.results.filter(r => r.status === 'passed').length;
    const failed = this.results.filter(r => r.status === 'failed').length;

    return {
      total: this.results.length,
      passed,
      failed,
      passRate: (passed / this.results.length * 100).toFixed(2),
      results: this.results
    };
  }
}

// 示例测试用例
const testRunner = new TestRunner();

testRunner.addTest('Form Detection', async () => {
  const mockForm = createMockLoginForm();
  const result = FormDetector.analyzeForm(mockForm);
  assert(result.isLoginForm === true, 'Should detect login form');
  assert(result.confidence > 0.5, 'Should have reasonable confidence');
});

testRunner.addTest('API Communication', async () => {
  const api = new EmailServerAPI();
  const result = await api.testConnection();
  assert(result.success === true, 'Should connect to server');
});
```

这个设计实现说明文档现在包含了完整的技术架构描述，涵盖了从核心算法到扩展性设计的各个方面，为开发团队提供了全面的技术参考和实现指导。
