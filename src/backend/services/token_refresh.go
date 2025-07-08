// Token refresh service - 令牌自动刷新服务
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"email_server/database"
	"email_server/integrations"
	"email_server/models"
)

// TokenRefreshService 令牌刷新服务
type TokenRefreshService struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewTokenRefreshService 创建新的令牌刷新服务
func NewTokenRefreshService() *TokenRefreshService {
	ctx, cancel := context.WithCancel(context.Background())
	return &TokenRefreshService{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动令牌刷新服务
func (s *TokenRefreshService) Start() {
	log.Println("[TokenRefreshService] Starting intelligent token refresh service...")
	
	// 启动时立即执行一次检查
	go s.refreshExpiredTokens()
	
	// 使用智能调度策略
	go s.intelligentScheduler()
}

// intelligentScheduler 智能调度器 - 根据令牌过期时间动态调整检查频率
func (s *TokenRefreshService) intelligentScheduler() {
	for {
		select {
		case <-s.ctx.Done():
			log.Println("[TokenRefreshService] Intelligent scheduler stopped")
			return
		default:
			// 计算下次检查时间
			nextCheckDuration := s.calculateNextCheckInterval()
			
			log.Printf("[TokenRefreshService] Next check scheduled in %v", nextCheckDuration)
			
			timer := time.NewTimer(nextCheckDuration)
			select {
			case <-s.ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				go s.refreshExpiredTokens()
			}
		}
	}
}

// calculateNextCheckInterval 计算下次检查间隔
func (s *TokenRefreshService) calculateNextCheckInterval() time.Duration {
	// 查询最近即将过期的令牌
	var earliestToken models.UserOAuthToken
	err := database.DB.Where("expiry > ?", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).
		Order("expiry ASC").
		First(&earliestToken).Error
	
	if err != nil {
		// 没有找到有效令牌，使用默认间隔
		log.Printf("[TokenRefreshService] No valid tokens found, using default interval")
		return 2 * time.Hour
	}
	
	now := time.Now()
	timeToExpiry := earliestToken.Expiry.Sub(now)
	
	// 智能调度策略
	var checkInterval time.Duration
	
	if timeToExpiry <= 30*time.Minute {
		// 30分钟内过期：每5分钟检查一次
		checkInterval = 5 * time.Minute
	} else if timeToExpiry <= 2*time.Hour {
		// 2小时内过期：每15分钟检查一次
		checkInterval = 15 * time.Minute
	} else if timeToExpiry <= 6*time.Hour {
		// 6小时内过期：每30分钟检查一次
		checkInterval = 30 * time.Minute
	} else if timeToExpiry <= 24*time.Hour {
		// 24小时内过期：每2小时检查一次
		checkInterval = 2 * time.Hour
	} else {
		// 24小时后过期：每6小时检查一次
		checkInterval = 6 * time.Hour
	}
	
	// 确保最小间隔为1分钟，最大间隔为6小时
	if checkInterval < time.Minute {
		checkInterval = time.Minute
	} else if checkInterval > 6*time.Hour {
		checkInterval = 6 * time.Hour
	}
	
	log.Printf("[TokenRefreshService] Token expires in %v, setting check interval to %v", 
		timeToExpiry, checkInterval)
	
	return checkInterval
}

// Stop 停止令牌刷新服务
func (s *TokenRefreshService) Stop() {
	log.Println("[TokenRefreshService] Stopping token refresh service...")
	s.cancel()
}

// refreshExpiredTokens 主动刷新即将过期的令牌 - 使用智能刷新阈值
func (s *TokenRefreshService) refreshExpiredTokens() {
	log.Println("[TokenRefreshService] Checking for tokens that need refresh...")
	
	// 动态计算刷新阈值
	var tokens []models.UserOAuthToken
	refreshThreshold := s.calculateRefreshThreshold()
	
	if err := database.DB.Where("expiry < ? AND expiry > ?", refreshThreshold, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).Find(&tokens).Error; err != nil {
		log.Printf("[TokenRefreshService] Failed to query tokens: %v", err)
		return
	}
	
	if len(tokens) == 0 {
		log.Println("[TokenRefreshService] No tokens need refresh at this time")
		return
	}
	
	log.Printf("[TokenRefreshService] Found %d tokens that need refresh", len(tokens))
	
	successCount := 0
	failureCount := 0
	
	for _, token := range tokens {
		timeToExpiry := token.Expiry.Sub(time.Now())
		log.Printf("[TokenRefreshService] Refreshing token for account %d (expires in %v)", 
			token.EmailAccountID, timeToExpiry)
			
		if err := s.refreshToken(token); err != nil {
			log.Printf("[TokenRefreshService] Failed to refresh token for account %d: %v", token.EmailAccountID, err)
			failureCount++
		} else {
			log.Printf("[TokenRefreshService] Successfully refreshed token for account %d", token.EmailAccountID)
			successCount++
		}
		
		// 根据令牌数量调整延迟
		if len(tokens) > 10 {
			time.Sleep(2 * time.Second) // 大量令牌时较长延迟
		} else {
			time.Sleep(500 * time.Millisecond) // 少量令牌时较短延迟
		}
	}
	
	log.Printf("[TokenRefreshService] Token refresh completed: %d success, %d failure", successCount, failureCount)
}

// calculateRefreshThreshold 动态计算刷新阈值
func (s *TokenRefreshService) calculateRefreshThreshold() time.Time {
	now := time.Now()
	
	// 查询令牌的平均生命周期来智能调整刷新阈值
	var avgLifetime struct {
		AvgHours float64 `json:"avg_hours"`
	}
	
	// 计算最近刷新的令牌的平均生命周期
	database.DB.Raw(`
		SELECT AVG(CAST((julianday(expiry) - julianday(created_at)) * 24 AS REAL)) as avg_hours 
		FROM user_o_auth_tokens 
		WHERE created_at > ? AND expiry > ?
	`, now.AddDate(0, 0, -7), time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).Scan(&avgLifetime)
	
	var refreshHours float64
	
	if avgLifetime.AvgHours > 0 {
		// 根据平均生命周期动态调整
		if avgLifetime.AvgHours <= 2 {
			// 短期令牌(≤2小时)：提前30分钟刷新
			refreshHours = 0.5
		} else if avgLifetime.AvgHours <= 24 {
			// 中期令牌(≤24小时)：提前2小时刷新
			refreshHours = 2
		} else if avgLifetime.AvgHours <= 168 { // 7天
			// 长期令牌(≤7天)：提前6小时刷新
			refreshHours = 6
		} else {
			// 超长期令牌(>7天)：提前24小时刷新
			refreshHours = 24
		}
		
		log.Printf("[TokenRefreshService] Average token lifetime: %.1f hours, refresh threshold: %.1f hours before expiry", 
			avgLifetime.AvgHours, refreshHours)
	} else {
		// 默认策略：提前2小时刷新
		refreshHours = 2
		log.Printf("[TokenRefreshService] Using default refresh threshold: %.1f hours before expiry", refreshHours)
	}
	
	return now.Add(time.Duration(refreshHours * float64(time.Hour)))
}

// refreshToken 刷新单个令牌
func (s *TokenRefreshService) refreshToken(token models.UserOAuthToken) error {
	// 获取邮件账户信息
	var emailAccount models.EmailAccount
	if err := database.DB.First(&emailAccount, token.EmailAccountID).Error; err != nil {
		return err
	}
	
	// 获取OAuth提供商信息
	var provider models.OAuthProvider
	if err := database.DB.First(&provider, token.ProviderID).Error; err != nil {
		return err
	}
	
	// 根据提供商类型使用相应的客户端进行令牌刷新
	// 这里会触发自动令牌刷新逻辑
	switch provider.Name {
	case "google":
		_, err := integrations.GetGmailOAuth2HTTPClient(emailAccount.ID)
		return err
	case "microsoft":
		_, err := integrations.GetMicrosoftOAuth2HTTPClient(emailAccount.ID)
		return err
	default:
		log.Printf("[TokenRefreshService] Unsupported provider: %s", provider.Name)
		return nil
	}
}

// RefreshAllTokens 手动刷新所有令牌
func (s *TokenRefreshService) RefreshAllTokens() {
	log.Println("[TokenRefreshService] Manual refresh of all tokens requested")
	go s.refreshExpiredTokens()
}

// GetTokenStatus 获取令牌状态统计 - 增强版本
func (s *TokenRefreshService) GetTokenStatus() (map[string]any, error) {
	log.Printf("[TokenRefreshService] Starting GetTokenStatus...")
	
	var stats struct {
		Total           int64 `json:"total"`
		Active          int64 `json:"active"`
		Expired         int64 `json:"expired"`
		ExpiringNext1h  int64 `json:"expiring_next_1h"`
		ExpiringNext6h  int64 `json:"expiring_next_6h"`
		ExpiringNext24h int64 `json:"expiring_next_24h"`
		Invalid         int64 `json:"invalid"`
	}
	
	now := time.Now()
	invalidTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// 总令牌数
	database.DB.Model(&models.UserOAuthToken{}).Count(&stats.Total)
	log.Printf("[TokenRefreshService] Total tokens: %d", stats.Total)
	
	// 活跃令牌数 (过期时间在当前时间之后)
	database.DB.Model(&models.UserOAuthToken{}).Where("expiry > ?", now).Count(&stats.Active)
	log.Printf("[TokenRefreshService] Active tokens: %d", stats.Active)
	
	// 过期令牌数 (过期时间在当前时间之前，但不是标记为无效的)
	database.DB.Model(&models.UserOAuthToken{}).Where("expiry < ? AND expiry > ?", now, invalidTime).Count(&stats.Expired)
	log.Printf("[TokenRefreshService] Expired tokens: %d", stats.Expired)
	
	// 即将过期令牌数 (不同时间窗口)
	database.DB.Model(&models.UserOAuthToken{}).Where("expiry BETWEEN ? AND ?", now, now.Add(1*time.Hour)).Count(&stats.ExpiringNext1h)
	database.DB.Model(&models.UserOAuthToken{}).Where("expiry BETWEEN ? AND ?", now, now.Add(6*time.Hour)).Count(&stats.ExpiringNext6h)
	database.DB.Model(&models.UserOAuthToken{}).Where("expiry BETWEEN ? AND ?", now, now.Add(24*time.Hour)).Count(&stats.ExpiringNext24h)
	
	// 无效令牌数 (标记为需要重新授权的)
	database.DB.Model(&models.UserOAuthToken{}).Where("expiry <= ?", invalidTime).Count(&stats.Invalid)
	log.Printf("[TokenRefreshService] Invalid tokens: %d", stats.Invalid)
	
	// 获取详细的令牌信息
	tokenDetails, err := s.getDetailedTokenInfo()
	if err != nil {
		log.Printf("[TokenRefreshService] Failed to get detailed token info: %v", err)
		tokenDetails = []map[string]any{}
	} else {
		log.Printf("[TokenRefreshService] Got %d token details", len(tokenDetails))
	}
	
	// 计算提供商分布
	providerStats, err := s.getProviderStats()
	if err != nil {
		log.Printf("[TokenRefreshService] Failed to get provider stats: %v", err)
		providerStats = map[string]any{}
	} else {
		log.Printf("[TokenRefreshService] Got provider stats: %+v", providerStats)
	}
	
	// 计算下次检查时间
	nextCheckDuration := s.calculateNextCheckInterval()
	
	// 获取最近的刷新阈值
	refreshThreshold := s.calculateRefreshThreshold()
	refreshHours := refreshThreshold.Sub(now).Hours()
	
	result := map[string]any{
		"total":              stats.Total,
		"active":             stats.Active,
		"expired":            stats.Expired,
		"expiring_next_1h":   stats.ExpiringNext1h,
		"expiring_next_6h":   stats.ExpiringNext6h,
		"expiring_next_24h":  stats.ExpiringNext24h,
		"invalid":            stats.Invalid,
		"last_check":         time.Now().Format(time.RFC3339),
		"next_check_in":      nextCheckDuration.String(),
		"refresh_threshold_hours": refreshHours,
		"health_score":       s.calculateHealthScore(stats.Active, stats.Total, stats.Invalid),
		"token_details":      tokenDetails,
		"provider_stats":     providerStats,
	}
	
	log.Printf("[TokenRefreshService] Returning result with %d token details", len(tokenDetails))
	return result, nil
}

// getDetailedTokenInfo 获取详细的令牌信息
func (s *TokenRefreshService) getDetailedTokenInfo() ([]map[string]any, error) {
	var results []models.UserOAuthToken
	
	// 使用GORM的预加载功能
	err := database.DB.
		Preload("EmailAccount").
		Preload("Provider").
		Find(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	now := time.Now()
	invalidTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	
	var tokenDetails []map[string]any
	for _, token := range results {
		timeToExpiry := token.Expiry.Sub(now)
		
		var status string
		if token.Expiry.Before(invalidTime.Add(24 * time.Hour)) {
			status = "invalid"
		} else if token.Expiry.Before(now) {
			status = "expired"
		} else if timeToExpiry <= 1*time.Hour {
			status = "expiring_soon"
		} else if timeToExpiry <= 24*time.Hour {
			status = "expiring_today"
		} else {
			status = "active"
		}
		
		ageHours := now.Sub(token.CreatedAt).Hours()
		lastRefreshHours := now.Sub(token.UpdatedAt).Hours()
		
		// 获取邮箱地址，如果预加载失败则显示ID
		email := "Unknown"
		if token.EmailAccount.EmailAddress != "" {
			email = token.EmailAccount.EmailAddress
		} else {
			email = fmt.Sprintf("Account #%d", token.EmailAccountID)
		}
		
		// 获取提供商名称，如果预加载失败则显示ID
		providerName := "Unknown"
		if token.Provider.Name != "" {
			providerName = token.Provider.Name
		} else {
			providerName = fmt.Sprintf("Provider #%d", token.ProviderID)
		}
		
		tokenDetails = append(tokenDetails, map[string]any{
			"account_id":       token.EmailAccountID,
			"email":           email,
			"provider":        providerName,
			"status":          status,
			"created_at":      token.CreatedAt.Format(time.RFC3339),
			"updated_at":      token.UpdatedAt.Format(time.RFC3339),
			"expiry":          token.Expiry.Format(time.RFC3339),
			"time_to_expiry":  timeToExpiry.String(),
			"age_hours":       int(ageHours),
			"last_refresh_hours": int(lastRefreshHours),
		})
	}
	
	return tokenDetails, nil
}

// getProviderStats 获取提供商统计信息
func (s *TokenRefreshService) getProviderStats() (map[string]any, error) {
	var results []struct {
		ProviderName string `json:"provider_name"`
		Total        int64  `json:"total"`
		Active       int64  `json:"active"`
		Expired      int64  `json:"expired"`
	}
	
	now := time.Now()
	invalidTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	
	err := database.DB.Raw(`
		SELECT 
			op.name as provider_name,
			COUNT(*) as total,
			SUM(CASE WHEN expiry > ? THEN 1 ELSE 0 END) as active,
			SUM(CASE WHEN expiry < ? AND expiry > ? THEN 1 ELSE 0 END) as expired
		FROM user_o_auth_tokens uot
		JOIN o_auth_providers op ON uot.provider_id = op.id
		GROUP BY op.name
	`, now, now, invalidTime).Scan(&results).Error
	
	if err != nil {
		return nil, err
	}
	
	providerStats := make(map[string]any)
	for _, stat := range results {
		providerStats[stat.ProviderName] = map[string]any{
			"total":   stat.Total,
			"active":  stat.Active,
			"expired": stat.Expired,
		}
	}
	
	return providerStats, nil
}

// calculateHealthScore 计算令牌健康评分 (0-100)
func (s *TokenRefreshService) calculateHealthScore(active, total, invalid int64) int {
	if total == 0 {
		return 100 // 没有令牌时认为是健康的
	}
	
	// 基础分数：活跃令牌占比
	activeRatio := float64(active) / float64(total)
	baseScore := activeRatio * 80 // 最高80分
	
	// 惩罚：无效令牌占比
	invalidRatio := float64(invalid) / float64(total)
	penalty := invalidRatio * 30 // 最多扣30分
	
	score := baseScore - penalty
	
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}
	
	return int(score)
}

// CleanupDuplicateTokens 清理重复的令牌记录 - 改进版本
func (s *TokenRefreshService) CleanupDuplicateTokens() (map[string]any, error) {
	log.Println("[TokenRefreshService] Starting token cleanup process...")
	
	// 1. 首先清理"Account #X"格式的测试记录 - 使用更直接的方法
	var testRecordsDeleted int64 = 0
	var accountRecords []struct {
		ID           uint   `json:"id"`
		EmailAddress string `json:"email_address"`
		TokenID      uint   `json:"token_id"`
	}
	
	err := database.DB.Raw(`
		SELECT ea.id, ea.email_address, uot.id as token_id
		FROM email_accounts ea
		JOIN user_o_auth_tokens uot ON ea.id = uot.email_account_id
		WHERE ea.email_address LIKE 'Account #%'
	`).Scan(&accountRecords).Error
	
	if err != nil {
		log.Printf("[TokenRefreshService] Failed to query Account # records: %v", err)
	} else if len(accountRecords) > 0 {
		var tokenIDs []uint
		for _, record := range accountRecords {
			tokenIDs = append(tokenIDs, record.TokenID)
		}
		
		result := database.DB.Delete(&models.UserOAuthToken{}, tokenIDs)
		if result.Error != nil {
			log.Printf("[TokenRefreshService] Failed to delete test tokens: %v", result.Error)
		} else {
			testRecordsDeleted = result.RowsAffected
			log.Printf("[TokenRefreshService] Deleted %d test tokens (Account #X format) with IDs: %v", 
				testRecordsDeleted, tokenIDs)
		}
	}
	
	// 2. 基于邮箱地址的重复检测（更实际的重复）
	type EmailGroup struct {
		EmailAddress string `json:"email_address"`
		Provider     string `json:"provider"`
		Count        int    `json:"count"`
	}
	
	var emailGroups []EmailGroup
	err = database.DB.Raw(`
		SELECT ea.email_address, op.name as provider, COUNT(*) as count
		FROM user_o_auth_tokens uot
		JOIN email_accounts ea ON uot.email_account_id = ea.id
		JOIN o_auth_providers op ON uot.provider_id = op.id
		WHERE ea.email_address != '' AND ea.email_address NOT LIKE 'Account #%'
		GROUP BY ea.email_address, op.name
		HAVING COUNT(*) > 1
	`).Scan(&emailGroups).Error
	
	if err != nil {
		return nil, fmt.Errorf("查询重复邮箱组失败: %v", err)
	}
	
	log.Printf("[TokenRefreshService] Found %d duplicate email groups", len(emailGroups))
	
	totalDuplicates := 0
	duplicatesDeleted := 0
	
	// 3. 对每个重复的邮箱+提供商组合，保留最新的记录
	for _, group := range emailGroups {
		var tokens []struct {
			ID             uint `json:"id"`
			EmailAccountID uint `json:"email_account_id"`
			CreatedAt      time.Time `json:"created_at"`
			UpdatedAt      time.Time `json:"updated_at"`
		}
		
		err := database.DB.Raw(`
			SELECT uot.id, uot.email_account_id, uot.created_at, uot.updated_at
			FROM user_o_auth_tokens uot
			JOIN email_accounts ea ON uot.email_account_id = ea.id
			JOIN o_auth_providers op ON uot.provider_id = op.id
			WHERE ea.email_address = ? AND op.name = ?
			ORDER BY uot.created_at DESC, uot.updated_at DESC
		`, group.EmailAddress, group.Provider).Scan(&tokens).Error
		
		if err != nil {
			log.Printf("[TokenRefreshService] Failed to query tokens for email %s provider %s: %v", 
				group.EmailAddress, group.Provider, err)
			continue
		}
		
		if len(tokens) <= 1 {
			continue // 没有重复，跳过
		}
		
		totalDuplicates += len(tokens)
		
		// 保留第一个（最新的），删除其他的
		var toDelete []uint
		for i := 1; i < len(tokens); i++ {
			toDelete = append(toDelete, tokens[i].ID)
		}
		
		if len(toDelete) > 0 {
			result := database.DB.Delete(&models.UserOAuthToken{}, toDelete)
			if result.Error != nil {
				log.Printf("[TokenRefreshService] Failed to delete duplicate tokens: %v", result.Error)
				continue
			}
			
			deletedCount := int(result.RowsAffected)
			duplicatesDeleted += deletedCount
			
			log.Printf("[TokenRefreshService] Deleted %d duplicate tokens for email %s provider %s", 
				deletedCount, group.EmailAddress, group.Provider)
		}
	}
	
	// 4. 清理过期超过30天的令牌
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	invalidTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	
	var expiredDeleted int64 = 0
	deleteResult := database.DB.Where("expiry < ? AND expiry > ?", thirtyDaysAgo, invalidTime).Delete(&models.UserOAuthToken{})
	if deleteResult.Error != nil {
		log.Printf("[TokenRefreshService] Failed to delete expired tokens: %v", deleteResult.Error)
	} else {
		expiredDeleted = deleteResult.RowsAffected
		log.Printf("[TokenRefreshService] Deleted %d expired tokens (older than 30 days)", expiredDeleted)
	}
	
	// 5. 清理孤儿记录（email_account_id不存在的记录）
	var orphanDeleted int64 = 0
	var orphanTokenIDs []uint
	err = database.DB.Raw(`
		SELECT uot.id 
		FROM user_o_auth_tokens uot
		WHERE uot.email_account_id NOT IN (SELECT id FROM email_accounts)
	`).Scan(&orphanTokenIDs).Error
	
	if err != nil {
		log.Printf("[TokenRefreshService] Failed to query orphan tokens: %v", err)
	} else if len(orphanTokenIDs) > 0 {
		orphanDeleteResult := database.DB.Delete(&models.UserOAuthToken{}, orphanTokenIDs)
		if orphanDeleteResult.Error != nil {
			log.Printf("[TokenRefreshService] Failed to delete orphan tokens: %v", orphanDeleteResult.Error)
		} else {
			orphanDeleted = orphanDeleteResult.RowsAffected
			log.Printf("[TokenRefreshService] Deleted %d orphan tokens (IDs: %v)", 
				orphanDeleted, orphanTokenIDs)
		}
	}
	
	result := map[string]any{
		"duplicate_email_groups":    len(emailGroups),
		"total_duplicates_found":    totalDuplicates,
		"duplicates_deleted":        duplicatesDeleted,
		"test_records_found":        len(accountRecords),
		"test_records_deleted":      testRecordsDeleted,
		"expired_tokens_deleted":    expiredDeleted,
		"orphan_tokens_deleted":     orphanDeleted,
		"total_tokens_deleted":      duplicatesDeleted + int(testRecordsDeleted) + int(expiredDeleted) + int(orphanDeleted),
	}
	
	log.Printf("[TokenRefreshService] Cleanup completed: %+v", result)
	return result, nil
}