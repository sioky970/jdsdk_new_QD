package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"jd-task-platform-go/internal/constants"
	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// GetCurrentUser 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=models.User}
// @Failure 401 {object} response.Response
// @Router /users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, constants.MsgUserNotFound)
		return
	}

	userData := gin.H{
		"id":              user.ID,
		"username":        user.Username,
		"nickname":        user.Nickname,
		"avatar":          user.Avatar,
		"role":            user.Role,
		"jingdou_balance": user.JingdouBalance,
		"created_at":      user.CreatedAt.Format(time.RFC3339),
		"last_login":      nil,
	}

	if user.LastLogin != nil {
		userData["last_login"] = user.LastLogin.Format(time.RFC3339)
	}

	response.Success(c, userData)
}

// ChangePassword 修改密码
// @Summary 修改用户密码
// @Description 修改当前用户密码
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /users/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, constants.MsgParamError)
		return
	}

	userID, _ := c.Get("user_id")
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, constants.MsgUserNotFound)
		return
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		response.Error(c, http.StatusUnauthorized, constants.MsgPasswordOldIncorrect)
		return
	}

	// 验证新密码长度
	if len(req.NewPassword) < 6 {
		response.Error(c, http.StatusBadRequest, constants.MsgPasswordTooShort)
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.MsgPasswordEncryptFailed)
		return
	}

	user.PasswordHash = string(hashedPassword)
	h.db.Save(&user)

	response.SuccessWithMsg(c, constants.MsgPasswordChanged, nil)
}

// GenerateAPIKey 生成API密钥
// @Summary 生成或重置API密钥
// @Description 为当前用户生成新的API密钥
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=map[string]string}
// @Failure 500 {object} response.Response
// @Router /users/api-key [post]
func (h *UserHandler) GenerateAPIKey(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, constants.MsgUserNotFound)
		return
	}

	// 生成新的API密钥
	apiKey, err := generateRandomAPIKey()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, constants.MsgApiKeyGenerateFailed)
		return
	}

	user.ApiKey = apiKey
	h.db.Save(&user)

	response.Success(c, gin.H{"api_key": apiKey})
}

// GetUsers 获取用户列表（管理员）
// @Summary 获取用户列表
// @Description 获取所有用户列表（仅管理员）
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param per_page query int false "每页数量" default(10)
// @Success 200 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	query := h.db.Model(&models.User{})
	if search != "" {
		query = query.Where("username LIKE ? OR nickname LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var users []models.User
	offset := (page - 1) * perPage
	query.Offset(offset).Limit(perPage).Find(&users)

	items := make([]gin.H, 0)
	for _, user := range users {
		// 查询用户统计信息
		var stats struct {
			PendingTaskCount     int64 // 未完成任务数
			PendingExecuteCount  int   // 未完成执行次数
			PendingJingdou       int   // 未完成任务京豆合计
			TotalTaskCount       int64 // 总任务数
			TotalExecuteCount    int   // 总执行次数
			TotalExecutedCount   int   // 已完成执行次数
			CompletedTaskCount   int64 // 已完成任务数
			HistoryConsumedJingdou int // 历史消耗京豆
		}

		// 未完成任务统计 (waiting, running)
		h.db.Model(&models.Task{}).Where("user_id = ? AND status IN ?", user.ID, []string{"waiting", "running"}).
			Select("COUNT(*) as pending_task_count, COALESCE(SUM(execute_count - executed_count), 0) as pending_execute_count, COALESCE(SUM(consume_jingdou), 0) as pending_jingdou").
			Scan(&stats)

		// 总任务统计
		h.db.Model(&models.Task{}).Where("user_id = ?", user.ID).
			Select("COUNT(*) as total_task_count, COALESCE(SUM(execute_count), 0) as total_execute_count, COALESCE(SUM(executed_count), 0) as total_executed_count").
			Scan(&stats)

		// 已完成任务数
		h.db.Model(&models.Task{}).Where("user_id = ? AND status = ?", user.ID, "completed").Count(&stats.CompletedTaskCount)

		// 历史消耗京豆 (从京豆日志中统计扣除类型)
		h.db.Model(&models.JingdouLog{}).Where("user_id = ? AND amount < 0", user.ID).
			Select("COALESCE(SUM(ABS(amount)), 0)").Scan(&stats.HistoryConsumedJingdou)

		// 计算百分比
		var pendingTaskPercent float64 = 0
		var pendingExecutePercent float64 = 0
		if stats.TotalTaskCount > 0 {
			pendingTaskPercent = float64(stats.TotalTaskCount-stats.CompletedTaskCount) / float64(stats.TotalTaskCount) * 100
		}
		if stats.TotalExecuteCount > 0 {
			pendingExecutePercent = float64(stats.TotalExecuteCount-stats.TotalExecutedCount) / float64(stats.TotalExecuteCount) * 100
		}

		items = append(items, gin.H{
			"id":                      user.ID,
			"username":                user.Username,
			"nickname":                user.Nickname,
			"role":                    user.Role,
			"jingdou_balance":         user.JingdouBalance,
			"is_active":               user.IsActive,
			"created_at":              user.CreatedAt.Format(time.RFC3339),
			"pending_task_count":      stats.PendingTaskCount,
			"pending_execute_count":   stats.PendingExecuteCount,
			"pending_jingdou":         stats.PendingJingdou,
			"pending_task_percent":    pendingTaskPercent,
			"pending_execute_percent": pendingExecutePercent,
			"history_consumed_jingdou": stats.HistoryConsumedJingdou,
		})
	}

	pages := int((total + int64(perPage) - 1) / int64(perPage))

	response.Success(c, gin.H{
		"items":    items,
		"page":     page,
		"per_page": perPage,
		"total":    total,
		"pages":    pages,
	})
}

// GetUserByID 获取用户详情（管理员）
// @Summary 获取用户详情
// @Description 根据ID获取用户详细信息（仅管理员）
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=models.User}
// @Failure 404 {object} response.Response
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, constants.MsgUserNotFound)
		return
	}

	response.Success(c, user)
}

// generateRandomAPIKey 生成随机API密钥
func generateRandomAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
