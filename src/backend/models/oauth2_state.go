// models/oauth2_state.go
package models

import (
	"time"
)

// OAuth2State 用于存储临时的OAuth2流程状态，持久化到数据库
type OAuth2State struct {
	State        string `gorm:"primaryKey"`
	AccountID    uint   // 用于关联邮箱账户，对LinuxDo流程可以为0
	PKCEVerifier string `gorm:"type:text"` // PKCE验证器，对LinuxDo流程可以为空
	ExpiresAt    time.Time
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

// 在数据库初始化时，请确保运行了迁移
// database.DB.AutoMigrate(&models.OAuth2State{})
