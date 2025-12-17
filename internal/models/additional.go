package models

import (
	"time"
)

// TaskType 任务类型模型
type TaskType struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	TypeCode          string    `gorm:"uniqueIndex;size:32;not null;column:type_code" json:"type_code"`
	TypeName          string    `gorm:"size:64;not null;column:type_name" json:"type_name"`
	JingdouPrice      int       `gorm:"not null;column:jingdou_price" json:"jingdou_price"`
	IsActive          bool      `gorm:"default:true;column:is_active" json:"is_active"`
	ExecuteMultiplier int       `gorm:"default:1;column:execute_multiplier" json:"-"` // 执行倍数，默认1（仅管理员可见）
	TimeSlot1Start    *string   `gorm:"size:5;column:time_slot1_start" json:"time_slot1_start"`        // 时间段1开始 HH:MM
	TimeSlot1End      *string   `gorm:"size:5;column:time_slot1_end" json:"time_slot1_end"`            // 时间段1结束 HH:MM
	TimeSlot2Start    *string   `gorm:"size:5;column:time_slot2_start" json:"time_slot2_start"`        // 时间段2开始 HH:MM
	TimeSlot2End      *string   `gorm:"size:5;column:time_slot2_end" json:"time_slot2_end"`            // 时间段2结束 HH:MM
	IsSystemPreset    bool      `gorm:"default:false;column:is_system_preset" json:"is_system_preset"` // 是否系统预设
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (TaskType) TableName() string {
	return "task_types"
}

// TaskLog 任务日志模型
type TaskLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TaskID    uint      `gorm:"not null;column:task_id" json:"task_id"`
	DeviceID  string    `gorm:"size:64;column:device_id" json:"device_id"`
	Status    string    `gorm:"size:20;not null" json:"status"`
	Message   string    `gorm:"type:text" json:"message"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 指定表名
func (TaskLog) TableName() string {
	return "task_logs"
}

// JingdouLog 京豆日志模型
type JingdouLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;column:user_id" json:"user_id"`
	Amount        int       `gorm:"not null" json:"amount"`
	Balance       int       `gorm:"not null" json:"balance"`
	OperationType string    `gorm:"size:20;not null;column:operation_type" json:"operation_type"`
	RelatedID     *uint     `gorm:"column:related_id" json:"related_id"`
	Remark        string    `gorm:"size:255" json:"remark"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 指定表名
func (JingdouLog) TableName() string {
	return "jingdou_logs"
}

// Setting 系统设置模型
type Setting struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ParamKey    string    `gorm:"uniqueIndex;size:64;not null;column:param_key" json:"param_key"`
	ParamValue  string    `gorm:"type:text;column:param_value" json:"param_value"`
	ParamType   string    `gorm:"size:20;column:param_type" json:"param_type"`
	Description string    `gorm:"size:255" json:"description"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (Setting) TableName() string {
	return "settings"
}

// APILog API调用日志模型
type APILog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       *uint     `gorm:"column:user_id" json:"user_id"`
	ApiKey       string    `gorm:"size:64;column:api_key" json:"api_key"`
	Endpoint     string    `gorm:"size:200;not null" json:"endpoint"`
	Method       string    `gorm:"size:10;not null" json:"method"`
	IP           string    `gorm:"size:45;column:ip" json:"ip"`
	UserAgent    string    `gorm:"type:text;column:user_agent" json:"user_agent"`
	ResponseCode int       `gorm:"not null;column:response_code" json:"response_code"`
	ResponseTime float64   `gorm:"column:response_time" json:"response_time"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 指定表名
func (APILog) TableName() string {
	return "api_logs"
}

// DeviceTaskHistory 设备任务历史
type DeviceTaskHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DeviceID    string    `gorm:"size:64;not null;column:device_id" json:"device_id"`
	TaskID      uint      `gorm:"not null;column:task_id" json:"task_id"`
	SKU         string    `gorm:"size:64;not null" json:"sku"`
	ExecuteTime time.Time `gorm:"not null;column:execute_time" json:"execute_time"`
	Status      string    `gorm:"size:20;not null" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName 指定表名
