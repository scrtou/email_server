package handlers

import (
	"github.com/gin-gonic/gin"
	"email_server/database"
	"email_server/models"
	"email_server/utils"
	"net/http"
	"time"
	"gorm.io/gorm"
)

// DashboardSummaryResponse 定义了仪表盘摘要API的响应结构
type DashboardSummaryResponse struct {
	ActiveSubscriptionsCount    int64                                  `json:"active_subscriptions_count"`
	EstimatedMonthlySpending    float64                                `json:"estimated_monthly_spending"`
	EstimatedYearlySpending     float64                                `json:"estimated_yearly_spending"`
	UpcomingRenewals            []models.ServiceSubscriptionResponse `json:"upcoming_renewals"` // 改为完整的Response
	SubscriptionsByPlatform     []PlatformSubscriptionCount            `json:"subscriptions_by_platform"`
	TotalEmailAccounts          int64                                  `json:"total_email_accounts"`
	TotalPlatforms              int64                                  `json:"total_platforms"`
	TotalPlatformRegistrations  int64                                  `json:"total_platform_registrations"`
	TotalServiceSubscriptions   int64                                  `json:"total_service_subscriptions"`
}

// PlatformSubscriptionCount 用于表示每个平台的订阅数量
type PlatformSubscriptionCount struct {
	PlatformName      string `json:"platform_name"`
	SubscriptionCount int64  `json:"subscription_count"`
}

// GetDashboardSummary 处理获取仪表盘摘要数据的请求
func GetDashboardSummary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}
	userIDRaw, ok := userID.(int64)
	if !ok {
		// Try float64 as JWT numbers can sometimes be parsed as float64
		userIDFloat, okFloat := userID.(float64)
		if !okFloat {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "无法将 user_id 转换为期望的数值类型")
			return
		}
		userIDRaw = int64(userIDFloat)
	}
	currentUserID := uint(userIDRaw)

	var summary DashboardSummaryResponse
	var err error

	// 1. 活跃订阅数
	err = database.DB.Model(&models.ServiceSubscription{}).
		Where("user_id = ? AND status = ?", currentUserID, "active").
		Count(&summary.ActiveSubscriptionsCount).Error
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取活跃订阅数失败: "+err.Error())
		return
	}

	// 获取所有活跃订阅用于计算支出和即将到期
	var activeSubscriptions []models.ServiceSubscription
	err = database.DB.Model(&models.ServiceSubscription{}).
		Where("user_id = ? AND status = ?", currentUserID, "active").
		Find(&activeSubscriptions).Error
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取活跃订阅列表失败: "+err.Error())
		return
	}

	// 2. 预估月度/年度总支出
	var monthlySpending float64
	var yearlySpending float64
	for _, sub := range activeSubscriptions {
		cost := sub.Cost
		switch sub.BillingCycle {
		case "monthly":
			monthlySpending += cost
			yearlySpending += cost * 12
		case "yearly":
			monthlySpending += cost / 12
			yearlySpending += cost
		case "onetime":
			// 一次性费用不计入周期性支出，或根据业务逻辑调整
		case "free":
			// 免费订阅不计入支出
		}
	}
	summary.EstimatedMonthlySpending = monthlySpending
	summary.EstimatedYearlySpending = yearlySpending

	// 3. 即将到期订阅列表 (未来30天内)
	thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)
	var upcomingRenewalsModels []models.ServiceSubscription
	err = database.DB.Model(&models.ServiceSubscription{}).
		Preload("PlatformRegistration.Platform"). // 预加载平台信息以获取平台名称
		Preload("PlatformRegistration.EmailAccount"). // 预加载邮箱信息
		Where("user_id = ? AND status = ? AND next_renewal_date IS NOT NULL AND next_renewal_date BETWEEN ? AND ?",
			currentUserID, "active", time.Now(), thirtyDaysFromNow).
		Order("next_renewal_date ASC").
		Find(&upcomingRenewalsModels).Error
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取即将到期订阅失败: "+err.Error())
		return
	}
	for _, sub := range upcomingRenewalsModels {
		// 直接使用完整的 Response，因为 UpcomingRenewals 的类型已经是 []models.ServiceSubscriptionResponse
		emailAccountForResp := models.EmailAccount{}
		if sub.PlatformRegistration.EmailAccount != nil {
			emailAccountForResp = *sub.PlatformRegistration.EmailAccount
		}
		summary.UpcomingRenewals = append(summary.UpcomingRenewals, sub.ToServiceSubscriptionResponse(sub.PlatformRegistration, sub.PlatformRegistration.Platform, emailAccountForResp))
	}


	// 4. (可选) 各平台订阅数量分布
	type platformSubscriptionResult struct {
		PlatformName      string
		SubscriptionCount int64
	}
	var platformCounts []platformSubscriptionResult
	err = database.DB.Model(&models.ServiceSubscription{}).
		Joins("JOIN platform_registrations ON platform_registrations.id = service_subscriptions.platform_registration_id AND platform_registrations.user_id = ?", currentUserID). //确保 platform_registrations 也是当前用户的
		Joins("JOIN platforms ON platforms.id = platform_registrations.platform_id"). // platforms 本身可能是共享的，也可能按用户隔离，这里假设平台名是关键
		Where("service_subscriptions.user_id = ? AND service_subscriptions.status = ?", currentUserID, "active").
		Group("platforms.name").
		Select("platforms.name as platform_name, count(service_subscriptions.id) as subscription_count").
		Scan(&platformCounts).Error

	if err != nil && err != gorm.ErrRecordNotFound { // 忽略记录未找到的错误，因为可能没有任何订阅
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取各平台订阅数量失败: "+err.Error())
		return
	}
	for _, pc := range platformCounts {
		summary.SubscriptionsByPlatform = append(summary.SubscriptionsByPlatform, PlatformSubscriptionCount{
			PlatformName:      pc.PlatformName,
			SubscriptionCount: pc.SubscriptionCount,
		})
	}
	
	// 5. 补充其他统计数据
	err = database.DB.Model(&models.EmailAccount{}).Where("user_id = ?", currentUserID).Count(&summary.TotalEmailAccounts).Error
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取邮箱账户总数失败: "+err.Error())
		return
	}
	// Platform 是共享的，但如果只想统计用户创建的，需要调整 Platform 模型或查询逻辑
	// 当前 Platform 模型有 UserID，所以可以统计用户创建的
	err = database.DB.Model(&models.Platform{}).Where("user_id = ?", currentUserID).Count(&summary.TotalPlatforms).Error
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台总数失败: "+err.Error())
		return
	}
	err = database.DB.Model(&models.PlatformRegistration{}).Where("user_id = ?", currentUserID).Count(&summary.TotalPlatformRegistrations).Error
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台注册总数失败: "+err.Error())
		return
	}
	err = database.DB.Model(&models.ServiceSubscription{}).Where("user_id = ?", currentUserID).Count(&summary.TotalServiceSubscriptions).Error
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取服务订阅总数失败: "+err.Error())
		return
	}


	utils.SendSuccessResponse(c, summary)
}


