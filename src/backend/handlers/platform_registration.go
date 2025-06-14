package handlers

import (
	"email_server/database"
	"email_server/models"
	"email_server/utils"
	"errors" // Added for errors.Is
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreatePlatformRegistrationInput 定义了创建平台注册信息的输入结构
type CreatePlatformRegistrationInput struct {
	EmailAddress  string `json:"email_address" binding:"omitempty,email"` // 修改为可选
	PlatformName  string `json:"platform_name" binding:"required"`
	LoginUsername string `json:"login_username"`
	LoginPassword string `json:"login_password" binding:"omitempty,min=6"` // 密码可选
	Notes         string `json:"notes"`
	PhoneNumber   string `json:"phone_number"` // 手机号码，可选
	// 可以根据需要添加 Provider (针对EmailAccount) 和 WebsiteURL (针对Platform)
	// EmailProvider    string `json:"email_provider"`
	// PlatformWebsiteURL string `json:"platform_website_url"`
}

// CreatePlatformRegistrationWithIDsInput 定义了通过ID创建平台注册信息的输入结构
type CreatePlatformRegistrationWithIDsInput struct {
	EmailAccountID uint   `json:"email_account_id"` // 移除 binding:"required"
	PlatformID     uint   `json:"platform_id" binding:"required"`
	LoginUsername  string `json:"login_username"`
	LoginPassword  string `json:"login_password" binding:"omitempty,min=6"` // 密码可选
	Notes          string `json:"notes"`
	PhoneNumber    string `json:"phone_number"` // 手机号码，可选
}

// CreatePlatformRegistrationWithIDs godoc
// @Summary 通过ID创建平台注册信息
// @Description 为当前用户创建一个新的平台注册信息，关联一个已有的邮箱账户ID和一个已有的平台ID。
// @Tags PlatformRegistrations
// @Accept json
// @Produce json
// @Param platformRegistration body handlers.CreatePlatformRegistrationWithIDsInput true "平台注册信息（包含EmailAccountID 和 PlatformID）。密码应为原始密码。"
// @Success 201 {object} models.SuccessResponse{data=models.PlatformRegistrationResponse} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误或关联资源无效"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 404 {object} models.ErrorResponse "关联的邮箱账户或平台未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /platform-registrations [post] // 这个是旧路由，现在由这个函数处理
// @Security BearerAuth
func CreatePlatformRegistrationWithIDs(c *gin.Context) {
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
	currentUserID := uint(userID)

	var input CreatePlatformRegistrationWithIDsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var err error // Declare err once

	// 仅当 EmailAccountID > 0 时，验证 EmailAccount 是否属于当前用户
	var emailAccount models.EmailAccount // 声明 emailAccount 变量，以便后续使用
	if input.EmailAccountID > 0 {
		err = tx.Where("id = ? AND user_id = ?", input.EmailAccountID, currentUserID).First(&emailAccount).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				utils.SendErrorResponse(c, http.StatusNotFound, "关联的邮箱账户未找到或不属于当前用户")
				return
			}
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询邮箱账户失败: "+err.Error())
			return
		}
	}
	// 如果 EmailAccountID 为 0，则跳过验证，emailAccount 将保持其零值

	// 验证 Platform 是否存在且属于当前用户
	var platform models.Platform
	err = tx.Where("id = ? AND user_id = ?", input.PlatformID, currentUserID).First(&platform).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "关联的平台未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台失败: "+err.Error())
		return
	}

	var encryptedPassword string
	if input.LoginPassword != "" {
		encryptedPassword, err = utils.EncryptPassword(input.LoginPassword)
		if err != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "密码加密失败: "+err.Error())
			return
		}
	}

	// --- 精确冲突检查 ---
	var existingRegistration models.PlatformRegistration
	query := tx.Where("user_id = ? AND platform_id = ?", currentUserID, input.PlatformID)

	conflictMsg := ""
	if input.LoginUsername != "" && input.EmailAccountID > 0 {
		// 检查用户名和邮箱组合
		query = query.Where("login_username = ? AND email_account_id = ?", input.LoginUsername, input.EmailAccountID)
		conflictMsg = "该用户名和邮箱组合已在此平台注册。"
	} else if input.LoginUsername != "" && input.EmailAccountID == 0 {
		// 仅检查用户名（邮箱为空或NULL）
		query = query.Where("login_username = ? AND (email_account_id = ? OR email_account_id IS NULL)", input.LoginUsername, 0)
		conflictMsg = "该用户名已在此平台注册（无关联邮箱）。"
	} else if input.LoginUsername == "" && input.EmailAccountID > 0 {
		// 仅检查邮箱（用户名为空或NULL）
		query = query.Where("(login_username = ? OR login_username IS NULL) AND email_account_id = ?", "", input.EmailAccountID)
		conflictMsg = "该邮箱账户已在此平台注册（无关联用户名）。"
	} else {
		// 理论上不应发生，因为前面应该有校验保证至少有一个不为空
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusBadRequest, "用户名和邮箱账户ID不能同时为空")
		return
	}

	err = query.First(&existingRegistration).Error
	if err == nil {
		// 明确找到记录，表示冲突
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusConflict, conflictMsg)
		return
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 明确未找到记录，无冲突，继续执行后续创建逻辑
	} else {
		// 发生其他数据库错误
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "检查平台注册冲突失败: "+err.Error())
		return
	}
	// --- 冲突检查结束，未找到冲突 (err == gorm.ErrRecordNotFound) ---

	// 创建新的 PlatformRegistration 记录
	var emailAccountIDPtr *uint
	if input.EmailAccountID > 0 {
		emailAccountIDPtr = &input.EmailAccountID
	}
	registration := models.PlatformRegistration{
		UserID:         currentUserID,
		EmailAccountID: emailAccountIDPtr,
		PlatformID:     input.PlatformID,
		LoginUsername: func() *string {
			if input.LoginUsername != "" {
				return &input.LoginUsername
			}
			return nil
		}(),
		LoginPasswordEncrypted: encryptedPassword,
		Notes:                  input.Notes,
		PhoneNumber:            input.PhoneNumber,
	}
	if createErr := tx.Create(&registration).Error; createErr != nil {
		tx.Rollback()
		if strings.Contains(createErr.Error(), "UNIQUE constraint failed") {
			// 根据输入推断是哪个约束冲突
			// 优先判断 EmailAccountID 是否可能导致冲突，因为它有一个涉及 EmailAccountID 的唯一索引组合
			// uq_user_platform_emailaccountid: (UserID, PlatformID, EmailAccountID)
			if input.EmailAccountID > 0 {
				// 实际冲突可能是 (UserID, PlatformID, EmailAccountID)
				// 此时返回邮箱账户相关的错误更准确
				utils.SendErrorResponse(c, http.StatusConflict, "此邮箱账户已在此平台注册。")
				return
			} else if input.LoginUsername != "" {
				// 如果 EmailAccountID 为 0 或 nil，再判断 LoginUsername 是否导致冲突
				// uq_user_platform_loginusername: (UserID, PlatformID, LoginUsername)
				utils.SendErrorResponse(c, http.StatusConflict, "此用户名已在该平台注册。")
				return
			}
			// 如果两者都为空（理论上不应发生，因为有预检查），或无法精确判断，则返回通用冲突消息
			utils.SendErrorResponse(c, http.StatusConflict, "创建失败，注册信息与现有记录冲突。")
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "创建平台注册信息失败: "+createErr.Error())
		}
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}

	// 创建成功，准备响应
	// 注意：如果 input.EmailAccountID 为 0，这里的 emailAccount 是零值
	response := registration.ToPlatformRegistrationResponse(emailAccount, platform)
	utils.SendSuccessResponse(c, response)
}

