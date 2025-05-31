package models

import (
	"time"

	"gorm.io/gorm"
	// "github.com/shopspring/decimal" // Consider for precise decimal arithmetic if needed
)

// ServiceSubscription 定义了用户在特定平台注册下的服务订阅详情
type ServiceSubscription struct {
	gorm.Model
	UserID                 uint   `gorm:"not null;index"`                                           // 外键，关联到 User 模型
	PlatformRegistrationID uint   `gorm:"not null;index;constraint:OnDelete:CASCADE"` // 外键，关联到 PlatformRegistration 模型
	ServiceName            string `gorm:"type:varchar(255);not null"`
	Description            string `gorm:"type:text"`
	Status                 string `gorm:"type:varchar(50)"` // e.g., active, cancelled, free_trial, expired
	Cost                   float64                         // 费用金额. For precision, consider decimal type or storing as integer (e.g., cents)
	// Cost                   decimal.Decimal `gorm:"type:decimal(10,2);"` // Example with shopspring/decimal
	BillingCycle      string     `gorm:"type:varchar(50)"` // e.g., monthly, yearly, onetime, free
	NextRenewalDate   *time.Time `gorm:"type:date"`        // 下次续费日期 (可空)
	PaymentMethodNotes string     `gorm:"type:text"`        // 支付方式备注
	IsRead                 bool      `gorm:"default:false"`     // 新增字段，标记是否已读

	User                 User                 `gorm:"foreignKey:UserID"`
	PlatformRegistration PlatformRegistration `gorm:"foreignKey:PlatformRegistrationID"`
}

// ServiceSubscriptionResponse 用于API响应
type ServiceSubscriptionResponse struct {
	ID                     uint    `json:"id"`
	UserID                 uint    `json:"user_id"`
	PlatformRegistrationID uint    `json:"platform_registration_id"`
	// Fields from PlatformRegistration for context
	PlatformName           string  `json:"platform_name,omitempty"` 
	EmailAddress           string  `json:"email_address,omitempty"`
	LoginUsername          string  `json:"login_username,omitempty"`
	// ServiceSubscription specific fields
	ServiceName            string  `json:"service_name"`
	Description            string  `json:"description"`
	Status                 string  `json:"status"`
	Cost                   float64 `json:"cost"`
	BillingCycle           string  `json:"billing_cycle"`
	NextRenewalDate        *string `json:"next_renewal_date"` // Pointer to string to handle null
	PaymentMethodNotes     string  `json:"payment_method_notes"`
	IsRead                 bool    `json:"is_read"` // 新增字段
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

// ToServiceSubscriptionResponse 将 ServiceSubscription 模型转换为 ServiceSubscriptionResponse
// Needs PlatformRegistration's associated Platform and EmailAccount for full context
func (ss *ServiceSubscription) ToServiceSubscriptionResponse(pr PlatformRegistration, p Platform, ea EmailAccount) ServiceSubscriptionResponse {
	var renewalDateStr *string
	if ss.NextRenewalDate != nil {
		s := ss.NextRenewalDate.Format("2006-01-02")
		renewalDateStr = &s
	}

	return ServiceSubscriptionResponse{
		ID:                     ss.ID,
		UserID:                 ss.UserID,
		PlatformRegistrationID: ss.PlatformRegistrationID,
		PlatformName:           p.Name,
		EmailAddress:           ea.EmailAddress,
		LoginUsername:          pr.LoginUsername,
		ServiceName:            ss.ServiceName,
		Description:            ss.Description,
		Status:                 ss.Status,
		Cost:                   ss.Cost,
		BillingCycle:           ss.BillingCycle,
		NextRenewalDate:        renewalDateStr,
		PaymentMethodNotes:     ss.PaymentMethodNotes,
		IsRead:                 ss.IsRead,
		CreatedAt:              ss.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:              ss.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ToServiceSubscriptionResponseLite is a simplified version for lists where full context might be loaded separately
func (ss *ServiceSubscription) ToServiceSubscriptionResponseLite() ServiceSubscriptionResponse {
    var renewalDateStr *string
    if ss.NextRenewalDate != nil {
        s := ss.NextRenewalDate.Format("2006-01-02")
        renewalDateStr = &s
    }
    return ServiceSubscriptionResponse{
        ID:                     ss.ID,
        UserID:                 ss.UserID,
        PlatformRegistrationID: ss.PlatformRegistrationID,
        ServiceName:            ss.ServiceName,
        Description:            ss.Description,
        Status:                 ss.Status,
        Cost:                   ss.Cost,
        BillingCycle:           ss.BillingCycle,
        NextRenewalDate:        renewalDateStr,
        PaymentMethodNotes:     ss.PaymentMethodNotes,
        IsRead:                 ss.IsRead,
        CreatedAt:              ss.CreatedAt.Format("2006-01-02 15:04:05"),
        UpdatedAt:              ss.UpdatedAt.Format("2006-01-02 15:04:05"),
       }
      }