// GetDashboard (保留旧的，或根据需要移除/重构)
func GetDashboard(c *gin.Context) {
	// ... (旧的 GetDashboard 实现可以保留，或者将其功能合并到 GetDashboardSummary，或者移除)
	// 为了清晰，这里暂时注释掉旧的实现，新的API是 GetDashboardSummary
	/*
	dashboard := &models.DashboardData{
		PlatformsByCategory:     make(map[string]int),
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

	recentEmailAccounts, err := GetRecentEmailAccountsWithCounts(c)
	if err != nil {
		return
	}
	dashboard.RecentEmailAccounts = recentEmailAccounts

	recentPlatforms, err := GetRecentPlatformsWithCounts(c)
	if err != nil {
		return
	}
	dashboard.RecentPlatforms = recentPlatforms

	utils.SendSuccessResponse(c, dashboard)
	*/
	utils.SendErrorResponse(c, http.StatusNotImplemented, "此 /dashboard 端点已弃用，请使用 /dashboard/summary")
}

// GetRecentEmailAccountsWithCounts - Fetches recent EmailAccount entries with platform counts.
// (此函数及 GetRecentPlatformsWithCounts 如果不再被新的 GetDashboardSummary 使用，可以考虑移除或重构)
func GetRecentEmailAccountsWithCounts(c *gin.Context) ([]models.EmailAccountResponse, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return nil, gorm.ErrRecordNotFound // 或者其他合适的错误
	}
	userIDRaw, ok := userID.(int64)
	if !ok {
		userIDFloat, okFloat := userID.(float64)
		if !okFloat {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "无法将 user_id 转换为期望的数值类型 (recent emails)")
			return nil, gorm.ErrInvalidData
		}
		userIDRaw = int64(userIDFloat)
	}
	currentUserID := uint(userIDRaw)

	var recentEmailAccounts []*models.EmailAccount
	
	dbResult := database.DB.Model(&models.EmailAccount{}).
		Where("user_id = ?", currentUserID).
		Order("created_at DESC").
		Limit(5).
		Find(&recentEmailAccounts)

	if dbResult.Error != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取最近邮箱账户失败: "+dbResult.Error.Error())
		return nil, dbResult.Error
	}

	var response []models.EmailAccountResponse
	for _, ea := range recentEmailAccounts {
		var platformCount int64
		countResult := database.DB.Model(&models.PlatformRegistration{}).
			Where("email_account_id = ? AND user_id = ?", ea.ID, currentUserID).
			Count(&platformCount)
		if countResult.Error != nil {
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
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return nil, gorm.ErrRecordNotFound // 或者其他合适的错误
	}
	userIDRaw, ok := userID.(int64)
	if !ok {
		userIDFloat, okFloat := userID.(float64)
		if !okFloat {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "无法将 user_id 转换为期望的数值类型 (recent platforms)")
			return nil, gorm.ErrInvalidData
		}
		userIDRaw = int64(userIDFloat)
	}
	currentUserID := uint(userIDRaw)
	var recentPlatforms []*models.Platform

	// 假设 Platform 是用户隔离的，或者有一个 user_id 字段
	// 根据 Platform 模型，它确实有 UserID
	dbResult := database.DB.Model(&models.Platform{}).
		Where("user_id = ?", currentUserID).
		Order("created_at DESC").
		Limit(5).
		Find(&recentPlatforms)

	if dbResult.Error != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取最近平台失败: "+dbResult.Error.Error())
		return nil, dbResult.Error
	}

	var response []models.PlatformResponse
	for _, p := range recentPlatforms {
		var emailAccountCount int64
		// 确保这里的查询也是用户隔离的
		countResult := database.DB.Model(&models.PlatformRegistration{}).
			Where("platform_id = ? AND user_id = ?", p.ID, currentUserID).
			Count(&emailAccountCount)
		if countResult.Error != nil {
			emailAccountCount = 0
		}

		pResponse := p.ToPlatformResponse()
		pResponse.EmailAccountCount = emailAccountCount
		response = append(response, pResponse)
	}
	return response, nil
}