// CreatePlatformRegistrationByNames godoc
// @Summary 通过名称创建平台注册信息（自动创建邮箱/平台）
// @Description 为当前用户创建一个新的平台注册信息。如果提供的邮箱地址或平台名称不存在，则会自动创建。
// @Tags PlatformRegistrations
// @Accept json
// @Produce json
// @Param platformRegistration body handlers.CreatePlatformRegistrationInput true "平台注册信息（包含邮箱地址和平台名称）。"
// @Success 201 {object} models.SuccessResponse{data=models.PlatformRegistrationResponse} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /platform-registrations/by-name [post] // 新路由
// @Security BearerAuth
func CreatePlatformRegistrationByNames(c *gin.Context) {
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
	currentUserID := uint(userID)

	var input CreatePlatformRegistrationInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// --- 新增校验：用户名和邮箱地址不能同时为空 ---
	if input.LoginUsername == "" && input.EmailAddress == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "用户名和邮箱地址不能同时为空")
		return
	}
	// --- 校验结束 ---

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var err error // Declare err once here

	// 查找或创建 EmailAccount
	var emailAccount models.EmailAccount
	err = tx.Where("email_address = ? AND user_id = ?", input.EmailAddress, currentUserID).First(&emailAccount).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound { // 完全不存在，创建新的
			emailAccount = models.EmailAccount{
				UserID:       currentUserID,
				EmailAddress: input.EmailAddress,
				Provider:     utils.ExtractProviderFromEmail(input.EmailAddress),
				// Notes 可以在创建 EmailAccount 时考虑是否从 PlatformRegistrationInput 传递，或留空
			}
			if createErr := tx.Create(&emailAccount).Error; createErr != nil {
				// 这里的 UNIQUE constraint failed 错误是预期的，如果并发创建或数据库状态不一致
				tx.Rollback()
				utils.SendErrorResponse(c, http.StatusInternalServerError, "创建邮箱账户失败: "+createErr.Error())
				return
			}
		} else { // 其他查询错误
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询邮箱账户失败: "+err.Error())
			return
		}
	} else { // 找到了记录
		// 直接使用 emailAccount
	}

	// 查找或创建 Platform
	var platform models.Platform
	err = tx.Where("name = ? AND user_id = ?", input.PlatformName, currentUserID).First(&platform).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound { // 完全不存在，创建新的
			platform = models.Platform{
				UserID:     currentUserID,
				Name:       input.PlatformName,
				WebsiteURL: "", // 可以从 input.PlatformWebsiteURL 获取 (如果 CreatePlatformRegistrationInput 有此字段)
				Notes:      "", // 可以从 input.Notes 获取 (如果 CreatePlatformRegistrationInput 有此字段，但通常notes是registration的)
			}
			if createErr := tx.Create(&platform).Error; createErr != nil {
				tx.Rollback()
				utils.SendErrorResponse(c, http.StatusInternalServerError, "创建平台失败: "+createErr.Error())
				return
			}
		} else { // 其他查询错误
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台失败: "+err.Error())
			return
		}
	} else { // 找到了记录
		// 直接使用 platform
	}

	var encryptedPassword string
	// var err error // Remove this redundant declaration
	if input.LoginPassword != "" {
		encryptedPassword, err = utils.EncryptPassword(input.LoginPassword)
		if err != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "密码加密失败: "+err.Error())
			return
		}
	}

	// --- 精确冲突检查 ---
	var existingRegistration models.PlatformRegistration
	query := tx.Where("user_id = ? AND platform_id = ?", currentUserID, platform.ID)

	conflictMsg := ""
	currentEmailAccountID := emailAccount.ID // 获取前面查找或创建的 EmailAccount ID

	if input.LoginUsername != "" && currentEmailAccountID > 0 {
		// 检查用户名和邮箱组合
		query = query.Where("login_username = ? AND email_account_id = ?", input.LoginUsername, currentEmailAccountID)
		conflictMsg = "该用户名和邮箱组合已在此平台注册。"
	} else if input.LoginUsername != "" && currentEmailAccountID == 0 { // EmailAddress 为空时 ID 为 0
		// 仅检查用户名（邮箱为空或NULL）
		query = query.Where("login_username = ? AND (email_account_id = ? OR email_account_id IS NULL)", input.LoginUsername, 0)
		conflictMsg = "该用户名已在此平台注册（无关联邮箱）。"
	} else if input.LoginUsername == "" && currentEmailAccountID > 0 {
		// 仅检查邮箱（用户名为空或NULL）
		query = query.Where("(login_username = ? OR login_username IS NULL) AND email_account_id = ?", "", currentEmailAccountID)
		conflictMsg = "该邮箱账户已在此平台注册（无关联用户名）。"
	} else {
		// 前面已有校验 input.LoginUsername == "" && input.EmailAddress == ""
		// 此处理论上不会执行，但保留以防万一
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusBadRequest, "内部错误：用户名和邮箱地址不能同时为空")
		return
	}

	err = query.First(&existingRegistration).Error
	if err == nil {
		// 明确找到记录，表示冲突
		tx.Rollback()
		// 返回更详细的冲突信息，包含现有记录的ID
		conflictResponse := map[string]interface{}{
			"message":       conflictMsg,
			"existing_id":   existingRegistration.ID,
			"conflict_type": "duplicate_registration",
			"can_update":    true, // 表示可以更新密码
		}
		c.JSON(http.StatusConflict, models.Response{
			Code:    http.StatusConflict,
			Message: conflictMsg,
			Data:    conflictResponse,
		})
		return
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 明确未找到记录，无冲突，继续执行后续创建逻辑
	} else {
		// 发生其他数据库错误
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "检查平台注册冲突失败: "+err.Error())
		return
	}
	// --- 冲突检查结束，未找到冲突 (err == gorm.ErrRecordNotFound) ---

	// 创建新的 PlatformRegistration 记录
	var currentEmailAccountIDPtr *uint
	if currentEmailAccountID > 0 {
		currentEmailAccountIDPtr = &currentEmailAccountID
	}
	registration := models.PlatformRegistration{
		UserID:         currentUserID,
		EmailAccountID: currentEmailAccountIDPtr,
		PlatformID:     platform.ID,
		LoginUsername: func() *string {
			if input.LoginUsername != "" {
				return &input.LoginUsername
			}
			return nil
		}(),
		LoginPasswordEncrypted: encryptedPassword,
		Notes:                  input.Notes,
		PhoneNumber:            input.PhoneNumber,
	}
	if createErr := tx.Create(&registration).Error; createErr != nil {
		tx.Rollback()
		if strings.Contains(createErr.Error(), "UNIQUE constraint failed") {
			// 根据输入推断是哪个约束冲突
			if input.LoginUsername != "" {
				// 假设是用户名冲突
				utils.SendErrorResponse(c, http.StatusConflict, "此用户名已在此平台注册。")
				return
			} else if currentEmailAccountID > 0 { // currentEmailAccountID 来自 emailAccount.ID
				// 假设是邮箱账户ID冲突
				utils.SendErrorResponse(c, http.StatusConflict, "此邮箱账户已在此平台注册。")
				return
			}
			// 通用冲突消息
			utils.SendErrorResponse(c, http.StatusConflict, "创建失败，注册信息与现有记录冲突。")
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "创建平台注册信息失败: "+createErr.Error())
		}
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}

	// 创建成功，准备响应
	response := registration.ToPlatformRegistrationResponse(emailAccount, platform)
	utils.SendSuccessResponse(c, response)
}

