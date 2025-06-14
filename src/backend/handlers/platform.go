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

// CreatePlatform godoc
// @Summary 创建平台
// @Description 创建一个新的平台信息
// @Tags Platforms
// @Accept json
// @Produce json
// @Param platform body models.Platform true "平台信息，ID、CreatedAt、UpdatedAt、DeletedAt 会被忽略"
// @Success 201 {object} models.SuccessResponse{data=models.PlatformResponse} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /platforms [post]
// @Security BearerAuth
func CreatePlatform(c *gin.Context) {
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

	var input struct {
		Name       string `json:"name" binding:"required,min=2,max=100"`
		WebsiteURL string `json:"website_url" binding:"omitempty,url"`
		Notes      string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// 首先检查是否存在同名的记录
	var existingPlatform models.Platform
	err := database.DB.Where("user_id = ? AND name = ?", currentUserID, input.Name).First(&existingPlatform).Error

	if err == nil {
		// 找到了同名记录
		utils.SendErrorResponse(c, http.StatusConflict, "您已创建过同名平台")
		return
	} else if err != gorm.ErrRecordNotFound {
		// 查询出错
		utils.SendErrorResponse(c, http.StatusInternalServerError, "检查平台是否存在失败: "+err.Error())
		return
	}

	// 没有找到同名记录，创建新平台
	platform := models.Platform{
		UserID:     currentUserID,
		Name:       input.Name,
		WebsiteURL: input.WebsiteURL,
		Notes:      input.Notes,
	}

	if err := database.DB.Create(&platform).Error; err != nil {
		if utils.IsUniqueConstraintError(err) {
			utils.SendErrorResponse(c, http.StatusConflict, "您已创建过同名平台")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "创建平台失败: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, platform.ToPlatformResponse())
}

// GetPlatforms godoc
// @Summary 获取所有平台信息
// @Description 获取所有平台信息，支持分页。pageSize > 0 时分页；pageSize <= 0 时获取所有；未提供或无效时使用默认值 10。
// @Tags Platforms
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量 (<= 0 表示获取所有)" default(10)
// @Param orderBy query string false "排序字段 (e.g., name, website_url, created_at, updated_at)" default(name)
// @Param sortDirection query string false "排序方向 (asc, desc)" default(asc)
// @Param name query string false "按平台名称进行模糊匹配筛选"
// @Success 200 {object} models.SuccessResponse{data=[]models.PlatformResponse,meta=models.PaginationMeta} "获取成功"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /platforms [get]
// @Security BearerAuth
func GetPlatforms(c *gin.Context) {
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

	orderBy := c.DefaultQuery("orderBy", "name")
	sortDirection := c.DefaultQuery("sortDirection", "asc")
	filterName := strings.ToLower(strings.TrimSpace(c.Query("name")))

	// Validate orderBy parameter
	allowedOrderByFields := map[string]string{
		"name":        "name",
		"website_url": "website_url",
		"notes":       "notes",
		"created_at":  "created_at",
		"updated_at":  "updated_at",
	}
	dbOrderByField, isValidField := allowedOrderByFields[orderBy]
	if !isValidField {
		dbOrderByField = "name" // Default to a safe field
	}

	// Validate sortDirection
	if strings.ToLower(sortDirection) != "asc" && strings.ToLower(sortDirection) != "desc" {
		sortDirection = "asc" // Default to asc
	}
	orderClause := dbOrderByField + " " + sortDirection

	var page int
	var pageSizeForQuery int // Used for Offset/Limit if pagination is active, and potentially for meta
	fetchAllWindows := false
	defaultPageSize := 10

	pageQuery := c.DefaultQuery("page", "1")
	page, _ = strconv.Atoi(pageQuery) // Error ignored for simplicity
	if page <= 0 {
		page = 1
	}

	pageSizeQuery := c.Query("pageSize")
	if pageSizeQuery == "" {
		// 3. pageSize 未传递: 使用默认值
		pageSizeForQuery = defaultPageSize
		fetchAllWindows = false
	} else {
		parsedPageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			// 3. pageSize 无法解析: 使用默认值
			pageSizeForQuery = defaultPageSize
			fetchAllWindows = false
		} else {
			if parsedPageSize > 0 {
				// 1. pageSize > 0: 使用传递的值
				pageSizeForQuery = parsedPageSize
				fetchAllWindows = false
			} else { // parsedPageSize <= 0 (包括 0 和负数)
				// 2. pageSize <= 0: 获取所有记录
				fetchAllWindows = true
				// pageSizeForQuery 在 fetchAllWindows 为 true 时不直接用于查询 LIMIT。
				// 它将在后面用于 meta data，并根据 totalRecords 进行调整。
				// 暂时设为 0 表示“全部”。
				pageSizeForQuery = 0 // Indicates "all", will be adjusted for meta later
			}
		}
	}

	var offset int
	if !fetchAllWindows {
		// 确保 pageSizeForQuery 是正数，用于 offset 计算和 Limit
		// 如果 fetchAllWindows 为 false，pageSizeForQuery 此时应为 > 0 的用户值或默认值
		if pageSizeForQuery <= 0 { // 安全检查，理论上不应发生
			pageSizeForQuery = defaultPageSize
		}
		offset = (page - 1) * pageSizeForQuery
	}
	// 如果 fetchAllWindows 为 true, offset 不用于主查询。

	var platforms []models.Platform
	var totalRecords int64

	query := database.DB.Model(&models.Platform{}).Where("user_id = ?", currentUserID)

	// Apply name filter if provided
	if filterName != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+filterName+"%")
	}

	if err := query.Count(&totalRecords).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台总数失败: "+err.Error())
		return
	}

	finalQuery := query.Order(orderClause)
	if !fetchAllWindows {
		// Apply pagination only if not fetching all items.
		// pageSizeForQuery here would be the user-defined positive value, or default 10.
		finalQuery = finalQuery.Offset(offset).Limit(pageSizeForQuery)
	}
	// If fetchAllWindows is true, no Offset or Limit is applied, getting all records.

	if err := finalQuery.Find(&platforms).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台列表失败: "+err.Error())
		return
	}

	var responses []models.PlatformResponse
	for _, p := range platforms {
		var emailAccountCount int64
		// 计算当前用户在此平台上有有效邮箱地址的注册数量 (JOIN with email_accounts)
		if err := database.DB.Model(&models.PlatformRegistration{}).
			Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id").
			Where("platform_registrations.platform_id = ? AND platform_registrations.user_id = ?", p.ID, currentUserID).
			Where("email_accounts.email_address IS NOT NULL AND email_accounts.email_address <> ''").
			Count(&emailAccountCount).Error; err != nil {
			emailAccountCount = 0 // Keep count as 0 on error
		}
		response := p.ToPlatformResponse()
		response.EmailAccountCount = emailAccountCount // 这个字段现在表示当前用户在该平台注册的邮箱数
		responses = append(responses, response)
	}

	var metaPage, metaPageSizeForResponse int
	if fetchAllWindows {
		metaPage = 1 // 当获取所有项目时，它实际上是第一页也是唯一一页
		// 当获取所有记录时，响应中的 pageSize 应为总记录数
		metaPageSizeForResponse = int(totalRecords)
	} else {
		metaPage = page
		// 当分页时，响应中的 pageSize 是用于查询的 pageSize (用户提供或默认值)
		// 此时 pageSizeForQuery 保证是 > 0 的值
		metaPageSizeForResponse = pageSizeForQuery
	}
	pagination := utils.CreatePaginationMeta(metaPage, metaPageSizeForResponse, int(totalRecords))
	utils.SendSuccessResponseWithMeta(c, responses, pagination)
}

