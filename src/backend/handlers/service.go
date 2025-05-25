package handlers

import (
    "strconv"

    "github.com/gin-gonic/gin"
    "email_server/database"
    "email_server/models"
    "email_server/utils"
)

func GetServices(c *gin.Context) {
    name := c.Query("name")
    category := c.Query("category")

    query := `
        SELECT s.id, s.name, s.website, s.category, s.description, s.logo_url,
               s.created_at, s.updated_at, COUNT(es.id) as email_count
        FROM services s
        LEFT JOIN email_services es ON s.id = es.service_id AND es.status = 1
        WHERE 1=1
    `
    args := []interface{}{}

    if name != "" {
        query += " AND s.name LIKE ?"
        args = append(args, "%"+name+"%")
    }
    if category != "" {
        query += " AND s.category = ?"
        args = append(args, category)
    }

    query += " GROUP BY s.id ORDER BY s.created_at DESC"

    rows, err := database.DB.Query(query, args...)
    if err != nil {
        utils.SendError(c, 500, "查询服务失败")
        return
    }
    defer rows.Close()

    var services []*models.Service
    for rows.Next() {
        service := &models.Service{}
        err := rows.Scan(&service.ID, &service.Name, &service.Website, &service.Category,
            &service.Description, &service.LogoURL, &service.CreatedAt, &service.UpdatedAt,
            &service.EmailCount)
        if err == nil {
            services = append(services, service)
        }
    }

    utils.Success(c, services)
}

func CreateService(c *gin.Context) {
    var service models.Service
    if err := c.ShouldBindJSON(&service); err != nil {
        utils.SendError(c, 400, "参数错误")
        return
    }

    query := `
        INSERT INTO services (name, website, category, description, logo_url)
        VALUES (?, ?, ?, ?, ?)
    `
    result, err := database.DB.Exec(query, service.Name, service.Website, service.Category,
        service.Description, service.LogoURL)
    if err != nil {
        utils.SendError(c, 500, "创建服务失败")
        return
    }

    id, _ := result.LastInsertId()
    service.ID = id
    utils.Success(c, service)
}

func UpdateService(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        utils.SendError(c, 400, "ID参数错误")
        return
    }

    var service models.Service
    if err := c.ShouldBindJSON(&service); err != nil {
        utils.SendError(c, 400, "参数错误")
        return
    }

    query := `
        UPDATE services SET name=?, website=?, category=?, description=?, 
            logo_url=?, updated_at=NOW()
        WHERE id=?
    `
    _, err = database.DB.Exec(query, service.Name, service.Website, service.Category,
        service.Description, service.LogoURL, id)
    if err != nil {
        utils.SendError(c, 500, "更新服务失败")
        return
    }

    utils.Success(c, "更新成功")
}

func DeleteService(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        utils.SendError(c, 400, "ID参数错误")
        return
    }

    _, err = database.DB.Exec("DELETE FROM services WHERE id=?", id)
    if err != nil {
        utils.SendError(c, 500, "删除服务失败")
        return
    }

    utils.Success(c, "删除成功")
}

func GetServiceByID(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        utils.SendError(c, 400, "ID参数错误")
        return
    }

    var service models.Service
    err = database.DB.QueryRow("SELECT id, name, website, category, description, logo_url, created_at, updated_at FROM services WHERE id=?", id).
        Scan(&service.ID, &service.Name, &service.Website, &service.Category,
            &service.Description, &service.LogoURL, &service.CreatedAt, &service.UpdatedAt)
    if err != nil {
        utils.SendError(c, 404, "服务不存在")
        return
    }

    utils.Success(c, service)
}

func GetServiceEmails(c *gin.Context) {
    serviceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        utils.SendError(c, 400, "ID参数错误")
        return
    }

    query := `
        SELECT es.id, es.email_id, es.service_id, es.username, es.phone,
               es.registration_date, es.subscription_type, es.subscription_expires,
               es.notes, es.status, es.created_at, es.updated_at, e.email
        FROM email_services es
        JOIN emails e ON es.email_id = e.id
        WHERE es.service_id = ? AND es.status = 1 AND e.status = 1
        ORDER BY es.created_at DESC
    `
    rows, err := database.DB.Query(query, serviceID)
    if err != nil {
        utils.SendError(c, 500, "查询服务邮箱失败")
        return
    }
    defer rows.Close()

    var serviceEmails []*models.EmailService
    for rows.Next() {
        se := &models.EmailService{}
        err := rows.Scan(&se.ID, &se.EmailID, &se.ServiceID, &se.Username, &se.Phone,
            &se.RegistrationDate, &se.SubscriptionType, &se.SubscriptionExpires,
            &se.Notes, &se.Status, &se.CreatedAt, &se.UpdatedAt, &se.EmailAddr)
        if err == nil {
            serviceEmails = append(serviceEmails, se)
        }
    }

    utils.Success(c, serviceEmails)
}
