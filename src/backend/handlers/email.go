package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"email_server/database"
	"email_server/integrations"
	"email_server/models"
	"email_server/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DEPRECATED section - no changes here
// ---
const deprecatedEmailHandlerMsg = "This API endpoint related to old 'Email' models is deprecated and will be removed. Please use new API endpoints for 'EmailAccount'."

func GetEmails(c *gin.Context) {
	log.Printf("Deprecated handler called: GetEmails")
	utils.SendErrorResponse(c, 501, deprecatedEmailHandlerMsg+" (GetEmails)")
}
func CreateEmail(c *gin.Context) {
	log.Printf("Deprecated handler called: CreateEmail")
	utils.SendErrorResponse(c, 501, deprecatedEmailHandlerMsg+" (CreateEmail)")
}
func UpdateEmail(c *gin.Context) {
	log.Printf("Deprecated handler called: UpdateEmail for ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedEmailHandlerMsg+" (UpdateEmail)")
}
func DeleteEmail(c *gin.Context) {
	log.Printf("Deprecated handler called: DeleteEmail for ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedEmailHandlerMsg+" (DeleteEmail)")
}
func GetEmailByID(c *gin.Context) {
	log.Printf("Deprecated handler called: GetEmailByID for ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedEmailHandlerMsg+" (GetEmailByID)")
}
func GetEmailServices(c *gin.Context) {
	log.Printf("Deprecated handler called: GetEmailServices for Email ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedEmailHandlerMsg+" (GetEmailServices - use PlatformRegistration/ServiceSubscription APIs)")
}

// ---

// GetInbox fetches emails from the user's specified email account.
// THIS FUNCTION HAS BEEN UPDATED TO USE MICROSOFT GRAPH API
func GetInbox(c *gin.Context) {
	log.Println("[GetInbox] Handler started.")

	// 1. Get User ID from context
	// ...
	userIDClaim, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated (user_id not in context)")
		return
	}

	var userID uint
	// The claim can be float64 or int64 depending on how it's parsed. Handle both.
	switch v := userIDClaim.(type) {
	case float64:
		userID = uint(v)
	case int64:
		userID = uint(v)
	case int:
		userID = uint(v)
	default:
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID type in context")
		return
	}

	if userID == 0 {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}
	// ...
	log.Printf("[GetInbox] Successfully retrieved userID: %d", userID)

	// 2. Get account_id from query string
	accountIDStr := c.Query("account_id")
	if accountIDStr == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "account_id query parameter is required")
		return
	}
	accountID, err := strconv.ParseUint(accountIDStr, 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid account_id format")
		return
	}
	log.Printf("[GetInbox] Target account_id: %d", accountID)

	// 3. Find the email account and check if it's OAuth2
	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&emailAccount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "Email account not found or access denied")
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve email account")
		}
		return
	}
	log.Printf("[GetInbox] Found email account: ID=%d, Email: %s", emailAccount.ID, emailAccount.EmailAddress)

	var oauthToken models.UserOAuthToken
	isOAuth2 := false
	if err := database.DB.Where("email_account_id = ?", emailAccount.ID).First(&oauthToken).Error; err == nil {
		isOAuth2 = true
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[GetInbox] Error checking for OAuth2 token: %v", err)
	}

	// 4. Decide which fetch method to use
	var emails []models.Email
	var total int
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	folder := c.DefaultQuery("folder", "inbox") // 支持文件夹参数，默认为inbox

	if isOAuth2 {
		// If it's OAuth2, we need to check the provider name
		var provider models.OAuthProvider
		if err := database.DB.First(&provider, oauthToken.ProviderID).Error; err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Could not determine OAuth provider for this account.")
			return
		}

		// ★★★ CORE LOGIC CHANGE IS HERE ★★★
		if provider.Name == "microsoft" {
			log.Printf("[GetInbox] Provider is '%s'. Using Microsoft Graph API to fetch emails from folder '%s'.", provider.Name, folder)
			emails, total, err = integrations.FetchEmailsWithGraphAPIFromFolder(emailAccount, page, pageSize, folder)
		} else if provider.Name == "google" {
			log.Printf("[GetInbox] Provider is '%s'. Using Gmail API to fetch emails from folder '%s'.", provider.Name, folder)
			// 将folder名称转换为Gmail标签
			gmailLabel := convertFolderToGmailLabel(folder)
			emails, total, err = integrations.FetchEmailsWithGmailAPIFromFolder(emailAccount, page, pageSize, gmailLabel)
		} else {
			// For other OAuth providers, we might still use IMAP
			log.Printf("[GetInbox] Provider is '%s'. Using standard IMAP to fetch emails.", provider.Name)
			emails, total, err = integrations.FetchEmails(emailAccount, page, pageSize)
		}
	} else {
		// Fallback to password-based IMAP
		log.Printf("[GetInbox] Account is not OAuth2. Using standard IMAP with password to fetch emails.")
		// Check for IMAP settings for non-OAuth accounts
		if emailAccount.IMAPServer == "" || emailAccount.IMAPPort == 0 {
			utils.SendErrorResponse(c, http.StatusBadRequest, "IMAP settings are not configured for this non-OAuth email account.")
			return
		}
		emails, total, err = integrations.FetchEmails(emailAccount, page, pageSize)
	}

	// 5. Handle potential errors from fetching
	if err != nil {
		log.Printf("[GetInbox] Fetching emails failed with error: %v", err)

		// 检查是否是刷新令牌失效错误
		if strings.Contains(err.Error(), "oauth2_refresh_token_expired") {
			// 刷新令牌失效，需要重新授权，但不要让用户登出
			var provider models.OAuthProvider
			providerName := "邮箱"
			if database.DB.First(&provider, oauthToken.ProviderID).Error == nil {
				providerName = provider.Name
			}

			// 使用特殊的状态码422表示需要重新授权
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":           "oauth2_reauth_required",
				"message":         fmt.Sprintf("%s账户的授权已过期，需要重新连接", providerName),
				"provider":        providerName,
				"account_id":      emailAccount.ID,
				"reauth_required": true,
			})
			return
		}

		// Provide more user-friendly error messages based on the error type
		if strings.Contains(err.Error(), "oauth2") || strings.Contains(err.Error(), "re-authenticate") {
			// 检查是哪个提供商
			var provider models.OAuthProvider
			if database.DB.First(&provider, oauthToken.ProviderID).Error == nil {
				if provider.Name == "google" {
					utils.SendErrorResponse(c, http.StatusUnauthorized, "Google账户认证已过期，请重新连接您的Google账户。")
				} else if provider.Name == "microsoft" {
					utils.SendErrorResponse(c, http.StatusUnauthorized, "Microsoft账户认证已过期，请重新连接您的Microsoft账户。")
				} else {
					utils.SendErrorResponse(c, http.StatusUnauthorized, "账户认证已过期，请重新连接您的邮箱账户。")
				}
			} else {
				utils.SendErrorResponse(c, http.StatusUnauthorized, "账户认证已过期，请重新连接您的邮箱账户。")
			}
		} else if strings.Contains(err.Error(), "graph api") {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "无法从Microsoft获取邮件，请稍后重试。")
		} else if strings.Contains(err.Error(), "gmail api") {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "无法从Gmail获取邮件，请稍后重试。")
		} else if strings.Contains(err.Error(), "invalid_grant") || strings.Contains(err.Error(), "Token has been expired or revoked") {
			// 特别处理Google的invalid_grant错误
			utils.SendErrorResponse(c, http.StatusUnauthorized, "Google账户授权已过期或被撤销，请重新连接您的Google账户。")
		} else {
			// Generic IMAP errors
			utils.SendErrorResponse(c, http.StatusInternalServerError, "无法从邮件服务商获取邮件，请检查网络连接或稍后重试。")
		}
		return
	}
	log.Printf("[GetInbox] Fetching successful. Fetched %d emails. Total reported: %d.", len(emails), total)

	// 6. Return the successful response
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"emails": emails,
			"total":  total,
		},
	})
}

