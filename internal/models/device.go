package models

import (
	"time"
)

// Device 设备模型
type Device struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	DeviceID      string     `gorm:"uniqueIndex;not null;size:64;column:device_id" json:"device_id"`
	DeviceName    string     `gorm:"not null;size:128;column:device_name" json:"device_name"`
	DeviceType    string     `gorm:"size:20;column:device_type" json:"device_type"`     // android 或 ios
	DeviceModel   string     `gorm:"size:64;column:device_model" json:"device_model"`   // 设备型号，如 iPhone 15 Pro
	OSVersion     string     `gorm:"size:32;column:os_version" json:"os_version"`       // 系统版本，如 iOS 17.2
	AppVersion    string     `gorm:"size:32;column:app_version" json:"app_version"`     // 应用版本
	IP            string     `gorm:"size:64" json:"ip"`
	Location      string     `gorm:"size:128;column:location" json:"location"` // 地理位置
	OSInfo        string     `gorm:"size:128;column:os_info" json:"os_info"`   // 兼容旧字段
	Version       string     `gorm:"size:32" json:"version"`                   // 兼容旧字段
	Status        string     `gorm:"size:20;not null" json:"status"`           // online, offline, working, idle
	IsBlocked     bool       `gorm:"default:false;column:is_blocked" json:"is_blocked"`
	LastHeartbeat *time.Time `gorm:"column:last_heartbeat" json:"last_heartbeat"`
	LastActive    *time.Time `gorm:"column:last_active" json:"last_active"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"created_at"`
	LastTaskTime  *time.Time `gorm:"column:last_task_time" json:"last_task_time"`
	TaskCount     int        `gorm:"default:0;column:task_count" json:"task_count"` // 任务执行次数
}

// TableName 指定表名
func (Device) TableName() string {
	return "devices"
}
