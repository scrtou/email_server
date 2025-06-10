package integrations

import (
	"context"
	"email_server/database"
	"email_server/models"
	"email_server/utils"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type xoauth2Client struct {
	user  string
	token string
}

func (a *xoauth2Client) Start() (string, []byte, error) {
	// The client-first mechanism for XOAUTH2 requires sending the initial
	// response formatted as "user=<user>\x01auth=Bearer <token>\x01\x01"
	initialResponse := []byte(fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.user, a.token))
	return "XOAUTH2", initialResponse, nil
}

func (a *xoauth2Client) Next(challenge []byte) ([]byte, error) {
	// XOAUTH2 is a client-first mechanism, so we don't expect a challenge.
	// If we receive one, it's likely an error from the server.
	// The spec says the client should respond with a single CRLF.
	return nil, fmt.Errorf("unexpected server challenge: %s", string(challenge))
}

// connectAndLogin handles the connection and authentication logic.
func connectAndLogin(emailAccount models.EmailAccount, token *models.UserOAuthToken) (*imapclient.Client, error) {
	var imapServer, imapPort = emailAccount.IMAPServer, emailAccount.IMAPPort

	// If it's an OAuth2 connection, we need to fetch the provider's IMAP details from the DB.
	if token != nil {
		var provider models.OAuthProvider
		if err := database.DB.First(&provider, token.ProviderID).Error; err != nil {
			return nil, fmt.Errorf("failed to find oauth provider with id %d for imap details: %w", token.ProviderID, err)
		}
		imapServer = provider.IMAPServer
		imapPort = provider.IMAPPort
	}

	imapServerAddr := fmt.Sprintf("%s:%d", imapServer, imapPort)
	authMethod := "Password"
	if token != nil {
		authMethod = "XOAUTH2"
	}
	log.Printf("Connecting to IMAP server: %s for user %s (Auth: %s)", imapServerAddr, emailAccount.EmailAddress, authMethod)

	c, err := imapclient.DialTLS(imapServerAddr, &imapclient.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
	}

	if token != nil {
		// OAuth2 authentication
		decryptedAccessToken, err := utils.Decrypt(token.AccessTokenEncrypted)
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("could not decrypt access token: %w", err)
		}

		decryptedRefreshToken, err := utils.Decrypt(token.RefreshTokenEncrypted)
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("could not decrypt refresh token: %w", err)
		}

		oauth2Token := &oauth2.Token{
			AccessToken:  string(decryptedAccessToken),
			RefreshToken: string(decryptedRefreshToken),
			Expiry:       token.Expiry,
			TokenType:    token.TokenType,
		}

		// Check if the token is expired and refresh if necessary
		if !oauth2Token.Valid() {
			log.Printf("OAuth2 token for %s is expired, refreshing...", emailAccount.EmailAddress)

			var provider models.OAuthProvider
			if err := database.DB.First(&provider, token.ProviderID).Error; err != nil {
				c.Close()
				return nil, fmt.Errorf("failed to find oauth provider with id %d: %w", token.ProviderID, err)
			}

			decryptedSecret, err := utils.Decrypt(provider.ClientSecretEncrypted)
			if err != nil {
				c.Close()
				return nil, fmt.Errorf("could not decrypt provider client secret: %w", err)
			}

			conf := &oauth2.Config{
				ClientID:     provider.ClientID,
				ClientSecret: string(decryptedSecret),
				Scopes:       strings.Split(provider.Scopes, ","), // <-- 多了这一行
				Endpoint: oauth2.Endpoint{
					AuthURL:  provider.AuthURL,
					TokenURL: provider.TokenURL,
				},
			}

			tokenSource := conf.TokenSource(context.Background(), oauth2Token)
			newToken, err := tokenSource.Token()
			if err != nil {
				c.Close()
				return nil, fmt.Errorf("failed to refresh token: %w", err)
			}

			// Update the local token variable and the database
			oauth2Token = newToken
			token.AccessTokenEncrypted, err = utils.Encrypt([]byte(newToken.AccessToken))
			if err != nil {
				c.Close()
				return nil, fmt.Errorf("failed to encrypt new access token: %w", err)
			}
			// Only update refresh token if a new one was provided
			if newToken.RefreshToken != "" {
				token.RefreshTokenEncrypted, err = utils.Encrypt([]byte(newToken.RefreshToken))
				if err != nil {
					c.Close()
					return nil, fmt.Errorf("failed to encrypt new refresh token: %w", err)
				}
			}
			token.Expiry = newToken.Expiry

			if err := database.DB.Save(token).Error; err != nil {
				c.Close()
				return nil, fmt.Errorf("failed to save updated token to database: %w", err)
			}
			log.Printf("Successfully refreshed and saved new token for %s", emailAccount.EmailAddress)
		}

		log.Printf("Attempting XOAUTH2 login for %s", emailAccount.EmailAddress)
		log.Printf("--- FULL ACCESS TOKEN FOR DEBUG ---\n%s\n---------------------------------", oauth2Token.AccessToken)
		auth := &xoauth2Client{
			user:  emailAccount.EmailAddress,
			token: oauth2Token.AccessToken,
		}
		if err := c.Authenticate(auth); err != nil {
			c.Close()
			return nil, fmt.Errorf("XOAUTH2 login failed: %w", err)
		}
		log.Printf("Successfully logged in with XOAUTH2 for %s", emailAccount.EmailAddress)
	} else {
		// Password (PLAIN) authentication
		password, err := utils.DecryptPassword(emailAccount.PasswordEncrypted)
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("could not decrypt password: %w", err)
		}
		if password == "" {
			c.Close()
			return nil, fmt.Errorf("password for email account %s is not set", emailAccount.EmailAddress)
		}

		if err := c.Login(emailAccount.EmailAddress, password).Wait(); err != nil {
			c.Close()
			return nil, fmt.Errorf("IMAP login failed: %w", err)
		}
		log.Printf("Successfully logged in with password for %s", emailAccount.EmailAddress)
	}

	return c, nil
}

