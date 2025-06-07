#!/bin/bash

# 数据备份脚本
# 用途: 备份Email Server的数据库和重要配置文件

set -e

# 配置
BACKUP_DIR="./backups"
DATE=$(date +"%Y%m%d_%H%M%S")
BACKUP_NAME="email_server_backup_${DATE}"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_NAME}"

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 创建备份目录
create_backup_dir() {
    log_info "创建备份目录..."
    mkdir -p "${BACKUP_PATH}"
    log_success "备份目录创建完成: ${BACKUP_PATH}"
}

# 备份数据库
backup_database() {
    log_info "备份数据库..."
    
    if [ -f "./data/backend/database.db" ]; then
        cp "./data/backend/database.db" "${BACKUP_PATH}/database.db"
        log_success "数据库备份完成"
    else
        log_warning "数据库文件不存在，跳过备份"
    fi
}

# 备份配置文件
backup_configs() {
    log_info "备份配置文件..."
    
    # 备份环境变量文件
    if [ -f ".env" ]; then
        cp ".env" "${BACKUP_PATH}/.env"
        log_success "环境变量文件备份完成"
    fi
    
    # 备份Docker配置
    cp "docker-compose.yml" "${BACKUP_PATH}/"
    if [ -f "docker-compose.prod.yml" ]; then
        cp "docker-compose.prod.yml" "${BACKUP_PATH}/"
    fi
    
    # 备份其他配置文件
    if [ -f "deploy.sh" ]; then
        cp "deploy.sh" "${BACKUP_PATH}/"
    fi
    if [ -f "backup.sh" ]; then
        cp "backup.sh" "${BACKUP_PATH}/"
    fi
    
    log_success "配置文件备份完成"
}

# 创建备份信息文件
create_backup_info() {
    log_info "创建备份信息文件..."
    
    cat > "${BACKUP_PATH}/backup_info.txt" << EOF
Email Server 备份信息
==================

备份时间: $(date)
备份版本: ${DATE}
备份内容:
- 数据库文件 (database.db)
- 环境变量配置 (.env)
- Docker配置文件
- 部署脚本

恢复说明:
1. 停止当前服务: docker-compose down
2. 恢复数据库: cp database.db ./data/backend/
3. 恢复配置: cp .env ./
4. 重启服务: docker-compose up -d

注意事项:
- 恢复前请确保服务已停止
- 建议在恢复前备份当前数据
- 检查配置文件中的域名和密钥设置
EOF
    
    log_success "备份信息文件创建完成"
}

# 压缩备份
compress_backup() {
    log_info "压缩备份文件..."
    
    pushd "${BACKUP_DIR}" > /dev/null # Push current dir, change to BACKUP_DIR
    tar -czf "${BACKUP_NAME}.tar.gz" "${BACKUP_NAME}"
    rm -rf "${BACKUP_NAME}"
    popd > /dev/null # Pop back to original dir
    
    log_success "备份压缩完成: ${BACKUP_DIR}/${BACKUP_NAME}.tar.gz"
}

# 清理旧备份 (保留最近7个)
cleanup_old_backups() {
    log_info "清理旧备份文件..."
    
    pushd "${BACKUP_DIR}" > /dev/null # Push current dir, change to BACKUP_DIR
    ls -t email_server_backup_*.tar.gz | tail -n +8 | xargs -r rm -f
    popd > /dev/null # Pop back to original dir
    
    log_success "旧备份清理完成"
}

# 主函数
main() {
    log_info "开始备份Email Server数据..."
    
    create_backup_dir
    backup_database
    backup_configs
    create_backup_info
    compress_backup
    cleanup_old_backups
    
    log_success "备份完成！"
    log_info "备份文件: ${BACKUP_DIR}/${BACKUP_NAME}.tar.gz"
}

# 执行主函数
main "$@"
