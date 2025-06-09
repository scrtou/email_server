package models

import (
	"time"

	"gorm.io/gorm"
)

// User model based on optimization_proposal.md
type User struct {
	gorm.Model        // Embeds ID, CreatedAt, UpdatedAt, DeletedAt
	Username   string `json:"username" gorm:"uniqueIndex:uq_username,priority:1;not null"`
	Email      string `json:"email" gorm:"uniqueIndex:uq_email,priority:1;not null"`
	Password   string `json:"-" gorm:""` // Password hash, json:"-" to omit from JSON responses by default, nullable for OAuth users
	// OAuth2 fields
	LinuxDoID   *int64  `json:"-" gorm:"uniqueIndex:uq_linuxdo_id,priority:1"`   // LinuxDo user ID, nullable
	GoogleID    *string `json:"-" gorm:"uniqueIndex:uq_google_id,priority:1"`    // Google user ID, nullable
	MicrosoftID *string `json:"-" gorm:"uniqueIndex:uq_microsoft_id,priority:1"` // Microsoft user ID, nullable
	Provider    *string `json:"provider,omitempty"`                              // OAuth provider (e.g., "linuxdo", "google", "microsoft")
	// Extended fields
	Role      string     `json:"role" gorm:"default:user"` // 用户角色: admin=管理员, user=普通用户
	Status    int        `json:"status" gorm:"default:1"`  // 用户状态: 1=激活, 0=封禁
	LastLogin *time.Time `json:"last_login"`               // 最后登录时间
}

// 用户角色常量
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// 用户状态常量
const (
	StatusBanned = 0 // 封禁
	StatusActive = 1 // 激活
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	User      *User  `json:"user"`
}

type UserResponse struct {
	ID        uint       `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Provider  *string    `json:"provider,omitempty"`
	Role      string     `json:"role"`
	Status    int        `json:"status"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// OAuth2 related structs
type OAuth2AuthRequest struct {
	Provider string `json:"provider" binding:"required"`
}

type OAuth2CallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

type LinuxDoUserInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar_url"`
}

// MicrosoftUserInfo 定义Microsoft OAuth2用户信息结构
type MicrosoftUserInfo struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
	UserPrincipalName string `json:"userPrincipalName"`
	Mail              string `json:"mail"`
}

// 转换为响应格式（隐藏密码）
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Provider:  u.Provider,
		Role:      u.Role,
		Status:    u.Status,
		LastLogin: u.LastLogin,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// 检查用户是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// 检查用户状态是否激活
func (u *User) IsStatusActive() bool {
	return u.Status == StatusActive
}
