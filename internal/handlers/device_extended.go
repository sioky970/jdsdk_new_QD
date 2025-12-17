package handlers

import (
	"net/http"
	"strings"
	"time"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
	"jd-task-platform-go/pkg/utils"

	"github.com/gin-gonic/gin"
)

// GetDeviceStatistics 获取设备统计信息
// @Summary 获取设备统计信息
// @Description 获取设备统计数据（仅管理员）
// @Tags 设备模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Router /devices/statistics [get]
func (h *DeviceHandler) GetDeviceStatistics(c *gin.Context) {
	var stats struct {
		TotalDevices   int64 `json:"total_devices"`
		OnlineDevices  int64 `json:"online_devices"`
		OfflineDevices int64 `json:"offline_devices"`
		WorkingDevices int64 `json:"working_devices"`
		IdleDevices    int64 `json:"idle_devices"`
	}

	h.db.Model(&models.Device{}).Count(&stats.TotalDevices)
	h.db.Model(&models.Device{}).Where("status = ?", "online").Count(&stats.OnlineDevices)
	h.db.Model(&models.Device{}).Where("status = ?", "offline").Count(&stats.OfflineDevices)
	h.db.Model(&models.Device{}).Where("status = ?", "working").Count(&stats.WorkingDevices)
	h.db.Model(&models.Device{}).Where("status = ?", "idle").Count(&stats.IdleDevices)

	response.Success(c, stats)
}

// ClearAllDevices 清空所有设备
// @Summary 清空所有设备
// @Description 清空所有设备记录（仅管理员）
// @Tags 设备模块
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /devices/clear-all [post]
func (h *DeviceHandler) ClearAllDevices(c *gin.Context) {
	result := h.db.Exec("DELETE FROM devices")

	response.SuccessWithMsg(c, "设备已清空", gin.H{
		"deleted_count": result.RowsAffected,
	})
}

