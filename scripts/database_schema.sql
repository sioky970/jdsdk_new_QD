-- ========================================
-- JD Task Platform 数据库结构
-- 导出时间: 2025-12-17
-- 数据库: MySQL
-- ========================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. 用户表 (users)
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `username` varchar(64) NOT NULL COMMENT '用户名',
  `password_hash` varchar(128) NOT NULL COMMENT '密码哈希',
  `nickname` varchar(64) DEFAULT NULL COMMENT '昵称',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像URL',
  `role` varchar(20) DEFAULT 'common' COMMENT '角色: admin/common',
  `api_key` varchar(64) DEFAULT NULL COMMENT 'API密钥',
  `api_key_created_at` datetime DEFAULT NULL COMMENT 'API密钥创建时间',
  `api_key_last_used_at` datetime DEFAULT NULL COMMENT 'API密钥最后使用时间',
  `jingdou_balance` int DEFAULT '0' COMMENT '京豆余额',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `last_login` datetime DEFAULT NULL COMMENT '最后登录时间',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否激活',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  UNIQUE KEY `idx_api_key` (`api_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ----------------------------
-- 2. 任务表 (tasks)
-- ----------------------------
DROP TABLE IF EXISTS `tasks`;
CREATE TABLE `tasks` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `task_type` varchar(32) NOT NULL COMMENT '任务类型: browse/search_order/add_cart等',
  `sku` varchar(64) NOT NULL COMMENT '商品SKU',
  `shop_name` varchar(128) DEFAULT NULL COMMENT '店铺名称',
  `keyword` varchar(128) DEFAULT NULL COMMENT '搜索关键词',
  `start_time` datetime NOT NULL COMMENT '开始执行时间',
  `execute_count` int NOT NULL COMMENT '需要执行次数',
  `executed_count` int DEFAULT '0' COMMENT '已执行次数',
  `priority` int DEFAULT '0' COMMENT '优先级',
  `status` varchar(20) NOT NULL COMMENT '状态: pending/running/completed/cancelled/expired',
  `consume_jingdou` int NOT NULL COMMENT '消耗京豆数',
  `remark` text COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_start_time` (`start_time`),
  KEY `idx_task_type` (`task_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务表';

-- ----------------------------
-- 3. 设备表 (devices)
-- ----------------------------
DROP TABLE IF EXISTS `devices`;
CREATE TABLE `devices` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `device_id` varchar(64) NOT NULL COMMENT '设备唯一标识',
  `device_name` varchar(128) NOT NULL COMMENT '设备名称',
  `device_type` varchar(20) DEFAULT NULL COMMENT '设备类型: android/ios',
  `device_model` varchar(64) DEFAULT NULL COMMENT '设备型号',
  `os_version` varchar(32) DEFAULT NULL COMMENT '系统版本',
  `app_version` varchar(32) DEFAULT NULL COMMENT '应用版本',
  `ip` varchar(64) DEFAULT NULL COMMENT 'IP地址',
  `location` varchar(128) DEFAULT NULL COMMENT '地理位置',
  `os_info` varchar(128) DEFAULT NULL COMMENT '系统信息(兼容旧字段)',
  `version` varchar(32) DEFAULT NULL COMMENT '版本(兼容旧字段)',
  `status` varchar(20) NOT NULL COMMENT '状态: online/offline/working/idle',
  `is_blocked` tinyint(1) DEFAULT '0' COMMENT '是否被封禁',
  `last_heartbeat` datetime DEFAULT NULL COMMENT '最后心跳时间',
  `last_active` datetime DEFAULT NULL COMMENT '最后活跃时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `last_task_time` datetime DEFAULT NULL COMMENT '最后执行任务时间',
  `task_count` int DEFAULT '0' COMMENT '任务执行次数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_device_id` (`device_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='设备表';

