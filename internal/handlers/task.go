package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"jd-task-platform-go/internal/constants"
	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
)

type TaskHandler struct {
	db *gorm.DB
}

func NewTaskHandler(db *gorm.DB) *TaskHandler {
	return &TaskHandler{db: db}
}

// GetTasks 获取任务列表
// @Summary 获取任务列表
// @Description 获取任务列表，支持分页和筛选
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "任务状态" Enums(waiting, running, completed, failed, cancelled)
// @Param task_type query string false "任务类型"
// @Success 200 {object} response.Response{data=object}
// @Failure 401 {object} response.Response
// @Router /tasks [get]
func (h *TaskHandler) GetTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	taskType := c.Query("task_type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	query := h.db.Model(&models.Task{})

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}

	var total int64
	query.Count(&total)

	var tasks []models.Task
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks)

	items := make([]gin.H, 0)
	for _, task := range tasks {
		var user models.User
		h.db.First(&user, task.UserID)

		items = append(items, gin.H{
			"id":              task.ID,
			"user_id":         task.UserID,
			"username":        user.Username,
			"task_type":       task.TaskType,
			"sku":             task.SKU,
			"shop_name":       task.ShopName,
			"keyword":         task.Keyword,
			"start_time":      task.StartTime.Format(time.RFC3339),
			"execute_count":   task.ExecuteCount,
			"executed_count":  task.ExecutedCount,
			"priority":        task.Priority,
			"status":          task.Status,
			"consume_jingdou": task.ConsumeJingdou,
			"remark":          task.Remark,
			"created_at":      task.CreatedAt.Format(time.RFC3339),
			"updated_at":      task.UpdatedAt.Format(time.RFC3339),
		})
	}

	pages := int((total + int64(pageSize) - 1) / int64(pageSize))

	response.Success(c, gin.H{
		"items":    items,
		"page":     page,
		"per_page": pageSize,
		"total":    total,
		"pages":    pages,
	})
}

// CreateTask 创建任务
// @Summary 创建任务
// @Description 创建新任务，管理员创建不消耗京豆
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateTaskRequest true "任务信息"
// @Success 201 {object} response.Response{data=object}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, constants.MsgTaskCreateParamError)
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, constants.MsgUserNotFound)
		return
	}

	// 查询任务类型配置
	var taskType models.TaskType
	if err := h.db.Where("type_code = ?", req.TaskType).First(&taskType).Error; err != nil {
		response.Error(c, http.StatusBadRequest, constants.MsgTaskTypeInvalid)
		return
	}

	// 检查任务类型是否启用
	if !taskType.IsActive {
		response.Error(c, http.StatusBadRequest, constants.MsgTaskTypeDisabled)
		return
	}

	// 只有关键词搜索任务需要关键词参数
	if req.TaskType == "search_browse" && req.Keyword == "" {
		response.Error(c, http.StatusBadRequest, "关键词搜索任务必须填写关键词")
		return
	}
	// 非关键词搜索任务不需要关键词，清空它
	if req.TaskType != "search_browse" {
		req.Keyword = ""
	}

	// 获取用户角色
	isAdmin := role == "admin"

	// 检查当前时间是否在允许的时间段内（管理员可以在任何时间创建任务）
	if !isAdmin && taskType.TimeSlot1Start != nil && taskType.TimeSlot1End != nil &&
		*taskType.TimeSlot1Start != "" && *taskType.TimeSlot1End != "" {
		now := time.Now()
		currentTime := now.Format("15:04")

		inTimeSlot := false

		// 检查时间段1
		if currentTime >= *taskType.TimeSlot1Start && currentTime <= *taskType.TimeSlot1End {
			inTimeSlot = true
		}

		// 检查时间段2（如果存在）
		if !inTimeSlot && taskType.TimeSlot2Start != nil && taskType.TimeSlot2End != nil &&
			*taskType.TimeSlot2Start != "" && *taskType.TimeSlot2End != "" {
			if currentTime >= *taskType.TimeSlot2Start && currentTime <= *taskType.TimeSlot2End {
				inTimeSlot = true
			}
		}

		if !inTimeSlot {
			timeSlotInfo := *taskType.TimeSlot1Start + "-" + *taskType.TimeSlot1End
			if taskType.TimeSlot2Start != nil && taskType.TimeSlot2End != nil &&
				*taskType.TimeSlot2Start != "" && *taskType.TimeSlot2End != "" {
				timeSlotInfo += ", " + *taskType.TimeSlot2Start + "-" + *taskType.TimeSlot2End
			}
			response.Errorf(c, http.StatusBadRequest, constants.MsgTaskTimeSlotLimit, timeSlotInfo)
			return
		}
	}

	// 使用任务类型配置的价格
	consumeJingdou := taskType.JingdouPrice * req.ExecuteCount

	if isAdmin {
		consumeJingdou = 0
	} else {
		if user.JingdouBalance < consumeJingdou {
			response.Errorf(c, http.StatusBadRequest, constants.MsgTaskBalanceInsufficient, consumeJingdou, user.JingdouBalance)
			return
		}
		user.JingdouBalance -= consumeJingdou
	}

	task := models.Task{
		UserID:         user.ID,
		TaskType:       req.TaskType,
		SKU:            req.SKU,
		ShopName:       req.ShopName,
		Keyword:        req.Keyword,
		StartTime:      req.StartTime,
		ExecuteCount:   req.ExecuteCount,
		ExecutedCount:  0,
		Priority:       req.Priority,
		Status:         "waiting",
		ConsumeJingdou: consumeJingdou,
		Remark:         req.Remark,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	tx := h.db.Begin()
	if err := tx.Create(&task).Error; err != nil {
		tx.Rollback()
		response.Error(c, http.StatusInternalServerError, constants.MsgTaskCreateFailed)
		return
	}

	// 非管理员扣除京豆并记录日志
	if !isAdmin && consumeJingdou > 0 {
		tx.Save(&user)

		// 创建京豆扣除日志
		jingdouLog := models.JingdouLog{
			UserID:        user.ID,
			Amount:        -consumeJingdou, // 负数表示扣除
			Balance:       user.JingdouBalance,
			OperationType: "task",
			RelatedID:     &task.ID,
			Remark:        "创建任务扣除 - SKU:" + task.SKU,
			CreatedAt:     time.Now(),
		}
		tx.Create(&jingdouLog)
	}

	// 更新或创建任务模板（仅普通用户）
	if !isAdmin {
		UpdateOrCreateTemplate(tx, user.ID, task.TaskType, task.SKU, task.ShopName, task.Keyword, task.ExecuteCount)
	}

	tx.Commit()

	response.SuccessWithDataAndMsgf(c, gin.H{
		"task_id":         task.ID,
		"consume_jingdou": consumeJingdou,
		"balance":         user.JingdouBalance,
		"is_admin":        isAdmin,
	}, constants.MsgTaskCreated, consumeJingdou)
}

