package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"email_server/database"
	"email_server/models"
	"email_server/utils"
)

// TestTokenData 测试令牌数据端点
func TestTokenData(c *gin.Context) {
	var tokens []models.UserOAuthToken
	
	// 简单查询所有令牌
	if err := database.DB.Find(&tokens).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询令牌失败: " + err.Error())
		return
	}
	
	var results []map[string]interface{}
	for _, token := range tokens {
		results = append(results, map[string]interface{}{
			"id":               token.ID,
			"email_account_id": token.EmailAccountID,
			"provider_id":      token.ProviderID,
			"token_type":       token.TokenType,
			"expiry":           token.Expiry,
			"created_at":       token.CreatedAt,
			"updated_at":       token.UpdatedAt,
		})
	}
	
	utils.SendSuccessResponse(c, gin.H{
		"status":      "success",
		"total_count": len(tokens),
		"tokens":      results,
	})
}