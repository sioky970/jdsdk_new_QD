package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
)

type APIKeyHandler struct {
	db *gorm.DB
}

func NewAPIKeyHandler(db *gorm.DB) *APIKeyHandler {
	return &APIKeyHandler{db: db}
}

// generateAPIKey 生成API密钥
func generateAPIKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "sk_" + hex.EncodeToString(b)
}

// GetAPIKey 获取API密钥（JWT认证）
// @Summary 获取API密钥
// @Description 获取当前用户的API密钥信息，包括创建时间和最后使用时间
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Router /apikey [get]
func (h *APIKeyHandler) GetAPIKey(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 如果没有API Key，返回空数据
	if user.ApiKey == "" {
		response.Success(c, gin.H{
			"api_key":      nil,
			"created_at":   nil,
			"last_used_at": nil,
		})
		return
	}

	// 返回 API Key 和时间戳
	data := gin.H{
		"api_key": user.ApiKey,
	}

	if user.ApiKeyCreatedAt != nil {
		data["created_at"] = user.ApiKeyCreatedAt.Format(time.RFC3339)
	} else {
		data["created_at"] = nil
	}

	if user.ApiKeyLastUsedAt != nil {
		data["last_used_at"] = user.ApiKeyLastUsedAt.Format(time.RFC3339)
	} else {
		data["last_used_at"] = nil
	}

	response.Success(c, data)
}

// GenerateAPIKey 生成API密钥（JWT认证）
// @Summary 生成API密钥
// @Description 为用户生成新的API密钥，旧密钥将失效
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Router /apikey/generate [post]
func (h *APIKeyHandler) GenerateAPIKey(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 生成新的API密钥
	now := time.Now()
	user.ApiKey = generateAPIKey()
	user.ApiKeyCreatedAt = &now
	user.ApiKeyLastUsedAt = nil // 重置最后使用时间

	if err := h.db.Save(&user).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "生成API密钥失败")
		return
	}

	response.Success(c, gin.H{
		"api_key":      user.ApiKey,
		"created_at":   now.Format(time.RFC3339),
		"last_used_at": nil,
	})
}

// ResetAPIKey 重置API密钥（JWT认证）
// @Summary 重置API密钥
// @Description 重新生成用户的API密钥，旧密钥将失效
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Router /apikey/reset [post]
func (h *APIKeyHandler) ResetAPIKey(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 重置API密钥
	now := time.Now()
	user.ApiKey = generateAPIKey()
	user.ApiKeyCreatedAt = &now
	user.ApiKeyLastUsedAt = nil // 重置最后使用时间

	if err := h.db.Save(&user).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "重置API密钥失败")
		return
	}

	response.Success(c, gin.H{
		"api_key":      user.ApiKey,
		"created_at":   now.Format(time.RFC3339),
		"last_used_at": nil,
	})
}

// DeleteAPIKey 删除API密钥（JWT认证）
// @Summary 删除API密钥
// @Description 删除用户的API密钥
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Router /apikey [delete]
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 删除API密钥
	user.ApiKey = ""
	user.ApiKeyCreatedAt = nil
	user.ApiKeyLastUsedAt = nil

	if err := h.db.Save(&user).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "删除API密钥失败")
		return
	}

	response.SuccessWithMsg(c, "API密钥已删除", nil)
}

// GetAPILogs 获取API调用记录（JWT认证）
// @Summary 获取API调用记录
// @Description 获取当前用户的API调用记录
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param endpoint query string false "接口端点"
// @Param status query string false "状态" Enums(success,failed)
// @Param start_date query string false "开始日期" format(date)
// @Param end_date query string false "结束日期" format(date)
// @Success 200 {object} response.Response{data=object}
// @Router /apikey/logs [get]
func (h *APIKeyHandler) GetAPILogs(c *gin.Context) {
	userID, _ := c.Get("user_id")

	page := 1
	pageSize := 20
	if p, ok := c.GetQuery("page"); ok {
		if pInt, err := strconv.Atoi(p); err == nil && pInt > 0 {
			page = pInt
		}
	}
	if ps, ok := c.GetQuery("page_size"); ok {
		if psInt, err := strconv.Atoi(ps); err == nil && psInt > 0 {
			pageSize = psInt
		}
	}

	query := h.db.Model(&models.APILog{}).Where("user_id = ?", userID)

	// 筛选条件
	if endpoint := c.Query("endpoint"); endpoint != "" {
		query = query.Where("endpoint LIKE ?", "%"+endpoint+"%")
	}

	if status := c.Query("status"); status != "" {
		if status == "success" {
			query = query.Where("response_code >= 200 AND response_code < 300")
		} else if status == "failed" {
			query = query.Where("response_code >= 400")
		}
	}

	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	var total int64
	query.Count(&total)

	var logs []models.APILog
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)

	items := make([]gin.H, 0)
	for _, log := range logs {
		status := "success"
		if log.ResponseCode >= 400 {
			status = "failed"
		}

		items = append(items, gin.H{
			"id":            log.ID,
			"endpoint":      log.Endpoint,
			"method":        log.Method,
			"ip_address":    log.IP,
			"status":        status,
			"response_code": log.ResponseCode,
			"details":       "",
			"created_at":    log.CreatedAt.Format(time.RFC3339),
		})
	}

	response.Success(c, gin.H{
		"items":    items,
		"page":     page,
		"per_page": pageSize,
		"total":    total,
		"pages":    (int(total) + pageSize - 1) / pageSize,
	})
}

// GetAPILogsByAPIKey 获取API调用记录（API Key认证）
// @Summary 获取API调用记录（API Key）
// @Description 使用API Key获取API调用记录
// @Tags API日志
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=object}
// @Router /logs/apikey [get]
func (h *APIKeyHandler) GetAPILogsByAPIKey(c *gin.Context) {
	apiKey, _ := c.Get("api_key")

	page := 1
	pageSize := 20
	if p, ok := c.GetQuery("page"); ok {
		if pInt, err := strconv.Atoi(p); err == nil && pInt > 0 {
			page = pInt
		}
	}
	if ps, ok := c.GetQuery("page_size"); ok {
		if psInt, err := strconv.Atoi(ps); err == nil && psInt > 0 {
			pageSize = psInt
		}
	}

	query := h.db.Model(&models.APILog{}).Where("api_key = ?", apiKey)

	var total int64
	query.Count(&total)

	var logs []models.APILog
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)

	items := make([]gin.H, 0)
	for _, log := range logs {
		status := "success"
		if log.ResponseCode >= 400 {
			status = "failed"
		}

		items = append(items, gin.H{
			"id":            log.ID,
			"endpoint":      log.Endpoint,
			"method":        log.Method,
			"ip_address":    log.IP,
			"status":        status,
			"response_code": log.ResponseCode,
			"created_at":    log.CreatedAt.Format(time.RFC3339),
		})
	}

	response.Success(c, gin.H{
		"items":    items,
		"page":     page,
		"per_page": pageSize,
		"total":    total,
		"pages":    (int(total) + pageSize - 1) / pageSize,
	})
}
