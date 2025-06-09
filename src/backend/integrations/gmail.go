// Gmail API 集成
package integrations

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"email_server/models"
)

// GmailMessage 对应 Gmail API 返回的邮件结构
type GmailMessage struct {
	ID           string           `json:"id"`
	ThreadID     string           `json:"threadId"`
	LabelIDs     []string         `json:"labelIds"`
	Snippet      string           `json:"snippet"`
	HistoryID    string           `json:"historyId"`
	InternalDate string           `json:"internalDate"`
	Payload      GmailMessagePart `json:"payload"`
	SizeEstimate int              `json:"sizeEstimate"`
}

type GmailMessagePart struct {
	PartID   string             `json:"partId"`
	MimeType string             `json:"mimeType"`
	Filename string             `json:"filename"`
	Headers  []GmailHeader      `json:"headers"`
	Body     GmailMessageBody   `json:"body"`
	Parts    []GmailMessagePart `json:"parts"`
}

type GmailHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type GmailMessageBody struct {
	AttachmentID string `json:"attachmentId"`
	Size         int    `json:"size"`
	Data         string `json:"data"`
}

// GmailListResponse 对应 Gmail API 列表请求的响应结构
type GmailListResponse struct {
	Messages           []GmailMessageRef `json:"messages"`
	NextPageToken      string            `json:"nextPageToken"`
	ResultSizeEstimate int               `json:"resultSizeEstimate"`
}

type GmailMessageRef struct {
	ID       string `json:"id"`
	ThreadID string `json:"threadId"`
}

// FetchEmailsWithGmailAPI 使用Gmail API获取邮件列表
func FetchEmailsWithGmailAPI(emailAccount models.EmailAccount, page, pageSize int) ([]models.Email, int, error) {
	return FetchEmailsWithGmailAPIFromFolder(emailAccount, page, pageSize, "INBOX")
}

// FetchEmailsWithGmailAPIFromFolder 从指定标签/文件夹获取邮件
func FetchEmailsWithGmailAPIFromFolder(emailAccount models.EmailAccount, page, pageSize int, labelName string) ([]models.Email, int, error) {
	client, err := GetOAuth2HTTPClient(emailAccount.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get oauth2 client: %w", err)
	}

	// 构建Gmail API请求URL
	// Gmail API使用pageToken而不是skip/offset
	url := fmt.Sprintf("https://gmail.googleapis.com/gmail/v1/users/me/messages?labelIds=%s&maxResults=%d", labelName, pageSize)

	// 如果不是第一页，需要处理分页token
	// 这里简化处理，实际应用中需要存储和管理pageToken
	if page > 1 {
		// Gmail API的分页比较复杂，这里先简化实现
		log.Printf("[FetchEmailsWithGmailAPIFromFolder] Gmail API pagination not fully implemented for page %d", page)
	}

	log.Printf("[FetchEmailsWithGmailAPIFromFolder] Requesting URL: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to call gmail api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("gmail api returned non-200 status: %s", resp.Status)
	}

	var listResponse GmailListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, 0, fmt.Errorf("failed to decode gmail api response: %w", err)
	}

	// 并发获取每个邮件的详细信息以提高性能
	var emails []models.Email
	emailChan := make(chan *models.Email, len(listResponse.Messages))
	errorChan := make(chan error, len(listResponse.Messages))

	// 限制并发数量以避免API配额问题
	maxConcurrent := 5
	semaphore := make(chan struct{}, maxConcurrent)

	for _, msgRef := range listResponse.Messages {
		go func(messageID string) {
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			email, err := fetchGmailMessageDetail(client, messageID)
			if err != nil {
				log.Printf("Failed to fetch message detail for %s: %v", messageID, err)
				errorChan <- err
				return
			}
			emailChan <- email
		}(msgRef.ID)
	}

	// 收集结果
	for i := 0; i < len(listResponse.Messages); i++ {
		select {
		case email := <-emailChan:
			if email != nil {
				emails = append(emails, *email)
			}
		case <-errorChan:
			// 记录错误但继续处理其他邮件
		}
	}

	total := listResponse.ResultSizeEstimate
	if total == 0 && len(emails) > 0 {
		total = len(emails)
	}

	return emails, total, nil
}

// fetchGmailMessageDetail 获取单个邮件的详细信息
func fetchGmailMessageDetail(client *http.Client, messageID string) (*models.Email, error) {
	// 明确请求包含标签信息的完整邮件数据
	url := fmt.Sprintf("https://gmail.googleapis.com/gmail/v1/users/me/messages/%s?format=full", messageID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gmail api returned status %d", resp.StatusCode)
	}

	var gmailMessage GmailMessage
	if err := json.NewDecoder(resp.Body).Decode(&gmailMessage); err != nil {
		return nil, err
	}

	return convertGmailMessageToEmail(&gmailMessage), nil
}

