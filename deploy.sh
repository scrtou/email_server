#!/bin/bash

# 生产环境部署脚本
# 用途: 自动化部署Email Server到生产环境

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查必要的工具
check_requirements() {
    log_info "检查部署环境..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    log_success "环境检查通过"
}

# 检查环境变量文件
check_env_file() {
    log_info "检查环境变量配置..."
    
    if [ ! -f ".env" ]; then
        log_warning ".env 文件不存在，正在创建示例文件..."
        cp src/backend/.env.example .env
        log_warning "请编辑 .env 文件并设置生产环境配置"
        log_warning "特别注意修改以下配置:"
        log_warning "  - JWT_SECRET: 设置强密钥"
        log_warning "  - LINUXDO_REDIRECT_URI: 修改为您的域名"
        log_warning "  - VUE_APP_API_BASE_URL: 修改为您的API地址"
        read -p "按回车键继续，或按 Ctrl+C 退出编辑 .env 文件..."
    fi
    
    log_success "环境变量文件检查完成"
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."

    mkdir -p data/backend

    log_success "目录创建完成"
}

# 构建和启动服务
deploy_services() {
    log_info "开始部署服务..."
    
    # 停止现有服务
    log_info "停止现有服务..."
    docker-compose down

    # 构建镜像
    log_info "构建Docker镜像..."
    docker-compose build --no-cache

    # 启动服务
    log_info "启动服务..."
    docker-compose up -d
    
    log_success "服务部署完成"
}

# 检查服务状态
check_services() {
    log_info "检查服务状态..."
    
    sleep 10  # 等待服务启动
    
    # 检查容器状态
    if docker-compose ps | grep -q "Up"; then
        log_success "服务启动成功"
        docker-compose ps
    else
        log_error "服务启动失败"
        docker-compose logs
        exit 1
    fi
}

# 显示部署信息
show_deployment_info() {
    log_success "部署完成！"
    echo ""
    log_info "服务访问信息:"
    log_info "  前端: http://localhost:\${FRONTEND_PORT:-80}"
    log_info "  后端API: http://localhost:\${BACKEND_PORT:-5555}/api/v1"
    echo ""
    log_info "常用命令:"
    log_info "  查看日志: docker-compose logs -f"
    log_info "  停止服务: docker-compose down"
    log_info "  重启服务: docker-compose restart"
    echo ""
    log_warning "生产环境部署注意事项:"
    log_warning "  1. 请配置防火墙，开放配置的端口 (默认: 前端80, 后端5555)"
    log_warning "  2. 如需HTTPS，建议在前端容器或负载均衡器配置SSL"
    log_warning "  3. 定期备份数据库文件"
    log_warning "  4. 监控服务运行状态"
    log_warning "  5. 确保前端能正确访问后端API (跨域配置)"
    log_warning "  6. 端口配置可在.env文件中修改 (FRONTEND_PORT, BACKEND_PORT)"
}

# 主函数
main() {
    log_info "开始生产环境部署..."
    
    check_requirements
    check_env_file
    create_directories
    deploy_services
    check_services
    show_deployment_info
    
    log_success "部署脚本执行完成！"
}

# 执行主函数
main "$@"
