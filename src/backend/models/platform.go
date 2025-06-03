package models

import "gorm.io/gorm"

// Platform 定义了注册平台的数据模型
type Platform struct {
	gorm.Model
	UserID     uint   `gorm:"not null"`                   // 外键，关联到 User 模型
	Name       string `gorm:"type:varchar(255);not null"` // 平台名称, 用户ID和平台名称组合唯一（通过自定义索引实现）
	WebsiteURL string `gorm:"type:varchar(255)"`          // 平台官方网址
	Notes      string `gorm:"type:text"`                  // 备注信息

	User User `gorm:"foreignKey:UserID"` // 定义关联关系
}

// PlatformResponse 用于API响应
type PlatformResponse struct {
	ID                uint   `json:"id"`
	UserID            uint   `json:"user_id"` // 添加 UserID
	Name              string `json:"name"`
	WebsiteURL        string `json:"website_url"`
	Notes             string `json:"notes"`
	EmailAccountCount int64  `json:"email_account_count"` // 添加关联邮箱数量字段
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

// ToPlatformResponse 将 Platform 模型转换为 PlatformResponse
// 注意：EmailAccountCount 需要在调用此方法前被填充
func (p *Platform) ToPlatformResponse() PlatformResponse {
	return PlatformResponse{
		ID:         p.ID,
		UserID:     p.UserID, // 添加 UserID
		Name:       p.Name,
		WebsiteURL: p.WebsiteURL,
		Notes:      p.Notes,
		// EmailAccountCount: 0, // 这里暂时不直接赋值，由 handler 处理
		CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