// parseEmailContent 解析邮件内容，支持MIME格式和编码解析
func parseEmailContent(rawContent string) (textBody, htmlBody string, err error) {
	// 创建一个reader来解析邮件
	reader := strings.NewReader(rawContent)

	// 解析邮件消息
	msg, err := mail.ReadMessage(reader)
	if err != nil {
		log.Printf("Failed to parse email message: %v", err)
		// 如果解析失败，返回原始内容作为文本
		return rawContent, "", nil
	}

	// 获取Content-Type头
	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		// 如果没有Content-Type，假设是纯文本
		body, err := io.ReadAll(msg.Body)
		if err != nil {
			return rawContent, "", nil
		}
		return string(body), "", nil
	}

	// 解析Content-Type
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.Printf("Failed to parse content type: %v", err)
		body, _ := io.ReadAll(msg.Body)
		return string(body), "", nil
	}

	// 处理不同的媒体类型
	switch {
	case strings.HasPrefix(mediaType, "text/plain"):
		body, err := decodeBody(msg.Body, msg.Header.Get("Content-Transfer-Encoding"))
		if err != nil {
			log.Printf("Failed to decode plain text body: %v", err)
			rawBody, _ := io.ReadAll(msg.Body)
			return string(rawBody), "", nil
		}
		return body, "", nil

	case strings.HasPrefix(mediaType, "text/html"):
		body, err := decodeBody(msg.Body, msg.Header.Get("Content-Transfer-Encoding"))
		if err != nil {
			log.Printf("Failed to decode HTML body: %v", err)
			rawBody, _ := io.ReadAll(msg.Body)
			return "", string(rawBody), nil
		}
		return "", body, nil

	case strings.HasPrefix(mediaType, "multipart/"):
		return parseMultipartContent(msg.Body, params["boundary"])

	default:
		// 对于其他类型，尝试读取原始内容
		body, _ := io.ReadAll(msg.Body)
		return string(body), "", nil
	}
}

