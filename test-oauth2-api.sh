#!/bin/bash

# OAuth2 API测试脚本
echo "🧪 测试LinuxDo OAuth2 API"

BASE_URL="http://localhost:5555/api/v1"

# 检查后端服务是否运行
echo "📡 检查后端服务状态..."
if ! curl -s "$BASE_URL/health" > /dev/null; then
    echo "❌ 后端服务未运行，请先启动后端服务"
    echo "   cd src/backend && go run main.go"
    exit 1
fi

echo "✅ 后端服务正常运行"

# 测试OAuth2登录URL生成
echo ""
echo "🔗 测试OAuth2登录URL生成..."
RESPONSE=$(curl -s "$BASE_URL/auth/oauth2/linuxdo/login")

if echo "$RESPONSE" | grep -q "auth_url"; then
    echo "✅ OAuth2登录URL生成成功"
    AUTH_URL=$(echo "$RESPONSE" | grep -o '"auth_url":"[^"]*"' | cut -d'"' -f4)
    echo "   授权URL: $AUTH_URL"
    
    # 检查URL是否包含必要参数
    if echo "$AUTH_URL" | grep -q "client_id" && echo "$AUTH_URL" | grep -q "redirect_uri"; then
        echo "✅ 授权URL包含必要参数"
    else
        echo "⚠️  授权URL可能缺少必要参数"
    fi
else
    echo "❌ OAuth2登录URL生成失败"
    echo "   响应: $RESPONSE"
fi

# 测试健康检查
echo ""
echo "💓 测试健康检查..."
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
if echo "$HEALTH_RESPONSE" | grep -q "ok"; then
    echo "✅ 健康检查通过"
else
    echo "❌ 健康检查失败"
fi

# 检查环境变量配置
echo ""
echo "⚙️  检查环境变量配置..."
if [ -f "src/backend/.env" ]; then
    echo "✅ 找到.env配置文件"
    
    if grep -q "LINUXDO_CLIENT_ID=" src/backend/.env && [ "$(grep "LINUXDO_CLIENT_ID=" src/backend/.env | cut -d'=' -f2)" != "" ]; then
        echo "✅ LINUXDO_CLIENT_ID已配置"
    else
        echo "⚠️  LINUXDO_CLIENT_ID未配置或为空"
    fi
    
    if grep -q "LINUXDO_CLIENT_SECRET=" src/backend/.env && [ "$(grep "LINUXDO_CLIENT_SECRET=" src/backend/.env | cut -d'=' -f2)" != "" ]; then
        echo "✅ LINUXDO_CLIENT_SECRET已配置"
    else
        echo "⚠️  LINUXDO_CLIENT_SECRET未配置或为空"
    fi
else
    echo "❌ 未找到.env配置文件"
    echo "   请复制 src/backend/.env.example 为 src/backend/.env 并配置OAuth2参数"
fi

echo ""
echo "📋 测试总结:"
echo "   - 后端服务: ✅ 运行中"
echo "   - OAuth2 API: $(echo "$RESPONSE" | grep -q "auth_url" && echo "✅ 正常" || echo "❌ 异常")"
echo "   - 配置文件: $([ -f "src/backend/.env" ] && echo "✅ 存在" || echo "❌ 缺失")"
echo ""
echo "🚀 如果所有检查都通过，您可以开始测试OAuth2登录功能了！"
echo "   1. 访问前端: http://localhost:8080"
echo "   2. 点击'使用 Linux.do 登录'按钮"
echo "   3. 完成LinuxDo授权流程"
