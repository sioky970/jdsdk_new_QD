package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
)

// 任务模块扩展handler

// CancelTask 取消任务
// @Summary 取消任务
// @Description 取消等待中的任务并退款京豆
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response{data=object}
// @Router /tasks/{id}/cancel [post]
func (h *TaskHandler) CancelTask(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var task models.Task
	if err := h.db.First(&task, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "任务不存在")
		return
	}

	// 检查权限
	if role != "admin" && task.UserID != userID.(uint) {
		response.Error(c, http.StatusForbidden, "无权取消此任务")
		return
	}

	// 只有waiting状态的任务可以取消
	if task.Status != "waiting" {
		response.Error(c, http.StatusBadRequest, "只有等待中的任务可以取消")
		return
	}

	// 获取用户
	var user models.User
	h.db.First(&user, task.UserID)

	// 退款
	refundAmount := task.ConsumeJingdou
	user.JingdouBalance += refundAmount

	tx := h.db.Begin()

	// 更新任务状态
	task.Status = "cancelled"
	tx.Save(&task)

	// 更新用户余额
	tx.Save(&user)

	// 记录京豆日志
	if refundAmount > 0 {
		jingdouLog := models.JingdouLog{
			UserID:        user.ID,
			Amount:        refundAmount,
			Balance:       user.JingdouBalance,
			OperationType: "refund",
			RelatedID:     &task.ID,
			Remark:        "取消任务退款 - SKU:" + task.SKU,
			CreatedAt:     time.Now(),
		}
		tx.Create(&jingdouLog)
	}

	tx.Commit()

	response.Success(c, gin.H{
		"task_id":        task.ID,
		"refund_jingdou": refundAmount,
		"balance":        user.JingdouBalance,
	})
}

// GetTaskTypes 获取任务类型列表
// @Summary 获取任务类型列表
// @Description 获取所有任务类型配置，包括价格、启用状态和时间段限制
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Router /tasks/types [get]
func (h *TaskHandler) GetTaskTypes(c *gin.Context) {
	// 获取当前用户角色
	role, _ := c.Get("role")
	isAdmin := role == "admin"

	var taskTypes []models.TaskType
	// 普通用户只能看到已启用的任务类型
	if isAdmin {
		// 管理员可以看到所有任务类型（包括禁用的）
		h.db.Order("id ASC").Find(&taskTypes)
	} else {
		// 普通用户只能看到已启用的任务类型
		h.db.Where("is_active = ?", true).Order("id ASC").Find(&taskTypes)
	}

	items := make([]gin.H, 0)
	for _, tt := range taskTypes {
		item := gin.H{
			"id":               tt.ID,
			"type_code":        tt.TypeCode,
			"type_name":        tt.TypeName,
			"jingdou_price":    tt.JingdouPrice,
			"is_active":        tt.IsActive,
			"is_system_preset": tt.IsSystemPreset,
			"created_at":       tt.CreatedAt.Format(time.RFC3339),
			"updated_at":       tt.UpdatedAt.Format(time.RFC3339),
		}

		// 仅管理员可见执行倍数
		if isAdmin {
			multiplier := tt.ExecuteMultiplier
			if multiplier < 1 {
				multiplier = 1 // 默认倍数为1
			}
			item["execute_multiplier"] = multiplier
		}

		// 添加时间段信息（如果配置了时间限制）
		if tt.TimeSlot1Start != nil && tt.TimeSlot1End != nil {
			item["time_slot1_start"] = *tt.TimeSlot1Start
			item["time_slot1_end"] = *tt.TimeSlot1End

			// 如果配置了第二个时间段
			if tt.TimeSlot2Start != nil && tt.TimeSlot2End != nil &&
				*tt.TimeSlot2Start != "" && *tt.TimeSlot2End != "" {
				item["time_slot2_start"] = *tt.TimeSlot2Start
				item["time_slot2_end"] = *tt.TimeSlot2End
				item["time_slots"] = []string{
					*tt.TimeSlot1Start + "-" + *tt.TimeSlot1End,
					*tt.TimeSlot2Start + "-" + *tt.TimeSlot2End,
				}
			} else {
				item["time_slots"] = []string{
					*tt.TimeSlot1Start + "-" + *tt.TimeSlot1End,
				}
			}
			item["has_time_limit"] = true
		} else {
			// 没有时间限制
			item["has_time_limit"] = false
			item["time_slots"] = []string{}
		}

		items = append(items, item)
	}

	response.Success(c, gin.H{
		"task_types": items,
		"total":      len(items),
	})
}

