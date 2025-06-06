# Email Server 账号管理器 - 浏览器插件

这是一个类似 Bitwarden 的浏览器插件，能够自动检测用户在网站上的注册/登录信息，并将其保存到 Email Server 的平台注册表中。

## 功能特性

### 🔍 自动检测
- 自动识别网页上的登录和注册表单
- 智能提取用户名、密码、邮箱等信息
- 基于域名自动识别平台名称

### 💾 数据管理
- 与 Email Server 后端 API 无缝集成
- 支持手动添加和编辑账号信息
- 查看已保存的账号列表

### 🛡️ 安全性
- 使用 JWT Token 进行身份验证
- 密码在传输前进行加密
- 支持自定义服务器地址

### ⚙️ 个性化设置
- 可配置自动检测行为
- 支持排除特定网站
- 灵活的通知设置

## 安装方法

### 开发模式安装

1. 打开 Chrome 浏览器，进入扩展程序管理页面：
   ```
   chrome://extensions/
   ```

2. 开启"开发者模式"（右上角开关）

3. 点击"加载已解压的扩展程序"

4. 选择 `browser-extension` 文件夹

5. 插件安装完成，会在工具栏显示图标

### 生产环境安装

1. 将整个 `browser-extension` 文件夹打包为 `.zip` 文件
2. 上传到 Chrome Web Store 或其他浏览器扩展商店
3. 用户可直接从商店安装

## 使用指南

### 初始设置

1. 点击插件图标，进入弹窗界面
2. 点击"设置"按钮，配置服务器地址
3. 输入 Email Server 的后端地址（如：`http://localhost:8080`）
4. 可选：输入用户名和密码以启用自动登录

### 自动检测使用

1. 访问任何网站的登录或注册页面
2. 填写表单信息（用户名、密码、邮箱等）
3. 提交表单时，插件会自动检测并显示保存提示
4. 点击"保存到服务器"确认保存

### 手动添加账号

1. 点击插件图标打开弹窗
2. 切换到"手动添加"标签页
3. 填写平台信息（会自动填充当前网站）
4. 点击"添加账号"保存

### 查看账号列表

1. 在弹窗中切换到"账号列表"标签页
2. 查看所有已保存的账号信息
3. 点击"刷新列表"获取最新数据

## 技术架构

### 文件结构
```
browser-extension/
├── manifest.json          # 插件配置文件
├── background.js          # 后台脚本，处理 API 通信
├── content.js            # 内容脚本，检测表单
├── popup.html            # 弹窗界面
├── popup.js              # 弹窗逻辑
├── options.html          # 设置页面
├── options.js            # 设置逻辑
├── icons/                # 图标文件
├── test-page.html        # 测试页面
├── package.sh            # 打包脚本
└── docs/                 # 文档目录
```

### 核心组件

#### Background Script (background.js)
- 处理与 Email Server API 的通信
- 管理用户认证状态
- 存储和检索配置信息

#### Content Script (content.js)
- 检测网页上的表单元素
- 提取用户输入的账号信息
- 显示保存提示界面

#### Popup Interface (popup.html/js)
- 用户交互界面
- 登录功能
- 手动添加账号
- 查看账号列表

#### Options Page (options.html/js)
- 插件设置配置
- 服务器连接测试
- 高级功能开关

### API 集成

插件与 Email Server 的以下 API 端点集成：

- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/platform-registrations/by-name` - 创建平台注册信息
- `GET /api/v1/platform-registrations` - 获取注册信息列表
- `GET /api/v1/health` - 健康检查

## 开发说明

### 本地开发

1. 修改代码后，在 Chrome 扩展程序页面点击刷新按钮
2. 使用浏览器开发者工具调试：
   - 右键插件图标 → 检查弹出内容（调试 popup）
   - 在网页上按 F12 → Console（查看 content script 日志）
   - 扩展程序页面 → 背景页面（调试 background script）

### 调试技巧

1. **Content Script 调试**：
   ```javascript
   console.log('Form detected:', formData);
   ```

2. **Background Script 调试**：
   ```javascript
   console.log('API response:', response);
   ```

3. **Popup 调试**：
   右键插件图标 → "检查弹出内容"

### 常见问题

1. **CORS 错误**：确保 Email Server 后端配置了正确的 CORS 策略
2. **权限问题**：检查 manifest.json 中的权限配置
3. **API 调用失败**：验证服务器地址和网络连接

## 安全考虑

1. **密码存储**：密码使用 Chrome 的安全存储 API
2. **HTTPS 支持**：建议生产环境使用 HTTPS
3. **权限最小化**：只请求必要的浏览器权限
4. **数据验证**：对所有用户输入进行验证

## 文档索引

### 📚 用户文档
- **[README.md](README.md)** - 项目概述和功能介绍
- **[INSTALL.md](INSTALL.md)** - 详细的安装和使用指南
- **[test-page.html](test-page.html)** - 功能测试页面

### 🔧 技术文档
- **[DESIGN_IMPLEMENTATION.md](DESIGN_IMPLEMENTATION.md)** - 详细的设计实现说明
- **[API_INTEGRATION.md](API_INTEGRATION.md)** - Email Server API 集成文档
- **[DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)** - 开发者指南和最佳实践
- **[PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)** - 项目技术总结

### 🛠️ 开发工具
- **[package.sh](package.sh)** - 自动化打包脚本
- **[icons/README.md](icons/README.md)** - 图标文件说明

## 快速导航

| 我想要... | 查看文档 |
|-----------|----------|
| 了解项目功能 | [README.md](README.md) |
| 安装和使用插件 | [INSTALL.md](INSTALL.md) |
| 了解技术架构 | [DESIGN_IMPLEMENTATION.md](DESIGN_IMPLEMENTATION.md) |
| 集成 API 接口 | [API_INTEGRATION.md](API_INTEGRATION.md) |
| 参与开发 | [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) |
| 测试功能 | [test-page.html](test-page.html) |
| 打包发布 | [package.sh](package.sh) |

## 更新日志

### v1.0.0
- 初始版本发布
- 支持自动表单检测
- 集成 Email Server API
- 提供完整的用户界面

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 发起 Pull Request

## 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。
