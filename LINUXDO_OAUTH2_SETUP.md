# LinuxDo OAuth2 第三方登录配置指南

本文档介绍如何配置LinuxDo OAuth2第三方登录功能。

## 1. 申请LinuxDo OAuth2应用

### 步骤1：访问LinuxDo Connect
访问 [https://connect.linux.do](https://connect.linux.do) 并登录您的LinuxDo账号。

### 步骤2：创建应用
1. 点击"我的应用"或"创建应用"
2. 填写应用信息：
   - **应用名称**: 您的应用名称（如：邮箱账号管理系统）
   - **应用描述**: 应用的简要描述
   - **回调URL**: `http://localhost:5555/api/v1/auth/oauth2/linuxdo/callback`
   - **应用网站**: 您的应用网站地址（可选）

### 步骤3：获取凭据
创建应用后，您将获得：
- **Client ID**: 客户端ID
- **Client Secret**: 客户端密钥

## 2. 配置后端

### 步骤1：设置环境变量
复制 `src/backend/.env.example` 为 `src/backend/.env`：

```bash
cp src/backend/.env.example src/backend/.env
```

### 步骤2：填写OAuth2配置
编辑 `src/backend/.env` 文件，填写从LinuxDo获得的凭据：

```env
# LinuxDo OAuth2 配置
LINUXDO_CLIENT_ID=your_actual_client_id
LINUXDO_CLIENT_SECRET=your_actual_client_secret
LINUXDO_REDIRECT_URI=http://localhost:5555/api/v1/auth/oauth2/linuxdo/callback
```

**注意**: 
- 如果您的应用运行在不同的端口或域名，请相应修改 `LINUXDO_REDIRECT_URI`
- 确保回调URL与LinuxDo应用配置中的回调URL完全一致

## 3. 功能说明

### OAuth2登录流程
1. 用户点击"使用 Linux.do 登录"按钮
2. 系统跳转到LinuxDo授权页面
3. 用户在LinuxDo确认授权
4. LinuxDo重定向回应用并携带授权码
5. 系统使用授权码获取访问令牌
6. 系统使用访问令牌获取用户信息
7. 系统创建或更新用户账号并完成登录

### 用户账号处理
- **新用户**: 自动创建账号，用户名和邮箱来自LinuxDo
- **已有邮箱**: 如果邮箱已存在，将绑定LinuxDo账号
- **已绑定用户**: 直接登录并更新用户信息

### 安全特性
- 使用state参数防止CSRF攻击
- 访问令牌仅用于获取用户信息，不存储
- 支持与传统用户名密码登录并存

## 4. 测试配置

### 启动应用
```bash
cd src/backend
go run main.go
```

### 测试OAuth2登录
1. 访问 `http://localhost:5555` (前端)
2. 点击登录页面的"使用 Linux.do 登录"按钮
3. 完成LinuxDo授权流程
4. 确认成功登录到系统

## 5. 生产环境配置

### 域名配置
在生产环境中，请：
1. 更新LinuxDo应用的回调URL为实际域名
2. 修改环境变量中的 `LINUXDO_REDIRECT_URI`
3. 确保HTTPS配置正确

### 安全建议
- 使用强随机字符串作为 `JWT_SECRET`
- 定期轮换 `LINUXDO_CLIENT_SECRET`
- 启用HTTPS以保护OAuth2流程
- 监控OAuth2登录日志

## 6. 故障排除

### 常见问题
1. **回调URL不匹配**: 确保LinuxDo应用配置与环境变量中的URL完全一致
2. **Client Secret错误**: 检查环境变量中的密钥是否正确
3. **网络问题**: 确保服务器可以访问 `connect.linux.do`

### 调试日志
查看后端日志以获取详细的错误信息：
```bash
tail -f /path/to/your/log/file
```

## 7. API端点

系统提供以下OAuth2相关的API端点：

- `GET /api/v1/auth/oauth2/linuxdo/login` - 获取授权URL
- `GET /api/v1/auth/oauth2/linuxdo/callback` - 处理OAuth2回调

这些端点已集成到前端登录流程中，通常不需要直接调用。
