package main

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Setting 系统设置模型
type Setting struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	ParamKey    string `gorm:"uniqueIndex;size:64;not null;column:param_key" json:"param_key"`
	ParamValue  string `gorm:"type:text;column:param_value" json:"param_value"`
	ParamType   string `gorm:"size:20;column:param_type" json:"param_type"`
	Description string `gorm:"size:255" json:"description"`
}

// TableName 指定表名
func (Setting) TableName() string {
	return "settings"
}

// TaskType 任务类型模型
type TaskType struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	TypeCode     string `gorm:"uniqueIndex;size:32;not null;column:type_code" json:"type_code"`
	TypeName     string `gorm:"size:64;not null;column:type_name" json:"type_name"`
	JingdouPrice int    `gorm:"not null;column:jingdou_price" json:"jingdou_price"`
	IsActive     bool   `gorm:"default:true;column:is_active" json:"is_active"`
}

// TableName 指定表名
func (TaskType) TableName() string {
	return "task_types"
}

// JingdouLog 京豆日志模型
type JingdouLog struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	UserID        uint   `gorm:"not null;column:user_id" json:"user_id"`
	Amount        int    `gorm:"not null" json:"amount"`
	Balance       int    `gorm:"not null" json:"balance"`
	OperationType string `gorm:"size:20;not null;column:operation_type" json:"operation_type"`
	RelatedID     *uint  `gorm:"column:related_id" json:"related_id"`
	Remark        string `gorm:"size:255" json:"remark"`
}

// TableName 指定表名
func (JingdouLog) TableName() string {
	return "jingdou_logs"
}

// APILog API调用日志模型
type APILog struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	UserID       *uint   `gorm:"column:user_id" json:"user_id"`
	ApiKey       string  `gorm:"size:64;column:api_key" json:"api_key"`
	Endpoint     string  `gorm:"size:200;not null" json:"endpoint"`
	Method       string  `gorm:"size:10;not null" json:"method"`
	IP           string  `gorm:"size:45;column:ip" json:"ip"`
	UserAgent    string  `gorm:"type:text;column:user_agent" json:"user_agent"`
	ResponseCode int     `gorm:"not null;column:response_code" json:"response_code"`
	ResponseTime float64 `gorm:"column:response_time" json:"response_time"`
}

// TableName 指定表名
func (APILog) TableName() string {
	return "api_logs"
}

// TaskLog 任务日志模型
type TaskLog struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	TaskID   uint   `gorm:"not null;column:task_id" json:"task_id"`
	DeviceID string `gorm:"size:64;column:device_id" json:"device_id"`
	Status   string `gorm:"size:20;not null" json:"status"`
	Message  string `gorm:"type:text" json:"message"`
}

// TableName 指定表名
func (TaskLog) TableName() string {
	return "task_logs"
}

// DeviceTaskHistory 设备任务历史
type DeviceTaskHistory struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	DeviceID string `gorm:"size:64;not null;column:device_id" json:"device_id"`
	TaskID   uint   `gorm:"not null;column:task_id" json:"task_id"`
	SKU      string `gorm:"size:64;not null" json:"sku"`
	Status   string `gorm:"size:20;not null" json:"status"`
}

// TableName 指定表名
func (DeviceTaskHistory) TableName() string {
	return "device_task_history"
}

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
		&Setting{},
		&TaskType{},
		&JingdouLog{},
		&APILog{},
		&TaskLog{},
		&DeviceTaskHistory{},
	}

	log.Println("\n========== 开始数据库表迁移 ==========")

	// 检查并创建每个表
	for _, model := range tables {
		tableName := db.NamingStrategy.TableName(db.Statement.Table)

		// 执行迁移
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("❌ 表迁移失败: %v", err)
			continue
		}

		// 获取表名
		stmt := &gorm.Statement{DB: db}
		stmt.Parse(model)
		tableName = stmt.Schema.Table

		log.Printf("✓ 表已就绪: %s", tableName)
	}

	log.Println("\n========== 数据库表迁移完成 ==========")
	log.Println("\n验证表结构...")

	// 验证settings表
	var count int64
	db.Table("settings").Count(&count)
	log.Printf("✓ settings 表记录数: %d", count)

	db.Table("task_types").Count(&count)
	log.Printf("✓ task_types 表记录数: %d", count)

	db.Table("jingdou_logs").Count(&count)
	log.Printf("✓ jingdou_logs 表记录数: %d", count)

	db.Table("api_logs").Count(&count)
	log.Printf("✓ api_logs 表记录数: %d", count)

	log.Println("\n✓ 所有表已成功创建并验证完成！")
}
