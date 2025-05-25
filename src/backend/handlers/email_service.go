package handlers

import (
    "database/sql"
    "log"
    "strconv"

    "github.com/gin-gonic/gin"
    "email_server/database"
    "email_server/models"
    "email_server/utils"
)

func GetAllEmailServices(c *gin.Context) {
    page := c.DefaultQuery("page", "1")
    pageSize := c.DefaultQuery("page_size", "20")
    
    emailID := c.Query("email_id")
    serviceID := c.Query("service_id")
    status := c.DefaultQuery("status", "1")

    query := `
        SELECT 
            es.id, es.email_id, es.service_id, es.username, es.phone,
            es.registration_date, es.subscription_type, es.subscription_expires,
            es.notes, es.status, es.created_at, es.updated_at,
            e.email as email_addr, e.display_name as email_display_name,
            s.name as service_name, s.website as service_website
        FROM email_services es
        LEFT JOIN emails e ON es.email_id = e.id
        LEFT JOIN services s ON es.service_id = s.id
        WHERE es.status = ?
    `
    args := []interface{}{status}

    if emailID != "" {
        query += " AND es.email_id = ?"
        args = append(args, emailID)
    }
    if serviceID != "" {
        query += " AND es.service_id = ?"
        args = append(args, serviceID)
    }

    query += " ORDER BY es.created_at DESC"
    
    if page != "all" {
        pageNum, _ := strconv.Atoi(page)
        pageSizeNum, _ := strconv.Atoi(pageSize)
        offset := (pageNum - 1) * pageSizeNum
        query += " LIMIT ? OFFSET ?"
        args = append(args, pageSizeNum, offset)
    }

    rows, err := database.DB.Query(query, args...)
    if err != nil {
        log.Printf("查询邮箱服务关联失败: %v", err)
        utils.SendError(c, 500, "查询关联数据失败")
        return
    }
    defer rows.Close()

    var emailServices []*models.EmailService
    for rows.Next() {
        es := &models.EmailService{}
        var emailDisplayName, serviceWebsite sql.NullString
        
        err := rows.Scan(
            &es.ID, &es.EmailID, &es.ServiceID, &es.Username, &es.Phone,
            &es.RegistrationDate, &es.SubscriptionType, &es.SubscriptionExpires,
            &es.Notes, &es.Status, &es.CreatedAt, &es.UpdatedAt,
            &es.EmailAddr, &emailDisplayName, &es.ServiceName, &serviceWebsite,
        )
        if err != nil {
            log.Printf("扫描关联数据失败: %v", err)
            continue
        }
        emailServices = append(emailServices, es)
    }

    if page != "all" {
        countQuery := `
            SELECT COUNT(*) 
            FROM email_services es
            WHERE es.status = ?
        `
        countArgs := []interface{}{status}
        
        if emailID != "" {
            countQuery += " AND es.email_id = ?"
            countArgs = append(countArgs, emailID)
        }
        if serviceID != "" {
            countQuery += " AND es.service_id = ?"
            countArgs = append(countArgs, serviceID)
        }

        var total int
        err = database.DB.QueryRow(countQuery, countArgs...).Scan(&total)
        if err != nil {
            log.Printf("查询总数失败: %v", err)
        }

        result := map[string]interface{}{
            "data":       emailServices,
            "total":      total,
            "page":       page,
            "page_size":  pageSize,
        }
        utils.Success(c, result)
        return
    }

    utils.Success(c, emailServices)
}

func CreateEmailService(c *gin.Context) {
    var es models.EmailService
    if err := c.ShouldBindJSON(&es); err != nil {
        utils.SendError(c, 400, "参数错误")
        return
    }

    query := `
        INSERT INTO email_services (email_id, service_id, username, password, phone,
            registration_date, subscription_type, subscription_expires, notes)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    result, err := database.DB.Exec(query, es.EmailID, es.ServiceID, es.Username, es.Password,
        es.Phone, es.RegistrationDate, es.SubscriptionType, es.SubscriptionExpires, es.Notes)
    if err != nil {
        utils.SendError(c, 500, "创建关联失败")
        return
    }

    id, _ := result.LastInsertId()
    es.ID = id
    utils.Success(c, es)
}

func UpdateEmailService(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        utils.SendError(c, 400, "ID参数错误")
        return
    }

    var es models.EmailService
    if err := c.ShouldBindJSON(&es); err != nil {
        utils.SendError(c, 400, "参数错误")
        return
    }

    query := `
        UPDATE email_services SET username=?, password=?, phone=?, registration_date=?,
            subscription_type=?, subscription_expires=?, notes=?, updated_at=NOW()
        WHERE id=?
    `
    _, err = database.DB.Exec(query, es.Username, es.Password, es.Phone, es.RegistrationDate,
        es.SubscriptionType, es.SubscriptionExpires, es.Notes, id)
    if err != nil {
        utils.SendError(c, 500, "更新关联失败")
        return
    }

    utils.Success(c, "更新成功")
}

func DeleteEmailService(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        utils.SendError(c, 400, "ID参数错误")
        return
    }

    _, err = database.DB.Exec("UPDATE email_services SET status=0 WHERE id=?", id)
    if err != nil {
        utils.SendError(c, 500, "删除关联失败")
        return
    }

    utils.Success(c, "删除成功")
}
