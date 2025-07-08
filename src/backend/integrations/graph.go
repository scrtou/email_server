// 微软接口 Microsoft Graph API
package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"email_server/database"
	"email_server/models"
	"email_server/utils"

	"golang.org/x/oauth2"
)

// GraphAPIMessage 对应 Graph API 返回的邮件结构
type GraphAPIMessage struct {
	ID               string            `json:"id"`
	ReceivedDateTime time.Time         `json:"receivedDateTime"`
	Subject          string            `json:"subject"`
	From             GraphAPIAddress   `json:"from"`
	ToRecipients     []GraphAPIAddress `json:"toRecipients"`
	IsRead           bool              `json:"isRead"`
	HasAttachments   bool              `json:"hasAttachments"`
}

type GraphAPIAddress struct {
	EmailAddress struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	} `json:"emailAddress"`
}

// GraphAPIResponse 对应 Graph API 列表请求的响应结构
type GraphAPIResponse struct {
	Value    []GraphAPIMessage `json:"value"`
	NextLink string            `json:"@odata.nextLink"`
	Count    int               `json:"@odata.count"`
}

// GetMicrosoftOAuth2HTTPClient 专门为Microsoft Graph API获取OAuth2客户端，包含自动令牌刷新功能
func GetMicrosoftOAuth2HTTPClient(accountID uint) (*http.Client, error) {
	var oauthToken models.UserOAuthToken
	if err := database.DB.Where("email_account_id = ?", accountID).First(&oauthToken).Error; err != nil {
		return nil, fmt.Errorf("no oauth token found for account %d: %w", accountID, err)
	}

	var provider models.OAuthProvider
	if err := database.DB.First(&provider, oauthToken.ProviderID).Error; err != nil {
		return nil, fmt.Errorf("failed to find oauth provider %d: %w", oauthToken.ProviderID, err)
	}

	decryptedSecret, err := utils.Decrypt(provider.ClientSecretEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt client secret: %w", err)
	}

	conf := &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: string(decryptedSecret),
		Scopes:       strings.Split(provider.Scopes, ","),
		Endpoint: oauth2.Endpoint{
			AuthURL:  provider.AuthURL,
			TokenURL: provider.TokenURL,
		},
	}

	decryptedAccessToken, err := utils.Decrypt(oauthToken.AccessTokenEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access token: %w", err)
	}

	decryptedRefreshToken, err := utils.Decrypt(oauthToken.RefreshTokenEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	token := &oauth2.Token{
		AccessToken:  string(decryptedAccessToken),
		RefreshToken: string(decryptedRefreshToken),
		TokenType:    oauthToken.TokenType,
		Expiry:       oauthToken.Expiry,
	}

	// 创建一个带有令牌刷新回调的TokenSource
	tokenSource := &microsoftRefreshableTokenSource{
		config:     conf,
		token:      token,
		oauthToken: &oauthToken,
		accountID:  accountID,
	}

	// 返回一个会自动刷新token的http.Client
	return &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}, nil
}

// microsoftRefreshableTokenSource 实现oauth2.TokenSource接口，支持Microsoft Graph API的自动刷新和数据库更新
type microsoftRefreshableTokenSource struct {
	config     *oauth2.Config
	token      *oauth2.Token
	oauthToken *models.UserOAuthToken
	accountID  uint
	mu         sync.Mutex
}

// Token 实现oauth2.TokenSource接口
func (ts *microsoftRefreshableTokenSource) Token() (*oauth2.Token, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// 检查令牌是否需要刷新
	if ts.token.Valid() {
		return ts.token, nil
	}

	log.Printf("[Microsoft OAuth2] Token expired for account %d, refreshing...", ts.accountID)

	// 使用配置的TokenSource进行刷新
	tokenSource := ts.config.TokenSource(context.Background(), ts.token)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Printf("[Microsoft OAuth2] Failed to refresh token for account %d: %v", ts.accountID, err)

		// 检查是否是刷新令牌失效（Microsoft特定错误码）
		if strings.Contains(err.Error(), "invalid_grant") ||
			strings.Contains(err.Error(), "AADSTS70008") || // Token expired
			strings.Contains(err.Error(), "AADSTS700082") || // Invalid refresh token
			strings.Contains(err.Error(), "Token has been expired or revoked") {
			// 标记令牌为需要重新授权状态
			ts.markTokenAsInvalid()
			return nil, fmt.Errorf("oauth2_refresh_token_expired: Microsoft账户的刷新令牌已过期或被撤销，需要重新授权")
		}

		return nil, fmt.Errorf("failed to refresh microsoft oauth2 token: %w", err)
	}

	// 更新内存中的令牌
	ts.token = newToken

	// 更新数据库中的令牌
	if err := ts.updateTokenInDatabase(newToken); err != nil {
		log.Printf("[Microsoft OAuth2] Failed to update token in database for account %d: %v", ts.accountID, err)
		// 不返回错误，因为令牌刷新成功了，只是数据库更新失败
	} else {
		log.Printf("[Microsoft OAuth2] Successfully refreshed and saved token for account %d", ts.accountID)
	}

	return newToken, nil
}

