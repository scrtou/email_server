# 开发者指南

## 1. 开发环境搭建

### 1.1 前置要求

- **Node.js**: 16.0+ (用于开发工具)
- **Chrome**: 88+ 或其他 Chromium 内核浏览器
- **Email Server**: 后端服务运行中
- **Git**: 版本控制

### 1.2 项目结构

```
browser-extension/
├── manifest.json              # 扩展配置文件
├── background.js             # 后台脚本
├── content.js               # 内容脚本
├── popup.html               # 弹窗界面
├── popup.js                 # 弹窗逻辑
├── options.html             # 设置页面
├── options.js               # 设置逻辑
├── icons/                   # 图标资源
├── test-page.html          # 测试页面
├── package.sh              # 打包脚本
├── docs/                   # 文档目录
│   ├── README.md
│   ├── INSTALL.md
│   ├── DESIGN_IMPLEMENTATION.md
│   ├── API_INTEGRATION.md
│   └── DEVELOPER_GUIDE.md
└── tests/                  # 测试文件（可选）
```

### 1.3 开发环境配置

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd email_server/browser-extension
   ```

2. **配置 Email Server**
   ```bash
   # 确保 Email Server 运行在 localhost:8080 或 localhost:5555
   cd ../src/backend
   go run main.go
   ```

3. **加载扩展到浏览器**
   - 打开 Chrome，访问 `chrome://extensions/`
   - 开启"开发者模式"
   - 点击"加载已解压的扩展程序"
   - 选择 `browser-extension` 文件夹

## 2. 核心概念

### 2.1 Manifest V3

本插件使用 Chrome Extension Manifest V3，主要特点：

- **Service Worker**: 替代 Background Pages
- **声明式权限**: 更严格的权限控制
- **动态内容脚本**: 按需注入脚本

### 2.2 组件通信

```javascript
// 消息传递示例
// Content Script → Background Script
chrome.runtime.sendMessage({
  action: 'saveRegistration',
  data: formData
}, (response) => {
  console.log('Response:', response);
});

// Background Script 处理消息
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'saveRegistration') {
    handleSaveRegistration(request.data)
      .then(result => sendResponse(result))
      .catch(error => sendResponse({ success: false, error: error.message }));
    return true; // 保持消息通道开放
  }
});
```

### 2.3 存储机制

```javascript
// 使用 Chrome Storage API
// 同步存储（跨设备同步）
chrome.storage.sync.set({ key: 'value' });
chrome.storage.sync.get(['key'], (result) => {
  console.log('Value:', result.key);
});

// 本地存储（仅本地）
chrome.storage.local.set({ key: 'value' });
chrome.storage.local.get(['key'], (result) => {
  console.log('Value:', result.key);
});
```

## 3. 开发工作流

### 3.1 代码修改流程

1. **修改代码**
   ```bash
   # 编辑相关文件
   vim background.js
   vim content.js
   vim popup.js
   ```

2. **重载扩展**
   - 在 `chrome://extensions/` 页面点击扩展的"刷新"按钮
   - 或使用快捷键 `Ctrl+R`（在扩展页面）

3. **测试功能**
   - 打开 `test-page.html` 进行功能测试
   - 检查浏览器控制台的错误信息

### 3.2 调试技巧

#### 3.2.1 Content Script 调试
```javascript
// 在网页控制台中调试
console.log('Content script loaded');
console.log('Detected forms:', detectedForms);

// 使用 debugger 断点
debugger;
```

#### 3.2.2 Background Script 调试
```javascript
// 在扩展页面点击"检查视图"→"Service Worker"
console.log('Background script event:', event);

// 查看存储数据
chrome.storage.sync.get(null, (data) => {
  console.log('All stored data:', data);
});
```

#### 3.2.3 Popup 调试
```javascript
// 右键扩展图标 → "检查弹出内容"
console.log('Popup opened');

// 查看元素状态
document.getElementById('login-form').addEventListener('submit', (e) => {
  console.log('Form submitted:', e);
});
```

### 3.3 常见问题解决

#### 3.3.1 权限问题
```json
// manifest.json 中添加必要权限
{
  "permissions": ["activeTab", "storage", "scripting"],
  "host_permissions": ["http://localhost:*/*"]
}
```

#### 3.3.2 CORS 问题
```javascript
// 确保 Email Server 配置了正确的 CORS
// 检查 src/backend/middleware/cors.go
```

#### 3.3.3 内容脚本注入失败
```javascript
// 检查页面是否支持内容脚本
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initContentScript);
} else {
  initContentScript();
}
```

## 4. 扩展开发

### 4.1 添加新的表单检测规则

```javascript
// 在 content.js 中添加新的检测规则
const customDetectionRules = {
  'example.com': {
    formSelector: '.custom-login-form',
    usernameField: 'input[name="user_id"]',
    passwordField: 'input[name="user_password"]',
    emailField: 'input[name="user_email"]'
  }
};

function applyCustomRules(hostname) {
  const rules = customDetectionRules[hostname];
  if (rules) {
    return detectFormWithRules(rules);
  }
  return detectFormDefault();
}
```

### 4.2 添加新的 API 端点

```javascript
// 在 background.js 中添加新的 API 方法
class EmailServerAPI {
  async getAccountStatistics() {
    try {
      const response = await fetch(`${this.baseURL}/api/v1/statistics`, {
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
```

### 4.3 添加新的 UI 组件

```html
<!-- 在 popup.html 中添加新标签页 -->
<div class="tabs">
  <div class="tab" data-tab="login">登录</div>
  <div class="tab" data-tab="accounts">账号列表</div>
  <div class="tab" data-tab="manual">手动添加</div>
  <div class="tab" data-tab="statistics">统计</div> <!-- 新增 -->
</div>

<div id="statistics-tab" class="tab-content">
  <div id="statistics-content">
    <!-- 统计内容 -->
  </div>
</div>
```

