package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Username         string     `gorm:"uniqueIndex;not null;size:64" json:"username"`
	PasswordHash     string     `gorm:"not null;size:128" json:"-"`
	Nickname         string     `gorm:"size:64" json:"nickname"`
	Avatar           string     `gorm:"size:255" json:"avatar"`
	Role             string     `gorm:"size:20;default:common" json:"role"`
	ApiKey           string     `gorm:"uniqueIndex;size:64;column:api_key" json:"api_key,omitempty"`
	ApiKeyCreatedAt  *time.Time `gorm:"column:api_key_created_at" json:"api_key_created_at,omitempty"`
	ApiKeyLastUsedAt *time.Time `gorm:"column:api_key_last_used_at" json:"api_key_last_used_at,omitempty"`
	JingdouBalance   int        `gorm:"default:0;column:jingdou_balance" json:"jingdou_balance"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"created_at"`
	LastLogin        *time.Time `gorm:"column:last_login" json:"last_login"`
	IsActive         bool       `gorm:"default:true;column:is_active" json:"is_active"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"testuser"`
	Password string `json:"password" binding:"required" example:"password123"`
	Nickname string `json:"nickname" example:"测试用户"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"testuser"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	ID           uint   `json:"id" example:"1"`
	Username     string `json:"username" example:"testuser"`
	Nickname     string `json:"nickname" example:"测试用户"`
	Role         string `json:"role" example:"common"`
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	Expires      int64  `json:"expires" example:"1625097600000"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpass123"`
	NewPassword string `json:"new_password" binding:"required" example:"newpass123"`
}
