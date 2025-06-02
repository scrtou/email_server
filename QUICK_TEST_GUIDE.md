# LinuxDo OAuth2 登录快速测试指南

## 🚀 快速开始

### 1. 配置LinuxDo OAuth2应用

根据您提供的截图，您已经在LinuxDo Connect创建了应用。请按以下步骤完成配置：

1. **应用名称**: 填写您的应用名称（如：邮箱账号管理系统）
2. **应用主页**: 填写 `https://linux.do` 或您的实际域名
3. **应用描述**: 填写应用的简要描述
4. **回调地址**: 填写 `http://localhost:5555/api/v1/auth/oauth2/linuxdo/callback`

点击"保存"后，您将获得：
- **Client ID** (客户端ID)
- **Client Secret** (客户端密钥)

### 2. 配置后端环境变量

```bash
# 复制环境变量模板
cp src/backend/.env.example src/backend/.env

# 编辑配置文件
nano src/backend/.env
```

在 `.env` 文件中填写您从LinuxDo获得的凭据：

```env
# LinuxDo OAuth2 配置
LINUXDO_CLIENT_ID=您的实际Client_ID
LINUXDO_CLIENT_SECRET=您的实际Client_Secret
LINUXDO_REDIRECT_URI=http://localhost:5555/api/v1/auth/oauth2/linuxdo/callback
```

### 3. 启动应用

使用提供的启动脚本：

```bash
./start-dev.sh
```

或者手动启动：

```bash
# 启动后端
cd src/backend
go run main.go &

# 启动前端
cd ../frontend
npm run serve &
```

### 4. 测试OAuth2登录

1. 打开浏览器访问 `http://localhost:8080`
2. 点击登录页面的"使用 Linux.do 登录"按钮
3. 系统会跳转到LinuxDo授权页面
4. 在LinuxDo页面点击"授权"
5. 系统会自动跳转回应用并完成登录

## 🔍 功能验证

### 验证点1：授权URL生成
- 点击"使用 Linux.do 登录"按钮
- 检查是否正确跳转到LinuxDo授权页面
- URL应包含正确的client_id和redirect_uri

### 验证点2：回调处理
- 在LinuxDo完成授权后
- 检查是否正确跳转回应用
- 检查是否显示登录成功消息

### 验证点3：用户信息
- 登录成功后，检查用户信息是否正确显示
- 用户名和邮箱应来自LinuxDo账号
- 用户类型应显示为OAuth2用户

## 🐛 故障排除

### 问题1：回调URL不匹配
**症状**: 授权后显示"redirect_uri_mismatch"错误
**解决**: 确保LinuxDo应用配置中的回调地址与 `.env` 文件中的完全一致

### 问题2：Client Secret错误
**症状**: 获取token时返回401错误
**解决**: 检查 `.env` 文件中的Client Secret是否正确

### 问题3：网络连接问题
**症状**: 无法连接到LinuxDo服务器
**解决**: 检查网络连接，确保可以访问 `connect.linux.do`

### 问题4：前端跳转失败
**症状**: 点击登录按钮没有反应
**解决**: 
- 检查浏览器控制台是否有错误
- 确认后端服务正常运行
- 检查API请求是否成功

## 📝 测试日志

查看后端日志以获取详细信息：

```bash
# 如果使用启动脚本，日志会直接显示在终端
# 或者查看具体的日志文件
tail -f /path/to/log/file
```

关键日志信息：
- OAuth2授权URL生成
- Token交换请求
- 用户信息获取
- 用户创建/更新

## 🎯 下一步

测试成功后，您可以：

1. **自定义UI**: 修改登录按钮样式和文案
2. **添加更多OAuth2提供商**: 扩展支持其他平台
3. **完善用户管理**: 添加账号绑定/解绑功能
4. **部署到生产环境**: 更新回调URL为实际域名

## 📞 技术支持

如果遇到问题，请检查：
1. 环境变量配置是否正确
2. LinuxDo应用配置是否匹配
3. 网络连接是否正常
4. 后端和前端服务是否都在运行

详细的配置说明请参考 `LINUXDO_OAUTH2_SETUP.md` 文件。
