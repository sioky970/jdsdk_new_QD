package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
)

// UserHomeHandler 用户首页处理器
type UserHomeHandler struct {
	db *gorm.DB
}

// NewUserHomeHandler 创建用户首页处理器
func NewUserHomeHandler(db *gorm.DB) *UserHomeHandler {
	return &UserHomeHandler{db: db}
}

// GetUserTodayStats 获取用户今日任务统计
// @Summary 获取用户今日任务统计
// @Description 获取当前用户今日任务统计信息，包括总数、已完成、待完成等
// @Tags 用户首页
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /user/home/today-stats [get]
func (h *UserHomeHandler) GetUserTodayStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 今天的开始和结束时间
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	todayEnd := todayStart.Add(24 * time.Hour)

	// 统计用户今日任务（排除已取消的任务）
	// 注意：使用 start_time 筛选今日执行的任务，而不是 created_at（创建时间）
	var totalTasks int64
	var completedTasks int64
	var runningTasks int64
	var waitingTasks int64
	var failedTasks int64
	var partialCompletedTasks int64

	// 统计执行次数（execute_count总和，executed_count总和）
	var totalExecuteCount int64
	var totalExecutedCount int64

	// 统计今日执行的任务（按 start_time 筛选，排除已取消cancelled状态）
	h.db.Model(&models.Task{}).Where("user_id = ? AND start_time >= ? AND start_time < ? AND status != ?", userID, todayStart, todayEnd, "cancelled").Count(&totalTasks)
	h.db.Model(&models.Task{}).Where("user_id = ? AND start_time >= ? AND start_time < ? AND status = ?", userID, todayStart, todayEnd, "completed").Count(&completedTasks)
	h.db.Model(&models.Task{}).Where("user_id = ? AND start_time >= ? AND start_time < ? AND status = ?", userID, todayStart, todayEnd, "running").Count(&runningTasks)
	h.db.Model(&models.Task{}).Where("user_id = ? AND start_time >= ? AND start_time < ? AND status = ?", userID, todayStart, todayEnd, "waiting").Count(&waitingTasks)
	h.db.Model(&models.Task{}).Where("user_id = ? AND start_time >= ? AND start_time < ? AND status = ?", userID, todayStart, todayEnd, "failed").Count(&failedTasks)
	h.db.Model(&models.Task{}).Where("user_id = ? AND start_time >= ? AND start_time < ? AND status = ?", userID, todayStart, todayEnd, "partial_completed").Count(&partialCompletedTasks)

	pendingTasks := waitingTasks + runningTasks

	// 统计今日总执行次数和已执行次数（按 start_time 筛选，排除已取消的任务）
	h.db.Model(&models.Task{}).Where("user_id = ? AND start_time >= ? AND start_time < ? AND status != ?", userID, todayStart, todayEnd, "cancelled").
		Select("COALESCE(SUM(execute_count), 0)").Scan(&totalExecuteCount)
	h.db.Model(&models.Task{}).Where("user_id = ? AND start_time >= ? AND start_time < ? AND status != ?", userID, todayStart, todayEnd, "cancelled").
		Select("COALESCE(SUM(executed_count), 0)").Scan(&totalExecutedCount)

	// 计算执行次数完成率
	executeCompletionRate := 0.0
	if totalExecuteCount > 0 {
		executeCompletionRate = float64(totalExecutedCount) / float64(totalExecuteCount) * 100
	}

	// 计算完成率
	completionRate := 0.0
	if totalTasks > 0 {
		completionRate = float64(completedTasks+partialCompletedTasks) / float64(totalTasks) * 100
	}

	// 统计今日消耗京豆（按 start_time 筛选，排除已取消的任务，因为取消会退还京豆）
	var todayConsumedJingdou int64
	h.db.Model(&models.Task{}).Where("user_id = ? AND start_time >= ? AND start_time < ? AND status != ?", userID, todayStart, todayEnd, "cancelled").
		Select("COALESCE(SUM(consume_jingdou), 0)").Scan(&todayConsumedJingdou)

	// 获取用户当前京豆余额
	var user models.User
	h.db.First(&user, userID)

	response.Success(c, gin.H{
		"today_total":             totalTasks,
		"today_completed":         completedTasks,
		"today_running":           runningTasks,
		"today_waiting":           waitingTasks,
		"today_pending":           pendingTasks,
		"today_failed":            failedTasks,
		"today_partial_completed": partialCompletedTasks,
		"completion_rate":         fmt.Sprintf("%.1f", completionRate),
		"today_consumed_jingdou":  todayConsumedJingdou,
		"jingdou_balance":         user.JingdouBalance,
		"total_execute_count":     totalExecuteCount,                          // 今日总执行次数
		"total_executed_count":    totalExecutedCount,                         // 今日已执行次数
		"execute_completion_rate": fmt.Sprintf("%.1f", executeCompletionRate), // 执行完成率
	})
}

