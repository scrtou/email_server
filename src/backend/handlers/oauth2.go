package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	pkce "github.com/nirasan/go-oauth-pkce-code-verifier"

	"email_server/config"
	"email_server/database"
	"email_server/models"
	"email_server/utils"
)

// generateRandomState 生成随机state字符串
func generateRandomState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// truncateString 截断字符串用于日志显示
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// --- LinuxDo OAuth2 流程 ---

// LinuxDoOAuth2Login 生成LinuxDo OAuth2登录URL
// 已更新为使用数据库存储state
func LinuxDoOAuth2Login(c *gin.Context) {
	state, err := generateRandomState()
	if err != nil {
		log.Printf("生成state失败: %v", err)
		utils.SendErrorResponse(c, 500, "系统错误")
		return
	}

	expiresAt := time.Now().Add(10 * time.Minute)

	// 创建并保存 state 到数据库
	oauthState := models.OAuth2State{
		State:     state,
		ExpiresAt: expiresAt,
	}
	if err := database.DB.Create(&oauthState).Error; err != nil {
		log.Printf("保存OAuth2 state到数据库失败: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "无法启动OAuth2流程")
		return
	}

	log.Printf("创建并保存LinuxDo OAuth2 state到数据库: %s, 过期时间: %v", state, expiresAt)

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
// 已更新为使用数据库验证state
func LinuxDoOAuth2Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		utils.SendErrorResponse(c, 400, "缺少授权码")
		return
	}

	// 1. 从数据库验证 state，并在事务中立即删除
	var stateInfo models.OAuth2State
	tx := database.DB.Begin()
	if err := tx.Where("state = ?", state).First(&stateInfo).Error; err != nil {
		tx.Rollback()
		log.Printf("State验证失败: state=%s 在数据库中不存在或查询出错: %v", state, err)
		c.Redirect(302, "http://localhost:8080/auth/login?error=invalid_state")
		return
	}
	if err := tx.Delete(&stateInfo).Error; err != nil {
		tx.Rollback()
		log.Printf("从数据库删除 state 失败: %v", err)
		c.Redirect(302, "http://localhost:8080/auth/login?error=internal_error")
		return
	}
	if err := tx.Commit().Error; err != nil {
		log.Printf("提交 state 删除事务失败: %v", err)
		c.Redirect(302, "http://localhost:8080/auth/login?error=internal_error")
		return
	}

	if time.Now().After(stateInfo.ExpiresAt) {
		log.Printf("State验证失败: state=%s 已过期", state)
		c.Redirect(302, "http://localhost:8080/auth/login?error=state_expired")
		return
	}

	log.Printf("State验证成功: %s", state)

	accessToken, err := exchangeCodeForToken(code)
	if err != nil {
		log.Printf("获取访问令牌失败: %v", err)
		c.Redirect(302, "http://localhost:8080/auth/login?error=token_exchange_failed")
		return
	}

	userInfo, err := getLinuxDoUserInfo(accessToken)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		c.Redirect(302, "http://localhost:8080/auth/login?error=user_info_failed")
		return
	}

	user, err := findOrCreateLinuxDoUser(userInfo)
	if err != nil {
		log.Printf("创建用户失败: %v", err)
		c.Redirect(302, "http://localhost:8080/auth/login?error=user_creation_failed")
		return
	}

	if !user.IsStatusActive() {
		log.Printf("用户账户已被封禁: user_id=%d, username=%s", user.ID, user.Username)
		c.Redirect(302, "http://localhost:8080/auth/login?error=account_banned")
		return
	}

	now := time.Now()
	if err := database.DB.Model(user).Update("last_login", now).Error; err != nil {
		log.Printf("更新最后登录时间失败: %v", err)
	}

	token, err := utils.GenerateToken(int64(user.ID), user.Username, user.Role)
	if err != nil {
		log.Printf("生成token失败: %v", err)
		c.Redirect(302, "http://localhost:8080/auth/login?error=token_generation_failed")
		return
	}

	log.Printf("LinuxDo OAuth2登录成功: user_id=%d, username=%s", user.ID, user.Username)

	frontendURL := config.AppConfig.Frontend.BaseURL
	if frontendURL == "" {
		frontendURL = "http://localhost:8080"
	}
	redirectURL := fmt.Sprintf("%s/oauth2/callback?token=%s&expires_in=%d", frontendURL, token, config.AppConfig.JWT.ExpiresIn)
	c.Redirect(302, redirectURL)
}

