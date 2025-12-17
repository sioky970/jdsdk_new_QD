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

// OpenAPIHandler 开放API处理器（API Key认证）
type OpenAPIHandler struct {
	db *gorm.DB
}

// NewOpenAPIHandler 创建开放API处理器
func NewOpenAPIHandler(db *gorm.DB) *OpenAPIHandler {
	return &OpenAPIHandler{db: db}
}

// =========================================
// 任务创建相关接口
// =========================================

// CreateTaskRequest 创建任务请求（开放API用）
type OpenAPICreateTaskRequest struct {
	TaskType     string    `json:"task_type" binding:"required"`     // 任务类型
	SKU          string    `json:"sku" binding:"required"`           // 商品SKU
	ShopName     string    `json:"shop_name"`                        // 店铺名称
	Keyword      string    `json:"keyword"`                          // 关键词
	StartTime    time.Time `json:"start_time" binding:"required"`    // 开始执行时间
	ExecuteCount int       `json:"execute_count" binding:"required"` // 执行次数
	Priority     int       `json:"priority"`                         // 优先级
	Remark       string    `json:"remark"`                           // 备注
}

// CreateTask 创建单个任务
// @Summary 创建单个任务（API Key）
// @Description 使用API Key创建单个任务
// @Tags 开放API-任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body OpenAPICreateTaskRequest true "任务信息"
// @Success 201 {object} response.Response{data=object}
// @Failure 400 {object} response.Response
// @Router /openapi/tasks [post]
func (h *OpenAPIHandler) CreateTask(c *gin.Context) {
	var req OpenAPICreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误：请检查任务类型(task_type)、商品SKU(sku)、开始时间(start_time)和执行次数(execute_count)是否填写正确")
		return
	}

	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在：请检查您的API Key是否有效")
		return
	}

	// 查询任务类型配置
	var taskType models.TaskType
	if err := h.db.Where("type_code = ?", req.TaskType).First(&taskType).Error; err != nil {
		response.Error(c, http.StatusBadRequest, "任务类型不存在：请使用有效的任务类型代码，可通过 GET /openapi/task-types 获取可用类型")
		return
	}

	// 检查任务类型是否启用
	if !taskType.IsActive {
		response.Error(c, http.StatusBadRequest, "该任务类型已被禁用：当前任务类型暂时不可使用，请选择其他类型或联系管理员")
		return
	}

	// 关键词搜索任务必须填写关键词
	if req.TaskType == "search_browse" && req.Keyword == "" {
		response.Error(c, http.StatusBadRequest, "参数错误：关键词搜索任务(search_browse)必须填写关键词(keyword)字段")
		return
	}
	// 非关键词搜索任务不需要关键词
	if req.TaskType != "search_browse" {
		req.Keyword = ""
	}

	// 检查执行次数
	if req.ExecuteCount <= 0 {
		response.Error(c, http.StatusBadRequest, "参数错误：执行次数(execute_count)必须大于0")
		return
	}

	// 检查时间段限制（使用API创建的任务也需要遵守时间限制）
	if taskType.TimeSlot1Start != nil && taskType.TimeSlot1End != nil &&
		*taskType.TimeSlot1Start != "" && *taskType.TimeSlot1End != "" {
		now := time.Now()
		currentTime := now.Format("15:04")

		inTimeSlot := false
		if currentTime >= *taskType.TimeSlot1Start && currentTime <= *taskType.TimeSlot1End {
			inTimeSlot = true
		}
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
			response.Error(c, http.StatusBadRequest, "创建失败：该任务类型仅在 "+timeSlotInfo+" 时段内允许创建任务")
			return
		}
	}

	// 计算京豆消耗
	consumeJingdou := taskType.JingdouPrice * req.ExecuteCount

	// 检查余额
	if user.JingdouBalance < consumeJingdou {
		response.Error(c, http.StatusBadRequest,
			"京豆余额不足：创建此任务需要 "+strconv.Itoa(consumeJingdou)+" 京豆，您当前余额为 "+strconv.Itoa(user.JingdouBalance)+" 京豆，请先充值")
		return
	}

	// 创建任务
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
		response.Error(c, http.StatusInternalServerError, "任务创建失败：系统内部错误，请稍后重试")
		return
	}

	// 扣除京豆
	user.JingdouBalance -= consumeJingdou
	tx.Save(&user)

	// 创建京豆日志
	jingdouLog := models.JingdouLog{
		UserID:        user.ID,
		Amount:        -consumeJingdou,
		Balance:       user.JingdouBalance,
		OperationType: "task",
		RelatedID:     &task.ID,
		Remark:        "API创建任务扣除 - SKU:" + task.SKU,
		CreatedAt:     time.Now(),
	}
	tx.Create(&jingdouLog)

	// 更新模板
	UpdateOrCreateTemplate(tx, user.ID, task.TaskType, task.SKU, task.ShopName, task.Keyword, task.ExecuteCount)

	tx.Commit()

	response.Success(c, gin.H{
		"task_id":         task.ID,
		"task_type":       task.TaskType,
		"sku":             task.SKU,
		"status":          task.Status,
		"consume_jingdou": consumeJingdou,
		"balance":         user.JingdouBalance,
		"created_at":      task.CreatedAt.Format(time.RFC3339),
		"message":         "任务创建成功",
	})
}