// GetTaskByID 获取任务详情
// @Summary 获取任务详情
// @Description 根据ID获取任务详细信息
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response{data=object}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var task models.Task
	if err := h.db.First(&task, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, constants.MsgTaskNotFound)
		return
	}

	if role != "admin" && task.UserID != userID.(uint) {
		response.Error(c, http.StatusForbidden, constants.MsgTaskNoPermission)
		return
	}

	var user models.User
	h.db.First(&user, task.UserID)

	response.Success(c, gin.H{
		"id":              task.ID,
		"user_id":         task.UserID,
		"user_name":       user.Username,
		"task_type":       task.TaskType,
		"sku":             task.SKU,
		"shop_name":       task.ShopName,
		"keyword":         task.Keyword,
		"start_time":      task.StartTime.Format(time.RFC3339),
		"execute_count":   task.ExecuteCount,
		"executed_count":  task.ExecutedCount,
		"priority":        task.Priority,
		"status":          task.Status,
		"consume_jingdou": task.ConsumeJingdou,
		"remark":          task.Remark,
		"created_at":      task.CreatedAt.Format(time.RFC3339),
		"updated_at":      task.UpdatedAt.Format(time.RFC3339),
	})
}

// UpdateTask 更新任务
// @Summary 更新任务信息
// @Description 更新任务的部分字段
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务ID"
// @Param request body models.UpdateTaskRequest true "任务信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, constants.MsgTaskUpdateParamError)
		return
	}

	var task models.Task
	if err := h.db.First(&task, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, constants.MsgTaskNotFound)
		return
	}

	if req.ShopName != nil {
		task.ShopName = *req.ShopName
	}
	if req.Keyword != nil {
		task.Keyword = *req.Keyword
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Status != nil {
		task.Status = *req.Status
	}

	task.UpdatedAt = time.Now()
	h.db.Save(&task)

	response.SuccessWithMsg(c, constants.MsgTaskUpdated, nil)
}

// DeleteTask 删除任务
// @Summary 删除任务
// @Description 根据ID删除任务
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	h.db.Delete(&models.Task{}, id)
	response.SuccessWithMsg(c, constants.MsgTaskDeleted, nil)
}

// GetTaskStats 获取任务统计
// @Summary 获取任务统计数据
// @Description 获取任务的统计信息，包括总数、运行中、等待中等
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Router /tasks/stats [get]
func (h *TaskHandler) GetTaskStats(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	query := h.db.Model(&models.Task{})
	if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	var stats struct {
		TotalTasks     int64 `json:"total_tasks"`
		RunningTasks   int64 `json:"running_tasks"`
		WaitingTasks   int64 `json:"waiting_tasks"`
		CompletedTasks int64 `json:"completed_tasks"`
		FailedTasks    int64 `json:"failed_tasks"`
	}

	query.Count(&stats.TotalTasks)
	h.db.Model(&models.Task{}).Where("status = ?", "running").Count(&stats.RunningTasks)
	h.db.Model(&models.Task{}).Where("status = ?", "waiting").Count(&stats.WaitingTasks)
	h.db.Model(&models.Task{}).Where("status = ?", "completed").Count(&stats.CompletedTasks)
	h.db.Model(&models.Task{}).Where("status = ?", "failed").Count(&stats.FailedTasks)

	response.Success(c, stats)
}
