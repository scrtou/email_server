package models

import "gorm.io/gorm"

// PlatformRegistration 定义了用户邮箱在特定平台上的注册信息
type PlatformRegistration struct {
	gorm.Model
	UserID                 uint    `gorm:"not null;uniqueIndex:uq_user_platform_loginusername,priority:1;uniqueIndex:uq_user_platform_emailaccountid,priority:1"`                             // 外键，关联到 User 模型
	EmailAccountID         *uint   `gorm:"uniqueIndex:uq_user_platform_emailaccountid,priority:3;constraint:OnDelete:CASCADE"`                                                                // 外键，关联到 EmailAccount 模型 (允许 NULL)
	PlatformID             uint    `gorm:"not null;uniqueIndex:uq_user_platform_loginusername,priority:2;uniqueIndex:uq_user_platform_emailaccountid,priority:2;constraint:OnDelete:CASCADE"` // 外键，关联到 Platform 模型
	LoginUsername          *string `gorm:"type:varchar(255);uniqueIndex:uq_user_platform_loginusername,priority:3"`                                                                           // 在该平台的登录用户名/ID (允许为空)
	LoginPasswordEncrypted string  `gorm:"type:varchar(255)"`                                                                                                                                 // 在该平台的登录密码 (加密存储)
	Notes                  string  `gorm:"type:text"`                                                                                                                                         // 备注信息
	PhoneNumber            string  `gorm:"type:varchar(50)"`                                                                                                                                  // 手机号码, 可选

	User         User          `gorm:"foreignKey:UserID"`
	EmailAccount *EmailAccount `gorm:"foreignKey:EmailAccountID"` // 指针类型以匹配可空外键
	Platform     Platform      `gorm:"foreignKey:PlatformID"`
}

// PlatformRegistrationResponse 用于API响应，可能包含关联模型的摘要信息
type PlatformRegistrationResponse struct {
	ID                 uint   `json:"id"`
	UserID             uint   `json:"user_id"`
	EmailAccountID     uint   `json:"email_account_id"`
	EmailAddress       string `json:"email_address"` // From EmailAccount
	PlatformID         uint   `json:"platform_id"`
	PlatformName       string `json:"platform_name"`        // From Platform
	PlatformWebsiteURL string `json:"platform_website_url"` // From Platform
	LoginUsername      string `json:"login_username"`
	Notes              string `json:"notes"`
	PhoneNumber        string `json:"phone_number,omitempty"`
	HasPassword        bool   `json:"has_password"` // 指示是否已设置密码
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

// ToPlatformRegistrationResponse 将 PlatformRegistration 模型转换为 PlatformRegistrationResponse
// 需要传入关联的 EmailAccount 和 Platform 以填充响应信息
func (pr *PlatformRegistration) ToPlatformRegistrationResponse(emailAccount EmailAccount, platform Platform) PlatformRegistrationResponse {
	return PlatformRegistrationResponse{
		ID:     pr.ID,
		UserID: pr.UserID,
		EmailAccountID: func(id *uint) uint {
			if id != nil {
				return *id
			}
			return 0
		}(pr.EmailAccountID),
		EmailAddress: func() string {
			// 只有当 EmailAccountID 不为 nil 时才返回邮箱地址
			if pr.EmailAccountID != nil {
				return emailAccount.EmailAddress
			}
			return "" // 没有关联邮箱时返回空字符串
		}(),
		PlatformID:         pr.PlatformID,
		PlatformName:       platform.Name,       // Populate from passed Platform
		PlatformWebsiteURL: platform.WebsiteURL, // Populate from passed Platform
		LoginUsername: func() string {
			if pr.LoginUsername != nil {
				return *pr.LoginUsername
			}
			return ""
		}(),
		Notes:       pr.Notes,
		PhoneNumber: pr.PhoneNumber,
		HasPassword: pr.LoginPasswordEncrypted != "", // 检查是否已设置密码
		CreatedAt:   pr.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   pr.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ToPlatformRegistrationResponseLite 是一个简化的转换，不立即加载关联对象，用于列表等场景
// 在实际使用中，handler层获取到PlatformRegistration后，需要单独查询EmailAccount和Platform来填充
func (pr *PlatformRegistration) ToPlatformRegistrationResponseLite() PlatformRegistrationResponse {
	return PlatformRegistrationResponse{
		ID:     pr.ID,
		UserID: pr.UserID,
		EmailAccountID: func(id *uint) uint {
			if id != nil {
				return *id
			}
			return 0
		}(pr.EmailAccountID),
		PlatformID: pr.PlatformID,
		LoginUsername: func() string {
			if pr.LoginUsername != nil {
				return *pr.LoginUsername
			}
			return ""
		}(),
		Notes:       pr.Notes,
		PhoneNumber: pr.PhoneNumber,
		HasPassword: pr.LoginPasswordEncrypted != "", // 检查是否已设置密码
		CreatedAt:   pr.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   pr.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
