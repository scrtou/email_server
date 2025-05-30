package handlers

import (
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
		utils.SendErrorResponse(c, 400, "参数错误: "+err.Error())
		return
	}

	// 验证密码强度
	if err := utils.ValidatePassword(req.Password); err != nil {
		utils.SendErrorResponse(c, 400, err.Error())
		return
	}

	// 检查用户名和邮箱是否已存在 (Using GORM)
	var existingUser models.User
	// Check for username
	// Use a new variable for error to avoid conflict with the outer scope 'err' if it exists or is reused.
	dbErr := database.DB.Where("username = ?", req.Username).First(&existingUser).Error
	if dbErr == nil { // User found with this username
		utils.SendErrorResponse(c, 400, "用户名已存在")
		return
	}
	// Important: Check if the error is specifically "record not found"
	if dbErr != nil && dbErr.Error() != "record not found" { // An actual DB error occurred
		log.Printf("查询用户失败 (username): %v", dbErr)
		utils.SendErrorResponse(c, 500, "系统错误")
		return
	}

	// Check for email
	dbErr = database.DB.Where("email = ?", req.Email).First(&existingUser).Error
	if dbErr == nil { // User found with this email
		utils.SendErrorResponse(c, 400, "邮箱已存在")
		return
	}
	if dbErr != nil && dbErr.Error() != "record not found" { // An actual DB error occurred
		log.Printf("查询用户失败 (email): %v", dbErr)
		utils.SendErrorResponse(c, 500, "系统错误")
		return
	}

	// 密码加密
	hashedPassword, err := utils.HashPassword(req.Password) // err from HashPassword
	if err != nil {
		log.Printf("密码加密失败: %v", err)
		utils.SendErrorResponse(c, 500, "系统错误")
		return
	}

	// 创建用户 (Using GORM, only core fields)
	newUser := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	// Use a new variable for GORM operation result
	createResult := database.DB.Create(&newUser) // newUser will be updated with ID, CreatedAt, UpdatedAt by GORM
	if createResult.Error != nil {
		log.Printf("创建用户失败: %v", createResult.Error)
		utils.SendErrorResponse(c, 500, "创建用户失败")
		return
	}

	// 生成token
	// newUser.ID is uint, GenerateToken might expect int64.
	token, err := utils.GenerateToken(int64(newUser.ID), newUser.Username, "user") // Assuming default role "user"
	if err != nil {
		log.Printf("生成token失败: %v", err)
		utils.SendErrorResponse(c, 500, "系统错误")
		return
	}
	
	// Construct user for response, ensuring it matches the LoginResponse.User type and omits sensitive data.
	// The User model itself has json:"-" for Password.
	// newUser already contains GORM-populated ID, CreatedAt, UpdatedAt.
	responseUser := &models.User{
		Model:    newUser.Model, // This includes ID, CreatedAt, UpdatedAt, DeletedAt
		Username: newUser.Username,
		Email:    newUser.Email,
		// Password is not included due to json:"-" in the model or by not explicitly setting it here.
	}

	response := models.LoginResponse{
		Token:     token,
		ExpiresIn: config.AppConfig.JWT.ExpiresIn, // Make sure config.AppConfig is properly loaded and accessible
		User:      responseUser,
	}

	utils.SendSuccessResponse(c, response)
}

// Login 用户登录
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, 400, "参数错误")
		return
	}

	var user models.User
	// 查询用户 (Using GORM)
	// Assuming 'status = 1' means active user. If User model has no Status field, this condition might change or be removed.
	// For now, sticking to core User model which doesn't have Status. If status is a business rule, it should be handled.
	// Let's assume for login, we only check username. Active status check can be added if User model supports it.
	dbErr := database.DB.Where("username = ?", req.Username).First(&user).Error

	if dbErr != nil {
		if dbErr.Error() == "record not found" {
			utils.SendErrorResponse(c, 401, "用户名或密码错误")
		} else {
			log.Printf("查询用户失败: %v", dbErr)
			utils.SendErrorResponse(c, 500, "系统错误")
		}
		return
	}

	// 验证密码
	// user.Password from DB is the hashed password
	if !utils.CheckPassword(req.Password, user.Password) {
		utils.SendErrorResponse(c, 401, "用户名或密码错误")
		return
	}

	// 更新最后登录时间 - This was part of the original logic.
	// The core User model from optimization_proposal.md does not have LastLogin.
	// If this is required, the User model needs to be extended, or this logic removed/rethought.
	// For now, I will comment it out to align with the strict core User model.
	/*
	now := time.Now()
	// GORM update: database.DB.Model(&user).Update("last_login", now)
	// However, last_login is not in the core User model.
	// If we were to update it, it would be:
	// database.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("last_login", now)
	// user.LastLogin = &now // This would also require LastLogin field in models.User
	*/

	// 生成token
	// user.Role is not in the core User model. Assuming "user" role for token generation.
	token, err := utils.GenerateToken(int64(user.ID), user.Username, "user") // Use user.ID from GORM
	if err != nil {
		log.Printf("生成token失败: %v", err)
		utils.SendErrorResponse(c, 500, "系统错误")
		return
	}

	// Prepare user data for response, adhering to the core User model.
	responseUser := &models.User{
		Model:    user.Model, // Includes ID, CreatedAt, UpdatedAt
		Username: user.Username,
		Email:    user.Email,
		// Password is not included (json:"-")
	}

	response := models.LoginResponse{
		Token:     token,
		ExpiresIn: config.AppConfig.JWT.ExpiresIn,
		User:      responseUser,
	}

	utils.SendSuccessResponse(c, response)
}

