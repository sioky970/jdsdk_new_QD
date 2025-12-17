package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
)

// UserTaskManageHandler 用户任务管理处理器
type UserTaskManageHandler struct {
	db *gorm.DB
}

// NewUserTaskManageHandler 创建用户任务管理处理器
func NewUserTaskManageHandler(db *gorm.DB) *UserTaskManageHandler {
	return &UserTaskManageHandler{db: db}
}

// GetUserTasks 获取用户任务列表
// @Summary 获取用户任务列表
// @Description 获取当前用户的任务列表，支持按状态和时间筛选，支持排序，管理员可指定user_id
// @Tags 用户任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "用户ID（仅管理员）"
// @Param status query string false "任务状态: waiting,running,completed,failed,cancelled,partial_completed"
// @Param start_date query string false "开始日期 YYYY-MM-DD"
// @Param end_date query string false "结束日期 YYYY-MM-DD"
// @Param sort_by query string false "排序字段: start_time,created_at" default(created_at)
// @Param sort_order query string false "排序方式: asc,desc" default(desc)
// @Param page query int false "页码" default(1)
// @Param per_page query int false "每页数量" default(20)
// @Success 200 {object} response.Response
// @Router /user/tasks [get]
func (h *UserTaskManageHandler) GetUserTasks(c *gin.Context) {
	userID := c.GetUint("user_id")
	role, _ := c.Get("role")

	// 确定要查询的用户ID
	targetUserID := userID
	// 如果是管理员且指定了user_id参数
	if role == "admin" {
		if targetUserIDParam := c.Query("user_id"); targetUserIDParam != "" {
			if parsedID, err := strconv.ParseUint(targetUserIDParam, 10, 32); err == nil {
				targetUserID = uint(parsedID)
			}
		}
	}

	// 解析查询参数
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	// 验证排序字段（只允许指定字段）
	allowedSortFields := map[string]bool{"start_time": true, "created_at": true}
	if !allowedSortFields[sortBy] {
		sortBy = "created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// 构建查询
	query := h.db.Model(&models.Task{}).Where("user_id = ?", targetUserID)

	// 状态筛选（支持多个状态，用逗号分隔）
	if status != "" {
		statuses := strings.Split(status, ",")
		if len(statuses) > 1 {
			query = query.Where("status IN ?", statuses)
		} else {
			query = query.Where("status = ?", status)
		}
	}

	// 时间筛选
	if startDate != "" {
		t, err := time.Parse("2006-01-02", startDate)
		if err == nil {
			query = query.Where("start_time >= ?", t)
		}
	}
	if endDate != "" {
		t, err := time.Parse("2006-01-02", endDate)
		if err == nil {
			// 包含结束日期的整天
			query = query.Where("start_time < ?", t.Add(24*time.Hour))
		}
	}

	// 统计总数
	var total int64
	query.Count(&total)

	// 分页查询（应用排序）
	var tasks []models.Task
	offset := (page - 1) * perPage
	orderClause := sortBy + " " + sortOrder
	query.Order(orderClause).Offset(offset).Limit(perPage).Find(&tasks)

	// 获取任务类型名称映射
	var taskTypes []models.TaskType
	h.db.Find(&taskTypes)
	typeNameMap := make(map[string]string)
	for _, tt := range taskTypes {
		typeNameMap[tt.TypeCode] = tt.TypeName
	}

	// 构建响应
	items := make([]gin.H, 0)
	for _, task := range tasks {
		items = append(items, gin.H{
			"id":              task.ID,
			"task_type":       task.TaskType,
			"task_type_name":  typeNameMap[task.TaskType],
			"sku":             task.SKU,
			"shop_name":       task.ShopName,
			"keyword":         task.Keyword,
			"start_time":      task.StartTime.Format(time.RFC3339),
			"execute_count":   task.ExecuteCount,
			"executed_count":  task.ExecutedCount,
			"priority":        task.Priority,
			"status":          task.Status,
			"status_text":     getStatusText(task.Status),
			"consume_jingdou": task.ConsumeJingdou,
			"can_cancel":      canCancelTask(task),
			"can_edit":        canEditTask(task),
			"created_at":      task.CreatedAt.Format(time.RFC3339),
		})
	}

	response.Success(c, gin.H{
		"tasks":    items,
		"total":    total,
		"page":     page,
		"per_page": perPage,
		"pages":    (total + int64(perPage) - 1) / int64(perPage),
	})
}

// CancelUserTask 取消用户任务
// @Summary 取消用户任务
// @Description 取消未开始的任务，退还京豆
// @Tags 用户任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response
// @Router /user/tasks/{id}/cancel [post]
func (h *UserTaskManageHandler) CancelUserTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID := c.Param("id")

	// 查找任务
	var task models.Task
	if err := h.db.First(&task, taskID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "任务不存在")
		return
	}

	// 验证是否属于当前用户
	if task.UserID != userID {
		response.Error(c, http.StatusForbidden, "无权操作此任务")
		return
	}

	// 检查是否可以取消
	if !canCancelTask(task) {
		response.Error(c, http.StatusBadRequest, "只能取消待开始的任务")
		return
	}

	tx := h.db.Begin()

	// 更新任务状态
	task.Status = "cancelled"
	task.UpdatedAt = time.Now()
	if err := tx.Save(&task).Error; err != nil {
		tx.Rollback()
		response.Error(c, http.StatusInternalServerError, "取消任务失败")
		return
	}

	// 退还京豆
	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		response.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	refundAmount := task.ConsumeJingdou
	user.JingdouBalance += refundAmount
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		response.Error(c, http.StatusInternalServerError, "退还京豆失败")
		return
	}

	// 记录京豆日志
	jingdouLog := models.JingdouLog{
		UserID:        userID,
		Amount:        refundAmount,
		Balance:       user.JingdouBalance,
		OperationType: "refund",
		RelatedID:     &task.ID,
		Remark:        "取消任务退还京豆",
		CreatedAt:     time.Now(),
	}
	tx.Create(&jingdouLog)

	tx.Commit()

	response.SuccessWithMsg(c, "任务取消成功，已退还"+strconv.Itoa(refundAmount)+"京豆", gin.H{
		"task_id":        task.ID,
		"refund_jingdou": refundAmount,
		"new_balance":    user.JingdouBalance,
	})
}

