package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
)

type AdminDashboardHandler struct {
	db *gorm.DB
}

func NewAdminDashboardHandler(db *gorm.DB) *AdminDashboardHandler {
	return &AdminDashboardHandler{db: db}
}

// GetTodayTaskStats 获取今日任务统计
// @Summary 获取今日任务统计
// @Description 获取今日任务的统计数据，支持按任务数或任务次数统计，支持按任务类型筛选（仅管理员）
// @Tags 管理员仪表盘
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param stat_mode query string false "统计模式: count(任务数) 或 execute(执行次数)" default(execute)
// @Param task_type query string false "任务类型，空表示全部"
// @Success 200 {object} response.Response{data=object}
// @Failure 401 {object} response.Response
// @Router /admin/dashboard/today-tasks [get]
func (h *AdminDashboardHandler) GetTodayTaskStats(c *gin.Context) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)

	// 获取查询参数
	statMode := c.DefaultQuery("stat_mode", "execute") // count 或 execute
	taskType := c.Query("task_type")

	// 构建基础查询条件
	baseQuery := func() *gorm.DB {
		query := h.db.Model(&models.Task{}).Where("start_time >= ? AND start_time < ?", todayStart, todayEnd)
		if taskType != "" {
			query = query.Where("task_type = ?", taskType)
		}
		return query
	}

	var pendingValue, runningValue, completedValue, totalValue int64

	if statMode == "count" {
		// 按任务数量统计
		baseQuery().Where("status = ?", "waiting").Count(&pendingValue)
		baseQuery().Where("status = ?", "running").Count(&runningValue)
		baseQuery().Where("status = ?", "completed").Count(&completedValue)
		totalValue = pendingValue + runningValue + completedValue
	} else {
		// 按执行次数统计
		type ExecuteStats struct {
			TotalCount    int64
			ExecutedCount int64
		}
		var pendingStats, runningStats, completedStats ExecuteStats

		baseQuery().Where("status = ?", "waiting").
			Select("COALESCE(SUM(execute_count), 0) as total_count, 0 as executed_count").
			Scan(&pendingStats)
		baseQuery().Where("status = ?", "running").
			Select("COALESCE(SUM(execute_count), 0) as total_count, COALESCE(SUM(executed_count), 0) as executed_count").
			Scan(&runningStats)
		baseQuery().Where("status = ?", "completed").
			Select("COALESCE(SUM(execute_count), 0) as total_count, COALESCE(SUM(executed_count), 0) as executed_count").
			Scan(&completedStats)

		pendingValue = pendingStats.TotalCount + (runningStats.TotalCount - runningStats.ExecutedCount)
		runningValue = runningStats.ExecutedCount
		completedValue = completedStats.TotalCount
		totalValue = pendingValue + runningValue + completedValue
	}

	// 计算百分比
	var pendingPercent, completedPercent, runningPercent float64
	if totalValue > 0 {
		pendingPercent = float64(pendingValue) / float64(totalValue) * 100
		completedPercent = float64(completedValue) / float64(totalValue) * 100
		runningPercent = float64(runningValue) / float64(totalValue) * 100
	}

	// 任务数量统计
	var totalTasks int64
	baseQuery().Count(&totalTasks)

	// 统计单位
	unit := "次"
	if statMode == "count" {
		unit = "个"
	}

	response.Success(c, gin.H{
		"stat_mode":         statMode,
		"task_type":         taskType,
		"unit":              unit,
		"total_tasks":       totalTasks,
		"total_value":       totalValue,
		"pending_value":     pendingValue,
		"completed_value":   completedValue,
		"running_value":     runningValue,
		"pending_percent":   pendingPercent,
		"completed_percent": completedPercent,
		"running_percent":   runningPercent,
	})
}

