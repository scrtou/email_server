package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"email_server/config"
	"email_server/database"
	"email_server/models"
	"email_server/utils"
)

// OAuth2 State管理
type OAuth2State struct {
	State     string    `json:"state"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// 内存中存储state，用于验证OAuth2回调
var (
	stateStore    = make(map[string]*OAuth2State)
	stateMutex    = sync.RWMutex{}
	cleanupTicker *time.Ticker
)

// 初始化state清理器
func init() {
	// 每5分钟清理一次过期的state
	cleanupTicker = time.NewTicker(5 * time.Minute)
	go func() {
		for range cleanupTicker.C {
			cleanupExpiredStates()
		}
	}()
}

// 清理过期的state
func cleanupExpiredStates() {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	now := time.Now()
	count := 0
	for state, stateInfo := range stateStore {
		if now.After(stateInfo.ExpiresAt) {
			delete(stateStore, state)
			count++
		}
	}

	if count > 0 {
		log.Printf("清理了 %d 个过期的OAuth2 state，当前存储数量: %d", count, len(stateStore))
	}

	// 如果存储的state数量过多，记录警告
	if len(stateStore) > 1000 {
		log.Printf("⚠️  OAuth2 state存储数量过多: %d，建议检查清理逻辑", len(stateStore))
	}
}

// LinuxDoOAuth2Login 生成LinuxDo OAuth2登录URL
func LinuxDoOAuth2Login(c *gin.Context) {
	// 生成随机state参数用于防止CSRF攻击
	state, err := generateRandomState()
	if err != nil {
		log.Printf("生成state失败: %v", err)
		utils.SendErrorResponse(c, 500, "系统错误")
		return
	}

	// 将state存储在内存中，10分钟有效期
	expiresAt := time.Now().Add(10 * time.Minute)
	stateMutex.Lock()
	stateStore[state] = &OAuth2State{
		State:     state,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	stateMutex.Unlock()

	log.Printf("创建OAuth2 state: %s, 过期时间: %v", state, expiresAt)

	// 构建授权URL
	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&state=%s&scope=read",
		config.AppConfig.OAuth2.LinuxDo.AuthURL,
		config.AppConfig.OAuth2.LinuxDo.ClientID,
		url.QueryEscape(config.AppConfig.OAuth2.LinuxDo.RedirectURI),
		state,
	)

	utils.SendSuccessResponse(c, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// LinuxDoOAuth2Callback 处理LinuxDo OAuth2回调
func LinuxDoOAuth2Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		utils.SendErrorResponse(c, 400, "缺少授权码")
		return
	}

	// 验证state参数
	stateMutex.Lock()
	stateInfo, exists := stateStore[state]
	if exists {
		delete(stateStore, state) // 使用后立即删除
	}
	stateMutex.Unlock()

	if !exists {
		log.Printf("State验证失败: state=%s 不存在", state)
		c.Redirect(302, "http://localhost:8080/auth/login?error=invalid_state")
		return
	}

	if time.Now().After(stateInfo.ExpiresAt) {
		log.Printf("State验证失败: state=%s 已过期", state)
		c.Redirect(302, "http://localhost:8080/auth/login?error=state_expired")
		return
	}

	log.Printf("State验证成功: %s", state)

	// 使用授权码获取访问令牌
	accessToken, err := exchangeCodeForToken(code)
	if err != nil {
		log.Printf("获取访问令牌失败: %v", err)
		c.Redirect(302, "http://localhost:8080/auth/login?error=token_exchange_failed")
		return
	}

	// 使用访问令牌获取用户信息
	userInfo, err := getLinuxDoUserInfo(accessToken)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		c.Redirect(302, "http://localhost:8080/auth/login?error=user_info_failed")
		return
	}

	// 查找或创建用户
	user, err := findOrCreateLinuxDoUser(userInfo)
	if err != nil {
		log.Printf("创建用户失败: %v", err)
		c.Redirect(302, "http://localhost:8080/auth/login?error=user_creation_failed")
		return
	}

	// 检查用户状态
	if !user.IsActive() {
		log.Printf("用户账户已被封禁: user_id=%d, username=%s", user.ID, user.Username)
		c.Redirect(302, "http://localhost:8080/auth/login?error=account_banned")
		return
	}

	// 更新最后登录时间
	now := time.Now()
	updateResult := database.DB.Model(user).Update("last_login", now)
	if updateResult.Error != nil {
		log.Printf("更新最后登录时间失败: %v", updateResult.Error)
		// 不阻断登录流程，只记录错误
	} else {
		user.LastLogin = &now
	}

	// 生成JWT token
	token, err := utils.GenerateToken(int64(user.ID), user.Username, user.Role)
	if err != nil {
		log.Printf("生成token失败: %v", err)
		// 重定向到前端错误页面
		c.Redirect(302, "http://localhost:8080/auth/login?error=token_generation_failed")
		return
	}

	log.Printf("LinuxDo OAuth2登录成功: user_id=%d, username=%s", user.ID, user.Username)

	// 重定向到前端页面，并在URL中携带token
	// 注意：在生产环境中，应该使用更安全的方式传递token，比如设置HttpOnly cookie
	frontendURL := config.AppConfig.Frontend.BaseURL
	if frontendURL == "" {
		frontendURL = "http://localhost:8080" // 默认值，仅用于开发环境
	}
	redirectURL := fmt.Sprintf("%s/oauth2/callback?token=%s&expires_in=%d", frontendURL, token, config.AppConfig.JWT.ExpiresIn)
	c.Redirect(302, redirectURL)
}

// generateRandomState 生成随机state字符串
func generateRandomState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// exchangeCodeForToken 使用授权码换取访问令牌
func exchangeCodeForToken(code string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", config.AppConfig.OAuth2.LinuxDo.ClientID)
	data.Set("client_secret", config.AppConfig.OAuth2.LinuxDo.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", config.AppConfig.OAuth2.LinuxDo.RedirectURI)

	req, err := http.NewRequest("POST", config.AppConfig.OAuth2.LinuxDo.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Token请求失败: status=%d, body=%s", resp.StatusCode, string(body))
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Token响应: %s", string(body))

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

// getLinuxDoUserInfo 使用访问令牌获取用户信息
func getLinuxDoUserInfo(accessToken string) (*models.LinuxDoUserInfo, error) {
	req, err := http.NewRequest("GET", config.AppConfig.OAuth2.LinuxDo.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo models.LinuxDoUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// findOrCreateLinuxDoUser 查找或创建LinuxDo用户
func findOrCreateLinuxDoUser(userInfo *models.LinuxDoUserInfo) (*models.User, error) {
	var user models.User

	// 首先尝试通过LinuxDoID查找用户
	err := database.DB.Where("linux_do_id = ?", userInfo.ID).First(&user).Error
	if err == nil {
		// 用户已存在，更新信息
		user.Username = userInfo.Username
		user.Email = userInfo.Email
		if err := database.DB.Save(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	// 如果通过LinuxDoID没找到，尝试通过邮箱查找
	err = database.DB.Where("email = ?", userInfo.Email).First(&user).Error
	if err == nil {
		// 邮箱已存在，绑定LinuxDo账号
		user.LinuxDoID = &userInfo.ID
		provider := "linuxdo"
		user.Provider = &provider
		if err := database.DB.Save(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	// 用户不存在，创建新用户
	provider := "linuxdo"
	user = models.User{
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		LinuxDoID: &userInfo.ID,
		Provider:  &provider,
		Role:      models.RoleUser,     // 默认为普通用户
		Status:    models.StatusActive, // 默认为激活状态
		// Password留空，因为是OAuth用户
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetOAuth2StateStats 获取OAuth2 state存储统计信息（仅用于监控）
func GetOAuth2StateStats(c *gin.Context) {
	stateMutex.RLock()
	defer stateMutex.RUnlock()

	now := time.Now()
	total := len(stateStore)
	expired := 0

	for _, stateInfo := range stateStore {
		if now.After(stateInfo.ExpiresAt) {
			expired++
		}
	}

	stats := gin.H{
		"total_states":          total,
		"expired_states":        expired,
		"active_states":         total - expired,
		"memory_usage_estimate": fmt.Sprintf("~%d KB", total*100/1024), // 粗略估算
	}

	utils.SendSuccessResponse(c, stats)
}
