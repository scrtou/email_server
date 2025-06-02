# Email Server - 邮箱账户管理系统


一个现代化的邮箱账户管理系统，帮助用户统一管理多个邮箱账户、平台注册信息和服务订阅。支持LinuxDo OAuth2登录，提供完整的Web界面和RESTful API。

## ✨ 主要功能


### 📧 邮箱账户管理
- **邮箱账户**：添加、编辑、删除邮箱账户信息
- **密码加密**：安全存储邮箱密码（可选）
- **服务商识别**：自动识别邮箱服务商（Gmail、Outlook等）
- **备注管理**：为每个邮箱账户添加备注信息

### 🌐 平台注册管理
- **平台信息**：管理各种网站和服务平台
- **注册记录**：记录邮箱在各平台的注册信息
- **登录凭据**：安全存储平台登录用户名和密码
- **关联管理**：邮箱账户与平台注册的关联关系

### 💰 服务订阅管理
- **订阅跟踪**：管理各种付费服务订阅
- **费用管理**：记录订阅费用和计费周期
- **续费提醒**：自动提醒即将到期的订阅
- **支付方式**：记录支付方式和相关备注

### 📊 数据统计
- **仪表板**：直观的数据统计和图表展示
- **搜索功能**：全局搜索邮箱、平台和订阅信息
- **数据导出**：支持数据备份和导出

### 部署
- **容器化**：Docker + Docker Compose
- **进程管理**：支持systemd服务
- **数据持久化**：Docker Volume

## 🚀 快速开始

### 环境要求
- Docker 20.0+
- Docker Compose 2.0+
- Git

### 1. 克隆项目
```bash
git clone https://github.com/yourusername/email_server.git
cd email_server
```

### 2. 配置环境变量
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置文件
nano .env
```

**重要配置项**：
```env
# JWT密钥（生产环境必须修改）
JWT_SECRET=your-production-super-secret-jwt-key-at-least-32-characters-long

# LinuxDo OAuth2配置（可选）
LINUXDO_CLIENT_ID=your_client_id
LINUXDO_CLIENT_SECRET=your_client_secret
LINUXDO_REDIRECT_URI=https://yourdomain.com/api/v1/auth/oauth2/linuxdo/callback

# 后端API地址配置
VUE_APP_API_BASE_URL=https://yourdomain.com/api/v1
FRONTEND_BASE_URL=https://yourdomain.com
```

### 3. 启动服务
```bash
# 开发环境
docker-compose up -d


### 4. 访问应用
- **前端界面**：http://localhost:80
- **后端API**：http://localhost:5555
- **健康检查**：http://localhost:5555/api/v1/health

### 5. 默认管理员账户
- **用户名**：admin
- **密码**：password

> ⚠️ **安全提醒**：首次登录后请立即修改默认密码！


### 构建部署

#### 使用构建脚本
```bash
cd src
chmod +x build-and-deploy.sh
./build-and-deploy.sh
```

#### 手动构建
```bash
# 构建前端
cd src/frontend
npm run build

# 构建后端
cd ../backend
go build -o email_server_app main.go


## 📊 监控和维护

### 健康检查
```bash
# 检查服务状态
curl http://localhost:5555/api/v1/health

# 检查OAuth2状态统计
curl http://localhost:5555/api/v1/auth/oauth2/stats
```

### 数据备份
```bash
# 执行备份
chmod +x backup.sh
./backup.sh

# 查看备份文件
ls -la backups/
```

### 日志管理
```bash
# 查看应用日志
docker-compose logs -f backend
docker-compose logs -f frontend

## 📝 更新日志

### v1.0.0 (2025-06-02)
- ✨ 初始版本发布
- 🔐 用户认证系统（JWT + OAuth2）
- 📧 邮箱账户管理
- 🌐 平台注册管理
- 💰 服务订阅管理
- 📊 数据统计仪表板
- 🐳 Docker容器化部署



## 📄 许可证

本项目采用 [MIT License](LICENSE) 许可证。

## 🙏 致谢

感谢以下开源项目：

- [Gin](https://github.com/gin-gonic/gin) - Go Web框架
- [Vue.js](https://vuejs.org/) - 渐进式JavaScript框架
- [Element Plus](https://element-plus.org/) - Vue 3组件库
- [GORM](https://gorm.io/) - Go ORM库
- [Pinia](https://pinia.vuejs.org/) - Vue状态管理
- [ECharts](https://echarts.apache.org/) - 数据可视化库

---

⭐ 如果这个项目对您有帮助，请给我们一个Star！