// BatchCreateTaskRequest 批量创建任务请求
type OpenAPIBatchCreateRequest struct {
	Tasks []OpenAPICreateTaskRequest `json:"tasks" binding:"required"`
}

// BatchCreateTasks 批量创建任务
// @Summary 批量创建任务（API Key）
// @Description 使用API Key批量创建任务，最多100个
// @Tags 开放API-任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body OpenAPIBatchCreateRequest true "批量任务信息"
// @Success 201 {object} response.Response{data=object}
// @Failure 400 {object} response.Response
// @Router /openapi/tasks/batch [post]
func (h *OpenAPIHandler) BatchCreateTasks(c *gin.Context) {
	var req OpenAPIBatchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误：请检查请求体格式，tasks字段必须是任务数组")
		return
	}

	if len(req.Tasks) == 0 {
		response.Error(c, http.StatusBadRequest, "任务列表为空：请至少提供一个任务")
		return
	}

	if len(req.Tasks) > 100 {
		response.Error(c, http.StatusBadRequest, "超出限制：单次最多创建100个任务，请分批提交")
		return
	}

	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在：请检查您的API Key是否有效")
		return
	}

	// 预先验证所有任务并计算总消耗
	totalConsume := 0
	taskTypeCache := make(map[string]models.TaskType)

	for i, taskReq := range req.Tasks {
		// 查询任务类型
		var taskType models.TaskType
		if cached, ok := taskTypeCache[taskReq.TaskType]; ok {
			taskType = cached
		} else {
			if err := h.db.Where("type_code = ? AND is_active = ?", taskReq.TaskType, true).First(&taskType).Error; err != nil {
				response.Error(c, http.StatusBadRequest,
					"第 "+strconv.Itoa(i+1)+" 个任务的任务类型无效："+taskReq.TaskType+" 不存在或已被禁用")
				return
			}
			taskTypeCache[taskReq.TaskType] = taskType
		}

		// 验证关键词
		if taskReq.TaskType == "search_browse" && taskReq.Keyword == "" {
			response.Error(c, http.StatusBadRequest,
				"第 "+strconv.Itoa(i+1)+" 个任务参数错误：关键词搜索任务必须填写关键词")
			return
		}

		// 验证执行次数
		if taskReq.ExecuteCount <= 0 {
			response.Error(c, http.StatusBadRequest,
				"第 "+strconv.Itoa(i+1)+" 个任务参数错误：执行次数必须大于0")
			return
		}

		totalConsume += taskType.JingdouPrice * taskReq.ExecuteCount
	}

	// 检查余额
	if user.JingdouBalance < totalConsume {
		response.Error(c, http.StatusBadRequest,
			"京豆余额不足：创建这 "+strconv.Itoa(len(req.Tasks))+" 个任务共需要 "+strconv.Itoa(totalConsume)+
				" 京豆，您当前余额为 "+strconv.Itoa(user.JingdouBalance)+" 京豆，请先充值")
		return
	}

	// 开始事务创建任务
	tx := h.db.Begin()
	createdTasks := make([]gin.H, 0)
	failedTasks := make([]gin.H, 0)
	successCount := 0
	actualConsume := 0

	for i, taskReq := range req.Tasks {
		taskType := taskTypeCache[taskReq.TaskType]

		// 非关键词搜索任务清空关键词
		keyword := taskReq.Keyword
		if taskReq.TaskType != "search_browse" {
			keyword = ""
		}

		consumeJingdou := taskType.JingdouPrice * taskReq.ExecuteCount

		task := models.Task{
			UserID:         user.ID,
			TaskType:       taskReq.TaskType,
			SKU:            taskReq.SKU,
			ShopName:       taskReq.ShopName,
			Keyword:        keyword,
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

		if err := tx.Create(&task).Error; err != nil {
			failedTasks = append(failedTasks, gin.H{
				"index":  i + 1,
				"sku":    taskReq.SKU,
				"reason": "创建失败：" + err.Error(),
			})
			continue
		}

		// 创建京豆日志
		jingdouLog := models.JingdouLog{
			UserID:        user.ID,
			Amount:        -consumeJingdou,
			Balance:       user.JingdouBalance - actualConsume - consumeJingdou,
			OperationType: "task",
			RelatedID:     &task.ID,
			Remark:        "API批量创建任务扣除 - SKU:" + task.SKU,
			CreatedAt:     time.Now(),
		}
		tx.Create(&jingdouLog)

		// 更新模板
		UpdateOrCreateTemplate(tx, user.ID, task.TaskType, task.SKU, task.ShopName, task.Keyword, task.ExecuteCount)

		createdTasks = append(createdTasks, gin.H{
			"task_id":         task.ID,
			"sku":             task.SKU,
			"consume_jingdou": consumeJingdou,
		})
		successCount++
		actualConsume += consumeJingdou
	}

	// 扣除京豆
	if actualConsume > 0 {
		user.JingdouBalance -= actualConsume
		tx.Save(&user)
	}

	tx.Commit()

	response.Success(c, gin.H{
		"total_submitted": len(req.Tasks),
		"success_count":   successCount,
		"failed_count":    len(failedTasks),
		"total_consume":   actualConsume,
		"balance":         user.JingdouBalance,
		"created_tasks":   createdTasks,
		"failed_tasks":    failedTasks,
		"message":         "批量创建完成：成功 " + strconv.Itoa(successCount) + " 个，失败 " + strconv.Itoa(len(failedTasks)) + " 个",
	})
}

