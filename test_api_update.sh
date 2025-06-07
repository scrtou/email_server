#!/bin/bash

# API测试脚本 - 测试更新平台注册信息的新逻辑
# 这个脚本用于验证后端修改后的邮箱地址处理逻辑

echo "🧪 开始测试更新平台注册信息API..."

# 配置
BASE_URL="http://localhost:8080"
API_ENDPOINT="/api/platform-registrations"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
    local test_name="$1"
    local registration_id="$2"
    local payload="$3"
    local expected_status="$4"
    
    echo -e "\n${YELLOW}🧪 测试: $test_name${NC}"
    echo "📤 请求数据: $payload"
    
    # 发送请求
    response=$(curl -s -w "\n%{http_code}" \
        -X PUT \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer YOUR_TOKEN_HERE" \
        -d "$payload" \
        "$BASE_URL$API_ENDPOINT/$registration_id")
    
    # 分离响应体和状态码
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    echo "📥 响应状态: $http_code"
    echo "📥 响应内容: $response_body"
    
    # 检查状态码
    if [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}✅ 状态码正确${NC}"
    else
        echo -e "${RED}❌ 状态码错误，期望: $expected_status，实际: $http_code${NC}"
    fi
}

echo "📋 注意：这个脚本需要："
echo "1. 后端服务运行在 $BASE_URL"
echo "2. 有效的认证令牌"
echo "3. 存在的平台注册记录ID"
echo ""
echo "请根据实际情况修改脚本中的配置。"
echo ""

# 测试用例1：添加新邮箱地址
test_api "添加新邮箱地址" "1" '{
    "email_address": "new@example.com",
    "login_username": "developer",
    "login_password": "newpassword123",
    "notes": "测试账号",
    "phone_number": "+86 139****5678"
}' "200"

# 测试用例2：使用现有邮箱地址
test_api "使用现有邮箱地址" "1" '{
    "email_address": "existing@example.com",
    "login_username": "developer",
    "notes": "测试账号",
    "phone_number": "+86 139****5678"
}' "200"

# 测试用例3：清空邮箱地址
test_api "清空邮箱地址" "1" '{
    "email_address": "",
    "login_username": "developer",
    "notes": "测试账号",
    "phone_number": "+86 139****5678"
}' "200"

# 测试用例4：无效邮箱格式
test_api "无效邮箱格式" "1" '{
    "email_address": "invalid-email",
    "login_username": "developer",
    "notes": "测试账号",
    "phone_number": "+86 139****5678"
}' "400"

# 测试用例5：密码太短
test_api "密码太短" "1" '{
    "email_address": "test@example.com",
    "login_username": "developer",
    "login_password": "123",
    "notes": "测试账号",
    "phone_number": "+86 139****5678"
}' "400"

echo -e "\n${YELLOW}📊 测试完成${NC}"
echo ""
echo "💡 使用说明："
echo "1. 将 YOUR_TOKEN_HERE 替换为有效的JWT令牌"
echo "2. 将平台注册ID替换为实际存在的ID"
echo "3. 确保后端服务正在运行"
echo "4. 根据需要调整BASE_URL"
echo ""
echo "🔧 手动测试示例："
echo "curl -X PUT \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -H \"Authorization: Bearer YOUR_TOKEN\" \\"
echo "  -d '{\"email_address\":\"test@example.com\",\"login_username\":\"user\"}' \\"
echo "  \"$BASE_URL$API_ENDPOINT/1\""
