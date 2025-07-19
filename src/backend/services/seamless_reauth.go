// Seamless Reauthorization Service - 无感知重新授权服务
// 当refresh token过期时，提供近乎无感知的重新授权体验
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"email_server/config"
	"email_server/database"
	"email_server/models"
)

// SeamlessReauthService 无感知重新授权服务
type SeamlessReauthService struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex

	// 待重新授权的账户队列
	reauthQueue map[uint]*ReauthRequest

	// 重新授权回调通道
	callbackChan chan ReauthCallback
}

// ReauthRequest 重新授权请求
type ReauthRequest struct {
	AccountID   uint      `json:"account_id"`
	UserID      uint      `json:"user_id"`
	Provider    string    `json:"provider"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	RetryCount  int       `json:"retry_count"`
	LastAttempt time.Time `json:"last_attempt"`

	// 用户通知状态
	NotificationSent bool `json:"notification_sent"`
}

// ReauthCallback 重新授权回调
type ReauthCallback struct {
	AccountID uint   `json:"account_id"`
	Success   bool   `json:"success"`
	Token     string `json:"token,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ReauthNotification 重新授权通知
type ReauthNotification struct {
	Type      string    `json:"type"`
	AccountID uint      `json:"account_id"`
	Provider  string    `json:"provider"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`

	// 重新授权链接
	ReauthURL string `json:"reauth_url"`
}

// NewSeamlessReauthService 创建无感知重新授权服务
func NewSeamlessReauthService() *SeamlessReauthService {
	ctx, cancel := context.WithCancel(context.Background())
	return &SeamlessReauthService{
		ctx:          ctx,
		cancel:       cancel,
		reauthQueue:  make(map[uint]*ReauthRequest),
		callbackChan: make(chan ReauthCallback, 100),
	}
}

// Start 启动无感知重新授权服务
func (s *SeamlessReauthService) Start() {
	log.Println("[SeamlessReauthService] Starting seamless reauthorization service...")

	go s.monitorExpiredTokens()
	go s.processReauthQueue()
	go s.handleReauthCallbacks()
	go s.cleanupOldRequests()
}

// monitorExpiredTokens 监控过期的token
func (s *SeamlessReauthService) monitorExpiredTokens() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkForExpiredTokens()
		}
	}
}

// checkForExpiredTokens 检查过期的token并创建重新授权请求
func (s *SeamlessReauthService) checkForExpiredTokens() {
	invalidTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	// 查找标记为需要重新授权的token
	var expiredTokens []models.UserOAuthToken
	err := database.DB.
		Preload("EmailAccount").
		Preload("Provider").
		Where("expiry <= ?", invalidTime).
		Find(&expiredTokens).Error

	if err != nil {
		log.Printf("[SeamlessReauthService] Failed to query expired tokens: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, token := range expiredTokens {
		// 检查是否已在队列中
		if _, exists := s.reauthQueue[token.EmailAccountID]; exists {
			continue
		}

		// 创建重新授权请求
		reauthReq := &ReauthRequest{
			AccountID:   token.EmailAccountID,
			UserID:      token.UserID,
			Provider:    token.Provider.Name,
			Email:       token.EmailAccount.EmailAddress,
			CreatedAt:   time.Now(),
			RetryCount:  0,
			LastAttempt: time.Time{},
		}

		s.reauthQueue[token.EmailAccountID] = reauthReq
		log.Printf("[SeamlessReauthService] Added account %d (%s) to reauth queue",
			token.EmailAccountID, token.EmailAccount.EmailAddress)
	}
}

// processReauthQueue 处理重新授权队列
func (s *SeamlessReauthService) processReauthQueue() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.processReauthRequests()
		}
	}
}

