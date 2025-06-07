#!/bin/bash

# 前后端集成测试脚本
# 测试修复后的邮箱地址字段处理

echo "🧪 开始前后端集成测试..."

# 配置
BASE_URL="http://localhost:8080"
API_ENDPOINT="/api/platform-registrations"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
    local test_name="$1"
    local method="$2"
    local endpoint="$3"
    local payload="$4"
    local expected_status="$5"
    
    echo -e "\n${YELLOW}🧪 测试: $test_name${NC}"
    echo -e "${BLUE}📤 请求: $method $endpoint${NC}"
    echo -e "${BLUE}📤 数据: $payload${NC}"
    
    # 发送请求
    if [ "$method" = "POST" ]; then
        response=$(curl -s -w "\n%{http_code}" \
            -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer YOUR_TOKEN_HERE" \
            -d "$payload" \
            "$BASE_URL$endpoint")
    elif [ "$method" = "PUT" ]; then
        response=$(curl -s -w "\n%{http_code}" \
            -X PUT \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer YOUR_TOKEN_HERE" \
            -d "$payload" \
            "$BASE_URL$endpoint")
    fi
    
    # 分离响应体和状态码
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    echo -e "${BLUE}📥 状态码: $http_code${NC}"
    echo -e "${BLUE}📥 响应: $response_body${NC}"
    
    # 检查状态码
    if [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}✅ 测试通过${NC}"
    else
        echo -e "${RED}❌ 测试失败，期望状态码: $expected_status，实际: $http_code${NC}"
    fi
}

echo -e "${YELLOW}📋 注意事项：${NC}"
echo "1. 确保后端服务运行在 $BASE_URL"
echo "2. 将 YOUR_TOKEN_HERE 替换为有效的JWT令牌"
echo "3. 根据实际情况调整测试数据"
echo ""

# 测试用例1：创建平台注册（按名称API）
test_api "创建平台注册 - 新邮箱地址" "POST" "/api/platform-registrations/by-name" '{
    "email_address": "test@example.com",
    "platform_name": "TestPlatform",
    "login_username": "testuser",
    "login_password": "password123",
    "notes": "测试账号",
    "phone_number": "+86 139****5678"
}' "201"

# 测试用例2：创建平台注册（现有邮箱地址）
test_api "创建平台注册 - 现有邮箱地址" "POST" "/api/platform-registrations/by-name" '{
    "email_address": "existing@example.com",
    "platform_name": "GitHub",
    "login_username": "existinguser",
    "login_password": "password123",
    "notes": "现有邮箱测试",
    "phone_number": "+86 139****5678"
}' "201"

# 测试用例3：更新平台注册（修改邮箱地址）
test_api "更新平台注册 - 修改邮箱地址" "PUT" "/api/platform-registrations/1" '{
    "email_address": "updated@example.com",
    "login_username": "updateduser",
    "notes": "更新后的账号",
    "phone_number": "+86 139****5678"
}' "200"

# 测试用例4：更新平台注册（清空邮箱地址）
test_api "更新平台注册 - 清空邮箱地址" "PUT" "/api/platform-registrations/1" '{
    "email_address": "",
    "login_username": "noemailuser",
    "notes": "无邮箱账号",
    "phone_number": "+86 139****5678"
}' "200"

# 测试用例5：无效邮箱格式
test_api "创建平台注册 - 无效邮箱格式" "POST" "/api/platform-registrations/by-name" '{
    "email_address": "invalid-email",
    "platform_name": "TestPlatform",
    "login_username": "testuser",
    "login_password": "password123",
    "notes": "无效邮箱测试"
}' "400"

# 测试用例6：密码太短
test_api "创建平台注册 - 密码太短" "POST" "/api/platform-registrations/by-name" '{
    "email_address": "test@example.com",
    "platform_name": "TestPlatform",
    "login_username": "testuser",
    "login_password": "123",
    "notes": "密码太短测试"
}' "400"

# 测试用例7：用户名和邮箱都为空
test_api "创建平台注册 - 用户名和邮箱都为空" "POST" "/api/platform-registrations/by-name" '{
    "email_address": "",
    "platform_name": "TestPlatform",
    "login_username": "",
    "login_password": "password123",
    "notes": "空字段测试"
}' "400"

echo -e "\n${YELLOW}📊 集成测试完成${NC}"
echo ""
echo -e "${BLUE}💡 使用说明：${NC}"
echo "1. 将 YOUR_TOKEN_HERE 替换为有效的JWT令牌"
echo "2. 根据实际数据调整平台注册ID"
echo "3. 确保后端服务正在运行"
echo "4. 根据需要调整BASE_URL"
echo ""
echo -e "${BLUE}🔧 手动测试示例：${NC}"
echo "# 创建平台注册"
echo "curl -X POST \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -H \"Authorization: Bearer YOUR_TOKEN\" \\"
echo "  -d '{\"email_address\":\"test@example.com\",\"platform_name\":\"GitHub\",\"login_username\":\"user\"}' \\"
echo "  \"$BASE_URL/api/platform-registrations/by-name\""
echo ""
echo "# 更新平台注册"
echo "curl -X PUT \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -H \"Authorization: Bearer YOUR_TOKEN\" \\"
echo "  -d '{\"email_address\":\"updated@example.com\",\"login_username\":\"user\"}' \\"
echo "  \"$BASE_URL/api/platform-registrations/1\""
echo ""
echo -e "${GREEN}✨ 修复总结：${NC}"
echo "1. 前端统一使用邮箱地址字段"
echo "2. 后端支持邮箱地址的创建和更新"
echo "3. 消除了类型不匹配的错误"
echo "4. 提升了系统的稳定性和用户体验"
