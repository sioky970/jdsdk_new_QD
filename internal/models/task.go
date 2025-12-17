package models

import (
	"time"
)

// Task 任务模型
type Task struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null;column:user_id" json:"user_id"`
	TaskType       string    `gorm:"size:32;not null;column:task_type" json:"task_type"`
	SKU            string    `gorm:"size:64;not null" json:"sku"`
	ShopName       string    `gorm:"size:128;column:shop_name" json:"shop_name"`
	Keyword        string    `gorm:"size:128" json:"keyword"`
	StartTime      time.Time `gorm:"not null;column:start_time" json:"start_time"`
	ExecuteCount   int       `gorm:"not null;column:execute_count" json:"execute_count"`
	ExecutedCount  int       `gorm:"default:0;column:executed_count" json:"executed_count"`
	Priority       int       `gorm:"default:0" json:"priority"`
	Status         string    `gorm:"size:20;not null" json:"status"`
	ConsumeJingdou int       `gorm:"not null;column:consume_jingdou" json:"consume_jingdou"`
	Remark         string    `gorm:"type:text" json:"remark"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	TaskType     string    `json:"task_type" binding:"required" example:"search_order"`
	SKU          string    `json:"sku" binding:"required" example:"100001234567"`
	ShopName     string    `json:"shop_name" example:"京东自营店"`
	Keyword      string    `json:"keyword" example:"手机"`
	StartTime    time.Time `json:"start_time" binding:"required" example:"2023-12-01T10:00:00Z"`
	ExecuteCount int       `json:"execute_count" binding:"required" example:"10"`
	Priority     int       `json:"priority" example:"1"`
	Remark       string    `json:"remark" example:"测试任务"`
	// 注意: consume_jingdou 由服务端根据任务类型和执行次数自动计算，不接受客户端传入
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	ShopName     *string    `json:"shop_name" example:"新店铺名称"`
	Keyword      *string    `json:"keyword" example:"新关键词"`
	StartTime    *time.Time `json:"start_time" example:"2023-12-01T10:00:00Z"`
	ExecuteCount *int       `json:"execute_count" example:"20"`
	Priority     *int       `json:"priority" example:"2"`
	Status       *string    `json:"status" example:"running"`
	Remark       *string    `json:"remark" example:"更新备注"`
}