// CheckOAuth2TokenStatus 检查OAuth2令牌状态并尝试刷新
func CheckOAuth2TokenStatus(c *gin.Context) {
	log.Println("[CheckOAuth2TokenStatus] Handler started.")

	// 1. Get User ID from context
	userIDClaim, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
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
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID type")
		return
	}

	// 2. Get account_id from query string
	accountIDStr := c.Query("account_id")
	if accountIDStr == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "account_id query parameter is required")
		return
	}
	accountID, err := strconv.ParseUint(accountIDStr, 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid account_id format")
		return
	}

	// 3. Find the email account and verify ownership
	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&emailAccount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "Email account not found or access denied")
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve email account")
		}
		return
	}

	// 4. Check OAuth2 token
	var oauthToken models.UserOAuthToken
	if err := database.DB.Where("email_account_id = ?", emailAccount.ID).First(&oauthToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "No OAuth2 token found for this account")
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to check OAuth2 token")
		}
		return
	}

	// 5. Get provider information
	var provider models.OAuthProvider
	if err := database.DB.First(&provider, oauthToken.ProviderID).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get provider information")
		return
	}

	// 6. Try to get a valid OAuth2 client (this will trigger token refresh if needed)
	var client *http.Client
	if provider.Name == "google" {
		client, err = integrations.GetGmailOAuth2HTTPClient(emailAccount.ID)
	} else {
		client, err = integrations.GetOAuth2HTTPClient(emailAccount.ID)
	}

	if err != nil {
		log.Printf("[CheckOAuth2TokenStatus] Failed to get OAuth2 client: %v", err)
		if strings.Contains(err.Error(), "oauth2") || strings.Contains(err.Error(), "invalid_grant") {
			c.JSON(http.StatusOK, gin.H{
				"status":          "expired",
				"message":         fmt.Sprintf("%s账户授权已过期，需要重新连接", provider.Name),
				"provider":        provider.Name,
				"reauth_required": true,
			})
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to validate token")
		}
		return
	}

	// 7. Test the token by making a simple API call
	var testURL string
	switch provider.Name {
	case "google":
		testURL = "https://gmail.googleapis.com/gmail/v1/users/me/profile"
	case "microsoft":
		testURL = "https://graph.microsoft.com/v1.0/me"
	default:
		// For other providers, just return success if we got a client
		c.JSON(http.StatusOK, gin.H{
			"status":   "valid",
			"message":  "令牌状态正常",
			"provider": provider.Name,
		})
		return
	}

	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create test request")
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[CheckOAuth2TokenStatus] Test API call failed: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"status":          "error",
			"message":         "令牌验证失败，可能需要重新连接",
			"provider":        provider.Name,
			"reauth_required": true,
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		c.JSON(http.StatusOK, gin.H{
			"status":          "expired",
			"message":         fmt.Sprintf("%s账户授权已过期，需要重新连接", provider.Name),
			"provider":        provider.Name,
			"reauth_required": true,
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{
			"status":   "error",
			"message":  "令牌验证失败",
			"provider": provider.Name,
		})
		return
	}

	// 8. Token is valid
	c.JSON(http.StatusOK, gin.H{
		"status":   "valid",
		"message":  "令牌状态正常",
		"provider": provider.Name,
	})
}

