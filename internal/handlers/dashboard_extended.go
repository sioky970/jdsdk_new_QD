package handlers

import (
	"time"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetDetailedStatistics 获取详细统计数据（管理员）
// @Summary 获取详细统计数据
// @Description 获取系统详细统计数据，包括趋势、分布等（仅管理员）
// @Tags 仪表板模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Router /dashboard/stat/details [get]
func (h *DashboardHandler) GetDetailedStatistics(c *gin.Context) {
	// 统计卡片数据
	var totalTasks, totalUsers, totalDevices int64
	h.db.Model(&models.Task{}).Count(&totalTasks)
	h.db.Model(&models.User{}).Count(&totalUsers)
	h.db.Model(&models.Device{}).Count(&totalDevices)

	var totalConsumed int
	h.db.Model(&models.JingdouLog{}).Where("operation_type = ?", "task").Select("SUM(ABS(amount))").Row().Scan(&totalConsumed)

	statCards := []gin.H{
		{
			"title":      "任务总数",
			"value":      totalTasks,
			"color":      "#409EFF",
			"trending":   "up",
			"trendValue": 15.5,
		},
		{
			"title":      "用户总数",
			"value":      totalUsers,
			"color":      "#67C23A",
			"trending":   "up",
			"trendValue": 8.2,
		},
		{
			"title":      "设备总数",
			"value":      totalDevices,
			"color":      "#E6A23C",
			"trending":   "down",
			"trendValue": 2.1,
		},
		{
			"title":      "京豆消耗",
			"value":      totalConsumed,
			"color":      "#F56C6C",
			"trending":   "up",
			"trendValue": 12.3,
		},
	}

	// 任务类型分布
	var taskTypeDistribution []struct {
		TaskType string `json:"task_type"`
		Count    int64  `json:"count"`
	}
	h.db.Model(&models.Task{}).Select("task_type, COUNT(*) as count").Group("task_type").Scan(&taskTypeDistribution)

	taskTypeDist := make([]gin.H, 0)
	for _, item := range taskTypeDistribution {
		typeName := item.TaskType
		switch item.TaskType {
		case "add_to_cart":
			typeName = "加购"
		case "collect_product":
			typeName = "收藏商品"
		case "click":
			typeName = "点击"
		case "keyword_click":
			typeName = "关键词点击"
		case "collect_shop":
			typeName = "收藏店铺"
		}
		taskTypeDist = append(taskTypeDist, gin.H{
			"name":  typeName,
			"value": item.Count,
		})
	}

	// 任务状态分布
	var taskStatusDistribution []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	h.db.Model(&models.Task{}).Select("status, COUNT(*) as count").Group("status").Scan(&taskStatusDistribution)

	taskStatusDist := make([]gin.H, 0)
	for _, item := range taskStatusDistribution {
		statusName := item.Status
		switch item.Status {
		case "waiting":
			statusName = "待执行"
		case "running":
			statusName = "执行中"
		case "completed":
			statusName = "已完成"
		case "failed":
			statusName = "失败"
		case "cancelled":
			statusName = "已取消"
		}
		taskStatusDist = append(taskStatusDist, gin.H{
			"name":  statusName,
			"value": item.Count,
		})
	}

	// 趋势数据（最近7天）
	dates := make([]string, 7)
	published := make([]int64, 7)
	executed := make([]int64, 7)

	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, -6+i)
		dates[i] = date.Format("01-02")
		dateStr := date.Format("2006-01-02")

		var pub, exec int64
		h.db.Model(&models.Task{}).Where("DATE(created_at) = ?", dateStr).Count(&pub)
		h.db.Model(&models.Task{}).Where("DATE(updated_at) = ? AND status IN (?)", dateStr, []string{"completed", "failed"}).Count(&exec)

		published[i] = pub
		executed[i] = exec
	}

	response.Success(c, gin.H{
		"stat_cards":               statCards,
		"task_type_distribution":   taskTypeDist,
		"task_status_distribution": taskStatusDist,
		"trend_data": gin.H{
			"dates":     dates,
			"published": published,
			"executed":  executed,
		},
		"executor_stats": []gin.H{}, // 简化版本，可后续扩展
	})
}

// GetFutureTrends 获取未来趋势数据
// @Summary 获取未来趋势数据
// @Description 获取未来7天的任务和京豆消耗预测
// @Tags 仪表板模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Router /dashboard/future-trends [get]
func (h *DashboardHandler) GetFutureTrends(c *gin.Context) {
	dates := make([]string, 7)
	futureTasks := make([]int64, 7)
	jingdouConsumption := make([]int, 7)

	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, i)
		dates[i] = date.Format("01-02")
		dateStr := date.Format("2006-01-02")

		// 统计该日期的待执行任务
		var taskCount int64
		h.db.Model(&models.Task{}).Where("DATE(start_time) = ? AND status IN (?)", dateStr, []string{"waiting", "running"}).Count(&taskCount)
		futureTasks[i] = taskCount

		// 计算预估京豆消耗
		var totalConsume int
		h.db.Model(&models.Task{}).Where("DATE(start_time) = ? AND status IN (?)", dateStr, []string{"waiting", "running"}).Select("SUM(consume_jingdou)").Row().Scan(&totalConsume)
		jingdouConsumption[i] = totalConsume
	}

	response.Success(c, gin.H{
		"dates":               dates,
		"future_tasks":        futureTasks,
		"jingdou_consumption": jingdouConsumption,
	})
}
