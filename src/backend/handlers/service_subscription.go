package handlers

import (
	"email_server/database"
	"email_server/models"
	"email_server/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateServiceSubscription godoc
// @Summary 创建服务订阅
// @Description 为当前用户创建一个新的服务订阅，关联一个平台注册信息
// @Tags ServiceSubscriptions
// @Accept json
// @Produce json
// @Param serviceSubscription body models.ServiceSubscription true "服务订阅信息。ID, UserID, CreatedAt 等会被忽略。"
// @Success 201 {object} models.SuccessResponse{data=models.ServiceSubscriptionResponse} "创建成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误或关联资源无效"
// @Failure 401 {object} models.ErrorResponse "用户未认证"
// @Failure 404 {object} models.ErrorResponse "关联的平台注册信息未找到"
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

	var input struct {
		PlatformRegistrationID uint    `json:"platform_registration_id" binding:"required"`
		ServiceName            string  `json:"service_name" binding:"required"`
		Description            string  `json:"description"`
		Status                 string  `json:"status" binding:"required"` // e.g., active, cancelled
		Cost                   float64 `json:"cost" binding:"min=0"`
		BillingCycle           string  `json:"billing_cycle" binding:"required"` // e.g., monthly, yearly
		NextRenewalDateStr     *string `json:"next_renewal_date"`                // Format: YYYY-MM-DD
		PaymentMethodNotes     string  `json:"payment_method_notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	// 验证 PlatformRegistration 是否属于当前用户
	var pr models.PlatformRegistration
	if err := database.DB.Where("id = ? AND user_id = ?", input.PlatformRegistrationID, currentUserID).
		Preload("Platform").Preload("EmailAccount").First(&pr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "关联的平台注册信息未找到或不属于当前用户")
			return
		}
		utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台注册信息失败: "+err.Error())
		return
	}

	var nextRenewalDate *time.Time
	if input.NextRenewalDateStr != nil && *input.NextRenewalDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", *input.NextRenewalDateStr)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusBadRequest, "下次续费日期格式无效，请使用 YYYY-MM-DD")
			return
		}
		nextRenewalDate = &parsedDate
	}

	subscription := models.ServiceSubscription{
		UserID:                 currentUserID,
		PlatformRegistrationID: input.PlatformRegistrationID,
		ServiceName:            input.ServiceName,
		Description:            input.Description,
		Status:                 input.Status,
		Cost:                   input.Cost,
		BillingCycle:           input.BillingCycle,
		NextRenewalDate:        nextRenewalDate,
		PaymentMethodNotes:     input.PaymentMethodNotes,
	}

	if err := database.DB.Create(&subscription).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "创建服务订阅失败: "+err.Error())
		return
	}
	
	// subscription will have its ID populated. We need pr, pr.Platform, pr.EmailAccount for the response.
	response := subscription.ToServiceSubscriptionResponse(pr, pr.Platform, pr.EmailAccount)
	utils.SendSuccessResponse(c, response)
}

// GetServiceSubscriptions godoc
// @Summary 获取当前用户的所有服务订阅
// @Description 获取当前登录用户的所有服务订阅，支持分页和筛选
// @Tags ServiceSubscriptions
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param platform_registration_id query int false "按平台注册ID筛选"
// @Param status query string false "按订阅状态筛选"
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
	statusFilter := c.Query("status")

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
		query = query.Where("status = ?", statusFilter)
		countQuery = countQuery.Where("status = ?", statusFilter)
	}
	
	if err := countQuery.Count(&totalRecords).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "获取服务订阅总数失败: "+err.Error())
		return
	}

	// Preload PlatformRegistration and its nested Platform and EmailAccount
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).
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
// @Description 更新当前用户拥有的指定ID的服务订阅信息
// @Tags ServiceSubscriptions
// @Accept json
// @Produce json
// @Param id path int true "服务订阅ID"
// @Param serviceSubscription body models.ServiceSubscription true "要更新的服务订阅信息。UserID, PlatformRegistrationID 不可更改。"
// @Success 200 {object} models.SuccessResponse{data=models.ServiceSubscriptionResponse} "更新成功"
// @Failure 400 {object} models.ErrorResponse "请求参数错误或无效的ID格式"
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

	var input struct {
		ServiceName        string  `json:"service_name" binding:"omitempty,required"` // omitempty allows partial, but service name is core
		Description        string  `json:"description"`
		Status             string  `json:"status" binding:"omitempty,required"`
		Cost               float64 `json:"cost" binding:"omitempty,min=0"`
		BillingCycle       string  `json:"billing_cycle" binding:"omitempty,required"`
		NextRenewalDateStr *string `json:"next_renewal_date"` // Format: YYYY-MM-DD
		PaymentMethodNotes string  `json:"payment_method_notes"`
		// PlatformRegistrationID is not updatable here.
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	if input.ServiceName != "" { ss.ServiceName = input.ServiceName }
	ss.Description = input.Description // Allow clearing
	if input.Status != "" { ss.Status = input.Status }
	// Cost needs careful handling for omitempty if 0 is a valid value but also "not provided"
    // Assuming if cost is in the payload, it's an intended update.
    // For simplicity, if the key is present in JSON, we update.
    // This requires client to send all fields they want to keep, or use PATCH.
    // Given PUT, typically means full replacement of updatable fields.
    // Let's assume client sends all updatable fields.
    ss.Cost = input.Cost
	if input.BillingCycle != "" { ss.BillingCycle = input.BillingCycle }
	ss.PaymentMethodNotes = input.PaymentMethodNotes // Allow clearing

	if input.NextRenewalDateStr != nil {
		if *input.NextRenewalDateStr == "" { // Explicitly clearing the date
			ss.NextRenewalDate = nil
		} else {
			parsedDate, err := time.Parse("2006-01-02", *input.NextRenewalDateStr)
			if err != nil {
				utils.SendErrorResponse(c, http.StatusBadRequest, "下次续费日期格式无效，请使用 YYYY-MM-DD")
				return
			}
			ss.NextRenewalDate = &parsedDate
		}
	}


	if err := database.DB.Save(&ss).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "更新服务订阅失败: "+err.Error())
		return
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