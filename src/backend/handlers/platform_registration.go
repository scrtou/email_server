package handlers

import (
	"email_server/database"
	"email_server/models"
	"email_server/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)
// CreatePlatformRegistrationInput 定义了创建平台注册信息的输入结构
type CreatePlatformRegistrationInput struct {
	EmailAddress  string `json:"email_address" binding:"required,email"`
	PlatformName  string `json:"platform_name" binding:"required"`
	LoginUsername string `json:"login_username"`
	LoginPassword string `json:"login_password" binding:"omitempty,min=6"` // 密码可选
	Notes         string `json:"notes"`
	// 可以根据需要添加 Provider (针对EmailAccount) 和 WebsiteURL (针对Platform)
	// EmailProvider    string `json:"email_provider"`
	// PlatformWebsiteURL string `json:"platform_website_url"`
}
// CreatePlatformRegistrationWithIDsInput 定义了通过ID创建平台注册信息的输入结构
type CreatePlatformRegistrationWithIDsInput struct {
	EmailAccountID uint   `json:"email_account_id" binding:"required"`
	PlatformID     uint   `json:"platform_id" binding:"required"`
	LoginUsername  string `json:"login_username"`
	LoginPassword  string `json:"login_password" binding:"omitempty,min=6"` // 密码可选
	Notes          string `json:"notes"`
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

	var err error // Declare err once

	// 验证 EmailAccount 是否属于当前用户
	var emailAccount models.EmailAccount
	err = database.DB.Where("id = ? AND user_id = ?", input.EmailAccountID, currentUserID).First(&emailAccount).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "关联的邮箱账户未找到或不属于当前用户")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询邮箱账户失败: "+err.Error())
		return
	}

	// 验证 Platform 是否存在且属于当前用户
	var platform models.Platform
	err = database.DB.Where("id = ? AND user_id = ?", input.PlatformID, currentUserID).First(&platform).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "关联的平台未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台失败: "+err.Error())
		return
	}

	var hashedPassword string
	if input.LoginPassword != "" {
		hashedPassword, err = utils.HashPassword(input.LoginPassword)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "密码加密失败: "+err.Error())
			return
		}
	}

	// 查找或创建/恢复 PlatformRegistration 记录
	var registration models.PlatformRegistration
	err = database.DB.Unscoped().Where(models.PlatformRegistration{
		UserID:         currentUserID,
		EmailAccountID: input.EmailAccountID,
		PlatformID:     input.PlatformID,
	}).First(&registration).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound { // 记录不存在，创建新的
			registration = models.PlatformRegistration{
				UserID:                 currentUserID,
				EmailAccountID:         input.EmailAccountID,
				PlatformID:             input.PlatformID,
				LoginUsername:          input.LoginUsername,
				LoginPasswordEncrypted: hashedPassword,
				Notes:                  input.Notes,
			}
			if createErr := database.DB.Create(&registration).Error; createErr != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "创建平台注册信息失败: "+createErr.Error())
				return
			}
		} else { // 其他查询错误
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台注册信息失败: "+err.Error())
			return
		}
	} else { // 找到了记录
		if registration.DeletedAt.Valid { // 如果是软删除的记录
			// 恢复记录并更新字段
			registration.LoginUsername = input.LoginUsername
			registration.LoginPasswordEncrypted = hashedPassword
			registration.Notes = input.Notes
			registration.DeletedAt = gorm.DeletedAt{} // 重置软删除标记
			if updateErr := database.DB.Unscoped().Save(&registration).Error; updateErr != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "恢复并更新平台注册信息失败: "+updateErr.Error())
				return
			}
		} else { // 记录已存在且未被软删除
			utils.SendErrorResponse(c, http.StatusConflict, "该邮箱账户已在此平台注册。")
			return
		}
	}
	
	// 确保 emailAccount 和 platform 包含最新的信息 (如果它们被恢复或创建)
	// GetPlatformRegistrationByID 预加载了这些，所以这里我们直接使用传入的或新创建/恢复的
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

	var err error // Declare err once here

	// 查找或创建 EmailAccount
	var emailAccount models.EmailAccount
	// 尝试查找包括软删除在内的记录
	err = database.DB.Unscoped().Where("email_address = ? AND user_id = ?", input.EmailAddress, currentUserID).First(&emailAccount).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound { // 完全不存在，创建新的
			emailAccount = models.EmailAccount{
				UserID:       currentUserID,
				EmailAddress: input.EmailAddress,
				Provider:     utils.ExtractProviderFromEmail(input.EmailAddress),
				// Notes 可以在创建 EmailAccount 时考虑是否从 PlatformRegistrationInput 传递，或留空
			}
			if createErr := database.DB.Create(&emailAccount).Error; createErr != nil {
				// 这里的 UNIQUE constraint failed 错误是预期的，如果并发创建或数据库状态不一致
				utils.SendErrorResponse(c, http.StatusInternalServerError, "创建邮箱账户失败: "+createErr.Error())
				return
			}
		} else { // 其他查询错误
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询邮箱账户失败: "+err.Error())
			return
		}
	} else { // 找到了记录
		if emailAccount.DeletedAt.Valid { // 如果是软删除的记录
			// 恢复该记录
			if updateErr := database.DB.Unscoped().Model(&emailAccount).Update("deleted_at", nil).Error; updateErr != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "恢复邮箱账户失败: "+updateErr.Error())
				return
			}
			// GORM 会自动更新 emailAccount 实例的 DeletedAt 字段为 nil (或使其 .Valid 为 false)
			// 如果需要确保其他字段也刷新，可以重新查询一次，但通常对于 Update("deleted_at", nil) 是不必要的。
		}
		// 如果不是软删除的，或者已成功恢复，则直接使用 emailAccount
	}

	// 查找或创建 Platform
	var platform models.Platform
	// 查找当前用户是否已创建同名平台 (包括软删除的)
	err = database.DB.Unscoped().Where("name = ? AND user_id = ?", input.PlatformName, currentUserID).First(&platform).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound { // 完全不存在，创建新的
			platform = models.Platform{
				UserID:     currentUserID,
				Name:       input.PlatformName,
				WebsiteURL: "", // 可以从 input.PlatformWebsiteURL 获取 (如果 CreatePlatformRegistrationInput 有此字段)
				Notes:      "", // 可以从 input.Notes 获取 (如果 CreatePlatformRegistrationInput 有此字段，但通常notes是registration的)
			}
			if createErr := database.DB.Create(&platform).Error; createErr != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "创建平台失败: "+createErr.Error())
				return
			}
		} else { // 其他查询错误
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台失败: "+err.Error())
			return
		}
	} else { // 找到了记录
		if platform.DeletedAt.Valid { // 如果是软删除的记录
			// 恢复该记录
			if updateErr := database.DB.Unscoped().Model(&platform).Update("deleted_at", nil).Error; updateErr != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "恢复平台失败: "+updateErr.Error())
				return
			}
		}
		// 如果不是软删除的，或者已成功恢复，则直接使用 platform
	}

	var hashedPassword string
	// var err error // Remove this redundant declaration
	if input.LoginPassword != "" {
		hashedPassword, err = utils.HashPassword(input.LoginPassword)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "密码加密失败: "+err.Error())
			return
		}
	}

	// 查找或创建/恢复 PlatformRegistration 记录
	var registration models.PlatformRegistration
	err = database.DB.Unscoped().Where(models.PlatformRegistration{
		UserID:         currentUserID,
		EmailAccountID: emailAccount.ID, // 使用已确定/创建/恢复的 emailAccount.ID
		PlatformID:     platform.ID,     // 使用已确定/创建/恢复的 platform.ID
	}).First(&registration).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound { // 记录不存在，创建新的
			registration = models.PlatformRegistration{
				UserID:                 currentUserID,
				EmailAccountID:         emailAccount.ID,
				PlatformID:             platform.ID,
				LoginUsername:          input.LoginUsername,
				LoginPasswordEncrypted: hashedPassword,
				Notes:                  input.Notes,
			}
			if createErr := database.DB.Create(&registration).Error; createErr != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "创建平台注册信息失败: "+createErr.Error())
				return
			}
		} else { // 其他查询错误
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台注册信息失败: "+err.Error())
			return
		}
	} else { // 找到了记录
		if registration.DeletedAt.Valid { // 如果是软删除的记录
			// 恢复记录并更新字段
			registration.LoginUsername = input.LoginUsername
			registration.LoginPasswordEncrypted = hashedPassword
			registration.Notes = input.Notes
			registration.DeletedAt = gorm.DeletedAt{} // 重置软删除标记
			if updateErr := database.DB.Unscoped().Save(&registration).Error; updateErr != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "恢复并更新平台注册信息失败: "+updateErr.Error())
				return
			}
		} else { // 记录已存在且未被软删除
			utils.SendErrorResponse(c, http.StatusConflict, "该邮箱账户已在此平台注册。")
			return
		}
	}
	
	// 为了返回完整的 PlatformRegistrationResponse，我们需要 emailAccount 和 platform 的信息
	// 在创建时我们已经查询过了
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// Get filter parameters
	emailAccountIDQuery := c.Query("email_account_id")
	platformIDQuery := c.Query("platform_id")
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
	}
	dbOrderByField, isValidField := allowedOrderByFields[orderBy]
	
	// Initialize query. We will add Joins to this query if needed.
	query := database.DB.Model(&models.PlatformRegistration{}).Where("platform_registrations.user_id = ?", currentUserID)
	countQuery := database.DB.Model(&models.PlatformRegistration{}).Where("user_id = ?", currentUserID) // countQuery doesn't need joins for sorting

	if !isValidField {
		dbOrderByField = "platform_registrations.created_at" // Default to created_at on the main table
	} else {
		if orderBy == "email_address" {
			query = query.Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id AND email_accounts.user_id = ?", currentUserID)
			// For countQuery, if filtering by email_account properties is ever needed, joins would be added there too.
			// But for sorting, countQuery remains simple.
		} else if orderBy == "platform_name" {
			query = query.Joins("JOIN platforms ON platforms.id = platform_registrations.platform_id AND platforms.user_id = ?", currentUserID)
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
	if page <= 0 {	page = 1 }
	if pageSize <= 0 { pageSize = 10	}
	if pageSize > 100 { pageSize = 100 }
	offset := (page - 1) * pageSize

	var registrations []models.PlatformRegistration
	var totalRecords int64

	// query and countQuery are now initialized and potentially filtered before this point.
	// The Joins for sorting are applied to 'query', not 'countQuery'.

	if err := countQuery.Count(&totalRecords).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台注册总数失败: "+err.Error())
		return
	}
	
	// Preload related data for the response
	// If sorting by related fields (e.g., email_accounts.email_address), a JOIN would be needed here.
	// For now, we sort by PlatformRegistration fields and then preload.
	// Example for JOIN and sort (more complex):
	// if orderBy == "email_address" {
	// query = query.Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id")
	// orderClause = "email_accounts.email_address " + sortDirection
	// } else if orderBy == "platform_name" {
	// query = query.Joins("JOIN platforms ON platforms.id = platform_registrations.platform_id")
	// orderClause = "platforms.name " + sortDirection
	// }

	if err := query.Order(orderClause).Offset(offset).Limit(pageSize).Preload("EmailAccount").Preload("Platform").Find(&registrations).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台注册列表失败: "+err.Error())
		return
	}

	var responses []models.PlatformRegistrationResponse
	for _, pr := range registrations {
		responses = append(responses, pr.ToPlatformRegistrationResponse(pr.EmailAccount, pr.Platform))
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
	
	response := registration.ToPlatformRegistrationResponse(registration.EmailAccount, registration.Platform)
	utils.SendSuccessResponse(c, response)
}

