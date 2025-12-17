package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 数据库连接配置
	dsn := "jduser:jdpass123@tcp(127.0.0.1:3306)/jd?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("  修改 device_id 字段类型为 VARCHAR")
	fmt.Println("========================================")
	fmt.Println()

	// 执行 SQL
	sql := "ALTER TABLE proxy_usage_logs MODIFY COLUMN device_id VARCHAR(100) NOT NULL"

	fmt.Println("执行 SQL:", sql)
	if err := db.Exec(sql).Error; err != nil {
		log.Fatalf("❌ 执行失败: %v", err)
	}

	fmt.Println("✅ 字段类型修改成功！")
	fmt.Println()

	// 验证
	fmt.Println("验证表结构...")
	var result []map[string]interface{}
	db.Raw("DESCRIBE proxy_usage_logs").Scan(&result)

	for _, row := range result {
		if row["Field"] == "device_id" {
			fmt.Printf("  device_id 字段类型: %s\n", row["Type"])
		}
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("✅ 完成！")
	fmt.Println("========================================")
}
