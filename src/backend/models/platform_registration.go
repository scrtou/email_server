package models

import "gorm.io/gorm"

// PlatformRegistration 定义了用户邮箱在特定平台上的注册信息
type PlatformRegistration struct {
	gorm.Model
	UserID                 uint   `gorm:"not null;index"`                                           // 外键，关联到 User 模型
	EmailAccountID         uint   `gorm:"not null;index;uniqueIndex:idx_email_platform;constraint:OnDelete:CASCADE"` // 外键，关联到 EmailAccount 模型
	PlatformID             uint   `gorm:"not null;index;uniqueIndex:idx_email_platform;constraint:OnDelete:CASCADE"` // 外键，关联到 Platform 模型
	LoginUsername          string `gorm:"type:varchar(255)"`                                      // 在该平台的登录用户名/ID
	LoginPasswordEncrypted string `gorm:"type:varchar(255)"`          // 在该平台的登录密码 (加密存储)
	Notes                  string `gorm:"type:text"`                  // 备注信息

	User         User         `gorm:"foreignKey:UserID"`
	EmailAccount EmailAccount `gorm:"foreignKey:EmailAccountID"`
	Platform     Platform     `gorm:"foreignKey:PlatformID"`
}

// PlatformRegistrationResponse 用于API响应，可能包含关联模型的摘要信息
type PlatformRegistrationResponse struct {
	ID                     uint                `json:"id"`
	UserID                 uint                `json:"user_id"`
	EmailAccountID         uint                `json:"email_account_id"`
	EmailAddress           string              `json:"email_address"` // From EmailAccount
	PlatformID             uint                `json:"platform_id"`
	PlatformName           string              `json:"platform_name"` // From Platform
	LoginUsername          string              `json:"login_username"`
	Notes                  string              `json:"notes"`
	CreatedAt              string              `json:"created_at"`
	UpdatedAt              string              `json:"updated_at"`
}

// ToPlatformRegistrationResponse 将 PlatformRegistration 模型转换为 PlatformRegistrationResponse
// 需要传入关联的 EmailAccount 和 Platform 以填充响应信息
func (pr *PlatformRegistration) ToPlatformRegistrationResponse(emailAccount EmailAccount, platform Platform) PlatformRegistrationResponse {
	return PlatformRegistrationResponse{
		ID:             pr.ID,
		UserID:         pr.UserID,
		EmailAccountID: pr.EmailAccountID,
		EmailAddress:   emailAccount.EmailAddress, // Populate from passed EmailAccount
		PlatformID:     pr.PlatformID,
		PlatformName:   platform.Name, // Populate from passed Platform
		LoginUsername:  pr.LoginUsername,
		Notes:          pr.Notes,
		CreatedAt:      pr.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      pr.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ToPlatformRegistrationResponseLite 是一个简化的转换，不立即加载关联对象，用于列表等场景
// 在实际使用中，handler层获取到PlatformRegistration后，需要单独查询EmailAccount和Platform来填充
func (pr *PlatformRegistration) ToPlatformRegistrationResponseLite() PlatformRegistrationResponse {
    return PlatformRegistrationResponse{
        ID:             pr.ID,
        UserID:         pr.UserID,
        EmailAccountID: pr.EmailAccountID,
        PlatformID:     pr.PlatformID,
        LoginUsername:  pr.LoginUsername,
        Notes:          pr.Notes,
        CreatedAt:      pr.CreatedAt.Format("2006-01-02 15:04:05"),
        UpdatedAt:      pr.UpdatedAt.Format("2006-01-02 15:04:05"),
    }
}