-- ----------------------------
-- 4. 任务类型表 (task_types)
-- ----------------------------
DROP TABLE IF EXISTS `task_types`;
CREATE TABLE `task_types` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `type_code` varchar(32) NOT NULL COMMENT '类型编码',
  `type_name` varchar(64) NOT NULL COMMENT '类型名称',
  `jingdou_price` int NOT NULL COMMENT '单次消耗京豆数',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否启用',
  `execute_multiplier` int DEFAULT '1' COMMENT '执行倍数',
  `time_slot1_start` varchar(5) DEFAULT NULL COMMENT '时间段1开始 HH:MM',
  `time_slot1_end` varchar(5) DEFAULT NULL COMMENT '时间段1结束 HH:MM',
  `time_slot2_start` varchar(5) DEFAULT NULL COMMENT '时间段2开始 HH:MM',
  `time_slot2_end` varchar(5) DEFAULT NULL COMMENT '时间段2结束 HH:MM',
  `is_system_preset` tinyint(1) DEFAULT '0' COMMENT '是否系统预设',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_type_code` (`type_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务类型表';

-- ----------------------------
-- 5. 任务日志表 (task_logs)
-- ----------------------------
DROP TABLE IF EXISTS `task_logs`;
CREATE TABLE `task_logs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `task_id` int unsigned NOT NULL COMMENT '任务ID',
  `device_id` varchar(64) DEFAULT NULL COMMENT '执行设备ID',
  `status` varchar(20) NOT NULL COMMENT '执行状态: success/failed',
  `message` text COMMENT '执行消息',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_device_id` (`device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务执行日志表';

-- ----------------------------
-- 6. 京豆日志表 (jingdou_logs)
-- ----------------------------
DROP TABLE IF EXISTS `jingdou_logs`;
CREATE TABLE `jingdou_logs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `amount` int NOT NULL COMMENT '变动数量(正数增加,负数减少)',
  `balance` int NOT NULL COMMENT '变动后余额',
  `operation_type` varchar(20) NOT NULL COMMENT '操作类型: recharge/consume/refund/admin_adjust',
  `related_id` int unsigned DEFAULT NULL COMMENT '关联ID(如任务ID)',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_operation_type` (`operation_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='京豆变动日志表';

-- ----------------------------
-- 7. 系统设置表 (settings)
-- ----------------------------
DROP TABLE IF EXISTS `settings`;
CREATE TABLE `settings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `param_key` varchar(64) NOT NULL COMMENT '参数键',
  `param_value` text COMMENT '参数值',
  `param_type` varchar(20) DEFAULT NULL COMMENT '参数类型: string/integer/boolean/json',
  `description` varchar(255) DEFAULT NULL COMMENT '参数描述',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_param_key` (`param_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统设置表';

-- ----------------------------
-- 8. API调用日志表 (api_logs)
-- ----------------------------
DROP TABLE IF EXISTS `api_logs`;
CREATE TABLE `api_logs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` int unsigned DEFAULT NULL COMMENT '用户ID',
  `api_key` varchar(64) DEFAULT NULL COMMENT 'API密钥',
  `endpoint` varchar(200) NOT NULL COMMENT '接口地址',
  `method` varchar(10) NOT NULL COMMENT '请求方法',
  `ip` varchar(45) DEFAULT NULL COMMENT '请求IP',
  `user_agent` text COMMENT 'User-Agent',
  `response_code` int NOT NULL COMMENT '响应状态码',
  `response_time` double DEFAULT NULL COMMENT '响应时间(ms)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_api_key` (`api_key`),
  KEY `idx_endpoint` (`endpoint`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API调用日志表';

-- ----------------------------
-- 9. 设备任务历史表 (device_task_history)
-- ----------------------------
DROP TABLE IF EXISTS `device_task_history`;
CREATE TABLE `device_task_history` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `device_id` varchar(64) NOT NULL COMMENT '设备ID',
  `task_id` int unsigned NOT NULL COMMENT '任务ID',
  `sku` varchar(64) NOT NULL COMMENT '商品SKU',
  `execute_time` datetime NOT NULL COMMENT '执行时间',
  `status` varchar(20) NOT NULL COMMENT '执行状态',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_device_id` (`device_id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_sku` (`sku`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='设备任务历史表';

-- ----------------------------
-- 10. 代理表 (proxies)
-- ----------------------------
DROP TABLE IF EXISTS `proxies`;
CREATE TABLE `proxies` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `ip` varchar(50) NOT NULL COMMENT '代理IP',
  `port` int NOT NULL COMMENT '代理端口',
  `username` varchar(100) NOT NULL COMMENT '用户名',
  `password` varchar(100) NOT NULL COMMENT '密码',
  `province` varchar(50) DEFAULT NULL COMMENT '省份',
  `city` varchar(50) DEFAULT NULL COMMENT '城市',
  `isp` varchar(50) DEFAULT NULL COMMENT '运营商',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `qrcode_url` text COMMENT 'Clash Mi二维码URL',
  `usage_count` int DEFAULT '0' COMMENT '使用次数',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否启用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_ip` (`ip`),
  KEY `idx_usage_count` (`usage_count`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代理表';

-- ----------------------------
-- 11. 代理使用日志表 (proxy_usage_logs)
-- ----------------------------
DROP TABLE IF EXISTS `proxy_usage_logs`;
CREATE TABLE `proxy_usage_logs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `proxy_id` int unsigned NOT NULL COMMENT '代理ID',
  `device_id` varchar(100) NOT NULL COMMENT '设备ID',
  `device_sn` varchar(100) DEFAULT NULL COMMENT '设备序列号',
  `ip` varchar(50) DEFAULT NULL COMMENT '代理IP',
  `port` int DEFAULT NULL COMMENT '代理端口',
  `assigned_at` datetime DEFAULT NULL COMMENT '分配时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_proxy_id` (`proxy_id`),
  KEY `idx_device_id` (`device_id`),
  KEY `idx_device_sn` (`device_sn`),
  KEY `idx_assigned_at` (`assigned_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代理使用日志表';

-- ----------------------------
-- 12. 任务模板表 (task_templates)
-- ----------------------------
DROP TABLE IF EXISTS `task_templates`;
CREATE TABLE `task_templates` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` int unsigned NOT NULL COMMENT '用户ID',
  `task_type` varchar(32) NOT NULL COMMENT '任务类型',
  `sku` varchar(64) NOT NULL COMMENT '商品SKU',
  `shop_name` varchar(128) DEFAULT NULL COMMENT '店铺名称',
  `keyword` varchar(256) DEFAULT NULL COMMENT '关键词(多个逗号分隔)',
  `remark` text COMMENT '模板备注',
  `total_created_count` int DEFAULT '0' COMMENT '使用此模板创建任务总次数',
  `last_used_at` datetime DEFAULT NULL COMMENT '最后使用时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_sku` (`sku`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务模板表';

-- ----------------------------
-- 初始化系统设置数据
-- ----------------------------
INSERT INTO `settings` (`param_key`, `param_value`, `param_type`, `description`) VALUES
('login_announcement', '欢迎使用JD任务平台！', 'string', '登录页公告'),
('system_name', 'JD任务平台', 'string', '系统名称'),
('default_jingdou', '100', 'integer', '新用户默认京豆数'),
('min_jingdou_balance', '0', 'integer', '最低京豆余额'),
('task_create_time_start', '08:00', 'string', '任务创建时间段开始'),
('task_create_time_end', '22:00', 'string', '任务创建时间段结束'),
('max_tasks_per_day', '50', 'integer', '每日最大任务数'),
('device_auth_key', 'KKNN778899', 'string', '设备认证密钥')
ON DUPLICATE KEY UPDATE `updated_at` = NOW();

-- ----------------------------
-- 初始化任务类型数据
-- ----------------------------
INSERT INTO `task_types` (`type_code`, `type_name`, `jingdou_price`, `is_active`, `is_system_preset`) VALUES
('browse', '浏览任务', 1, 1, 1),
('search_order', '关键词搜索下单', 3, 1, 1),
('add_cart', '加购任务', 2, 1, 1),
('collect', '收藏任务', 1, 1, 1),
('follow_shop', '关注店铺', 1, 1, 1)
ON DUPLICATE KEY UPDATE `updated_at` = NOW();

SET FOREIGN_KEY_CHECKS = 1;

-- ========================================
-- 数据库结构导出完成
-- 共计 12 张表
-- ========================================
