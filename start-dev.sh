#!/bin/bash

# 开发环境启动脚本
echo "🚀 启动邮箱账号管理系统开发环境"

# 检查是否存在 .env 文件
if [ ! -f "src/backend/.env" ]; then
    echo "⚠️  未找到 .env 文件，正在创建..."
    cp src/backend/.env.example src/backend/.env
    echo "✅ 已创建 .env 文件，请编辑 src/backend/.env 配置LinuxDo OAuth2参数"
    echo "📖 详细配置说明请查看 LINUXDO_OAUTH2_SETUP.md"
fi

# 启动后端
echo "🔧 启动后端服务..."
cd src/backend
go run main.go &
BACKEND_PID=$!
cd ../..

# 等待后端启动
echo "⏳ 等待后端服务启动..."
sleep 3

# 启动前端
echo "🎨 启动前端服务..."
cd src/frontend
npm run serve &
FRONTEND_PID=$!
cd ../..

echo "✅ 服务启动完成！"
echo "📱 前端地址: http://localhost:8080"
echo "🔧 后端地址: http://localhost:5555"
echo "📚 OAuth2配置文档: LINUXDO_OAUTH2_SETUP.md"
echo ""
echo "按 Ctrl+C 停止所有服务"

# 等待用户中断
trap "echo '🛑 正在停止服务...'; kill $BACKEND_PID $FRONTEND_PID; exit" INT
wait
