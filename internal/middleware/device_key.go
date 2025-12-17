package middleware

import (
	"net/http"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DeviceKeyMiddleware 设备固定密钥认证中间件
func DeviceKeyMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceKey := c.GetHeader("X-Device-Key")
		if deviceKey == "" {
			response.Error(c, http.StatusUnauthorized, "未提供设备密钥")
			c.Abort()
			return
		}

		// 从数据库获取设备密钥配置
		var setting models.Setting
		if err := db.Where("param_key = ?", "device_auth_key").First(&setting).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, "设备密钥配置不存在")
			c.Abort()
			return
		}

		// 验证密钥
		if deviceKey != setting.ParamValue {
			response.Error(c, http.StatusUnauthorized, "无效的设备密钥")
			c.Abort()
			return
		}

		// 设置标识，表示这是设备认证
		c.Set("auth_type", "device")
		c.Set("device_key", deviceKey)

		c.Next()
	}
}
