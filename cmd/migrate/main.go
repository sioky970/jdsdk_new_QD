package main

import (
	"log"

	"jd-task-platform-go/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 数据库连接配置
	dsn := "jduser:jdpass123@tcp(localhost:3306)/jd?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("❌ 数据库连接失败:", err)
	}

	log.Println("✓ 数据库连接成功")

	// 需要迁移的表
	tables := []interface{}{
		&models.Setting{},
		&models.TaskType{},
		&models.JingdouLog{},
		&models.APILog{},
		&models.TaskLog{},
		&models.DeviceTaskHistory{},
	}

	log.Println("\n========== 开始数据库表迁移 ==========")

	// 检查并创建每个表
	for _, model := range tables {
		// 执行迁移
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("❌ 表迁移失败: %v", err)
			continue
		}

		// 获取表名
		stmt := &gorm.Statement{DB: db}
		stmt.Parse(model)
		tableName := stmt.Schema.Table

		log.Printf("✓ 表已就绪: %s", tableName)
	}

	log.Println("\n========== 数据库表迁移完成 ==========")
	log.Println("\n验证表结构...")

	// 验证表
	var count int64
	db.Table("settings").Count(&count)
	log.Printf("✓ settings 表记录数: %d", count)

	db.Table("task_types").Count(&count)
	log.Printf("✓ task_types 表记录数: %d", count)

	db.Table("jingdou_logs").Count(&count)
	log.Printf("✓ jingdou_logs 表记录数: %d", count)

	db.Table("api_logs").Count(&count)
	log.Printf("✓ api_logs 表记录数: %d", count)

	db.Table("task_logs").Count(&count)
	log.Printf("✓ task_logs 表记录数: %d", count)

	db.Table("device_task_history").Count(&count)
	log.Printf("✓ device_task_history 表记录数: %d", count)

	log.Println("\n✓ 所有表已成功创建并验证完成！")
}