// GetTaskTemplates 获取用户任务模板列表
// @Summary 获取用户任务模板列表
// @Description 获取当前用户的任务模板列表，按创建次数降序排列
// @Tags 用户首页
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} response.Response
// @Router /user/home/templates [get]
func (h *UserHomeHandler) GetTaskTemplates(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if limit > 50 {
		limit = 50
	}

	var templates []models.TaskTemplate
	h.db.Where("user_id = ?", userID).
		Order("total_created_count DESC, last_used_at DESC").
		Limit(limit).
		Find(&templates)

	// 获取任务类型价格信息
	var taskTypes []models.TaskType
	h.db.Where("is_active = ?", true).Find(&taskTypes)
	typePrice := make(map[string]int)
	typeName := make(map[string]string)
	for _, t := range taskTypes {
		typePrice[t.TypeCode] = t.JingdouPrice
		typeName[t.TypeCode] = t.TypeName
	}

	// 组装返回数据
	result := make([]gin.H, 0, len(templates))
	for _, t := range templates {
		result = append(result, gin.H{
			"id":                  t.ID,
			"task_type":           t.TaskType,
			"task_type_name":      typeName[t.TaskType],
			"sku":                 t.SKU,
			"shop_name":           t.ShopName,
			"keyword":             t.Keyword,
			"remark":              t.Remark,
			"total_created_count": t.TotalCreatedCount,
			"last_used_at":        t.LastUsedAt,
			"jingdou_price":       typePrice[t.TaskType],
		})
	}

	response.Success(c, gin.H{
		"templates": result,
		"total":     len(result),
	})
}