// GetPlatformRegistrations godoc
// @Summary 获取当前用户的所有平台注册信息
// @Description 获取当前登录用户的所有平台注册信息，支持分页
// @Tags PlatformRegistrations
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param email_account_id query int false "按邮箱账户ID筛选"
// @Param platform_id query int false "按平台ID筛选"
// @Param username query string false "按平台用户名筛选"
// @Param orderBy query string false "排序字段 (e.g., login_username, created_at)" default(created_at)
// @Param sortDirection query string false "排序方向 (asc, desc)" default(desc)
// @Success 200 {object} models.SuccessResponse{data=[]models.PlatformRegistrationResponse,meta=models.PaginationMeta} "获取成功"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /platform-registrations [get]
// @Security BearerAuth
func GetPlatformRegistrations(c *gin.Context) {
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
	currentUserID := uint(userID)

	// --- Pagination and Filtering ---
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.Query("pageSize") // Get pageSize query param as string

	// Get filter parameters before calculating total records
	emailAccountIDQuery := c.Query("email_account_id")
	platformIDQuery := c.Query("platform_id")
	usernameFilter := strings.ToLower(strings.TrimSpace(c.Query("username"))) // 读取 username 参数
	var emailAccountIDFilter uint64
	if emailAccountIDQuery != "" {
		emailAccountIDFilter, _ = strconv.ParseUint(emailAccountIDQuery, 10, 32)
	}
	var platformIDFilter uint64
	if platformIDQuery != "" {
		platformIDFilter, _ = strconv.ParseUint(platformIDQuery, 10, 32)
	}

	orderBy := c.DefaultQuery("orderBy", "created_at")
	sortDirection := c.DefaultQuery("sortDirection", "desc")

	// Validate orderBy parameter
	// Note: Sorting by EmailAccount.email_address or Platform.name would require joins or subqueries
	// For now, only allow sorting by PlatformRegistration's own fields.
	allowedOrderByFields := map[string]string{
		"login_username": "platform_registrations.login_username",
		"notes":          "platform_registrations.notes",
		"created_at":     "platform_registrations.created_at",
		"updated_at":     "platform_registrations.updated_at",
		"email_address":  "email_accounts.email_address",
		"platform_name":  "platforms.name",
		"phone_number":   "platform_registrations.phone_number",
	}
	dbOrderByField, isValidField := allowedOrderByFields[orderBy]

	// Initialize query. We will add Joins to this query if needed.
	query := database.DB.Model(&models.PlatformRegistration{}).Where("platform_registrations.user_id = ?", currentUserID)
	countQuery := database.DB.Model(&models.PlatformRegistration{}).Where("user_id = ?", currentUserID) // countQuery doesn't need joins for sorting

	if !isValidField {
		dbOrderByField = "platform_registrations.created_at" // Default to created_at on the main table
	} else {
		if orderBy == "email_address" {
			query = query.Joins("LEFT JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id AND email_accounts.user_id = ?", currentUserID)
			// For countQuery, if filtering by email_account properties is ever needed, joins would be added there too.
			// But for sorting, countQuery remains simple.
		} else if orderBy == "platform_name" {
			query = query.Joins("LEFT JOIN platforms ON platforms.id = platform_registrations.platform_id AND platforms.user_id = ?", currentUserID)
		}
		// For other valid fields (login_username, notes, created_at, updated_at), no join is needed beyond what's already handled by allowedOrderByFields.
	}

	// Validate sortDirection
	if strings.ToLower(sortDirection) != "asc" && strings.ToLower(sortDirection) != "desc" {
		sortDirection = "desc" // Default to desc
	}
	orderClause := dbOrderByField + " " + sortDirection

	// Apply filters (these were originally applied to a query initialized later, moving them up)
	if emailAccountIDFilter > 0 {
		query = query.Where("platform_registrations.email_account_id = ?", uint(emailAccountIDFilter))
		countQuery = countQuery.Where("email_account_id = ?", uint(emailAccountIDFilter))
	}
	if platformIDFilter > 0 {
		query = query.Where("platform_registrations.platform_id = ?", uint(platformIDFilter))
		countQuery = countQuery.Where("platform_id = ?", uint(platformIDFilter))
	}
	if usernameFilter != "" { // 应用 username 筛选
		// 使用 LOWER on both sides for case-insensitive comparison
		query = query.Where("LOWER(platform_registrations.login_username) = LOWER(?)", usernameFilter)
		countQuery = countQuery.Where("LOWER(login_username) = LOWER(?)", usernameFilter)
	}
	// --- Calculate Total Records (needs to be done after filtering) ---
	var totalRecords int64
	if err := countQuery.Count(&totalRecords).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台注册总数失败: "+err.Error())
		return
	}

	// --- Process Pagination Parameters ---
	page, _ := strconv.Atoi(pageStr)
	if page <= 0 {
		page = 1
	}

	pageSize := 10 // Default page size
	fetchAll := false

	if pageSizeStr != "" {
		parsedPageSize, err := strconv.Atoi(pageSizeStr)
		if err == nil {
			if parsedPageSize > 0 {
				pageSize = parsedPageSize // Use provided positive value, no upper limit
			} else {
				// pageSize <= 0 means fetch all
				fetchAll = true
			}
		}
		// If parsing fails, pageSize remains the default 10
	}

	var offset int
	if fetchAll {
		pageSize = int(totalRecords) // Set pageSize to total records
		if pageSize < 0 {
			pageSize = 0
		} // Ensure pageSize is not negative if totalRecords is somehow negative (shouldn't happen)
		page = 1 // Force page to 1 when fetching all
		offset = 0
	} else {
		offset = (page - 1) * pageSize
	}

	// --- Fetch Paginated Data ---
	var registrations []models.PlatformRegistration

	// Apply sorting, preloading, and potentially pagination limits
	query = query.Order(orderClause) // Apply sorting order

	if !fetchAll && totalRecords > 0 { // Only apply limit and offset if not fetching all and there are records
		query = query.Offset(offset).Limit(pageSize)
	}
	// If fetchAll is true, no Offset or Limit is applied, retrieving all records matching filters.
	// If totalRecords is 0, Offset/Limit don't matter but skipping them is cleaner.

	// Preload related data for the response
	// For now, we sort by PlatformRegistration fields and then preload.
	// Example for JOIN and sort (more complex):
	// if orderBy == "email_address" {
	// query = query.Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id")
	// orderClause = "email_accounts.email_address " + sortDirection
	// } else if orderBy == "platform_name" {
	// query = query.Joins("JOIN platforms ON platforms.id = platform_registrations.platform_id")
	// orderClause = "platforms.name " + sortDirection
	// }

	// Always preload after potential pagination
	query = query.Preload("EmailAccount").Preload("Platform")

	if err := query.Find(&registrations).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台注册列表失败: "+err.Error())
		return
	}

	var responses []models.PlatformRegistrationResponse
	for _, pr := range registrations {
		emailAccountForResp := models.EmailAccount{}
		if pr.EmailAccount != nil {
			emailAccountForResp = *pr.EmailAccount
		}
		responses = append(responses, pr.ToPlatformRegistrationResponse(emailAccountForResp, pr.Platform))
	}

	pagination := utils.CreatePaginationMeta(page, pageSize, int(totalRecords))
	utils.SendSuccessResponseWithMeta(c, responses, pagination)
}

