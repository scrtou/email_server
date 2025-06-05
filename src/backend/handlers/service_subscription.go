package handlers

import (
	"email_server/database"
	"email_server/models"
	"email_server/utils"
	"fmt"
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
// @Param serviceSubscription body object{platform_id=uint,platform_name=string,email_address=string,email_account_id=uint,platform_registration_id=uint,selected_username_registration_id=uint,login_username=string,service_name=string,description=string,status=string,cost=float64,billing_cycle=string,next_renewal_date=string,payment_method_notes=string} true "服务订阅信息"
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

	var input models.CreateServiceSubscriptionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// --- 数据库事务开始 ---
	tx := database.DB.Begin()
	if tx.Error != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "开启数据库事务失败: "+tx.Error.Error())
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if tx.Error != nil {
			// tx.Rollback() // 已经被 tx.Commit() 尝试处理或已回滚
		}
	}()

	// 根据数据处理原则简化逻辑
	// 1. 查询前端传递的平台名称是否存在，不存在则新增平台记录
	var platform models.Platform
	if input.PlatformName == "" {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusBadRequest, "平台名称不能为空")
		return
	}

	// 查询平台
	err := tx.Where("name = ? AND user_id = ?", input.PlatformName, currentUserID).First(&platform).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 平台不存在，创建新平台
			platform = models.Platform{
				Name:   input.PlatformName,
				UserID: currentUserID,
			}
			if createErr := tx.Create(&platform).Error; createErr != nil {
				tx.Rollback()
				utils.SendErrorResponse(c, http.StatusInternalServerError, "创建新平台失败: "+createErr.Error())
				return
			}
		} else {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台失败: "+err.Error())
			return
		}
	}

	// 2. 根据数据处理原则的三种情况处理
	emailAddress := input.EmailAddress
	loginUsername := input.LoginUsername

	var platformReg models.PlatformRegistration
	var emailAccount models.EmailAccount

	if emailAddress == "" && loginUsername != "" {
		// 情况1: 邮箱地址为空，用户名不为空
		err := createServiceSubscriptionCase1(tx, currentUserID, platform, loginUsername, &platformReg)
		if err != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
	} else if emailAddress != "" && loginUsername == "" {
		// 情况2: 邮箱地址不为空，用户名为空
		err := createServiceSubscriptionCase2(tx, currentUserID, platform, emailAddress, &platformReg, &emailAccount)
		if err != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
	} else if emailAddress != "" && loginUsername != "" {
		// 情况3: 邮箱地址和用户名都不为空
		err := createServiceSubscriptionCase3(tx, currentUserID, platform, emailAddress, loginUsername, &platformReg, &emailAccount)
		if err != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
	} else {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusBadRequest, "邮箱地址和用户名不能都为空")
		return
	}

	// 3. 处理下次续费日期
	var nextRenewalDate *time.Time
	if input.NextRenewalDateStr != nil && *input.NextRenewalDateStr != "" {
		parsedDate, errDate := time.Parse("2006-01-02", *input.NextRenewalDateStr)
		if errDate != nil {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusBadRequest, "下次续费日期格式无效，请使用 YYYY-MM-DD")
			return
		}
		nextRenewalDate = &parsedDate
	}

	// 4. 创建服务订阅记录
	var subscription models.ServiceSubscription
	err = tx.Where(models.ServiceSubscription{
		UserID:                 currentUserID,
		PlatformRegistrationID: platformReg.ID,
		ServiceName:            input.ServiceName,
	}).First(&subscription).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			subscription = models.ServiceSubscription{
				UserID:                 currentUserID,
				PlatformRegistrationID: platformReg.ID,
				ServiceName:            input.ServiceName,
				Description:            input.Description,
				Status:                 input.Status,
				Cost:                   input.Cost,
				BillingCycle:           input.BillingCycle,
				NextRenewalDate:        nextRenewalDate,
				PaymentMethodNotes:     input.PaymentMethodNotes,
			}
			if createErr := tx.Create(&subscription).Error; createErr != nil {
				tx.Rollback()
				utils.SendErrorResponse(c, http.StatusInternalServerError, "创建服务订阅失败: "+createErr.Error())
				return
			}
		} else {
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询服务订阅失败: "+err.Error())
			return
		}
	} else {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusConflict, "该服务订阅已存在。")
		return
	}

	// --- 提交事务 ---
	if err := tx.Commit().Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}

	// 准备响应数据
	finalPlatformForResp := platform
	finalEmailAccountForResp := emailAccount

	response := subscription.ToServiceSubscriptionResponse(platformReg, finalPlatformForResp, finalEmailAccountForResp)
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
// @Param platform_name query string false "按平台名称筛选"
// @Param email query string false "按邮箱地址筛选"
// @Param username query string false "按平台用户名筛选"
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
	// --- pageSize 处理逻辑 ---
	pageSizeStr := c.Query("pageSize")
	var pageSize int
	fetchAll := false
	if pageSizeStr == "" {
		pageSize = 10 // 默认值
	} else {
		parsedPageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			pageSize = 10 // 解析失败，使用默认值
		} else {
			if parsedPageSize > 0 {
				pageSize = parsedPageSize // 使用前端提供的值
			} else {
				// pageSize <= 0，获取所有记录
				fetchAll = true
				pageSize = 0 // 标记为获取所有，实际查询时不使用 limit
			}
		}
	}
	// --- End pageSize 处理逻辑 ---
	prIDFilter, _ := strconv.Atoi(c.Query("platform_registration_id"))
	statusFilter := strings.ToLower(strings.TrimSpace(c.Query("status")))
	billingCycleFilter := strings.ToLower(strings.TrimSpace(c.Query("billing_cycle")))
	renewalDateStartStr := strings.TrimSpace(c.Query("renewal_date_start"))
	renewalDateEndStr := strings.TrimSpace(c.Query("renewal_date_end"))
	platformNameFilter := strings.TrimSpace(c.Query("platform_name"))
	emailFilter := strings.TrimSpace(c.Query("email"))
	usernameFilter := strings.TrimSpace(c.Query("username"))
	orderBy := c.DefaultQuery("orderBy", "created_at")
	sortDirection := c.DefaultQuery("sortDirection", "desc")

	// Validate orderBy parameter
	allowedOrderByFields := map[string]string{
		"service_name":      "service_name",
		"status":            "status",
		"cost":              "cost",
		"billing_cycle":     "billing_cycle",
		"next_renewal_date": "next_renewal_date",
		"created_at":        "service_subscriptions.created_at",
		"updated_at":        "service_subscriptions.updated_at", // Also qualify updated_at for consistency if it's from service_subscriptions
	}
	dbOrderByField, isValidField := allowedOrderByFields[orderBy]
	if !isValidField {
		dbOrderByField = "service_subscriptions.created_at" // Default to a safe field
	}

	// Validate sortDirection
	if strings.ToLower(sortDirection) != "asc" && strings.ToLower(sortDirection) != "desc" {
		sortDirection = "desc" // Default to desc
	}
	orderClause := dbOrderByField + " " + sortDirection

	if page <= 0 {
		page = 1
	}
	// 移除旧的 pageSize <= 0 和 > 100 的限制
	// if pageSize <= 0 { pageSize = 10 } // 由上面的新逻辑处理
	// if pageSize > 100 { pageSize = 100 } // 移除上限
	offset := 0
	if !fetchAll {
		offset = (page - 1) * pageSize
	}

	var subscriptions []models.ServiceSubscription
	var totalRecords int64

	query := database.DB.Model(&models.ServiceSubscription{}).
		Joins("JOIN platform_registrations ON platform_registrations.id = service_subscriptions.platform_registration_id").
		Where("service_subscriptions.user_id = ?", currentUserID)

	countQuery := database.DB.Model(&models.ServiceSubscription{}).
		Joins("JOIN platform_registrations ON platform_registrations.id = service_subscriptions.platform_registration_id").
		Where("service_subscriptions.user_id = ?", currentUserID)

	if prIDFilter > 0 {
		query = query.Where("service_subscriptions.platform_registration_id = ?", prIDFilter)
		countQuery = countQuery.Where("service_subscriptions.platform_registration_id = ?", prIDFilter)
	}
	if statusFilter != "" {
		query = query.Where("LOWER(service_subscriptions.status) = ?", statusFilter)
		countQuery = countQuery.Where("LOWER(service_subscriptions.status) = ?", statusFilter)
	}
	if billingCycleFilter != "" {
		query = query.Where("LOWER(service_subscriptions.billing_cycle) = ?", billingCycleFilter)
		countQuery = countQuery.Where("LOWER(service_subscriptions.billing_cycle) = ?", billingCycleFilter)
	}

	if renewalDateStartStr != "" {
		startDate, err := time.Parse("2006-01-02", renewalDateStartStr)
		if err == nil {
			query = query.Where("service_subscriptions.next_renewal_date >= ?", startDate)
			countQuery = countQuery.Where("service_subscriptions.next_renewal_date >= ?", startDate)
		}
	}
	if renewalDateEndStr != "" {
		endDate, err := time.Parse("2006-01-02", renewalDateEndStr)
		if err == nil {
			endDate = endDate.AddDate(0, 0, 1)
			query = query.Where("service_subscriptions.next_renewal_date < ?", endDate)
			countQuery = countQuery.Where("service_subscriptions.next_renewal_date < ?", endDate)
		}
	}

	// New filters for platform_name, email, username
	if platformNameFilter != "" {
		query = query.Joins("JOIN platforms ON platforms.id = platform_registrations.platform_id").
			Where("LOWER(platforms.name) LIKE ?", "%"+strings.ToLower(platformNameFilter)+"%")
		countQuery = countQuery.Joins("JOIN platforms ON platforms.id = platform_registrations.platform_id").
			Where("LOWER(platforms.name) LIKE ?", "%"+strings.ToLower(platformNameFilter)+"%")
	}
	if emailFilter != "" {
		query = query.Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id").
			Where("LOWER(email_accounts.email_address) LIKE ?", "%"+strings.ToLower(emailFilter)+"%")
		countQuery = countQuery.Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id").
			Where("LOWER(email_accounts.email_address) LIKE ?", "%"+strings.ToLower(emailFilter)+"%")
	}
	if usernameFilter != "" {
		query = query.Where("LOWER(platform_registrations.login_username) = ?", strings.ToLower(usernameFilter))
		countQuery = countQuery.Where("LOWER(platform_registrations.login_username) = ?", strings.ToLower(usernameFilter))
	}

	if err := countQuery.Count(&totalRecords).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取服务订阅总数失败: "+err.Error())
		return
	}

	// Preload PlatformRegistration and its nested Platform and EmailAccount
	dbQuery := query.Order(orderClause)
	if !fetchAll {
		dbQuery = dbQuery.Offset(offset).Limit(pageSize)
	}
	err := dbQuery.Preload("PlatformRegistration.Platform").
		Preload("PlatformRegistration.EmailAccount").
		Find(&subscriptions).Error
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取服务订阅列表失败: "+err.Error())
		return
	}

	var responses []models.ServiceSubscriptionResponse
	for _, ss := range subscriptions {
		emailAccountForResp := models.EmailAccount{}
		if ss.PlatformRegistration.EmailAccount != nil {
			emailAccountForResp = *ss.PlatformRegistration.EmailAccount
		}
		platformForResp := models.Platform{}          // Assuming Platform is models.Platform in PlatformRegistration
		if ss.PlatformRegistration.Platform.ID != 0 { // Corrected: Check ID for value type
			platformForResp = ss.PlatformRegistration.Platform // Corrected: No indirection for value type
		} else {
			// If Platform is models.Platform (not a pointer) in PlatformRegistration, this else might not be needed
			// or handle it if ss.PlatformRegistration.Platform could be a zero struct.
			// For now, assuming ss.PlatformRegistration.Platform is *models.Platform based on preload behavior.
		}
		responses = append(responses, ss.ToServiceSubscriptionResponse(ss.PlatformRegistration, platformForResp, emailAccountForResp))
	}

	metaPageSize := pageSize
	if fetchAll {
		metaPageSize = int(totalRecords) // 如果获取所有，meta 中的 pageSize 为总数
	}
	pagination := utils.CreatePaginationMeta(page, metaPageSize, int(totalRecords))
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

	emailAccountForRespGetByID := models.EmailAccount{}
	if ss.PlatformRegistration.EmailAccount != nil {
		emailAccountForRespGetByID = *ss.PlatformRegistration.EmailAccount
	}
	platformForRespGetByID := models.Platform{}
	if ss.PlatformRegistration.Platform.ID != 0 { // Corrected: Check ID for value type
		platformForRespGetByID = ss.PlatformRegistration.Platform // Corrected: No indirection for value type
	}

	response := ss.ToServiceSubscriptionResponse(ss.PlatformRegistration, platformForRespGetByID, emailAccountForRespGetByID)
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
	// Removed check that prevented updating email_address per request.
	// Note: The current logic doesn't actually persist email changes here,
	// as email is tied to PlatformRegistration.
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

	emailAccountForRespUpdate := models.EmailAccount{}
	if ss.PlatformRegistration.EmailAccount != nil {
		emailAccountForRespUpdate = *ss.PlatformRegistration.EmailAccount
	}
	platformForRespUpdate := models.Platform{}
	if ss.PlatformRegistration.Platform.ID != 0 { // Corrected: Check ID for value type
		platformForRespUpdate = ss.PlatformRegistration.Platform // Corrected: No indirection for value type
	}
	response := ss.ToServiceSubscriptionResponse(ss.PlatformRegistration, platformForRespUpdate, emailAccountForRespUpdate)
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

	if err := database.DB.Unscoped().Delete(&ss).Error; err != nil {
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
	// --- pageSize 处理逻辑 ---
	pageSizeStr := c.Query("pageSize")
	var pageSize int
	fetchAll := false
	if pageSizeStr == "" {
		pageSize = 10 // 默认值
	} else {
		parsedPageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			pageSize = 10 // 解析失败，使用默认值
		} else {
			if parsedPageSize > 0 {
				pageSize = parsedPageSize // 使用前端提供的值
			} else {
				// pageSize <= 0，获取所有记录
				fetchAll = true
				pageSize = 0 // 标记为获取所有，实际查询时不使用 limit
			}
		}
	}
	// --- End pageSize 处理逻辑 ---
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
	// 移除旧的 pageSize <= 0 和 > 100 的限制
	// if pageSize <= 0 { pageSize = 10 } // 由上面的新逻辑处理
	// if pageSize > 100 { pageSize = 100 } // 移除上限
	offset := 0
	if !fetchAll {
		offset = (page - 1) * pageSize
	}

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
	finalDbQuery := database.DB.Model(&models.ServiceSubscription{}).
		Where("platform_registration_id = ? AND user_id = ?", platformRegistrationID, uint(currentUserID)).
		Order(orderClause)

	if !fetchAll {
		finalDbQuery = finalDbQuery.Offset(offset).Limit(pageSize)
	}

	err = finalDbQuery.Find(&subscriptions).Error
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取服务订阅信息失败: "+err.Error())
		return
	}

	var responses []models.ServiceSubscriptionResponse
	for _, ss := range subscriptions {
		// For each subscription, we use the already fetched `pr` for its related data.
		// pr.Platform and pr.EmailAccount were preloaded.
		// pr.EmailAccount is *models.EmailAccount
		// pr.Platform is *models.Platform (based on Preload("Platform") behavior with GORM, it usually preloads into a pointer field if the field is a pointer, or directly if not)
		// Let's assume pr.Platform is *models.Platform for consistency with other fixes.

		emailAccountForRespLoop := models.EmailAccount{}
		if pr.EmailAccount != nil {
			emailAccountForRespLoop = *pr.EmailAccount
		}
		platformForRespLoop := models.Platform{}
		if pr.Platform.ID != 0 {
			platformForRespLoop = pr.Platform
		}
		responses = append(responses, ss.ToServiceSubscriptionResponse(pr, platformForRespLoop, emailAccountForRespLoop))
	}

	metaPageSize := pageSize
	if fetchAll {
		metaPageSize = int(totalRecords) // 如果获取所有，meta 中的 pageSize 为总数
	}
	pagination := utils.CreatePaginationMeta(page, metaPageSize, int(totalRecords))
	utils.SendSuccessResponseWithMeta(c, responses, pagination)
}

