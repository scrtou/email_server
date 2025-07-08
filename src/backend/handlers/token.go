package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"email_server/services"
	"email_server/utils"
)

// GetTokenStatus 获取令牌状态统计
func GetTokenStatus(c *gin.Context) {
	tokenRefreshService := services.NewTokenRefreshService()
	
	status, err := tokenRefreshService.GetTokenStatus()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取令牌状态失败")
		return
	}
	
	utils.SendSuccessResponse(c, gin.H{
		"status": "success",
		"data":   status,
	})
}

// RefreshTokens 手动刷新令牌
func RefreshTokens(c *gin.Context) {
	tokenRefreshService := services.NewTokenRefreshService()
	
	// 异步刷新令牌
	go tokenRefreshService.RefreshAllTokens()
	
	utils.SendSuccessResponse(c, gin.H{
		"status":  "success",
		"message": "令牌刷新任务已启动",
	})
}

// CleanupTokens 清理重复和过期的令牌
func CleanupTokens(c *gin.Context) {
	tokenRefreshService := services.NewTokenRefreshService()
	
	result, err := tokenRefreshService.CleanupDuplicateTokens()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "清理令牌失败")
		return
	}
	
	utils.SendSuccessResponse(c, gin.H{
		"status": "success",
		"data":   result,
	})
}