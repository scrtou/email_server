package handlers

import (
	"fmt" // Added import for fmt.Sscan
	"log"
	"time"
	"net/http"
	

	"email_server/database"
	"email_server/models"
	
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

const layoutISO = "2006-01-02" // 日期格式

// ReminderInfo 用于API响应的结构体
type ReminderInfo struct {
	ID               uint      `json:"id"`
	ServiceName      string    `json:"serviceName"`
	PlatformName     string    `json:"platformName"`
	NextRenewalDate      string    `json:"renewalDate"` // 使用字符串表示日期
	DaysRemaining    int       `json:"daysRemaining"`
	Status           string    `json:"status"`
	IsRead           bool      `json:"is_read"` // 新增字段，表示是否已读
}


// StartSubscriptionReminderJob 初始化并启动订阅提醒的定时任务
func StartSubscriptionReminderJob() {
	c := cron.New()
	// 每天凌晨1点执行 (可根据需求调整 "0 1 * * *")
	_, err := c.AddFunc("0 1 * * *", checkUpcomingRenewals)
	if err != nil {
		log.Fatalf("Error adding cron job: %v", err)
	}
	c.Start()
	log.Println("Subscription reminder job started.")
}

// checkUpcomingRenewals 检查即将到期的订阅并记录日志
func checkUpcomingRenewals() {
	log.Println("Running daily check for upcoming renewals...")
	db := database.DB

	var upcomingSubscriptions []models.ServiceSubscription
	thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)

	// 查询未来30天内到期且状态为 "active" 的订阅
	// 注意：NextRenewalDate 在数据库中可能是 time.Time 类型，比较时需要确保格式正确
	// GORM 通常能处理好 time.Time 类型的比较
	err := db.Preload("PlatformRegistration.Platform").
		Where("next_renewal_date <= ? AND next_renewal_date >= ? AND status = ?", thirtyDaysFromNow, time.Now(), "active").
		Find(&upcomingSubscriptions).Error

	if err != nil {
		log.Printf("Error querying upcoming renewals: %v", err)
		return
	}

	if len(upcomingSubscriptions) > 0 {
		log.Printf("Found %d upcoming renewals:", len(upcomingSubscriptions))
		for _, sub := range upcomingSubscriptions {
if sub.NextRenewalDate == nil {
				log.Printf("  - Subscription ID: %d, Service: %s (Platform: %s) has no renewal date.",
					sub.ID,
					sub.ServiceName,
					sub.PlatformRegistration.Platform.Name)
				continue
			}
			daysRemaining := int(sub.NextRenewalDate.Sub(time.Now()).Hours() / 24)
			log.Printf("  - Subscription ID: %d, Service: %s (Platform: %s), Renewal Date: %s, Days Remaining: %d",
				sub.ID,
				sub.ServiceName, // 假设 ServiceSubscription 直接有 ServiceName
				sub.PlatformRegistration.Platform.Name,
				sub.NextRenewalDate.Format(layoutISO),
				daysRemaining)
		}
	} else {
		log.Println("No upcoming renewals found in the next 30 days.")
	}
}