// GetDistinctPlatformNames godoc
// @Summary 获取所有去重的平台名称列表
// @Description 获取与当前用户服务订阅相关的、所有去重后的平台名称列表
// @Tags ServiceSubscriptions
// @Produce json
// @Success 200 {object} models.SuccessResponse{data=[]string} "获取成功"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /service-subscriptions/distinct-platform-names [get]
// @Security BearerAuth
func GetDistinctPlatformNames(c *gin.Context) {
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
	var platformNames []string
	query := database.DB.Table("service_subscriptions ss").
		Select("DISTINCT p.name").
		Joins("JOIN platform_registrations pr ON pr.id = ss.platform_registration_id").
		Joins("JOIN platforms p ON p.id = pr.platform_id").
		Where("ss.user_id = ?", currentUserID).
		Where("p.name IS NOT NULL AND p.name != ''").
		Order("p.name ASC")

	if err := query.Scan(&platformNames).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取平台名称列表失败: "+err.Error())
		return
	}

	if platformNames == nil {
		platformNames = []string{} //确保在没有结果时返回空数组而不是null
	}
	utils.SendSuccessResponse(c, platformNames)
}

// GetDistinctEmails godoc
// @Summary 获取所有去重的邮箱地址列表
// @Description 获取与当前用户服务订阅相关的、所有去重后的邮箱地址列表
// @Tags ServiceSubscriptions
// @Produce json
// @Success 200 {object} models.SuccessResponse{data=[]string} "获取成功"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /service-subscriptions/distinct-emails [get]
// @Security BearerAuth
func GetDistinctEmails(c *gin.Context) {
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
	var emailAddresses []string
	query := database.DB.Table("service_subscriptions ss").
		Select("DISTINCT ea.email_address").
		Joins("JOIN platform_registrations pr ON pr.id = ss.platform_registration_id").
		Joins("JOIN email_accounts ea ON ea.id = pr.email_account_id").
		Where("ss.user_id = ?", currentUserID).
		Where("ea.email_address IS NOT NULL AND ea.email_address != ''").
		Order("ea.email_address ASC")

	if err := query.Scan(&emailAddresses).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取邮箱地址列表失败: "+err.Error())
		return
	}
	if emailAddresses == nil {
		emailAddresses = []string{}
	}
	utils.SendSuccessResponse(c, emailAddresses)
}