// GetProfile 获取当前用户信息 ( entspricht /users/me )
func GetProfile(c *gin.Context) {
	userIDContext, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, 401, "用户未认证") // Should be caught by AuthRequired middleware anyway
		return
	}

	userID, ok := userIDContext.(int64)
	if !ok {
		log.Printf("user_id in context is not int64: %T", userIDContext)
		utils.SendErrorResponse(c, 500, "服务器内部错误")
		return
	}

	var user models.User
	// 查询用户 (Using GORM)
	// The core User model (gorm.Model) has ID, CreatedAt, UpdatedAt.
	// Username, Email, Password (hashed) are other fields.
	// Password should not be sent in response.
	dbErr := database.DB.First(&user, userID).Error // GORM uses primary key by default with First

	if dbErr != nil {
		if dbErr.Error() == "record not found" { // gorm.ErrRecordNotFound
			utils.SendErrorResponse(c, 404, "用户不存在")
		} else {
			log.Printf("查询用户信息失败: %v", dbErr)
			utils.SendErrorResponse(c, 500, "系统错误")
		}
		return
	}

	// Prepare user data for response, adhering to the core User model.
	// The User model itself has json:"-" for Password.
	// user variable already contains GORM-populated ID, CreatedAt, UpdatedAt, Username, Email.
	
	// If UserResponse struct is preferred and aligned with core model:
	// utils.Success(c, user.ToResponse())
	// For now, directly returning the core user model fields (GORM model + Username, Email)
	// The User struct itself will be marshalled to JSON, respecting `json:"-"` for Password.
	
	responseUser := models.UserResponse{
		ID:        user.ID, // user.ID is uint, UserResponse.ID is uint
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	utils.SendSuccessResponse(c, responseUser)
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
		utils.SendErrorResponse(c, 400, "参数错误")
		return
	}

	// 检查邮箱是否被其他用户使用
	var count int64
	// Assuming userID from context is int64, convert to uint for GORM query if user.ID is uint
	userIDUint, ok := userID.(uint)
	if !ok {
		// If userID from context was int64, try to convert.
		// This depends on how user_id is set in the context.
		// For now, assuming it's compatible or needs conversion.
		// Let's assume it's already uint for simplicity with GORM's user.ID.
		// If it was int64, it should be: userIDUint = uint(userID.(int64))
		// For now, let's assume userID is already uint or compatible.
		// If user_id in context is int64, and models.User.ID is uint:
		idFromContext, idOk := userID.(int64)
		if !idOk {
			log.Printf("User ID in context is not of expected type: %T", userID)
			utils.SendErrorResponse(c, 500, "服务器内部错误")
			return
		}
		userIDUint = uint(idFromContext)
	}


	err := database.DB.Model(&models.User{}).Where("email = ? AND id != ?", req.Email, userIDUint).Count(&count).Error
	if err != nil {
		log.Printf("检查邮箱失败: %v", err)
		utils.SendErrorResponse(c, 500, "系统错误")
		return
	}
	if count > 0 {
		utils.SendErrorResponse(c, 400, "邮箱已被其他用户使用")
		return
	}

	// 更新用户信息 - GORM handles nil for empty strings if the field is a pointer or nullable type.
	// For string fields, GORM will save an empty string.
	// The User model does not have RealName or Phone in the core definition.
	// This UpdateProfile function needs to align with the User model.
	// If RealName and Phone are to be updated, they must be part of models.User.
	// Based on models/user.go, User only has Username, Email, Password.
	// This function seems to be updating fields (real_name, phone) not in the current User model.
	// This will likely fail or do nothing for those fields.
	// For now, I will proceed assuming the User model *might* be extended later or these are custom fields.
	// However, the strict instruction is to adhere to the simplified model.
	// Let's assume this function is intended to update Email, and other fields are ignored if not in model.
	// Or, if RealName and Phone were meant to be part of User, the model definition is the source of truth.
	// Given the error messages are about DB operations, I'll fix the GORM syntax.
	// The fields `real_name` and `phone` are not in the `models.User` struct.
	// GORM will ignore these fields during update if they are not in the struct.
	// We should only update fields that exist in `models.User`.
	// The request includes RealName, Phone, Email.
	// The core User model only has Email that can be updated here (Username is unique, Password has its own flow).

	updateData := map[string]interface{}{
		"email": req.Email,
		// "real_name": req.RealName, // Not in core User model
		// "phone": req.Phone,       // Not in core User model
		"updated_at": time.Now(),
	}
	// If req.Phone is empty and phone field was nullable string (e.g. *string),
	// then `phoneValue = nil` would be appropriate.
	// Since User.Phone is not in the model, this part is moot for now.

	result := database.DB.Model(&models.User{}).Where("id = ?", userIDUint).Updates(updateData)
	if result.Error != nil {
		log.Printf("更新用户信息失败: %v", result.Error)
		utils.SendErrorResponse(c, 500, "更新失败")
		return
	}
	if result.RowsAffected == 0 {
		// This could mean user not found, or data was the same.
		// For simplicity, not treating as error, but could be a 404 if ID must exist.
		log.Printf("更新用户信息时，没有行受到影响 (ID: %d)", userIDUint)
	}

	utils.SendSuccessResponse(c, "更新成功")
}

