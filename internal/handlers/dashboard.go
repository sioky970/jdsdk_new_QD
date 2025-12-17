package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// GetOverview 获取概览数据
// @Summary 获取概览数据
// @Description 获取系统概览数据，包括用户、任务、设备总数
// @Tags 仪表板模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Failure 401 {object} response.Response
// @Router /dashboard/overview [get]
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	var totalUsers int64
	var totalTasks int64
	var totalDevices int64

	h.db.Model(&models.User{}).Count(&totalUsers)
	h.db.Model(&models.Task{}).Count(&totalTasks)
	h.db.Model(&models.Device{}).Count(&totalDevices)

	response.Success(c, gin.H{
		"total_users":   totalUsers,
		"total_tasks":   totalTasks,
		"total_devices": totalDevices,
	})
}

// GetStatistics 获取统计数据
// @Summary 获取统计数据
// @Description 获取详细的统计数据，包括任务状态、设备状态等
// @Tags 仪表板模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Failure 401 {object} response.Response
// @Router /dashboard/statistics [get]
func (h *DashboardHandler) GetStatistics(c *gin.Context) {
	var stats struct {
		TotalTasks     int64 `json:"total_tasks"`
		RunningTasks   int64 `json:"running_tasks"`
		WaitingTasks   int64 `json:"waiting_tasks"`
		CompletedTasks int64 `json:"completed_tasks"`
		OnlineDevices  int64 `json:"online_devices"`
		ActiveUsers    int64 `json:"active_users"`
	}

	h.db.Model(&models.Task{}).Count(&stats.TotalTasks)
	h.db.Model(&models.Task{}).Where("status = ?", "running").Count(&stats.RunningTasks)
	h.db.Model(&models.Task{}).Where("status = ?", "waiting").Count(&stats.WaitingTasks)
	h.db.Model(&models.Task{}).Where("status = ?", "completed").Count(&stats.CompletedTasks)
	h.db.Model(&models.Device{}).Where("status = ?", "online").Count(&stats.OnlineDevices)
	h.db.Model(&models.User{}).Where("is_active = ?", true).Count(&stats.ActiveUsers)

	response.Success(c, stats)
}
