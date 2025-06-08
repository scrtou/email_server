package models

import "gorm.io/gorm"

// EmailAccount 定义了邮箱账户的数据模型
type EmailAccount struct {
	gorm.Model
	UserID            uint   `gorm:"not null;index;uniqueIndex:uq_user_email,priority:1"`             // 外键，关联到 User 模型
	EmailAddress      string `gorm:"type:varchar(255);not null;uniqueIndex:uq_user_email,priority:2"` // 与UserID组合唯一
	PasswordEncrypted string `gorm:"type:varchar(255)"`                                               // 加密存储的密码, 允许为空
	Provider          string `gorm:"type:varchar(100)"`                                               // 邮箱服务商，例如 Gmail, Outlook 等
	IMAPServer        string `gorm:"type:varchar(255)"`                                               // IMAP 服务器地址
	IMAPPort          int    `gorm:"type:int"`                                                        // IMAP 服务器端口
	Notes             string `gorm:"type:text"`                                                       // 备注信息
	PhoneNumber       string `gorm:"type:varchar(50)"`                                                // 手机号码, 可选

	User User `gorm:"foreignKey:UserID"` // 定义关联关系
}

// EmailAccountResponse 用于API响应，不包含敏感信息
type EmailAccountResponse struct {
	ID            uint   `json:"id"`
	EmailAddress  string `json:"email_address"`
	Provider      string `json:"provider"`
	IMAPServer    string `json:"imap_server"`
	IMAPPort      int    `json:"imap_port"`
	Notes         string `json:"notes"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	PlatformCount    int64  `json:"platform_count"` // 添加关联平台数量字段
	IsOAuthConnected bool   `json:"is_oauth_connected"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// ToEmailAccountResponse 将 EmailAccount 模型转换为 EmailAccountResponse
// 注意：PlatformCount 需要在调用此方法前被填充
func (ea *EmailAccount) ToEmailAccountResponse() EmailAccountResponse {
	return EmailAccountResponse{
		ID:           ea.ID,
		EmailAddress: ea.EmailAddress,
		Provider:     ea.Provider,
		IMAPServer:   ea.IMAPServer,
		IMAPPort:     ea.IMAPPort,
		Notes:        ea.Notes,
		PhoneNumber:  ea.PhoneNumber,
		// PlatformCount: 0, // 这里暂时不直接赋值，由 handler 处理
		CreatedAt: ea.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: ea.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
