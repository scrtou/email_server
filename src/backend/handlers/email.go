package handlers

import (
	"errors"
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
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if isOAuth2 {
		// If it's OAuth2, we need to check the provider name
		var provider models.OAuthProvider
		if err := database.DB.First(&provider, oauthToken.ProviderID).Error; err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Could not determine OAuth provider for this account.")
			return
		}

		// ★★★ CORE LOGIC CHANGE IS HERE ★★★
		if provider.Name == "microsoft" {
			log.Printf("[GetInbox] Provider is '%s'. Using Microsoft Graph API to fetch emails.", provider.Name)
			emails, total, err = integrations.FetchEmailsWithGraphAPI(emailAccount, page, pageSize)
			// client, err := integrations.GetOAuth2HTTPClient(emailAccount.ID) // 使用您已有的辅助函数
			// if err != nil {
			// 	// ... handle error
			// 	return
			// }
			// resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
			// if err != nil {
			// 	// ... handle error
			// 	return
			// }
			// defer resp.Body.Close()

			// if resp.StatusCode != http.StatusOK {
			// 	// ... handle non-200 error, print body
			// 	utils.SendErrorResponse(c, http.StatusInternalServerError, "Graph API /me call failed with status: "+resp.Status)
			// 	return
			// }

			// var userInfo map[string]interface{}
			// json.NewDecoder(resp.Body).Decode(&userInfo)
			// log.Printf("SUCCESSFULLY CALLED /me ENDPOINT. User Principal Name: %v", userInfo["userPrincipalName"])

			// // 暂时返回成功，不获取邮件
			// c.JSON(http.StatusOK, gin.H{"message": "Basic Graph API test successful!", "user_info": userInfo})
			// return

		} else {
			// For other OAuth providers (e.g., Google), we might still use IMAP
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
		// Provide more user-friendly error messages based on the error type
		if strings.Contains(err.Error(), "oauth2") || strings.Contains(err.Error(), "re-authenticate") {
			utils.SendErrorResponse(c, http.StatusUnauthorized, "Authentication failed. Please try re-connecting your Microsoft account.")
		} else if strings.Contains(err.Error(), "graph api") {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get emails from Microsoft. Please try again later.")
		} else {
			// Generic IMAP errors
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to fetch emails from provider.")
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
