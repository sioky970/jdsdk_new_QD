-- 修改 proxy_usage_logs 表的 device_id 字段类型
-- 从 int 改为 varchar(100)

USE jd;

-- 1. 修改 device_id 字段类型
ALTER TABLE proxy_usage_logs MODIFY COLUMN device_id VARCHAR(100) NOT NULL;

-- 2. 确认修改
DESCRIBE proxy_usage_logs;
