# 认证错误故障排除指南

## 🚨 错误现象
进入密码库时提示："加载失败: 认证格式错误"

## 🔍 可能原因

### 1. Token格式问题
- Token为空或undefined
- Token格式不正确
- Token已过期

### 2. 认证头部问题
- Authorization头部格式错误
- 缺少Bearer前缀
- 头部编码问题

### 3. 服务器配置问题
- 服务器期望不同的认证格式
- CORS配置问题
- API端点变更

## 🛠️ 诊断步骤

### 步骤1：使用认证调试工具

1. **打开扩展popup**
2. **右键点击 → 检查**
3. **在Console中复制粘贴** `auth-debug.js` 的全部内容
4. **执行自动诊断**

### 步骤2：检查认证状态

```javascript
// 检查当前认证信息
authDebug.checkAuthStatus()
```

应该显示：
```
存储中的认证信息: {
  serverURL: "https://accountback.azhen.de",
  hasToken: true,
  tokenLength: 200+,
  tokenPreview: "eyJhbGciOiJIUzI1NiIs...",
  username: "your_username"
}
```

### 步骤3：测试服务器连接

```javascript
// 测试服务器是否可访问
authDebug.testServerConnection()
```

### 步骤4：重新登录测试

```javascript
// 完整的认证流程测试
authDebug.fullAuthTest('你的用户名', '你的密码')
```

## 🔧 解决方案

### 方案1：重新登录

如果token有问题，重新登录：

```javascript
// 1. 清除旧认证信息
authDebug.clearAuth()

// 2. 重新登录
authDebug.testLogin('用户名', '密码')

// 3. 测试数据获取
authDebug.testGetData()
```

### 方案2：手动设置Token

如果您有有效的token：

```javascript
// 手动设置token
authDebug.setToken('your_valid_token_here')

// 测试数据获取
authDebug.testGetData()
```

### 方案3：检查服务器响应

查看详细的错误信息：

1. **打开扩展管理页面**
2. **点击"检查视图: Service Worker"**
3. **查看Console中的详细日志**

应该看到类似：
```
📡 发送请求: {
  url: "https://accountback.azhen.de/api/v1/platform-registrations",
  headers: { Authorization: "Bearer eyJhbGci..." }
}
📨 响应状态: 401 Unauthorized
❌ 服务器错误: { message: "认证格式错误" }
```

## 🔍 常见问题

### Q: Token长度为0或undefined
**A**: 登录过程中token没有正确保存
```javascript
// 重新登录
authDebug.testLogin('用户名', '密码')
```

### Q: 服务器返回401错误
**A**: Token无效或已过期
```javascript
// 清除并重新登录
authDebug.clearAuth()
authDebug.testLogin('用户名', '密码')
```

### Q: 服务器返回400错误
**A**: 请求格式问题，检查API端点和参数

### Q: 网络连接错误
**A**: 检查服务器地址和网络连接
```javascript
authDebug.testServerConnection()
```

## 📋 检查清单

- [ ] 服务器地址配置正确
- [ ] 服务器可以正常访问
- [ ] 用户名和密码正确
- [ ] 登录成功并返回token
- [ ] Token正确保存到存储
- [ ] Background.js正确加载token
- [ ] 认证头部格式正确

## 🔄 完整修复流程

### 1. 重置认证状态
```javascript
authDebug.clearAuth()
```

### 2. 重新登录
```javascript
authDebug.testLogin('你的用户名', '你的密码')
```

### 3. 验证认证状态
```javascript
authDebug.checkAuthStatus()
```

### 4. 测试数据获取
```javascript
authDebug.testGetData()
```

### 5. 如果仍然失败
- 检查服务器日志
- 确认API端点是否正确
- 检查服务器的认证配置

## 💡 预防措施

1. **定期检查token有效性**
2. **实现token自动刷新机制**
3. **添加更详细的错误处理**
4. **监控认证状态变化**

## 🆘 如果问题仍然存在

1. **收集调试信息**：
   - 执行 `authDebug.fullAuthTest()`
   - 复制所有Console输出

2. **检查服务器端**：
   - 确认API端点正确
   - 检查认证中间件配置
   - 查看服务器错误日志

3. **网络问题排查**：
   - 检查CORS配置
   - 确认防火墙设置
   - 测试其他网络环境

使用这个调试工具应该能帮您快速定位和解决认证问题！
