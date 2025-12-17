package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 连接数据库
	db, err := sql.Open("sqlite3", "jd_task.db")
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 添加 qrcode_url 字段到 proxies 表
	log.Println("开始添加 qrcode_url 字段...")

	// 检查字段是否已存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('proxies') WHERE name='qrcode_url'").Scan(&count)
	if err != nil {
		log.Fatalf("检查字段失败: %v", err)
	}

	if count > 0 {
		log.Println("✅ qrcode_url 字段已存在，无需添加")
		return
	}

	// 添加字段
	_, err = db.Exec("ALTER TABLE proxies ADD COLUMN qrcode_url TEXT")
	if err != nil {
		log.Fatalf("❌ 添加字段失败: %v", err)
	}

	log.Println("✅ 成功添加 qrcode_url 字段")

	// 统计代理总数
	var totalProxies int
	err = db.QueryRow("SELECT COUNT(*) FROM proxies").Scan(&totalProxies)
	if err != nil {
		log.Printf("⚠️  统计代理记录失败: %v", err)
	} else if totalProxies > 0 {
		log.Printf("⚠️  检测到 %d 条现有代理记录", totalProxies)
		log.Println("提示: 需要重新启动后端服务，现有代理记录会在下次更新时自动生成二维码URL")
	}

	fmt.Println("\n数据库迁移完成！")
}
