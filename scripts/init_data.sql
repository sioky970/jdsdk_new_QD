-- ========================================
-- 数据库初始化脚本
-- ========================================

-- 1. 初始化系统设置
-- ========================================
INSERT INTO settings (param_key, param_value, param_type, description, updated_at) VALUES
('login_announcement', '欢迎使用JD任务平台！', 'string', '登录页公告', NOW()),
('system_name', 'JD任务平台', 'string', '系统名称', NOW()),
('default_jingdou', '100', 'integer', '新用户默认京豆数', NOW()),
('min_jingdou_balance', '0', 'integer', '最低京豆余额', NOW()),
('task_create_time_start', '08:00', 'string', '任务创建时间段开始', NOW()),
('task_create_time_end', '22:00', 'string', '任务创建时间段结束', NOW()),
('max_tasks_per_day', '50', 'integer', '每日最大任务数', NOW()),
('device_auth_key', 'KKNN778899', 'string', '设备认证密钥（用于设备端API认证）', NOW())
ON DUPLICATE KEY UPDATE 
    param_value = VALUES(param_value),
    updated_at = NOW();

-- 2. 验证数据
SELECT '系统设置初始化完成' as status;
SELECT COUNT(*) as settings_count FROM settings;

SELECT '任务类型初始化状态' as status;
SELECT COUNT(*) as task_types_count FROM task_types;

SELECT '用户状态' as status;
SELECT COUNT(*) as total_users, 
       SUM(CASE WHEN role='admin' THEN 1 ELSE 0 END) as admin_count,
       SUM(CASE WHEN role='user' THEN 1 ELSE 0 END) as user_count
FROM users;