// CreateTaskType 创建任务类型
// @Summary 创建任务类型
// @Description 创建新的任务类型（仅管理员，不允许创建系统预设类型）
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateTaskTypeRequest true "任务类型信息"
// @Success 201 {object} response.Response{data=object}
// @Router /tasks/types [post]
func (h *TaskHandler) CreateTaskType(c *gin.Context) {
	// 禁止创建任务类型，系统已预设所有类型
	response.Error(c, http.StatusForbidden, "系统已预设所有任务类型，不允许创建新类型")
}

// UpdateTaskType 更新任务类型
// @Summary 更新任务类型
// @Description 更新任务类型信息（仅管理员，系统预设类型不可修改名称）
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务类型ID"
// @Param request body models.UpdateTaskTypeRequest true "任务类型信息"
// @Success 200 {object} response.Response
// @Router /tasks/types/{id} [put]
func (h *TaskHandler) UpdateTaskType(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateTaskTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("更新任务类型参数解析失败: %v", err)
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 调试日志：打印接收到的参数
	log.Printf("更新任务类型 ID=%s, 请求参数: TypeName=%v, JingdouPrice=%v, IsActive=%v, ExecuteMultiplier=%v",
		id,
		req.TypeName,
		req.JingdouPrice,
		req.IsActive,
		req.ExecuteMultiplier)

	var taskType models.TaskType
	if err := h.db.First(&taskType, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "任务类型不存在")
		return
	}

	// 系统预设类型不允许修改代码，但允许修改名称
	if taskType.IsSystemPreset {
		// 系统预设类型可以修改名称、价格、启用状态、时间段和执行倍数
		if req.TypeName != nil && *req.TypeName != "" {
			taskType.TypeName = *req.TypeName
		}
		if req.JingdouPrice != nil {
			taskType.JingdouPrice = *req.JingdouPrice
		}
		if req.IsActive != nil {
			taskType.IsActive = *req.IsActive
		}
		if req.ExecuteMultiplier != nil && *req.ExecuteMultiplier >= 1 {
			taskType.ExecuteMultiplier = *req.ExecuteMultiplier
		}
		if req.TimeSlot1Start != nil {
			taskType.TimeSlot1Start = req.TimeSlot1Start
		}
		if req.TimeSlot1End != nil {
			taskType.TimeSlot1End = req.TimeSlot1End
		}
		if req.TimeSlot2Start != nil {
			taskType.TimeSlot2Start = req.TimeSlot2Start
		}
		if req.TimeSlot2End != nil {
			taskType.TimeSlot2End = req.TimeSlot2End
		}
	} else {
		// 非预设类型可以修改所有字段（理论上不存在，但保留逻辑）
		if req.TypeName != nil && *req.TypeName != "" {
			taskType.TypeName = *req.TypeName
		}
		if req.JingdouPrice != nil {
			taskType.JingdouPrice = *req.JingdouPrice
		}
		if req.IsActive != nil {
			taskType.IsActive = *req.IsActive
		}
		if req.ExecuteMultiplier != nil && *req.ExecuteMultiplier >= 1 {
			taskType.ExecuteMultiplier = *req.ExecuteMultiplier
		}
		if req.TimeSlot1Start != nil {
			taskType.TimeSlot1Start = req.TimeSlot1Start
		}
		if req.TimeSlot1End != nil {
			taskType.TimeSlot1End = req.TimeSlot1End
		}
		if req.TimeSlot2Start != nil {
			taskType.TimeSlot2Start = req.TimeSlot2Start
		}
		if req.TimeSlot2End != nil {
			taskType.TimeSlot2End = req.TimeSlot2End
		}
	}

	taskType.UpdatedAt = time.Now()
	if err := h.db.Save(&taskType).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "保存失败，请重试")
		return
	}

	response.SuccessWithMsg(c, "任务类型更新成功", nil)
}