func (DeviceTaskHistory) TableName() string {
	return "device_task_history"
}

// CreateTaskTypeRequest 创建任务类型请求
type CreateTaskTypeRequest struct {
	TypeCode     string `json:"type_code" binding:"required" example:"add_to_cart"`
	TypeName     string `json:"type_name" binding:"required" example:"加购"`
	JingdouPrice int    `json:"jingdou_price" binding:"required" example:"2"`
	IsActive     *bool  `json:"is_active" example:"true"`
}

// UpdateTaskTypeRequest 更新任务类型请求
type UpdateTaskTypeRequest struct {
	TypeName          *string `json:"type_name" example:"浏览任务"`        // 任务类型名称（管理员可修改）
	JingdouPrice      *int    `json:"jingdou_price" example:"3"`
	IsActive          *bool   `json:"is_active" example:"false"`
	ExecuteMultiplier *int    `json:"execute_multiplier" example:"1"` // 执行倍数（仅管理员可见可修改）
	TimeSlot1Start    *string `json:"time_slot1_start" example:"08:00"` // 时间段1开始 HH:MM
	TimeSlot1End      *string `json:"time_slot1_end" example:"12:00"`   // 时间段1结束 HH:MM
	TimeSlot2Start    *string `json:"time_slot2_start" example:"14:00"` // 时间段2开始 HH:MM
	TimeSlot2End      *string `json:"time_slot2_end" example:"18:00"`   // 时间段2结束 HH:MM
}

// BatchCreateTaskRequest 批量创建任务请求
type BatchCreateTaskRequest struct {
	Tasks []CreateTaskRequest `json:"tasks" binding:"required"`
}

// TaskTemplate 任务模板模型（保存用户曾经创建的任务模板）
type TaskTemplate struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	UserID            uint      `gorm:"not null;index;column:user_id" json:"user_id"`
	TaskType          string    `gorm:"size:32;not null;column:task_type" json:"task_type"`
	SKU               string    `gorm:"size:64;not null;index" json:"sku"`
	ShopName          string    `gorm:"size:128;column:shop_name" json:"shop_name"`
	Keyword           string    `gorm:"size:256" json:"keyword"`                                         // 可能多个关键词，逗号分隔
	Remark            string    `gorm:"type:text;column:remark" json:"remark"`                           // 模板备注
	TotalCreatedCount int       `gorm:"default:0;column:total_created_count" json:"total_created_count"` // 使用此模板创建任务总次数
	LastUsedAt        time.Time `gorm:"column:last_used_at" json:"last_used_at"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName 指定表名
func (TaskTemplate) TableName() string {
	return "task_templates"
}

// QuickCreateTaskRequest 快速创建任务请求（从模板创建或直接创建）
type QuickCreateTaskRequest struct {
	TemplateID   *uint     `json:"template_id" example:"1"`    // 可选，模板ID（使用模板创建时传递）
	SKU          string    `json:"sku" example:"100001234567"` // 可选，直接创建时传递
	ExecuteCount int       `json:"execute_count" binding:"required" example:"10"`
	StartTime    time.Time `json:"start_time" binding:"required" example:"2023-12-01T10:00:00Z"`
	TaskType     string    `json:"task_type" example:"browse"` // 可选，覆盖模板的任务类型
	Keyword      string    `json:"keyword" example:"手机"`       // 可选，搜索关键词
	ShopName     string    `json:"shop_name" example:"京东自营"`   // 可选，店铺名称
}

// UpdateTemplateRemarkRequest 更新模板备注请求
type UpdateTemplateRemarkRequest struct {
	Remark string `json:"remark" binding:"max=500" example:"这是一个常用模板"` // 备注，最多500字符
}
