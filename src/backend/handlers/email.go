package handlers

import (
	"log"

	"github.com/gin-gonic/gin"
	// "gorm.io/gorm" // No longer needed if not using gorm specific errors here
	// "email_server/database" // No longer directly used by these stubbed functions
	// "email_server/models" // IMPORTANT: Removed to avoid compile errors from deleted models
	"email_server/utils"
)

const deprecatedEmailHandlerMsg = "This API endpoint related to old 'Email' models is deprecated and will be removed. Please use new API endpoints for 'EmailAccount'."

// GetEmails retrieves a paginated list of emails for the logged-in user. (DEPRECATED)
func GetEmails(c *gin.Context) {
	log.Printf("Deprecated handler called: GetEmails")
	utils.SendErrorResponse(c, 501, deprecatedEmailHandlerMsg + " (GetEmails)")
}

// CreateEmail creates a new email entry. (DEPRECATED)
func CreateEmail(c *gin.Context) {
	log.Printf("Deprecated handler called: CreateEmail")
	utils.SendErrorResponse(c, 501, deprecatedEmailHandlerMsg + " (CreateEmail)")
}

// UpdateEmail updates an existing email entry. (DEPRECATED)
func UpdateEmail(c *gin.Context) {
	log.Printf("Deprecated handler called: UpdateEmail for ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedEmailHandlerMsg + " (UpdateEmail)")
}

// DeleteEmail marks an email as inactive. (DEPRECATED)
func DeleteEmail(c *gin.Context) {
	log.Printf("Deprecated handler called: DeleteEmail for ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedEmailHandlerMsg + " (DeleteEmail)")
}

// GetEmailByID retrieves a single email by its ID. (DEPRECATED)
func GetEmailByID(c *gin.Context) {
	log.Printf("Deprecated handler called: GetEmailByID for ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedEmailHandlerMsg + " (GetEmailByID)")
}

// GetEmailServices retrieves services associated with a specific email. (DEPRECATED)
func GetEmailServices(c *gin.Context) {
	log.Printf("Deprecated handler called: GetEmailServices for Email ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedEmailHandlerMsg + " (GetEmailServices - use PlatformRegistration/ServiceSubscription APIs)")
}
