// 微软接口 Microsoft Graph API
package integrations

import (
	"context"
	"encoding/json"
	"fmt"
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
	client, err := GetOAuth2HTTPClient(emailAccount.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get oauth2 client: %w", err)
	}

	// 构建Graph API请求URL
	skip := (page - 1) * pageSize
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/mailFolders/inbox/messages?$top=%d&$skip=%d&$orderby=receivedDateTime desc&$count=true", pageSize, skip)

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

	return emails, graphResponse.Count, nil
}