// UpdateTaskPriority 修改任务优先级
// @Summary 修改任务优先级
// @Description 修改任务优先级（仅管理员）
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务ID"
// @Param request body object{priority=int} true "优先级"
// @Success 200 {object} response.Response
// @Router /tasks/{id}/priority [put]
func (h *TaskHandler) UpdateTaskPriority(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Priority int `json:"priority" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	var task models.Task
	if err := h.db.First(&task, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "任务不存在")
		return
	}

	task.Priority = req.Priority
	task.UpdatedAt = time.Now()
	h.db.Save(&task)

	response.SuccessWithMsg(c, "任务优先级更新成功", nil)
}

// GetTaskStatistics 获取任务统计（支持user_id参数）
// @Summary 获取任务统计
// @Description 获取任务统计数据，管理员可指定user_id
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "用户ID（仅管理员）"
// @Success 200 {object} response.Response{data=object}
// @Router /tasks/statistics [get]
func (h *TaskHandler) GetTaskStatistics(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	targetUserID := c.Query("user_id")

	query := h.db.Model(&models.Task{})

	// 如果是管理员且指定了user_id
	if role == "admin" && targetUserID != "" {
		query = query.Where("user_id = ?", targetUserID)
	} else if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	var stats struct {
		TotalTasks      int64 `json:"total_tasks"`
		RunningTasks    int64 `json:"running_tasks"`
		WaitingTasks    int64 `json:"waiting_tasks"`
		CompletedTasks  int64 `json:"completed_tasks"`
		FailedTasks     int64 `json:"failed_tasks"`
		TodayTasks      int64 `json:"today_tasks"`
		ConsumedJingdou int   `json:"consumed_jingdou"`
	}

	query.Count(&stats.TotalTasks)

	var queryClone = h.db.Model(&models.Task{})
	if role == "admin" && targetUserID != "" {
		queryClone = queryClone.Where("user_id = ?", targetUserID)
	} else if role != "admin" {
		queryClone = queryClone.Where("user_id = ?", userID)
	}

	queryClone.Where("status = ?", "running").Count(&stats.RunningTasks)
	queryClone.Where("status = ?", "waiting").Count(&stats.WaitingTasks)
	queryClone.Where("status = ?", "completed").Count(&stats.CompletedTasks)
	queryClone.Where("status = ?", "failed").Count(&stats.FailedTasks)

	// 今日任务
	today := time.Now().Format("2006-01-02")
	queryClone.Where("DATE(created_at) = ?", today).Count(&stats.TodayTasks)

	// 京豆消耗
	var totalConsumed int
	queryClone.Select("SUM(consume_jingdou)").Row().Scan(&totalConsumed)
	stats.ConsumedJingdou = totalConsumed

	// 获取用户京豆余额
	var user models.User
	h.db.First(&user, userID)

	result := gin.H{
		"total_tasks":      stats.TotalTasks,
		"running_tasks":    stats.RunningTasks,
		"waiting_tasks":    stats.WaitingTasks,
		"completed_tasks":  stats.CompletedTasks,
		"failed_tasks":     stats.FailedTasks,
		"today_tasks":      stats.TodayTasks,
		"consumed_jingdou": stats.ConsumedJingdou,
		"jingdou_balance":  user.JingdouBalance,
	}

	response.Success(c, result)
}

// BatchCreateTasks 批量创建任务
// @Summary 批量创建任务
// @Description 批量创建任务，最多100个
// @Tags 任务模块
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.BatchCreateTaskRequest true "批量任务信息"
// @Success 201 {object} response.Response{data=object}
// @Router /tasks/apikey/batch [post]
func (h *TaskHandler) BatchCreateTasks(c *gin.Context) {
	var req models.BatchCreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	if len(req.Tasks) == 0 {
		response.Error(c, http.StatusBadRequest, "任务列表不能为空")
		return
	}

	if len(req.Tasks) > 100 {
		response.Error(c, http.StatusBadRequest, "单次最多创建100个任务")
		return
	}

	// 从API Key中间件获取用户
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 计算总消耗
	totalConsume := 0
	for i, taskReq := range req.Tasks {
		var taskType models.TaskType
		if err := h.db.Where("type_code = ? AND is_active = ?", taskReq.TaskType, true).First(&taskType).Error; err != nil {
			response.Error(c, http.StatusBadRequest, "无效的任务类型: "+taskReq.TaskType)
			return
		}
		// 只有关键词搜索任务需要关键词参数
		if taskReq.TaskType == "search_browse" && taskReq.Keyword == "" {
			response.Error(c, http.StatusBadRequest, "关键词搜索任务必须填写关键词")
			return
		}
		// 非关键词搜索任务不需要关键词，清空它
		if taskReq.TaskType != "search_browse" {
			req.Tasks[i].Keyword = ""
		}
		totalConsume += taskType.JingdouPrice * taskReq.ExecuteCount
	}

	// 检查余额
	isAdmin := user.Role == "admin"
	if !isAdmin && user.JingdouBalance < totalConsume {
		response.Error(c, http.StatusBadRequest, "京豆余额不足")
		return
	}

	// 开始事务
	tx := h.db.Begin()
	createdIDs := make([]uint, 0)
	successCount := 0

	for _, taskReq := range req.Tasks {
		var taskType models.TaskType
		tx.Where("type_code = ?", taskReq.TaskType).First(&taskType)

		consumeJingdou := taskType.JingdouPrice * taskReq.ExecuteCount
		if isAdmin {
			consumeJingdou = 0
		}

		task := models.Task{
			UserID:         user.ID,
			TaskType:       taskReq.TaskType,
			SKU:            taskReq.SKU,
			ShopName:       taskReq.ShopName,
			Keyword:        taskReq.Keyword,
			StartTime:      taskReq.StartTime,
			ExecuteCount:   taskReq.ExecuteCount,
			ExecutedCount:  0,
			Priority:       taskReq.Priority,
			Status:         "waiting",
			ConsumeJingdou: consumeJingdou,
			Remark:         taskReq.Remark,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := tx.Create(&task).Error; err == nil {
			createdIDs = append(createdIDs, task.ID)
			successCount++

			// 非管理员记录每个任务的京豆扣除日志
			if !isAdmin && consumeJingdou > 0 {
				jingdouLog := models.JingdouLog{
					UserID:        user.ID,
					Amount:        -consumeJingdou,
					Balance:       user.JingdouBalance - consumeJingdou, // 估算余额
					OperationType: "task",
					RelatedID:     &task.ID,
					Remark:        "批量创建任务扣除 - SKU:" + task.SKU,
					CreatedAt:     time.Now(),
				}
				tx.Create(&jingdouLog)
			}

			// 更新或创建任务模板（仅普通用户）
			if !isAdmin {
				UpdateOrCreateTemplate(tx, user.ID, task.TaskType, task.SKU, task.ShopName, task.Keyword, task.ExecuteCount)
			}
		}
	}

	// 扣除京豆
	if !isAdmin && totalConsume > 0 {
		user.JingdouBalance -= totalConsume
		tx.Save(&user)
	}

	tx.Commit()

	response.Success(c, gin.H{
		"total_tasks":           len(req.Tasks),
		"successful_tasks":      successCount,
		"failed_tasks":          len(req.Tasks) - successCount,
		"total_consume_jingdou": totalConsume,
		"balance":               user.JingdouBalance,
		"is_admin":              isAdmin,
		"created_task_ids":      createdIDs,
	})
}
