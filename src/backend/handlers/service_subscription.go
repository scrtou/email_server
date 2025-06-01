package handlers

import (
	"email_server/database"
	"email_server/models"
	"email_server/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateServiceSubscription godoc
// @Summary 创建服务订阅
// @Description 为当前用户创建一个新的服务订阅，通过平台名称和邮箱地址关联或创建平台和邮箱账户信息
// @Tags ServiceSubscriptions
// @Accept json
// @Produce json
// @Param serviceSubscription body object{platform_name=string,email_address=string,service_name=string,description=string,status=string,cost=float64,billing_cycle=string,next_renewal_date=string,payment_method_notes=string} true "服务订阅信息，包含平台名称和邮箱地址。 platform_name, email_address, service_name, status, billing_cycle 为必填项。"
// @Success 201 {object} models.SuccessResponse{data=models.ServiceSubscriptionResponse} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误或关联资源无效"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /service-subscriptions [post]
// @Security BearerAuth
func CreateServiceSubscription(c *gin.Context) {
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

	var input models.CreateServiceSubscriptionRequest // 使用新的 DTO

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// 1. 查找或创建 EmailAccount
	var emailAccount models.EmailAccount
	err := database.DB.Where("email_address = ? AND user_id = ?", input.EmailAddress, currentUserID).First(&emailAccount).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			emailAccount = models.EmailAccount{
				UserID:       currentUserID,
				EmailAddress: input.EmailAddress,
				Provider:     "", // Per requirement, set Provider to empty string
			}
			if createErr := database.DB.Create(&emailAccount).Error; createErr != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "创建邮箱账户失败: "+createErr.Error())
				return
			}
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询邮箱账户失败: "+err.Error())
			return
		}
	}

	// 2. 查找或创建 Platform
	var platform models.Platform
	err = database.DB.Where("name = ? AND user_id = ?", input.PlatformName, currentUserID).First(&platform).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			platform = models.Platform{
				Name:   input.PlatformName,
				UserID: currentUserID, // Assign current user's ID
			}
			if createErr := database.DB.Create(&platform).Error; createErr != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "创建平台失败: "+createErr.Error())
				return
			}
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台失败: "+err.Error())
			return
		}
	}

	// 3. 查找或创建 PlatformRegistration
	var platformRegistration models.PlatformRegistration
	// Try to find existing PlatformRegistration and preload its associations
	err = database.DB.Where("user_id = ? AND email_account_id = ? AND platform_id = ?", currentUserID, emailAccount.ID, platform.ID).
		Preload("Platform").Preload("EmailAccount").First(&platformRegistration).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Not found, create a new one
			platformRegistration = models.PlatformRegistration{
				UserID:         currentUserID,
				EmailAccountID: emailAccount.ID,
				PlatformID:     platform.ID,
				// Manually assign the Platform and EmailAccount objects for the response structure
				// as GORM might not link them immediately on create for the current object instance
				Platform:     platform,
				EmailAccount: emailAccount,
			}
			if createErr := database.DB.Create(&platformRegistration).Error; createErr != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "创建平台注册信息失败: "+createErr.Error())
				return
			}
			// After creation, platformRegistration.ID is populated.
			// The Platform and EmailAccount fields were manually set above.
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台注册信息失败: "+err.Error())
			return
		}
	}
	// If found, Preload("Platform").Preload("EmailAccount") should have populated them.
	// If created, we manually assigned platform and emailAccount to platformRegistration.Platform and platformRegistration.EmailAccount.

	var nextRenewalDate *time.Time
	if input.NextRenewalDateStr != nil && *input.NextRenewalDateStr != "" {
		parsedDate, errDate := time.Parse("2006-01-02", *input.NextRenewalDateStr)
		if errDate != nil {
			utils.SendErrorResponse(c, http.StatusBadRequest, "下次续费日期格式无效，请使用 YYYY-MM-DD")
			return
		}
		nextRenewalDate = &parsedDate
	}

	// 4. 查找或创建/恢复 ServiceSubscription 记录
	var subscription models.ServiceSubscription
	err = database.DB.Unscoped().Where(models.ServiceSubscription{
		UserID:                 currentUserID,
		PlatformRegistrationID: platformRegistration.ID, // Use the ID from the found/created PlatformRegistration
		ServiceName:            input.ServiceName,
	}).First(&subscription).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound { // 记录不存在，创建新的
			subscription = models.ServiceSubscription{
				UserID:                 currentUserID,
				PlatformRegistrationID: platformRegistration.ID,
				ServiceName:            input.ServiceName,
				Description:            input.Description,
				Status:                 input.Status,
				Cost:                   input.Cost,
				BillingCycle:           input.BillingCycle,
				NextRenewalDate:        nextRenewalDate,
				PaymentMethodNotes:     input.PaymentMethodNotes,
			}
			if createErr := database.DB.Create(&subscription).Error; createErr != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "创建服务订阅失败: "+createErr.Error())
				return
			}
		} else { // 其他查询错误
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询服务订阅失败: "+err.Error())
			return
		}
	} else { // 找到了记录
		if subscription.DeletedAt.Valid { // 如果是软删除的记录
			subscription.Description = input.Description
			subscription.Status = input.Status
			subscription.Cost = input.Cost
			subscription.BillingCycle = input.BillingCycle
			subscription.NextRenewalDate = nextRenewalDate
			subscription.PaymentMethodNotes = input.PaymentMethodNotes
			subscription.DeletedAt = gorm.DeletedAt{} // 重置软删除标记
			if updateErr := database.DB.Unscoped().Save(&subscription).Error; updateErr != nil {
				utils.SendErrorResponse(c, http.StatusInternalServerError, "恢复并更新服务订阅失败: "+updateErr.Error())
				return
			}
		} else { // 记录已存在且未被软删除
			utils.SendErrorResponse(c, http.StatusConflict, "该服务订阅已存在。")
			return
		}
	}

	// platformRegistration should have .Platform and .EmailAccount populated for the response
	response := subscription.ToServiceSubscriptionResponse(platformRegistration, platformRegistration.Platform, platformRegistration.EmailAccount)
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": response})
}