// GetPlatformByID godoc
// @Summary 获取指定ID的平台详情
// @Description 获取指定ID的平台详情
// @Tags Platforms
// @Produce json
// @Param id path int true "平台ID"
// @Success 200 {object} models.SuccessResponse{data=models.PlatformResponse} "获取成功"
// @Failure 400 {object} models.ErrorResponse "无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "平台未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /platforms/{id} [get]
// @Security BearerAuth
func GetPlatformByID(c *gin.Context) {
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

	platformID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的平台ID格式")
		return
	}

	var platform models.Platform
	if err := database.DB.Where("id = ? AND user_id = ?", platformID, currentUserID).First(&platform).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "平台未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台详情失败: "+err.Error())
		return
	}
	// 计算当前用户在此平台上的注册数量
	var emailAccountCount int64
	// 计算当前用户在此平台上有有效邮箱地址的注册数量 (JOIN with email_accounts)
	if errDb := database.DB.Model(&models.PlatformRegistration{}).
		Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id").
		Where("platform_registrations.platform_id = ? AND platform_registrations.user_id = ?", platform.ID, currentUserID).
		Where("email_accounts.email_address IS NOT NULL AND email_accounts.email_address <> ''").
		Count(&emailAccountCount).Error; errDb != nil {
		emailAccountCount = 0 // Keep count as 0 on error
	}
	response := platform.ToPlatformResponse()
	response.EmailAccountCount = emailAccountCount
	utils.SendSuccessResponse(c, response)
}