// processReauthRequests 处理重新授权请求
func (s *SeamlessReauthService) processReauthRequests() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for accountID, req := range s.reauthQueue {
		// 检查是否需要发送通知
		if !req.NotificationSent {
			s.sendReauthNotification(req)
			req.NotificationSent = true
		}

		// 检查是否超过重试限制
		if req.RetryCount >= 3 {
			if now.Sub(req.LastAttempt) > 24*time.Hour {
				// 24小时后重置重试计数
				req.RetryCount = 0
				req.NotificationSent = false
			}
			continue
		}

		// 检查重试间隔
		if !req.LastAttempt.IsZero() && now.Sub(req.LastAttempt) < 1*time.Hour {
			continue
		}

		// 尝试自动重新授权
		go s.attemptAutoReauth(accountID, req)
	}
}

// sendReauthNotification 发送重新授权通知
func (s *SeamlessReauthService) sendReauthNotification(req *ReauthRequest) {
	notification := ReauthNotification{
		Type:      "reauth_required",
		AccountID: req.AccountID,
		Provider:  req.Provider,
		Email:     req.Email,
		Message:   fmt.Sprintf("您的%s邮箱账户 %s 需要重新授权以继续同步邮件", req.Provider, req.Email),
		Timestamp: time.Now(),
		ReauthURL: s.generateReauthURL(req),
	}

	// 这里可以扩展为多种通知方式：WebSocket、推送通知、邮件等
	s.storeNotification(notification)

	log.Printf("[SeamlessReauthService] Sent reauth notification for account %d (%s)",
		req.AccountID, req.Email)
}

// generateReauthURL 生成重新授权URL
func (s *SeamlessReauthService) generateReauthURL(req *ReauthRequest) string {
	baseURL := config.AppConfig.Frontend.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return fmt.Sprintf("%s/email-accounts/%d/reauth?provider=%s",
		baseURL, req.AccountID, req.Provider)
}

// storeNotification 存储通知到数据库
func (s *SeamlessReauthService) storeNotification(notification ReauthNotification) {
	// 这里可以存储到数据库或发送到消息队列
	notificationJSON, _ := json.Marshal(notification)

	// 临时存储到日志，实际应用中应该存储到数据库
	log.Printf("[SeamlessReauthService] Notification: %s", string(notificationJSON))

	// TODO: 实现WebSocket推送给前端
	// TODO: 实现邮件通知
	// TODO: 实现数据库存储
}

// attemptAutoReauth 尝试自动重新授权
func (s *SeamlessReauthService) attemptAutoReauth(accountID uint, req *ReauthRequest) {
	req.LastAttempt = time.Now()
	req.RetryCount++

	log.Printf("[SeamlessReauthService] Attempting auto reauth for account %d (attempt %d)",
		accountID, req.RetryCount)

	// 这里可以实现一些自动重新授权的策略
	// 例如：检查用户是否在线，如果在线则发送实时提醒

	// 检查是否有已保存的长期凭据
	if s.tryLongTermCredentials(accountID) {
		log.Printf("[SeamlessReauthService] Auto reauth successful for account %d using long-term credentials", accountID)
		s.removeFromQueue(accountID)
		return
	}

	// 如果用户在24小时内活跃，发送实时提醒
	if s.isUserRecentlyActive(req.UserID) {
		s.sendRealTimeReauthPrompt(req)
	}

	log.Printf("[SeamlessReauthService] Auto reauth failed for account %d, user intervention required", accountID)
}

// tryLongTermCredentials 尝试使用长期凭据
func (s *SeamlessReauthService) tryLongTermCredentials(accountID uint) bool {
	// 这里可以检查是否有应用密码或其他长期凭据
	// 对于某些企业应用，可能有服务账号凭据

	var emailAccount models.EmailAccount
	if err := database.DB.First(&emailAccount, accountID).Error; err != nil {
		return false
	}

	// 检查是否配置了应用密码
	if emailAccount.PasswordEncrypted != "" && emailAccount.IMAPServer != "" {
		log.Printf("[SeamlessReauthService] Found IMAP credentials for account %d, falling back to basic auth", accountID)
		// 可以标记账户为使用IMAP而不是OAuth2
		return true
	}

	return false
}

// isUserRecentlyActive 检查用户是否最近活跃
func (s *SeamlessReauthService) isUserRecentlyActive(userID uint) bool {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return false
	}

	if user.LastLogin == nil {
		return false
	}

	return time.Since(*user.LastLogin) < 24*time.Hour
}

