# Email Server API 集成文档

## 1. API 概览

浏览器插件与 Email Server 后端通过 RESTful API 进行通信，使用 JWT Token 进行身份验证。

### 1.1 基础信息

- **协议**: HTTP/HTTPS
- **数据格式**: JSON
- **认证方式**: JWT Bearer Token
- **编码**: UTF-8

### 1.2 服务器配置

```javascript
// 默认配置
const DEFAULT_CONFIG = {
  serverURL: '',  // 需要用户在设置中配置
  apiVersion: 'v1',
  timeout: 10000,  // 10秒超时
  retryAttempts: 3
};

// 生产环境配置示例
const PRODUCTION_CONFIG = {
  serverURL: 'https://your-domain.com',
  apiVersion: 'v1',
  timeout: 15000,
  retryAttempts: 2
};
```

## 2. 认证 API

### 2.1 用户登录

**端点**: `POST /api/v1/auth/login`

**请求头**:
```
Content-Type: application/json
```

**请求体**:
```json
{
  "username": "string",
  "password": "string"
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin"
    }
  }
}
```

**错误响应**:
```json
{
  "code": 401,
  "message": "用户名或密码错误"
}
```

### 2.2 Token 刷新

**端点**: `POST /api/v1/auth/refresh`

**请求头**:
```
Content-Type: application/json
Authorization: Bearer <refresh_token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "Token刷新成功",
  "data": {
    "access_token": "new_access_token_here",
    "expires_in": 86400
  }
}
```

## 3. 平台注册 API

### 3.1 创建平台注册信息

**端点**: `POST /api/v1/platform-registrations/by-name`

**请求头**:
```
Content-Type: application/json
Authorization: Bearer <access_token>
```

**请求体**:
```json
{
  "platform_name": "github.com",
  "email_address": "user@example.com",
  "login_username": "username",
  "login_password": "password123",
  "notes": "自动检测于 2024-01-01 12:00:00",
  "phone_number": "+1234567890"
}
```

**字段说明**:
- `platform_name`: 平台名称（必填）
- `email_address`: 邮箱地址（可选）
- `login_username`: 登录用户名（可选）
- `login_password`: 登录密码（可选，最少6位）
- `notes`: 备注信息（可选）
- `phone_number`: 手机号码（可选）

**响应示例**:
```json
{
  "code": 201,
  "message": "创建成功",
  "data": {
    "id": 123,
    "user_id": 1,
    "platform_id": 45,
    "platform_name": "github.com",
    "email_account_id": 67,
    "email_address": "user@example.com",
    "login_username": "username",
    "notes": "自动检测于 2024-01-01 12:00:00",
    "phone_number": "+1234567890",
    "created_at": "2024-01-01 12:00:00",
    "updated_at": "2024-01-01 12:00:00"
  }
}
```

**错误响应**:
```json
{
  "code": 400,
  "message": "请求参数错误: platform_name不能为空"
}
```

```json
{
  "code": 409,
  "message": "创建失败，注册信息与现有记录冲突"
}
```

### 3.2 获取平台注册列表

**端点**: `GET /api/v1/platform-registrations`

**请求头**:
```
Authorization: Bearer <access_token>
```

**查询参数**:
- `page`: 页码（默认: 1）
- `pageSize`: 每页数量（默认: 10，最大: 100）
- `platform_name`: 平台名称过滤（可选）
- `email_address`: 邮箱地址过滤（可选）

**请求示例**:
```
GET /api/v1/platform-registrations?page=1&pageSize=20&platform_name=github
```

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "id": 123,
      "platform_name": "github.com",
      "email_address": "user@example.com",
      "login_username": "username",
      "notes": "自动检测于 2024-01-01 12:00:00",
      "created_at": "2024-01-01 12:00:00",
      "updated_at": "2024-01-01 12:00:00"
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 1,
    "totalPages": 1
  }
}
```

### 3.3 获取单个注册信息

**端点**: `GET /api/v1/platform-registrations/{id}`

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "id": 123,
    "platform_name": "github.com",
    "email_address": "user@example.com",
    "login_username": "username",
    "notes": "自动检测于 2024-01-01 12:00:00",
    "phone_number": "+1234567890",
    "created_at": "2024-01-01 12:00:00",
    "updated_at": "2024-01-01 12:00:00"
  }
}
```

