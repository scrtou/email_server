// Enhanced Token Refresh Service - 增强版令牌刷新服务
// 实现接近"永不过期"的用户体验
package services

import (
	"context"
	"fmt"
	"log"
	"time"
	"sync"
	"math/rand"

	"email_server/database"
	"email_server/integrations"
	"email_server/models"
)

// EnhancedTokenRefreshService 增强版令牌刷新服务
type EnhancedTokenRefreshService struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	
	// 刷新策略配置
	config RefreshConfig
}

type RefreshConfig struct {
	// 预防性刷新：在token过期前多长时间就开始刷新
	PreventiveRefreshThreshold time.Duration
	
	// 最小检查间隔
	MinCheckInterval time.Duration
	
	// 失败重试配置
	MaxRetryAttempts int
	RetryBackoffBase time.Duration
	
	// 是否启用随机抖动避免同时刷新
	EnableJitter bool
}

// NewEnhancedTokenRefreshService 创建增强版令牌刷新服务
func NewEnhancedTokenRefreshService() *EnhancedTokenRefreshService {
	ctx, cancel := context.WithCancel(context.Background())
	return &EnhancedTokenRefreshService{
		ctx:    ctx,
		cancel: cancel,
		config: RefreshConfig{
			// 非常激进的预防性刷新：token过期前75%的时间就开始刷新
			PreventiveRefreshThreshold: 45 * time.Minute, // 假设token 1小时过期，提前45分钟刷新
			MinCheckInterval:          30 * time.Second,   // 30秒检查一次
			MaxRetryAttempts:          5,
			RetryBackoffBase:          2 * time.Second,
			EnableJitter:             true,
		},
	}
}

// Start 启动增强版令牌刷新服务
func (s *EnhancedTokenRefreshService) Start() {
	log.Println("[EnhancedTokenRefreshService] Starting ultra-aggressive token refresh service for near-永不过期 experience...")
	
	// 启动多个goroutine并行处理
	go s.continuousRefreshMonitor()
	go s.proactiveRefreshWorker()
	go s.emergencyRefreshHandler()
}

// continuousRefreshMonitor 连续刷新监控器 - 主要刷新逻辑
func (s *EnhancedTokenRefreshService) continuousRefreshMonitor() {
	ticker := time.NewTicker(s.config.MinCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			log.Println("[EnhancedTokenRefreshService] Continuous refresh monitor stopped")
			return
		case <-ticker.C:
			s.performPreventiveRefresh()
		}
	}
}

// proactiveRefreshWorker 主动刷新工作器 - 处理即将过期的token
func (s *EnhancedTokenRefreshService) proactiveRefreshWorker() {
	// 每10秒检查一次即将过期的token
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.handleExpiringTokens()
		}
	}
}

// emergencyRefreshHandler 紧急刷新处理器 - 处理已过期但还有refresh token的情况
func (s *EnhancedTokenRefreshService) emergencyRefreshHandler() {
	// 每5秒检查一次过期的token
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.emergencyRefresh()
		}
	}
}

// performPreventiveRefresh 执行预防性刷新
func (s *EnhancedTokenRefreshService) performPreventiveRefresh() {
	now := time.Now()
	threshold := now.Add(s.config.PreventiveRefreshThreshold)
	
	var tokens []models.UserOAuthToken
	if err := database.DB.Where("expiry < ? AND expiry > ?", 
		threshold, 
		time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).Find(&tokens).Error; err != nil {
		log.Printf("[EnhancedTokenRefreshService] Failed to query tokens for preventive refresh: %v", err)
		return
	}
	
	if len(tokens) == 0 {
		return
	}
	
	log.Printf("[EnhancedTokenRefreshService] Found %d tokens for preventive refresh", len(tokens))
	
	// 并发刷新，但限制并发数
	semaphore := make(chan struct{}, 5) // 最多5个并发
	var wg sync.WaitGroup
	
	for _, token := range tokens {
		wg.Add(1)
		go func(t models.UserOAuthToken) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			if err := s.refreshTokenWithRetry(t); err != nil {
				log.Printf("[EnhancedTokenRefreshService] Preventive refresh failed for account %d: %v", 
					t.EmailAccountID, err)
			} else {
				log.Printf("[EnhancedTokenRefreshService] Preventive refresh successful for account %d", 
					t.EmailAccountID)
			}
		}(token)
	}
	
	wg.Wait()
}

// handleExpiringTokens 处理即将过期的token（更激进的策略）
func (s *EnhancedTokenRefreshService) handleExpiringTokens() {
	now := time.Now()
	
	// 查找未来15分钟内过期的token
	var tokens []models.UserOAuthToken
	if err := database.DB.Where("expiry BETWEEN ? AND ?", 
		now, now.Add(15*time.Minute)).Find(&tokens).Error; err != nil {
		return
	}
	
	for _, token := range tokens {
		go func(t models.UserOAuthToken) {
			timeToExpiry := t.Expiry.Sub(time.Now())
			log.Printf("[EnhancedTokenRefreshService] URGENT: Token expires in %v for account %d, refreshing immediately...", 
				timeToExpiry, t.EmailAccountID)
				
			if err := s.refreshTokenWithRetry(t); err != nil {
				log.Printf("[EnhancedTokenRefreshService] URGENT refresh failed for account %d: %v", 
					t.EmailAccountID, err)
			}
		}(token)
	}
}