// =========================================
// 任务查询相关接口
// =========================================

// GetTasks 查询任务列表
// @Summary 查询任务列表（API Key）
// @Description 使用API Key查询任务列表，支持检索和分页
// @Tags 开放API-任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "任务状态(waiting/running/completed/failed/cancelled)"
// @Param task_type query string false "任务类型"
// @Param sku query string false "商品SKU（模糊搜索）"
// @Param shop_name query string false "店铺名称（模糊搜索）"
// @Param keyword query string false "关键词（模糊搜索）"
// @Param start_date query string false "开始日期(YYYY-MM-DD)"
// @Param end_date query string false "结束日期(YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=object}
// @Router /openapi/tasks [get]
func (h *OpenAPIHandler) GetTasks(c *gin.Context) {
	userID, _ := c.Get("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := h.db.Model(&models.Task{}).Where("user_id = ?", userID)

	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 任务类型筛选
	if taskType := c.Query("task_type"); taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}

	// SKU模糊搜索
	if sku := c.Query("sku"); sku != "" {
		query = query.Where("sku LIKE ?", "%"+sku+"%")
	}

	// 店铺名称模糊搜索
	if shopName := c.Query("shop_name"); shopName != "" {
		query = query.Where("shop_name LIKE ?", "%"+shopName+"%")
	}

	// 关键词模糊搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("keyword LIKE ?", "%"+keyword+"%")
	}

	// 日期范围筛选
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	// 计算总数
	var total int64
	query.Count(&total)

	// 分页查询
	var tasks []models.Task
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks)

	// 构建响应
	items := make([]gin.H, 0)
	for _, task := range tasks {
		items = append(items, gin.H{
			"id":              task.ID,
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

	response.Success(c, gin.H{
		"items":     items,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"pages":     (int(total) + pageSize - 1) / pageSize,
	})
}

// GetTaskByID 查询单个任务详情
// @Summary 查询任务详情（API Key）
// @Description 使用API Key查询单个任务的详细信息
// @Tags 开放API-任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response{data=object}
// @Failure 404 {object} response.Response
// @Router /openapi/tasks/{id} [get]
func (h *OpenAPIHandler) GetTaskByID(c *gin.Context) {
	userID, _ := c.Get("user_id")
	taskID := c.Param("id")

	var task models.Task
	if err := h.db.First(&task, taskID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "任务不存在：未找到ID为 "+taskID+" 的任务")
		return
	}

	// 验证权限
	if task.UserID != userID.(uint) {
		response.Error(c, http.StatusForbidden, "无权访问：您没有权限查看此任务")
		return
	}

	response.Success(c, gin.H{
		"id":              task.ID,
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

// GetTaskTypes 获取可用任务类型
// @Summary 获取任务类型列表（API Key）
// @Description 获取所有可用的任务类型及其价格
// @Tags 开放API-任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=object}
// @Router /openapi/task-types [get]
func (h *OpenAPIHandler) GetTaskTypes(c *gin.Context) {
	var taskTypes []models.TaskType
	h.db.Where("is_active = ?", true).Order("id ASC").Find(&taskTypes)

	items := make([]gin.H, 0)
	for _, tt := range taskTypes {
		item := gin.H{
			"type_code":     tt.TypeCode,
			"type_name":     tt.TypeName,
			"jingdou_price": tt.JingdouPrice,
		}

		// 添加时间限制信息
		if tt.TimeSlot1Start != nil && tt.TimeSlot1End != nil {
			timeSlots := []string{*tt.TimeSlot1Start + "-" + *tt.TimeSlot1End}
			if tt.TimeSlot2Start != nil && tt.TimeSlot2End != nil {
				timeSlots = append(timeSlots, *tt.TimeSlot2Start+"-"+*tt.TimeSlot2End)
			}
			item["time_slots"] = timeSlots
			item["has_time_limit"] = true
		} else {
			item["time_slots"] = []string{}
			item["has_time_limit"] = false
		}

		items = append(items, item)
	}

	response.Success(c, gin.H{
		"task_types": items,
		"total":      len(items),
	})
}

// =========================================
// 余额和明细查询相关接口
// =========================================

// GetBalance 查询京豆余额
// @Summary 查询京豆余额（API Key）
// @Description 使用API Key查询当前账户的京豆余额
// @Tags 开放API-京豆
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=object}
// @Router /openapi/balance [get]
func (h *OpenAPIHandler) GetBalance(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在：请检查您的API Key是否有效")
		return
	}

	// 获取最后一次京豆变动记录
	var lastLog models.JingdouLog
	h.db.Where("user_id = ?", userID).Order("created_at DESC").First(&lastLog)

	lastUpdated := ""
	if lastLog.ID > 0 {
		lastUpdated = lastLog.CreatedAt.Format(time.RFC3339)
	}

	response.Success(c, gin.H{
		"user_id":         user.ID,
		"username":        username,
		"jingdou_balance": user.JingdouBalance,
		"last_updated":    lastUpdated,
	})
}

// GetJingdouRecords 查询京豆明细
// @Summary 查询京豆明细（API Key）
// @Description 使用API Key查询京豆变动明细
// @Tags 开放API-京豆
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param type query string false "类型(task/recharge/refund/admin)"
// @Param start_date query string false "开始日期(YYYY-MM-DD)"
// @Param end_date query string false "结束日期(YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=object}
// @Router /openapi/jingdou/records [get]
func (h *OpenAPIHandler) GetJingdouRecords(c *gin.Context) {
	userID, _ := c.Get("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := h.db.Model(&models.JingdouLog{}).Where("user_id = ?", userID)

	// 类型筛选
	if opType := c.Query("type"); opType != "" {
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
			"amount":         log.Amount,
			"balance":        log.Balance,
			"operation_type": log.OperationType,
			"related_id":     log.RelatedID,
			"remark":         log.Remark,
			"created_at":     log.CreatedAt.Format(time.RFC3339),
		})
	}

	response.Success(c, gin.H{
		"items":     items,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"pages":     (int(total) + pageSize - 1) / pageSize,
	})
}

// =========================================
// 任务修改和取消相关接口
// =========================================

// UpdateTaskRequest 更新任务请求（开放API用）
type OpenAPIUpdateTaskRequest struct {
	ShopName *string `json:"shop_name"` // 店铺名称
	Keyword  *string `json:"keyword"`   // 关键词
	Priority *int    `json:"priority"`  // 优先级
	Remark   *string `json:"remark"`    // 备注
}

// UpdateTask 修改任务
// @Summary 修改任务（API Key）
// @Description 使用API Key修改指定任务（只能修改等待中的任务）
// @Tags 开放API-任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "任务ID"
// @Param request body OpenAPIUpdateTaskRequest true "更新信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /openapi/tasks/{id} [put]
func (h *OpenAPIHandler) UpdateTask(c *gin.Context) {
	userID, _ := c.Get("user_id")
	taskID := c.Param("id")

	var req OpenAPIUpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误：请检查请求体格式")
		return
	}

	// 检查是否有任何更新
	if req.ShopName == nil && req.Keyword == nil && req.Priority == nil && req.Remark == nil {
		response.Error(c, http.StatusBadRequest, "参数错误：请至少提供一个需要更新的字段（shop_name/keyword/priority/remark）")
		return
	}

	var task models.Task
	if err := h.db.First(&task, taskID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "任务不存在：未找到ID为 "+taskID+" 的任务")
		return
	}

	// 验证权限
	if task.UserID != userID.(uint) {
		response.Error(c, http.StatusForbidden, "无权操作：您没有权限修改此任务")
		return
	}

	// 只能修改等待中的任务
	if task.Status != "waiting" {
		statusMap := map[string]string{
			"running":   "执行中",
			"completed": "已完成",
			"failed":    "已失败",
			"cancelled": "已取消",
		}
		statusName := statusMap[task.Status]
		if statusName == "" {
			statusName = task.Status
		}
		response.Error(c, http.StatusBadRequest, "无法修改：只能修改等待中的任务，当前任务状态为"+statusName)
		return
	}

	// 更新字段
	if req.ShopName != nil {
		task.ShopName = *req.ShopName
	}
	if req.Keyword != nil {
		// 检查是否是关键词搜索任务
		if task.TaskType != "search_browse" && *req.Keyword != "" {
			response.Error(c, http.StatusBadRequest, "参数错误：只有关键词搜索任务(search_browse)可以设置关键词")
			return
		}
		task.Keyword = *req.Keyword
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Remark != nil {
		task.Remark = *req.Remark
	}

	task.UpdatedAt = time.Now()
	h.db.Save(&task)

	response.SuccessWithMsg(c, "任务更新成功", gin.H{
		"task_id":    task.ID,
		"shop_name":  task.ShopName,
		"keyword":    task.Keyword,
		"priority":   task.Priority,
		"remark":     task.Remark,
		"updated_at": task.UpdatedAt.Format(time.RFC3339),
	})
}

