package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
)

// CreateUser 创建用户（管理员）
// @Summary 创建用户
// @Description 创建新用户（仅管理员）
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{username=string,password=string,nickname=string,role=string,jingdou_balance=int} true "用户信息"
// @Success 201 {object} response.Response{data=object}
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username       string `json:"username" binding:"required"`
		Password       string `json:"password" binding:"required,min=6"`
		Nickname       string `json:"nickname"`
		Role           string `json:"role"`
		JingdouBalance int    `json:"jingdou_balance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("创建用户参数绑定失败: %v", err)
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 检查用户名是否已存在
	var existing models.User
	if err := h.db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		response.Error(c, http.StatusBadRequest, "用户名已存在")
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("密码加密失败: %v", err)
		response.Error(c, http.StatusInternalServerError, "密码加密失败: "+err.Error())
		return
	}

	role := req.Role
	if role == "" {
		role = "common"
	}

	user := models.User{
		Username:       req.Username,
		PasswordHash:   string(hashedPassword),
		Nickname:       req.Nickname,
		Role:           role,
		IsActive:       true,
		JingdouBalance: req.JingdouBalance,
		CreatedAt:      time.Now(),
		// ApiKey 不设置，让数据库默认为 NULL，避免唯一索引冲突
	}

	// 使用 Omit 跳过 ApiKey 字段，让其为 NULL
	if err := h.db.Omit("ApiKey", "ApiKeyCreatedAt", "ApiKeyLastUsedAt").Create(&user).Error; err != nil {
		log.Printf("数据库创建用户失败 [username=%s]: %v", req.Username, err)
		response.Error(c, http.StatusInternalServerError, "创建用户失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

// UpdateUser 更新用户信息（管理员）
// @Summary 更新用户信息
// @Description 更新用户信息（仅管理员）
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body object{nickname=string,role=string,is_active=bool,jingdou_balance=int,password=string} true "用户信息"
// @Success 200 {object} response.Response
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Nickname       *string `json:"nickname"`
		Role           *string `json:"role"`
		IsActive       *bool   `json:"is_active"`
		JingdouBalance *int    `json:"jingdou_balance"`
		Password       *string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.JingdouBalance != nil {
		user.JingdouBalance = *req.JingdouBalance
	}
	if req.Password != nil && len(*req.Password) >= 6 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err == nil {
			user.PasswordHash = string(hashedPassword)
		}
	}

	if err := h.db.Save(&user).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "更新用户信息失败")
		return
	}

	response.SuccessWithMsg(c, "用户信息更新成功", nil)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除用户（仅管理员）
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	if err := h.db.Delete(&user).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "删除用户失败")
		return
	}

	response.SuccessWithMsg(c, "用户删除成功", nil)
}

// AdjustJingdou 调整用户京豆余额
// @Summary 调整用户京豆余额
// @Description 充值或扣除用户京豆（仅管理员）
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body object{amount=int,operation_type=string,remark=string} true "调整信息"
// @Success 200 {object} response.Response{data=object}
// @Router /users/{id}/jingdou [post]
func (h *UserHandler) AdjustJingdou(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Amount        int    `json:"amount" binding:"required"`
		OperationType string `json:"operation_type"` // recharge, deduct
		Remark        string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	tx := h.db.Begin()

	// 更新余额
	user.JingdouBalance += req.Amount
	if user.JingdouBalance < 0 {
		tx.Rollback()
		response.Error(c, http.StatusBadRequest, "京豆余额不足")
		return
	}

	tx.Save(&user)

	// 记录日志
	opType := req.OperationType
	if opType == "" {
		if req.Amount > 0 {
			opType = "recharge"
		} else {
			opType = "deduct"
		}
	}

	log := models.JingdouLog{
		UserID:        user.ID,
		Amount:        req.Amount,
		Balance:       user.JingdouBalance,
		OperationType: opType,
		Remark:        req.Remark,
		CreatedAt:     time.Now(),
	}
	tx.Create(&log)

	tx.Commit()

	response.Success(c, gin.H{
		"balance": user.JingdouBalance,
		"amount":  req.Amount,
	})
}

// SearchUsers 搜索用户
// @Summary 搜索用户
// @Description 按用户名搜索用户（仅管理员）
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Response{data=object}
// @Router /users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")

	query := h.db.Model(&models.User{})
	if keyword != "" {
		query = query.Where("username LIKE ?", "%"+keyword+"%")
	}

	var users []models.User
	query.Limit(20).Find(&users)

	items := make([]gin.H, 0)
	for _, user := range users {
		items = append(items, gin.H{
			"id":              user.ID,
			"username":        user.Username,
			"role":            user.Role,
			"jingdou_balance": user.JingdouBalance,
		})
	}

	response.Success(c, gin.H{"users": items})
}

// UpdateProfile 修改个人资料
// @Summary 修改个人资料
// @Description 修改当前用户的个人资料
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{nickname=string,avatar=string} true "个人资料"
// @Success 200 {object} response.Response
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req struct {
		Nickname *string `json:"nickname"`
		Avatar   *string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}

	h.db.Save(&user)

	response.SuccessWithMsg(c, "个人资料更新成功", nil)
}

// GetRechargeStatistics 充值统计
// @Summary 充值统计
// @Description 获取用户充值统计信息（仅管理员）
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "用户ID"
// @Success 200 {object} response.Response{data=object}
// @Router /users/recharge-statistics [get]
func (h *UserHandler) GetRechargeStatistics(c *gin.Context) {
	userIDParam := c.Query("user_id")

	query := h.db.Model(&models.JingdouLog{}).Where("operation_type = ?", "recharge")

	if userIDParam != "" {
		if uid, err := strconv.Atoi(userIDParam); err == nil {
			query = query.Where("user_id = ?", uid)
		}
	}

	var totalAmount int
	var count int64
	query.Select("SUM(amount)").Row().Scan(&totalAmount)
	query.Count(&count)

	response.Success(c, gin.H{
		"total_amount": totalAmount,
		"count":        count,
	})
}

// GetUserApiKey 获取用户API Key（管理员）
// @Summary 获取用户API Key
// @Description 获取指定用户的API Key信息（仅管理员）
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=object}
// @Router /users/{id}/apikey [get]
func (h *UserHandler) GetUserApiKey(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	data := gin.H{
		"api_key":      user.ApiKey,
		"created_at":   nil,
		"last_used_at": nil,
	}

	if user.ApiKeyCreatedAt != nil {
		data["created_at"] = user.ApiKeyCreatedAt.Format(time.RFC3339)
	}
	if user.ApiKeyLastUsedAt != nil {
		data["last_used_at"] = user.ApiKeyLastUsedAt.Format(time.RFC3339)
	}

	response.Success(c, data)
}

// ResetUserApiKey 重置用户API Key（管理员）
// @Summary 重置用户API Key
// @Description 为指定用户重新生成API Key（仅管理员）
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=object}
// @Router /users/{id}/apikey [post]
func (h *UserHandler) ResetUserApiKey(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 生成新的API Key
	apiKey, err := generateRandomAPIKey()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "API Key生成失败")
		return
	}

	now := time.Now()
	user.ApiKey = apiKey
	user.ApiKeyCreatedAt = &now
	user.ApiKeyLastUsedAt = nil

	h.db.Save(&user)

	response.SuccessWithMsg(c, "API Key重置成功", gin.H{
		"api_key":    apiKey,
		"created_at": now.Format(time.RFC3339),
	})
}

// DeleteUserApiKey 删除用户API Key（管理员）
// @Summary 删除用户API Key
// @Description 删除指定用户的API Key（仅管理员）
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /users/{id}/apikey [delete]
func (h *UserHandler) DeleteUserApiKey(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	user.ApiKey = ""
	user.ApiKeyCreatedAt = nil
	user.ApiKeyLastUsedAt = nil

	h.db.Save(&user)

	response.SuccessWithMsg(c, "API Key已删除", nil)
}
