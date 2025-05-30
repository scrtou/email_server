package handlers

import (
	"log"
	// "time" // Not used by stubs

	"github.com/gin-gonic/gin"
	// "email_server/database" // No longer directly used by these stubbed functions
	// "email_server/models" // IMPORTANT: Removed to avoid compile errors from deleted models
	"email_server/utils"
)

const deprecatedEmailServiceHandlerMsg = "This API endpoint related to old 'EmailService' models is deprecated and will be removed. Please use new API endpoints for 'PlatformRegistration' and 'ServiceSubscription'."

// GetAllEmailServices retrieves a list of email-service associations. (DEPRECATED)
func GetAllEmailServices(c *gin.Context) {
	log.Printf("Deprecated handler called: GetAllEmailServices")
	utils.SendErrorResponse(c, 501, deprecatedEmailServiceHandlerMsg+" (GetAllEmailServices)")
}

// CreateEmailService creates a new email-service association. (DEPRECATED)
func CreateEmailService(c *gin.Context) {
	log.Printf("Deprecated handler called: CreateEmailService")
	utils.SendErrorResponse(c, 501, deprecatedEmailServiceHandlerMsg+" (CreateEmailService)")
}

// UpdateEmailService updates an existing email-service association. (DEPRECATED)
func UpdateEmailService(c *gin.Context) {
	log.Printf("Deprecated handler called: UpdateEmailService for ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedEmailServiceHandlerMsg+" (UpdateEmailService)")
}

// DeleteEmailService deletes an email-service association. (DEPRECATED)
func DeleteEmailService(c *gin.Context) {
	log.Printf("Deprecated handler called: DeleteEmailService for ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedEmailServiceHandlerMsg+" (DeleteEmailService)")
}
