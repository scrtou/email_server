# 贡献指南

感谢您对Email Server项目的关注！我们欢迎所有形式的贡献，包括但不限于：

- 🐛 Bug报告
- ✨ 新功能建议
- 📝 文档改进
- 🔧 代码贡献
- 🌐 翻译工作

## 🚀 快速开始

### 1. Fork项目
在GitHub上Fork本项目到您的账户。

### 2. 克隆项目
```bash
git clone https://github.com/yourusername/email_server.git
cd email_server
```

### 3. 设置开发环境
```bash
# 后端开发环境
cd src/backend
go mod download

# 前端开发环境
cd ../frontend
npm install
```

### 4. 创建功能分支
```bash
git checkout -b feature/your-feature-name
# 或
git checkout -b bugfix/issue-number
```

## 📋 开发规范

### 代码风格

#### Go代码规范
- 使用`gofmt`格式化代码
- 遵循Go官方代码规范
- 使用有意义的变量和函数名
- 添加必要的注释

```bash
# 格式化代码
go fmt ./...

# 检查代码质量
go vet ./...

# 运行测试
go test ./...
```

#### Vue.js代码规范
- 遵循Vue.js官方风格指南
- 使用ESLint进行代码检查
- 组件名使用PascalCase
- 文件名使用kebab-case

```bash
# 代码检查和修复
npm run lint

# 运行测试
npm run test
```

### 提交信息规范

使用[Conventional Commits](https://conventionalcommits.org/)格式：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### 类型说明
- `feat`: 新功能
- `fix`: Bug修复
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

#### 示例
```
feat(auth): add OAuth2 login support

Add LinuxDo OAuth2 authentication to allow users to login
using their LinuxDo accounts.

Closes #123
```

## 🐛 Bug报告

在提交Bug报告前，请：

1. 检查是否已有相关Issue
2. 确保使用最新版本
3. 提供详细的复现步骤

### Bug报告模板
```markdown
## Bug描述
简要描述遇到的问题

## 复现步骤
1. 进入...
2. 点击...
3. 看到错误...

## 预期行为
描述您期望发生的情况

## 实际行为
描述实际发生的情况

## 环境信息
- 操作系统: [e.g. Ubuntu 20.04]
- 浏览器: [e.g. Chrome 91]
- 项目版本: [e.g. v1.0.0]

## 附加信息
添加任何其他有助于解决问题的信息
```

## ✨ 功能请求

### 功能请求模板
```markdown
## 功能描述
清晰简洁地描述您想要的功能

## 问题背景
描述这个功能要解决的问题

## 解决方案
描述您希望的解决方案

## 替代方案
描述您考虑过的其他解决方案

## 附加信息
添加任何其他相关信息或截图
```

## 🔧 代码贡献流程

### 1. 开发前准备
- 确保您的Fork是最新的
- 创建新的功能分支
- 了解相关的代码结构

### 2. 开发过程
- 编写清晰的代码
- 添加必要的测试
- 更新相关文档
- 确保代码通过所有检查

### 3. 提交Pull Request
- 填写详细的PR描述
- 关联相关的Issue
- 确保CI检查通过
- 响应代码审查意见

### Pull Request模板
```markdown
## 变更描述
简要描述这个PR的变更内容

## 变更类型
- [ ] Bug修复
- [ ] 新功能
- [ ] 文档更新
- [ ] 代码重构
- [ ] 性能优化

## 测试
- [ ] 已添加测试用例
- [ ] 所有测试通过
- [ ] 手动测试通过

## 检查清单
- [ ] 代码遵循项目规范
- [ ] 已更新相关文档
- [ ] 提交信息符合规范
- [ ] 已自测功能正常

## 关联Issue
Closes #(issue number)

## 截图
如果适用，请添加截图说明变更
```

## 🧪 测试指南

### 后端测试
```bash
cd src/backend

# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./handlers

# 生成测试覆盖率报告
go test -cover ./...
```

### 前端测试
```bash
cd src/frontend

# 运行单元测试
npm run test:unit

# 运行端到端测试
npm run test:e2e
```

## 📚 文档贡献

### 文档类型
- API文档
- 用户指南
- 开发文档
- 部署文档

### 文档规范
- 使用Markdown格式
- 保持简洁明了
- 提供实际示例
- 及时更新过时内容

## 🌐 国际化

我们欢迎翻译贡献：

1. 复制`src/frontend/src/locales/zh-CN.js`
2. 翻译为目标语言
3. 更新语言配置
4. 提交PR

## 🎯 项目路线图

查看[GitHub Projects](https://github.com/yourusername/email_server/projects)了解项目规划。

## 💬 社区交流

- GitHub Discussions: 技术讨论和问答
- GitHub Issues: Bug报告和功能请求
- Email: 私密问题和安全报告

## 🏆 贡献者

感谢所有为项目做出贡献的开发者！

## 📄 许可证

通过贡献代码，您同意您的贡献将在[MIT License](LICENSE)下授权。
