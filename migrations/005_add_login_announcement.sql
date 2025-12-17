-- 公告配置迁移脚本
-- 在 settings 表中添加公告配置

-- 插入公告配置（如果不存在）
INSERT INTO settings (param_key, param_value, param_type, description, updated_at)
SELECT 'login_announcement', '', 'string', '登录成功后显示的公告内容，为空则不显示', NOW()
WHERE NOT EXISTS (SELECT 1 FROM settings WHERE param_key = 'login_announcement');

-- 验证
SELECT * FROM settings WHERE param_key = 'login_announcement';
