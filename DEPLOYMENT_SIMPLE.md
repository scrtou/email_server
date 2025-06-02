# Email Server 简化部署指南 (不使用Nginx)

## 🚀 快速部署

### 1. 环境准备
```bash
# 安装 Docker 和 Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 2. 配置环境变量
```bash
# 复制环境变量模板
cp .env.production.example .env

# 编辑配置文件 (重要!)
nano .env
```

**必须修改的配置**:
```bash
# 端口配置 (可选，使用默认值即可)
FRONTEND_PORT=80        # 前端端口，默认80
BACKEND_PORT=5555       # 后端端口，默认5555

# 强JWT密钥 (至少32个字符)
JWT_SECRET=your-production-super-secret-jwt-key-at-least-32-characters-long

# LinuxDo OAuth2 配置
LINUXDO_CLIENT_ID=your_client_id
LINUXDO_CLIENT_SECRET=your_client_secret

# 回调地址 (修改为您的域名，注意端口号)
LINUXDO_REDIRECT_URI=http://yourdomain.com:5555/api/v1/auth/oauth2/linuxdo/callback

# 前端API地址 (修改为您的域名，注意端口号)
VUE_APP_API_BASE_URL=http://yourdomain.com:5555/api/v1
```

### 3. 一键部署
```bash
chmod +x deploy.sh
./deploy.sh
```

## 📋 服务架构

```
┌─────────────────┐    ┌─────────────────┐
│   前端服务       │    │   后端服务       │
│   Port: 80      │    │   Port: 5555    │
│   (Vue.js)      │    │   (Go API)      │
└─────────────────┘    └─────────────────┘
         │                       │
         └───────────────────────┘
              直接通信 (HTTP)
```

## 🌐 访问地址

- **前端**: `http://yourdomain.com:${FRONTEND_PORT}` (默认80端口可省略)
- **后端API**: `http://yourdomain.com:${BACKEND_PORT}/api/v1` (默认5555)

## ⚙️ 端口配置

您可以在 `.env` 文件中自定义端口：

```bash
# 前端端口配置
FRONTEND_PORT=80        # 生产环境推荐80 (HTTP标准端口)
FRONTEND_PORT=8080      # 或者使用其他端口

# 后端端口配置
BACKEND_PORT=5555       # 默认端口
BACKEND_PORT=3000       # 或者使用其他端口

# HTTPS端口 (可选)
FRONTEND_HTTPS_PORT=443 # HTTPS标准端口
```

**端口选择建议**:
- **前端**: 80 (HTTP) 或 443 (HTTPS) 用于生产环境
- **后端**: 5555 (默认) 或其他非标准端口

## 🔧 部署

```bash
docker-compose up -d
# 包含资源限制、日志配置等生产环境优化
```

## 🛡️ 防火墙配置

```bash
# 开放必要端口
sudo ufw allow 22     # SSH
sudo ufw allow 80     # 前端
sudo ufw allow 5555   # 后端API
sudo ufw enable
```

## 📊 常用命令

```bash
# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 重启服务
docker-compose restart

# 停止服务
docker-compose down

# 数据备份
./backup.sh
```

## ⚠️ 注意事项

1. **跨域配置**: 前端和后端运行在不同端口，确保CORS配置正确
2. **OAuth2回调**: LinuxDo应用配置中的回调地址要包含端口号
3. **防火墙**: 确保开放80和5555端口
4. **域名解析**: 确保域名正确解析到服务器IP
5. **SSL证书**: 如需HTTPS，建议使用云服务商的负载均衡器或CDN

## 🔍 故障排除

### 前端无法访问后端API
```bash
# 检查后端服务状态
docker-compose logs backend

# 检查网络连通性
curl http://localhost:5555/api/v1/health
```

### OAuth2登录失败
1. 检查LinuxDo应用配置中的回调地址
2. 确认环境变量中的CLIENT_ID和SECRET正确
3. 检查防火墙是否阻止了配置的端口

### 容器启动失败
```bash
# 查看详细错误信息
docker-compose logs

# 检查端口占用
sudo netstat -tlnp | grep :80
sudo netstat -tlnp | grep :5555
```

## 📞 技术支持

如遇问题，请检查：
1. 环境变量配置是否正确
2. 防火墙端口是否开放
3. 域名解析是否正确
4. Docker服务是否正常运行