// emergencyRefresh 紧急刷新已过期但可能仍可刷新的token
func (s *EnhancedTokenRefreshService) emergencyRefresh() {
	now := time.Now()
	fiveMinutesAgo := now.Add(-5 * time.Minute)
	
	// 查找5分钟内刚过期的token
	var tokens []models.UserOAuthToken
	if err := database.DB.Where("expiry BETWEEN ? AND ? AND refresh_token_encrypted != ''", 
		fiveMinutesAgo, now).Find(&tokens).Error; err != nil {
		return
	}
	
	for _, token := range tokens {
		go func(t models.UserOAuthToken) {
			log.Printf("[EnhancedTokenRefreshService] EMERGENCY: Attempting to refresh recently expired token for account %d", 
				t.EmailAccountID)
				
			if err := s.refreshTokenWithRetry(t); err != nil {
				log.Printf("[EnhancedTokenRefreshService] Emergency refresh failed for account %d: %v", 
					t.EmailAccountID, err)
				
				// 记录需要用户重新授权的账户
				s.markAccountForReauth(t)
			} else {
				log.Printf("[EnhancedTokenRefreshService] Emergency refresh SUCCESS for account %d!", 
					t.EmailAccountID)
			}
		}(token)
	}
}

// refreshTokenWithRetry 带重试机制的token刷新
func (s *EnhancedTokenRefreshService) refreshTokenWithRetry(token models.UserOAuthToken) error {
	var lastErr error
	
	for attempt := 0; attempt < s.config.MaxRetryAttempts; attempt++ {
		if attempt > 0 {
			// 指数退避 + 随机抖动
			backoff := s.config.RetryBackoffBase * time.Duration(1<<uint(attempt))
			if s.config.EnableJitter {
				jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
				backoff += jitter
			}
			
			log.Printf("[EnhancedTokenRefreshService] Retry attempt %d for account %d after %v", 
				attempt+1, token.EmailAccountID, backoff)
			
			select {
			case <-s.ctx.Done():
				return lastErr
			case <-time.After(backoff):
			}
		}
		
		if err := s.refreshSingleToken(token); err != nil {
			lastErr = err
			log.Printf("[EnhancedTokenRefreshService] Refresh attempt %d failed for account %d: %v", 
				attempt+1, token.EmailAccountID, err)
			continue
		}
		
		log.Printf("[EnhancedTokenRefreshService] Refresh successful on attempt %d for account %d", 
			attempt+1, token.EmailAccountID)
		return nil
	}
	
	return fmt.Errorf("refresh failed after %d attempts: %v", s.config.MaxRetryAttempts, lastErr)
}

// refreshSingleToken 刷新单个token（复用现有逻辑）
func (s *EnhancedTokenRefreshService) refreshSingleToken(token models.UserOAuthToken) error {
	var emailAccount models.EmailAccount
	if err := database.DB.First(&emailAccount, token.EmailAccountID).Error; err != nil {
		return err
	}
	
	var provider models.OAuthProvider
	if err := database.DB.First(&provider, token.ProviderID).Error; err != nil {
		return err
	}
	
	switch provider.Name {
	case "google":
		_, err := integrations.GetGmailOAuth2HTTPClient(emailAccount.ID)
		return err
	case "microsoft":
		_, err := integrations.GetMicrosoftOAuth2HTTPClient(emailAccount.ID)
		return err
	default:
		return fmt.Errorf("unsupported provider: %s", provider.Name)
	}
}

// markAccountForReauth 标记账户需要重新授权
func (s *EnhancedTokenRefreshService) markAccountForReauth(token models.UserOAuthToken) {
	invalidTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	database.DB.Model(&token).Update("expiry", invalidTime)
	
	log.Printf("[EnhancedTokenRefreshService] Marked account %d as requiring re-authorization", 
		token.EmailAccountID)
}

// Stop 停止服务
func (s *EnhancedTokenRefreshService) Stop() {
	log.Println("[EnhancedTokenRefreshService] Stopping enhanced token refresh service...")
	s.cancel()
}

// GetRefreshStats 获取刷新统计信息
func (s *EnhancedTokenRefreshService) GetRefreshStats() map[string]interface{} {
	now := time.Now()
	stats := make(map[string]interface{})
	
	var counts struct {
		Total           int64 `json:"total"`
		Active          int64 `json:"active"`
		ExpiringNext5m  int64 `json:"expiring_next_5m"`
		ExpiringNext15m int64 `json:"expiring_next_15m"`
		ExpiringNext1h  int64 `json:"expiring_next_1h"`
		RequireReauth   int64 `json:"require_reauth"`
	}
	
	database.DB.Model(&models.UserOAuthToken{}).Count(&counts.Total)
	database.DB.Model(&models.UserOAuthToken{}).Where("expiry > ?", now).Count(&counts.Active)
	database.DB.Model(&models.UserOAuthToken{}).Where("expiry BETWEEN ? AND ?", now, now.Add(5*time.Minute)).Count(&counts.ExpiringNext5m)
	database.DB.Model(&models.UserOAuthToken{}).Where("expiry BETWEEN ? AND ?", now, now.Add(15*time.Minute)).Count(&counts.ExpiringNext15m)
	database.DB.Model(&models.UserOAuthToken{}).Where("expiry BETWEEN ? AND ?", now, now.Add(1*time.Hour)).Count(&counts.ExpiringNext1h)
	database.DB.Model(&models.UserOAuthToken{}).Where("expiry <= ?", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).Count(&counts.RequireReauth)
	
	stats["token_counts"] = counts
	stats["config"] = s.config
	stats["status"] = "running"
	
	return stats
} 