// GetTaskPressure 获取任务执行压力统计
// @Summary 获取任务执行压力统计
// @Description 获取任务执行压力水平，支持按任务数或任务次数统计，支持按任务类型筛选（仅管理员）
// @Tags 管理员仪表盘
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param stat_mode query string false "统计模式: count(任务数) 或 execute(执行次数)" default(execute)
// @Param task_type query string false "任务类型，空表示全部"
// @Success 200 {object} response.Response{data=object}
// @Failure 401 {object} response.Response
// @Router /admin/dashboard/task-pressure [get]
func (h *AdminDashboardHandler) GetTaskPressure(c *gin.Context) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 获取查询参数
	statMode := c.DefaultQuery("stat_mode", "execute")
	taskType := c.Query("task_type")

	// 构建基础查询条件
	baseQueryWithTime := func(start, end time.Time) *gorm.DB {
		query := h.db.Model(&models.Task{}).Where("start_time >= ? AND start_time < ?", start, end)
		if taskType != "" {
			query = query.Where("task_type = ?", taskType)
		}
		return query
	}

	baseQueryFromTime := func(start time.Time) *gorm.DB {
		query := h.db.Model(&models.Task{}).Where("start_time >= ?", start)
		if taskType != "" {
			query = query.Where("task_type = ?", taskType)
		}
		return query
	}

	// 未杧7天待执行的任务（按天分组）
	type DayCount struct {
		Day   string `json:"day"`
		Count int64  `json:"count"`
	}
	var futureTasks []DayCount

	for i := 0; i < 7; i++ {
		dayStart := todayStart.Add(time.Duration(i) * 24 * time.Hour)
		dayEnd := dayStart.Add(24 * time.Hour)
		var count int64

		if statMode == "count" {
			baseQueryWithTime(dayStart, dayEnd).Where("status IN ?", []string{"waiting", "running"}).Count(&count)
		} else {
			// 统计待执行次数 = waiting的execute_count + running的(execute_count - executed_count)
			var waitingCount, runningPending int64
			baseQueryWithTime(dayStart, dayEnd).Where("status = ?", "waiting").
				Select("COALESCE(SUM(execute_count), 0)").Scan(&waitingCount)
			baseQueryWithTime(dayStart, dayEnd).Where("status = ?", "running").
				Select("COALESCE(SUM(execute_count - executed_count), 0)").Scan(&runningPending)
			count = waitingCount + runningPending
		}
		futureTasks = append(futureTasks, DayCount{
			Day:   dayStart.Format("2006-01-02"),
			Count: count,
		})
	}

	// 未来总待执行
	var totalFuturePending int64
	if statMode == "count" {
		baseQueryFromTime(todayStart).Where("status IN ?", []string{"waiting", "running"}).Count(&totalFuturePending)
	} else {
		var waitingCount, runningPending int64
		baseQueryFromTime(todayStart).Where("status = ?", "waiting").
			Select("COALESCE(SUM(execute_count), 0)").Scan(&waitingCount)
		baseQueryFromTime(todayStart).Where("status = ?", "running").
			Select("COALESCE(SUM(execute_count - executed_count), 0)").Scan(&runningPending)
		totalFuturePending = waitingCount + runningPending
	}

	// 昨日完成
	yesterdayStart := todayStart.Add(-24 * time.Hour)
	var yesterdayCompleted int64
	if statMode == "count" {
		baseQueryWithTime(yesterdayStart, todayStart).Where("status = ?", "completed").Count(&yesterdayCompleted)
	} else {
		baseQueryWithTime(yesterdayStart, todayStart).Where("status = ?", "completed").
			Select("COALESCE(SUM(executed_count), 0)").Scan(&yesterdayCompleted)
	}

	// 过去3天平均完成
	threeDaysAgo := todayStart.Add(-72 * time.Hour)
	var last3DaysCompleted int64
	if statMode == "count" {
		baseQueryWithTime(threeDaysAgo, todayStart).Where("status = ?", "completed").Count(&last3DaysCompleted)
	} else {
		baseQueryWithTime(threeDaysAgo, todayStart).Where("status = ?", "completed").
			Select("COALESCE(SUM(executed_count), 0)").Scan(&last3DaysCompleted)
	}
	avgCompleted := float64(last3DaysCompleted) / 3.0

	// 计算执行压力水平
	var pressureLevel string
	var pressureValue float64
	if avgCompleted > 0 {
		pressureValue = float64(totalFuturePending) / (avgCompleted * 7) * 100
		if pressureValue <= 50 {
			pressureLevel = "低"
		} else if pressureValue <= 80 {
			pressureLevel = "中"
		} else if pressureValue <= 100 {
			pressureLevel = "高"
		} else {
			pressureLevel = "超载"
		}
	} else {
		if totalFuturePending > 0 {
			pressureLevel = "无历史数据"
			pressureValue = 100
		} else {
			pressureLevel = "无任务"
			pressureValue = 0
		}
	}

	// 统计单位
	unit := "次"
	if statMode == "count" {
		unit = "个"
	}

	response.Success(c, gin.H{
		"stat_mode":            statMode,
		"task_type":            taskType,
		"unit":                 unit,
		"future_tasks":         futureTasks,
		"total_future_pending": totalFuturePending,
		"yesterday_completed":  yesterdayCompleted,
		"avg_3days_completed":  avgCompleted,
		"pressure_level":       pressureLevel,
		"pressure_value":       pressureValue,
	})
}