// GetServiceSubscriptions godoc
// @Summary 获取当前用户的所有服务订阅
// @Description 获取当前登录用户的所有服务订阅，支持分页和筛选
// @Tags ServiceSubscriptions
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param platform_registration_id query int false "按平台注册ID筛选"
// @Param status query string false "按订阅状态筛选 (e.g., active, inactive, expired)"
// @Param billing_cycle query string false "按计费周期筛选 (e.g., monthly, annually)"
// @Param renewal_date_start query string false "续费日期开始 (YYYY-MM-DD)"
// @Param renewal_date_end query string false "续费日期结束 (YYYY-MM-DD)"
// @Param orderBy query string false "排序字段 (e.g., service_name, status, cost, next_renewal_date, created_at)" default(created_at)
// @Param sortDirection query string false "排序方向 (asc, desc)" default(desc)
// @Success 200 {object} models.SuccessResponse{data=[]models.ServiceSubscriptionResponse,meta=models.PaginationMeta} "获取成功"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /service-subscriptions [get]
// @Security BearerAuth
func GetServiceSubscriptions(c *gin.Context) {
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
	prIDFilter, _ := strconv.Atoi(c.Query("platform_registration_id"))
	statusFilter := strings.ToLower(strings.TrimSpace(c.Query("status")))
	billingCycleFilter := strings.ToLower(strings.TrimSpace(c.Query("billing_cycle")))
	renewalDateStartStr := strings.TrimSpace(c.Query("renewal_date_start"))
	renewalDateEndStr := strings.TrimSpace(c.Query("renewal_date_end"))
	orderBy := c.DefaultQuery("orderBy", "created_at")
	sortDirection := c.DefaultQuery("sortDirection", "desc")

	// Validate orderBy parameter
	allowedOrderByFields := map[string]string{
		"service_name":      "service_name",
		"status":            "status",
		"cost":              "cost",
		"billing_cycle":     "billing_cycle",
		"next_renewal_date": "next_renewal_date",
		"created_at":        "created_at",
		"updated_at":        "updated_at",
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

	if page <= 0 { page = 1 }
	if pageSize <= 0 { pageSize = 10 }
	if pageSize > 100 { pageSize = 100 }
	offset := (page - 1) * pageSize

	var subscriptions []models.ServiceSubscription
	var totalRecords int64

	query := database.DB.Model(&models.ServiceSubscription{}).Where("user_id = ?", currentUserID)
	countQuery := database.DB.Model(&models.ServiceSubscription{}).Where("user_id = ?", currentUserID)

	if prIDFilter > 0 {
		query = query.Where("platform_registration_id = ?", prIDFilter)
		countQuery = countQuery.Where("platform_registration_id = ?", prIDFilter)
	}
	if statusFilter != "" {
		query = query.Where("LOWER(status) = ?", statusFilter)
		countQuery = countQuery.Where("LOWER(status) = ?", statusFilter)
	}
	if billingCycleFilter != "" {
		query = query.Where("LOWER(billing_cycle) = ?", billingCycleFilter)
		countQuery = countQuery.Where("LOWER(billing_cycle) = ?", billingCycleFilter)
	}

	if renewalDateStartStr != "" {
		startDate, err := time.Parse("2006-01-02", renewalDateStartStr)
		if err == nil {
			query = query.Where("next_renewal_date >= ?", startDate)
			countQuery = countQuery.Where("next_renewal_date >= ?", startDate)
		} else {
			// Optionally handle or log date parsing error for filter
		}
	}
	if renewalDateEndStr != "" {
		endDate, err := time.Parse("2006-01-02", renewalDateEndStr)
		if err == nil {
			// To include the end date, typically query for dates less than the day after
			endDate = endDate.AddDate(0, 0, 1)
			query = query.Where("next_renewal_date < ?", endDate)
			countQuery = countQuery.Where("next_renewal_date < ?", endDate)
		} else {
			// Optionally handle or log date parsing error for filter
		}
	}
	
	if err := countQuery.Count(&totalRecords).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取服务订阅总数失败: "+err.Error())
		return
	}

	// Preload PlatformRegistration and its nested Platform and EmailAccount
	err := query.Order(orderClause).Offset(offset).Limit(pageSize).
		Preload("PlatformRegistration.Platform").
		Preload("PlatformRegistration.EmailAccount").
		Find(&subscriptions).Error
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取服务订阅列表失败: "+err.Error())
		return
	}

	var responses []models.ServiceSubscriptionResponse
	for _, ss := range subscriptions {
		responses = append(responses, ss.ToServiceSubscriptionResponse(ss.PlatformRegistration, ss.PlatformRegistration.Platform, ss.PlatformRegistration.EmailAccount))
	}
	
	pagination := utils.CreatePaginationMeta(page, pageSize, int(totalRecords))
	utils.SendSuccessResponseWithMeta(c, responses, pagination)
}

