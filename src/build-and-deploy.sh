#!/bin/bash
set -e # 如果命令失败，则立即退出

# 构建和部署脚本
#
# 用途:
# 1. 构建前端 Vue.js 应用
# 2. 构建后端 Go 应用 (生产环境二进制文件)
# 3. 提供启动后端应用的指令
#
# 前提:
# - 确保已安装 Node.js, npm, 和 Go
# - 确保后端代码中已处理静态文件服务 (如果选择Go服务前端静态文件) 或已配置Nginx

# --- 配置变量 (可根据实际情况修改) ---
BACKEND_OUTPUT_DIR="../../dist_backend" # 后端编译输出目录 (相对于 src/backend)
BACKEND_APP_NAME="email_server_app"    # 后端编译后的应用名称
# API_BASE_URL (示例, 用于前端构建)
# 如果需要通过脚本设置前端环境变量，可以在下面取消注释并设置
# export VUE_APP_API_BASE_URL="https://api.example.com/v1"

echo "----------------------------------------"
echo "INFO: 开始构建前端项目..."
echo "----------------------------------------"

# 进入前端目录
cd src/frontend

echo "INFO: 安装前端依赖..."
npm install

echo "INFO: 开始构建前端生产版本..."
echo "提示: 前端构建会使用 .env.production 文件中的环境变量 (例如 VUE_APP_API_BASE_URL)。"
echo "      如果需要覆盖，可以在执行此脚本前设置相应的环境变量，例如："
echo "      export VUE_APP_API_BASE_URL=\"your_api_url\""
npm run build

echo "INFO: 前端构建完成。静态文件位于 src/frontend/dist/"
echo "----------------------------------------"

# 返回项目根目录 (假设脚本在项目根目录的 src/ 下执行)
cd ../.. # 从 src/frontend 返回到项目根目录

echo "----------------------------------------"
echo "INFO: 开始构建后端项目..."
echo "----------------------------------------"

# 进入后端目录
cd src/backend

echo "INFO: 准备编译 Go 后端应用为生产环境二进制文件..."
echo "提示: 后端应用在运行时会读取操作系统环境变量。"
echo "      例如，数据库连接信息、JWT密钥等应通过环境变量配置。"
echo "      示例: export DATABASE_URL=\"./data/app.db\" JWT_SECRET=\"your_secret_key\""

# 创建后端输出目录 (如果不存在)
mkdir -p "${BACKEND_OUTPUT_DIR}"

echo "INFO: 编译 Go 应用..."
# 使用 -ldflags 来减小二进制文件大小并移除调试信息
# CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "${BACKEND_OUTPUT_DIR}/${BACKEND_APP_NAME}" main.go
# 默认情况下，为当前操作系统编译
go build -ldflags="-s -w" -o "${BACKEND_OUTPUT_DIR}/${BACKEND_APP_NAME}" main.go

echo "INFO: 后端构建完成。二进制文件位于 ${BACKEND_OUTPUT_DIR}/${BACKEND_APP_NAME}"
echo "----------------------------------------"

echo "----------------------------------------"
echo "部署和运行说明:"
echo "----------------------------------------"
echo ""
echo "1. 前端静态文件服务:"
echo "   前端构建后的静态文件位于 src/frontend/dist/。"
echo "   你可以选择以下任一方式提供服务:"
echo "   a) Go 后端服务: "
echo "      - 确保你的 Go 后端应用 ([`src/backend/main.go`](src/backend/main.go:1)) 已配置为服务这些静态文件。"
echo "      - 例如，使用 Gin 的 StaticFS 或 net/http 的 FileServer 指向 'src/frontend/dist' 目录。"
echo "      - 如果后端服务静态文件，则前端的 API 请求路径应配置为相对路径或与后端服务同源。"
echo "   b) Nginx 或其他 Web 服务器:"
echo "      - 配置 Nginx 将 'src/frontend/dist/' 作为静态文件根目录。"
echo "      - 同时配置 Nginx 将 API 请求 (例如 /api/v1/*) 反向代理到后端 Go 应用运行的端口。"
echo ""
echo "2. 后端环境变量:"
echo "   在运行后端应用前，请确保已设置必要的环境变量，例如:"
echo "   export DATABASE_PATH=\"./data/app.db\"       # SQLite数据库文件路径"
echo "   export JWT_SECRET_KEY=\"your_very_secret_key\" # JWT 密钥"
echo "   # 注意：应用固定监听5555端口，如需修改请在代码中调整"
echo "   # 根据你的应用实际需要配置更多环境变量"
echo ""
echo "3. 运行后端应用:"
echo "   你可以直接运行编译后的二进制文件:"
echo "   cd src/backend && ./${BACKEND_OUTPUT_DIR}/${BACKEND_APP_NAME}"
echo "   或者从项目根目录运行:"
echo "   ./src/backend/${BACKEND_OUTPUT_DIR}/${BACKEND_APP_NAME}"
echo ""
echo "----------------------------------------"
echo "构建和部署脚本执行完毕。"
echo "----------------------------------------"