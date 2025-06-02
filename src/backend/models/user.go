package models

import (
	"time"
	"gorm.io/gorm"
)

// User model based on optimization_proposal.md
type User struct {
	gorm.Model        // Embeds ID, CreatedAt, UpdatedAt, DeletedAt
	Username   string `json:"username" gorm:"unique;not null"`
	Email      string `json:"email" gorm:"unique;not null"`
	Password   string `json:"-" gorm:""` // Password hash, json:"-" to omit from JSON responses by default, nullable for OAuth users
	// OAuth2 fields
	LinuxDoID  *int64  `json:"-" gorm:"unique"` // LinuxDo user ID, nullable
	Provider   *string `json:"provider,omitempty"` // OAuth provider (e.g., "linuxdo")
	// RealName  string     `json:"real_name"` // Not in core model per optimization_proposal.md
	// Phone     string     `json:"phone"`     // Not in core model
	// Role      string     `json:"role"`      // Not in core model
	// Status    int        `json:"status"`    // Not in core model
	// LastLogin *time.Time `json:"last_login"`// Not in core model
}

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

// 转换为响应格式（隐藏密码）
func (u *User) ToResponse() *UserResponse {
    return &UserResponse{
        ID:        u.ID,
        Username:  u.Username,
        Email:     u.Email,
        Provider:  u.Provider,
        CreatedAt: u.CreatedAt,
        UpdatedAt: u.UpdatedAt,
    }
}
