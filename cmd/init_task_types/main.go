package main

import (
	"io/ioutil"
	"log"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 数据库连接配置
	dsn := "jduser:jdpass123@tcp(localhost:3306)/jd?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		log.Fatal("❌ 数据库连接失败:", err)
	}

	log.Println("✓ 数据库连接成功")

	// 读取SQL文件
	sqlFile := "scripts/init_task_types.sql"
	content, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		log.Fatal("❌ 读取SQL文件失败:", err)
	}

	log.Println("\n========== 执行SQL脚本 ==========\n")

	// 分割SQL语句
	sqlStatements := strings.Split(string(content), ";")

	for _, sql := range sqlStatements {
		sql = strings.TrimSpace(sql)
		// 跳过空语句和注释
		if sql == "" || strings.HasPrefix(sql, "--") {
			continue
		}

		if err := db.Exec(sql).Error; err != nil {
			// 忽略字段已存在的错误
			if !strings.Contains(err.Error(), "Duplicate column") {
				log.Printf("⚠️ 警告: %v", err)
			}
		}
	}

	log.Println("✓ SQL脚本执行完成")

	// 验证结果
	log.Println("\n========== 验证初始化结果 ==========\n")

	type TaskType struct {
		ID             uint
		TypeCode       string
		TypeName       string
		JingdouPrice   int
		IsActive       bool
		TimeSlot1Start *string
		TimeSlot1End   *string
		TimeSlot2Start *string
		TimeSlot2End   *string
		IsSystemPreset bool
	}

	var types []TaskType
	db.Table("task_types").Find(&types)

	if len(types) == 0 {
		log.Fatal("❌ 没有找到任务类型")
	}

	log.Printf("✓ 任务类型总数: %d\n", len(types))
	log.Println("所有任务类型:")

	for _, t := range types {
		timeSlot := "无时间限制"
		if t.TimeSlot1Start != nil && t.TimeSlot1End != nil {
			timeSlot = *t.TimeSlot1Start + "-" + *t.TimeSlot1End
			if t.TimeSlot2Start != nil && t.TimeSlot2End != nil {
				timeSlot += ", " + *t.TimeSlot2Start + "-" + *t.TimeSlot2End
			}
		}
		presetStr := ""
		if t.IsSystemPreset {
			presetStr = " [系统预设]"
		}
		log.Printf("  [%d] %s (%s) - %d京豆 - %s - 启用: %v%s",
			t.ID, t.TypeName, t.TypeCode, t.JingdouPrice, timeSlot, t.IsActive, presetStr)
	}

	log.Println("\n========================================")
	log.Println("✓ 任务类型初始化完成！")
	log.Println("========================================")
}
