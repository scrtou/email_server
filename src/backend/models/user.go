package models

import (
    "time"
)

type User struct {
    ID        int64      `json:"id" db:"id"`
    Username  string     `json:"username" db:"username"`
    Email     string     `json:"email" db:"email"`
    Password  string     `json:"password,omitempty" db:"password"`
    RealName  string     `json:"real_name" db:"real_name"`
    Phone     string     `json:"phone" db:"phone"`
    Role      string     `json:"role" db:"role"`
    Status    int        `json:"status" db:"status"`
    LastLogin *time.Time `json:"last_login" db:"last_login"`
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=20"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    RealName string `json:"real_name" binding:"required"`
    Phone    string `json:"phone"`
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
    ID        int64      `json:"id"`
    Username  string     `json:"username"`
    Email     string     `json:"email"`
    RealName  string     `json:"real_name"`
    Phone     string     `json:"phone"`
    Role      string     `json:"role"`
    Status    int        `json:"status"`
    LastLogin *time.Time `json:"last_login"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}

// 转换为响应格式（隐藏密码）
func (u *User) ToResponse() *UserResponse {
    return &UserResponse{
        ID:        u.ID,
        Username:  u.Username,
        Email:     u.Email,
        RealName:  u.RealName,
        Phone:     u.Phone,
        Role:      u.Role,
        Status:    u.Status,
        LastLogin: u.LastLogin,
        CreatedAt: u.CreatedAt,
        UpdatedAt: u.UpdatedAt,
    }
}
