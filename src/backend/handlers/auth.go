package handlers

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"email_server/config"
	"email_server/database"
	"email_server/models"
	"email_server/utils"

	"github.com/gin-gonic/gin"
)

// Register 用户注册
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, 400, "参数错误: "+err.Error())
		return
	}

	// 验证密码强度
	if err := utils.ValidatePassword(req.Password); err != nil {
		utils.SendError(c, 400, err.Error())
		return
	}

	// 检查用户名和邮箱是否已存在
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?",
		req.Username, req.Email).Scan(&count)
	if err != nil {
		log.Printf("查询用户失败: %v", err)
		utils.SendError(c, 500, "系统错误")
		return
	}
	if count > 0 {
		utils.SendError(c, 400, "用户名或邮箱已存在")
		return
	}

	// 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("密码加密失败: %v", err)
		utils.SendError(c, 500, "系统错误")
		return
	}

	// 创建用户
	query := `
        INSERT INTO users (username, email, password, real_name, phone)
        VALUES (?, ?, ?, ?, ?)
    `
	result, err := database.DB.Exec(query, req.Username, req.Email, hashedPassword,
		req.RealName, req.Phone)
	if err != nil {
		log.Printf("创建用户失败: %v", err)
		utils.SendError(c, 500, "创建用户失败")
		return
	}

	userID, _ := result.LastInsertId()

	// 生成token
	token, err := utils.GenerateToken(userID, req.Username, "user")
	if err != nil {
		log.Printf("生成token失败: %v", err)
		utils.SendError(c, 500, "系统错误")
		return
	}

	user := &models.User{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
		RealName: req.RealName,
		Phone:    req.Phone,
		Role:     "user",
		Status:   1,
	}

	response := models.LoginResponse{
		Token:     token,
		ExpiresIn: config.AppConfig.JWT.ExpiresIn,
		User:      user,
	}

	utils.Success(c, response)
}

// Login 用户登录 - 修复NULL值处理
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, 400, "参数错误")
		return
	}

	// 查询用户 - 使用sql.NullString处理可能为NULL的字段
	var user models.User
	var hashedPassword string
	var phone sql.NullString
	var lastLogin sql.NullTime

	query := `
        SELECT id, username, email, password, real_name, phone, role, status, 
               last_login, created_at, updated_at
        FROM users WHERE username = ? AND status = 1
    `
	err := database.DB.QueryRow(query, req.Username).Scan(
		&user.ID, &user.Username, &user.Email, &hashedPassword,
		&user.RealName, &phone, &user.Role, &user.Status,
		&lastLogin, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		utils.SendError(c, 401, "用户名或密码错误")
		return
	} else if err != nil {
		log.Printf("查询用户失败: %v", err)
		utils.SendError(c, 500, "系统错误")
		return
	}

	// 处理可能为NULL的字段
	if phone.Valid {
		user.Phone = phone.String
	} else {
		user.Phone = ""
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	} else {
		user.LastLogin = nil
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, hashedPassword) {
		utils.SendError(c, 401, "用户名或密码错误")
		return
	}

	// 更新最后登录时间
	now := time.Now()
	database.DB.Exec("UPDATE users SET last_login = ? WHERE id = ?", now, user.ID)
	user.LastLogin = &now

	// 生成token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		log.Printf("生成token失败: %v", err)
		utils.SendError(c, 500, "系统错误")
		return
	}

	response := models.LoginResponse{
		Token:     token,
		ExpiresIn: config.AppConfig.JWT.ExpiresIn,
		User:      &user,
	}

	utils.Success(c, response)
}

// GetProfile 获取用户信息 - 修复NULL值处理
func GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	var phone sql.NullString
	var lastLogin sql.NullTime

	query := `
        SELECT id, username, email, real_name, phone, role, status, 
               last_login, created_at, updated_at
        FROM users WHERE id = ?
    `
	err := database.DB.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.RealName,
		&phone, &user.Role, &user.Status, &lastLogin,
		&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		log.Printf("查询用户信息失败: %v", err)
		utils.SendError(c, 404, "用户不存在")
		return
	}

	// 处理可能为NULL的字段
	if phone.Valid {
		user.Phone = phone.String
	} else {
		user.Phone = ""
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	} else {
		user.LastLogin = nil
	}

	utils.Success(c, user.ToResponse())
}

// UpdateProfile 更新用户信息
func UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		RealName string `json:"real_name" binding:"required"`
		Phone    string `json:"phone"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, 400, "参数错误")
		return
	}

	// 检查邮箱是否被其他用户使用
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND id != ?",
		req.Email, userID).Scan(&count)
	if err != nil {
		log.Printf("检查邮箱失败: %v", err)
		utils.SendError(c, 500, "系统错误")
		return
	}
	if count > 0 {
		utils.SendError(c, 400, "邮箱已被其他用户使用")
		return
	}

	// 更新用户信息 - 处理空phone值
	var phoneValue interface{}
	if req.Phone == "" {
		phoneValue = nil
	} else {
		phoneValue = req.Phone
	}

	query := `
        UPDATE users SET real_name = ?, phone = ?, email = ?, updated_at = NOW()
        WHERE id = ?
    `
	_, err = database.DB.Exec(query, req.RealName, phoneValue, req.Email, userID)
	if err != nil {
		log.Printf("更新用户信息失败: %v", err)
		utils.SendError(c, 500, "更新失败")
		return
	}

	utils.Success(c, "更新成功")
}

// ChangePassword 修改密码
func ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, 400, "参数错误")
		return
	}

	// 验证新密码强度
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		utils.SendError(c, 400, err.Error())
		return
	}

	// 查询当前密码
	var currentPassword string
	err := database.DB.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&currentPassword)
	if err != nil {
		log.Printf("查询用户密码失败: %v", err)
		utils.SendError(c, 500, "系统错误")
		return
	}

	// 验证旧密码
	if !utils.CheckPassword(req.OldPassword, currentPassword) {
		utils.SendError(c, 400, "原密码错误")
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("密码加密失败: %v", err)
		utils.SendError(c, 500, "系统错误")
		return
	}

	// 更新密码
	_, err = database.DB.Exec("UPDATE users SET password = ?, updated_at = NOW() WHERE id = ?",
		hashedPassword, userID)
	if err != nil {
		log.Printf("更新密码失败: %v", err)
		utils.SendError(c, 500, "修改密码失败")
		return
	}

	utils.Success(c, "密码修改成功")
}

// RefreshToken 刷新token
func RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.SendError(c, 401, "缺少认证token")
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		utils.SendError(c, 401, "认证格式错误")
		return
	}

	newToken, err := utils.RefreshToken(parts[1])
	if err != nil {
		utils.SendError(c, 401, "token刷新失败")
		return
	}

	utils.Success(c, gin.H{
		"token":      newToken,
		"expires_in": config.AppConfig.JWT.ExpiresIn,
	})
}

// Logout 登出（前端删除token即可，这里做日志记录）
func Logout(c *gin.Context) {
	username, _ := c.Get("username")
	log.Printf("用户 %s 登出", username)
	utils.Success(c, "登出成功")
}