```javascript
// 在 popup.js 中添加对应逻辑
async loadStatistics() {
  const result = await this.sendMessage({ action: 'getStatistics' });
  if (result.success) {
    this.displayStatistics(result.data);
  } else {
    this.showError('statistics', result.error);
  }
}
```

## 5. 测试指南

### 5.1 单元测试

```javascript
// 创建 tests/unit.test.js
describe('FormDetector', () => {
  beforeEach(() => {
    // 设置测试环境
    document.body.innerHTML = '';
  });
  
  test('should detect login form', () => {
    // 创建测试表单
    const form = document.createElement('form');
    const usernameInput = document.createElement('input');
    usernameInput.type = 'text';
    usernameInput.name = 'username';
    
    const passwordInput = document.createElement('input');
    passwordInput.type = 'password';
    passwordInput.name = 'password';
    
    form.appendChild(usernameInput);
    form.appendChild(passwordInput);
    document.body.appendChild(form);
    
    // 测试检测功能
    const detector = new FormDetector();
    const result = detector.analyzeForm(form);
    
    expect(result.isLoginForm).toBe(true);
    expect(result.fields.length).toBe(2);
  });
});
```

### 5.2 集成测试

```javascript
// 创建 tests/integration.test.js
describe('API Integration', () => {
  let api;
  
  beforeEach(() => {
    api = new EmailServerAPI();
    api.baseURL = 'http://localhost:8080';
  });
  
  test('should login successfully', async () => {
    const result = await api.login('admin', 'password');
    expect(result.success).toBe(true);
    expect(result.data.access_token).toBeDefined();
  });
  
  test('should save registration', async () => {
    // 先登录
    await api.login('admin', 'password');
    
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

### 5.3 端到端测试

```javascript
// 使用 Puppeteer 进行端到端测试
const puppeteer = require('puppeteer');

describe('E2E Tests', () => {
  let browser, page;
  
  beforeAll(async () => {
    browser = await puppeteer.launch({
      headless: false,
      args: [
        '--load-extension=./browser-extension',
        '--disable-extensions-except=./browser-extension'
      ]
    });
    page = await browser.newPage();
  });
  
  afterAll(async () => {
    await browser.close();
  });
  
  test('should detect form and show save prompt', async () => {
    await page.goto('file://' + __dirname + '/test-page.html');
    
    // 填写表单
    await page.type('#login-username', 'testuser');
    await page.type('#login-password', 'testpass');
    
    // 提交表单
    await page.click('button[type="submit"]');
    
    // 等待保存提示出现
    await page.waitForSelector('#email-server-save-prompt', { timeout: 5000 });
    
    // 验证提示内容
    const promptText = await page.$eval('#email-server-save-prompt', el => el.textContent);
    expect(promptText).toContain('检测到账号信息');
  });
});
```

## 6. 性能优化

### 6.1 内存优化

```javascript
// 使用 WeakMap 避免内存泄漏
const formCache = new WeakMap();

function cacheFormData(form, data) {
  formCache.set(form, data);
}

function getCachedFormData(form) {
  return formCache.get(form);
}

// 定期清理
setInterval(() => {
  // 清理已失效的 DOM 引用
  cleanupDetachedElements();
}, 5 * 60 * 1000);
```

### 6.2 性能监控

```javascript
// 性能监控
class PerformanceTracker {
  static startTimer(name) {
    performance.mark(`${name}_start`);
  }
  
  static endTimer(name) {
    performance.mark(`${name}_end`);
    performance.measure(name, `${name}_start`, `${name}_end`);
    
    const measure = performance.getEntriesByName(name)[0];
    if (measure.duration > 100) { // 超过100ms的操作
      console.warn(`Slow operation: ${name} took ${measure.duration}ms`);
    }
  }
}

// 使用示例
PerformanceTracker.startTimer('form_detection');
detectForms();
PerformanceTracker.endTimer('form_detection');
```

## 7. 发布流程

### 7.1 版本管理

```bash
# 更新版本号
vim manifest.json  # 修改 version 字段

# 创建发布标签
git tag v1.0.1
git push origin v1.0.1
```

### 7.2 打包发布

```bash
# 使用打包脚本
./package.sh

# 手动打包
zip -r email-server-extension-v1.0.1.zip . \
  -x "*.git*" "*.md" "test-*" "tests/*" "docs/*"
```

### 7.3 发布检查清单

- [ ] 功能测试通过
- [ ] 性能测试通过
- [ ] 安全审查完成
- [ ] 文档更新完成
- [ ] 版本号已更新
- [ ] 打包文件已生成
- [ ] 发布说明已准备

## 8. 贡献指南

### 8.1 代码规范

```javascript
// 使用 JSDoc 注释
/**
 * 检测表单类型
 * @param {HTMLFormElement} form - 要检测的表单元素
 * @returns {Object} 检测结果
 */
function analyzeForm(form) {
  // 实现代码
}

// 使用一致的命名规范
const API_ENDPOINTS = {
  LOGIN: '/api/v1/auth/login',
  REGISTRATIONS: '/api/v1/platform-registrations'
};
```

### 8.2 提交规范

```bash
# 提交消息格式
git commit -m "feat: 添加新的表单检测规则"
git commit -m "fix: 修复登录状态检查问题"
git commit -m "docs: 更新 API 文档"
git commit -m "test: 添加表单检测单元测试"
```

### 8.3 Pull Request 流程

1. Fork 项目
2. 创建功能分支
3. 编写代码和测试
4. 提交 Pull Request
5. 代码审查
6. 合并到主分支

这个开发者指南为团队协作和项目维护提供了完整的技术指导，确保代码质量和开发效率。