// GetPlatformRegistrationByID godoc
// @Summary 获取指定ID的平台注册信息详情
// @Description 获取当前用户拥有的指定ID的平台注册信息详情
// @Tags PlatformRegistrations
// @Produce json
// @Param id path int true "平台注册ID"
// @Success 200 {object} models.SuccessResponse{data=models.PlatformRegistrationResponse} "获取成功"
// @Failure 400 {object} models.ErrorResponse "无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "平台注册信息未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /platform-registrations/{id} [get]
// @Security BearerAuth
func GetPlatformRegistrationByID(c *gin.Context) {
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
	currentUserID := uint(userID)

	registrationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的平台注册ID格式")
		return
	}

	var registration models.PlatformRegistration
	if err := database.DB.Where("id = ? AND user_id = ?", registrationID, currentUserID).Preload("EmailAccount").Preload("Platform").First(&registration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "平台注册信息未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台注册详情失败: "+err.Error())
		return
	}

	emailAccountForRespGetByID := models.EmailAccount{}
	if registration.EmailAccount != nil {
		emailAccountForRespGetByID = *registration.EmailAccount
	}
	response := registration.ToPlatformRegistrationResponse(emailAccountForRespGetByID, registration.Platform)
	utils.SendSuccessResponse(c, response)
}

// UpdatePlatformRegistration godoc
// @Summary 更新指定ID的平台注册信息
// @Description 更新当前用户拥有的指定ID的平台注册信息。支持通过邮箱地址自动查找或创建邮箱账户。
// @Tags PlatformRegistrations
// @Accept json
// @Produce json
// @Param id path int true "平台注册ID"
// @Param platformRegistration body object{email_address=string,login_username=string,login_password=string,notes=string,phone_number=string} true "要更新的平台注册信息。邮箱地址会自动查找或创建对应的邮箱账户。密码可选。"
// @Success 200 {object} models.SuccessResponse{data=models.PlatformRegistrationResponse} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误或无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "平台注册信息未找到"
// @Failure 409 {object} models.ErrorResponse "邮箱账户已在此平台注册"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /platform-registrations/{id} [put]
// @Security BearerAuth
func UpdatePlatformRegistration(c *gin.Context) {
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
	currentUserID := uint(userID)

	registrationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的平台注册ID格式")
		return
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var registration models.PlatformRegistration
	// Preload EmailAccount and Platform to be used in the response
	if err := tx.Where("id = ? AND user_id = ?", registrationID, currentUserID).Preload("EmailAccount").Preload("Platform").First(&registration).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "平台注册信息未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询待更新平台注册信息失败: "+err.Error())
		return
	}

	var input struct {
		EmailAddress  string `json:"email_address" binding:"omitempty,email"` // 修改为接受邮箱地址
		LoginUsername string `json:"login_username"`
		LoginPassword string `json:"login_password" binding:"omitempty,min=6"` // 密码可选
		Notes         string `json:"notes"`
		PhoneNumber   string `json:"phone_number"` // 手机号码，可选
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	originalLoginUsername := registration.LoginUsername
	// 更新基本字段
	if input.LoginUsername != "" {
		registration.LoginUsername = &input.LoginUsername
	} else {
		registration.LoginUsername = nil
	}
	registration.Notes = input.Notes
	registration.PhoneNumber = input.PhoneNumber

	// 更新密码（如果提供）
	if input.LoginPassword != "" {
		encryptedPassword, err := utils.EncryptPassword(input.LoginPassword)
		if err != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "密码加密失败: "+err.Error())
			return
		}
		registration.LoginPasswordEncrypted = encryptedPassword
	}

	// --- 新增：如果 LoginUsername 发生变化且非空，则检查冲突 ---
	loginUsernameChanged := func() bool {
		if originalLoginUsername == nil && registration.LoginUsername == nil {
			return false
		}
		if originalLoginUsername == nil || registration.LoginUsername == nil {
			return true
		}
		return *originalLoginUsername != *registration.LoginUsername
	}()

	if loginUsernameChanged && registration.LoginUsername != nil && *registration.LoginUsername != "" {
		var existingUserReg models.PlatformRegistration
		errCheckUser := tx.Where("login_username = ? AND platform_id = ? AND user_id = ? AND id != ?",
			*registration.LoginUsername, registration.PlatformID, currentUserID, registration.ID).First(&existingUserReg).Error
		if errCheckUser == nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusConflict, "此用户名已在此平台注册，无法更新。")
			return
		} else if errCheckUser != gorm.ErrRecordNotFound {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "检查用户名唯一性失败: "+errCheckUser.Error())
			return
		}
		// 如果 errCheckUser == gorm.ErrRecordNotFound，说明用户名可用，继续
	}
	// --- 预检查结束 ---

	// 处理邮箱地址更新逻辑
	var newEmailAccount models.EmailAccount
	var newEmailAccountID *uint

	if input.EmailAddress != "" {
		// 查找或创建邮箱账户
		err = tx.Where("user_id = ? AND email_address = ?", currentUserID, input.EmailAddress).First(&newEmailAccount).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 邮箱账户不存在，创建新的
				newEmailAccount = models.EmailAccount{
					UserID:       currentUserID,
					EmailAddress: input.EmailAddress,
					Provider:     utils.ExtractProviderFromEmail(input.EmailAddress),
					Notes:        "", // 可以根据需要设置
				}
				if createErr := tx.Create(&newEmailAccount).Error; createErr != nil {
					tx.Rollback()
					utils.SendErrorResponse(c, http.StatusInternalServerError, "创建邮箱账户失败: "+createErr.Error())
					return
				}
			} else {
				// 其他查询错误
				tx.Rollback()
				utils.SendErrorResponse(c, http.StatusInternalServerError, "查询邮箱账户失败: "+err.Error())
				return
			}
		}

		newEmailAccountID = &newEmailAccount.ID

		// 检查新的 (EmailAccountID, PlatformID) 组合是否已存在（违反唯一约束）
		var existingRegistration models.PlatformRegistration
		// 确保不与自身比较
		err = tx.Where("email_account_id = ? AND platform_id = ? AND user_id = ? AND id != ?", newEmailAccount.ID, registration.PlatformID, currentUserID, registration.ID).First(&existingRegistration).Error
		if err == nil {
			// 找到了一个存在的记录，不允许更新，因为会违反唯一约束
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusConflict, "该邮箱账户已在此平台注册，无法将当前注册信息更新为该邮箱。")
			return
		} else if err != gorm.ErrRecordNotFound {
			// 查询时发生其他错误
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "检查唯一约束失败: "+err.Error())
			return
		}

		// 更新 EmailAccountID
		registration.EmailAccountID = newEmailAccountID
		registration.EmailAccount = &newEmailAccount
	} else {
		// 邮箱地址为空，设置为 nil
		registration.EmailAccountID = nil
		registration.EmailAccount = nil
	}

	// --- 新增校验：更新后用户名和关联邮箱不能都为空/无效 ---
	if (registration.LoginUsername == nil || *registration.LoginUsername == "") && (registration.EmailAccountID == nil || *registration.EmailAccountID == 0) {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusBadRequest, "更新失败：用户名和关联邮箱不能同时为空或无效")
		return
	}
	// --- 校验结束 ---

	if err := tx.Save(&registration).Error; err != nil {
		tx.Rollback()
		// 理论上，如果前面的检查都通过了，这里的保存不应再触发唯一约束错误，但以防万一
		if strings.Contains(err.Error(), "UNIQUE constraint failed") { // 检查 SQLite 特有的唯一约束错误
			// 尝试判断是哪个字段引起的冲突
			// 注意：这里的判断可能不完美，因为Save操作可能同时更新多个字段
			// 更好的做法是进行更细致的预检查
			if input.LoginUsername != "" { // 如果尝试更新的用户名非空
				utils.SendErrorResponse(c, http.StatusConflict, "更新失败，此用户名可能已在该平台注册。")
			} else if input.EmailAddress != "" { // 如果尝试更新的邮箱地址非空
				utils.SendErrorResponse(c, http.StatusConflict, "更新失败，此邮箱账户可能已在该平台注册。")
			} else {
				utils.SendErrorResponse(c, http.StatusConflict, "更新失败，可能导致重复的平台注册信息。")
			}

		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "更新平台注册信息失败: "+err.Error())
		}
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}

	emailAccountForRespUpdate := models.EmailAccount{}
	if registration.EmailAccount != nil {
		emailAccountForRespUpdate = *registration.EmailAccount
	}
	response := registration.ToPlatformRegistrationResponse(emailAccountForRespUpdate, registration.Platform)
	utils.SendSuccessResponse(c, response)
}