// GetServiceSubscriptionByID godoc
// @Summary 获取指定ID的服务订阅详情
// @Description 获取当前用户拥有的指定ID的服务订阅详情
// @Tags ServiceSubscriptions
// @Produce json
// @Param id path int true "服务订阅ID"
// @Success 200 {object} models.SuccessResponse{data=models.ServiceSubscriptionResponse} "获取成功"
// @Failure 400 {object} models.ErrorResponse "无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "服务订阅未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /service-subscriptions/{id} [get]
// @Security BearerAuth
func GetServiceSubscriptionByID(c *gin.Context) {
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

	subscriptionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的服务订阅ID格式")
		return
	}

	var ss models.ServiceSubscription
	err = database.DB.Where("id = ? AND user_id = ?", subscriptionID, currentUserID).
		Preload("PlatformRegistration.Platform").
		Preload("PlatformRegistration.EmailAccount").
		First(&ss).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "服务订阅未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取服务订阅详情失败: "+err.Error())
		return
	}
	
	response := ss.ToServiceSubscriptionResponse(ss.PlatformRegistration, ss.PlatformRegistration.Platform, ss.PlatformRegistration.EmailAccount)
	utils.SendSuccessResponse(c, response)
}

// UpdateServiceSubscription godoc
// @Summary 更新指定ID的服务订阅
// @Description 更新当前用户拥有的指定ID的服务订阅信息。平台和邮箱不可更改。
// @Tags ServiceSubscriptions
// @Accept json
// @Produce json
// @Param id path int true "服务订阅ID"
// @Param serviceSubscription body object{service_name=string,description=string,status=string,cost=float64,billing_cycle=string,next_renewal_date=string,payment_method_notes=string} true "要更新的服务订阅信息。UserID, PlatformRegistrationID, PlatformName, EmailAddress 不可更改。"
// @Success 200 {object} models.SuccessResponse{data=models.ServiceSubscriptionResponse} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误、无效的ID格式或尝试修改不可变字段"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "服务订阅未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /service-subscriptions/{id} [put]
// @Security BearerAuth
func UpdateServiceSubscription(c *gin.Context) {
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

	subscriptionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的服务订阅ID格式")
		return
	}

	var ss models.ServiceSubscription
	err = database.DB.Where("id = ? AND user_id = ?", subscriptionID, currentUserID).
		Preload("PlatformRegistration.Platform").
		Preload("PlatformRegistration.EmailAccount").
		First(&ss).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "服务订阅未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询待更新服务订阅失败: "+err.Error())
		return
	}

	var rawInput map[string]interface{}
	if err := c.ShouldBindJSON(&rawInput); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// 检查是否尝试修改不可变字段
	if _, rok := rawInput["platform_name"]; rok {
		utils.SendErrorResponse(c, http.StatusBadRequest, "不允许修改平台名称(platform_name)")
		return
	}
	if _, rok := rawInput["email_address"]; rok {
		utils.SendErrorResponse(c, http.StatusBadRequest, "不允许修改邮箱地址(email_address)")
		return
	}
	if _, rok := rawInput["platform_registration_id"]; rok {
		utils.SendErrorResponse(c, http.StatusBadRequest, "不允许修改平台注册ID(platform_registration_id)")
		return
	}

	// 更新允许修改的字段
	updated := false // Track if any field is actually updated

	if val, rok := rawInput["service_name"]; rok {
		if strVal, okAssert := val.(string); okAssert {
			// Add validation if service_name cannot be empty, e.g.
			// if strVal == "" { utils.SendErrorResponse(c, http.StatusBadRequest, "服务名称不能为空"); return }
			if ss.ServiceName != strVal {
				ss.ServiceName = strVal
				updated = true
			}
		} else {
			// utils.SendErrorResponse(c, http.StatusBadRequest, "service_name 格式错误")
			// return
		}
	}

	if val, rok := rawInput["description"]; rok { // Allows null or empty string to clear
		currentDesc := ss.Description
		newDesc := ""
		if strVal, okAssert := val.(string); okAssert {
			newDesc = strVal
		} else if val == nil {
			newDesc = "" // Treat null as empty string for description
		}
		if currentDesc != newDesc {
			ss.Description = newDesc
			updated = true
		}
	}

	if val, rok := rawInput["status"]; rok {
		if strVal, okAssert := val.(string); okAssert && strVal != "" { // Status usually required
			if ss.Status != strVal {
				ss.Status = strVal
				updated = true
			}
		}
	}

	if val, rok := rawInput["cost"]; rok { // Allows 0 as a valid cost
		if numVal, okAssert := val.(float64); okAssert {
			// Add validation if cost cannot be negative, e.g.
			// if numVal < 0 { utils.SendErrorResponse(c, http.StatusBadRequest, "cost 不能为负数"); return }
			if ss.Cost != numVal {
				ss.Cost = numVal
				updated = true
			}
		}
	}

	if val, rok := rawInput["billing_cycle"]; rok {
		if strVal, okAssert := val.(string); okAssert && strVal != "" { // Billing cycle usually required
			if ss.BillingCycle != strVal {
				ss.BillingCycle = strVal
				updated = true
			}
		}
	}

	if val, rok := rawInput["payment_method_notes"]; rok {
		currentPmn := ss.PaymentMethodNotes
		newPmn := ""
		if strVal, okAssert := val.(string); okAssert {
			newPmn = strVal
		} else if val == nil {
			newPmn = "" // Treat null as empty string
		}
		if currentPmn != newPmn {
			ss.PaymentMethodNotes = newPmn
			updated = true
		}
	}

	if val, rok := rawInput["next_renewal_date"]; rok {
		var newDate *time.Time
		changed := false
		if val == nil { // Explicitly set to null
			if ss.NextRenewalDate != nil { // Only change if it was not already null
				newDate = nil
				changed = true
			}
		} else if strVal, okAssert := val.(string); okAssert {
			if strVal == "" { // Empty string also clears the date
				if ss.NextRenewalDate != nil {
					newDate = nil
					changed = true
				}
			} else {
				parsedDate, errDate := time.Parse("2006-01-02", strVal)
				if errDate != nil {
					utils.SendErrorResponse(c, http.StatusBadRequest, "下次续费日期格式无效，请使用 YYYY-MM-DD")
					return
				}
				// Compare with existing date if it exists
				if ss.NextRenewalDate == nil || !ss.NextRenewalDate.Equal(parsedDate) {
					newDate = &parsedDate
					changed = true
				} else if ss.NextRenewalDate != nil && ss.NextRenewalDate.Equal(parsedDate) {
                    // No change if dates are the same
                }
			}
		} else {
			// utils.SendErrorResponse(c, http.StatusBadRequest, "next_renewal_date 格式错误")
			// return
		}
		if changed {
			ss.NextRenewalDate = newDate
			updated = true
		}
	}
	
	if updated { // Only save if there were actual changes to prevent unnecessary DB write and updated_at bump
		if err := database.DB.Save(&ss).Error; err != nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "更新服务订阅失败: "+err.Error())
			return
		}
	}
	
	response := ss.ToServiceSubscriptionResponse(ss.PlatformRegistration, ss.PlatformRegistration.Platform, ss.PlatformRegistration.EmailAccount)
	utils.SendSuccessResponse(c, response)
}