// RequestTask 设备请求任务
// @Summary 设备请求任务
// @Description 设备请求待执行的任务
// @Tags 设备模块
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body object{device_id=string,device_name=string,device_type=string,device_model=string,os_version=string,app_version=string} true "设备信息"
// @Success 200 {object} response.Response{data=object}
// @Router /devices/request-task [post]
func (h *DeviceHandler) RequestTask(c *gin.Context) {
	var req struct {
		DeviceID    string `json:"device_id" binding:"required"`
		DeviceName  string `json:"device_name"`
		DeviceType  string `json:"device_type"`  // android 或 ios
		DeviceModel string `json:"device_model"` // 设备型号
		OSVersion   string `json:"os_version"`   // 系统版本
		AppVersion  string `json:"app_version"`  // 应用版本
		// 兼容旧字段
		OSInfo  string `json:"os_info"`
		Version string `json:"version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 获取客户端IP
	clientIP := c.ClientIP()
	if clientIP == "" {
		clientIP = c.Request.RemoteAddr
		if colonIndex := strings.LastIndex(clientIP, ":"); colonIndex != -1 {
			clientIP = clientIP[:colonIndex]
		}
	}

	// 解析地理位置
	location := utils.GetLocationByIP(clientIP)

	// 更新或创建设备
	var device models.Device
	if err := h.db.Where("device_id = ?", req.DeviceID).First(&device).Error; err != nil {
		// 设备不存在，创建新设备
		now := time.Now()
		deviceName := req.DeviceName
		if deviceName == "" {
			deviceName = req.DeviceID
		}
		device = models.Device{
			DeviceID:      req.DeviceID,
			DeviceName:    deviceName,
			DeviceType:    req.DeviceType,
			DeviceModel:   req.DeviceModel,
			OSVersion:     req.OSVersion,
			AppVersion:    req.AppVersion,
			IP:            clientIP,
			Location:      location,
			OSInfo:        req.OSInfo,
			Version:       req.Version,
			Status:        "idle",
			LastHeartbeat: &now,
			LastActive:    &now,
			CreatedAt:     time.Now(),
		}
		h.db.Create(&device)
	} else {
		// 更新设备信息
		now := time.Now()
		device.LastHeartbeat = &now
		device.LastActive = &now
		device.IP = clientIP
		device.Location = location
		if req.DeviceName != "" {
			device.DeviceName = req.DeviceName
		}
		if req.DeviceType != "" {
			device.DeviceType = req.DeviceType
		}
		if req.DeviceModel != "" {
			device.DeviceModel = req.DeviceModel
		}
		if req.OSVersion != "" {
			device.OSVersion = req.OSVersion
		}
		if req.AppVersion != "" {
			device.AppVersion = req.AppVersion
		}
		// 兼容旧字段
		if req.OSInfo != "" {
			device.OSInfo = req.OSInfo
		}
		if req.Version != "" {
			device.Version = req.Version
		}
		if device.Status == "offline" {
			device.Status = "idle"
		}
		h.db.Save(&device)
	}

	// 检查设备是否被封禁
	if device.IsBlocked {
		response.Success(c, gin.H{
			"has_task": false,
			"message":  "设备已被封禁",
		})
		return
	}

	// 查找可执行的任务，按优先级和创建时间排序
	// 包含条件：
	// 1. waiting 或 running 状态的任务（只要未达到完成数量即可继续下发）
	// 2. 未完成的任务（executed_count < execute_count）
	// 3. 未过期的任务（start_time + 24小时 > 当前时间）
	expireThreshold := time.Now().Add(-24 * time.Hour)
	var task models.Task
	if err := h.db.Where(
		"status IN (?, ?) AND executed_count < execute_count AND (start_time IS NULL OR start_time <= ?) AND (start_time IS NULL OR start_time > ?)",
		"waiting", "running", time.Now(), expireThreshold,
	).Order("priority DESC, created_at ASC").First(&task).Error; err != nil {
		// 没有可执行任务
		response.Success(c, gin.H{
			"has_task": false,
			"message":  "暂无待执行任务",
		})
		return
	}

	// 检查该设备是否最近执行过同样的SKU（仅针对特定任务类型）
	// 需要防重复的任务类型：加购、店铺关注、商品关注
	needCheckDuplicate := task.TaskType == "add_to_cart" ||
		task.TaskType == "follow_shop" ||
		task.TaskType == "follow_product"

	if needCheckDuplicate {
		var history models.DeviceTaskHistory
		recentTime := time.Now().Add(-24 * time.Hour)
		if err := h.db.Where("device_id = ? AND sku = ? AND execute_time > ?",
			req.DeviceID, task.SKU, recentTime).First(&history).Error; err == nil {
			// 24小时内执行过同样的SKU，跳过
			response.Success(c, gin.H{
				"has_task": false,
				"message":  "24小时内已执行过相同SKU任务",
			})
			return
		}
	}

	// 更新任务状态
	task.Status = "running"
	task.UpdatedAt = time.Now()
	h.db.Save(&task)

	// 更新设备状态
	device.Status = "working"
	now := time.Now()
	device.LastActive = &now
	h.db.Save(&device)

	response.Success(c, gin.H{
		"has_task":  true,
		"task_id":   task.ID,
		"task_type": task.TaskType,
		"sku":       task.SKU,
		"shop_name": task.ShopName,
		"keyword":   task.Keyword,
		"remark":    task.Remark,
	})
}

// TaskFeedback 任务反馈
// @Summary 任务执行反馈
// @Description 设备提交任务执行结果
// @Tags 设备模块
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body object{device_id=string,task_id=int,status=string,message=string} true "反馈信息"
// @Success 200 {object} response.Response
// @Router /devices/task-feedback [post]
func (h *DeviceHandler) TaskFeedback(c *gin.Context) {
	var req struct {
		DeviceID string `json:"device_id" binding:"required"`
		TaskID   uint   `json:"task_id" binding:"required"`
		Status   string `json:"status" binding:"required"` // success, failed
		Message  string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 查找任务
	var task models.Task
	if err := h.db.First(&task, req.TaskID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "任务不存在")
		return
	}

	// 获取任务类型的执行倍数
	var taskType models.TaskType
	multiplier := 1 // 默认倍数为1
	if err := h.db.Where("type_code = ?", task.TaskType).First(&taskType).Error; err == nil {
		if taskType.ExecuteMultiplier >= 1 {
			multiplier = taskType.ExecuteMultiplier
		}
	}

	tx := h.db.Begin()

	// 任务下发后，无论设备是否成功执行，直接认为本次任务已完成
	// 根据任务类型设置的任务倍数来提交到数据库已完成次数
	task.ExecutedCount += multiplier
	if task.ExecutedCount >= task.ExecuteCount {
		task.Status = "completed"
	} else {
		task.Status = "waiting" // 还未完成，回到等待状态
	}
	task.UpdatedAt = time.Now()
	tx.Save(&task)

	// 记录任务日志
	taskLog := models.TaskLog{
		TaskID:    task.ID,
		DeviceID:  req.DeviceID,
		Status:    req.Status,
		Message:   req.Message,
		CreatedAt: time.Now(),
	}
	tx.Create(&taskLog)

	// 记录设备任务历史
	history := models.DeviceTaskHistory{
		DeviceID:    req.DeviceID,
		TaskID:      task.ID,
		SKU:         task.SKU,
		ExecuteTime: time.Now(),
		Status:      req.Status,
		CreatedAt:   time.Now(),
	}
	tx.Create(&history)

	// 更新设备状态和任务计数
	var device models.Device
	if err := tx.Where("device_id = ?", req.DeviceID).First(&device).Error; err == nil {
		device.Status = "idle"
		now := time.Now()
		device.LastHeartbeat = &now
		device.LastActive = &now
		device.LastTaskTime = &now
		device.TaskCount += multiplier // 增加任务执行次数
		tx.Save(&device)
	}

	tx.Commit()

	response.SuccessWithMsg(c, "任务反馈已记录", nil)
}