// GetEmailDetail fetches a single email's detailed information by messageId
func GetEmailDetail(c *gin.Context) {
	log.Println("[GetEmailDetail] Handler started.")

	// 1. Get User ID from context
	userIDClaim, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated (user_id not in context)")
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
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID type in context")
		return
	}

	if userID == 0 {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid user ID in token")
		return
	}

	// 2. Get messageId from URL parameter
	messageId := c.Param("messageId")
	if messageId == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "messageId parameter is required")
		return
	}

	// 3. Get account_id from query string
	accountIDStr := c.Query("account_id")
	if accountIDStr == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "account_id query parameter is required")
		return
	}
	accountID, err := strconv.ParseUint(accountIDStr, 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid account_id format")
		return
	}

	log.Printf("[GetEmailDetail] Target messageId: %s, account_id: %d", messageId, accountID)

	// 4. Find the email account and verify ownership
	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&emailAccount).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "Email account not found or access denied")
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve email account")
		}
		return
	}

	// 5. Check if this is an OAuth2 account or password-based IMAP account
	var oauthToken models.UserOAuthToken
	isOAuth2 := false
	if err := database.DB.Where("email_account_id = ?", emailAccount.ID).First(&oauthToken).Error; err == nil {
		isOAuth2 = true
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[GetEmailDetail] Error checking for OAuth2 token: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to check account authentication type")
		return
	}

	// 6. Fetch email detail using appropriate method
	var email *models.Email
	if isOAuth2 {
		// OAuth2 account - use provider-specific API
		var provider models.OAuthProvider
		if err := database.DB.First(&provider, oauthToken.ProviderID).Error; err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve provider information")
			return
		}

		if provider.Name == "microsoft" {
			log.Printf("[GetEmailDetail] Provider is '%s'. Using Microsoft Graph API to fetch email detail.", provider.Name)
			email, err = integrations.FetchEmailDetailWithGraphAPI(emailAccount, messageId)
			if err != nil {
				log.Printf("[GetEmailDetail] Error fetching email detail: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch email detail: "+err.Error())
				return
			}
		} else if provider.Name == "google" {
			log.Printf("[GetEmailDetail] Provider is '%s'. Using Gmail API to fetch email detail.", provider.Name)
			email, err = integrations.FetchGmailMessageDetail(emailAccount, messageId)
			if err != nil {
				log.Printf("[GetEmailDetail] Error fetching Gmail email detail: %v", err)
				utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch email detail: "+err.Error())
				return
			}
		} else {
			utils.SendErrorResponse(c, http.StatusNotImplemented, "Provider not supported for email detail fetching")
			return
		}
	} else {
		// Password-based IMAP account - use IMAP to fetch email detail
		log.Printf("[GetEmailDetail] Account is not OAuth2. Using standard IMAP to fetch email detail.")
		// Check for IMAP settings for non-OAuth accounts
		if emailAccount.IMAPServer == "" || emailAccount.IMAPPort == 0 {
			utils.SendErrorResponse(c, http.StatusBadRequest, "IMAP settings are not configured for this non-OAuth email account.")
			return
		}
		email, err = integrations.FetchEmailDetailWithIMAP(emailAccount, messageId)
		if err != nil {
			log.Printf("[GetEmailDetail] Error fetching email detail with IMAP: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch email detail: "+err.Error())
			return
		}
	}

	if email == nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Email not found")
		return
	}

	log.Printf("[GetEmailDetail] Successfully fetched email detail for messageId: %s", messageId)

	// 7. Return the successful response
	c.JSON(http.StatusOK, gin.H{
		"data": email,
	})
}

