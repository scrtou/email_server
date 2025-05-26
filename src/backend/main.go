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
        user := protected.Group("/user")
        {
            user.GET("/profile", handlers.GetProfile)
            user.PUT("/profile", handlers.UpdateProfile)
            user.POST("/change-password", handlers.ChangePassword)
            user.POST("/logout", handlers.Logout)
        }

        // 仪表板
        protected.GET("/dashboard", handlers.GetDashboard)

        // 邮箱管理
        emails := protected.Group("/emails")
        {
            emails.GET("", handlers.GetEmails)
            emails.POST("", handlers.CreateEmail)
            emails.GET("/:id", handlers.GetEmailByID)
            emails.PUT("/:id", handlers.UpdateEmail)
            emails.DELETE("/:id", handlers.DeleteEmail)
            emails.GET("/:id/services", handlers.GetEmailServices)
        }

        // 服务管理
        services := protected.Group("/services")
        {
            services.GET("", handlers.GetServices)
            services.POST("", handlers.CreateService)
            services.GET("/:id", handlers.GetServiceByID)
            services.PUT("/:id", handlers.UpdateService)
            services.DELETE("/:id", handlers.DeleteService)
            services.GET("/:id/emails", handlers.GetServiceEmails)
        }

        // 邮箱服务关联管理
        emailServices := protected.Group("/email-services")
        {
            emailServices.GET("", handlers.GetAllEmailServices)
            emailServices.POST("", handlers.CreateEmailService)
            emailServices.PUT("/:id", handlers.UpdateEmailService)
            emailServices.DELETE("/:id", handlers.DeleteEmailService)
        }
    }

    // 管理员路由
    admin := r.Group("/api/v1/admin")
    admin.Use(middleware.AuthRequired())
    admin.Use(middleware.AdminRequired())
    {
        // 管理员可以查看所有用户的数据
        admin.GET("/users", handlers.GetAllUsers)
        admin.PUT("/users/:id/status", handlers.UpdateUserStatus)
    }

    // 静态文件服务
    middleware.ServeStaticFiles(r)

    return r
}

func main() {
    // 初始化配置
    config.Init()
    
    // 初始化数据库
    database.Init()
    defer database.Close()

    // 设置路由
    r := setupRouter() //短变量声明

    // 启动服务器
    fmt.Printf("服务器启动在 http://localhost:%s\n", config.AppConfig.Server.Port)
    fmt.Println("默认管理员账户: admin / password")
    log.Fatal(r.Run(":" + config.AppConfig.Server.Port))
}
