package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "jd-task-platform-go/docs" // 导入自动生成的docs
	"jd-task-platform-go/internal/handlers"
	"jd-task-platform-go/internal/middleware"
	"jd-task-platform-go/internal/models"
	"jd-task-platform-go/internal/services"
)

// @title JD任务平台 API
// @version 1.0
// @description JD任务平台后端API服务 - Go语言实现版本
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@jdtask.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:5001
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
// @description API Key for authentication

func main() {
	// 数据库连接配置
	dsn := "jduser:jdpass123@tcp(localhost:3306)/jd?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai&collation=utf8mb4_unicode_ci"

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			// 使用北京时间
			loc, _ := time.LoadLocation("Asia/Shanghai")
			return time.Now().In(loc)
		},
	})

	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	log.Println("✓ 数据库连接成功")

	// 自动迁移所有表
	db.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.Device{},
		&models.JingdouLog{},
		&models.Setting{},
		&models.APILog{},
		&models.TaskTemplate{},
		&models.TaskType{},
		&models.Proxy{},
		&models.ProxyUsageLog{},
	)
	log.Println("✓ 数据库表迁移完成")

	// 测试查询
	var count int64
	db.Model(&models.User{}).Count(&count)
	log.Printf("✓ 数据库表验证成功，当前用户数: %d", count)

	// 设置连接池
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 初始化 Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS 中间件
	r.Use(middleware.CORSMiddleware())

	// 请求日志中间件
	r.Use(middleware.LoggerMiddleware())

	// 根路径
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"msg":     "JD任务平台后端API服务运行中 (Go版本)",
			"version": "1.0.0",
			"docs":    "/swagger/index.html",
		})
	})

	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API 路由组
	api := r.Group("/api")
	{
		// 认证路由 (无需认证)
		auth := api.Group("/auth")
		{
			authHandler := handlers.NewAuthHandler(db)
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.AuthMiddleware(), authHandler.Logout)
		}

		// 用户路由 (需要认证)
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware())
		{
			userHandler := handlers.NewUserHandler(db)
			users.GET("/me", userHandler.GetCurrentUser)
			users.PUT("/password", userHandler.ChangePassword)
			users.POST("/api-key", userHandler.GenerateAPIKey)
			users.PUT("/profile", userHandler.UpdateProfile)

			// 管理员路由
			users.GET("", middleware.AdminMiddleware(), userHandler.GetUsers)
			users.POST("", middleware.AdminMiddleware(), userHandler.CreateUser)
			users.GET("/search", middleware.AdminMiddleware(), userHandler.SearchUsers)
			users.GET("/recharge-statistics", middleware.AdminMiddleware(), userHandler.GetRechargeStatistics)
			users.GET("/:id", middleware.AdminMiddleware(), userHandler.GetUserByID)
			users.PUT("/:id", middleware.AdminMiddleware(), userHandler.UpdateUser)
			users.DELETE("/:id", middleware.AdminMiddleware(), userHandler.DeleteUser)
			users.POST("/:id/jingdou", middleware.AdminMiddleware(), userHandler.AdjustJingdou)
			users.GET("/:id/apikey", middleware.AdminMiddleware(), userHandler.GetUserApiKey)
			users.POST("/:id/apikey", middleware.AdminMiddleware(), userHandler.ResetUserApiKey)
			users.DELETE("/:id/apikey", middleware.AdminMiddleware(), userHandler.DeleteUserApiKey)
		}

		// 任务路由 (需要认证)
		tasks := api.Group("/tasks")
		tasks.Use(middleware.AuthMiddleware())
		{
			taskHandler := handlers.NewTaskHandler(db)
			tasks.GET("", taskHandler.GetTasks)
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("/stats", taskHandler.GetTaskStats)
			tasks.GET("/statistics", taskHandler.GetTaskStatistics)
			tasks.GET("/:id", taskHandler.GetTaskByID)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)
			tasks.POST("/:id/cancel", taskHandler.CancelTask)
			tasks.PUT("/:id/priority", middleware.AdminMiddleware(), taskHandler.UpdateTaskPriority)

			// 任务类型管理
			tasks.GET("/types", taskHandler.GetTaskTypes)
			tasks.POST("/types", middleware.AdminMiddleware(), taskHandler.CreateTaskType)
			tasks.PUT("/types/:id", middleware.AdminMiddleware(), taskHandler.UpdateTaskType)
		}

		// 任务API Key路由
		tasksApiKey := api.Group("/tasks/apikey")
		tasksApiKey.Use(middleware.APIKeyMiddleware(db))
		{
			taskHandler := handlers.NewTaskHandler(db)
			tasksApiKey.GET("", taskHandler.GetTasks)
			tasksApiKey.POST("", taskHandler.CreateTask)
			tasksApiKey.GET("/statistics", taskHandler.GetTaskStatistics)
			tasksApiKey.GET("/types", taskHandler.GetTaskTypes)
			tasksApiKey.GET("/:id", taskHandler.GetTaskByID)
			tasksApiKey.POST("/batch", taskHandler.BatchCreateTasks)
			tasksApiKey.POST("/:id/cancel", taskHandler.CancelTask)
		}

		// 设备路由 (需要认证)
		devices := api.Group("/devices")
		devices.Use(middleware.AuthMiddleware())
		{
			deviceHandler := handlers.NewDeviceHandler(db)
			devices.GET("", deviceHandler.GetDevices)
			devices.GET("/statistics", middleware.AdminMiddleware(), deviceHandler.GetDeviceStatistics)
			devices.GET("/:id", deviceHandler.GetDeviceByID)
			devices.PUT("/:id/status", deviceHandler.UpdateDeviceStatus)
			devices.POST("/clear-all", middleware.AdminMiddleware(), deviceHandler.ClearAllDevices)
		}

		// 设备API Key路由 - 改为设备密钥认证
		devicesApiKey := api.Group("/devices")
		devicesApiKey.Use(middleware.DeviceKeyMiddleware(db)) // 使用设备密钥认证
		{
			deviceHandler := handlers.NewDeviceHandler(db)
			devicesApiKey.POST("/request-task", deviceHandler.RequestTask)
			devicesApiKey.POST("/task-feedback", deviceHandler.TaskFeedback)
			devicesApiKey.GET("/apikey", deviceHandler.GetDevices)
		}

		// 京豆路由 (需要认证)
		jingdou := api.Group("/jingdou")
		jingdou.Use(middleware.AuthMiddleware())
		{
			jingdouHandler := handlers.NewJingdouHandler(db)
			jingdou.GET("/logs", jingdouHandler.GetJingdouLogs)
			jingdou.GET("/records", jingdouHandler.GetJingdouRecords) // 新增：前端京豆明细接口
			jingdou.GET("/balance", jingdouHandler.GetJingdouBalance)
			jingdou.GET("/statistics", middleware.AdminMiddleware(), jingdouHandler.GetJingdouStatistics)
		}

		// 京豆API Key路由
		jingdouApiKey := api.Group("/jingdou")
		jingdouApiKey.Use(middleware.APIKeyMiddleware(db))
		{
			jingdouHandler := handlers.NewJingdouHandler(db)
			jingdouApiKey.GET("/balance/apikey", jingdouHandler.GetJingdouBalanceByAPIKey)
		}

		// 系统设置路由
		settings := api.Group("/settings")
		{
			settingHandler := handlers.NewSettingHandler(db)
			settings.GET("/frontend", settingHandler.GetFrontendSettings)      // 公开接口
			settings.GET("/announcement", settingHandler.GetLoginAnnouncement) // 公开接口 - 登录公告

			// 需要认证的路由
			settingsAuth := settings.Group("")
			settingsAuth.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
			settingsAuth.GET("", settingHandler.GetSettings)
			settingsAuth.POST("/frontend", settingHandler.SaveFrontendSettings)
			settingsAuth.PUT("/:id", settingHandler.UpdateSetting)
			settingsAuth.PUT("/batch", settingHandler.BatchUpdateSettings)
			settingsAuth.POST("/init", settingHandler.InitDefaultSettings)
			settingsAuth.PUT("/announcement", settingHandler.UpdateLoginAnnouncement) // 管理员更新公告
			// 设备密钥管理接口
			settingsAuth.GET("/device-key", settingHandler.GetDeviceAuthKey)    // 获取设备密钥
			settingsAuth.PUT("/device-key", settingHandler.UpdateDeviceAuthKey) // 更新设备密钥
		}

		// API密钥路由 (需要认证)
		apikey := api.Group("/apikey")
		apikey.Use(middleware.AuthMiddleware())
		{
			apikeyHandler := handlers.NewAPIKeyHandler(db)
			apikey.GET("", apikeyHandler.GetAPIKey)
			apikey.POST("/generate", apikeyHandler.GenerateAPIKey)
			apikey.POST("/reset", apikeyHandler.ResetAPIKey)
			apikey.DELETE("", apikeyHandler.DeleteAPIKey)
			apikey.GET("/logs", apikeyHandler.GetAPILogs)
		}

		// API日志路由 (API Key认证)
		logs := api.Group("/logs")
		logs.Use(middleware.APIKeyMiddleware(db))
		{
			apikeyHandler := handlers.NewAPIKeyHandler(db)
			logs.GET("/apikey", apikeyHandler.GetAPILogsByAPIKey)
		}

		// 仪表板路由 (需要认证)
		dashboard := api.Group("/dashboard")
		dashboard.Use(middleware.AuthMiddleware())
		{
			dashboardHandler := handlers.NewDashboardHandler(db)
			dashboard.GET("/overview", dashboardHandler.GetOverview)
			dashboard.GET("/statistics", dashboardHandler.GetStatistics)
			dashboard.GET("/stat/details", middleware.AdminMiddleware(), dashboardHandler.GetDetailedStatistics)
			dashboard.GET("/future-trends", dashboardHandler.GetFutureTrends)
		}

		// 管理员仪表盘路由 (仅管理员)
		adminDashboard := api.Group("/admin/dashboard")
		adminDashboard.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			adminDashboardHandler := handlers.NewAdminDashboardHandler(db)
			adminDashboard.GET("/today-tasks", adminDashboardHandler.GetTodayTaskStats)
			adminDashboard.GET("/task-pressure", adminDashboardHandler.GetTaskPressure)
			adminDashboard.GET("/finance", adminDashboardHandler.GetFinanceStats)
			adminDashboard.POST("/trigger-expire-check", adminDashboardHandler.TriggerExpiredTaskCheck)
			adminDashboard.POST("/trigger-cleanup", adminDashboardHandler.TriggerDataCleanup)
		}

		// 用户首页路由 (普通用户)
		userHome := api.Group("/user/home")
		userHome.Use(middleware.AuthMiddleware())
		{
			userHomeHandler := handlers.NewUserHomeHandler(db)
			userHome.GET("/today-stats", userHomeHandler.GetUserTodayStats)
			userHome.GET("/templates", userHomeHandler.GetTaskTemplates)
			userHome.POST("/quick-create", userHomeHandler.QuickCreateTask)
			userHome.GET("/template-price", userHomeHandler.GetTemplatePrice)
			userHome.GET("/jingdou-stats", userHomeHandler.GetJingdouStats)
			userHome.PUT("/templates/:id/remark", userHomeHandler.UpdateTemplateRemark)
		}

		// 用户任务管理路由 (普通用户)
		userTasks := api.Group("/user/tasks")
		userTasks.Use(middleware.AuthMiddleware())
		{
			userTaskHandler := handlers.NewUserTaskManageHandler(db)
			userTasks.GET("", userTaskHandler.GetUserTasks)
			userTasks.GET("/status-options", userTaskHandler.GetTaskStatusOptions)
			userTasks.POST("/:id/cancel", userTaskHandler.CancelUserTask)
			userTasks.PUT("/:id", userTaskHandler.UpdateUserTask)
		}

		// 代理池管理路由 (仅管理员)
		proxies := api.Group("/proxies")
		proxies.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			proxyHandler := handlers.NewProxyHandler(db)
			proxies.GET("", proxyHandler.GetProxies)
			proxies.GET("/statistics", proxyHandler.GetProxyStatistics)
			proxies.GET("/usage-logs", proxyHandler.GetProxyUsageLogs)
			proxies.POST("", proxyHandler.CreateProxy)
			proxies.POST("/batch-import", proxyHandler.BatchImportProxies)
			proxies.POST("/batch-delete", proxyHandler.BatchDeleteProxies)
			proxies.GET("/:id", proxyHandler.GetProxyByID)
			proxies.PUT("/:id", proxyHandler.UpdateProxy)
			proxies.DELETE("/:id", proxyHandler.DeleteProxy)
		}

		// Clash 配置公开接口（无需认证，供 Clash Mi 扫码访问）
		proxyHandler := handlers.NewProxyHandler(db)
		api.GET("/proxies/:id/clash-config", proxyHandler.GetClashConfig)
		// v2rayN 配置公开接口（无需认证，供 v2rayN 扫码访问）
		api.GET("/proxies/:id/v2ray-config", proxyHandler.GetV2rayConfig)

		// 代理分配路由 (设备密钥认证)
		proxyApiKey := api.Group("/proxy")
		proxyApiKey.Use(middleware.DeviceKeyMiddleware(db)) // 改为设备密钥认证
		{
			proxyHandler := handlers.NewProxyHandler(db)
			proxyApiKey.POST("/assign", proxyHandler.AssignProxy)
		}

		// =========================================
		// 开放API路由（API Key认证 + 限流）
		// =========================================
		openapi := api.Group("/openapi")
		openapi.Use(middleware.APIKeyMiddleware(db))
		openapi.Use(middleware.APIKeyRateLimitMiddleware())
		{
			openapiHandler := handlers.NewOpenAPIHandler(db)

			// 任务管理接口
			openapi.POST("/tasks", openapiHandler.CreateTask)             // 创建单个任务
			openapi.POST("/tasks/batch", openapiHandler.BatchCreateTasks) // 批量创建任务
			openapi.GET("/tasks", openapiHandler.GetTasks)                // 查询任务列表
			openapi.GET("/tasks/:id", openapiHandler.GetTaskByID)         // 查询任务详情
			openapi.PUT("/tasks/:id", openapiHandler.UpdateTask)          // 修改任务
			openapi.POST("/tasks/:id/cancel", openapiHandler.CancelTask)  // 取消任务
			openapi.GET("/task-types", openapiHandler.GetTaskTypes)       // 获取任务类型

			// 京豆相关接口
			openapi.GET("/balance", openapiHandler.GetBalance)                // 查询余额
			openapi.GET("/jingdou/records", openapiHandler.GetJingdouRecords) // 查询京豆明细
		}
	}

	// 启动任务过期检查服务
	taskExpiryService := services.NewTaskExpiryService(db)
	taskExpiryService.Start()

	// 启动数据清理服务（60天保留期，每日0点执行）
	dataCleanupService := services.NewDataCleanupService(db, 60, 0)
	dataCleanupService.Start()

	// 启动设备状态监控服务（3分钟无活动设为离线）
	deviceStatusService := services.NewDeviceStatusService(db)
	go deviceStatusService.Start()

	// 启动服务器
	port := ":5001"
	log.Println("========================================")
	log.Println("  JD任务平台 Go 后端启动成功")
	log.Println("========================================")
	log.Printf("  服务地址: http://localhost%s\n", port)
	log.Printf("  API文档: http://localhost%s/swagger/index.html\n", port)
	log.Printf("  数据库: MySQL (jd)\n")
	log.Println("========================================")

	if err := r.Run(port); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