// GetFinanceStats 获取财务统计
// @Summary 获取财务统计
// @Description 获取财务统计数据，包括每日平均充值、消耗京豆，以及京豆最低的用户列表（仅管理员）
// @Tags 管理员仪表板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Failure 401 {object} response.Response
// @Router /admin/dashboard/finance [get]
func (h *AdminDashboardHandler) GetFinanceStats(c *gin.Context) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 过去30天的数据
	thirtyDaysAgo := todayStart.Add(-30 * 24 * time.Hour)

	// 计算每日平均充值京豆
	var totalRecharge int64
	h.db.Model(&models.JingdouLog{}).
		Where("created_at >= ? AND change_type = ?", thirtyDaysAgo, "recharge").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalRecharge)
	avgDailyRecharge := float64(totalRecharge) / 30.0

	// 计算每日平均消耗京豆
	var totalConsume int64
	h.db.Model(&models.JingdouLog{}).
		Where("created_at >= ? AND change_type = ?", thirtyDaysAgo, "consume").
		Select("COALESCE(SUM(ABS(amount)), 0)").
		Scan(&totalConsume)
	avgDailyConsume := float64(totalConsume) / 30.0

	// 京豆最低的10名用户
	type LowBalanceUser struct {
		ID             uint   `json:"id"`
		Username       string `json:"username"`
		Nickname       string `json:"nickname"`
		JingdouBalance int    `json:"jingdou_balance"`
	}
	var lowBalanceUsers []LowBalanceUser
	result := h.db.Model(&models.User{}).
		Select("id, username, nickname, jingdou_balance").
		Where("role != ?", "admin").
		Order("jingdou_balance ASC").
		Limit(10).
		Scan(&lowBalanceUsers)

	// 如果查询出错，设置为空数组而不是nil
	if result.Error != nil {
		lowBalanceUsers = []LowBalanceUser{}
	}

	// 今日充值总额
	var todayRecharge int64
	h.db.Model(&models.JingdouLog{}).
		Where("created_at >= ? AND change_type = ?", todayStart, "recharge").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&todayRecharge)

	// 今日消耗总额
	var todayConsume int64
	h.db.Model(&models.JingdouLog{}).
		Where("created_at >= ? AND change_type = ?", todayStart, "consume").
		Select("COALESCE(SUM(ABS(amount)), 0)").
		Scan(&todayConsume)

	response.Success(c, gin.H{
		"avg_daily_recharge": avgDailyRecharge,
		"avg_daily_consume":  avgDailyConsume,
		"today_recharge":     todayRecharge,
		"today_consume":      todayConsume,
		"low_balance_users":  lowBalanceUsers, // 确保即使为空也返回空数组而不是nil
	})
}

// TriggerExpiredTaskCheck 手动触发过期任务检查
// @Summary 手动触发过期任务检查
// @Description 手动触发过期任务检查和退款处理（仅管理员）
// @Tags 管理员仪表板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Failure 401 {object} response.Response
// @Router /admin/dashboard/trigger-expire-check [post]
func (h *AdminDashboardHandler) TriggerExpiredTaskCheck(c *gin.Context) {
	// 计算24小时前的时间点
	expireThreshold := time.Now().Add(-24 * time.Hour)

	// 查找过期任务
	var expiredTasks []models.Task
	if err := h.db.Where(
		"start_time < ? AND status IN (?, ?) AND executed_count < execute_count",
		expireThreshold, "waiting", "running",
	).Find(&expiredTasks).Error; err != nil {
		response.Error(c, 500, "查询过期任务失败: "+err.Error())
		return
	}

	if len(expiredTasks) == 0 {
		response.Success(c, gin.H{
			"message":         "没有发现过期任务",
			"processed_count": 0,
		})
		return
	}

	processedCount := 0
	var processedTasks []gin.H

	for _, task := range expiredTasks {
		result := h.processExpiredTask(&task)
		if result != nil {
			processedTasks = append(processedTasks, result)
			processedCount++
		}
	}

	response.Success(c, gin.H{
		"message":         fmt.Sprintf("已处理 %d 个过期任务", processedCount),
		"processed_count": processedCount,
		"processed_tasks": processedTasks,
	})
}