// --- LinuxDo 辅助函数 ---

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

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}
	return tokenResponse.AccessToken, nil
}

func getLinuxDoUserInfo(accessToken string) (*models.LinuxDoUserInfo, error) {
	req, err := http.NewRequest("GET", config.AppConfig.OAuth2.LinuxDo.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := (&http.Client{}).Do(req)
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

func findOrCreateLinuxDoUser(userInfo *models.LinuxDoUserInfo) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("linux_do_id = ?", userInfo.ID).First(&user).Error; err == nil {
		user.Username = userInfo.Username
		user.Email = userInfo.Email
		return &user, database.DB.Save(&user).Error
	}

	if err := database.DB.Where("email = ?", userInfo.Email).First(&user).Error; err == nil {
		user.LinuxDoID = &userInfo.ID
		provider := "linuxdo"
		user.Provider = &provider
		return &user, database.DB.Save(&user).Error
	}

	provider := "linuxdo"
	user = models.User{
		Username:  userInfo.Username,
		Email:     userInfo.Email,
		LinuxDoID: &userInfo.ID,
		Provider:  &provider,
		Role:      models.RoleUser,
		Status:    models.StatusActive,
	}
	return &user, database.DB.Create(&user).Error
}

// --- 通用 OAuth2 提供商流程 (Microsoft, Google, etc.) ---

// getOAuth2Config 从数据库动态构建 oauth2.Config
func getOAuth2Config(providerName string) (*oauth2.Config, error) {
	var provider models.OAuthProvider
	if err := database.DB.Where("name = ?", providerName).First(&provider).Error; err != nil {
		return nil, fmt.Errorf("provider '%s' not found in database", providerName)
	}

	decryptedSecret, err := utils.Decrypt(provider.ClientSecretEncrypted)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt client secret for provider '%s'", providerName)
	}

	baseURL := config.AppConfig.Backend.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:5555"
	}
	redirectURL := fmt.Sprintf("%s/api/v1/oauth2/callback/%s", baseURL, provider.Name)

	log.Printf("[DEBUG] Preparing OAuth2 config for provider '%s'", providerName)
	log.Printf("[DEBUG]   -> ClientID: %s", provider.ClientID)
	log.Printf("[DEBUG]   -> ClientSecret (decrypted): %s", string(decryptedSecret))
	log.Printf("[DEBUG]   -> RedirectURL: %s", redirectURL)

	return &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: string(decryptedSecret),
		RedirectURL:  redirectURL,
		Scopes:       strings.Split(provider.Scopes, ","),
		Endpoint: oauth2.Endpoint{
			AuthURL:  provider.AuthURL,
			TokenURL: provider.TokenURL,
		},
	}, nil
}

