package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex"`
	PasswordHash string
	Role         string
	Nickname     *string
}

func (User) TableName() string {
	return "users"
}

func main() {
	// 连接数据库
	dsn := "jduser:jdpass123@tcp(localhost:3306)/jd?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 查找admin用户
	var user User
	result := db.Where("username = ?", "admin").First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		// 用户不存在，创建新用户
		fmt.Println("admin用户不存在，正在创建...")

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("密码加密失败:", err)
		}

		nickname := "系统管理员"
		newUser := User{
			Username:     "admin",
			PasswordHash: string(hashedPassword),
			Role:         "admin",
			Nickname:     &nickname,
		}

		if err := db.Create(&newUser).Error; err != nil {
			log.Fatal("创建用户失败:", err)
		}

		fmt.Printf("✓ 管理员用户创建成功 (ID: %d)\n", newUser.ID)
	} else if result.Error != nil {
		log.Fatal("查询用户失败:", result.Error)
	} else {
		// 用户已存在，更新角色和密码
		fmt.Printf("找到用户: %s (ID: %d, 当前角色: %s)\n", user.Username, user.ID, user.Role)

		// 更新密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("密码加密失败:", err)
		}

		user.PasswordHash = string(hashedPassword)
		user.Role = "admin"

		if err := db.Save(&user).Error; err != nil {
			log.Fatal("更新用户失败:", err)
		}

		fmt.Printf("✓ 用户已更新为管理员角色，密码已重置为: admin123\n")
	}

	// 验证结果
	db.Where("username = ?", "admin").First(&user)
	fmt.Printf("\n最终用户信息:\n")
	fmt.Printf("  ID: %d\n", user.ID)
	fmt.Printf("  用户名: %s\n", user.Username)
	fmt.Printf("  角色: %s\n", user.Role)
	if user.Nickname != nil {
		fmt.Printf("  昵称: %s\n", *user.Nickname)
	}
	fmt.Println("\n✓ 管理员账户配置完成！")
}
