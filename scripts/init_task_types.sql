-- 1. 禁用外键检查
SET FOREIGN_KEY_CHECKS=0;

-- 2. 添加新字段（如果不存在）
ALTER TABLE task_types 
ADD COLUMN time_slot1_start VARCHAR(5) DEFAULT NULL COMMENT '时间段1开始',
ADD COLUMN time_slot1_end VARCHAR(5) DEFAULT NULL COMMENT '时间段1结束',
ADD COLUMN time_slot2_start VARCHAR(5) DEFAULT NULL COMMENT '时间段2开始',
ADD COLUMN time_slot2_end VARCHAR(5) DEFAULT NULL COMMENT '时间段2结束',
ADD COLUMN is_system_preset TINYINT(1) DEFAULT 0 COMMENT '是否系统预设';

-- 3. 删除现有任务类型
DELETE FROM task_types;

-- 4. 插入6种预设任务类型
INSERT INTO task_types (type_code, type_name, jingdou_price, is_active, time_slot1_start, time_slot1_end, time_slot2_start, time_slot2_end, is_system_preset, created_at, updated_at) VALUES
('browse', '浏览任务', 2, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()),
('search_browse', '关键词搜索浏览任务', 3, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()),
('add_to_cart', '加入购物车任务', 5, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()),
('follow_shop', '关注店铺任务', 4, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()),
('follow_product', '收藏商品任务', 4, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()),
('purchase', '购买商品任务', 10, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW());

-- 5. 重新启用外键检查
SET FOREIGN_KEY_CHECKS=1;

-- 6. 查看结果
SELECT * FROM task_types;
