package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"email_server/config"
	"email_server/database"
	"email_server/handlers"
	"email_server/middleware"
)

func setupRouter() *gin.Engine { //函数签名 返回指针类型
	r := gin.Default()

	// 应用CORS中间件
	r.Use(middleware.CORS())

	// 公开路由（不需要认证）
	public := r.Group("/api/v1")
	{
		// 认证相关
		auth := public.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/refresh", handlers.RefreshToken)

			// OAuth2 相关路由
			oauth2 := auth.Group("/oauth2")
			{
				oauth2.GET("/linuxdo/login", handlers.LinuxDoOAuth2Login)
				oauth2.GET("/linuxdo/callback", handlers.LinuxDoOAuth2Callback)
				oauth2.GET("/stats", handlers.GetOAuth2StateStats) // 监控端点
			}
		}

		// 健康检查
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now()})
		})
	}

	// 需要认证的路由
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthRequired())
	{
		// 用户相关
		// user := protected.Group("/user") // Grouping /users together
		// {
		// user.GET("/profile", handlers.GetProfile) // Old path
		protected.GET("/users/me", handlers.GetProfile)                      // New path as per plan /api/v1/users/me
		protected.PUT("/users/me", handlers.UpdateProfile)                   // 更新用户资料路由
		protected.POST("/users/me/change-password", handlers.ChangePassword) // 修改密码路由
		// user.POST("/logout", handlers.Logout) // Moved to /auth/logout
		// }

		// Auth related routes that need protection (e.g. logout)
		authProtected := protected.Group("/auth")
		{
			authProtected.POST("/logout", handlers.Logout) // New path as per plan /api/v1/auth/logout
		}

		// EmailAccount 模块
		emailAccounts := protected.Group("/email-accounts")
		{
			emailAccounts.POST("", handlers.CreateEmailAccount)
			emailAccounts.GET("", handlers.GetEmailAccounts)
			emailAccounts.GET("/:id", handlers.GetEmailAccountByID)
			emailAccounts.PUT("/:id", handlers.UpdateEmailAccount)
			emailAccounts.DELETE("/:id", handlers.DeleteEmailAccount)
			emailAccounts.GET("/providers", handlers.GetEmailAccountProviders)                                  // 新增：获取唯一服务商列表
			emailAccounts.GET("/:id/platform-registrations", handlers.GetPlatformRegistrationsByEmailAccountID) // 修改参数名
		}

		// Platform 模块
		platforms := protected.Group("/platforms")
		{
			platforms.POST("", handlers.CreatePlatform)
			platforms.GET("", handlers.GetPlatforms)
			platforms.GET("/:id", handlers.GetPlatformByID)
			platforms.PUT("/:id", handlers.UpdatePlatform)
			platforms.DELETE("/:id", handlers.DeletePlatform)
			platforms.GET("/:id/email-registrations", handlers.GetEmailRegistrationsByPlatformID) // 修改参数名
		}

		// PlatformRegistration 模块
		platformRegistrations := protected.Group("/platform-registrations")
		{
			platformRegistrations.POST("", handlers.CreatePlatformRegistrationWithIDs)         // 通过ID创建
			platformRegistrations.POST("/by-name", handlers.CreatePlatformRegistrationByNames) // 通过名称创建
			platformRegistrations.GET("", handlers.GetPlatformRegistrations)
			platformRegistrations.GET("/:id", handlers.GetPlatformRegistrationByID)
			platformRegistrations.GET("/:id/password", handlers.GetPlatformRegistrationPassword) // 获取密码
			platformRegistrations.PUT("/:id", handlers.UpdatePlatformRegistration)
			platformRegistrations.DELETE("/:id", handlers.DeletePlatformRegistration)
			platformRegistrations.GET("/:id/service-subscriptions", handlers.GetServiceSubscriptionsByPlatformRegistrationID)
		}

		// ServiceSubscription 模块
		serviceSubscriptions := protected.Group("/service-subscriptions")
		{
			serviceSubscriptions.POST("", handlers.CreateServiceSubscription)
			serviceSubscriptions.GET("", handlers.GetServiceSubscriptions)
			serviceSubscriptions.GET("/distinct-platform-names", handlers.GetDistinctPlatformNames) // 新增
			serviceSubscriptions.GET("/distinct-emails", handlers.GetDistinctEmails)                // 新增
			serviceSubscriptions.GET("/distinct-usernames", handlers.GetDistinctUsernames)          // 新增
			serviceSubscriptions.GET("/:id", handlers.GetServiceSubscriptionByID)
			serviceSubscriptions.PUT("/:id", handlers.UpdateServiceSubscription)
			serviceSubscriptions.DELETE("/:id", handlers.DeleteServiceSubscription)
		}

		// 导入模块
		importerGroup := protected.Group("/import") // 使用 importer 而不是 import 避免与 Go 关键字冲突
		{
			importerGroup.POST("/bitwarden-csv", handlers.ImportBitwardenCSVHandler)
		}

		// 仪表板
		protected.GET("/dashboard", handlers.GetDashboard)                // 旧的仪表盘API，已在handler中标记为弃用
		protected.GET("/dashboard/summary", handlers.GetDashboardSummary) // 新的仪表盘摘要API

		// 全局搜索
		protected.GET("/search", handlers.SearchHandler)

		// 用户提醒
		protected.GET("/users/me/reminders", handlers.GetUserReminders)
		protected.PUT("/users/me/reminders/:id/read", handlers.MarkReminderAsRead) // 新增：标记提醒为已读

		// 邮箱管理 (DEPRECATED - Use /email-accounts)
		// emails := protected.Group("/emails")
		// {
		// 	emails.GET("", handlers.GetEmails)
		// 	emails.POST("", handlers.CreateEmail)
		// 	emails.GET("/:id", handlers.GetEmailByID)
		// 	emails.PUT("/:id", handlers.UpdateEmail)
		// 	emails.DELETE("/:id", handlers.DeleteEmail)
		// 	emails.GET("/:id/services", handlers.GetEmailServices)
		// }

		// 服务管理 (DEPRECATED - Use /platforms)
		// services := protected.Group("/services")
		// {
		// 	services.GET("", handlers.GetServices)
		// 	services.POST("", handlers.CreateService)
		// 	services.GET("/:id", handlers.GetServiceByID)
		// 	services.PUT("/:id", handlers.UpdateService)
		// 	services.DELETE("/:id", handlers.DeleteService)
		// 	services.GET("/:id/emails", handlers.GetServiceEmails)
		// }

		// 邮箱服务关联管理 (DEPRECATED - Use /platform-registrations and /service-subscriptions)
		// emailServices := protected.Group("/email-services")
		// {
		// 	emailServices.GET("", handlers.GetAllEmailServices)
		// 	emailServices.POST("", handlers.CreateEmailService)
		// 	emailServices.PUT("/:id", handlers.UpdateEmailService)
		// 	emailServices.DELETE("/:id", handlers.DeleteEmailService)
		// }
	}

	// 管理员路由
	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.AuthRequired())
	admin.Use(middleware.AdminRequired())
	{
		// 用户管理
		admin.GET("/users", handlers.GetAllUsers)
		admin.PUT("/users/:id/status", handlers.UpdateUserStatus) // 更新用户状态
		admin.PUT("/users/:id/role", handlers.UpdateUserRole)     // 更新用户角色
	}

	// 静态文件服务
	middleware.ServeStaticFiles(r)

	return r
}

func main() {
	// 初始化配置
	config.Init()

	// 初始化数据库
	database.Init(config.AppConfig.Database.File) // Pass SQLite file path
	// defer database.Close() // GORM typically doesn't require explicit close in this manner for app lifecycle

	// 初始化并启动定时任务
	handlers.StartSubscriptionReminderJob() // 新增：启动定时任务

	// 设置路由
	r := setupRouter() //短变量声明

	// 启动服务器
	fmt.Printf("服务器启动在 http://localhost:%s\n", config.AppConfig.Server.Port)
	fmt.Println("默认管理员账户: admin / password")
	log.Fatal(r.Run(":" + config.AppConfig.Server.Port))
}
