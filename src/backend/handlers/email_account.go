package handlers

import (
	"email_server/database"
	"email_server/models"
	"email_server/utils"
	"net/http"
	"strconv"
	"strings" // 新增导入

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateEmailAccount godoc
// @Summary 创建邮箱账户
// @Description 为当前登录用户创建一个新的邮箱账户
// @Tags EmailAccounts
// @Accept json
// @Produce json
// @Param emailAccount body models.EmailAccount true "邮箱账户信息，ID、UserID、CreatedAt、UpdatedAt、DeletedAt 会被忽略"
// @Success 201 {object} models.SuccessResponse{data=models.EmailAccountResponse} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /email-accounts [post]
// @Security BearerAuth
func CreateEmailAccount(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	userID, ok := userIDRaw.(int64) // Assert to int64 first
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "用户ID类型错误")
		return
	}
	actualUserID := uint(userID) // Convert to uint

	var input struct {
		EmailAddress string `json:"email_address" binding:"required,email"`
		Password     string `json:"password" binding:"omitempty,min=6"` // 密码可选
		// Provider     string `json:"provider"` // Provider 将从 EmailAddress 自动提取
		Notes        string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	var hashedPassword string
	var err error
	if input.Password != "" {
		hashedPassword, err = utils.HashPassword(input.Password)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "密码加密失败: "+err.Error())
			return
		}
	}

	// 从 EmailAddress 提取 Provider
	var provider string
	parts := strings.Split(input.EmailAddress, "@")
	if len(parts) == 2 {
		provider = parts[1]
	}

	emailAccount := models.EmailAccount{
		UserID:            actualUserID,
		EmailAddress:      input.EmailAddress,
		PasswordEncrypted: hashedPassword,
		Provider:          provider, // 使用提取的 Provider
		Notes:             input.Notes,
	}

	if err := database.DB.Create(&emailAccount).Error; err != nil {
		// 检查是否是唯一约束冲突 (例如邮箱地址已存在)
		// GORM 对于 SQLite 的唯一约束错误可能不会返回特定的错误类型，需要更通用的错误检查
		// 或者在模型层面使用 gorm:"uniqueIndex" 并依赖数据库返回错误
		// 对于更复杂的错误处理，可能需要检查 err.Error() 的内容
		utils.SendErrorResponse(c, http.StatusInternalServerError, "创建邮箱账户失败: "+err.Error())
		return
	}

	utils.SendCreatedResponse(c, emailAccount.ToEmailAccountResponse())
}

// GetEmailAccounts godoc
// @Summary 获取当前用户的所有邮箱账户
// @Description 获取当前登录用户的所有邮箱账户，支持分页
// @Tags EmailAccounts
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} models.SuccessResponse{data=[]models.EmailAccountResponse,meta=models.PaginationMeta} "获取成功"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /email-accounts [get]
// @Security BearerAuth
func GetEmailAccounts(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "用户ID类型错误")
		return
	}
	actualUserID := uint(userID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 { // Max page size limit
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	var emailAccounts []models.EmailAccount
	var totalRecords int64

	// Count total records for pagination meta
	if err := database.DB.Model(&models.EmailAccount{}).Where("user_id = ?", actualUserID).Count(&totalRecords).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取邮箱账户总数失败: "+err.Error())
		return
	}

	// Fetch paginated records
	if err := database.DB.Where("user_id = ?", actualUserID).Order("created_at desc").Offset(offset).Limit(pageSize).Find(&emailAccounts).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取邮箱账户列表失败: "+err.Error())
		return
	}

	var responses []models.EmailAccountResponse
	for _, ea := range emailAccounts {
		var platformCount int64
		// 计算关联的平台数量
		// 这里假设 PlatformRegistration 模型中有一个 EmailAccountID 字段关联到 EmailAccount 的 ID
		if err := database.DB.Model(&models.PlatformRegistration{}).Where("email_account_id = ?", ea.ID).Count(&platformCount).Error; err != nil {
			// 如果查询数量失败，可以记录错误，但为了不中断整个列表的返回，这里可以给一个默认值或忽略错误
			// log.Printf("Failed to count platforms for email account %d: %v", ea.ID, err)
			platformCount = 0 // 或者根据业务需求处理
		}
		
		response := ea.ToEmailAccountResponse()
		response.PlatformCount = platformCount
		responses = append(responses, response)
	}

	pagination := utils.CreatePaginationMeta(page, pageSize, int(totalRecords))
	utils.SendSuccessResponseWithMeta(c, responses, pagination)
}