// 根据数据处理原则实现的辅助函数

// createServiceSubscriptionCase1 处理情况1: 邮箱地址为空，用户名不为空
func createServiceSubscriptionCase1(tx *gorm.DB, userID uint, platform models.Platform, loginUsername string, platformReg *models.PlatformRegistration) error {
	// 查询平台注册表中该平台和用户名是否存在
	err := tx.Where("user_id = ? AND platform_id = ? AND login_username = ?", userID, platform.ID, loginUsername).
		Preload("Platform").Preload("EmailAccount").First(platformReg).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 不存在，新增一条平台和用户名的注册记录
			*platformReg = models.PlatformRegistration{
				UserID:        userID,
				PlatformID:    platform.ID,
				LoginUsername: &loginUsername,
			}
			if createErr := tx.Create(platformReg).Error; createErr != nil {
				return fmt.Errorf("创建平台注册记录失败: %v", createErr)
			}
			platformReg.Platform = platform
		} else {
			return fmt.Errorf("查询平台注册记录失败: %v", err)
		}
	}
	// 如果存在，直接使用现有记录
	return nil
}

// createServiceSubscriptionCase2 处理情况2: 邮箱地址不为空，用户名为空
func createServiceSubscriptionCase2(tx *gorm.DB, userID uint, platform models.Platform, emailAddress string, platformReg *models.PlatformRegistration, emailAccount *models.EmailAccount) error {
	// 查询邮箱表中该邮箱是否存在
	err := tx.Where("user_id = ? AND email_address = ?", userID, emailAddress).First(emailAccount).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 邮箱不存在，新增一条邮箱记录
			*emailAccount = models.EmailAccount{
				UserID:       userID,
				EmailAddress: emailAddress,
				Provider:     utils.ExtractProviderFromEmail(emailAddress),
			}
			if createErr := tx.Create(emailAccount).Error; createErr != nil {
				return fmt.Errorf("创建邮箱记录失败: %v", createErr)
			}
		} else {
			return fmt.Errorf("查询邮箱记录失败: %v", err)
		}
	}

	// 查询平台注册表中该平台和邮箱是否存在
	err = tx.Where("user_id = ? AND platform_id = ? AND email_account_id = ?", userID, platform.ID, emailAccount.ID).
		Preload("Platform").Preload("EmailAccount").First(platformReg).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 不存在，新增一条平台和邮箱的注册记录
			*platformReg = models.PlatformRegistration{
				UserID:         userID,
				PlatformID:     platform.ID,
				EmailAccountID: &emailAccount.ID,
			}
			if createErr := tx.Create(platformReg).Error; createErr != nil {
				return fmt.Errorf("创建平台注册记录失败: %v", createErr)
			}
			platformReg.Platform = platform
			platformReg.EmailAccount = emailAccount
		} else {
			return fmt.Errorf("查询平台注册记录失败: %v", err)
		}
	}
	// 如果存在，直接使用现有记录
	return nil
}

