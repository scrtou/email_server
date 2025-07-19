// App Password Manager - 应用密码管理器
// 支持使用Gmail、Outlook等服务商的应用密码，实现永不过期的邮箱访问
package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/mail"
	"strings"
	"time"

	"email_server/database"
	"email_server/models"
	"email_server/utils"

	"github.com/emersion/go-imap/v2/imapclient"
)

// AppPasswordManager 应用密码管理器
type AppPasswordManager struct {
	// 支持的邮件服务商配置
	providers map[string]ProviderConfig
}

// ProviderConfig 邮件服务商配置
type ProviderConfig struct {
	Name             string `json:"name"`
	IMAPServer       string `json:"imap_server"`
	IMAPPort         int    `json:"imap_port"`
	SMTPServer       string `json:"smtp_server"`
	SMTPPort         int    `json:"smtp_port"`
	AppPasswordGuide string `json:"app_password_guide"`
	DocumentationURL string `json:"documentation_url"`
	SupportsTLS      bool   `json:"supports_tls"`
}

// AppPasswordSetupRequest 应用密码设置请求
type AppPasswordSetupRequest struct {
	EmailAddress string `json:"email_address" binding:"required,email"`
	AppPassword  string `json:"app_password" binding:"required"`
	Provider     string `json:"provider"`
}

// AppPasswordTestResult 应用密码测试结果
type AppPasswordTestResult struct {
	Success       bool     `json:"success"`
	EmailAddress  string   `json:"email_address"`
	Provider      string   `json:"provider"`
	Message       string   `json:"message"`
	Capabilities  []string `json:"capabilities,omitempty"`
	TestTimestamp string   `json:"test_timestamp"`
}

// NewAppPasswordManager 创建应用密码管理器
func NewAppPasswordManager() *AppPasswordManager {
	return &AppPasswordManager{
		providers: map[string]ProviderConfig{
			"gmail": {
				Name:             "Gmail",
				IMAPServer:       "imap.gmail.com",
				IMAPPort:         993,
				SMTPServer:       "smtp.gmail.com",
				SMTPPort:         587,
				AppPasswordGuide: "前往 Google 账户设置 > 安全性 > 两步验证 > 应用密码，生成16位应用密码",
				DocumentationURL: "https://support.google.com/accounts/answer/185833",
				SupportsTLS:      true,
			},
			"outlook": {
				Name:             "Outlook/Hotmail",
				IMAPServer:       "outlook.office365.com",
				IMAPPort:         993,
				SMTPServer:       "smtp-mail.outlook.com",
				SMTPPort:         587,
				AppPasswordGuide: "前往 Microsoft 账户设置 > 安全性 > 应用密码，生成应用密码",
				DocumentationURL: "https://support.microsoft.com/account-billing/using-app-passwords-with-apps-that-don-t-support-two-step-verification-5896ed9b-4263-e681-128a-a6f2979a7944",
				SupportsTLS:      true,
			},
			"yahoo": {
				Name:             "Yahoo Mail",
				IMAPServer:       "imap.mail.yahoo.com",
				IMAPPort:         993,
				SMTPServer:       "smtp.mail.yahoo.com",
				SMTPPort:         587,
				AppPasswordGuide: "前往 Yahoo 账户设置 > 账户安全 > 应用密码，生成应用密码",
				DocumentationURL: "https://help.yahoo.com/kb/generate-manage-third-party-passwords-sln15241.html",
				SupportsTLS:      true,
			},
			"icloud": {
				Name:             "iCloud Mail",
				IMAPServer:       "imap.mail.me.com",
				IMAPPort:         993,
				SMTPServer:       "smtp.mail.me.com",
				SMTPPort:         587,
				AppPasswordGuide: "前往 Apple ID 设置 > 登录与安全 > 应用专用密码，生成应用专用密码",
				DocumentationURL: "https://support.apple.com/zh-cn/HT204397",
				SupportsTLS:      true,
			},
		},
	}
}

// DetectProvider 自动检测邮箱服务商
func (apm *AppPasswordManager) DetectProvider(email string) string {
	address, err := mail.ParseAddress(email)
	if err != nil {
		return ""
	}

	domain := strings.ToLower(address.Address)

	if strings.Contains(domain, "@gmail.com") {
		return "gmail"
	} else if strings.Contains(domain, "@outlook.com") || strings.Contains(domain, "@hotmail.com") || strings.Contains(domain, "@live.com") {
		return "outlook"
	} else if strings.Contains(domain, "@yahoo.com") || strings.Contains(domain, "@ymail.com") {
		return "yahoo"
	} else if strings.Contains(domain, "@icloud.com") || strings.Contains(domain, "@me.com") || strings.Contains(domain, "@mac.com") {
		return "icloud"
	}

	return ""
}

