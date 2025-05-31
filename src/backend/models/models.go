package models

import (
	//"time"
)

// Email, Service, and EmailService structs have been removed as they are replaced by
// EmailAccount, Platform, PlatformRegistration, and ServiceSubscription respectively.
// Ensure that EmailAccount and Platform types are available in this package (e.g., defined in other .go files within this package).

type DashboardData struct {
	EmailAccountCount       int64             `json:"email_account_count"`      // Renamed from EmailCount, reflects EmailAccount model
	PlatformCount           int64             `json:"platform_count"`           // Renamed from ServiceCount, reflects Platform model
	RelationCount           int64             `json:"relation_count"`           // Represents general relations, may need specific review based on new models
	PlatformsByCategory     map[string]int    `json:"platforms_by_category"`      // Renamed from ServicesByCategory
	RecentEmailAccounts     []EmailAccountResponse `json:"recent_email_accounts"`    // Changed from RecentEmails to use EmailAccountResponse
	RecentPlatforms         []PlatformResponse     `json:"recent_platforms"`         // Changed from RecentServices to use PlatformResponse
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}
// GlobalSearchResult defines the structure for items in global search results.
type GlobalSearchResultItem struct {
	ID          uint        `json:"id"`
	Type        string      `json:"type"`         // e.g., "user", "email_account", "platform", "platform_registration", "service_subscription"
	DisplayName string      `json:"display_name"` // A user-friendly name for the item
	Details     interface{} `json:"details,omitempty"` // Additional details specific to the item type
}

// GlobalSearchResponse defines the structure for the global search API response.
type GlobalSearchResponse struct {
	Results []GlobalSearchResultItem `json:"results"`
}