// updateTokenInDatabase 更新数据库中的令牌
func (ts *microsoftRefreshableTokenSource) updateTokenInDatabase(newToken *oauth2.Token) error {
	// 加密新的访问令牌
	encryptedAccessToken, err := utils.Encrypt([]byte(newToken.AccessToken))
	if err != nil {
		return fmt.Errorf("failed to encrypt new access token: %w", err)
	}

	// 更新数据库记录
	updates := map[string]interface{}{
		"access_token_encrypted": encryptedAccessToken,
		"expiry":                 newToken.Expiry,
		"token_type":             newToken.TokenType,
	}

	// 只有在提供了新的刷新令牌时才更新
	if newToken.RefreshToken != "" {
		encryptedRefreshToken, err := utils.Encrypt([]byte(newToken.RefreshToken))
		if err != nil {
			return fmt.Errorf("failed to encrypt new refresh token: %w", err)
		}
		updates["refresh_token_encrypted"] = encryptedRefreshToken
	}

	return database.DB.Model(ts.oauthToken).Updates(updates).Error
}

// markTokenAsInvalid 标记令牌为无效状态，需要重新授权
func (ts *microsoftRefreshableTokenSource) markTokenAsInvalid() {
	// 在数据库中添加一个标记字段来表示令牌需要重新授权
	// 这里我们可以设置一个很早的过期时间来标记令牌无效
	invalidTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	updates := map[string]interface{}{
		"expiry": invalidTime,
	}

	if err := database.DB.Model(ts.oauthToken).Updates(updates).Error; err != nil {
		log.Printf("[Microsoft OAuth2] Failed to mark token as invalid for account %d: %v", ts.accountID, err)
	} else {
		log.Printf("[Microsoft OAuth2] Marked token as invalid for account %d, requires re-authorization", ts.accountID)
	}
}

// FetchEmailsWithGraphAPI 是新的邮件获取实现
func FetchEmailsWithGraphAPI(emailAccount models.EmailAccount, page, pageSize int) ([]models.Email, int, error) {
	return FetchEmailsWithGraphAPIFromFolder(emailAccount, page, pageSize, "inbox")
}

// FetchEmailsWithGraphAPIFromFolder 从指定文件夹获取邮件
func FetchEmailsWithGraphAPIFromFolder(emailAccount models.EmailAccount, page, pageSize int, folderName string) ([]models.Email, int, error) {
	client, err := GetMicrosoftOAuth2HTTPClient(emailAccount.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get oauth2 client: %w", err)
	}

	// 构建Graph API请求URL，支持不同的文件夹，包含isRead字段
	skip := (page - 1) * pageSize
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/mailFolders/%s/messages?$top=%d&$skip=%d&$orderby=receivedDateTime desc&$count=true&$select=id,receivedDateTime,subject,from,toRecipients,isRead,hasAttachments", folderName, pageSize, skip)

	log.Printf("[FetchEmailsWithGraphAPIFromFolder] Requesting URL: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	// Graph API 要求返回总数时，加上这个Header
	req.Header.Add("ConsistencyLevel", "eventual")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to call graph api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("graph api returned non-200 status: %s", resp.Status)
	}

	var graphResponse GraphAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&graphResponse); err != nil {
		return nil, 0, fmt.Errorf("failed to decode graph api response: %w", err)
	}

	// 将Graph API的返回结果转换为我们自己的models.Email格式
	var emails []models.Email
	for _, msg := range graphResponse.Value {
		var from []models.EmailAddress
		if msg.From.EmailAddress.Address != "" {
			from = append(from, models.EmailAddress{
				Name:    msg.From.EmailAddress.Name,
				Address: msg.From.EmailAddress.Address,
			})
		}

		var to []models.EmailAddress
		for _, recipient := range msg.ToRecipients {
			to = append(to, models.EmailAddress{
				Name:    recipient.EmailAddress.Name,
				Address: recipient.EmailAddress.Address,
			})
		}

		emails = append(emails, models.Email{
			MessageID:     msg.ID,
			Subject:       msg.Subject,
			Date:          msg.ReceivedDateTime,
			From:          from,
			To:            to,
			IsRead:        msg.IsRead,
			HasAttachment: msg.HasAttachments,
		})

		// 调试日志：输出已读状态
		log.Printf("[Graph] Message %s IsRead: %v", msg.ID, msg.IsRead)
	}

	total := graphResponse.Count
	// 如果API没有返回总数，但返回了邮件，我们至少可以用当前获取的数量，以避免前端出问题
	if total == 0 && len(emails) > 0 {
		total = len(emails)
	}

	return emails, total, nil
}