// convertFolderToGmailLabel 将通用文件夹名称转换为Gmail标签
func convertFolderToGmailLabel(folder string) string {
	switch strings.ToLower(folder) {
	case "inbox":
		return "INBOX"
	case "sent", "sentitems":
		return "SENT"
	case "drafts":
		return "DRAFT"
	case "trash", "deleteditems":
		return "TRASH"
	case "spam", "junkemail":
		return "SPAM"
	case "important":
		return "IMPORTANT"
	case "starred":
		return "STARRED"
	default:
		return "INBOX" // 默认返回收件箱
	}
}

// MarkEmailAsRead 标记邮件为已读
func MarkEmailAsRead(c *gin.Context) {
	// 1. 获取用户ID
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	userID, ok := userIDRaw.(int64)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "用户ID类型错误")
		return
	}
	actualUserID := uint(userID)

	// 2. 获取邮件ID和账户ID
	messageId := c.Param("messageId")
	accountIdStr := c.Query("account_id")

	if messageId == "" || accountIdStr == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "缺少必要参数")
		return
	}

	accountId, err := strconv.ParseUint(accountIdStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的账户ID")
		return
	}

	// 3. 验证账户权限
	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", accountId, actualUserID).First(&emailAccount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "邮箱账户未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取邮箱账户失败: "+err.Error())
		return
	}

	// 4. 根据提供商调用相应的API标记为已读
	log.Printf("[MarkEmailAsRead] EmailAccount Provider: '%s'", emailAccount.Provider)

	// 根据Provider字段判断邮箱服务商类型
	providerLower := strings.ToLower(emailAccount.Provider)

	if providerLower == "microsoft" || providerLower == "outlook.com" || strings.Contains(providerLower, "outlook") || strings.Contains(providerLower, "hotmail") {
		log.Printf("[MarkEmailAsRead] Using Microsoft Graph API for provider: %s", emailAccount.Provider)
		err = integrations.MarkMicrosoftEmailAsRead(emailAccount, messageId)
	} else if providerLower == "google" || providerLower == "gmail.com" || strings.Contains(providerLower, "gmail") {
		log.Printf("[MarkEmailAsRead] Using Gmail API for provider: %s", emailAccount.Provider)
		err = integrations.MarkGmailAsRead(emailAccount, messageId)
	} else {
		log.Printf("[MarkEmailAsRead] Unsupported provider: '%s'", emailAccount.Provider)
		utils.SendErrorResponse(c, http.StatusNotImplemented, fmt.Sprintf("该提供商 '%s' 暂不支持标记已读功能", emailAccount.Provider))
		return
	}

	if err != nil {
		log.Printf("[MarkEmailAsRead] Error marking email as read: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "标记邮件已读失败: "+err.Error())
		return
	}

	// 6. 返回成功响应
	utils.SendSuccessResponse(c, gin.H{"message": "邮件已标记为已读"})
}