// DeleteServiceSubscription godoc
// @Summary 删除指定ID的服务订阅
// @Description 删除当前用户拥有的指定ID的服务订阅
// @Tags ServiceSubscriptions
// @Produce json
// @Param id path int true "服务订阅ID"
// @Success 200 {object} models.SuccessResponse "删除成功"
// @Failure 400 {object} models.ErrorResponse "无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该资源"
// @Failure 404 {object} models.ErrorResponse "服务订阅未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /service-subscriptions/{id} [delete]
// @Security BearerAuth
func DeleteServiceSubscription(c *gin.Context) {
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

	subscriptionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的服务订阅ID格式")
		return
	}

	var ss models.ServiceSubscription
	if err := database.DB.Where("id = ? AND user_id = ?", subscriptionID, currentUserID).First(&ss).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "服务订阅未找到或无权访问")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询待删除服务订阅失败: "+err.Error())
		return
	}
	
	if err := database.DB.Delete(&ss).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "删除服务订阅失败: "+err.Error())
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "服务订阅删除成功"})
}

// GetServiceSubscriptionsByPlatformRegistrationID godoc
// @Summary 获取指定平台注册信息关联的所有服务订阅
// @Description 获取当前用户拥有的指定平台注册信息所关联的所有服务订阅
// @Tags ServiceSubscriptions
// @Produce json
// @Param id path int true "平台注册ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param orderBy query string false "排序字段 (e.g., service_name, status, cost, next_renewal_date, created_at)" default(created_at)
// @Param sortDirection query string false "排序方向 (asc, desc)" default(desc)
// @Success 200 {object} models.SuccessResponse{data=[]models.ServiceSubscriptionResponse,meta=models.PaginationMeta} "获取成功"
// @Failure 400 {object} models.ErrorResponse "无效的ID格式"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 403 {object} models.ErrorResponse "无权访问该平台注册信息"
// @Failure 404 {object} models.ErrorResponse "平台注册信息未找到"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /platform-registrations/{id}/service-subscriptions [get]
// @Security BearerAuth
func GetServiceSubscriptionsByPlatformRegistrationID(c *gin.Context) {
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

	platformRegistrationIDParam := c.Param("id")
	platformRegistrationID, err := strconv.ParseUint(platformRegistrationIDParam, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "无效的平台注册ID格式")
		return
	}

	// 验证平台注册信息是否属于当前用户
	var pr models.PlatformRegistration
	if err := database.DB.Where("id = ? AND user_id = ?", platformRegistrationID, uint(currentUserID)).
		Preload("Platform").Preload("EmailAccount").First(&pr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusForbidden, "无权访问该平台注册信息或平台注册信息不存在")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台注册信息失败: "+err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	orderBy := c.DefaultQuery("orderBy", "created_at")
	sortDirection := c.DefaultQuery("sortDirection", "desc")

	// Validate orderBy parameter (same as GetServiceSubscriptions)
	allowedOrderByFields := map[string]string{
		"service_name":      "service_name",
		"status":            "status",
		"cost":              "cost",
		"billing_cycle":     "billing_cycle",
		"next_renewal_date": "next_renewal_date",
		"created_at":        "created_at",
		"updated_at":        "updated_at",
	}
	dbOrderByField, isValidField := allowedOrderByFields[orderBy]
	if !isValidField {
		dbOrderByField = "created_at"
	}
	if strings.ToLower(sortDirection) != "asc" && strings.ToLower(sortDirection) != "desc" {
		sortDirection = "desc"
	}
	orderClause := dbOrderByField + " " + sortDirection

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

	var subscriptions []models.ServiceSubscription
	var totalRecords int64

	dbQuery := database.DB.Model(&models.ServiceSubscription{}).Where("platform_registration_id = ? AND user_id = ?", platformRegistrationID, uint(currentUserID))

	// Count total records
	if err := dbQuery.Count(&totalRecords).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "统计关联服务订阅总数失败: "+err.Error())
		return
	}

	// Fetch paginated records
	// Preload PlatformRegistration and its nested Platform and EmailAccount for the response
	// Since we already have `pr` (PlatformRegistration with its preloads), we can use it.
	// However, the subscriptions themselves need to be fetched.
	err = database.DB.Model(&models.ServiceSubscription{}).
		Where("platform_registration_id = ? AND user_id = ?", platformRegistrationID, uint(currentUserID)).
		Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&subscriptions).Error
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取服务订阅信息失败: "+err.Error())
		return
	}

	var responses []models.ServiceSubscriptionResponse
	for _, ss := range subscriptions {
		// For each subscription, we use the already fetched `pr` for its related data.
		responses = append(responses, ss.ToServiceSubscriptionResponse(pr, pr.Platform, pr.EmailAccount))
	}

	pagination := utils.CreatePaginationMeta(page, pageSize, int(totalRecords))
	utils.SendSuccessResponseWithMeta(c, responses, pagination)
}