package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"email_server/database"
	"email_server/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() {
	var err error
	database.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移数据表
	err = database.DB.AutoMigrate(
		&models.User{},
		&models.EmailAccount{},
		&models.Platform{},
		&models.PlatformRegistration{},
		&models.ServiceSubscription{},
	)
	if err != nil {
		panic("failed to migrate database")
	}
}

func createTestUser() *models.User {
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     "user",
		Status:   1,
	}
	database.DB.Create(user)
	return user
}

func TestServiceSubscriptionCreation(t *testing.T) {
	// 设置测试数据库
	setupTestDB()
	
	// 创建测试用户
	user := createTestUser()

	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	// 测试用例1: 邮箱地址为空，用户名不为空
	t.Run("Case1: Empty email, with username", func(t *testing.T) {
		payload := models.CreateServiceSubscriptionRequest{
			PlatformName:  "GitHub",
			LoginUsername: "testuser123",
			ServiceName:   "GitHub Pro",
			Description:   "GitHub专业版订阅",
			Status:        "active",
			Cost:          4.0,
			BillingCycle:  "monthly",
		}

		jsonPayload, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/service-subscriptions", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("user_id", "1") // 模拟认证中间件设置的用户ID

		w := httptest.NewRecorder()
		
		// 创建一个带有认证中间件模拟的上下文
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", int64(user.ID))

		// 直接调用处理函数
		CreateServiceSubscription(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])
		
		// 验证数据库中的记录
		var platform models.Platform
		database.DB.Where("name = ? AND user_id = ?", "GitHub", user.ID).First(&platform)
		assert.Equal(t, "GitHub", platform.Name)
		
		var platformReg models.PlatformRegistration
		database.DB.Where("user_id = ? AND platform_id = ? AND login_username = ?", user.ID, platform.ID, "testuser123").First(&platformReg)
		assert.Equal(t, "testuser123", *platformReg.LoginUsername)
		
		var subscription models.ServiceSubscription
		database.DB.Where("user_id = ? AND platform_registration_id = ? AND service_name = ?", user.ID, platformReg.ID, "GitHub Pro").First(&subscription)
		assert.Equal(t, "GitHub Pro", subscription.ServiceName)
	})

	// 测试用例2: 邮箱地址不为空，用户名为空
	t.Run("Case2: With email, empty username", func(t *testing.T) {
		payload := models.CreateServiceSubscriptionRequest{
			PlatformName: "Netflix",
			EmailAddress: "user@example.com",
			ServiceName:  "Netflix Premium",
			Description:  "Netflix高级订阅",
			Status:       "active",
			Cost:         15.99,
			BillingCycle: "monthly",
		}

		jsonPayload, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/service-subscriptions", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", int64(user.ID))

		CreateServiceSubscription(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		// 验证邮箱记录被创建
		var emailAccount models.EmailAccount
		database.DB.Where("user_id = ? AND email_address = ?", user.ID, "user@example.com").First(&emailAccount)
		assert.Equal(t, "user@example.com", emailAccount.EmailAddress)
	})

	// 测试用例3: 邮箱地址和用户名都不为空
	t.Run("Case3: With both email and username", func(t *testing.T) {
		payload := models.CreateServiceSubscriptionRequest{
			PlatformName:  "Discord",
			EmailAddress:  "discord@example.com",
			LoginUsername: "discorduser",
			ServiceName:   "Discord Nitro",
			Description:   "Discord Nitro订阅",
			Status:        "active",
			Cost:          9.99,
			BillingCycle:  "monthly",
		}

		jsonPayload, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/service-subscriptions", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", int64(user.ID))

		CreateServiceSubscription(c)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	// 测试用例4: 邮箱地址和用户名都为空（错误用例）
	t.Run("Case4: Both email and username empty - should fail", func(t *testing.T) {
		payload := models.CreateServiceSubscriptionRequest{
			PlatformName: "Steam",
			ServiceName:  "Steam Game",
			Description:  "Steam游戏订阅",
			Status:       "active",
			Cost:         59.99,
			BillingCycle: "onetime",
		}

		jsonPayload, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "/api/v1/service-subscriptions", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("user_id", int64(user.ID))

		CreateServiceSubscription(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["message"], "邮箱地址和用户名不能都为空")
	})
}