// processExpiredTask 处理单个过期任务
func (h *AdminDashboardHandler) processExpiredTask(task *models.Task) gin.H {
	tx := h.db.Begin()

	// 获取任务类型价格
	var taskType models.TaskType
	if err := tx.Where("type_code = ?", task.TaskType).First(&taskType).Error; err != nil {
		tx.Rollback()
		return nil
	}

	// 计算退款金额
	remainingCount := task.ExecuteCount - task.ExecutedCount
	refundAmount := 0

	if task.ConsumeJingdou > 0 && task.ExecuteCount > 0 {
		refundAmount = task.ConsumeJingdou * remainingCount / task.ExecuteCount
	}

	// 获取用户
	var user models.User
	if err := tx.First(&user, task.UserID).Error; err != nil {
		tx.Rollback()
		return nil
	}

	// 更新用户余额
	if refundAmount > 0 {
		user.JingdouBalance += refundAmount
		tx.Save(&user)

		// 创建京豆日志
		jingdouLog := models.JingdouLog{
			UserID:        user.ID,
			Amount:        refundAmount,
			Balance:       user.JingdouBalance,
			OperationType: "refund",
			RelatedID:     &task.ID,
			Remark:        fmt.Sprintf("任务过期自动退款 - SKU:%s (完成%d/%d)", task.SKU, task.ExecutedCount, task.ExecuteCount),
			CreatedAt:     time.Now(),
		}
		tx.Create(&jingdouLog)
	}

	// 更新任务状态
	oldStatus := task.Status
	task.Status = "partial_completed"
	task.UpdatedAt = time.Now()

	expireRemark := fmt.Sprintf("【系统自动处理】任务过期，完成%d/%d次", task.ExecutedCount, task.ExecuteCount)
	if refundAmount > 0 {
		expireRemark += fmt.Sprintf("，退还%d京豆", refundAmount)
	}
	if task.Remark != "" {
		task.Remark = task.Remark + " | " + expireRemark
	} else {
		task.Remark = expireRemark
	}
	tx.Save(task)

	// 创建任务日志
	taskLog := models.TaskLog{
		TaskID:    task.ID,
		Status:    "partial_completed",
		Message:   expireRemark,
		CreatedAt: time.Now(),
	}
	tx.Create(&taskLog)

	tx.Commit()

	return gin.H{
		"task_id":        task.ID,
		"sku":            task.SKU,
		"old_status":     oldStatus,
		"new_status":     "partial_completed",
		"executed":       task.ExecutedCount,
		"total":          task.ExecuteCount,
		"refund_jingdou": refundAmount,
	}
}

// TriggerDataCleanup 手动触发数据清理
// @Summary 手动触发数据清理
// @Description 手动触发清理过期数据，默认60天保留期（仅管理员）
// @Tags 管理员仪表板
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param retention_days query int false "保留天数（默认60）"
// @Success 200 {object} response.Response{data=object}
// @Failure 401 {object} response.Response
// @Router /admin/dashboard/trigger-cleanup [post]
func (h *AdminDashboardHandler) TriggerDataCleanup(c *gin.Context) {
	// 获取保留天数参数，默认60天
	retentionDays := 60
	if days := c.Query("retention_days"); days != "" {
		if d, err := strconv.Atoi(days); err == nil && d > 0 {
			retentionDays = d
		}
	}

	// 计算清理阈值时间
	threshold := time.Now().AddDate(0, 0, -retentionDays)

	// 统计清理结果
	var tasksDeleted, taskLogsDeleted, deviceHistoryDeleted, apiLogsDeleted int64

	// 只清理已完成/已取消/部分完成/失败的任务
	safeStatuses := []string{"completed", "cancelled", "partial_completed", "failed"}

	// 1. 查询要删除的任务ID（用于删除关联日志）
	var taskIDs []uint
	h.db.Model(&models.Task{}).
		Where("created_at < ? AND status IN ?", threshold, safeStatuses).
		Pluck("id", &taskIDs)

	// 2. 删除任务日志
	if len(taskIDs) > 0 {
		result := h.db.Where("task_id IN ?", taskIDs).Delete(&models.TaskLog{})
		taskLogsDeleted = result.RowsAffected
	}

	// 3. 删除任务
	result := h.db.Where("created_at < ? AND status IN ?", threshold, safeStatuses).Delete(&models.Task{})
	tasksDeleted = result.RowsAffected

	// 4. 删除设备任务历史
	result = h.db.Where("execute_time < ?", threshold).Delete(&models.DeviceTaskHistory{})
	deviceHistoryDeleted = result.RowsAffected

	// 5. 删除API日志
	result = h.db.Where("created_at < ?", threshold).Delete(&models.APILog{})
	apiLogsDeleted = result.RowsAffected

	response.Success(c, gin.H{
		"message":                fmt.Sprintf("已清理 %s 之前的数据", threshold.Format("2006-01-02")),
		"retention_days":         retentionDays,
		"threshold_date":         threshold.Format("2006-01-02"),
		"tasks_deleted":          tasksDeleted,
		"task_logs_deleted":      taskLogsDeleted,
		"device_history_deleted": deviceHistoryDeleted,
		"api_logs_deleted":       apiLogsDeleted,
		"total_deleted":          tasksDeleted + taskLogsDeleted + deviceHistoryDeleted + apiLogsDeleted,
	})
}