// GetPlatformRegistrationPassword godoc
// @Summary 获取平台注册密码
// @Description 获取指定平台注册信息的解密密码
// @Tags PlatformRegistrations
// @Produce json
// @Param id path int true "平台注册ID"
// @Success 200 {object} models.SuccessResponse{data=map[string]string} "获取成功"
// @Failure 400 {object} models.ErrorResponse "无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "平台注册信息未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /platform-registrations/{id}/password [get]
// @Security BearerAuth
func GetPlatformRegistrationPassword(c *gin.Context) {
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
	currentUserID := uint(userID)

	registrationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的平台注册ID格式")
		return
	}

	// 查询平台注册信息
	var registration models.PlatformRegistration
	if err := database.DB.Where("id = ? AND user_id = ?", registrationID, currentUserID).First(&registration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "平台注册信息未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台注册信息失败: "+err.Error())
		return
	}

	// 检查是否有密码
	if registration.LoginPasswordEncrypted == "" {
		utils.SendErrorResponse(c, http.StatusNotFound, "该注册信息未设置密码")
		return
	}

	// 解密密码
	var decryptedPassword string
	if utils.IsEncryptedPassword(registration.LoginPasswordEncrypted) {
		// 新的加密格式，可以解密
		decryptedPassword, err = utils.DecryptPassword(registration.LoginPasswordEncrypted)
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

// DeletePlatformRegistration godoc
// @Summary 删除指定ID的平台注册信息
// @Description 删除当前用户拥有的指定ID的平台注册信息
// @Tags PlatformRegistrations
// @Produce json
// @Param id path int true "平台注册ID"
// @Success 200 {object} models.SuccessResponse "删除成功"
// @Failure 400 {object} models.ErrorResponse "无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "平台注册信息未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /platform-registrations/{id} [delete]
// @Security BearerAuth
func DeletePlatformRegistration(c *gin.Context) {
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
	currentUserID := uint(userID)

	registrationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的平台注册ID格式")
		return
	}

	var registration models.PlatformRegistration
	if err := database.DB.Where("id = ? AND user_id = ?", registrationID, currentUserID).First(&registration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "平台注册信息未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询待删除平台注册信息失败: "+err.Error())
		return
	}

	// 使用事务确保操作的原子性
	tx := database.DB.Begin()
	if tx.Error != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "开启事务失败: "+tx.Error.Error())
		return
	}

	// 1. 硬删除关联的 ServiceSubscriptions
	if err := tx.Unscoped().Where("platform_registration_id = ? AND user_id = ?", registration.ID, currentUserID).Delete(&models.ServiceSubscription{}).Error; err != nil {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "删除服务订阅失败: "+err.Error())
		return
	}

	// 2. 硬删除 PlatformRegistration
	if err := tx.Unscoped().Delete(&registration).Error; err != nil {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "删除平台注册信息失败: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "平台注册信息删除成功"})
}

