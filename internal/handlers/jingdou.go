package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
)

type JingdouHandler struct {
	db *gorm.DB
}

func NewJingdouHandler(db *gorm.DB) *JingdouHandler {
	return &JingdouHandler{db: db}
}

// GetJingdouLogs 获取京豆变动记录
// @Summary 获取京豆变动记录
// @Description 获取当前用户的京豆变动记录，支持分页和筛选
// @Tags 京豆模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param operation_type query string false "操作类型" Enums(task, recharge, refund, deduct)
// @Param start_date query string false "开始日期" format(date)
// @Param end_date query string false "结束日期" format(date)
// @Success 200 {object} response.Response{data=object}
// @Router /jingdou/logs [get]
func (h *JingdouHandler) GetJingdouLogs(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

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

	query := h.db.Model(&models.JingdouLog{})

	// 管理员可查看所有，普通用户只能查看自己的
	targetUserID := c.Query("user_id")
	if role == "admin" && targetUserID != "" {
		query = query.Where("user_id = ?", targetUserID)
	} else if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	// 筛选操作类型
	if opType := c.Query("operation_type"); opType != "" {
		query = query.Where("operation_type = ?", opType)
	}

	// 日期范围筛选
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	var total int64
	query.Count(&total)

	var logs []models.JingdouLog
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)

	items := make([]gin.H, 0)
	for _, log := range logs {
		items = append(items, gin.H{
			"id":             log.ID,
			"user_id":        log.UserID,
			"amount":         log.Amount,
			"balance":        log.Balance,
			"operation_type": log.OperationType,
			"related_id":     log.RelatedID,
			"remark":         log.Remark,
			"created_at":     log.CreatedAt.Format(time.RFC3339),
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

// GetJingdouBalance 获取京豆余额（JWT认证）
// @Summary 获取京豆余额
// @Description 获取当前用户的京豆余额
// @Tags 京豆模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Router /jingdou/balance [get]
func (h *JingdouHandler) GetJingdouBalance(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	response.Success(c, gin.H{
		"jingdou_balance": user.JingdouBalance,
	})
}

// GetJingdouRecords 获取京豆变动记录（前端用）
// @Summary 获取京豆变动记录
// @Description 获取当前用户的京豆变动记录，支持分页和筛选
// @Tags 京豆模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param per_page query int false "每页数量" default(20)
// @Param type query string false "类型" Enums(task_consume, task_refund, recharge, withdraw, admin_adjust)
// @Param start_date query string false "开始日期" format(date)
// @Param end_date query string false "结束日期" format(date)
// @Success 200 {object} response.Response{data=object}
// @Router /jingdou/records [get]
func (h *JingdouHandler) GetJingdouRecords(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	page := 1
	perPage := 20
	if p, ok := c.GetQuery("page"); ok {
		if pInt, err := strconv.Atoi(p); err == nil && pInt > 0 {
			page = pInt
		}
	}
	if ps, ok := c.GetQuery("per_page"); ok {
		if psInt, err := strconv.Atoi(ps); err == nil && psInt > 0 {
			perPage = psInt
		}
	}

	query := h.db.Model(&models.JingdouLog{})

	// 管理员可查看所有，普通用户只能查看自己的
	targetUserID := c.Query("user_id")
	if role == "admin" && targetUserID != "" {
		query = query.Where("user_id = ?", targetUserID)
	} else if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	// 筛选类型（将前端类型映射到数据库字段）
	if recordType := c.Query("type"); recordType != "" {
		// 前端类型映射：
		// task_consume -> task, consume (任务消耗)
		// task_refund -> refund (任务退还)
		// recharge -> recharge (充值)
		// withdraw -> withdraw (提现)
		// admin_adjust -> admin (管理员调整)
		typeMapping := map[string][]string{
			"task_consume": {"task", "consume"},
			"task_refund":  {"refund"},
			"recharge":     {"recharge"},
			"withdraw":     {"withdraw"},
			"admin_adjust": {"admin"},
		}
		if dbTypes, ok := typeMapping[recordType]; ok {
			if len(dbTypes) == 1 {
				query = query.Where("operation_type = ?", dbTypes[0])
			} else {
				query = query.Where("operation_type IN ?", dbTypes)
			}
		} else {
			query = query.Where("operation_type = ?", recordType)
		}
	}

	// 日期范围筛选
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	var total int64
	query.Count(&total)

	var logs []models.JingdouLog
	query.Order("created_at DESC").Offset((page - 1) * perPage).Limit(perPage).Find(&logs)

	// 管理员查看时，获取用户信息
	userMap := make(map[uint]models.User)
	if role == "admin" {
		var userIDs []uint
		for _, log := range logs {
			userIDs = append(userIDs, log.UserID)
		}
		if len(userIDs) > 0 {
			var users []models.User
			h.db.Where("id IN ?", userIDs).Find(&users)
			for _, u := range users {
				userMap[u.ID] = u
			}
		}
	}

	// 将数据库类型映射回前端类型
	typeReverseMapping := map[string]string{
		"task":     "task_consume",
		"consume":  "task_consume", // consume 也是任务消耗
		"refund":   "task_refund",
		"recharge": "recharge",
		"withdraw": "withdraw",
		"admin":    "admin_adjust",
	}

	records := make([]gin.H, 0)
	for _, log := range logs {
		recordType := log.OperationType
		if mappedType, ok := typeReverseMapping[log.OperationType]; ok {
			recordType = mappedType
		}

		record := gin.H{
			"id":            log.ID,
			"user_id":       log.UserID,
			"type":          recordType, // 前端使用 type 字段
			"amount":        log.Amount,
			"balance_after": log.Balance,   // 前端使用 balance_after 字段
			"task_id":       log.RelatedID, // 前端使用 task_id 字段
			"remark":        log.Remark,
			"created_at":    log.CreatedAt.Format(time.RFC3339),
		}

		// 管理员查看时添加用户信息
		if role == "admin" {
			if user, ok := userMap[log.UserID]; ok {
				record["username"] = user.Username
				if user.Nickname != "" {
					record["nickname"] = user.Nickname
				} else {
					record["nickname"] = user.Username
				}
			}
		}

		records = append(records, record)
	}

	response.Success(c, gin.H{
		"records":  records,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// GetJingdouBalanceByAPIKey 获取京豆余额（API Key认证）
// @Summary 获取京豆余额（API Key）
// @Description 使用API Key获取京豆余额
// @Tags 京豆模块
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=object}
// @Router /jingdou/balance/apikey [get]
func (h *JingdouHandler) GetJingdouBalanceByAPIKey(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 获取最新的京豆日志
	var lastLog models.JingdouLog
	h.db.Where("user_id = ?", userID).Order("created_at DESC").First(&lastLog)

	response.Success(c, gin.H{
		"user_id":         user.ID,
		"username":        username,
		"jingdou_balance": user.JingdouBalance,
		"last_updated":    lastLog.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// GetJingdouStatistics 获取京豆统计信息（管理员）
// @Summary 获取京豆统计信息
// @Description 获取京豆统计信息，包括消耗、充值等（仅管理员）
// @Tags 京豆模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "用户ID"
// @Success 200 {object} response.Response{data=object}
// @Router /jingdou/statistics [get]
func (h *JingdouHandler) GetJingdouStatistics(c *gin.Context) {
	targetUserID := c.Query("user_id")

	query := h.db.Model(&models.JingdouLog{})
	if targetUserID != "" {
		query = query.Where("user_id = ?", targetUserID)
	}

	// 总消耗和总充值
	var totalConsumed, totalRecharged int
	h.db.Model(&models.JingdouLog{}).Where("operation_type IN (?) AND amount < 0", []string{"task", "deduct"}).Select("SUM(ABS(amount))").Row().Scan(&totalConsumed)
	h.db.Model(&models.JingdouLog{}).Where("operation_type IN (?)", []string{"recharge", "refund"}).Select("SUM(amount)").Row().Scan(&totalRecharged)

	// 今日统计
	today := time.Now().Format("2006-01-02")
	var todayConsumed, todayRecharged int
	h.db.Model(&models.JingdouLog{}).Where("operation_type IN (?) AND amount < 0 AND DATE(created_at) = ?", []string{"task", "deduct"}, today).Select("SUM(ABS(amount))").Row().Scan(&todayConsumed)
	h.db.Model(&models.JingdouLog{}).Where("operation_type IN (?) AND DATE(created_at) = ?", []string{"recharge", "refund"}, today).Select("SUM(amount)").Row().Scan(&todayRecharged)

	// 按操作类型统计
	operationStats := make(map[string]gin.H)
	operations := []string{"task", "recharge", "refund", "deduct"}
	for _, op := range operations {
		var amount int
		var count int64
		h.db.Model(&models.JingdouLog{}).Where("operation_type = ?", op).Select("SUM(ABS(amount))").Row().Scan(&amount)
		h.db.Model(&models.JingdouLog{}).Where("operation_type = ?", op).Count(&count)

		operationStats[op] = gin.H{
			"total_amount": amount,
			"count":        count,
		}
	}

	response.Success(c, gin.H{
		"total_consumed":  totalConsumed,
		"total_recharged": totalRecharged,
		"today_consumed":  todayConsumed,
		"today_recharged": todayRecharged,
		"operation_stats": operationStats,
	})
}
