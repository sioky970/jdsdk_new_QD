package models

import (
	"time"
)

// Proxy SK5代理信息
type Proxy struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	IP         string    `gorm:"size:50;not null;index" json:"ip" example:"42.101.12.24"`
	Port       int       `gorm:"not null" json:"port" example:"11011"`
	Username   string    `gorm:"size:100;not null" json:"username" example:"chtJZ0530135"`
	Password   string    `gorm:"size:100;not null" json:"password" example:"3678"`
	Province   string    `gorm:"size:50" json:"province" example:"北京"`
	City       string    `gorm:"size:50" json:"city" example:"北京市"`
	ISP        string    `gorm:"size:50" json:"isp" example:"中国电信"` // 运营商
	Remark     string    `gorm:"size:500" json:"remark" example:"备注信息"`
	QRCodeURL  string    `gorm:"type:text" json:"qrcode_url" example:"clash://install-config?url=..."` // Clash Mi 二维码 URL
	UsageCount int       `gorm:"default:0;index" json:"usage_count" example:"5"`                       // 使用次数
	IsActive   bool      `gorm:"default:true" json:"is_active" example:"true"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ProxyUsageLog 代理使用记录
type ProxyUsageLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ProxyID    uint      `gorm:"not null;index" json:"proxy_id"`
	DeviceID   string    `gorm:"size:100;not null;index" json:"device_id"`
	DeviceSN   string    `gorm:"size:100;index" json:"device_sn" example:"DEVICE001"`
	IP         string    `gorm:"size:50" json:"ip" example:"42.101.12.24"`
	Port       int       `json:"port" example:"11011"`
	AssignedAt time.Time `gorm:"index" json:"assigned_at"` // 分配时间
	CreatedAt  time.Time `json:"created_at"`
}

// ProxyBatchImportRequest 批量导入请求
type ProxyBatchImportRequest struct {
	ProxyList string `json:"proxy_list" binding:"required" example:"42.101.12.24|11011|user|pass\n110.166.73.212|11006|user2|pass2"`
}

// ProxyBatchDeleteRequest 批量删除请求
type ProxyBatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required" example:"[1,2,3]"`
}

// ProxyCreateRequest 创建代理请求
type ProxyCreateRequest struct {
	IP       string `json:"ip" binding:"required" example:"42.101.12.24"`
	Port     int    `json:"port" binding:"required,min=1,max=65535" example:"11011"`
	Username string `json:"username" binding:"required" example:"chtJZ0530135"`
	Password string `json:"password" binding:"required" example:"3678"`
	Remark   string `json:"remark" example:"备注信息"`
}

// ProxyUpdateRequest 更新代理请求
type ProxyUpdateRequest struct {
	IP       *string `json:"ip" example:"42.101.12.24"`
	Port     *int    `json:"port" example:"11011"`
	Username *string `json:"username" example:"chtJZ0530135"`
	Password *string `json:"password" example:"3678"`
	Remark   *string `json:"remark" example:"备注信息"`
	IsActive *bool   `json:"is_active" example:"true"`
}

// ProxyAssignRequest 代理分配请求
type ProxyAssignRequest struct {
	DeviceID string `json:"device_id" binding:"required" example:"a6a5ba66839ecfa1d3350155e8b1db8d"`
	DeviceSN string `json:"device_sn" binding:"required" example:"DEVICE001"`
}

// ProxyAssignResponse 代理分配响应
type ProxyAssignResponse struct {
	IP       string `json:"ip" example:"42.101.12.24"`
	Port     int    `json:"port" example:"11011"`
	Username string `json:"username" example:"chtJZ0530135"`
	Password string `json:"password" example:"3678"`
	ProxyID  uint   `json:"proxy_id" example:"1"`
}

// ProxyStatistics 代理统计信息
type ProxyStatistics struct {
	TotalCount    int64   `json:"total_count" example:"100"`
	ActiveCount   int64   `json:"active_count" example:"95"`
	InactiveCount int64   `json:"inactive_count" example:"5"`
	TotalUsage    int64   `json:"total_usage" example:"1500"`
	AvgUsage      float64 `json:"avg_usage" example:"15.5"`
}