// GetPlatformRegistrationsByEmailAccountID godoc
// @Summary 获取指定邮箱账户关联的所有平台注册信息
// @Description 获取当前用户拥有的指定邮箱账户所关联的所有平台注册信息
// @Tags PlatformRegistrations
// @Produce json
// @Param email_account_id path int true "邮箱账户ID"
// @Success 200 {object} models.SuccessResponse{data=[]map[string]interface{}} "获取成功"
// @Failure 400 {object} models.ErrorResponse "无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该邮箱账户"
// @Failure 404 {object} models.ErrorResponse "邮箱账户未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Router /email-accounts/{id}/platform-registrations [get] // Path updated in main.go
// @Security BearerAuth
func GetPlatformRegistrationsByEmailAccountID(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}
	currentUserID, ok := userIDRaw.(int64)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "用户ID类型错误")
		return
	}

	emailAccountIDParam := c.Param("id") // 修改参数名
	emailAccountID, err := strconv.ParseUint(emailAccountIDParam, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的邮箱账户ID格式")
		return
	}

	// 验证邮箱账户是否属于当前用户
	var emailAccount models.EmailAccount
	if err := database.DB.Where("id = ? AND user_id = ?", emailAccountID, uint(currentUserID)).First(&emailAccount).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusForbidden, "无权访问该邮箱账户或邮箱账户不存在")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询邮箱账户失败: "+err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	} // Max page size limit
	offset := (page - 1) * pageSize

	var registrations []models.PlatformRegistration
	var totalRecords int64

	dbQuery := database.DB.Where("email_account_id = ? AND user_id = ?", emailAccountID, uint(currentUserID))

	// Count total records
	if err := dbQuery.Model(&models.PlatformRegistration{}).Count(&totalRecords).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "统计关联平台注册总数失败: "+err.Error())
		return
	}

	// Fetch paginated records
	if err := dbQuery.Preload("Platform").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&registrations).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台注册信息失败: "+err.Error())
		return
	}

	type ResponseItem struct {
		PlatformID         uint   `json:"platform_id"`
		PlatformName       string `json:"platform_name"`
		PlatformWebsiteURL string `json:"platform_website_url"`
		RegistrationNotes  string `json:"registration_notes"`
	}
	var responseData []ResponseItem

	for _, reg := range registrations {
		responseData = append(responseData, ResponseItem{
			PlatformID:         reg.PlatformID,
			PlatformName:       reg.Platform.Name,
			PlatformWebsiteURL: reg.Platform.WebsiteURL,
			RegistrationNotes:  reg.Notes,
		})
	}

	pagination := utils.CreatePaginationMeta(page, pageSize, int(totalRecords))
	utils.SendSuccessResponseWithMeta(c, responseData, pagination)
}

