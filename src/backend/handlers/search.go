package handlers

import (
	"net/http"
	"strings"

	"email_server/database"
	"email_server/models"
	"email_server/utils"

	"github.com/gin-gonic/gin"
	// "gorm.io/gorm" // Removed as it's not directly used in this file after recent changes
)

// SearchHandler handles global search requests.
// GET /search?q=...&type=...
func SearchHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User ID not found in token")
		return
	}
	currentUserID := userID.(uint)

	query := strings.ToLower(strings.TrimSpace(c.Query("q")))
	searchType := strings.ToLower(strings.TrimSpace(c.Query("type")))

	if query == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Search query 'q' cannot be empty")
		return
	}

	var results []models.GlobalSearchResultItem
	db := database.DB

	// Search Users (example: by username, email) - Assuming User model has Username and Email fields
	// Note: In a real application, searching all users might be an admin-only feature.
	// For this task, we'll allow searching users but the results might be limited or need careful consideration for privacy.
	if searchType == "users" || searchType == "all" || searchType == "" {
		var users []models.User
		// Adjust fields based on your User model
		if err := db.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ?", "%"+query+"%", "%"+query+"%").Find(&users).Error; err == nil {
			for _, u := range users {
				// Only return current user's information if searching users, or if it's an admin search (not implemented here)
				// This is a simplified approach. For non-admin users, they should probably only find themselves.
				if u.ID == currentUserID { // Simplified: only show current user if "users" type is searched by non-admin
					results = append(results, models.GlobalSearchResultItem{
						ID:          u.ID,
						Type:        "user",
						DisplayName: u.Username,
						Details: gin.H{
							"email": u.Email,
						},
					})
				}
			}
		}
	}

	// Search EmailAccounts (by email address, provider)
	if searchType == "email_accounts" || searchType == "all" || searchType == "" {
		var emailAccounts []models.EmailAccount
		if err := db.Where("user_id = ? AND (LOWER(email_address) LIKE ? OR LOWER(provider) LIKE ?)", currentUserID, "%"+query+"%", "%"+query+"%").Find(&emailAccounts).Error; err == nil {
			for _, ea := range emailAccounts {
				results = append(results, models.GlobalSearchResultItem{
					ID:          ea.ID,
					Type:        "email_account",
					DisplayName: ea.EmailAddress,
					Details: gin.H{
						"provider": ea.Provider,
					},
				})
			}
		}
	}

	// Search Platforms (by name, url) - Public data
	if searchType == "platforms" || searchType == "all" || searchType == "" {
		var platforms []models.Platform
		if err := db.Where("LOWER(name) LIKE ? OR LOWER(website) LIKE ?", "%"+query+"%", "%"+query+"%").Find(&platforms).Error; err == nil {
			for _, p := range platforms {
				results = append(results, models.GlobalSearchResultItem{
					ID:          p.ID,
					Type:        "platform",
					DisplayName: p.Name,
					Details: gin.H{
						"website_url": p.WebsiteURL,
						// "category": p.Category, // Category field does not exist in Platform model
					},
				})
			}
		}
	}

	// Search PlatformRegistrations (by username, notes)
	if searchType == "platform_registrations" || searchType == "all" || searchType == "" {
		var platformRegistrations []models.PlatformRegistration
		// Need to join with EmailAccount to ensure UserID match for ownership
		// Or, if PlatformRegistration directly has UserID (which it should for direct ownership)
		// Assuming PlatformRegistration has UserID. If not, this query needs adjustment.
		// Let's assume PlatformRegistration has a UserID field for direct ownership.
		// If PlatformRegistration is linked via EmailAccount, the query would be more complex.
		// For simplicity, assuming direct UserID on PlatformRegistration or a join is handled by GORM's Preload or similar.
		// The task implies user-specific entities, so PlatformRegistration should be filtered by currentUserID.
		// A direct UserID on PlatformRegistration is the most straightforward for this.
		// If it's indirect (PlatformRegistration -> EmailAccount -> User), the query needs joins.
		// Let's assume PlatformRegistration has UserID.
		// If not, we'll need to adjust. Reading the model file for PlatformRegistration would clarify.
		// For now, proceeding with assumption of direct UserID.

		// Correct approach: PlatformRegistrations are linked to EmailAccounts, which are linked to Users.
		// So we need a join.
		if err := db.Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id").
			Where("email_accounts.user_id = ? AND (LOWER(platform_registrations.username) LIKE ? OR LOWER(platform_registrations.notes) LIKE ?)", currentUserID, "%"+query+"%", "%"+query+"%").
			Preload("Platform"). // Preload platform to get its name for display
			Find(&platformRegistrations).Error; err == nil {
			for _, pr := range platformRegistrations {
				displayName := ""
				if pr.LoginUsername != nil {
					displayName = *pr.LoginUsername
				}
				if pr.Platform.Name != "" {
					if displayName != "" {
						displayName = displayName + " @ " + pr.Platform.Name
					} else {
						displayName = pr.Platform.Name
					}
				}
				results = append(results, models.GlobalSearchResultItem{
					ID:          pr.ID,
					Type:        "platform_registration",
					DisplayName: displayName,
					Details: gin.H{
						"platform_id":   pr.PlatformID,
						"platform_name": pr.Platform.Name, // Assumes Platform is preloaded
						"notes":         pr.Notes,
					},
				})
			}
		}
	}

	// Search ServiceSubscriptions (by service name, notes)
	if searchType == "service_subscriptions" || searchType == "all" || searchType == "" {
		var serviceSubscriptions []models.ServiceSubscription
		// Assuming ServiceSubscription has UserID.
		// If linked via PlatformRegistration -> EmailAccount -> User, this also needs joins.
		// The task implies user-specific entities.
		// Let's assume ServiceSubscription has UserID.
		// Reading the model file for ServiceSubscription would clarify.
		// For now, proceeding with assumption of direct UserID.

		// Correct approach: ServiceSubscriptions are linked to PlatformRegistrations,
		// which are linked to EmailAccounts, which are linked to Users. This is a multi-level join.
		// Or, if ServiceSubscription has a direct UserID.
		// Given the structure, it's more likely linked.
		// Let's assume ServiceSubscription has a UserID field for simplicity of this search handler first.
		// If not, this will need significant join clauses.
		// The prompt says "ensure on user specific entities ... all ownership validation"
		// This implies ServiceSubscription should be filtered by currentUserID.

		// Simpler: Assume ServiceSubscription has UserID.
		// If not, the join would be:
		// JOIN platform_registrations ON platform_registrations.id = service_subscriptions.platform_registration_id
		// JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id
		// WHERE email_accounts.user_id = ?
		// This is getting complex for a single search handler without helper functions or more direct UserID links.
		// For now, let's assume ServiceSubscription has a UserID field.
		// This is a common simplification if performance of deep joins is a concern for search.
		// If ServiceSubscription.UserID does not exist, this part will fail and need rework.

		// Revisiting: The most robust way is to join.
		err := db.Joins("JOIN platform_registrations ON platform_registrations.id = service_subscriptions.platform_registration_id").
			Joins("JOIN email_accounts ON email_accounts.id = platform_registrations.email_account_id").
			Where("email_accounts.user_id = ? AND (LOWER(service_subscriptions.service_name) LIKE ? OR LOWER(service_subscriptions.notes) LIKE ?)", currentUserID, "%"+query+"%", "%"+query+"%").
			Preload("PlatformRegistration.Platform"). // For context
			Find(&serviceSubscriptions).Error

		if err == nil {
			for _, ss := range serviceSubscriptions {
				displayName := ss.ServiceName
				if ss.PlatformRegistration.Platform.Name != "" {
					displayName = ss.ServiceName + " (" + ss.PlatformRegistration.Platform.Name + ")"
				}

				results = append(results, models.GlobalSearchResultItem{
					ID:          ss.ID,
					Type:        "service_subscription",
					DisplayName: displayName,
					Details: gin.H{
						"status":                   ss.Status,
						"billing_cycle":            ss.BillingCycle,
						"next_renewal_date":        ss.NextRenewalDate,
						"platform_registration_id": ss.PlatformRegistrationID,
					},
				})
			}
		} else {
			// Log error if any during DB operation for service_subscriptions
			// c.Error(err) // Or some other logging
		}
	}

	if len(results) == 0 {
		utils.SendSuccessResponse(c, models.GlobalSearchResponse{Results: []models.GlobalSearchResultItem{}})
		return
	}

	utils.SendSuccessResponse(c, models.GlobalSearchResponse{Results: results})
}
