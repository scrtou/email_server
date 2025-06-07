# 快速修复指南

## 🚨 问题现象
扩展提示："登录失败: 请先在设置中配置服务器地址"，即使已经在设置页面配置了服务器地址。

## 🔧 快速解决方案

### 方案1：手动设置配置（推荐）

1. **打开扩展popup页面**
2. **按F12打开开发者工具**
3. **在Console中执行以下代码**：

```javascript
// 直接设置配置
chrome.storage.sync.set({
  serverURL: 'https://accountback.azhen.de'
}, () => {
  console.log('配置已设置');
  
  // 通知background.js
  chrome.runtime.sendMessage({
    action: 'saveConfig',
    config: { serverURL: 'https://accountback.azhen.de' }
  }, (response) => {
    console.log('Background已更新:', response);
  });
});
```

4. **刷新扩展**：
   - 在扩展管理页面点击刷新按钮
   - 或者重新加载扩展

5. **测试登录**

### 方案2：使用调试工具

1. **在扩展环境中打开**：
   ```
   chrome-extension://[扩展ID]/storage-debug.html
   ```

2. **保存配置**：
   - 输入服务器地址
   - 点击"保存配置"
   - 确认显示"配置保存成功"

3. **测试Background消息**：
   - 点击"测试Background消息"
   - 确认返回正确的配置

### 方案3：重新安装扩展

1. **卸载当前扩展**
2. **重新加载扩展文件夹**
3. **重新配置设置**

## 🔍 诊断步骤

### 检查配置是否保存

在popup的开发者工具Console中执行：

```javascript
chrome.storage.sync.get(['serverURL', 'token'], (result) => {
  console.log('当前配置:', result);
});
```

应该显示：
```
当前配置: {serverURL: "https://accountback.azhen.de"}
```

### 检查Background状态

在popup的开发者工具Console中执行：

```javascript
chrome.runtime.sendMessage({ action: 'getConfig' }, (response) => {
  console.log('Background配置:', response);
});
```

### 强制重新初始化Background

```javascript
chrome.runtime.sendMessage({
  action: 'saveConfig',
  config: { serverURL: 'https://accountback.azhen.de' }
}, (response) => {
  console.log('强制更新结果:', response);
});
```

## 🛠️ 临时解决方案

如果上述方法都不行，可以修改popup.js，跳过配置检查：

1. **找到popup.js中的登录函数**
2. **临时注释掉配置检查**：

```javascript
async handleLogin() {
  const username = document.getElementById('username').value;
  const password = document.getElementById('password').value;
  
  // 临时跳过配置检查，直接尝试登录
  const result = await fetch('https://accountback.azhen.de/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  });
  
  // 处理响应...
}
```

## 📋 验证清单

完成修复后，验证以下项目：

- [ ] 在Console中能看到正确的配置
- [ ] Background消息返回正确配置
- [ ] 登录不再提示配置错误
- [ ] 能够成功连接到服务器

## 🔄 如果问题仍然存在

1. **检查网络连接**
2. **确认服务器地址正确**
3. **检查扩展权限**
4. **重启浏览器**
5. **清除扩展数据并重新配置**

## 💡 预防措施

为避免此问题再次发生：

1. **每次修改代码后重新加载扩展**
2. **在设置页面保存后等待几秒再测试**
3. **使用调试工具验证配置状态**
4. **定期检查Chrome存储权限**

这个快速修复指南应该能解决大部分配置相关的登录问题。如果问题仍然存在，可能需要更深入的调试。