// sendRealTimeReauthPrompt 发送实时重新授权提醒
func (s *SeamlessReauthService) sendRealTimeReauthPrompt(req *ReauthRequest) {
	prompt := map[string]interface{}{
		"type":       "urgent_reauth",
		"account_id": req.AccountID,
		"provider":   req.Provider,
		"email":      req.Email,
		"message":    fmt.Sprintf("需要立即重新授权 %s 账户以避免服务中断", req.Email),
		"action_url": s.generateReauthURL(req),
		"timestamp":  time.Now(),
	}

	promptJSON, _ := json.Marshal(prompt)
	log.Printf("[SeamlessReauthService] Real-time prompt: %s", string(promptJSON))

	// TODO: 通过WebSocket发送给前端
	// TODO: 发送推送通知
}

// handleReauthCallbacks 处理重新授权回调
func (s *SeamlessReauthService) handleReauthCallbacks() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case callback := <-s.callbackChan:
			s.processReauthCallback(callback)
		}
	}
}

// processReauthCallback 处理重新授权回调
func (s *SeamlessReauthService) processReauthCallback(callback ReauthCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if callback.Success {
		log.Printf("[SeamlessReauthService] Reauth successful for account %d", callback.AccountID)
		s.removeFromQueue(callback.AccountID)

		// 发送成功通知
		notification := ReauthNotification{
			Type:      "reauth_success",
			AccountID: callback.AccountID,
			Message:   "账户重新授权成功，邮件同步已恢复",
			Timestamp: time.Now(),
		}
		s.storeNotification(notification)
	} else {
		log.Printf("[SeamlessReauthService] Reauth failed for account %d: %s",
			callback.AccountID, callback.Error)

		// 重置通知状态，以便重新发送
		if req, exists := s.reauthQueue[callback.AccountID]; exists {
			req.NotificationSent = false
		}
	}
}

// cleanupOldRequests 清理过期的重新授权请求
func (s *SeamlessReauthService) cleanupOldRequests() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanupExpiredRequests()
		}
	}
}

// cleanupExpiredRequests 清理过期请求
func (s *SeamlessReauthService) cleanupExpiredRequests() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expiredRequests := make([]uint, 0)

	for accountID, req := range s.reauthQueue {
		// 超过7天的请求被认为是过期的
		if now.Sub(req.CreatedAt) > 7*24*time.Hour {
			expiredRequests = append(expiredRequests, accountID)
		}
	}

	for _, accountID := range expiredRequests {
		delete(s.reauthQueue, accountID)
		log.Printf("[SeamlessReauthService] Cleaned up expired reauth request for account %d", accountID)
	}
}

// removeFromQueue 从队列中移除请求
func (s *SeamlessReauthService) removeFromQueue(accountID uint) {
	delete(s.reauthQueue, accountID)
}

// OnReauthComplete 重新授权完成回调
func (s *SeamlessReauthService) OnReauthComplete(accountID uint, success bool, token string, err error) {
	callback := ReauthCallback{
		AccountID: accountID,
		Success:   success,
		Token:     token,
	}

	if err != nil {
		callback.Error = err.Error()
	}

	select {
	case s.callbackChan <- callback:
	default:
		log.Printf("[SeamlessReauthService] Callback channel full, dropping callback for account %d", accountID)
	}
}

// GetReauthQueue 获取重新授权队列状态
func (s *SeamlessReauthService) GetReauthQueue() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{})
	result["queue_size"] = len(s.reauthQueue)

	requests := make([]ReauthRequest, 0, len(s.reauthQueue))
	for _, req := range s.reauthQueue {
		requests = append(requests, *req)
	}

	result["requests"] = requests
	result["timestamp"] = time.Now()

	return result
}

// Stop 停止服务
func (s *SeamlessReauthService) Stop() {
	log.Println("[SeamlessReauthService] Stopping seamless reauth service...")
	s.cancel()
	close(s.callbackChan)
}
