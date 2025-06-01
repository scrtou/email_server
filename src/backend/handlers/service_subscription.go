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

	var platform models.Platform
	var platformReg models.PlatformRegistration 
	var emailAccount models.EmailAccount 
	var emailAccountIDToUse uint        


	// 1. 尝试通过 ID 加载 PlatformRegistration
	if input.PlatformRegistrationID != nil && *input.PlatformRegistrationID != 0 {
		if err := tx.Where("id = ? AND user_id = ?", *input.PlatformRegistrationID, currentUserID).
			Preload("Platform").Preload("EmailAccount").First(&platformReg).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				tx.Rollback()
				utils.SendErrorResponse(c, http.StatusBadRequest, "提供的 PlatformRegistrationID 无效或不属于当前用户")
				return
			}
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台注册信息失败 (by PlatformRegistrationID): "+err.Error())
			return
		}
		if platformReg.Platform.ID != 0 {
			platform = platformReg.Platform
		}
		if platformReg.EmailAccount != nil && platformReg.EmailAccount.ID != 0 {
			emailAccount = *platformReg.EmailAccount
			emailAccountIDToUse = emailAccount.ID
		}

	} else if input.SelectedUsernameRegistrationID != 0 {
		if err := tx.Where("id = ? AND user_id = ?", input.SelectedUsernameRegistrationID, currentUserID).
			Preload("Platform").Preload("EmailAccount").First(&platformReg).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				tx.Rollback()
				utils.SendErrorResponse(c, http.StatusBadRequest, "提供的 SelectedUsernameRegistrationID 无效或不属于当前用户")
				return
			}
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台注册信息失败 (by SelectedUsernameRegistrationID): "+err.Error())
			return
		}
		if platformReg.Platform.ID != 0 {
			platform = platformReg.Platform
		}
		if platformReg.EmailAccount != nil && platformReg.EmailAccount.ID != 0 {
			emailAccount = *platformReg.EmailAccount
			emailAccountIDToUse = emailAccount.ID
		}
	}

	// 2. 自定义校验逻辑 (如果 platformReg 未通过 ID 加载)
	if platformReg.ID == 0 {
		emailInfoProvided := input.EmailAccountID != 0 || input.EmailAddress != ""
		loginUsernameStringProvided := input.LoginUsername != ""
		if !emailInfoProvided && !loginUsernameStringProvided {
			tx.Rollback() 
			utils.SendErrorResponse(c, http.StatusBadRequest, "当未通过ID指定平台注册时，必须提供邮箱信息(ID或地址)或登录用户名字符串中的至少一个")
			return
		}
	}
	
	// 3. 确定 emailAccountIDToUse (如果尚未通过已加载的 platformReg 确定)
	if platformReg.ID == 0 && emailAccountIDToUse == 0 { 
		if input.EmailAccountID != 0 {
			if err := tx.Where("id = ? AND user_id = ?", input.EmailAccountID, currentUserID).First(&emailAccount).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					tx.Rollback()
					utils.SendErrorResponse(c, http.StatusBadRequest, "提供的邮箱账户ID无效或不属于当前用户")
					return
				}
				tx.Rollback()
				utils.SendErrorResponse(c, http.StatusInternalServerError, "查询邮箱账户失败 (by ID): "+err.Error())
				return
			}
			emailAccountIDToUse = emailAccount.ID
		} else if input.EmailAddress != "" {
			if err := tx.Where("email_address = ? AND user_id = ?", input.EmailAddress, currentUserID).First(&emailAccount).Error; err != nil {
				if err != gorm.ErrRecordNotFound { 
					tx.Rollback()
					utils.SendErrorResponse(c, http.StatusInternalServerError, "查询邮箱账户失败 (by Address): "+err.Error())
					return
				}
			} else {
				emailAccountIDToUse = emailAccount.ID
			}
		}
	}

	// 4. 平台处理 (如果尚未通过已加载的 platformReg 确定 platform)
	if platformReg.ID == 0 || platform.ID == 0 { 
		if input.PlatformID != nil && *input.PlatformID > 0 { 
			if err := tx.Where("id = ? AND user_id = ?", *input.PlatformID, currentUserID).First(&platform).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					tx.Rollback()
					utils.SendErrorResponse(c, http.StatusBadRequest, "指定的平台ID不存在或不属于当前用户")
					return
				}
				tx.Rollback()
				utils.SendErrorResponse(c, http.StatusInternalServerError, "查询平台失败: "+err.Error())
				return
			}
		} else if input.PlatformName != "" { 
			err := tx.Where("name = ? AND user_id = ?", input.PlatformName, currentUserID).First(&platform).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound { 
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
		} else {
			if platformReg.ID == 0 { 
				tx.Rollback()
				utils.SendErrorResponse(c, http.StatusBadRequest, "必须提供平台ID或平台名称")
				return
			}
		}
	}

	// 5. 通过组件查找/创建 PlatformRegistration (如果 platformReg.ID == 0 且上述校验通过后执行)
	if platformReg.ID == 0 {
		if platform.ID == 0 { 
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusInternalServerError, "未能确定或创建平台信息")
			return
		}

		query := tx.Where("user_id = ? AND platform_id = ?", currentUserID, platform.ID)
		if input.LoginUsername != "" {
			query = query.Where("login_username = ?", input.LoginUsername)
		}
		
		if emailAccountIDToUse != 0 {
			query = query.Where("email_account_id = ?", emailAccountIDToUse)
		} else {
			query = query.Where("email_account_id IS NULL")
		}
		
		err := query.Preload("Platform").Preload("EmailAccount").First(&platformReg).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound { 
				var eaIDForNewReg *uint
				if emailAccountIDToUse != 0 {
					eaIDForNewReg = &emailAccountIDToUse
				}

				platformReg = models.PlatformRegistration{
					UserID:         currentUserID,
					EmailAccountID: eaIDForNewReg,
					PlatformID:     platform.ID,
					LoginUsername:  input.LoginUsername, 
				}
				if createErr := tx.Create(&platformReg).Error; createErr != nil {
					tx.Rollback()
					if strings.Contains(createErr.Error(), "UNIQUE constraint failed") {
						// 根据输入推断是哪个约束冲突
						// 优先判断 EmailAccountID 是否可能导致冲突
						if emailAccountIDToUse > 0 {
							// 实际冲突可能是 (UserID, PlatformID, EmailAccountID, IsActive)
							utils.SendErrorResponse(c, http.StatusConflict, "创建关联平台注册失败：此邮箱账户已在该平台注册。")
							return
						} else if input.LoginUsername != "" {
							// 如果 EmailAccountID 为 0 或 nil，再判断 LoginUsername 是否导致冲突
							utils.SendErrorResponse(c, http.StatusConflict, "创建关联平台注册失败：此用户名已在该平台注册。")
							return
						}
						// 通用冲突消息
						utils.SendErrorResponse(c, http.StatusConflict, "创建关联平台注册失败：注册信息与现有记录冲突。")
					} else {
						utils.SendErrorResponse(c, http.StatusInternalServerError, "创建平台注册信息失败: "+createErr.Error())
					}
					return
				}
				platformReg.Platform = platform
				if emailAccountIDToUse != 0 && emailAccount.ID != 0 { 
					platformReg.EmailAccount = &emailAccount
				} else {
					platformReg.EmailAccount = nil 
				}

			} else { 
				tx.Rollback()
				utils.SendErrorResponse(c, http.StatusInternalServerError, "查询或创建平台注册信息失败: "+err.Error())
				return
			}
		}
	}

	// 6. 最终检查 platformReg
	if platformReg.ID == 0 {
		tx.Rollback()
		utils.SendErrorResponse(c, http.StatusInternalServerError, "未能确定或创建平台注册信息")
		return
	}
	
	if platformReg.Platform.ID == 0 && platform.ID != 0 {
		platformReg.Platform = platform
	}
	if platformReg.EmailAccount == nil && emailAccountIDToUse != 0 && emailAccount.ID != 0 {
		platformReg.EmailAccount = &emailAccount
	}


	// 7. 处理下次续费日期
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

	// 8. 查找或创建/恢复 ServiceSubscription 记录
	var subscription models.ServiceSubscription
	err := tx.Unscoped().Where(models.ServiceSubscription{
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
		if subscription.DeletedAt.Valid { 
			subscription.Description = input.Description
			subscription.Status = input.Status
			subscription.Cost = input.Cost
			subscription.BillingCycle = input.BillingCycle
			subscription.NextRenewalDate = nextRenewalDate
			subscription.PaymentMethodNotes = input.PaymentMethodNotes
			subscription.DeletedAt = gorm.DeletedAt{} 
			if updateErr := tx.Unscoped().Save(&subscription).Error; updateErr != nil {
				tx.Rollback()
				utils.SendErrorResponse(c, http.StatusInternalServerError, "恢复并更新服务订阅失败: "+updateErr.Error())
				return
			}
		} else { 
			tx.Rollback()
			utils.SendErrorResponse(c, http.StatusConflict, "该服务订阅已存在。")
			return
		}
	}

	// --- 提交事务 ---
	if err := tx.Commit().Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	
	// 准备响应数据
	finalPlatformForResp := models.Platform{}
	if platformReg.Platform.ID != 0 { 
		finalPlatformForResp = platformReg.Platform
	} else if platform.ID != 0 { 
		finalPlatformForResp = platform
	}


	finalEmailAccountForResp := models.EmailAccount{}
	if platformReg.EmailAccount != nil && platformReg.EmailAccount.ID != 0 { 
		finalEmailAccountForResp = *platformReg.EmailAccount
	} else if emailAccount.ID != 0 { 
		finalEmailAccountForResp = emailAccount
	}
	

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
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
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

	if page <= 0 { page = 1 }
	if pageSize <= 0 { pageSize = 10 }
	if pageSize > 100 { pageSize = 100 }
	offset := (page - 1) * pageSize

	var subscriptions []models.ServiceSubscription
	var totalRecords int64

	query := database.DB.Model(&models.ServiceSubscription{}).
		Joins("JOIN platform_registrations ON platform_registrations.id = service_subscriptions.platform_registration_id AND platform_registrations.deleted_at IS NULL"). // Ensure platform_registration is not soft-deleted
		Where("service_subscriptions.user_id = ?", currentUserID)

	countQuery := database.DB.Model(&models.ServiceSubscription{}).
		Joins("JOIN platform_registrations ON platform_registrations.id = service_subscriptions.platform_registration_id AND platform_registrations.deleted_at IS NULL").
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
		query = query.Joins("JOIN platforms ON platforms.id = platform_registrations.platform_id AND platforms.deleted_at IS NULL").
			Where("LOWER(platforms.name) LIKE ?", "%"+strings.ToLower(platformNameFilter)+"%")
		countQuery = countQuery.Joins("JOIN platforms ON platforms.id = platform_registrations.platform_id AND platforms.deleted_at IS NULL").
			Where("LOWER(platforms.name) LIKE ?", "%"+strings.ToLower(platformNameFilter)+"%")
	}
	if emailFilter != "" {
		query = query.Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id AND email_accounts.deleted_at IS NULL").
			Where("LOWER(email_accounts.email_address) LIKE ?", "%"+strings.ToLower(emailFilter)+"%")
		countQuery = countQuery.Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id AND email_accounts.deleted_at IS NULL").
			Where("LOWER(email_accounts.email_address) LIKE ?", "%"+strings.ToLower(emailFilter)+"%")
	}
	if usernameFilter != "" {
		query = query.Where("LOWER(platform_registrations.login_username) LIKE ?", "%"+strings.ToLower(usernameFilter)+"%")
		countQuery = countQuery.Where("LOWER(platform_registrations.login_username) LIKE ?", "%"+strings.ToLower(usernameFilter)+"%")
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
		emailAccountForResp := models.EmailAccount{}
		if ss.PlatformRegistration.EmailAccount != nil {
			emailAccountForResp = *ss.PlatformRegistration.EmailAccount
		}
		platformForResp := models.Platform{} // Assuming Platform is models.Platform in PlatformRegistration
		if ss.PlatformRegistration.Platform.ID != 0 { // Corrected: Check ID for value type
			platformForResp = ss.PlatformRegistration.Platform // Corrected: No indirection for value type
		} else {
			// If Platform is models.Platform (not a pointer) in PlatformRegistration, this else might not be needed
			// or handle it if ss.PlatformRegistration.Platform could be a zero struct.
			// For now, assuming ss.PlatformRegistration.Platform is *models.Platform based on preload behavior.
		}
		responses = append(responses, ss.ToServiceSubscriptionResponse(ss.PlatformRegistration, platformForResp, emailAccountForResp))
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

	pagination := utils.CreatePaginationMeta(page, pageSize, int(totalRecords))
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
		Joins("JOIN platform_registrations pr ON pr.id = ss.platform_registration_id AND pr.deleted_at IS NULL").
		Joins("JOIN platforms p ON p.id = pr.platform_id AND p.deleted_at IS NULL").
		Where("ss.user_id = ? AND ss.deleted_at IS NULL", currentUserID).
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
		Joins("JOIN platform_registrations pr ON pr.id = ss.platform_registration_id AND pr.deleted_at IS NULL").
		Joins("JOIN email_accounts ea ON ea.id = pr.email_account_id AND ea.deleted_at IS NULL").
		Where("ss.user_id = ? AND ss.deleted_at IS NULL", currentUserID).
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
		Joins("JOIN platform_registrations pr ON pr.id = ss.platform_registration_id AND pr.deleted_at IS NULL").
		Where("ss.user_id = ? AND ss.deleted_at IS NULL", currentUserID).
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