// GetProviderConfig 获取服务商配置
func (apm *AppPasswordManager) GetProviderConfig(provider string) (ProviderConfig, bool) {
	config, exists := apm.providers[strings.ToLower(provider)]
	return config, exists
}

// GetSupportedProviders 获取支持的服务商列表
func (apm *AppPasswordManager) GetSupportedProviders() map[string]ProviderConfig {
	return apm.providers
}

// SetupAppPassword 设置应用密码
func (apm *AppPasswordManager) SetupAppPassword(userID uint, req AppPasswordSetupRequest) error {
	// 检测服务商
	if req.Provider == "" {
		req.Provider = apm.DetectProvider(req.EmailAddress)
		if req.Provider == "" {
			return fmt.Errorf("无法自动检测邮箱服务商，请手动指定")
		}
	}

	// 获取服务商配置
	config, exists := apm.GetProviderConfig(req.Provider)
	if !exists {
		return fmt.Errorf("不支持的邮箱服务商: %s", req.Provider)
	}

	// 测试应用密码
	testResult := apm.TestAppPassword(req.EmailAddress, req.AppPassword, config)
	if !testResult.Success {
		return fmt.Errorf("应用密码测试失败: %s", testResult.Message)
	}

	// 加密应用密码
	encryptedPassword, err := utils.Encrypt([]byte(req.AppPassword))
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 查找或创建邮箱账户
	var emailAccount models.EmailAccount
	result := database.DB.Where("user_id = ? AND email_address = ?", userID, req.EmailAddress).First(&emailAccount)

	if result.Error != nil {
		// 创建新账户
		emailAccount = models.EmailAccount{
			UserID:            userID,
			EmailAddress:      req.EmailAddress,
			PasswordEncrypted: encryptedPassword,
			Provider:          config.Name,
			IMAPServer:        config.IMAPServer,
			IMAPPort:          config.IMAPPort,
			Notes:             fmt.Sprintf("使用应用密码配置 - %s", time.Now().Format("2006-01-02 15:04:05")),
		}

		if err := database.DB.Create(&emailAccount).Error; err != nil {
			return fmt.Errorf("创建邮箱账户失败: %w", err)
		}

		log.Printf("[AppPasswordManager] Created new email account %d for %s with app password",
			emailAccount.ID, req.EmailAddress)
	} else {
		// 更新现有账户
		updates := map[string]interface{}{
			"password_encrypted": encryptedPassword,
			"provider":           config.Name,
			"imap_server":        config.IMAPServer,
			"imap_port":          config.IMAPPort,
			"notes":              fmt.Sprintf("已更新应用密码 - %s", time.Now().Format("2006-01-02 15:04:05")),
		}

		if err := database.DB.Model(&emailAccount).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新邮箱账户失败: %w", err)
		}

		log.Printf("[AppPasswordManager] Updated email account %d for %s with app password",
			emailAccount.ID, req.EmailAddress)
	}

	return nil
}

// TestAppPassword 测试应用密码连接
func (apm *AppPasswordManager) TestAppPassword(email, appPassword string, config ProviderConfig) AppPasswordTestResult {
	result := AppPasswordTestResult{
		EmailAddress:  email,
		Provider:      config.Name,
		TestTimestamp: time.Now().Format(time.RFC3339),
	}

	log.Printf("[AppPasswordManager] Testing app password for %s using %s", email, config.Name)

	// 连接到IMAP服务器
	serverAddr := fmt.Sprintf("%s:%d", config.IMAPServer, config.IMAPPort)

	var c *imapclient.Client
	var err error

	if config.SupportsTLS {
		// 使用TLS连接
		options := &imapclient.Options{
			TLSConfig: &tls.Config{ServerName: config.IMAPServer},
		}
		c, err = imapclient.DialTLS(serverAddr, options)
	} else {
		// 使用明文连接
		c, err = imapclient.DialInsecure(serverAddr, nil)
	}

	if err != nil {
		result.Message = fmt.Sprintf("无法连接到服务器: %v", err)
		return result
	}
	defer c.Close()

	// 尝试登录
	if err = c.Login(email, appPassword).Wait(); err != nil {
		result.Message = fmt.Sprintf("登录失败: %v", err)
		return result
	}

	// 获取服务器能力
	caps, err := c.Capability().Wait()
	if err != nil {
		log.Printf("[AppPasswordManager] Failed to get capabilities: %v", err)
	} else {
		for cap := range caps {
			result.Capabilities = append(result.Capabilities, string(cap))
		}
	}

	// 测试列出文件夹
	mailboxes, err := c.List("", "%", nil).Collect()
	if err != nil {
		log.Printf("[AppPasswordManager] Failed to list mailboxes: %v", err)
	} else {
		log.Printf("[AppPasswordManager] Successfully listed %d mailboxes", len(mailboxes))
	}

	result.Success = true
	result.Message = "应用密码连接测试成功"

	log.Printf("[AppPasswordManager] App password test successful for %s", email)
	return result
}

