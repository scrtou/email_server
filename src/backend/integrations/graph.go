// 微软接口 Microsoft Graph API
package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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

// getOAuth2HTTPClient 是一个核心辅助函数，用于获取一个可用的、能自动刷新token的http客户端
func GetOAuth2HTTPClient(accountID uint) (*http.Client, error) {
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
		return nil, err
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
		return nil, err
	}
	decryptedRefreshToken, err := utils.Decrypt(oauthToken.RefreshTokenEncrypted)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  string(decryptedAccessToken),
		RefreshToken: string(decryptedRefreshToken),
		TokenType:    oauthToken.TokenType,
		Expiry:       oauthToken.Expiry,
	}

	// 返回一个会自动刷新token的http.Client
	return conf.Client(context.Background(), token), nil
}

// FetchEmailsWithGraphAPI 是新的邮件获取实现
func FetchEmailsWithGraphAPI(emailAccount models.EmailAccount, page, pageSize int) ([]models.Email, int, error) {
	return FetchEmailsWithGraphAPIFromFolder(emailAccount, page, pageSize, "inbox")
}

// FetchEmailsWithGraphAPIFromFolder 从指定文件夹获取邮件
func FetchEmailsWithGraphAPIFromFolder(emailAccount models.EmailAccount, page, pageSize int, folderName string) ([]models.Email, int, error) {
	client, err := GetOAuth2HTTPClient(emailAccount.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get oauth2 client: %w", err)
	}

	// 构建Graph API请求URL，支持不同的文件夹
	skip := (page - 1) * pageSize
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/mailFolders/%s/messages?$top=%d&$skip=%d&$orderby=receivedDateTime desc&$count=true", folderName, pageSize, skip)

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
			MessageID: msg.ID,
			Subject:   msg.Subject,
			Date:      msg.ReceivedDateTime,
			From:      from,
			To:        to,
		})
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
	client, err := GetOAuth2HTTPClient(emailAccount.ID)
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