// UpdatePlatform godoc
// @Summary 更新指定ID的平台信息
// @Description 更新指定ID的平台信息
// @Tags Platforms
// @Accept json
// @Produce json
// @Param id path int true "平台ID"
// @Param platform body models.Platform true "要更新的平台信息，ID 会被忽略"
// @Success 200 {object} models.SuccessResponse{data=models.PlatformResponse} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误或无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "平台未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /platforms/{id} [put]
// @Security BearerAuth
func UpdatePlatform(c *gin.Context) {
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

	platformID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的平台ID格式")
		return
	}

	var platform models.Platform
	if err := database.DB.Where("id = ? AND user_id = ?", platformID, currentUserID).First(&platform).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "平台未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询待更新平台失败: "+err.Error())
		return
	}

	var input struct {
		Name       string `json:"name" binding:"omitempty,min=2,max=100"`
		WebsiteURL string `json:"website_url" binding:"omitempty,url"`
		Notes      string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	if input.Name != "" {
		platform.Name = input.Name
	}
	// Allow clearing WebsiteURL and Notes by providing them, even if empty
	if c.GetHeader("Content-Type") == "application/json" { // Check if fields were actually sent
		if _, ok := c.GetPostForm("website_url"); ok || input.WebsiteURL != "" || platform.WebsiteURL != "" && input.WebsiteURL == "" && c.Request.ContentLength > 0 { // More robust check needed
			// A bit complex to detect if a field was explicitly sent as empty string vs not sent.
			// For simplicity, if omitempty is not used, it means we always update.
			// If omitempty is used, empty means "don't update".
			// Here, we assume if name is sent, other fields are also intended for update.
		}
	}
	// Simpler: always update if field is in input struct, even if empty string (allows clearing)
	// For omitempty fields, they won't be in 'input' if not sent or empty.
	// Let's adjust binding and logic for clarity.

	updateData := make(map[string]interface{})
	if input.Name != "" {
		updateData["name"] = input.Name
	}
	if input.WebsiteURL != "" || (input.WebsiteURL == "" && platform.WebsiteURL != "") { // Allow clearing
		updateData["website_url"] = input.WebsiteURL
	}
	// For notes, always update from input if provided
	// This logic for partial updates can be tricky. GORM's Update vs Updates.
	// Using struct for updates:
	updates := models.Platform{Notes: input.Notes} // Default to updating notes
	if input.Name != "" {
		updates.Name = input.Name
	}
	if input.WebsiteURL != "" || (input.WebsiteURL == "" && platform.WebsiteURL != "") { // If input URL is empty and existing is not, it's an attempt to clear
		updates.WebsiteURL = input.WebsiteURL
	} else if input.WebsiteURL == "" && platform.WebsiteURL == "" {
		// If both are empty, do nothing for URL to avoid GORM trying to set it to empty again if not in struct
	} else if input.WebsiteURL != "" { // If input URL is not empty, update
		updates.WebsiteURL = input.WebsiteURL
	}
	// A cleaner way for partial updates with GORM is to use a map[string]interface{}
	// or to fetch the record, modify fields, and then Save().

	// Let's use the fetch, modify, save pattern for clarity
	if input.Name != "" {
		platform.Name = input.Name
	}
	platform.WebsiteURL = input.WebsiteURL // Allow clearing
	platform.Notes = input.Notes           // Allow clearing

	if err := database.DB.Save(&platform).Error; err != nil {
		if utils.IsUniqueConstraintError(err) {
			utils.SendErrorResponse(c, http.StatusConflict, "您已创建过同名平台")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "更新平台失败: "+err.Error())
		return
	}
	// 计算当前用户在此平台上的注册数量
	var emailAccountCount int64
	// 计算当前用户在此平台上有有效邮箱地址的注册数量 (JOIN with email_accounts)
	if errDb := database.DB.Model(&models.PlatformRegistration{}).
		Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id").
		Where("platform_registrations.platform_id = ? AND platform_registrations.user_id = ?", platform.ID, currentUserID).
		Where("email_accounts.email_address IS NOT NULL AND email_accounts.email_address <> ''").
		Count(&emailAccountCount).Error; errDb != nil {
		emailAccountCount = 0 // Keep count as 0 on error
	}
	response := platform.ToPlatformResponse()
	response.EmailAccountCount = emailAccountCount
	utils.SendSuccessResponse(c, response)
}