// parseMultipartContent 解析多部分邮件内容
func parseMultipartContent(body io.Reader, boundary string) (textBody, htmlBody string, err error) {
	if boundary == "" {
		rawBody, _ := io.ReadAll(body)
		return string(rawBody), "", nil
	}

	mr := multipart.NewReader(body, boundary)

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading multipart: %v", err)
			break
		}

		// 获取部分的Content-Type
		partContentType := part.Header.Get("Content-Type")
		mediaType, _, err := mime.ParseMediaType(partContentType)
		if err != nil {
			log.Printf("Failed to parse part content type: %v", err)
			part.Close()
			continue
		}

		// 解码部分内容
		encoding := part.Header.Get("Content-Transfer-Encoding")
		content, err := decodeBody(part, encoding)
		if err != nil {
			log.Printf("Failed to decode part body: %v", err)
			part.Close()
			continue
		}

		// 根据媒体类型分配内容
		switch {
		case strings.HasPrefix(mediaType, "text/plain"):
			if textBody == "" { // 只取第一个文本部分
				textBody = content
			}
		case strings.HasPrefix(mediaType, "text/html"):
			if htmlBody == "" { // 只取第一个HTML部分
				htmlBody = content
			}
		case strings.HasPrefix(mediaType, "multipart/"):
			// 递归处理嵌套的multipart
			nestedText, nestedHTML, _ := parseMultipartContent(part, "")
			if textBody == "" && nestedText != "" {
				textBody = nestedText
			}
			if htmlBody == "" && nestedHTML != "" {
				htmlBody = nestedHTML
			}
		}

		part.Close()
	}

	return textBody, htmlBody, nil
}

