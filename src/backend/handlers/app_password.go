package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"email_server/services"
	"email_server/utils"
)

var appPasswordManager *services.AppPasswordManager

// InitAppPasswordManager 初始化应用密码管理器
func InitAppPasswordManager(manager *services.AppPasswordManager) {
	appPasswordManager = manager
}

// GetAppPasswordGuide 获取应用密码设置指南
func GetAppPasswordGuide(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "邮箱地址不能为空")
		return
	}
	
	guide := appPasswordManager.GetAppPasswordGuide(email)
	utils.SendSuccessResponse(c, guide)
}

// GetSupportedProviders 获取支持的邮件服务商
func GetSupportedProviders(c *gin.Context) {
	providers := appPasswordManager.GetSupportedProviders()
	utils.SendSuccessResponse(c, providers)
}

// TestAppPassword 测试应用密码连接
func TestAppPassword(c *gin.Context) {
	var req services.AppPasswordSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	
	// 检测服务商
	if req.Provider == "" {
		req.Provider = appPasswordManager.DetectProvider(req.EmailAddress)
	}
	
	config, exists := appPasswordManager.GetProviderConfig(req.Provider)
	if !exists {
		utils.SendErrorResponse(c, http.StatusBadRequest, "不支持的邮箱服务商")
		return
	}
	
	result := appPasswordManager.TestAppPassword(req.EmailAddress, req.AppPassword, config)
	
	if result.Success {
		utils.SendSuccessResponse(c, result)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "应用密码测试失败",
			"data":    result,
		})
	}
}

// SetupAppPassword 设置应用密码
func SetupAppPassword(c *gin.Context) {
	// 获取用户ID
	userIDClaim, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}
	
	var userID uint
	switch v := userIDClaim.(type) {
	case float64:
		userID = uint(v)
	case int64:
		userID = uint(v)
	case int:
		userID = uint(v)
	default:
		utils.SendErrorResponse(c, http.StatusInternalServerError, "用户ID类型错误")
		return
	}
	
	var req services.AppPasswordSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	
	if err := appPasswordManager.SetupAppPassword(userID, req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	
	utils.SendSuccessResponse(c, gin.H{
		"message": "应用密码设置成功，您的邮箱现在使用永不过期的应用密码",
		"email":   req.EmailAddress,
	})
}

// MigrateToAppPassword 从OAuth2迁移到应用密码
func MigrateToAppPassword(c *gin.Context) {
	// 获取账户ID
	accountIDStr := c.Param("id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "账户ID格式错误")
		return
	}
	
	var req struct {
		AppPassword string `json:"app_password" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	
	if err := appPasswordManager.MigrateFromOAuth2ToAppPassword(uint(accountID), req.AppPassword); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	
	utils.SendSuccessResponse(c, gin.H{
		"message":    "成功从OAuth2迁移到应用密码，您的邮箱现在使用永不过期的应用密码",
		"account_id": accountID,
	})
}

// CheckAccountAuthType 检查账户认证类型
func CheckAccountAuthType(c *gin.Context) {
	// 获取账户ID
	accountIDStr := c.Param("id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "账户ID格式错误")
		return
	}
	
	isUsingAppPassword := appPasswordManager.IsUsingAppPassword(uint(accountID))
	
	authType := "oauth2"
	if isUsingAppPassword {
		authType = "app_password"
	}
	
	utils.SendSuccessResponse(c, gin.H{
		"account_id":            accountID,
		"auth_type":            authType,
		"is_using_app_password": isUsingAppPassword,
		"never_expires":        isUsingAppPassword,
	})
} 