// MigrateFromOAuth2ToAppPassword 从OAuth2迁移到应用密码
func (apm *AppPasswordManager) MigrateFromOAuth2ToAppPassword(accountID uint, appPassword string) error {
	var emailAccount models.EmailAccount
	if err := database.DB.First(&emailAccount, accountID).Error; err != nil {
		return fmt.Errorf("邮箱账户不存在: %w", err)
	}

	// 检测服务商
	provider := apm.DetectProvider(emailAccount.EmailAddress)
	if provider == "" {
		return fmt.Errorf("不支持的邮箱服务商")
	}

	config, _ := apm.GetProviderConfig(provider)

	// 测试应用密码
	testResult := apm.TestAppPassword(emailAccount.EmailAddress, appPassword, config)
	if !testResult.Success {
		return fmt.Errorf("应用密码测试失败: %s", testResult.Message)
	}

	// 加密应用密码
	encryptedPassword, err := utils.Encrypt([]byte(appPassword))
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新账户配置
	updates := map[string]interface{}{
		"password_encrypted": encryptedPassword,
		"provider":           config.Name,
		"imap_server":        config.IMAPServer,
		"imap_port":          config.IMAPPort,
		"notes":              fmt.Sprintf("从OAuth2迁移到应用密码 - %s", time.Now().Format("2006-01-02 15:04:05")),
	}

	if err := database.DB.Model(&emailAccount).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新账户失败: %w", err)
	}

	// 删除相关的OAuth2 token记录
	database.DB.Where("email_account_id = ?", accountID).Delete(&models.UserOAuthToken{})

	log.Printf("[AppPasswordManager] Successfully migrated account %d from OAuth2 to app password", accountID)
	return nil
}

// IsUsingAppPassword 检查账户是否使用应用密码
func (apm *AppPasswordManager) IsUsingAppPassword(accountID uint) bool {
	var emailAccount models.EmailAccount
	if err := database.DB.First(&emailAccount, accountID).Error; err != nil {
		return false
	}

	// 检查是否有加密的密码且没有OAuth2 token
	hasPassword := emailAccount.PasswordEncrypted != ""

	var tokenCount int64
	database.DB.Model(&models.UserOAuthToken{}).Where("email_account_id = ?", accountID).Count(&tokenCount)
	hasOAuth2 := tokenCount > 0

	return hasPassword && !hasOAuth2
}

// GetAppPasswordGuide 获取应用密码设置指南
func (apm *AppPasswordManager) GetAppPasswordGuide(email string) map[string]interface{} {
	provider := apm.DetectProvider(email)
	if provider == "" {
		return map[string]interface{}{
			"supported": false,
			"message":   "暂不支持此邮箱服务商的应用密码",
		}
	}

	config, exists := apm.GetProviderConfig(provider)
	if !exists {
		return map[string]interface{}{
			"supported": false,
			"message":   "暂不支持此邮箱服务商的应用密码",
		}
	}

	return map[string]interface{}{
		"supported":     true,
		"provider":      config.Name,
		"guide":         config.AppPasswordGuide,
		"documentation": config.DocumentationURL,
		"imap_server":   config.IMAPServer,
		"imap_port":     config.IMAPPort,
		"benefits": []string{
			"永不过期：应用密码不会自动过期",
			"更稳定：避免OAuth2令牌刷新失败问题",
			"更简单：无需复杂的重新授权流程",
			"更安全：可以随时撤销和重新生成",
		},
		"steps": []string{
			"确保您的邮箱已启用两步验证",
			"前往邮箱服务商的应用密码设置页面",
			"生成一个专用于此应用的密码",
			"使用生成的应用密码而不是您的账户密码",
		},
	}
}