// GetUserReminders 获取当前认证用户的订阅提醒列表
func GetUserReminders(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id") // Corrected: Changed "userID" to "user_id"
	if !exists {
		log.Printf("[GetUserReminders] User ID not found in context. Expected key 'user_id'. Available keys: %v", c.Keys)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context after auth"})
		return
	}

	var userID uint
	switch v := userIDInterface.(type) {
	case float64: // Common for JWT numbers when parsed from JSON
		if v < 0 {
			log.Printf("[GetUserReminders] Invalid user ID value (float64 negative) in context: %v", v)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}
		userID = uint(v)
		log.Printf("[GetUserReminders] User ID successfully converted from float64: %d", userID)
	case int64:   // As per the error log and JWT claims struct
		if v < 0 {
			log.Printf("[GetUserReminders] Invalid user ID value (int64 negative) in context: %v", v)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}
		userID = uint(v)
		log.Printf("[GetUserReminders] User ID successfully converted from int64: %d", userID)
	case int:
		if v < 0 {
			log.Printf("[GetUserReminders] Invalid user ID value (int negative) in context: %v", v)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}
		userID = uint(v)
		log.Printf("[GetUserReminders] User ID successfully converted from int: %d", userID)
	case uint:
		userID = v
		log.Printf("[GetUserReminders] User ID successfully retrieved as uint: %d", userID)
	default:
		log.Printf("[GetUserReminders] Unexpected user ID type in context: %T for value %v", userIDInterface, userIDInterface)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error processing user ID"})
		return
	}


	db := database.DB
	var userSubscriptions []models.ServiceSubscription
	thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)

	// 查询属于该用户、未来30天内到期且状态为 "active" 的订阅
	// 我们需要通过 PlatformRegistration 来关联到 UserID
	err := db.Joins("JOIN platform_registrations ON platform_registrations.id = service_subscriptions.platform_registration_id").
		Joins("JOIN platforms ON platforms.id = platform_registrations.platform_id"). // 用于获取 PlatformName 和 Icon
		Preload("PlatformRegistration.Platform"). // 预加载以方便获取平台信息
		Where("platform_registrations.user_id = ? AND service_subscriptions.next_renewal_date <= ? AND service_subscriptions.next_renewal_date >= ? AND service_subscriptions.status = ?",
			userID, thirtyDaysFromNow, time.Now(), "active").
		Select("service_subscriptions.*, platforms.name as platform_name"). // 选择平台名称和图标
		Find(&userSubscriptions).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{"reminders": []ReminderInfo{}}) // 返回空列表
			return
		}
		log.Printf("Error fetching user reminders for user ID %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reminders"})
		return
	}

	var reminders []ReminderInfo
	for _, sub := range userSubscriptions {
if sub.NextRenewalDate == nil {
				// Optionally log or handle subscriptions without a renewal date
				continue
			}
		daysRemaining := int(sub.NextRenewalDate.Sub(time.Now()).Hours() / 24)
		if daysRemaining < 0 { // 确保不会显示负数天数 (虽然查询条件应该已经过滤了)
			daysRemaining = 0
		}
		reminders = append(reminders, ReminderInfo{
			ID:               sub.ID,
			ServiceName:      sub.ServiceName, // ServiceSubscription 直接有 ServiceName
			PlatformName:     sub.PlatformRegistration.Platform.Name,
			NextRenewalDate:      sub.NextRenewalDate.Format(layoutISO),
			DaysRemaining:    daysRemaining,
			Status:           sub.Status,
			IsRead:           sub.IsRead, // 填充 IsRead 字段
		})
	}

	if len(reminders) == 0 {
		c.JSON(http.StatusOK, gin.H{"reminders": []ReminderInfo{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reminders": reminders})
}

// MarkReminderAsRead 将指定的订阅提醒标记为已读
func MarkReminderAsRead(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		log.Printf("[MarkReminderAsRead] User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	var userID uint
	switch v := userIDInterface.(type) {
	case float64:
		userID = uint(v)
	case int64:
		userID = uint(v)
	case int:
		userID = uint(v)
	case uint:
		userID = v
	default:
		log.Printf("[MarkReminderAsRead] Unexpected user ID type in context: %T", userIDInterface)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error processing user ID"})
		return
	}

	reminderIDStr := c.Param("id")
	if reminderIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reminder ID is required"})
		return
	}

	// string 到 uint 的转换 (reminderIDStr to reminderID)
	var reminderID uint
	// 注意: 实际应用中应进行更健壮的错误处理
	_, err := fmt.Sscan(reminderIDStr, &reminderID)
	if err != nil {
		log.Printf("[MarkReminderAsRead] Invalid Reminder ID format: %s", reminderIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Reminder ID format"})
		return
	}


	db := database.DB
	var subscription models.ServiceSubscription

	// 查找属于该用户且具有指定ID的订阅
	// UserID 在 ServiceSubscription 模型中是直接的字段
	if err := db.Where("id = ? AND user_id = ?", reminderID, userID).First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found or not owned by user"})
			return
		}
		log.Printf("[MarkReminderAsRead] Error fetching reminder %d for user %d: %v", reminderID, userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reminder"})
		return
	}

	// 标记为已读并保存
	if err := db.Model(&subscription).Update("is_read", true).Error; err != nil {
		log.Printf("[MarkReminderAsRead] Error updating reminder %d to read: %v", reminderID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark reminder as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reminder marked as read successfully"})
}