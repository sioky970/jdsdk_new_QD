package middleware

import (
	"net/http"
	"time"

	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// APIKeyMiddleware API Key认证中间件，并更新最后使用时间
func APIKeyMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		if apiKey == "" {
			response.Error(c, http.StatusUnauthorized, "未提供API密钥")
			c.Abort()
			return
		}

		var user models.User
		if err := db.Where("api_key = ?", apiKey).First(&user).Error; err != nil {
			response.Error(c, http.StatusUnauthorized, "无效的API密钥")
			c.Abort()
			return
		}

		if !user.IsActive {
			response.Error(c, http.StatusForbidden, "用户已被禁用")
			c.Abort()
			return
		}

		// 更新最后使用时间
		now := time.Now()
		user.ApiKeyLastUsedAt = &now
		// 异步更新，不阻塞请求
		go func() {
			db.Model(&user).Update("api_key_last_used_at", now)
		}()

		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("role", user.Role)
		c.Set("api_key", apiKey)

		c.Next()
	}
}
