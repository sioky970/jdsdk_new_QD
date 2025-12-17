// +build ignore

package main

import (
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 数据库连接
	dsn := "jduser:jdpass123@tcp(localhost:3306)/jd?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		dsn = envDSN
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	log.Println("✓ 数据库连接成功")

	// 检查列是否存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = 'jd' AND table_name = 'task_types' AND column_name = 'execute_multiplier'").Count(&count)

	if count == 0 {
		// 添加列
		result := db.Exec("ALTER TABLE task_types ADD COLUMN execute_multiplier INT DEFAULT 1")
		if result.Error != nil {
			log.Fatal("添加列失败:", result.Error)
		}
		log.Println("✓ 成功添加 execute_multiplier 列")
	} else {
		log.Println("✓ execute_multiplier 列已存在")
	}

	log.Println("迁移完成！")
}
