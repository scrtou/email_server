package models

import (
    "time"
)

type Email struct {
    ID           int64     `json:"id" db:"id"`
    Email        string    `json:"email" db:"email"`
    Password     string    `json:"password,omitempty" db:"password"`
    DisplayName  string    `json:"display_name" db:"display_name"`
    Provider     string    `json:"provider" db:"provider"`
    Phone        string    `json:"phone" db:"phone"`
    BackupEmail  string    `json:"backup_email" db:"backup_email"`
    SecurityQ    string    `json:"security_question" db:"security_question"`
    SecurityA    string    `json:"security_answer,omitempty" db:"security_answer"`
    Notes        string    `json:"notes" db:"notes"`
    Status       int       `json:"status" db:"status"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
    ServiceCount int       `json:"service_count,omitempty"`
}

type Service struct {
    ID          int64     `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    Website     string    `json:"website" db:"website"`
    Category    string    `json:"category" db:"category"`
    Description string    `json:"description" db:"description"`
    LogoURL     string    `json:"logo_url" db:"logo_url"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
    EmailCount  int       `json:"email_count,omitempty"`
}

type EmailService struct {
    ID                  int64      `json:"id" db:"id"`
    EmailID             int64      `json:"email_id" db:"email_id"`
    ServiceID           int64      `json:"service_id" db:"service_id"`
    Username            string     `json:"username" db:"username"`
    Password            string     `json:"password,omitempty" db:"password"`
    Phone               string     `json:"phone" db:"phone"`
    RegistrationDate    *time.Time `json:"registration_date" db:"registration_date"`
    SubscriptionType    string     `json:"subscription_type" db:"subscription_type"`
    SubscriptionExpires *time.Time `json:"subscription_expires" db:"subscription_expires"`
    Notes               string     `json:"notes" db:"notes"`
    Status              int        `json:"status" db:"status"`
    CreatedAt           time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
    Email               *Email     `json:"email,omitempty"`
    Service             *Service   `json:"service,omitempty"`
    ServiceName         string     `json:"service_name,omitempty"`
    EmailAddr           string     `json:"email_addr,omitempty"`

    // 关联字段
    EmailDisplayName      string    `json:"email_display_name,omitempty"`
    ServiceWebsite        string    `json:"service_website,omitempty"`
}

type DashboardData struct {
    EmailCount         int            `json:"email_count"`
    ServiceCount       int            `json:"service_count"`
    RelationCount      int            `json:"relation_count"`
    EmailsByProvider   map[string]int `json:"emails_by_provider"`
    ServicesByCategory map[string]int `json:"services_by_category"`
    RecentEmails       []*Email       `json:"recent_emails"`
    RecentServices     []*Service     `json:"recent_services"`
}

type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