// UpdatePlatformRegistration godoc
// @Summary 更新指定ID的平台注册信息
// @Description 更新当前用户拥有的指定ID的平台注册信息
// @Tags PlatformRegistrations
// @Accept json
// @Produce json
// @Param id path int true "平台注册ID"
// @Param platformRegistration body models.PlatformRegistration true "要更新的平台注册信息。UserID, EmailAccountID, PlatformID 不可更改。密码可选。"
// @Success 200 {object} models.SuccessResponse{data=models.PlatformRegistrationResponse} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误或无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "平台注册信息未找到"
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

	var registration models.PlatformRegistration
	// Preload EmailAccount and Platform to be used in the response
	if err := database.DB.Where("id = ? AND user_id = ?", registrationID, currentUserID).Preload("EmailAccount").Preload("Platform").First(&registration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "平台注册信息未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询待更新平台注册信息失败: "+err.Error())
		return
	}

	var input struct {
		LoginUsername string `json:"login_username"`
		LoginPassword string `json:"login_password" binding:"omitempty,min=6"` // 密码可选
		Notes         string `json:"notes"`
		// EmailAccountID and PlatformID are not updatable via this endpoint to maintain integrity.
		// If they need to be changed, it's conceptually a new registration.
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	registration.LoginUsername = input.LoginUsername
	registration.Notes = input.Notes

	if input.LoginPassword != "" {
		hashedPassword, err := utils.HashPassword(input.LoginPassword)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "密码加密失败: "+err.Error())
			return
		}
		registration.LoginPasswordEncrypted = hashedPassword
	}

	if err := database.DB.Save(&registration).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "更新平台注册信息失败: "+err.Error())
		return
	}
	
	response := registration.ToPlatformRegistrationResponse(registration.EmailAccount, registration.Platform)
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
	
	// Consider if deleting a PlatformRegistration should cascade to ServiceSubscriptions
	// GORM's default Delete is a soft delete if gorm.DeletedAt field exists.
	if err := database.DB.Delete(&registration).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "删除平台注册信息失败: "+err.Error())
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
	if page <= 0 {	page = 1 }
	if pageSize <= 0 { pageSize = 10	}
	if pageSize > 100 { pageSize = 100 } // Max page size limit
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
	if page <= 0 {	page = 1 }
	if pageSize <= 0 { pageSize = 10	}
	if pageSize > 100 { pageSize = 100 } // Max page size limit
	offset := (page - 1) * pageSize

	var registrations []models.PlatformRegistration
	var totalRecords int64
	
	dbQuery := database.DB.Where("platform_id = ? AND user_id = ?", platformID, uint(currentUserID))

	// Count total records
	if err := dbQuery.Model(&models.PlatformRegistration{}).Count(&totalRecords).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "统计关联邮箱注册总数失败: "+err.Error())
		return
	}

	// Fetch paginated records
	if err := dbQuery.Preload("EmailAccount").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&registrations).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取邮箱注册信息失败: "+err.Error())
		return
	}

	type ResponseItem struct {
		EmailAccountID    uint   `json:"email_account_id"`
		EmailAddress      string `json:"email_address"`
		RegistrationNotes string `json:"registration_notes"`
	}
	var responseData []ResponseItem

	for _, reg := range registrations {
		responseData = append(responseData, ResponseItem{
			EmailAccountID:    reg.EmailAccountID,
			EmailAddress:      reg.EmailAccount.EmailAddress,
			RegistrationNotes: reg.Notes,
		})
	}

	pagination := utils.CreatePaginationMeta(page, pageSize, int(totalRecords))
	utils.SendSuccessResponseWithMeta(c, responseData, pagination)
}