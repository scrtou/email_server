package models

import "time"

// OAuthProvider stores configuration for each supported OAuth2 provider.
type OAuthProvider struct {
	ID                   uint      `gorm:"primaryKey;autoIncrement"`
	Name                 string    `gorm:"type:varchar(50);not null;unique"` // e.g., 'google', 'microsoft'
	ClientID             string    `gorm:"type:varchar(255);not null"`
	ClientSecretEncrypted string    `gorm:"type:varchar(512);not null"`      // Encrypted client secret
	AuthURL              string    `gorm:"type:varchar(255);not null"`
	TokenURL             string    `gorm:"type:varchar(255);not null"`
	Scopes               string    `gorm:"type:text;not null"` // Comma-separated list of scopes
	IMAPServer           string    `gorm:"type:varchar(255);not null;default:''"`
	IMAPPort             int       `gorm:"not null;default:0"`
	CreatedAt            time.Time `gorm:"autoCreateTime"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime"`
}