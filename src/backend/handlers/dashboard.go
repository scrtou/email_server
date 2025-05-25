package handlers

import (
    "github.com/gin-gonic/gin"
    "email_server/database"
    "email_server/models"
    "email_server/utils"
)

func GetDashboard(c *gin.Context) {
    dashboard := &models.DashboardData{
        EmailsByProvider:   make(map[string]int),
        ServicesByCategory: make(map[string]int),
    }

    err := database.DB.QueryRow("SELECT COUNT(*) FROM emails WHERE status = 1").Scan(&dashboard.EmailCount)
    if err != nil {
        utils.SendError(c, 500, "获取邮箱数量失败")
        return
    }

    err = database.DB.QueryRow("SELECT COUNT(*) FROM services").Scan(&dashboard.ServiceCount)
    if err != nil {
        utils.SendError(c, 500, "获取服务数量失败")
        return
    }

    err = database.DB.QueryRow("SELECT COUNT(*) FROM email_services WHERE status = 1").Scan(&dashboard.RelationCount)
    if err != nil {
        utils.SendError(c, 500, "获取关联数量失败")
        return
    }

    rows, err := database.DB.Query("SELECT provider, COUNT(*) FROM emails WHERE status = 1 GROUP BY provider")
    if err == nil {
        defer rows.Close()
        for rows.Next() {
            var provider string
            var count int
            if err := rows.Scan(&provider, &count); err == nil {
                dashboard.EmailsByProvider[provider] = count
            }
        }
    }

    rows, err = database.DB.Query("SELECT category, COUNT(*) FROM services GROUP BY category")
    if err == nil {
        defer rows.Close()
        for rows.Next() {
            var category string
            var count int
            if err := rows.Scan(&category, &count); err == nil {
                dashboard.ServicesByCategory[category] = count
            }
        }
    }

    dashboard.RecentEmails = GetRecentEmails()
    dashboard.RecentServices = GetRecentServices()

    utils.Success(c, dashboard)
}

func GetRecentEmails() []*models.Email {
    query := `
        SELECT e.id, e.email, e.display_name, e.provider, e.phone, e.created_at, COUNT(es.id) as service_count
        FROM emails e
        LEFT JOIN email_services es ON e.id = es.email_id AND es.status = 1
        WHERE e.status = 1
        GROUP BY e.id
        ORDER BY e.created_at DESC
        LIMIT 5
    `
    rows, err := database.DB.Query(query)
    if err != nil {
        return []*models.Email{}
    }
    defer rows.Close()

    var emails []*models.Email
    for rows.Next() {
        email := &models.Email{}
        err := rows.Scan(&email.ID, &email.Email, &email.DisplayName, &email.Provider,
            &email.Phone, &email.CreatedAt, &email.ServiceCount)
        if err == nil {
            emails = append(emails, email)
        }
    }
    return emails
}

func GetRecentServices() []*models.Service {
    query := `
        SELECT s.id, s.name, s.website, s.category, s.description, s.created_at, COUNT(es.id) as email_count
        FROM services s
        LEFT JOIN email_services es ON s.id = es.service_id AND es.status = 1
        GROUP BY s.id
        ORDER BY s.created_at DESC
        LIMIT 5
    `
    rows, err := database.DB.Query(query)
    if err != nil {
        return []*models.Service{}
    }
    defer rows.Close()

    var services []*models.Service
    for rows.Next() {
        service := &models.Service{}
        err := rows.Scan(&service.ID, &service.Name, &service.Website, &service.Category,
            &service.Description, &service.CreatedAt, &service.EmailCount)
        if err == nil {
            services = append(services, service)
        }
    }
    return services
}
