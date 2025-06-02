# Email Server 生产环境部署指南

## 📋 部署前准备

### 1. 服务器要求
- **操作系统**: Linux (推荐 Ubuntu 20.04+ 或 CentOS 8+)
- **内存**: 最少 2GB RAM (推荐 4GB+)
- **存储**: 最少 10GB 可用空间
- **网络**: 公网IP地址，开放 80/443 端口

### 2. 必要软件
```bash
# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

## 🚀 快速部署

### 1. 克隆项目
```bash
git clone <your-repository-url>
cd email_server
```

### 2. 配置环境变量
```bash
# 复制环境变量模板
cp .env.production.example .env

# 编辑配置文件
nano .env
```

**重要配置项**:
- `JWT_SECRET`: 设置强密钥 (至少32个字符)
- `LINUXDO_CLIENT_ID/SECRET`: LinuxDo OAuth2 应用信息
- `LINUXDO_REDIRECT_URI`: 修改为您的域名
- `VUE_APP_API_BASE_URL`: 修改为您的API地址

### 3. 修改域名配置
```bash
# 编辑 Nginx 配置
nano nginx/conf.d/default.conf
# 将 yourdomain.com 替换为您的实际域名
```

### 4. 执行部署
```bash
# 给脚本执行权限
chmod +x deploy.sh

# 执行部署
./deploy.sh
```

## 🔧 手动部署步骤

### 1. 创建必要目录
```bash
mkdir -p data/backend logs/nginx ssl
```

### 2. 构建和启动服务
```bash
# 开发环境
docker-compose up -d

# 生产环境
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### 3. 检查服务状态
```bash
docker-compose ps
docker-compose logs -f
```

## 🔒 HTTPS 配置

### 1. 获取 SSL 证书
```bash
# 使用 Let's Encrypt (推荐)
sudo apt install certbot
sudo certbot certonly --standalone -d yourdomain.com
```

### 2. 配置证书
```bash
# 复制证书到项目目录
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem ./ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem ./ssl/key.pem
sudo chown $USER:$USER ./ssl/*.pem
```

### 3. 启用 HTTPS
编辑 `nginx/conf.d/default.conf`，取消注释 HTTPS 服务器配置部分。

## 📊 监控和维护

### 1. 查看日志
```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f nginx
```

### 2. 数据备份
```bash
# 执行备份
chmod +x backup.sh
./backup.sh

# 备份文件位置
ls -la backups/
```

### 3. 服务管理
```bash
# 重启服务
docker-compose restart

# 停止服务
docker-compose down

# 更新服务
docker-compose pull
docker-compose up -d
```

## 🛡️ 安全配置

### 1. 防火墙设置
```bash
# Ubuntu/Debian
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP
sudo ufw allow 443   # HTTPS
sudo ufw enable

# CentOS/RHEL
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

### 2. 定期更新
```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 更新 Docker 镜像
docker-compose pull
docker-compose up -d
```

## 🔍 故障排除

### 1. 常见问题

**服务无法启动**:
```bash
# 检查端口占用
sudo netstat -tlnp | grep :80
sudo netstat -tlnp | grep :443

# 检查 Docker 状态
sudo systemctl status docker
```

**数据库连接失败**:
```bash
# 检查数据目录权限
ls -la data/backend/
sudo chown -R 1001:1001 data/backend/
```

**OAuth2 回调失败**:
- 检查 LinuxDo 应用配置中的回调地址
- 确认域名解析正确
- 检查防火墙设置

### 2. 性能优化

**内存优化**:
- 调整 Docker 容器资源限制
- 监控内存使用情况

**数据库优化**:
- 定期清理日志
- 优化查询语句
- 考虑使用 PostgreSQL 替代 SQLite

## 📞 技术支持

如果遇到部署问题，请：
1. 查看日志文件
2. 检查配置文件
3. 参考故障排除部分
4. 提交 Issue 到项目仓库