// GetEmailRegistrationsByPlatformID godoc
// @Summary 获取指定平台关联的所有邮箱注册信息
// @Description 获取当前用户在指定平台上注册的所有邮箱账户信息
// @Tags PlatformRegistrations
// @Produce json
// @Param platform_id path int true "平台ID"
// @Success 200 {object} models.SuccessResponse{data=[]map[string]interface{}} "获取成功"
// @Failure 400 {object} models.ErrorResponse "无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 404 {object} models.ErrorResponse "平台未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Router /platforms/{id}/email-registrations [get] // Path updated in main.go
// @Security BearerAuth
func GetEmailRegistrationsByPlatformID(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "用户未认证")
		return
	}
	currentUserID, ok := userIDRaw.(int64)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "用户ID类型错误")
		return
	}

	platformIDParam := c.Param("id") // 修改参数名
	platformID, err := strconv.ParseUint(platformIDParam, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的平台ID格式")
		return
	}

	// 验证平台是否存在且属于当前用户
	var platform models.Platform
	if err := database.DB.Where("id = ? AND user_id = ?", platformID, uint(currentUserID)).First(&platform).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "平台未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台失败: "+err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	} // Max page size limit
	offset := (page - 1) * pageSize

	var totalRecords int64

	// Base query joining with email_accounts and filtering for valid addresses
	baseQuery := database.DB.Model(&models.PlatformRegistration{}).
		Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id").
		Where("platform_registrations.platform_id = ? AND platform_registrations.user_id = ? AND platform_registrations.email_account_id IS NOT NULL AND platform_registrations.email_account_id > 0 AND email_accounts.email_address IS NOT NULL AND email_accounts.email_address <> ''", platformID, uint(currentUserID))

	// Count total records using the base query
	countQuery := baseQuery
	// We need to be careful with Count() after Joins, sometimes it requires specifying the count column.
	// Let's count on the primary key of the main table to be safe.
	if err := countQuery.Distinct("platform_registrations.id").Count(&totalRecords).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "统计关联邮箱注册总数失败: "+err.Error())
		return
	}

	// Define the structure for the response items
	type ResponseItem struct {
		EmailAccountID    uint   `json:"email_account_id"`
		EmailAddress      string `json:"email_address"`
		RegistrationNotes string `json:"registration_notes"`
	}
	var responseData []ResponseItem

	// Define a temporary structure to scan query results into
	type QueryResultItem struct {
		EmailAccountID    uint   `gorm:"column:email_account_id"`
		EmailAddress      string `gorm:"column:email_address"`
		RegistrationNotes string `gorm:"column:notes"`
	}
	var queryResults []QueryResultItem

	// Fetch paginated records using the base query, selecting specific fields
	fetchQuery := baseQuery
	if err := fetchQuery.
		Select("platform_registrations.email_account_id, email_accounts.email_address, platform_registrations.notes").
		Offset(offset).
		Limit(pageSize).
		Order("platform_registrations.created_at DESC").
		Scan(&queryResults).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取邮箱注册信息失败: "+err.Error())
		return
	}

	// Populate responseData from queryResults
	for _, qr := range queryResults {
		responseData = append(responseData, ResponseItem{
			EmailAccountID:    qr.EmailAccountID,
			EmailAddress:      qr.EmailAddress,
			RegistrationNotes: qr.RegistrationNotes,
		})
	}

	pagination := utils.CreatePaginationMeta(page, pageSize, int(totalRecords))
	utils.SendSuccessResponseWithMeta(c, responseData, pagination)
}
