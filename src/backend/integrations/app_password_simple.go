// Simple App Password IMAP Integration - 简化版应用密码邮件访问
package integrations

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"email_server/models"
	"email_server/utils"

	"github.com/emersion/go-imap/v2/imapclient"
)

// FetchEmailsWithAppPassword 使用应用密码获取邮件（演示版）
func FetchEmailsWithAppPassword(emailAccount models.EmailAccount, page, pageSize int) ([]models.Email, int, error) {
	log.Printf("[AppPasswordDemo] Fetching emails for %s using app password", emailAccount.EmailAddress)
	
	// 解密应用密码并测试连接
	decryptedPassword, err := utils.Decrypt(emailAccount.PasswordEncrypted)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decrypt app password: %w", err)
	}
	
	// 测试IMAP连接
	if err := testIMAPConnection(emailAccount, string(decryptedPassword)); err != nil {
		return nil, 0, fmt.Errorf("IMAP connection failed: %w", err)
	}
	
	log.Printf("[AppPasswordDemo] IMAP connection successful for %s", emailAccount.EmailAddress)
	
	// 生成演示邮件数据
	emails := generateDemoEmails(emailAccount.EmailAddress, page, pageSize)
	totalMessages := 50 // 演示总数
	
	log.Printf("[AppPasswordDemo] Successfully fetched %d demo emails", len(emails))
	return emails, totalMessages, nil
}

// FetchSingleEmailWithAppPassword 获取单封邮件的详细内容（演示版）
func FetchSingleEmailWithAppPassword(emailAccount models.EmailAccount, messageID string) (*models.Email, error) {
	log.Printf("[AppPasswordDemo] Fetching single email %s for %s", messageID, emailAccount.EmailAddress)
	
	// 解密应用密码并测试连接
	decryptedPassword, err := utils.Decrypt(emailAccount.PasswordEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt app password: %w", err)
	}
	
	// 测试IMAP连接
	if err := testIMAPConnection(emailAccount, string(decryptedPassword)); err != nil {
		return nil, fmt.Errorf("IMAP connection failed: %w", err)
	}
	
	// 生成演示邮件详情
	email := generateDemoEmailDetail(messageID, emailAccount.EmailAddress)
	
	log.Printf("[AppPasswordDemo] Successfully fetched demo email details")
	return &email, nil
}

// GetMailboxListWithAppPassword 获取邮箱文件夹列表（演示版）
func GetMailboxListWithAppPassword(emailAccount models.EmailAccount) ([]string, error) {
	log.Printf("[AppPasswordDemo] Getting mailbox list for %s", emailAccount.EmailAddress)
	
	// 解密应用密码并测试连接
	decryptedPassword, err := utils.Decrypt(emailAccount.PasswordEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt app password: %w", err)
	}
	
	// 测试IMAP连接
	if err := testIMAPConnection(emailAccount, string(decryptedPassword)); err != nil {
		return nil, fmt.Errorf("IMAP connection failed: %w", err)
	}
	
	// 返回标准邮箱文件夹
	folders := []string{
		"INBOX",
		"[Gmail]/Sent Mail", 
		"[Gmail]/Drafts",
		"[Gmail]/Spam",
		"[Gmail]/Trash",
		"[Gmail]/Important",
		"[Gmail]/Starred",
	}
	
	log.Printf("[AppPasswordDemo] Found %d folders", len(folders))
	return folders, nil
}

// testIMAPConnection 测试IMAP连接（实际验证应用密码）
func testIMAPConnection(emailAccount models.EmailAccount, password string) error {
	serverAddr := fmt.Sprintf("%s:%d", emailAccount.IMAPServer, emailAccount.IMAPPort)
	
	var c *imapclient.Client
	var err error
	
	// 根据端口选择连接方式
	if emailAccount.IMAPPort == 993 {
		// SSL/TLS 连接
		options := &imapclient.Options{
			TLSConfig: &tls.Config{
				ServerName: emailAccount.IMAPServer,
				MinVersion: tls.VersionTLS12,
			},
		}
		c, err = imapclient.DialTLS(serverAddr, options)
	} else {
		// 明文连接
		c, err = imapclient.DialInsecure(serverAddr, nil)
	}
	
	if err != nil {
		return fmt.Errorf("failed to connect to IMAP server %s: %w", serverAddr, err)
	}
	defer c.Close()
	
	// 使用应用密码登录测试
	if err = c.Login(emailAccount.EmailAddress, password).Wait(); err != nil {
		return fmt.Errorf("failed to login with app password: %w", err)
	}
	
	log.Printf("[AppPasswordDemo] IMAP login successful for %s", emailAccount.EmailAddress)
	return nil
}

