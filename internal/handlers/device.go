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

type DeviceHandler struct {
	db *gorm.DB
}

func NewDeviceHandler(db *gorm.DB) *DeviceHandler {
	return &DeviceHandler{db: db}
}

// GetDevices 获取设备列表
// @Summary 获取设备列表
// @Description 获取所有设备列表，支持分页（仅管理员）
// @Tags 设备模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=object}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /devices [get]
func (h *DeviceHandler) GetDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var total int64
	h.db.Model(&models.Device{}).Count(&total)

	var devices []models.Device
	offset := (page - 1) * pageSize
	h.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&devices)

	items := make([]gin.H, 0)
	for _, device := range devices {
		// 计算设备运行时长（小时）
		runningHours := time.Since(device.CreatedAt).Hours()
		if runningHours < 1 {
			runningHours = 1 // 最少算为1小时
		}
		
		// 计算每小时任务执行率
		hourlyRate := float64(device.TaskCount) / runningHours
		// 预估每天可完成任务数（24小时）
		dailyEstimate := hourlyRate * 24
		
		item := gin.H{
			"id":             device.ID,
			"device_id":      device.DeviceID,
			"device_name":    device.DeviceName,
			"device_type":    device.DeviceType,   // android 或 ios
			"device_model":   device.DeviceModel,  // 设备型号
			"os_version":     device.OSVersion,    // 系统版本
			"app_version":    device.AppVersion,   // 应用版本
			"ip":             device.IP,
			"location":       device.Location,
			"os_info":        device.OSInfo,       // 兼容旧字段
			"version":        device.Version,      // 兼容旧字段
			"status":         device.Status,
			"is_blocked":     device.IsBlocked,
			"task_count":     device.TaskCount,
			"hourly_rate":    hourlyRate,     // 每小时任务执行数
			"daily_estimate": dailyEstimate,  // 预估每天可完成数
			"created_at":     device.CreatedAt.Format(time.RFC3339),
		}
		if device.LastHeartbeat != nil {
			item["last_heartbeat"] = device.LastHeartbeat.Format(time.RFC3339)
		}
		if device.LastActive != nil {
			item["last_active"] = device.LastActive.Format(time.RFC3339)
		}
		if device.LastTaskTime != nil {
			item["last_task_time"] = device.LastTaskTime.Format(time.RFC3339)
		}
		items = append(items, item)
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

// GetDeviceByID 获取设备详情
// @Summary 获取设备详情
// @Description 根据ID获取设备详细信息（仅管理员）
// @Tags 设备模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "设备ID"
// @Success 200 {object} response.Response{data=models.Device}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /devices/{id} [get]
func (h *DeviceHandler) GetDeviceByID(c *gin.Context) {
	id := c.Param("id")

	var device models.Device
	if err := h.db.First(&device, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "设备不存在")
		return
	}

	response.Success(c, device)
}

// UpdateDeviceStatus 更新设备状态
// @Summary 更新设备状态
// @Description 更新设备的封禁状态（仅管理员）
// @Tags 设备模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "设备ID"
// @Param request body object{is_blocked=bool} true "状态信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /devices/{id}/status [put]
func (h *DeviceHandler) UpdateDeviceStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		IsBlocked *bool `json:"is_blocked"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	var device models.Device
	if err := h.db.First(&device, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "设备不存在")
		return
	}

	if req.IsBlocked != nil {
		device.IsBlocked = *req.IsBlocked
	}

	h.db.Save(&device)

	response.SuccessWithMsg(c, "设备状态更新成功", nil)
}