// UpdateUserTask 修改用户任务
// @Summary 修改用户任务
// @Description 修改未开始的任务参数，同时更新对应模板
// @Tags 用户任务管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务ID"
// @Param request body object true "修改参数"
// @Success 200 {object} response.Response
// @Router /user/tasks/{id} [put]
func (h *UserTaskManageHandler) UpdateUserTask(c *gin.Context) {
	userID := c.GetUint("user_id")
	taskID := c.Param("id")

	var req struct {
		ShopName     *string    `json:"shop_name"`
		Keyword      *string    `json:"keyword"`
		StartTime    *time.Time `json:"start_time"`
		ExecuteCount *int       `json:"execute_count"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 查找任务
	var task models.Task
	if err := h.db.First(&task, taskID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "任务不存在")
		return
	}

	// 验证是否属于当前用户
	if task.UserID != userID {
		response.Error(c, http.StatusForbidden, "无权操作此任务")
		return
	}

	// 检查是否可以编辑
	if !canEditTask(task) {
		response.Error(c, http.StatusBadRequest, "只能修改待开始的任务")
		return
	}

	// 不能减少任务次数
	if req.ExecuteCount != nil && *req.ExecuteCount < task.ExecuteCount {
		response.Error(c, http.StatusBadRequest, "不能减少任务执行次数")
		return
	}

	tx := h.db.Begin()

	// 计算京豆差额（如果增加次数）
	var additionalJingdou int
	if req.ExecuteCount != nil && *req.ExecuteCount > task.ExecuteCount {
		// 获取任务类型单价
		var taskType models.TaskType
		if err := h.db.Where("type_code = ?", task.TaskType).First(&taskType).Error; err != nil {
			tx.Rollback()
			response.Error(c, http.StatusInternalServerError, "获取任务类型失败")
			return
		}

		additionalCount := *req.ExecuteCount - task.ExecuteCount
		additionalJingdou = additionalCount * taskType.JingdouPrice

		// 检查用户余额
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			tx.Rollback()
			response.Error(c, http.StatusInternalServerError, "获取用户信息失败")
			return
		}

		if user.JingdouBalance < additionalJingdou {
			tx.Rollback()
			response.Error(c, http.StatusBadRequest, "京豆余额不足，需要额外"+strconv.Itoa(additionalJingdou)+"京豆")
			return
		}

		// 扣除京豆
		user.JingdouBalance -= additionalJingdou
		tx.Save(&user)

		// 记录京豆日志
		jingdouLog := models.JingdouLog{
			UserID:        userID,
			Amount:        -additionalJingdou,
			Balance:       user.JingdouBalance,
			OperationType: "consume",
			RelatedID:     &task.ID,
			Remark:        "修改任务增加执行次数",
			CreatedAt:     time.Now(),
		}
		tx.Create(&jingdouLog)

		task.ExecuteCount = *req.ExecuteCount
		task.ConsumeJingdou += additionalJingdou
	}

	// 更新任务参数
	if req.ShopName != nil {
		task.ShopName = *req.ShopName
	}
	if req.Keyword != nil {
		// 只有关键词搜索任务才能设置关键词
		if task.TaskType == "search_browse" {
			task.Keyword = *req.Keyword
		}
	}
	if req.StartTime != nil {
		task.StartTime = *req.StartTime
	}

	task.UpdatedAt = time.Now()
	if err := tx.Save(&task).Error; err != nil {
		tx.Rollback()
		response.Error(c, http.StatusInternalServerError, "更新任务失败")
		return
	}

	// 同步更新任务模板
	var template models.TaskTemplate
	if err := h.db.Where("user_id = ? AND sku = ? AND task_type = ?", userID, task.SKU, task.TaskType).First(&template).Error; err == nil {
		// 找到模板，更新参数
		if req.ShopName != nil {
			template.ShopName = *req.ShopName
		}
		if req.Keyword != nil && task.TaskType == "search_browse" {
			template.Keyword = *req.Keyword
		}
		template.UpdatedAt = time.Now()
		tx.Save(&template)
	}

	tx.Commit()

	response.SuccessWithMsg(c, "任务修改成功", gin.H{
		"task_id":            task.ID,
		"additional_jingdou": additionalJingdou,
	})
}

// GetTaskStatusOptions 获取任务状态选项
// @Summary 获取任务状态选项
// @Description 获取所有可用的任务状态选项
// @Tags 用户任务管理
// @Produce json
// @Success 200 {object} response.Response
// @Router /user/tasks/status-options [get]
func (h *UserTaskManageHandler) GetTaskStatusOptions(c *gin.Context) {
	options := []gin.H{
		{"value": "", "label": "全部状态"},
		{"value": "waiting", "label": "待开始"},
		{"value": "running", "label": "执行中"},
		{"value": "completed", "label": "已完成"},
		{"value": "partial_completed", "label": "部分完成"},
		{"value": "failed", "label": "失败"},
		{"value": "cancelled", "label": "已取消"},
	}
	response.Success(c, gin.H{
		"options": options,
	})
}

// 获取状态文本
func getStatusText(status string) string {
	switch status {
	case "waiting":
		return "待开始"
	case "running":
		return "执行中"
	case "completed":
		return "已完成"
	case "partial_completed":
		return "部分完成"
	case "failed":
		return "失败"
	case "cancelled":
		return "已取消"
	default:
		return status
	}
}

// 判断任务是否可以取消
func canCancelTask(task models.Task) bool {
	// 只有待开始状态且未到开始时间的任务可以取消
	return task.Status == "waiting" && time.Now().Before(task.StartTime)
}

// 判断任务是否可以编辑
func canEditTask(task models.Task) bool {
	// 只有待开始状态且未到开始时间的任务可以编辑
	return task.Status == "waiting" && time.Now().Before(task.StartTime)
}