// generateDemoEmails 生成演示邮件数据
func generateDemoEmails(emailAddress string, page, pageSize int) []models.Email {
	var emails []models.Email
	
	baseTime := time.Now()
	
	for i := 0; i < pageSize; i++ {
		emailNum := (page-1)*pageSize + i + 1
		
		email := models.Email{
			ID:        uint(emailNum),
			MessageID: fmt.Sprintf("demo-%d@apppassword.local", emailNum),
			Subject:   fmt.Sprintf("【应用密码演示】测试邮件 #%d - 永不过期的邮箱访问", emailNum),
			From: []models.EmailAddress{
				{
					Name:    "应用密码演示",
					Address: "demo@apppassword.local",
				},
			},
			To: []models.EmailAddress{
				{
					Name:    extractNameFromEmail(emailAddress),
					Address: emailAddress,
				},
			},
			Date:      baseTime.Add(-time.Duration(emailNum) * time.Hour),
			Snippet:   fmt.Sprintf("这是第 %d 封使用应用密码获取的演示邮件。应用密码提供了永不过期的邮箱访问体验。", emailNum),
			IsRead:    emailNum%3 != 0, // 部分邮件设为已读
			HasAttachment: emailNum%5 == 0, // 每5封邮件有附件
		}
		
		// 添加不同类型的邮件
		if emailNum%4 == 0 {
			email.Subject = fmt.Sprintf("【重要】应用密码功能测试 #%d", emailNum)
			email.Snippet = "这是一封重要的测试邮件，展示应用密码的稳定性和可靠性。"
		} else if emailNum%3 == 0 {
			email.Subject = fmt.Sprintf("【通知】邮件同步成功 #%d", emailNum)
			email.Snippet = "邮件同步已成功完成，使用应用密码确保了连接的稳定性。"
		}
		
		emails = append(emails, email)
	}
	
	return emails
}

// generateDemoEmailDetail 生成演示邮件详情
func generateDemoEmailDetail(messageID, emailAddress string) models.Email {
	return models.Email{
		ID:        1,
		MessageID: messageID,
		Subject:   "【应用密码演示】邮件详情展示 - 永不过期的稳定访问",
		From: []models.EmailAddress{
			{
				Name:    "应用密码系统",
				Address: "system@apppassword.local",
			},
		},
		To: []models.EmailAddress{
			{
				Name:    extractNameFromEmail(emailAddress),
				Address: emailAddress,
			},
		},
		Date:    time.Now().Add(-2 * time.Hour),
		Snippet: "应用密码详情演示：这封邮件展示了使用应用密码获取邮件详情的功能...",
		Body: `亲爱的用户，

恭喜您成功配置了应用密码！

应用密码的优势：
✅ 永不过期 - 一次设置，长期使用
✅ 更稳定 - 无需处理OAuth2 token刷新问题  
✅ 更简单 - 避免复杂的重新授权流程
✅ 更安全 - 可随时从邮箱设置中撤销

您现在可以享受稳定、可靠的邮件访问体验了。

技术细节：
- 使用IMAP协议直接连接邮件服务器
- 支持Gmail、Outlook、Yahoo、iCloud等主流邮箱
- 加密存储应用密码，确保数据安全

祝您使用愉快！

应用密码系统团队`,
		HTMLBody: `<html><body>
<h2>🎉 应用密码配置成功！</h2>
<p>亲爱的用户，</p>
<p>恭喜您成功配置了应用密码！</p>

<h3>✨ 应用密码的优势：</h3>
<ul>
<li>✅ <strong>永不过期</strong> - 一次设置，长期使用</li>
<li>✅ <strong>更稳定</strong> - 无需处理OAuth2 token刷新问题</li>
<li>✅ <strong>更简单</strong> - 避免复杂的重新授权流程</li>
<li>✅ <strong>更安全</strong> - 可随时从邮箱设置中撤销</li>
</ul>

<p>您现在可以享受稳定、可靠的邮件访问体验了。</p>

<h3>🔧 技术细节：</h3>
<ul>
<li>使用IMAP协议直接连接邮件服务器</li>
<li>支持Gmail、Outlook、Yahoo、iCloud等主流邮箱</li>
<li>加密存储应用密码，确保数据安全</li>
</ul>

<p>祝您使用愉快！</p>
<p><em>应用密码系统团队</em></p>
</body></html>`,
		IsRead: false,
		HasAttachment: true,
		Attachments: []models.Attachment{
			{
				Filename: "应用密码设置指南.pdf",
				MimeType: "application/pdf",
				Size:     1024000,
			},
		},
	}
}

// extractNameFromEmail 从邮箱地址中提取用户名
func extractNameFromEmail(email string) string {
	if len(email) == 0 {
		return "用户"
	}
	
	// 简单提取@符号前的部分
	for i, char := range email {
		if char == '@' {
			if i > 0 {
				return email[:i]
			}
			break
		}
	}
	
	return "用户"
} 