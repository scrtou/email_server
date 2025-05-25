#!/bin/bash

echo "开始构建前端项目..."

# 进入前端目录
cd frontend

# 安装依赖
npm install

# 构建生产版本
npm run build

echo "前端构建完成"

# 返回项目根目录
cd ..
cd backend
# 启动Go服务器
echo "启动Go服务器..."
go run main.go