// FetchEmailDetailWithGraphAPI fetches detailed information for a single email by messageId
func FetchEmailDetailWithGraphAPI(emailAccount models.EmailAccount, messageId string) (*models.Email, error) {
	client, err := GetMicrosoftOAuth2HTTPClient(emailAccount.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth2 client: %w", err)
	}

	// Construct Graph API request URL for a specific message
	// We need to get the full message content including body
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/messages/%s?$select=id,receivedDateTime,subject,from,toRecipients,ccRecipients,isRead,hasAttachments,body,attachments", messageId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Prefer", "outlook.body-content-type=\"html\"") // Request HTML body

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graph API returned status %d", resp.StatusCode)
	}

	var graphMessage struct {
		ID               string    `json:"id"`
		ReceivedDateTime time.Time `json:"receivedDateTime"`
		Subject          string    `json:"subject"`
		From             struct {
			EmailAddress struct {
				Name    string `json:"name"`
				Address string `json:"address"`
			} `json:"emailAddress"`
		} `json:"from"`
		ToRecipients []struct {
			EmailAddress struct {
				Name    string `json:"name"`
				Address string `json:"address"`
			} `json:"emailAddress"`
		} `json:"toRecipients"`
		CcRecipients []struct {
			EmailAddress struct {
				Name    string `json:"name"`
				Address string `json:"address"`
			} `json:"emailAddress"`
		} `json:"ccRecipients"`
		IsRead         bool `json:"isRead"`
		HasAttachments bool `json:"hasAttachments"`
		Body           struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		} `json:"body"`
		Attachments struct {
			Value []struct {
				Name        string `json:"name"`
				ContentType string `json:"contentType"`
				Size        int64  `json:"size"`
				ContentId   string `json:"contentId"`
			} `json:"value"`
		} `json:"attachments"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&graphMessage); err != nil {
		return nil, err
	}

	// Convert to our Email model
	email := &models.Email{
		MessageID:     graphMessage.ID,
		Subject:       graphMessage.Subject,
		Date:          graphMessage.ReceivedDateTime,
		IsRead:        graphMessage.IsRead,
		HasAttachment: graphMessage.HasAttachments,
	}

	// Convert From address
	if graphMessage.From.EmailAddress.Address != "" {
		email.From = []models.EmailAddress{
			{
				Name:    graphMessage.From.EmailAddress.Name,
				Address: graphMessage.From.EmailAddress.Address,
			},
		}
	}

	// Convert To addresses
	for _, to := range graphMessage.ToRecipients {
		email.To = append(email.To, models.EmailAddress{
			Name:    to.EmailAddress.Name,
			Address: to.EmailAddress.Address,
		})
	}

	// Convert Cc addresses
	for _, cc := range graphMessage.CcRecipients {
		email.Cc = append(email.Cc, models.EmailAddress{
			Name:    cc.EmailAddress.Name,
			Address: cc.EmailAddress.Address,
		})
	}

	// Set body content based on content type
	if graphMessage.Body.ContentType == "html" {
		email.HTMLBody = graphMessage.Body.Content
	} else {
		email.Body = graphMessage.Body.Content
	}

	// Convert attachments
	for _, att := range graphMessage.Attachments.Value {
		email.Attachments = append(email.Attachments, models.Attachment{
			Filename:  att.Name,
			MimeType:  att.ContentType,
			Size:      att.Size,
			ContentID: att.ContentId,
		})
	}

	return email, nil
}

// MarkMicrosoftEmailAsRead 标记Microsoft邮件为已读
func MarkMicrosoftEmailAsRead(emailAccount models.EmailAccount, messageID string) error {
	client, err := GetMicrosoftOAuth2HTTPClient(emailAccount.ID)
	if err != nil {
		return fmt.Errorf("failed to get oauth2 client: %w", err)
	}

	// Microsoft Graph API修改邮件的请求体
	requestBody := map[string]interface{}{
		"isRead": true,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 调用Microsoft Graph API修改邮件
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/messages/%s", messageID)
	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call graph api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("graph api returned non-200 status: %s", resp.Status)
	}

	log.Printf("[MarkMicrosoftEmailAsRead] Successfully marked message %s as read", messageID)
	return nil
}