// RedirectToOAuthProvider 将用户重定向到所选提供商的授权页面
func RedirectToOAuthProvider(c *gin.Context) {
	providerName := c.Param("provider")
	accountIDStr := c.Query("account_id")

	if accountIDStr == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "account_id is required")
		return
	}
	accountID, err := utils.StringToUint(accountIDStr)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid account_id")
		return
	}

	userID, _ := c.Get("user_id")
	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", accountID, userID).First(&emailAccount).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusForbidden, "Access denied to this email account")
		return
	}

	conf, err := getOAuth2Config(providerName)
	if err != nil {
		log.Printf("Error getting OAuth2 config for %s: %v", providerName, err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "Unsupported or misconfigured OAuth2 provider")
		return
	}

	state, err := generateRandomState()
	if err != nil {
		log.Printf("生成state失败: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "无法启动OAuth2流程")
		return
	}

	pkceVerifier, err := pkce.CreateCodeVerifier()
	if err != nil {
		log.Printf("生成PKCE verifier失败: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "无法启动OAuth2流程")
		return
	}

	expiresAt := time.Now().Add(10 * time.Minute)

	// 创建并保存 state 到数据库
	oauthState := models.OAuth2State{
		State:        state,
		AccountID:    accountID,
		PKCEVerifier: pkceVerifier.Value,
		ExpiresAt:    expiresAt,
	}
	if err := database.DB.Create(&oauthState).Error; err != nil {
		log.Printf("保存OAuth2 state到数据库失败: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "无法启动OAuth2流程")
		return
	}
	log.Printf("创建并保存OAuth2 state到数据库: %s, 过期时间: %v", state, expiresAt)

	var authURL string
	if providerName == "microsoft" {
		authURL = conf.AuthCodeURL(state,
			oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("code_challenge", pkceVerifier.CodeChallengeS256()),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
			oauth2.SetAuthURLParam("prompt", "consent"),
			oauth2.SetAuthURLParam("response_mode", "query"),
		)
	} else {
		authURL = conf.AuthCodeURL(state,
			oauth2.AccessTypeOffline,
			oauth2.ApprovalForce,
			oauth2.SetAuthURLParam("code_challenge", pkceVerifier.CodeChallengeS256()),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
	}

	log.Printf("[DEBUG] 生成的授权URL: %s", authURL)
	utils.SendSuccessResponse(c, gin.H{"auth_url": authURL})
}

// HandleOAuth2Callback 处理来自提供商的回调
// HandleOAuth2Callback 处理来自提供商的回调
func HandleOAuth2Callback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	log.Printf("[DEBUG] OAuth2回调 - Provider: %s, Code: %s, State: %s, Error: %s", provider, truncateString(code, 20), state, errorParam)

	if errorParam != "" {
		errorDesc := c.Query("error_description")
		log.Printf("OAuth2授权被拒绝: error=%s, description=%s", errorParam, errorDesc)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=authorization_denied&details="+errorParam)
		return
	}
	if code == "" {
		log.Printf("OAuth2回调缺少授权码")
		c.Redirect(http.StatusTemporaryRedirect, "/?error=missing_code")
		return
	}

	// 1. 从数据库验证 state，并在事务中立即删除
	var stateInfo models.OAuth2State
	tx := database.DB.Begin()
	if err := tx.Where("state = ?", state).First(&stateInfo).Error; err != nil {
		tx.Rollback()
		log.Printf("State验证失败: state=%s 在数据库中不存在或查询出错: %v", state, err)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=invalid_state")
		return
	}
	if err := tx.Delete(&stateInfo).Error; err != nil {
		tx.Rollback()
		log.Printf("从数据库删除 state 失败: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=internal_error")
		return
	}
	if err := tx.Commit().Error; err != nil {
		log.Printf("提交 state 删除事务失败: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=internal_error")
		return
	}

	if time.Now().After(stateInfo.ExpiresAt) {
		log.Printf("State验证失败: state=%s 已过期", state)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=state_expired")
		return
	}
	log.Printf("[DEBUG] State从数据库验证成功: %s", state)

	// 2. 准备交换token
	conf, err := getOAuth2Config(provider)
	if err != nil {
		log.Printf("Error getting OAuth2 config for %s: %v", provider, err)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=provider_not_configured")
		return
	}

	pkceVerifier := stateInfo.PKCEVerifier
	if pkceVerifier == "" {
		log.Printf("State验证失败: state=%s 缺少PKCE verifier", state)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=internal_error")
		return
	}
	log.Printf("[DEBUG] 准备交换token - Code: %s, PKCE Verifier: %s", truncateString(code, 20), truncateString(pkceVerifier, 10))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	// (可选) 调试代码
	httpClient := &http.Client{Transport: &loggingTransport{T: http.DefaultTransport}}
	debugCtx := context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_verifier", pkceVerifier),
	}

	if provider == "microsoft" {
		opts = append(opts, oauth2.SetAuthURLParam("scope", strings.Join(conf.Scopes, " ")))
		log.Printf("[DEBUG] Adding required 'scope' parameter for Microsoft: %s", strings.Join(conf.Scopes, " "))
	}

	log.Printf("[DEBUG] Calling conf.Exchange with all required parameters...")
	token, err := conf.Exchange(debugCtx, code, opts...) // 使用 debugCtx
	if err != nil {
		log.Printf("用code交换token失败: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=token_exchange_failed&details="+url.QueryEscape(err.Error()))
		return
	}

	// 多余的代码块已被删除

	log.Printf("[DEBUG] Token交换成功, Expiry: %v", token.Expiry)

	// 3. 获取用户信息
	client := conf.Client(ctx, token) // 这里用回原始的ctx即可
	var userInfoURL string
	switch provider {
	case "google":
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "microsoft":
		userInfoURL = "https://graph.microsoft.com/v1.0/me"
	default:
		c.Redirect(http.StatusTemporaryRedirect, "/?error=unsupported_provider")
		return
	}

	// ... 后续代码完全不变 ...
	resp, err := client.Get(userInfoURL)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=user_info_failed")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("获取用户信息失败: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
		c.Redirect(http.StatusTemporaryRedirect, "/?error=user_info_failed")
		return
	}

	var userInfo struct {
		Email             string `json:"email"`
		UserPrincipalName string `json:"userPrincipalName,omitempty"`
		Mail              string `json:"mail,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Printf("解析用户信息失败: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=user_info_parsing_failed")
		return
	}

	email := userInfo.Email
	if provider == "microsoft" && email == "" {
		if userInfo.Mail != "" {
			email = userInfo.Mail
		} else {
			email = userInfo.UserPrincipalName
		}
	}
	if email == "" {
		log.Printf("无法从provider获取邮箱信息")
		c.Redirect(http.StatusTemporaryRedirect, "/?error=email_not_provided")
		return
	}
	log.Printf("[DEBUG] 获取到用户信息: Email=%s", email)

	userID, _ := c.Get("user_id")
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/?error=user_not_found")
		return
	}
	var oauthProvider models.OAuthProvider
	if err := database.DB.Where("name = ?", provider).First(&oauthProvider).Error; err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/?error=provider_not_configured")
		return
	}
	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", stateInfo.AccountID, user.ID).First(&emailAccount).Error; err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/?error=email_account_not_found")
		return
	}
	if !strings.EqualFold(email, emailAccount.EmailAddress) {
		log.Printf("OAuth email mismatch: token email (%s) does not match account email (%s)", email, emailAccount.EmailAddress)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=email_mismatch")
		return
	}

	encryptedAccessToken, err := utils.Encrypt([]byte(token.AccessToken))
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/?error=internal_error_encrypt_at")
		return
	}
	var encryptedRefreshToken string
	if token.RefreshToken != "" {
		encryptedRefreshToken, err = utils.Encrypt([]byte(token.RefreshToken))
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/?error=internal_error_encrypt_rt")
			return
		}
	}

	userOAuthToken := models.UserOAuthToken{
		UserID:         user.ID,
		EmailAccountID: emailAccount.ID,
		ProviderID:     oauthProvider.ID,
	}
	err = database.DB.Where(&userOAuthToken).Assign(models.UserOAuthToken{
		AccessTokenEncrypted:  encryptedAccessToken,
		RefreshTokenEncrypted: encryptedRefreshToken,
		TokenType:             token.TokenType,
		Expiry:                token.Expiry,
	}).FirstOrCreate(&userOAuthToken).Error
	if err != nil {
		log.Printf("保存OAuth token失败: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, "/?error=database_error")
		return
	}

	log.Printf("用户 %s 的 %s 账号已成功关联", user.Email, provider)
	c.Redirect(http.StatusTemporaryRedirect, "/email-accounts?status=success")
}

// GetDBStateStats 获取数据库中OAuth2 state的统计信息
func GetDBStateStats(c *gin.Context) {
	var totalCount int64
	var expiredCount int64

	// 统计总数
	if err := database.DB.Model(&models.OAuth2State{}).Count(&totalCount).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "无法统计总数")
		return
	}

	// 统计已过期的数量
	if err := database.DB.Model(&models.OAuth2State{}).Where("expires_at < ?", time.Now()).Count(&expiredCount).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "无法统计过期数量")
		return
	}

	stats := gin.H{
		"source":         "database",
		"total_states":   totalCount,
		"expired_states": expiredCount,
		"active_states":  totalCount - expiredCount,
	}

	utils.SendSuccessResponse(c, stats)
}

// --- 在文件末尾添加这个辅助类型和方法，用于打印请求 ---
type loggingTransport struct {
	T http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 打印请求细节
	reqDump, dumpErr := httputil.DumpRequestOut(req, true)
	if dumpErr != nil {
		log.Printf("Error dumping request: %v", dumpErr)
	} else {
		log.Printf("\n--- OAUTH2 TOKEN REQUEST ---\n%s\n--------------------------", string(reqDump))
	}

	// 执行原始请求
	resp, roundTripErr := t.T.RoundTrip(req)
	if roundTripErr != nil {
		return nil, roundTripErr
	}

	// 打印响应细节
	respDump, dumpErr := httputil.DumpResponse(resp, true)
	if dumpErr != nil {
		log.Printf("Error dumping response: %v", dumpErr)
	} else {
		log.Printf("\n--- OAUTH2 TOKEN RESPONSE ---\n%s\n---------------------------", string(respDump))
	}

	return resp, nil
}
