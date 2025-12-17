package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"
)

type SettingHandler struct {
	db *gorm.DB
}

func NewSettingHandler(db *gorm.DB) *SettingHandler {
	return &SettingHandler{db: db}
}

// GetSettings 获取系统设置列表
// @Summary 获取系统设置列表
// @Description 获取所有系统设置（仅管理员）
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Router /settings [get]
func (h *SettingHandler) GetSettings(c *gin.Context) {
	var settings []models.Setting
	h.db.Order("id ASC").Find(&settings)

	items := make([]gin.H, 0)
	for _, s := range settings {
		items = append(items, gin.H{
			"id":          s.ID,
			"param_key":   s.ParamKey,
			"param_value": s.ParamValue,
			"param_type":  s.ParamType,
			"description": s.Description,
			"updated_at":  s.UpdatedAt.Format(time.RFC3339),
		})
	}

	response.Success(c, gin.H{"settings": items})
}

// GetFrontendSettings 获取前端设置
// @Summary 获取前端设置
// @Description 获取前端配置信息，无需认证
// @Tags 系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=object}
// @Router /settings/frontend [get]
func (h *SettingHandler) GetFrontendSettings(c *gin.Context) {
	var setting models.Setting
	if err := h.db.Where("param_key = ?", "frontend_config").First(&setting).Error; err != nil {
		// 如果不存在，返回默认配置
		defaultConfig := gin.H{
			"app_name":        "JD任务平台",
			"theme":           "dark",
			"primary_color":   "#1890ff",
			"logo_url":        "",
			"enable_register": true,
		}
		response.Success(c, defaultConfig)
		return
	}

	// 解析JSON配置
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(setting.ParamValue), &config); err != nil {
		response.Error(c, http.StatusInternalServerError, "配置解析失败")
		return
	}

	response.Success(c, config)
}

// SaveFrontendSettings 保存前端设置
// @Summary 保存前端设置
// @Description 保存前端配置信息（仅管理员）
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param config body object true "前端配置"
// @Success 200 {object} response.Response
// @Router /settings/frontend [post]
func (h *SettingHandler) SaveFrontendSettings(c *gin.Context) {
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "配置序列化失败")
		return
	}

	var setting models.Setting
	if err := h.db.Where("param_key = ?", "frontend_config").First(&setting).Error; err != nil {
		// 不存在则创建
		setting = models.Setting{
			ParamKey:    "frontend_config",
			ParamValue:  string(configJSON),
			ParamType:   "json",
			Description: "前端配置",
			UpdatedAt:   time.Now(),
		}
		h.db.Create(&setting)
	} else {
		// 存在则更新
		setting.ParamValue = string(configJSON)
		setting.UpdatedAt = time.Now()
		h.db.Save(&setting)
	}

	response.SuccessWithMsg(c, "前端设置保存成功", nil)
}

// UpdateSetting 修改单个设置
// @Summary 修改单个设置
// @Description 修改单个系统设置（仅管理员）
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "设置ID"
// @Param request body object{param_value=string} true "设置值"
// @Success 200 {object} response.Response
// @Router /settings/{id} [put]
func (h *SettingHandler) UpdateSetting(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ParamValue string `json:"param_value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	var setting models.Setting
	if err := h.db.First(&setting, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "设置不存在")
		return
	}

	setting.ParamValue = req.ParamValue
	setting.UpdatedAt = time.Now()
	h.db.Save(&setting)

	response.SuccessWithMsg(c, "设置更新成功", nil)
}

// BatchUpdateSettings 批量修改设置
// @Summary 批量修改设置
// @Description 批量修改系统设置（仅管理员）
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param settings body object true "设置列表"
// @Success 200 {object} response.Response
// @Router /settings/batch [put]
func (h *SettingHandler) BatchUpdateSettings(c *gin.Context) {
	var req struct {
		Settings []struct {
			ID         uint   `json:"id"`
			ParamValue string `json:"param_value"`
		} `json:"settings"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	tx := h.db.Begin()
	for _, s := range req.Settings {
		var setting models.Setting
		if err := tx.First(&setting, s.ID).Error; err == nil {
			setting.ParamValue = s.ParamValue
			setting.UpdatedAt = time.Now()
			tx.Save(&setting)
		}
	}
	tx.Commit()

	response.SuccessWithMsg(c, "批量更新成功", nil)
}

