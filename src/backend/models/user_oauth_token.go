package models

import "time"

// UserOAuthToken stores the access and refresh tokens for each user and connected account.
type UserOAuthToken struct {
	ID                   uint      `gorm:"primaryKey;autoIncrement"`
	UserID               uint      `gorm:"not null"`
	EmailAccountID       uint      `gorm:"not null"`
	ProviderID           uint      `gorm:"not null"`
	AccessTokenEncrypted  string    `gorm:"type:varchar(2048);not null"` // Encrypted access token
	RefreshTokenEncrypted string    `gorm:"type:varchar(2048)"`          // Encrypted refresh token (can be null)
	TokenType            string    `gorm:"type:varchar(50);default:'Bearer'"`
	Expiry               time.Time `gorm:"not null"`                      // Expiry date/time of the access token
	CreatedAt            time.Time `gorm:"autoCreateTime"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime"`

	User         User          `gorm:"foreignKey:UserID"`
	EmailAccount EmailAccount  `gorm:"foreignKey:EmailAccountID"`
	Provider     OAuthProvider `gorm:"foreignKey:ProviderID"`
}