// DeletePlatform godoc
// @Summary 删除指定ID的平台
// @Description 删除指定ID的平台信息
// @Tags Platforms
// @Produce json
// @Param id path int true "平台ID"
// @Success 200 {object} models.SuccessResponse "删除成功"
// @Failure 400 {object} models.ErrorResponse "无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "平台未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误 (例如，如果平台仍被引用)"
// @Router /platforms/{id} [delete]
// @Security BearerAuth
func DeletePlatform(c *gin.Context) {
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

	platformID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的平台ID格式")
		return
	}

	var platform models.Platform
	if err := database.DB.Where("id = ? AND user_id = ?", platformID, currentUserID).First(&platform).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "平台未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询待删除平台失败: "+err.Error())
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
	// 注意：由于 Platform 现在是用户特定的，我们只删除当前用户与此平台相关的注册信息
	if err := tx.Where("platform_id = ? AND user_id = ?", platform.ID, currentUserID).Find(&registrations).Error; err != nil {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查找关联的平台注册信息失败: "+err.Error())
		return
	}

	for _, reg := range registrations {
		// 1a. 硬删除关联的 ServiceSubscriptions
		// ServiceSubscription 也应该有 UserID，确保只删除当前用户的订阅
		if err := tx.Unscoped().Where("platform_registration_id = ? AND user_id = ?", reg.ID, currentUserID).Delete(&models.ServiceSubscription{}).Error; err != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "删除服务订阅失败: "+err.Error())
			return
		}
		// 1b. 硬删除 PlatformRegistration
		if err := tx.Unscoped().Delete(&reg).Error; err != nil { // reg 已经包含了 UserID，所以 GORM 的钩子或条件应该能正确处理
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "删除平台注册信息失败: "+err.Error())
			return
		}
	}

	// 2. 硬删除 Platform 本身
	if err := tx.Unscoped().Delete(&platform).Error; err != nil { // platform 已经通过 platformID 和 currentUserID 查询得到，是正确的记录
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "删除平台失败: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "平台及关联信息删除成功"})
}
