package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"email_server/database"
	"email_server/integrations"
	"email_server/models"
	"email_server/utils"
)

// GetEmailsWithAppPassword 使用应用密码获取邮件列表
func GetEmailsWithAppPassword(c *gin.Context) {
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
	
	// 获取账户ID
	accountIDStr := c.Param("id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "账户ID格式错误")
		return
	}
	
	// 查找邮箱账户并验证权限
	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&emailAccount).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "邮箱账户不存在或无权访问")
		return
	}
	
	// 验证是否使用应用密码
	if emailAccount.PasswordEncrypted == "" || emailAccount.IMAPServer == "" || emailAccount.IMAPPort == 0 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "该账户未配置应用密码")
		return
	}
	
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	// 使用应用密码获取邮件
	emails, total, err := integrations.FetchEmailsWithAppPassword(emailAccount, page, pageSize)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取邮件失败: "+err.Error())
		return
	}
	
	utils.SendSuccessResponse(c, gin.H{
		"emails": emails,
		"total":  total,
		"page":   page,
		"page_size": pageSize,
		"auth_type": "app_password",
		"message": "邮件获取成功（使用应用密码）",
	})
}

// GetEmailDetailWithAppPassword 使用应用密码获取邮件详情
func GetEmailDetailWithAppPassword(c *gin.Context) {
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
	
	// 获取账户ID和消息ID
	accountIDStr := c.Param("id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "账户ID格式错误")
		return
	}
	
	messageID := c.Param("messageId")
	if messageID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "消息ID不能为空")
		return
	}
	
	// 查找邮箱账户并验证权限
	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&emailAccount).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "邮箱账户不存在或无权访问")
		return
	}
	
	// 验证是否使用应用密码
	if emailAccount.PasswordEncrypted == "" || emailAccount.IMAPServer == "" || emailAccount.IMAPPort == 0 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "该账户未配置应用密码")
		return
	}
	
	// 使用应用密码获取邮件详情
	email, err := integrations.FetchSingleEmailWithAppPassword(emailAccount, messageID)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取邮件详情失败: "+err.Error())
		return
	}
	
	utils.SendSuccessResponse(c, gin.H{
		"email": email,
		"auth_type": "app_password",
		"message": "邮件详情获取成功（使用应用密码）",
	})
}

// GetMailboxFoldersWithAppPassword 使用应用密码获取邮箱文件夹列表
func GetMailboxFoldersWithAppPassword(c *gin.Context) {
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
	
	// 获取账户ID
	accountIDStr := c.Param("id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "账户ID格式错误")
		return
	}
	
	// 查找邮箱账户并验证权限
	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&emailAccount).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "邮箱账户不存在或无权访问")
		return
	}
	
	// 验证是否使用应用密码
	if emailAccount.PasswordEncrypted == "" || emailAccount.IMAPServer == "" || emailAccount.IMAPPort == 0 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "该账户未配置应用密码")
		return
	}
	
	// 使用应用密码获取文件夹列表
	folders, err := integrations.GetMailboxListWithAppPassword(emailAccount)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取文件夹列表失败: "+err.Error())
		return
	}
	
	utils.SendSuccessResponse(c, gin.H{
		"folders": folders,
		"auth_type": "app_password",
		"message": "文件夹列表获取成功（使用应用密码）",
	})
} 