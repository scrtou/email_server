package handlers

import (
    "strconv"

    "github.com/gin-gonic/gin"
    "email_server/database"
    "email_server/models"
    "email_server/utils"
)

func GetEmails(c *gin.Context) {
    userID, _ := c.Get("user_id")
    email := c.Query("email")
    provider := c.Query("provider")

    query := `
        SELECT e.id, e.email, e.display_name, e.provider, e.phone, e.backup_email, 
               e.notes, e.status, e.created_at, e.updated_at, COUNT(es.id) as service_count
        FROM emails e
        LEFT JOIN email_services es ON e.id = es.email_id AND es.status = 1
        WHERE e.status = 1 AND e.user_id = ?
    `
    args := []interface{}{userID}

    if email != "" {
        query += " AND e.email LIKE ?"
        args = append(args, "%"+email+"%")
    }
    if provider != "" {
        query += " AND e.provider = ?"
        args = append(args, provider)
    }

    query += " GROUP BY e.id ORDER BY e.created_at DESC"

    rows, err := database.DB.Query(query, args...)
    if err != nil {
        utils.SendError(c, 500, "查询邮箱失败")
        return
    }
    defer rows.Close()

    var emails []*models.Email
    for rows.Next() {
        email := &models.Email{}
        err := rows.Scan(&email.ID, &email.Email, &email.DisplayName, &email.Provider,
            &email.Phone, &email.BackupEmail, &email.Notes, &email.Status,
            &email.CreatedAt, &email.UpdatedAt, &email.ServiceCount)
        if err == nil {
            emails = append(emails, email)
        }
    }

    utils.Success(c, emails)
}

func CreateEmail(c *gin.Context) {
    userID, _ := c.Get("user_id")
    
    var email models.Email
    if err := c.ShouldBindJSON(&email); err != nil {
        utils.SendError(c, 400, "参数错误")
        return
    }

    query := `
        INSERT INTO emails (user_id, email, password, display_name, provider, phone, backup_email, 
            security_question, security_answer, notes)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    result, err := database.DB.Exec(query, userID, email.Email, email.Password, email.DisplayName,
        email.Provider, email.Phone, email.BackupEmail, email.SecurityQ,
        email.SecurityA, email.Notes)
    if err != nil {
        utils.SendError(c, 500, "创建邮箱失败")
        return
    }

    id, _ := result.LastInsertId()
    email.ID = id
    utils.Success(c, email)
}

func UpdateEmail(c *gin.Context) {
    userID, _ := c.Get("user_id")
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        utils.SendError(c, 400, "ID参数错误")
        return
    }

    var email models.Email
    if err := c.ShouldBindJSON(&email); err != nil {
        utils.SendError(c, 400, "参数错误")
        return
    }

    query := `
        UPDATE emails SET email=?, password=?, display_name=?, provider=?, phone=?, 
            backup_email=?, security_question=?, security_answer=?, notes=?, updated_at=NOW()
        WHERE id=? AND user_id=?
    `
    result, err := database.DB.Exec(query, email.Email, email.Password, email.DisplayName,
        email.Provider, email.Phone, email.BackupEmail, email.SecurityQ,
        email.SecurityA, email.Notes, id, userID)
    if err != nil {
        utils.SendError(c, 500, "更新邮箱失败")
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        utils.SendError(c, 404, "邮箱不存在或无权限")
        return
    }

    utils.Success(c, "更新成功")
}

func DeleteEmail(c *gin.Context) {
    userID, _ := c.Get("user_id")
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        utils.SendError(c, 400, "ID参数错误")
        return
    }

    result, err := database.DB.Exec("UPDATE emails SET status=0 WHERE id=? AND user_id=?", id, userID)
    if err != nil {
        utils.SendError(c, 500, "删除邮箱失败")
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        utils.SendError(c, 404, "邮箱不存在或无权限")
        return
    }

    utils.Success(c, "删除成功")
}

func GetEmailByID(c *gin.Context) {
    userID, _ := c.Get("user_id")
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        utils.SendError(c, 400, "ID参数错误")
        return
    }

    var email models.Email
    query := `
        SELECT id, email, display_name, provider, phone, backup_email, 
               security_question, notes, status, created_at, updated_at
        FROM emails WHERE id=? AND user_id=? AND status=1
    `
    err = database.DB.QueryRow(query, id, userID).Scan(&email.ID, &email.Email, &email.DisplayName,
        &email.Provider, &email.Phone, &email.BackupEmail, &email.SecurityQ,
        &email.Notes, &email.Status, &email.CreatedAt, &email.UpdatedAt)
    if err != nil {
        utils.SendError(c, 404, "邮箱不存在")
        return
    }

    utils.Success(c, email)
}

func GetEmailServices(c *gin.Context) {
    userID, _ := c.Get("user_id")
    emailID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        utils.SendError(c, 400, "ID参数错误")
        return
    }

    // 首先验证邮箱是否属于当前用户
    var count int
    err = database.DB.QueryRow("SELECT COUNT(*) FROM emails WHERE id = ? AND user_id = ? AND status = 1", 
        emailID, userID).Scan(&count)
    if err != nil || count == 0 {
        utils.SendError(c, 404, "邮箱不存在或无权限")
        return
    }

    query := `
        SELECT es.id, es.email_id, es.service_id, es.username, es.phone, 
               es.registration_date, es.subscription_type, es.subscription_expires,
               es.notes, es.status, es.created_at, es.updated_at, s.name as service_name
        FROM email_services es
        JOIN services s ON es.service_id = s.id
        WHERE es.email_id = ? AND es.status = 1
        ORDER BY es.created_at DESC
    `
    rows, err := database.DB.Query(query, emailID)
    if err != nil {
        utils.SendError(c, 500, "查询邮箱服务失败")
        return
    }
    defer rows.Close()

    var emailServices []*models.EmailService
    for rows.Next() {
        es := &models.EmailService{}
        err := rows.Scan(&es.ID, &es.EmailID, &es.ServiceID, &es.Username, &es.Phone,
            &es.RegistrationDate, &es.SubscriptionType, &es.SubscriptionExpires,
            &es.Notes, &es.Status, &es.CreatedAt, &es.UpdatedAt, &es.ServiceName)
        if err == nil {
            emailServices = append(emailServices, es)
        }
    }

    utils.Success(c, emailServices)
}