### 3.4 获取注册密码

**端点**: `GET /api/v1/platform-registrations/{id}/password`

**请求头**:
```
Authorization: Bearer <access_token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "login_password": "decrypted_password_here"
  }
}
```

**安全说明**: 密码在数据库中加密存储，只有在明确请求时才解密返回。

## 4. 健康检查 API

### 4.1 服务器状态检查

**端点**: `GET /api/v1/health`

**请求头**: 无需认证

**响应示例**:
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00.000Z"
}
```

## 5. 错误处理

### 5.1 HTTP 状态码

| 状态码 | 说明 | 处理方式 |
|--------|------|----------|
| 200 | 成功 | 正常处理响应数据 |
| 201 | 创建成功 | 正常处理响应数据 |
| 400 | 请求参数错误 | 显示错误信息，要求用户修正 |
| 401 | 未认证或认证过期 | 重新登录 |
| 403 | 权限不足 | 显示权限错误信息 |
| 404 | 资源不存在 | 显示资源不存在信息 |
| 409 | 数据冲突 | 显示冲突信息，提供解决方案 |
| 500 | 服务器内部错误 | 显示通用错误信息，建议重试 |

### 5.2 错误响应格式

```json
{
  "code": 400,
  "message": "具体的错误描述信息",
  "details": {
    "field": "field_name",
    "error": "field_specific_error"
  }
}
```

### 5.3 插件错误处理

```javascript
class APIErrorHandler {
  static handle(error, context) {
    if (error.response) {
      // HTTP 错误响应
      const { status, data } = error.response;
      
      switch (status) {
        case 401:
          this.handleAuthError();
          break;
        case 403:
          this.handlePermissionError(data.message);
          break;
        case 409:
          this.handleConflictError(data.message);
          break;
        default:
          this.handleGenericError(data.message || '请求失败');
      }
    } else if (error.request) {
      // 网络错误
      this.handleNetworkError();
    } else {
      // 其他错误
      this.handleGenericError(error.message);
    }
  }
  
  static handleAuthError() {
    // 清除本地 token
    chrome.storage.sync.remove('token');
    // 显示登录提示
    this.showMessage('登录已过期，请重新登录', 'error');
  }
  
  static handleNetworkError() {
    this.showMessage('网络连接失败，请检查网络设置', 'error');
  }
  
  static handleConflictError(message) {
    this.showMessage(`数据冲突: ${message}`, 'warning');
  }
  
  static showMessage(message, type) {
    // 实现消息显示逻辑
    console.log(`[${type.toUpperCase()}] ${message}`);
  }
}
```

## 6. 请求重试机制

```javascript
class APIRetryHandler {
  static async request(url, options, maxRetries = 3) {
    let lastError;
    
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        const response = await fetch(url, options);
        
        if (response.ok) {
          return response;
        }
        
        // 不重试的状态码
        if ([400, 401, 403, 404].includes(response.status)) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        lastError = new Error(`HTTP ${response.status}: ${response.statusText}`);
      } catch (error) {
        lastError = error;
        
        // 网络错误才重试
        if (error.name === 'TypeError' && attempt < maxRetries) {
          await this.delay(Math.pow(2, attempt) * 1000); // 指数退避
          continue;
        }
      }
    }
    
    throw lastError;
  }
  
  static delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

## 7. 数据验证

### 7.1 客户端验证

```javascript
class DataValidator {
  static validateRegistrationData(data) {
    const errors = [];
    
    // 平台名称验证
    if (!data.platform_name || data.platform_name.trim().length === 0) {
      errors.push('平台名称不能为空');
    }
    
    // 邮箱验证
    if (data.email_address && !this.isValidEmail(data.email_address)) {
      errors.push('邮箱格式不正确');
    }
    
    // 密码验证
    if (data.login_password && data.login_password.length < 6) {
      errors.push('密码长度不能少于6位');
    }
    
    return {
      isValid: errors.length === 0,
      errors
    };
  }
  
  static isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }
}
```

这个 API 集成文档详细说明了浏览器插件与 Email Server 后端的所有交互接口，包括请求格式、响应格式、错误处理和最佳实践，为开发和维护提供了完整的参考。