// CancelTask 取消任务
// @Summary 取消任务（API Key）
// @Description 使用API Key取消指定任务（只能取消等待中的任务，会退还京豆）
// @Tags 开放API-任务
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response{data=object}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /openapi/tasks/{id}/cancel [post]
func (h *OpenAPIHandler) CancelTask(c *gin.Context) {
	userID, _ := c.Get("user_id")
	taskID := c.Param("id")

	var task models.Task
	if err := h.db.First(&task, taskID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "任务不存在：未找到ID为 "+taskID+" 的任务")
		return
	}

	// 验证权限
	if task.UserID != userID.(uint) {
		response.Error(c, http.StatusForbidden, "无权操作：您没有权限取消此任务")
		return
	}

	// 只能取消等待中的任务
	if task.Status != "waiting" {
		statusMap := map[string]string{
			"running":   "执行中",
			"completed": "已完成",
			"failed":    "已失败",
			"cancelled": "已取消",
		}
		statusName := statusMap[task.Status]
		if statusName == "" {
			statusName = task.Status
		}
		response.Error(c, http.StatusBadRequest, "无法取消：只能取消等待中的任务，当前任务状态为"+statusName)
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
	task.UpdatedAt = time.Now()
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
			Remark:        "API取消任务退款 - SKU:" + task.SKU,
			CreatedAt:     time.Now(),
		}
		tx.Create(&jingdouLog)
	}

	tx.Commit()

	response.SuccessWithMsg(c, "任务取消成功，京豆已退还", gin.H{
		"task_id":        task.ID,
		"refund_jingdou": refundAmount,
		"balance":        user.JingdouBalance,
	})
}