// ChangePassword 修改密码
func ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, 400, "参数错误")
		return
	}

	// 验证新密码强度
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		utils.SendErrorResponse(c, 400, err.Error())
		return
	}

	// 查询当前密码
	var user models.User
	// Assuming userID from context is int64, convert to uint for GORM query if user.ID is uint
	userIDUint, ok := userID.(uint)
	if !ok {
		idFromContext, idOk := userID.(int64)
		if !idOk {
			log.Printf("User ID in context is not of expected type: %T", userID)
			utils.SendErrorResponse(c, 500, "服务器内部错误")
			return
		}
		userIDUint = uint(idFromContext)
	}

	err := database.DB.Model(&models.User{}).Select("password").Where("id = ?", userIDUint).First(&user).Error
	if err != nil {
		if err.Error() == "record not found" { // gorm.ErrRecordNotFound
			utils.SendErrorResponse(c, 404, "用户不存在")
		} else {
			log.Printf("查询用户密码失败: %v", err)
			utils.SendErrorResponse(c, 500, "系统错误")
		}
		return
	}
	currentPassword := user.Password

	// 验证旧密码
	if !utils.CheckPassword(req.OldPassword, currentPassword) {
		utils.SendErrorResponse(c, 400, "原密码错误")
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("密码加密失败: %v", err)
		utils.SendErrorResponse(c, 500, "系统错误")
		return
	}

	// 更新密码
	// userIDUint is already defined and converted above
	result := database.DB.Model(&models.User{}).Where("id = ?", userIDUint).Update("password", hashedPassword)
	if result.Error != nil {
		log.Printf("更新密码失败: %v", result.Error)
		utils.SendErrorResponse(c, 500, "修改密码失败")
		return
	}
	if result.RowsAffected == 0 {
		log.Printf("更新密码时，没有行受到影响 (ID: %d)", userIDUint)
		// This could be an issue if the user must exist.
		// For now, not treating as a fatal error if GORM itself doesn't error.
	}
	utils.SendSuccessResponse(c, "密码修改成功")
}

// RefreshToken 刷新token
func RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.SendErrorResponse(c, 401, "缺少认证token")
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		utils.SendErrorResponse(c, 401, "认证格式错误")
		return
	}

	newToken, err := utils.RefreshToken(parts[1])
	if err != nil {
		utils.SendErrorResponse(c, 401, "token刷新失败")
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"token":      newToken,
		"expires_in": config.AppConfig.JWT.ExpiresIn,
	})
}

// Logout 登出（前端删除token即可，这里做日志记录）
func Logout(c *gin.Context) {
	username, _ := c.Get("username")
	log.Printf("用户 %s 登出", username)
	utils.SendSuccessResponse(c, "登出成功")
}