// QuickCreateTask 从模板快速创建任务或直接创建任务
// @Summary 从模板快速创建任务或直接创建任务
// @Description 使用已保存的任务模板快速创建任务，或直接通过SKU创建新任务
// @Tags 用户首页
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.QuickCreateTaskRequest true "快速创建任务请求"
// @Success 200 {object} response.Response
// @Router /user/home/quick-create [post]
func (h *UserHomeHandler) QuickCreateTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	var req models.QuickCreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 判断是模板模式还是新任务模式
	var sku, taskTypeCode, keyword, shopName string
	var template *models.TaskTemplate

	if req.TemplateID != nil && *req.TemplateID > 0 {
		// 模板模式：从模板获取信息
		var t models.TaskTemplate
		if err := h.db.Where("id = ? AND user_id = ?", *req.TemplateID, userID).First(&t).Error; err != nil {
			response.Error(c, http.StatusNotFound, "任务模板不存在")
			return
		}
		template = &t
		sku = template.SKU
		taskTypeCode = template.TaskType
		keyword = template.Keyword
		shopName = template.ShopName
	} else if req.SKU != "" {
		// 新任务模式：直接使用请求中的SKU
		sku = req.SKU
		taskTypeCode = req.TaskType
		if taskTypeCode == "" {
			response.Error(c, http.StatusBadRequest, "新任务必须指定任务类型")
			return
		}
	} else {
		response.Error(c, http.StatusBadRequest, "必须提供模板ID或SKU编号")
		return
	}

	// 如果请求中指定了任务类型，覆盖默认值
	if req.TaskType != "" {
		taskTypeCode = req.TaskType
	}

	// 获取任务类型信息
	var taskType models.TaskType
	if err := h.db.Where("type_code = ? AND is_active = ?", taskTypeCode, true).First(&taskType).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "任务类型不可用或未启用")
		return
	}

	// 处理关键词：优先使用请求中的，否则使用默认值
	if req.Keyword != "" {
		keyword = req.Keyword
	}
	// 搜索关键词浏览任务必须有关键词
	if taskTypeCode == "search_browse" && keyword == "" {
		response.Error(c, http.StatusBadRequest, "关键词搜索浏览任务必须填写搜索关键词")
		return
	}
	// 非关键词搜索任务不需要关键词
	if taskTypeCode != "search_browse" {
		keyword = ""
	}

	// 处理店铺名称：优先使用请求中的，否则使用默认值
	if req.ShopName != "" {
		shopName = req.ShopName
	}
	// 搜索关键词浏览任务必须有店铺名称
	if taskTypeCode == "search_browse" && shopName == "" {
		response.Error(c, http.StatusBadRequest, "关键词搜索浏览任务必须填写店铺名称")
		return
	}

	// 获取用户信息
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 计算京豆消耗
	consumeJingdou := taskType.JingdouPrice * req.ExecuteCount
	isAdmin := user.Role == "admin"
	if isAdmin {
		consumeJingdou = 0
	}

	// 检查余额
	if !isAdmin && user.JingdouBalance < consumeJingdou {
		response.Error(c, http.StatusBadRequest, "京豆余额不足")
		return
	}

	// 开始事务
	tx := h.db.Begin()

	// 创建任务
	task := models.Task{
		UserID:         user.ID,
		TaskType:       taskTypeCode,
		SKU:            sku,
		ShopName:       shopName,
		Keyword:        keyword,
		StartTime:      req.StartTime,
		ExecuteCount:   req.ExecuteCount,
		ExecutedCount:  0,
		Priority:       0,
		Status:         "waiting",
		ConsumeJingdou: consumeJingdou,
		Remark:         "快速创建",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := tx.Create(&task).Error; err != nil {
		tx.Rollback()
		response.Error(c, http.StatusInternalServerError, "创建任务失败")
		return
	}

	// 扣除京豆
	if !isAdmin && consumeJingdou > 0 {
		user.JingdouBalance -= consumeJingdou
		tx.Save(&user)

		// 记录京豆日志
		jingdouLog := models.JingdouLog{
			UserID:        user.ID,
			Amount:        -consumeJingdou,
			Balance:       user.JingdouBalance,
			OperationType: "task",
			RelatedID:     &task.ID,
			Remark:        "快速创建任务扣除 - SKU:" + task.SKU,
			CreatedAt:     time.Now(),
		}
		tx.Create(&jingdouLog)
	}

	// 如果是模板模式，更新模板使用次数和参数
	if template != nil {
		template.TotalCreatedCount += req.ExecuteCount
		template.LastUsedAt = time.Now()
		template.UpdatedAt = time.Now()
		// 如果提供了新的关键词或店铺名，更新到模板
		if req.Keyword != "" && req.Keyword != template.Keyword {
			template.Keyword = req.Keyword
		}
		if req.ShopName != "" && req.ShopName != template.ShopName {
			template.ShopName = req.ShopName
		}
		tx.Save(template)
	} else {
		// 新任务模式，创建模板（仅普通用户）
		if !isAdmin {
			UpdateOrCreateTemplate(tx, user.ID, taskTypeCode, sku, shopName, keyword, req.ExecuteCount)
		}
	}

	tx.Commit()

	response.Success(c, gin.H{
		"message":         "任务创建成功",
		"task_id":         task.ID,
		"consume_jingdou": consumeJingdou,
		"jingdou_balance": user.JingdouBalance,
	})
}

