package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"jd-task-platform-go/internal/constants"
	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
	"jd-task-platform-go/pkg/utils"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户账号
// @Tags 认证模块
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "注册信息"
// @Success 200 {object} response.Response{data=map[string]interface{}} "注册成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, constants.MsgRegisterParamError)
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		response.Error(c, http.StatusBadRequest, constants.MsgUsernameExists)
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.MsgPasswordEncryptFailed)
		return
	}

	// 创建用户
	user := models.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		Nickname:     req.Nickname,
		Role:         "common",
		CreatedAt:    time.Now(),
	}

	if err := h.db.Create(&user).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, constants.MsgRegisterFailed)
		return
	}

	response.SuccessWithMsg(c, constants.MsgRegisterSuccess, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
	})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取访问令牌
// @Tags 认证模块
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=models.LoginResponse} "登录成功"
// @Failure 401 {object} response.Response "用户名或密码错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, constants.MsgLoginParamError)
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		response.Error(c, http.StatusUnauthorized, constants.MsgLoginFailed)
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.Error(c, http.StatusUnauthorized, constants.MsgLoginFailed)
		return
	}

	// 生成 Token
	accessToken, err := utils.GenerateToken(user.ID, user.Username, user.Role, time.Hour)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.MsgLoginSessionFailed)
		return
	}

	refreshToken, err := utils.GenerateToken(user.ID, user.Username, user.Role, 30*24*time.Hour)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.MsgLoginSessionFailed)
		return
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	h.db.Save(&user)

	loginResp := models.LoginResponse{
		ID:           user.ID,
		Username:     user.Username,
		Nickname:     user.Nickname,
		Role:         user.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expires:      time.Now().Add(time.Hour).Unix() * 1000,
	}

	response.SuccessWithMsg(c, constants.MsgLoginSuccess, loginResp)
}

// RefreshToken 刷新令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证模块
// @Accept json
// @Produce json
// @Param request body map[string]string true "刷新令牌" example({"refresh_token": "eyJhbGci..."})
// @Success 200 {object} response.Response{data=map[string]interface{}} "刷新成功"
// @Failure 401 {object} response.Response "令牌无效"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, constants.MsgTokenExpired)
		return
	}

	// 验证 refresh token
	claims, err := utils.ParseToken(req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, constants.MsgTokenExpired)
		return
	}

	// 生成新的 token
	accessToken, err := utils.GenerateToken(claims.UserID, claims.Username, claims.Role, time.Hour)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.MsgTokenGenFailed)
		return
	}

	newRefreshToken, err := utils.GenerateToken(claims.UserID, claims.Username, claims.Role, 30*24*time.Hour)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.MsgTokenGenFailed)
		return
	}

	response.SuccessWithMsg(c, constants.MsgTokenRefreshed, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"expires":       time.Now().Add(time.Hour).Unix() * 1000,
	})
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出（前端清除Token）
// @Tags 认证模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "登出成功"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Go版本为stateless JWT，登出由前端删除token实现
	response.SuccessWithMsg(c, constants.MsgLogoutSuccess, nil)
}

// GenerateAPIKey 生成随机API密钥
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
