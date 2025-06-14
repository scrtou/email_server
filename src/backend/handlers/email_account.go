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
		Password     string `json:"password" binding:"omitempty,min=6"`
		IMAPServer   string `json:"imap_server"`
		IMAPPort     *int   `json:"imap_port"`
		Notes        string `json:"notes"`
		PhoneNumber  string `json:"phone_number"`
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
	provider := utils.ExtractProviderFromEmail(input.EmailAddress)

	// 检查是否存在同邮箱地址的记录
	var existingEmailAccount models.EmailAccount
	err = database.DB.Where("user_id = ? AND email_address = ?", actualUserID, input.EmailAddress).First(&existingEmailAccount).Error

	if err == nil {
		// 记录存在
		utils.SendErrorResponse(c, http.StatusConflict, "该邮箱地址已被注册")
		return
	} else if err != gorm.ErrRecordNotFound {
		// 查询出错
		utils.SendErrorResponse(c, http.StatusInternalServerError, "检查邮箱账户是否存在失败: "+err.Error())
		return
	}

	// 没有找到同邮箱地址记录，创建新邮箱账户
	emailAccount := models.EmailAccount{
		UserID:            actualUserID,
		EmailAddress:      input.EmailAddress,
		PasswordEncrypted: hashedPassword,
		Provider:          provider,
		IMAPServer:        input.IMAPServer,
		Notes:             input.Notes,
		PhoneNumber:       input.PhoneNumber,
	}

	// Only set port if it's not nil
	if input.IMAPPort != nil {
		emailAccount.IMAPPort = *input.IMAPPort
	}

	if err := database.DB.Create(&emailAccount).Error; err != nil {
		if utils.IsUniqueConstraintError(err) {
			utils.SendErrorResponse(c, http.StatusConflict, "该邮箱地址已被注册")
			return
		}
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
// @Param orderBy query string false "排序字段 (e.g., email_address, created_at, updated_at)" default(created_at)
// @Param sortDirection query string false "排序方向 (asc, desc)" default(desc)
// @Param provider query string false "按服务商名称进行模糊匹配筛选"
// @Param email_address query string false "按邮箱地址进行模糊匹配筛选"
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
	// pageSize 的获取和转换将在后面根据具体逻辑处理
	rawPageSize := c.Query("pageSize")
	orderBy := c.DefaultQuery("orderBy", "created_at")
	sortDirection := c.DefaultQuery("sortDirection", "desc")
	filterProvider := strings.ToLower(strings.TrimSpace(c.Query("provider")))
	filterEmailAddress := strings.ToLower(strings.TrimSpace(c.Query("email_address")))

	// Validate orderBy parameter to prevent SQL injection
	allowedOrderByFields := map[string]string{
		"email_address": "email_address",
		"provider":      "provider",
		"notes":         "notes",
		"created_at":    "created_at",
		"updated_at":    "updated_at",
		"phone_number":  "phone_number",
	}
	dbOrderByField, isValidField := allowedOrderByFields[orderBy]
	if !isValidField {
		dbOrderByField = "created_at" // Default to a safe field
	}

	// Validate sortDirection
	if strings.ToLower(sortDirection) != "asc" && strings.ToLower(sortDirection) != "desc" {
		sortDirection = "desc" // Default to desc
	}

	orderClause := dbOrderByField + " " + sortDirection

	if page <= 0 {
		page = 1
	}
	var offset int
	var limitApplied bool
	var pageSize int

	// pageSize 参数处理逻辑调整
	if rawPageSize == "" {
		// 前端未传递 pageSize 参数，设置为默认值
		pageSize = 10
		limitApplied = true
	} else {
		parsedPageSize, err := strconv.Atoi(rawPageSize)
		if err != nil {
			// 解析错误，可以视为无效输入，按默认处理或返回错误
			// 这里我们按默认值处理
			pageSize = 10
			limitApplied = true
		} else {
			if parsedPageSize > 0 {
				// 前端传递的 pageSize > 0，直接使用
				pageSize = parsedPageSize
				limitApplied = true
			} else {
				// 前端传递的 pageSize <= 0，获取所有记录
				pageSize = 0 // 稍后会用 totalRecords 更新
				limitApplied = false
			}
		}
	}

	if limitApplied {
		offset = (page - 1) * pageSize
	} else {
		offset = 0 // 获取所有记录时，offset 为 0
	}

	var emailAccounts []models.EmailAccount
	var totalRecords int64

	// Count total records for pagination meta
	query := database.DB.Model(&models.EmailAccount{}).Where("user_id = ?", actualUserID)

	// Apply provider filter if provided
	if filterProvider != "" {
		query = query.Where("LOWER(provider) LIKE ?", "%"+filterProvider+"%")
	}
	if filterEmailAddress != "" {
		query = query.Where("LOWER(email_address) LIKE ?", "%"+filterEmailAddress+"%")
	}

	countQuery := query // Create a new query builder for count to avoid issues with Order, Limit, Offset
	if err := countQuery.Count(&totalRecords).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取邮箱账户总数失败: "+err.Error())
		return
	}

	// 如果 pageSize <= 0 (即 limitApplied 为 false)，则 pageSize 更新为 totalRecords
	if !limitApplied {
		pageSize = int(totalRecords) // 获取所有记录时，pageSize 为总记录数
		// 如果 totalRecords 为 0，pageSize 也为 0，这是合理的
	}

	// Fetch records
	dbQuery := query.Order(orderClause)
	if limitApplied { // 只有当 limitApplied 为 true 时才应用 Offset 和 Limit
		dbQuery = dbQuery.Offset(offset).Limit(pageSize)
	}
	// 如果 limitApplied 为 false，则不应用 Limit 和 Offset，获取所有记录

	if err := dbQuery.Find(&emailAccounts).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取邮箱账户列表失败: "+err.Error())
		return
	}

	var responses []models.EmailAccountResponse
	for _, ea := range emailAccounts {
		var platformCount int64
		// 计算关联的平台数量
		if err := database.DB.Model(&models.PlatformRegistration{}).Where("email_account_id = ?", ea.ID).Count(&platformCount).Error; err != nil {
			platformCount = 0
		}

		response := ea.ToEmailAccountResponse()
		response.PlatformCount = platformCount

		// Check if an OAuth token exists for this account
		var token models.UserOAuthToken
		err := database.DB.Where("email_account_id = ?", ea.ID).First(&token).Error
		response.IsOAuthConnected = err == nil

		responses = append(responses, response)
	}

	// API响应的 meta 数据中返回实际使用的 pageSize 值
	// finalPageSizeForMeta 现在就是 pageSize，因为它已经被正确设置
	var finalPageSizeForMeta int = pageSize
	// 如果 pageSize 为 0 (例如 totalRecords 为 0 且之前 pageSize <= 0),
	// CreatePaginationMeta 应该能处理这种情况，或者我们在这里确保它至少为1（如果需要分页）。
	// 但如果 totalRecords 为 0，那么 pageSize 为 0 是合理的，表示没有数据，每页0条。
	// 如果 totalRecords > 0 但 pageSize 仍然是0（这不应该发生，因为如果 limitApplied=false, pageSize=totalRecords），
	// 那么 CreatePaginationMeta 可能会出问题。
	// 确保 finalPageSizeForMeta 在 totalRecords > 0 时至少为 1，除非我们确实想表示“所有”（此时 pageSize = totalRecords）。
	if totalRecords > 0 && finalPageSizeForMeta == 0 {
		// 这种情况理论上不应该发生，因为如果 limitApplied 为 false，pageSize 会被设为 totalRecords。
		// 如果 limitApplied 为 true，pageSize > 0。
		// 但作为安全措施，如果 totalRecords > 0 而 pageSize 仍然是0，则设为 totalRecords。
		finalPageSizeForMeta = int(totalRecords)
	} else if totalRecords == 0 && finalPageSizeForMeta == 0 {
		// 如果没有记录，pageSize 为 0 是可以接受的。
		// 但 CreatePaginationMeta 可能期望 pageSize 至少为 1 来计算 totalPages。
		// 如果 CreatePaginationMeta 能够处理 pageSize 为 0 的情况，则无需更改。
		// 假设 CreatePaginationMeta 在 pageSize 为 0 时，totalPages 也为 0 或 1（取决于实现）。
		// 为了安全，如果 pageSize 为0，且有记录，则设为 totalRecords。如果没记录，pageSize 为0。
		// 如果 CreatePaginationMeta 要求 pageSize >= 1，那么即使 totalRecords = 0，也应设为1。
		// 暂时维持 pageSize 可能为0的情况，依赖 CreatePaginationMeta 的处理。
		// 或者，如果 CreatePaginationMeta 要求 pageSize >= 1:
		// if finalPageSizeForMeta == 0 { finalPageSizeForMeta = 1 }
	}

	pagination := utils.CreatePaginationMeta(page, finalPageSizeForMeta, int(totalRecords))
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
		EmailAddress string `json:"email_address" binding:"omitempty,email"`
		Password     string `json:"password" binding:"omitempty,min=6"`
		IMAPServer   string `json:"imap_server"`
		IMAPPort     *int   `json:"imap_port"`
		Notes        string `json:"notes"`
		PhoneNumber  string `json:"phone_number"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// 更新字段
	if input.EmailAddress != "" {
		emailAccount.EmailAddress = input.EmailAddress
		// 如果 EmailAddress 更新了，也需要更新 Provider
		emailAccount.Provider = utils.ExtractProviderFromEmail(input.EmailAddress)
	}
	if input.Password != "" {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "密码加密失败: "+err.Error())
			return
		}
		emailAccount.PasswordEncrypted = hashedPassword
	}
	// 如果 Password 和 ConfirmPassword 均未提供，则不更新密码
	// Notes 总是更新，即使是空字符串，允许用户清空这些字段
	// Provider 的更新已在 EmailAddress 更新时处理
	emailAccount.Notes = input.Notes
	emailAccount.PhoneNumber = input.PhoneNumber
	emailAccount.IMAPServer = input.IMAPServer
	if input.IMAPPort != nil {
		emailAccount.IMAPPort = *input.IMAPPort
	} else {
		// If the input is null, explicitly set it to 0 or another default
		emailAccount.IMAPPort = 0
	}

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

	// 1. 查找并硬删除关联的 PlatformRegistrations 及其下的 ServiceSubscriptions
	var registrations []models.PlatformRegistration
	if err := tx.Where("email_account_id = ?", emailAccount.ID).Find(&registrations).Error; err != nil {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查找关联的平台注册信息失败: "+err.Error())
		return
	}

	for _, reg := range registrations {
		// 1a. 硬删除关联的 ServiceSubscriptions
		if err := tx.Unscoped().Where("platform_registration_id = ?", reg.ID).Delete(&models.ServiceSubscription{}).Error; err != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "删除服务订阅失败: "+err.Error())
			return
		}
		// 1b. 硬删除 PlatformRegistration
		if err := tx.Unscoped().Delete(&reg).Error; err != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "删除平台注册信息失败: "+err.Error())
			return
		}
	}

	// 2. 硬删除 EmailAccount
	if err := tx.Unscoped().Delete(&emailAccount).Error; err != nil {
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

// GetEmailAccountProviders godoc
// @Summary 获取当前用户邮箱账户的所有唯一服务商列表
// @Description 获取当前登录用户的所有邮箱账户中不重复的服务商名称列表
// @Tags EmailAccounts
// @Produce json
// @Success 200 {object} models.SuccessResponse{data=[]string} "获取成功"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /email-accounts/providers [get]
// @Security BearerAuth
func GetEmailAccountProviders(c *gin.Context) {
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

	var providers []string
	if err := database.DB.Model(&models.EmailAccount{}).
		Where("user_id = ?", actualUserID).
		Distinct().Pluck("provider", &providers).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取服务商列表失败: "+err.Error())
		return
	}

	// Filter out empty or null providers if any were stored that way
	var uniqueProviders []string
	seenProviders := make(map[string]bool)
	for _, p := range providers {
		if p != "" && !seenProviders[p] {
			uniqueProviders = append(uniqueProviders, p)
			seenProviders[p] = true
		}
	}

	utils.SendSuccessResponse(c, uniqueProviders)
}

// GetEmailAccountPassword 获取邮箱账户的密码
func GetEmailAccountPassword(c *gin.Context) {
	// 获取用户ID
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

	// 获取邮箱账户ID
	emailAccountID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的邮箱账户ID格式")
		return
	}

	// 查询邮箱账户
	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", emailAccountID, actualUserID).First(&emailAccount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "邮箱账户未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询邮箱账户失败: "+err.Error())
		return
	}

	// 检查是否有密码
	if emailAccount.PasswordEncrypted == "" {
		utils.SendErrorResponse(c, http.StatusNotFound, "该邮箱账户未设置密码")
		return
	}

	// 解密密码
	var decryptedPassword string
	if utils.IsEncryptedPassword(emailAccount.PasswordEncrypted) {
		// 新的加密格式，可以解密
		decryptedPassword, err = utils.DecryptPassword(emailAccount.PasswordEncrypted)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "密码解密失败: "+err.Error())
			return
		}
	} else {
		// 旧的bcrypt格式，无法解密
		utils.SendErrorResponse(c, http.StatusBadRequest, "该密码使用旧格式存储，无法查看。请重新设置密码。")
		return
	}

	response := map[string]string{
		"password": decryptedPassword,
	}
	utils.SendSuccessResponse(c, response)
}
