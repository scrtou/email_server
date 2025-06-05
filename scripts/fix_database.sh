#!/bin/bash

# 数据库索引修复脚本
# 解决软删除记录阻止创建新记录的问题

set -e

echo "🔧 开始修复数据库索引问题..."

# 检查数据库文件是否存在
DB_PATH="${SQLITE_FILE:-/data/database.db}"
echo "📍 数据库路径: $DB_PATH"

if [ ! -f "$DB_PATH" ]; then
    echo "❌ 数据库文件不存在: $DB_PATH"
    exit 1
fi

# 备份数据库
BACKUP_PATH="${DB_PATH}.backup.$(date +%Y%m%d_%H%M%S)"
echo "💾 备份数据库到: $BACKUP_PATH"
cp "$DB_PATH" "$BACKUP_PATH"

# 执行SQL修复脚本
echo "🗃️ 执行索引修复..."
sqlite3 "$DB_PATH" << 'EOF'
-- 删除旧的唯一索引
DROP INDEX IF EXISTS idx_unique_user_platform_username_not_empty;
DROP INDEX IF EXISTS idx_unique_user_platform_email_id_not_null_zero;
DROP INDEX IF EXISTS idx_user_platform_name;
DROP INDEX IF EXISTS idx_unique_user_platform_name_not_deleted;
DROP INDEX IF EXISTS idx_user_email;
DROP INDEX IF EXISTS idx_unique_user_email_not_deleted;

-- 显示当前索引状态
.echo on
SELECT 'Current indexes on platform_registrations:';
.schema platform_registrations
EOF

echo "✅ 数据库索引修复完成"
echo "🎉 现在软删除记录不会阻止创建新记录了"
echo "📁 数据库备份保存在: $BACKUP_PATH"