// GetTemplatePrice 计算模板任务价格
// @Summary 计算模板任务价格
// @Description 根据模板和执行次数计算京豆消耗，支持指定任务类型
// @Tags 用户首页
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param template_id query int true "模板ID"
// @Param execute_count query int true "执行次数"
// @Param task_type query string false "任务类型(可选，覆盖模板的任务类型)"
// @Success 200 {object} response.Response
// @Router /user/home/template-price [get]
func (h *UserHomeHandler) GetTemplatePrice(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	var templateID, executeCount int
	fmt.Sscanf(c.Query("template_id"), "%d", &templateID)
	fmt.Sscanf(c.Query("execute_count"), "%d", &executeCount)
	taskTypeCode := c.Query("task_type") // 可选的任务类型参数

	if templateID == 0 || executeCount == 0 {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	// 获取模板
	var template models.TaskTemplate
	if err := h.db.Where("id = ? AND user_id = ?", templateID, userID).First(&template).Error; err != nil {
		response.Error(c, http.StatusNotFound, "任务模板不存在")
		return
	}

	// 使用指定的任务类型，否则使用模板的任务类型
	typeCode := template.TaskType
	if taskTypeCode != "" {
		typeCode = taskTypeCode
	}

	// 获取任务类型价格
	var taskType models.TaskType
	if err := h.db.Where("type_code = ? AND is_active = ?", typeCode, true).First(&taskType).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "任务类型不存在或未启用")
		return
	}

	consumeJingdou := taskType.JingdouPrice * executeCount

	// 获取用户余额
	var user models.User
	h.db.First(&user, userID)

	response.Success(c, gin.H{
		"template_id":     templateID,
		"execute_count":   executeCount,
		"task_type":       typeCode,
		"task_type_name":  taskType.TypeName,
		"jingdou_price":   taskType.JingdouPrice,
		"consume_jingdou": consumeJingdou,
		"jingdou_balance": user.JingdouBalance,
		"is_sufficient":   user.JingdouBalance >= consumeJingdou,
	})
}