// decodeBody 根据编码类型解码邮件正文
func decodeBody(body io.Reader, encoding string) (string, error) {
	var reader io.Reader = body

	// 根据编码类型创建相应的解码器
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "quoted-printable":
		reader = quotedprintable.NewReader(body)
	case "base64":
		// 对于base64，我们需要先读取所有内容然后解码
		content, err := io.ReadAll(body)
		if err != nil {
			return "", err
		}
		// 移除换行符
		cleanContent := strings.ReplaceAll(string(content), "\n", "")
		cleanContent = strings.ReplaceAll(cleanContent, "\r", "")

		// 解码base64
		decoded, err := base64.StdEncoding.DecodeString(cleanContent)
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	case "7bit", "8bit", "binary", "":
		// 这些编码不需要特殊处理
		reader = body
	default:
		log.Printf("Unknown encoding: %s, treating as plain text", encoding)
		reader = body
	}

	// 读取解码后的内容
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// FetchEmails connects to an IMAP server and fetches emails with pagination.
func FetchEmails(emailAccount models.EmailAccount, page, pageSize int) ([]models.Email, int, error) {
	var token models.UserOAuthToken
	err := database.DB.Where("email_account_id = ?", emailAccount.ID).First(&token).Error

	var c *imapclient.Client
	if err == nil {
		// Token found, use OAuth2
		c, err = connectAndLogin(emailAccount, &token)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// No token found, use password
		log.Printf("No OAuth token found for %s, falling back to password authentication.", emailAccount.EmailAddress)
		c, err = connectAndLogin(emailAccount, nil)
	} else {
		// Database error
		return nil, 0, fmt.Errorf("failed to query for oauth token: %w", err)
	}

	if err != nil {
		log.Printf("Failed to connect or login for %s: %v", emailAccount.EmailAddress, err)
		return nil, 0, err
	}
	defer c.Close()

	// Select INBOX
	mailbox, err := c.Select("INBOX", nil).Wait()
	if err != nil {
		log.Printf("Failed to select INBOX for %s: %v", emailAccount.EmailAddress, err)
		return nil, 0, fmt.Errorf("failed to select INBOX: %w", err)
	}
	totalMessages := int(mailbox.NumMessages)
	log.Printf("INBOX selected. Total messages: %d", totalMessages)

	if totalMessages == 0 {
		return []models.Email{}, 0, nil
	}

	// Calculate message sequence numbers for the requested page
	start := totalMessages - (page * pageSize) + 1
	end := totalMessages - ((page - 1) * pageSize)

	if start < 1 {
		start = 1
	}
	if end < 1 || start > end {
		log.Printf("No messages to fetch for page %d, pageSize %d", page, pageSize)
		return []models.Email{}, totalMessages, nil
	}

	// Create a sequence set using the correct API for go-imap v2
	var seqSet imap.SeqSet
	seqSet.AddRange(uint32(start), uint32(end))
	log.Printf("Fetching messages in range: %d-%d", start, end)

	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		Flags:    true,
		// Fetching body sections can be added here if needed
	}

	// Fetch messages using Collect, which waits for the command to finish.
	messages, err := c.Fetch(seqSet, fetchOptions).Collect()
	if err != nil {
		log.Printf("Failed to fetch emails for %s: %v", emailAccount.EmailAddress, err)
		return nil, 0, fmt.Errorf("IMAP fetch command failed: %w", err)
	}

	var emails []models.Email
	// Reverse the order of messages to have the newest first
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg == nil {
			log.Println("Received a nil message from fetch command")
			continue
		}

		from := make([]models.EmailAddress, len(msg.Envelope.From))
		for i, addr := range msg.Envelope.From {
			from[i] = models.EmailAddress{Name: addr.Mailbox, Address: addr.Host}
		}

		to := make([]models.EmailAddress, len(msg.Envelope.To))
		for i, addr := range msg.Envelope.To {
			to[i] = models.EmailAddress{Name: addr.Mailbox, Address: addr.Host}
		}

		date := time.Time{}
		if !msg.Envelope.Date.IsZero() {
			date = msg.Envelope.Date
		}

		email := models.Email{
			MessageID: msg.Envelope.MessageID,
			Subject:   msg.Envelope.Subject,
			From:      from,
			To:        to,
			Date:      date,
		}
		emails = append(emails, email)
	}

	log.Printf("Successfully fetched %d emails for %s", len(emails), emailAccount.EmailAddress)
	return emails, totalMessages, nil
}

