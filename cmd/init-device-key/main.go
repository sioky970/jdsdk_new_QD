package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DeviceSetting struct {
	ID          uint      `gorm:"primaryKey"`
	ParamKey    string    `gorm:"column:param_key;uniqueIndex;not null"`
	ParamValue  string    `gorm:"column:param_value;type:text"`
	ParamType   string    `gorm:"column:param_type;type:varchar(20)"`
	Description string    `gorm:"column:description;type:varchar(200)"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (DeviceSetting) TableName() string {
	return "settings"
}

func initDeviceKey() {
	// 数据库连接配置
	dsn := "jduser:jdpass123@tcp(127.0.0.1:3306)/jd?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("  插入设备认证密钥配置")
	fmt.Println("========================================")
	fmt.Println()

	// 检查是否已存在
	var existingSetting DeviceSetting
	result := db.Where("param_key = ?", "device_auth_key").First(&existingSetting)

	if result.Error == nil {
		// 已存在，更新
		fmt.Println("✅ 设备密钥配置已存在，更新中...")
		existingSetting.ParamValue = "KKNN778899"
		existingSetting.UpdatedAt = time.Now()

		if err := db.Save(&existingSetting).Error; err != nil {
			log.Fatalf("❌ 更新失败: %v", err)
		}

		fmt.Println("✅ 更新成功！")
	} else {
		// 不存在，插入
		fmt.Println("⚠️  设备密钥配置不存在，插入中...")
		newSetting := DeviceSetting{
			ParamKey:    "device_auth_key",
			ParamValue:  "KKNN778899",
			ParamType:   "string",
			Description: "设备认证密钥（用于设备端API认证）",
			UpdatedAt:   time.Now(),
		}

		if err := db.Create(&newSetting).Error; err != nil {
			log.Fatalf("❌ 插入失败: %v", err)
		}

		fmt.Println("✅ 插入成功！")
	}

	// 验证
	var setting DeviceSetting
	db.Where("param_key = ?", "device_auth_key").First(&setting)

	fmt.Println()
	fmt.Println("当前配置:")
	fmt.Printf("  - 参数名: %s\n", setting.ParamKey)
	fmt.Printf("  - 参数值: %s\n", setting.ParamValue)
	fmt.Printf("  - 参数类型: %s\n", setting.ParamType)
	fmt.Printf("  - 描述: %s\n", setting.Description)
	fmt.Printf("  - 更新时间: %s\n", setting.UpdatedAt.Format("2006-01-02 15:04:05"))

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("✅ 完成！")
	fmt.Println("========================================")
}

func main() {
	initDeviceKey()
}