// InitDefaultSettings 初始化默认设置
// @Summary 初始化默认设置
// @Description 初始化系统默认设置（仅管理员）
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /settings/init [post]
func (h *SettingHandler) InitDefaultSettings(c *gin.Context) {
	defaultSettings := []models.Setting{
		{
			ParamKey:    "app_name",
			ParamValue:  "JD任务平台",
			ParamType:   "string",
			Description: "应用名称",
			UpdatedAt:   time.Now(),
		},
		{
			ParamKey:    "enable_register",
			ParamValue:  "true",
			ParamType:   "boolean",
			Description: "是否允许注册",
			UpdatedAt:   time.Now(),
		},
		{
			ParamKey:    "default_jingdou",
			ParamValue:  "0",
			ParamType:   "integer",
			Description: "新用户默认京豆",
			UpdatedAt:   time.Now(),
		},
		{
			ParamKey:    "login_announcement",
			ParamValue:  "",
			ParamType:   "string",
			Description: "登录成功后显示的公告内容，为空则不显示",
			UpdatedAt:   time.Now(),
		},
	}

	for _, s := range defaultSettings {
		var existing models.Setting
		if err := h.db.Where("param_key = ?", s.ParamKey).First(&existing).Error; err != nil {
			h.db.Create(&s)
		}
	}

	response.SuccessWithMsg(c, "默认设置初始化成功", nil)
}

// GetLoginAnnouncement 获取登录公告
// @Summary 获取登录公告
// @Description 获取登录成功后显示的公告内容，无需认证
// @Tags 系统设置
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=object}
// @Router /settings/announcement [get]
func (h *SettingHandler) GetLoginAnnouncement(c *gin.Context) {
	var setting models.Setting
	if err := h.db.Where("param_key = ?", "login_announcement").First(&setting).Error; err != nil {
		// 不存在则返回空
		response.Success(c, gin.H{"announcement": ""})
		return
	}

	response.Success(c, gin.H{"announcement": setting.ParamValue})
}

// UpdateLoginAnnouncement 更新登录公告
// @Summary 更新登录公告
// @Description 更新登录成功后显示的公告内容（仅管理员）
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{announcement=string} true "公告内容"
// @Success 200 {object} response.Response
// @Router /settings/announcement [put]
func (h *SettingHandler) UpdateLoginAnnouncement(c *gin.Context) {
	var req struct {
		Announcement string `json:"announcement"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	var setting models.Setting
	if err := h.db.Where("param_key = ?", "login_announcement").First(&setting).Error; err != nil {
		// 不存在则创建
		setting = models.Setting{
			ParamKey:    "login_announcement",
			ParamValue:  req.Announcement,
			ParamType:   "string",
			Description: "登录成功后显示的公告内容，为空则不显示",
			UpdatedAt:   time.Now(),
		}
		h.db.Create(&setting)
	} else {
		// 存在则更新
		setting.ParamValue = req.Announcement
		setting.UpdatedAt = time.Now()
		h.db.Save(&setting)
	}

	if req.Announcement == "" {
		response.SuccessWithMsg(c, "公告已清空", nil)
	} else {
		response.SuccessWithMsg(c, "公告更新成功", nil)
	}
}

// GetDeviceAuthKey 获取设备认证密钥
// @Summary 获取设备认证密钥
// @Description 获取当前的设备认证密钥（仅管理员）
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object{device_key=string}}
// @Router /settings/device-key [get]
func (h *SettingHandler) GetDeviceAuthKey(c *gin.Context) {
	var setting models.Setting
	if err := h.db.Where("param_key = ?", "device_auth_key").First(&setting).Error; err != nil {
		// 不存在则返回默认值
		response.Success(c, gin.H{"device_key": "KKNN778899"})
		return
	}

	response.Success(c, gin.H{"device_key": setting.ParamValue})
}

// UpdateDeviceAuthKey 更新设备认证密钥
// @Summary 更新设备认证密钥
// @Description 更新设备认证密钥（仅管理员，密钥长度必须大于6位）
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{device_key=string} true "设备密钥"
// @Success 200 {object} response.Response
// @Router /settings/device-key [put]
func (h *SettingHandler) UpdateDeviceAuthKey(c *gin.Context) {
	var req struct {
		DeviceKey string `json:"device_key" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "设备密钥长度必须大于6位")
		return
	}

	var setting models.Setting
	if err := h.db.Where("param_key = ?", "device_auth_key").First(&setting).Error; err != nil {
		// 不存在则创建
		setting = models.Setting{
			ParamKey:    "device_auth_key",
			ParamValue:  req.DeviceKey,
			ParamType:   "string",
			Description: "设备认证密钥（用于设备端API认证）",
			UpdatedAt:   time.Now(),
		}
		h.db.Create(&setting)
	} else {
		// 存在则更新
		setting.ParamValue = req.DeviceKey
		setting.UpdatedAt = time.Now()
		h.db.Save(&setting)
	}

	response.SuccessWithMsg(c, "设备密钥更新成功", nil)
}