// FetchEmailDetailWithIMAP fetches a single email's detailed information using IMAP
// FetchEmailDetailWithIMAP fetches a single email's detailed information using IMAP
func FetchEmailDetailWithIMAP(emailAccount models.EmailAccount, messageID string) (*models.Email, error) {
	var token models.UserOAuthToken
	err := database.DB.Where("email_account_id = ?", emailAccount.ID).First(&token).Error

	var c *imapclient.Client
	if err == nil {
		c, err = connectAndLogin(emailAccount, &token)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("No OAuth token found for %s, falling back to password authentication.", emailAccount.EmailAddress)
		c, err = connectAndLogin(emailAccount, nil)
	} else {
		return nil, fmt.Errorf("failed to query for oauth token: %w", err)
	}

	if err != nil {
		log.Printf("Failed to connect or login for %s: %v", emailAccount.EmailAddress, err)
		return nil, err
	}
	defer c.Close()

	_, err = c.Select("INBOX", nil).Wait()
	if err != nil {
		log.Printf("Failed to select INBOX for %s: %v", emailAccount.EmailAddress, err)
		return nil, fmt.Errorf("failed to select INBOX: %w", err)
	}

	// --- NEW ROBUST STRATEGY: Scan-then-Fetch ---

	// Step 1: Scan all message envelopes to find the sequence number.
	log.Printf("IMAP SEARCH is unreliable on this server. Starting full envelope scan for Message-ID: %s", messageID)
	var allMessagesSeqSet imap.SeqSet
	allMessagesSeqSet.AddRange(1, 0) // 0 means the last message

	scanFetchOptions := &imap.FetchOptions{Envelope: true}
	envelopes, err := c.Fetch(allMessagesSeqSet, scanFetchOptions).Collect()
	if err != nil {
		return nil, fmt.Errorf("failed to scan envelopes: %w", err)
	}

	var targetSeqNum uint32
	for _, msg := range envelopes {
		if msg.Envelope != nil && msg.Envelope.MessageID == messageID {
			targetSeqNum = msg.SeqNum
			log.Printf("Found message with ID %s at sequence number: %d", messageID, targetSeqNum)
			break
		}
	}

	if targetSeqNum == 0 {
		log.Printf("Message with ID %s not found after scanning all envelopes.", messageID)
		return nil, fmt.Errorf("message with ID %s not found", messageID)
	}

	// Step 2: Fetch the full details for the found sequence number.
	var detailSeqSet imap.SeqSet
	detailSeqSet.AddNum(targetSeqNum)

	// 获取完整的邮件内容，包括所有部分
	fullBodySectionItem := &imap.FetchItemBodySection{Specifier: imap.PartSpecifierNone}
	detailFetchOptions := &imap.FetchOptions{
		Envelope:      true,
		Flags:         true,
		BodyStructure: &imap.FetchItemBodyStructure{Extended: true},
		BodySection:   []*imap.FetchItemBodySection{fullBodySectionItem},
	}

	detailMessages, err := c.Fetch(detailSeqSet, detailFetchOptions).Collect()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch details for seqnum %d: %w", targetSeqNum, err)
	}

	if len(detailMessages) == 0 {
		return nil, fmt.Errorf("message with seqnum %d vanished after being found", targetSeqNum)
	}

	// Step 3: Process the detailed message.
	msg := detailMessages[0]

	from := make([]models.EmailAddress, len(msg.Envelope.From))
	for i, addr := range msg.Envelope.From {
		from[i] = models.EmailAddress{Name: addr.Mailbox, Address: fmt.Sprintf("%s@%s", addr.Mailbox, addr.Host)}
	}

	to := make([]models.EmailAddress, len(msg.Envelope.To))
	for i, addr := range msg.Envelope.To {
		to[i] = models.EmailAddress{Name: addr.Mailbox, Address: fmt.Sprintf("%s@%s", addr.Mailbox, addr.Host)}
	}

	date := time.Time{}
	if !msg.Envelope.Date.IsZero() {
		date = msg.Envelope.Date
	}

	// 获取原始邮件内容并解析
	var textBody, htmlBody string
	if bodyBytes := msg.FindBodySection(fullBodySectionItem); bodyBytes != nil {
		rawContent := string(bodyBytes)
		log.Printf("Raw email content length: %d", len(rawContent))

		// 使用新的MIME解析函数
		parsedText, parsedHTML, err := parseEmailContent(rawContent)
		if err != nil {
			log.Printf("Failed to parse email content: %v", err)
			// 如果解析失败，使用原始内容
			textBody = rawContent
		} else {
			textBody = parsedText
			htmlBody = parsedHTML
			log.Printf("Parsed email - Text length: %d, HTML length: %d", len(textBody), len(htmlBody))
		}
	}

	email := &models.Email{
		MessageID:     msg.Envelope.MessageID,
		Subject:       msg.Envelope.Subject,
		From:          from,
		To:            to,
		Date:          date,
		Body:          textBody,
		HTMLBody:      htmlBody,
		IsRead:        false,
		HasAttachment: false,
	}

	for _, flag := range msg.Flags {
		if flag == imap.FlagSeen {
			email.IsRead = true
			break
		}
	}

	if msg.BodyStructure != nil {
		fullMediaType := msg.BodyStructure.MediaType()
		if strings.HasPrefix(strings.ToLower(fullMediaType), "multipart/") {
			email.HasAttachment = true
		}
	}

	log.Printf("Successfully fetched email detail for messageID: %s using scan-then-fetch method.", messageID)
	return email, nil
}
