package middleware

import (
    "strings"
    "log"

    "github.com/gin-gonic/gin"
    "email_server/utils"
)

// AuthRequired 需要登录认证的中间件
func AuthRequired() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            utils.SendErrorResponse(c, 401, "请先登录")
            c.Abort()
            return
        }

        // Bearer token格式
        parts := strings.SplitN(authHeader, " ", 2)
        if !(len(parts) == 2 && parts[0] == "Bearer") {
            utils.SendErrorResponse(c, 401, "认证格式错误")
            c.Abort()
            return
        }

        claims, err := utils.ParseToken(parts[1])
        if err != nil {
            log.Printf("Token解析失败: %v", err)
            utils.SendErrorResponse(c, 401, "登录已过期，请重新登录")
            c.Abort()
            return
        }

        // 将用户信息存储到context中
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)
        c.Next()
    })
}

// AdminRequired 需要管理员权限的中间件
func AdminRequired() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists || role != "admin" {
            utils.SendErrorResponse(c, 403, "需要管理员权限")
            c.Abort()
            return
        }
        c.Next()
    })
}

// OptionalAuth 可选认证中间件（登录和未登录都可以访问）
func OptionalAuth() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" {
            parts := strings.SplitN(authHeader, " ", 2)
            if len(parts) == 2 && parts[0] == "Bearer" {
                if claims, err := utils.ParseToken(parts[1]); err == nil {
                    c.Set("user_id", claims.UserID)
                    c.Set("username", claims.Username)
                    c.Set("role", claims.Role)
                }
            }
        }
        c.Next()
    })
}