// UpdateOrCreateTemplate 更新或创建任务模板（内部方法，任务创建时调用）
func UpdateOrCreateTemplate(db *gorm.DB, userID uint, taskType, sku, shopName, keyword string, executeCount int) {
	var template models.TaskTemplate
	err := db.Where("user_id = ? AND sku = ? AND task_type = ?", userID, sku, taskType).First(&template).Error

	now := time.Now()
	if err == gorm.ErrRecordNotFound {
		// 创建新模板
		template = models.TaskTemplate{
			UserID:            userID,
			TaskType:          taskType,
			SKU:               sku,
			ShopName:          shopName,
			Keyword:           keyword,
			TotalCreatedCount: executeCount,
			LastUsedAt:        now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		db.Create(&template)
	} else if err == nil {
		// 更新已有模板
		template.TotalCreatedCount += executeCount
		template.LastUsedAt = now
		template.UpdatedAt = now
		// 如果关键词不同，追加
		if keyword != "" && template.Keyword != keyword {
			if template.Keyword == "" {
				template.Keyword = keyword
			} else {
				// 检查是否已包含
				keywords := template.Keyword
				if !containsKeyword(keywords, keyword) {
					template.Keyword = keywords + "," + keyword
				}
			}
		}
		// 更新店铺名
		if shopName != "" && template.ShopName != shopName {
			template.ShopName = shopName
		}
		db.Save(&template)
	}
}

// containsKeyword 检查关键词列表是否包含指定关键词
func containsKeyword(keywords, target string) bool {
	// 简单实现，可以改进
	return keywords == target ||
		len(keywords) > len(target) && (keywords[:len(target)+1] == target+"," ||
			keywords[len(keywords)-len(target)-1:] == ","+target ||
			contains(keywords, ","+target+","))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetJingdouStats 获取用户京豆统计
// @Summary 获取用户京豆统计
// @Description 获取当前用户京豆余额、过去消耗、未来预计消耗、可消耗天数等
// @Tags 用户首页
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /user/home/jingdou-stats [get]
func (h *UserHomeHandler) GetJingdouStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	todayEnd := todayStart.Add(24 * time.Hour)

	// 获取用户信息
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 今日消耗京豆
	var todayConsumed int64
	h.db.Model(&models.Task{}).Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, todayStart, todayEnd).
		Select("COALESCE(SUM(consume_jingdou), 0)").Scan(&todayConsumed)

	// 过去7天消耗京豆
	past7DaysStart := todayStart.AddDate(0, 0, -7)
	var past7DaysConsumed int64
	h.db.Model(&models.Task{}).Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, past7DaysStart, todayEnd).
		Select("COALESCE(SUM(consume_jingdou), 0)").Scan(&past7DaysConsumed)

	// 过去30天消耗京豆
	past30DaysStart := todayStart.AddDate(0, 0, -30)
	var past30DaysConsumed int64
	h.db.Model(&models.Task{}).Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, past30DaysStart, todayEnd).
		Select("COALESCE(SUM(consume_jingdou), 0)").Scan(&past30DaysConsumed)

	// 未来待执行任务预计消耗（waiting/running状态的任务）
	var futureConsumed int64
	h.db.Model(&models.Task{}).Where("user_id = ? AND status IN (?, ?)", userID, "waiting", "running").
		Select("COALESCE(SUM(consume_jingdou), 0)").Scan(&futureConsumed)

	// 计算日均消耗（基于过去7天）
	dailyAvgConsumed := 0.0
	if past7DaysConsumed > 0 {
		dailyAvgConsumed = float64(past7DaysConsumed) / 7.0
	}

	// 计算当前余额可消耗天数
	estimatedDays := 0.0
	if dailyAvgConsumed > 0 {
		estimatedDays = float64(user.JingdouBalance) / dailyAvgConsumed
	}

	// 计算扣除未来任务后可用余额
	availableBalance := user.JingdouBalance - int(futureConsumed)
	if availableBalance < 0 {
		availableBalance = 0
	}

	// 计算扣除未来任务后可消耗天数
	estimatedDaysAfterFuture := 0.0
	if dailyAvgConsumed > 0 {
		estimatedDaysAfterFuture = float64(availableBalance) / dailyAvgConsumed
	}

	response.Success(c, gin.H{
		"jingdou_balance":             user.JingdouBalance,                           // 当前余额
		"today_consumed":              todayConsumed,                                 // 今日已消耗
		"past_7_days_consumed":        past7DaysConsumed,                             // 过去7天消耗
		"past_30_days_consumed":       past30DaysConsumed,                            // 过去30天消耗
		"future_consumed":             futureConsumed,                                // 未来待执行任务预计消耗
		"available_balance":           availableBalance,                              // 扣除未来任务后可用余额
		"daily_avg_consumed":          fmt.Sprintf("%.1f", dailyAvgConsumed),         // 日均消耗
		"estimated_days":              fmt.Sprintf("%.1f", estimatedDays),            // 预计可消耗天数
		"estimated_days_after_future": fmt.Sprintf("%.1f", estimatedDaysAfterFuture), // 扣除未来任务后可消耗天数
	})
}

// UpdateTemplateRemark 更新任务模板备注
// @Summary 更新任务模板备注
// @Description 更新用户的任务模板备注
// @Tags 用户首页
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param request body models.UpdateTemplateRemarkRequest true "备注信息"
// @Success 200 {object} response.Response
// @Router /user/home/templates/{id}/remark [put]
func (h *UserHomeHandler) UpdateTemplateRemark(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	templateID := c.Param("id")
	var req models.UpdateTemplateRemarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 查找模板
	var template models.TaskTemplate
	if err := h.db.Where("id = ? AND user_id = ?", templateID, userID).First(&template).Error; err != nil {
		response.Error(c, http.StatusNotFound, "任务模板不存在")
		return
	}

	// 更新备注
	template.Remark = req.Remark
	template.UpdatedAt = time.Now()

	if err := h.db.Save(&template).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}

	response.Success(c, gin.H{
		"message": "备注更新成功",
		"remark":  template.Remark,
	})
}