// convertGmailMessageToEmail 将Gmail API的消息转换为我们的Email模型
func convertGmailMessageToEmail(gmailMsg *GmailMessage) *models.Email {
	email := &models.Email{
		MessageID: gmailMsg.ID,
		Subject:   getHeaderValue(gmailMsg.Payload.Headers, "Subject"),
	}

	// 解析日期
	if dateStr := getHeaderValue(gmailMsg.Payload.Headers, "Date"); dateStr != "" {
		if parsedDate, err := time.Parse(time.RFC1123Z, dateStr); err == nil {
			email.Date = parsedDate
		} else if parsedDate, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", dateStr); err == nil {
			email.Date = parsedDate
		}
	}

	// 解析发件人
	if fromStr := getHeaderValue(gmailMsg.Payload.Headers, "From"); fromStr != "" {
		email.From = parseEmailAddresses(fromStr)
	}

	// 解析收件人
	if toStr := getHeaderValue(gmailMsg.Payload.Headers, "To"); toStr != "" {
		email.To = parseEmailAddresses(toStr)
	}

	// 解析抄送
	if ccStr := getHeaderValue(gmailMsg.Payload.Headers, "Cc"); ccStr != "" {
		email.Cc = parseEmailAddresses(ccStr)
	}

	// 检查是否已读（Gmail使用标签系统）
	email.IsRead = !contains(gmailMsg.LabelIDs, "UNREAD")

	// 调试日志：输出标签信息
	log.Printf("[Gmail] Message %s labels: %v, IsRead: %v", gmailMsg.ID, gmailMsg.LabelIDs, email.IsRead)

	// 检查是否有附件
	email.HasAttachment = hasAttachments(&gmailMsg.Payload)

	// 提取邮件正文
	extractEmailBody(&gmailMsg.Payload, email)

	return email
}

// getHeaderValue 从headers中获取指定名称的值
func getHeaderValue(headers []GmailHeader, name string) string {
	for _, header := range headers {
		if strings.EqualFold(header.Name, name) {
			return header.Value
		}
	}
	return ""
}

// parseEmailAddresses 解析邮件地址字符串
func parseEmailAddresses(addressStr string) []models.EmailAddress {
	// 简化的邮件地址解析，实际应用中可能需要更复杂的解析逻辑
	addresses := strings.Split(addressStr, ",")
	var result []models.EmailAddress

	for _, addr := range addresses {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}

		// 简单解析 "Name <email@domain.com>" 格式
		if strings.Contains(addr, "<") && strings.Contains(addr, ">") {
			parts := strings.Split(addr, "<")
			if len(parts) == 2 {
				name := strings.TrimSpace(strings.Trim(parts[0], "\""))
				email := strings.TrimSpace(strings.Trim(parts[1], ">"))
				result = append(result, models.EmailAddress{
					Name:    name,
					Address: email,
				})
			}
		} else {
			result = append(result, models.EmailAddress{
				Address: addr,
			})
		}
	}

	return result
}

// contains 检查slice中是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// hasAttachments 检查邮件是否有附件
func hasAttachments(part *GmailMessagePart) bool {
	if part.Filename != "" && part.Body.AttachmentID != "" {
		return true
	}

	for _, subPart := range part.Parts {
		if hasAttachments(&subPart) {
			return true
		}
	}

	return false
}

// extractEmailBody 提取邮件正文
func extractEmailBody(part *GmailMessagePart, email *models.Email) {
	if part.MimeType == "text/plain" && part.Body.Data != "" {
		if decoded, err := decodeBase64URL(part.Body.Data); err == nil {
			email.Body = string(decoded)
		}
	} else if part.MimeType == "text/html" && part.Body.Data != "" {
		if decoded, err := decodeBase64URL(part.Body.Data); err == nil {
			email.HTMLBody = string(decoded)
		}
	}

	// 递归处理子部分
	for _, subPart := range part.Parts {
		extractEmailBody(&subPart, email)
	}
}

// decodeBase64URL 解码Gmail API使用的base64url编码
func decodeBase64URL(data string) ([]byte, error) {
	// Gmail API使用base64url编码，需要特殊处理
	data = strings.ReplaceAll(data, "-", "+")
	data = strings.ReplaceAll(data, "_", "/")

	// 添加必要的padding
	switch len(data) % 4 {
	case 2:
		data += "=="
	case 3:
		data += "="
	}

	return base64.StdEncoding.DecodeString(data)
}

// FetchGmailMessageDetail 获取Gmail邮件的详细信息
func FetchGmailMessageDetail(emailAccount models.EmailAccount, messageID string) (*models.Email, error) {
	client, err := GetOAuth2HTTPClient(emailAccount.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth2 client: %w", err)
	}

	return fetchGmailMessageDetail(client, messageID)
}

// MarkGmailAsRead 标记Gmail邮件为已读
func MarkGmailAsRead(emailAccount models.EmailAccount, messageID string) error {
	client, err := GetOAuth2HTTPClient(emailAccount.ID)
	if err != nil {
		return fmt.Errorf("failed to get oauth2 client: %w", err)
	}

	// Gmail API修改邮件标签的请求体
	requestBody := map[string]interface{}{
		"removeLabelIds": []string{"UNREAD"},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 调用Gmail API修改邮件标签
	url := fmt.Sprintf("https://gmail.googleapis.com/gmail/v1/users/me/messages/%s/modify", messageID)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call gmail api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gmail api returned non-200 status: %s", resp.Status)
	}

	log.Printf("[MarkGmailAsRead] Successfully marked message %s as read", messageID)
	return nil
}