// createServiceSubscriptionCase3 处理情况3: 邮箱地址和用户名都不为空
func createServiceSubscriptionCase3(tx *gorm.DB, userID uint, platform models.Platform, emailAddress, loginUsername string, platformReg *models.PlatformRegistration, emailAccount *models.EmailAccount) error {
	// 查询平台注册表中该平台、邮箱和用户名的组合是否存在
	err := tx.Where("user_id = ? AND platform_id = ? AND login_username = ?", userID, platform.ID, loginUsername).
		Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id").
		Where("email_accounts.email_address = ?", emailAddress).
		Preload("Platform").Preload("EmailAccount").First(platformReg).Error

	if err == nil {
		// 存在完整组合，直接使用
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("查询平台注册记录失败: %v", err)
	}

	// 不存在，检查是否存在冲突
	var conflictReg models.PlatformRegistration

	// 检查平台和邮箱是否存在（用户名不同）
	err = tx.Where("user_id = ? AND platform_id = ?", userID, platform.ID).
		Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id").
		Where("email_accounts.email_address = ? AND (platform_registrations.login_username != ? OR platform_registrations.login_username IS NULL)", emailAddress, loginUsername).
		First(&conflictReg).Error

	if err == nil {
		return fmt.Errorf("该邮箱已注册在该平台，但用户名不一致")
	}

	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("检查邮箱冲突失败: %v", err)
	}

	// 检查平台和用户名是否存在（邮箱不同）
	err = tx.Where("user_id = ? AND platform_id = ? AND login_username = ?", userID, platform.ID, loginUsername).
		Joins("LEFT JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id").
		Where("email_accounts.email_address != ? OR email_accounts.email_address IS NULL", emailAddress).
		First(&conflictReg).Error

	if err == nil {
		return fmt.Errorf("该用户名已注册在该平台，但邮箱不一致")
	}

	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("检查用户名冲突失败: %v", err)
	}

	// 都不存在冲突，查询邮箱表中该邮箱是否存在
	err = tx.Where("user_id = ? AND email_address = ?", userID, emailAddress).First(emailAccount).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 邮箱不存在，新增一条邮箱记录
			*emailAccount = models.EmailAccount{
				UserID:       userID,
				EmailAddress: emailAddress,
				Provider:     utils.ExtractProviderFromEmail(emailAddress),
			}
			if createErr := tx.Create(emailAccount).Error; createErr != nil {
				return fmt.Errorf("创建邮箱记录失败: %v", createErr)
			}
		} else {
			return fmt.Errorf("查询邮箱记录失败: %v", err)
		}
	}

	// 新增一条平台注册记录（包含平台、邮箱、用户名）
	*platformReg = models.PlatformRegistration{
		UserID:         userID,
		PlatformID:     platform.ID,
		EmailAccountID: &emailAccount.ID,
		LoginUsername:  &loginUsername,
	}
	if createErr := tx.Create(platformReg).Error; createErr != nil {
		return fmt.Errorf("创建平台注册记录失败: %v", createErr)
	}

	platformReg.Platform = platform
	platformReg.EmailAccount = emailAccount

	return nil
}

// GetDistinctUsernames godoc
// @Summary 获取所有去重的用户名列表
// @Description 获取与当前用户服务订阅相关的、所有去重后的平台用户名列表
// @Tags ServiceSubscriptions
// @Produce json
// @Success 200 {object} models.SuccessResponse{data=[]string} "获取成功"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 500 {object} models.ErrorResponse "服务器内部错误"
// @Router /service-subscriptions/distinct-usernames [get]
// @Security BearerAuth
func GetDistinctUsernames(c *gin.Context) {
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
	var usernames []string
	query := database.DB.Table("service_subscriptions ss").
		Select("DISTINCT pr.login_username").
		Joins("JOIN platform_registrations pr ON pr.id = ss.platform_registration_id").
		Where("ss.user_id = ?", currentUserID).
		Where("pr.login_username IS NOT NULL AND pr.login_username != ''").
		Order("pr.login_username ASC")

	if err := query.Scan(&usernames).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取用户名列表失败: "+err.Error())
		return
	}
	if usernames == nil {
		usernames = []string{}
	}
	utils.SendSuccessResponse(c, usernames)
}
