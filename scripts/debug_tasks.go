package main

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Task struct {
	ID            uint      `gorm:"primarykey"`
	UserID        uint      `json:"user_id"`
	TaskType      string    `json:"task_type"`
	SKU           string    `json:"sku"`
	Status        string    `json:"status"`
	ExecuteCount  int       `json:"execute_count"`
	ExecutedCount int       `json:"executed_count"`
	StartTime     time.Time `json:"start_time"`
	CreatedAt     time.Time
}

type User struct {
	ID       uint   `gorm:"primarykey"`
	Username string `json:"username"`
}

func main() {
	dsn := "root:root@tcp(127.0.0.1:3306)/jd?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		return
	}

	// 获取 user001 的 ID
	var user User
	if err := db.Where("username = ?", "user001").First(&user).Error; err != nil {
		fmt.Printf("找不到 user001: %v\n", err)
		return
	}
	fmt.Printf("用户 user001 ID: %d\n\n", user.ID)

	// 今天的时间范围
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	todayEnd := todayStart.Add(24 * time.Hour)

	fmt.Printf("今日时间范围: %s ~ %s\n\n", todayStart.Format("2006-01-02 15:04:05"), todayEnd.Format("2006-01-02 15:04:05"))

	// 查询今天 created_at 的所有任务（不排除已取消）
	var tasksCreatedToday []Task
	db.Where("user_id = ? AND created_at >= ? AND created_at < ?", user.ID, todayStart, todayEnd).Find(&tasksCreatedToday)
	fmt.Printf("=== 今日创建的任务 (按 created_at 筛选，共 %d 条) ===\n", len(tasksCreatedToday))
	for _, t := range tasksCreatedToday {
		fmt.Printf("ID: %d, Status: %s, ExecuteCount: %d, ExecutedCount: %d, CreatedAt: %s, StartTime: %s\n",
			t.ID, t.Status, t.ExecuteCount, t.ExecutedCount,
			t.CreatedAt.Format("2006-01-02 15:04:05"),
			t.StartTime.Format("2006-01-02 15:04:05"))
	}

	// 计算统计
	var totalExecuteCount, totalExecutedCount int64
	for _, t := range tasksCreatedToday {
		if t.Status != "cancelled" {
			totalExecuteCount += int64(t.ExecuteCount)
			totalExecutedCount += int64(t.ExecutedCount)
		}
	}
	fmt.Printf("\n排除 cancelled 后的统计:\n")
	fmt.Printf("任务数: %d\n", len(tasksCreatedToday))
	fmt.Printf("总执行次数(execute_count): %d\n", totalExecuteCount)
	fmt.Printf("已执行次数(executed_count): %d\n", totalExecutedCount)

	// 查询今天 start_time 的所有任务
	var tasksStartToday []Task
	db.Where("user_id = ? AND start_time >= ? AND start_time < ?", user.ID, todayStart, todayEnd).Find(&tasksStartToday)
	fmt.Printf("\n=== 今日开始的任务 (按 start_time 筛选，共 %d 条) ===\n", len(tasksStartToday))
	for _, t := range tasksStartToday {
		fmt.Printf("ID: %d, Status: %s, ExecuteCount: %d, ExecutedCount: %d, CreatedAt: %s, StartTime: %s\n",
			t.ID, t.Status, t.ExecuteCount, t.ExecutedCount,
			t.CreatedAt.Format("2006-01-02 15:04:05"),
			t.StartTime.Format("2006-01-02 15:04:05"))
	}

	// 查询用户所有任务
	var allTasks []Task
	db.Where("user_id = ?", user.ID).Order("created_at DESC").Limit(20).Find(&allTasks)
	fmt.Printf("\n=== 用户最近20条任务 ===\n")
	for _, t := range allTasks {
		fmt.Printf("ID: %d, Status: %s, ExecuteCount: %d, ExecutedCount: %d, CreatedAt: %s, StartTime: %s\n",
			t.ID, t.Status, t.ExecuteCount, t.ExecutedCount,
			t.CreatedAt.Format("2006-01-02 15:04:05"),
			t.StartTime.Format("2006-01-02 15:04:05"))
	}
}
