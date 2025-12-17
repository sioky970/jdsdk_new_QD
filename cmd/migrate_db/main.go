package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 连接数据库
	dsn := "jduser:jdpass123@tcp(localhost:3306)/jd?charset=utf8mb4&parseTime=True"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("❌ 数据库连接失败:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("❌ 数据库Ping失败:", err)
	}

	log.Println("✓ 数据库连接成功\n")

	// 执行SQL语句
	sqlStatements := []string{
		// 禁用外键检查
		"SET FOREIGN_KEY_CHECKS=0",

		// 添加新字段
		"ALTER TABLE task_types ADD COLUMN time_slot1_start VARCHAR(5) DEFAULT NULL COMMENT '时间段1开始'",
		"ALTER TABLE task_types ADD COLUMN time_slot1_end VARCHAR(5) DEFAULT NULL COMMENT '时间段1结束'",
		"ALTER TABLE task_types ADD COLUMN time_slot2_start VARCHAR(5) DEFAULT NULL COMMENT '时间段2开始'",
		"ALTER TABLE task_types ADD COLUMN time_slot2_end VARCHAR(5) DEFAULT NULL COMMENT '时间段2结束'",
		"ALTER TABLE task_types ADD COLUMN is_system_preset TINYINT(1) DEFAULT 0 COMMENT '是否系统预设'",

		// 清空现有数据
		"DELETE FROM task_types",

		// 插入预设任务类型
		`INSERT INTO task_types (type_code, type_name, jingdou_price, is_active, time_slot1_start, time_slot1_end, time_slot2_start, time_slot2_end, is_system_preset, created_at, updated_at) VALUES
		('browse', '浏览任务', 2, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()),
		('search_browse', '关键词搜索浏览任务', 3, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()),
		('add_to_cart', '加入购物车任务', 5, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()),
		('follow_shop', '关注店铺任务', 4, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()),
		('follow_product', '收藏商品任务', 4, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW()),
		('purchase', '购买商品任务', 10, 1, '08:00', '12:00', '14:00', '18:00', 1, NOW(), NOW())`,

		// 重新启用外键检查
		"SET FOREIGN_KEY_CHECKS=1",
	}

	log.Println("========== 执行SQL迁移 ==========\n")

	for i, sqlStmt := range sqlStatements {
		_, err := db.Exec(sqlStmt)
		if err != nil {
			// 忽略字段已存在的错误
			if err.Error() != "Error 1060: Duplicate column name 'time_slot1_start'" &&
				err.Error() != "Error 1060: Duplicate column name 'time_slot1_end'" &&
				err.Error() != "Error 1060: Duplicate column name 'time_slot2_start'" &&
				err.Error() != "Error 1060: Duplicate column name 'time_slot2_end'" &&
				err.Error() != "Error 1060: Duplicate column name 'is_system_preset'" {
				log.Printf("⚠️  SQL #%d 执行警告: %v", i+1, err)
			}
		}
	}

	log.Println("✓ SQL迁移执行完成\n")

	// 验证结果
	log.Println("========== 验证迁移结果 ==========\n")

	rows, err := db.Query("SELECT id, type_code, type_name, jingdou_price, is_active, time_slot1_start, time_slot1_end, time_slot2_start, time_slot2_end, is_system_preset FROM task_types ORDER BY id")
	if err != nil {
		log.Fatal("❌ 查询失败:", err)
	}
	defer rows.Close()

	count := 0
	fmt.Println("任务类型列表:")
	fmt.Println("================================================================================")

	for rows.Next() {
		var (
			id             int
			typeCode       string
			typeName       string
			jingdouPrice   int
			isActive       bool
			timeSlot1Start sql.NullString
			timeSlot1End   sql.NullString
			timeSlot2Start sql.NullString
			timeSlot2End   sql.NullString
			isSystemPreset bool
		)

		if err := rows.Scan(&id, &typeCode, &typeName, &jingdouPrice, &isActive, &timeSlot1Start, &timeSlot1End, &timeSlot2Start, &timeSlot2End, &isSystemPreset); err != nil {
			log.Fatal("❌ 扫描失败:", err)
		}

		timeSlot := "无时间限制"
		if timeSlot1Start.Valid && timeSlot1End.Valid {
			timeSlot = fmt.Sprintf("%s-%s", timeSlot1Start.String, timeSlot1End.String)
			if timeSlot2Start.Valid && timeSlot2End.Valid {
				timeSlot += fmt.Sprintf(", %s-%s", timeSlot2Start.String, timeSlot2End.String)
			}
		}

		presetMark := ""
		if isSystemPreset {
			presetMark = " [系统预设]"
		}

		fmt.Printf("[%d] %s (%s)\n", id, typeName, typeCode)
		fmt.Printf("    价格: %d京豆 | 启用: %v | 时间段: %s%s\n", jingdouPrice, isActive, timeSlot, presetMark)
		fmt.Println("--------------------------------------------------------------------------------")

		count++
	}

	fmt.Printf("\n✓ 共 %d 个任务类型\n", count)

	if count != 6 {
		log.Fatal("❌ 任务类型数量不正确，应该是6个")
	}

	log.Println("\n========================================")
	log.Println("✓ 数据库迁移成功完成！")
	log.Println("========================================")
}