// GetEmailAccountByID godoc
// @Summary 获取指定ID的邮箱账户详情
// @Description 获取当前登录用户拥有的指定ID的邮箱账户详情
// @Tags EmailAccounts
// @Produce json
// @Param id path int true "邮箱账户ID"
// @Success 200 {object} models.SuccessResponse{data=models.EmailAccountResponse} "获取成功"
// @Failure 400 {object} models.ErrorResponse "无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "邮箱账户未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /email-accounts/{id} [get]
// @Security BearerAuth
func GetEmailAccountByID(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "用户ID类型错误")
		return
	}
	actualUserID := uint(userID)

	emailAccountID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的邮箱账户ID格式")
		return
	}

	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", emailAccountID, actualUserID).First(&emailAccount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "邮箱账户未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取邮箱账户详情失败: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, emailAccount.ToEmailAccountResponse())
}

// UpdateEmailAccount godoc
// @Summary 更新指定ID的邮箱账户
// @Description 更新当前登录用户拥有的指定ID的邮箱账户信息
// @Tags EmailAccounts
// @Accept json
// @Produce json
// @Param id path int true "邮箱账户ID"
// @Param emailAccount body models.EmailAccount true "要更新的邮箱账户信息，ID、UserID、CreatedAt、UpdatedAt、DeletedAt 会被忽略。密码字段可选，如果提供则更新。"
// @Success 200 {object} models.SuccessResponse{data=models.EmailAccountResponse} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误或无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "邮箱账户未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /email-accounts/{id} [put]
// @Security BearerAuth
func UpdateEmailAccount(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "用户ID类型错误")
		return
	}
	actualUserID := uint(userID)

	emailAccountID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的邮箱账户ID格式")
		return
	}

	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", emailAccountID, actualUserID).First(&emailAccount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "邮箱账户未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询待更新邮箱账户失败: "+err.Error())
		return
	}

	var input struct {
		EmailAddress string `json:"email_address" binding:"omitempty,email"` // omitempty 允许部分更新
		Password     string `json:"password" binding:"omitempty,min=6"`     // 密码可选
		// Provider     string `json:"provider"` // Provider 将从 EmailAddress 自动提取
		Notes        string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// 更新字段
	if input.EmailAddress != "" {
		emailAccount.EmailAddress = input.EmailAddress
		// 如果 EmailAddress 更新了，也需要更新 Provider
		parts := strings.Split(input.EmailAddress, "@")
		if len(parts) == 2 {
			emailAccount.Provider = parts[1]
		} else {
			emailAccount.Provider = "" // 或者保持旧值，根据业务需求
		}
	}
	if input.Password != "" {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "密码加密失败: "+err.Error())
			return
		}
		emailAccount.PasswordEncrypted = hashedPassword
	}
	// Notes 总是更新，即使是空字符串，允许用户清空这些字段
	// Provider 的更新已在 EmailAddress 更新时处理
	emailAccount.Notes = input.Notes


	if err := database.DB.Save(&emailAccount).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "更新邮箱账户失败: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, emailAccount.ToEmailAccountResponse())
}

// DeleteEmailAccount godoc
// @Summary 删除指定ID的邮箱账户
// @Description 删除当前登录用户拥有的指定ID的邮箱账户
// @Tags EmailAccounts
// @Produce json
// @Param id path int true "邮箱账户ID"
// @Success 200 {object} models.SuccessResponse "删除成功"
// @Failure 400 {object} models.ErrorResponse "无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "邮箱账户未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /email-accounts/{id} [delete]
// @Security BearerAuth
func DeleteEmailAccount(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}
	userID, ok := userIDRaw.(int64)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "用户ID类型错误")
		return
	}
	actualUserID := uint(userID)

	emailAccountID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的邮箱账户ID格式")
		return
	}

	// 检查记录是否存在并且属于该用户
	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", emailAccountID, actualUserID).First(&emailAccount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "邮箱账户未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询待删除邮箱账户失败: "+err.Error())
		return
	}
	
	// 使用事务确保操作的原子性
	tx := database.DB.Begin()
	if tx.Error != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "开启事务失败: "+tx.Error.Error())
		return
	}

	// 1. 查找并软删除关联的 PlatformRegistrations 及其下的 ServiceSubscriptions
	var registrations []models.PlatformRegistration
	if err := tx.Where("email_account_id = ?", emailAccount.ID).Find(&registrations).Error; err != nil {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查找关联的平台注册信息失败: "+err.Error())
		return
	}

	for _, reg := range registrations {
		// 1a. 软删除关联的 ServiceSubscriptions
		if err := tx.Where("platform_registration_id = ?", reg.ID).Delete(&models.ServiceSubscription{}).Error; err != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "软删除服务订阅失败: "+err.Error())
			return
		}
		// 1b. 软删除 PlatformRegistration
		if err := tx.Delete(&reg).Error; err != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "软删除平台注册信息失败: "+err.Error())
			return
		}
	}

	// 2. 软删除 EmailAccount
	if err := tx.Delete(&emailAccount).Error; err != nil {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "删除邮箱账户失败: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "邮箱账户及关联信息删除成功"})
}