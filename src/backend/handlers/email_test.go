package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"email_server/database"
	"email_server/integrations"
	"email_server/models"
	"email_server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestRouter sets up a test router with a mock database and necessary routes.
func setupTestRouter() (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Setup in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto-migrate schema
	db.AutoMigrate(&models.User{}, &models.EmailAccount{})
	database.DB = db

	// Setup routes
	r.GET("/inbox", AuthRequiredTest(), GetInbox)

	return r, db
}

// AuthRequiredTest is a test middleware to simulate an authenticated user.
func AuthRequiredTest() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", uint(1)) // Set a mock user ID
		c.Next()
	}
}

// This is a package-level variable that can be swapped out in tests.
var fetchEmails = integrations.FetchEmails

func TestGetInbox_Success(t *testing.T) {
	router, db := setupTestRouter()

	// Create a test user and email account
	testUser := models.User{Model: gorm.Model{ID: 1}, Username: "testuser"}
	db.Create(&testUser)
	testEmailAccount := models.EmailAccount{
		UserID:            1,
		EmailAddress:      "test@example.com",
		PasswordEncrypted: "encrypted_password",
		IMAPServer:        "imap.example.com",
		IMAPPort:          993,
	}
	db.Create(&testEmailAccount)

	// Mock the FetchEmails function
	originalFetchEmails := fetchEmails
	fetchEmails = func(account models.EmailAccount, page, pageSize int) ([]models.Email, int, error) {
		return []models.Email{{Subject: "Mocked Email"}}, 1, nil
	}
	defer func() { fetchEmails = originalFetchEmails }()

	// Mock password decryption
	originalDecryptPassword := utils.DecryptPassword
	utils.DecryptPassword = func(encrypted string) (string, error) {
		return "password", nil
	}
	defer func() { utils.DecryptPassword = originalDecryptPassword }()


	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/inbox", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response gin.H
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.NotNil(t, response["data"])
	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)

	assert.Equal(t, float64(1), data["total"])
	emails, ok := data["emails"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, emails, 1)
}


func TestGetInbox_NoEmailAccount(t *testing.T) {
	router, db := setupTestRouter()

	// Create a test user but no email account
	testUser := models.User{Model: gorm.Model{ID: 1}, Username: "testuser"}
	db.Create(&testUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/inbox", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetInbox_FetchError(t *testing.T) {
	router, db := setupTestRouter()

	// Create a test user and email account
	testUser := models.User{Model: gorm.Model{ID: 1}, Username: "testuser"}
	db.Create(&testUser)
	testEmailAccount := models.EmailAccount{
		UserID:            1,
		EmailAddress:      "test@example.com",
		PasswordEncrypted: "encrypted_password",
		IMAPServer:        "imap.example.com",
		IMAPPort:          993,
	}
	db.Create(&testEmailAccount)

	// Mock FetchEmails to return an error
	originalFetchEmails := fetchEmails
	fetchEmails = func(account models.EmailAccount, page, pageSize int) ([]models.Email, int, error) {
		return nil, 0, errors.New("failed to fetch")
	}
	defer func() { fetchEmails = originalFetchEmails }()

	// Mock password decryption
	originalDecryptPassword := utils.DecryptPassword
	utils.DecryptPassword = func(encrypted string) (string, error) {
		return "password", nil
	}
	defer func() { utils.DecryptPassword = originalDecryptPassword }()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/inbox", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}