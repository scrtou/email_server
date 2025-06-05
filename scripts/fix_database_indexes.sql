-- 修复数据库唯一索引问题
-- 删除不包含IsActive字段的旧唯一索引，这些索引会阻止软删除记录的正确处理

-- 删除旧的平台注册表唯一索引
DROP INDEX IF EXISTS idx_unique_user_platform_username_not_empty;
DROP INDEX IF EXISTS idx_unique_user_platform_email_id_not_null_zero;

-- 删除其他可能存在的旧索引
DROP INDEX IF EXISTS idx_user_platform_name;
DROP INDEX IF EXISTS idx_unique_user_platform_name_not_deleted;
DROP INDEX IF EXISTS idx_user_email;
DROP INDEX IF EXISTS idx_unique_user_email_not_deleted;

-- 验证正确的GORM生成的索引是否存在
-- 这些索引包含IsActive字段，能正确处理软删除
-- uq_user_platform_emailaccountid_active: (user_id, platform_id, email_account_id, is_active)
-- uq_user_platform_loginusername_active: (user_id, platform_id, login_username, is_active)

-- 查看当前索引状态
.schema platform_registrations
