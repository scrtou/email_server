package handlers

import (
	"github.com/gin-gonic/gin"
	"email_server/database"
	"email_server/models"
	"email_server/utils"
)

func GetDashboard(c *gin.Context) {
	dashboard := &models.DashboardData{
		PlatformsByCategory:     make(map[string]int), // Initialized, but logic to fill it is removed as Platform has no Category
	}
	var err error

	// Get EmailAccount Count
	err = database.DB.Model(&models.EmailAccount{}).Count(&dashboard.EmailAccountCount).Error
	if err != nil {
		utils.SendErrorResponse(c, 500, "获取邮箱账户数量失败: "+err.Error())
		return
	}

	// Get Platform Count
	err = database.DB.Model(&models.Platform{}).Count(&dashboard.PlatformCount).Error
	if err != nil {
		utils.SendErrorResponse(c, 500, "获取平台数量失败: "+err.Error())
		return
	}

	// Get Relation Count (e.g., PlatformRegistrations)
	err = database.DB.Model(&models.PlatformRegistration{}).Count(&dashboard.RelationCount).Error
	if err != nil {
		utils.SendErrorResponse(c, 500, "获取关联数量 (PlatformRegistrations) 失败: "+err.Error())
		return
	}

	// PlatformsByCategory is initialized in DashboardData but not populated here as Platform has no Category.

	recentEmailAccounts, err := GetRecentEmailAccountsWithCounts(c)
	if err != nil {
		// Error already handled and response sent in GetRecentEmailAccountsWithCounts
		return
	}
	dashboard.RecentEmailAccounts = recentEmailAccounts

	recentPlatforms, err := GetRecentPlatformsWithCounts(c)
	if err != nil {
		// Error already handled and response sent in GetRecentPlatformsWithCounts
		return
	}
	dashboard.RecentPlatforms = recentPlatforms

	utils.SendSuccessResponse(c, dashboard)
}

// GetRecentEmailAccountsWithCounts - Fetches recent EmailAccount entries with platform counts.
func GetRecentEmailAccountsWithCounts(c *gin.Context) ([]models.EmailAccountResponse, error) {
	var recentEmailAccounts []*models.EmailAccount
	
	dbResult := database.DB.Model(&models.EmailAccount{}).
		Order("created_at DESC").
		Limit(5).
		Find(&recentEmailAccounts)

	if dbResult.Error != nil {
		utils.SendErrorResponse(c, 500, "获取最近邮箱账户失败: "+dbResult.Error.Error())
		return nil, dbResult.Error
	}

	var response []models.EmailAccountResponse
	for _, ea := range recentEmailAccounts {
		var platformCount int64
		// Count related PlatformRegistrations
		// Assuming PlatformRegistration has an EmailAccountID field
		countResult := database.DB.Model(&models.PlatformRegistration{}).Where("email_account_id = ?", ea.ID).Count(&platformCount)
		if countResult.Error != nil {
			// Log error or handle. For simplicity, we'll set count to 0 if query fails.
			platformCount = 0
		}
		
		eaResponse := ea.ToEmailAccountResponse()
		eaResponse.PlatformCount = platformCount
		response = append(response, eaResponse)
	}
	return response, nil
}

// GetRecentPlatformsWithCounts - Fetches recent Platform entries with email account counts.
func GetRecentPlatformsWithCounts(c *gin.Context) ([]models.PlatformResponse, error) {
	var recentPlatforms []*models.Platform

	dbResult := database.DB.Model(&models.Platform{}).
		Order("created_at DESC").
		Limit(5).
		Find(&recentPlatforms)

	if dbResult.Error != nil {
		utils.SendErrorResponse(c, 500, "获取最近平台失败: "+dbResult.Error.Error())
		return nil, dbResult.Error
	}

	var response []models.PlatformResponse
	for _, p := range recentPlatforms {
		var emailAccountCount int64
		// Count related PlatformRegistrations
		// Assuming PlatformRegistration has a PlatformID field
		countResult := database.DB.Model(&models.PlatformRegistration{}).Where("platform_id = ?", p.ID).Count(&emailAccountCount)
		if countResult.Error != nil {
			// Log error or handle. For simplicity, we'll set count to 0 if query fails.
			emailAccountCount = 0
		}

		pResponse := p.ToPlatformResponse()
		pResponse.EmailAccountCount = emailAccountCount
		response = append(response, pResponse)
	}
	return response, nil
}
