package handlers

import (
	"log"
	// "encoding/json" // Not used by stubs

	"github.com/gin-gonic/gin"
	// "email_server/database" // No longer directly used by these stubbed functions
	// "email_server/models" // IMPORTANT: Removed to avoid compile errors from deleted models
	"email_server/utils"
)

const deprecatedServiceHandlerMsg = "This API endpoint related to old 'Service' models is deprecated and will be removed. Please use new API endpoints for 'Platform'."

// GetServices retrieves a paginated list of services. (DEPRECATED)
func GetServices(c *gin.Context) {
	log.Printf("Deprecated handler called: GetServices")
	utils.SendErrorResponse(c, 501, deprecatedServiceHandlerMsg+" (GetServices)")
}

// CreateService creates a new service. (DEPRECATED)
func CreateService(c *gin.Context) {
	log.Printf("Deprecated handler called: CreateService")
	utils.SendErrorResponse(c, 501, deprecatedServiceHandlerMsg+" (CreateService)")
}

// UpdateService updates an existing service. (DEPRECATED)
func UpdateService(c *gin.Context) {
	log.Printf("Deprecated handler called: UpdateService for ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedServiceHandlerMsg+" (UpdateService)")
}

// DeleteService deletes a service. (DEPRECATED)
func DeleteService(c *gin.Context) {
	log.Printf("Deprecated handler called: DeleteService for ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedServiceHandlerMsg+" (DeleteService)")
}

// GetServiceByID retrieves a single service by its ID. (DEPRECATED)
func GetServiceByID(c *gin.Context) {
	log.Printf("Deprecated handler called: GetServiceByID for ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedServiceHandlerMsg+" (GetServiceByID)")
}

// GetServiceEmails retrieves emails associated with a specific service. (DEPRECATED)
func GetServiceEmails(c *gin.Context) {
	log.Printf("Deprecated handler called: GetServiceEmails for Service ID: %s", c.Param("id"))
	utils.SendErrorResponse(c, 501, deprecatedServiceHandlerMsg+" (GetServiceEmails - use PlatformRegistration/ServiceSubscription APIs)")
}
