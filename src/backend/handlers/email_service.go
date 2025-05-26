package handlers

import (
    //"database/sql" // 保留这个导入，因为需要处理NULL值
    "log"
    "strconv"
    //"strings"
    "time"

    "github.com/gin-gonic/gin"
    "email_server/database"
    "email_server/models"
    "email_server/utils"
)

func GetAllEmailServices(c *gin.Context) {
    startTime := time.Now()
    defer func() {
        log.Printf("GetAllEmailServices 耗时: %v", time.Since(startTime))
    }()

    page := c.DefaultQuery("page", "1")
    pageSize := c.DefaultQuery("page_size", "20")
    
    emailID := c.Query("email_id")
    serviceID := c.Query("service_id")
    status := c.DefaultQuery("status", "1")

    // 参数验证
    if status != "0" && status != "1" {
        utils.SendError(c, 400, "状态参数无效")
        return
    }

    query := `
        SELECT 
            es.id, es.email_id, es.service_id, es.username, es.phone,
            es.registration_date, es.subscription_type, es.subscription_expires,
            es.notes, es.status, es.created_at, es.updated_at,
            COALESCE(e.email, '') as email_addr, 
            COALESCE(e.display_name, '') as email_display_name,
            COALESCE(s.name, '') as service_name, 
            COALESCE(s.website, '') as service_website
        FROM email_services es
        LEFT JOIN emails e ON es.email_id = e.id AND e.status = 1
        LEFT JOIN services s ON es.service_id = s.id
        WHERE es.status = ?
    `
    args := []interface{}{status}

    if emailID != "" {
        if _, err := strconv.ParseInt(emailID, 10, 64); err != nil {
            utils.SendError(c, 400, "邮箱ID参数无效")
            return
        }
        query += " AND es.email_id = ?"
        args = append(args, emailID)
    }
    
    if serviceID != "" {
        if _, err := strconv.ParseInt(serviceID, 10, 64); err != nil {
            utils.SendError(c, 400, "服务ID参数无效")
            return
        }
        query += " AND es.service_id = ?"
        args = append(args, serviceID)
    }

    query += " ORDER BY es.created_at DESC"
    
    var total int
    var emailServices []*models.EmailService
    
    if page != "all" {
        // 参数验证和处理
        pageNum, err := strconv.Atoi(page)
        if err != nil || pageNum < 1 {
            utils.SendError(c, 400, "页码参数无效")
            return
        }
        
        pageSizeNum, err := strconv.Atoi(pageSize)
        if err != nil || pageSizeNum < 1 || pageSizeNum > 100 {
            utils.SendError(c, 400, "页面大小参数无效")
            return
        }

        // 查询总数
        countQuery := `
            SELECT COUNT(*) 
            FROM email_services es
            LEFT JOIN emails e ON es.email_id = e.id AND e.status = 1
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

        err = database.DB.QueryRow(countQuery, countArgs...).Scan(&total)
        if err != nil {
            log.Printf("查询总数失败: %v", err)
            utils.SendError(c, 500, "查询总数失败")
            return
        }

        // 添加分页
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

    for rows.Next() {
        es := &models.EmailService{}
        
        err := rows.Scan(
            &es.ID, &es.EmailID, &es.ServiceID, &es.Username, &es.Phone,
            &es.RegistrationDate, &es.SubscriptionType, &es.SubscriptionExpires,
            &es.Notes, &es.Status, &es.CreatedAt, &es.UpdatedAt,
            &es.EmailAddr, &es.EmailDisplayName, &es.ServiceName, &es.ServiceWebsite,
        )
        if err != nil {
            log.Printf("扫描关联数据失败: %v", err)
            continue
        }
        emailServices = append(emailServices, es)
    }

    if err = rows.Err(); err != nil {
        log.Printf("遍历结果集时出错: %v", err)
        utils.SendError(c, 500, "处理查询结果失败")
        return
    }

    if page != "all" {
        pageNum, _ := strconv.Atoi(page)
        pageSizeNum, _ := strconv.Atoi(pageSize)
        
        result := map[string]interface{}{
            "data":       emailServices,
            "total":      total,
            "page":       pageNum,
            "page_size":  pageSizeNum,
            "has_next":   (pageNum * pageSizeNum) < total,
            "has_prev":   pageNum > 1,
        }
        utils.Success(c, result)
    } else {
        utils.Success(c, emailServices)
    }
}

func CreateEmailService(c *gin.Context) {
    var es models.EmailService
    if err := c.ShouldBindJSON(&es); err != nil {
        log.Printf("参数绑定失败: %v", err)
        utils.SendError(c, 400, "参数格式错误")
        return
    }

    // 参数验证
    if es.EmailID <= 0 {
        utils.SendError(c, 400, "邮箱ID无效")
        return
    }
    if es.ServiceID <= 0 {
        utils.SendError(c, 400, "服务ID无效")
        return
    }

    // 检查邮箱和服务是否存在
    var emailExists, serviceExists bool
    err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM emails WHERE id = ? AND status = 1)", 
        es.EmailID).Scan(&emailExists)
    if err != nil || !emailExists {
        utils.SendError(c, 400, "指定的邮箱不存在")
        return
    }

    err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM services WHERE id = ?)", 
        es.ServiceID).Scan(&serviceExists)
    if err != nil || !serviceExists {
        utils.SendError(c, 400, "指定的服务不存在")
        return
    }

    // 检查是否已存在相同的关联
    var existsCount int
    err = database.DB.QueryRow(
        "SELECT COUNT(*) FROM email_services WHERE email_id = ? AND service_id = ? AND status = 1",
        es.EmailID, es.ServiceID).Scan(&existsCount)
    if err != nil {
        log.Printf("检查关联是否存在失败: %v", err)
        utils.SendError(c, 500, "检查关联状态失败")
        return
    }
    if existsCount > 0 {
        utils.SendError(c, 400, "该邮箱与服务的关联已存在")
        return
    }

    query := `
        INSERT INTO email_services (email_id, service_id, username, password, phone,
            registration_date, subscription_type, subscription_expires, notes, status, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, NOW(), NOW())
    `
    result, err := database.DB.Exec(query, es.EmailID, es.ServiceID, es.Username, es.Password,
        es.Phone, es.RegistrationDate, es.SubscriptionType, es.SubscriptionExpires, es.Notes)
    if err != nil {
        log.Printf("创建关联失败: %v", err)
        utils.SendError(c, 500, "创建关联失败")
        return
    }

    id, _ := result.LastInsertId()
    es.ID = id
    utils.Success(c, es)
}

func UpdateEmailService(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil || id <= 0 {
        utils.SendError(c, 400, "ID参数无效")
        return
    }

    var es models.EmailService
    if err := c.ShouldBindJSON(&es); err != nil {
        log.Printf("参数绑定失败: %v", err)
        utils.SendError(c, 400, "参数格式错误")
        return
    }

    // 检查记录是否存在
    var exists bool
    err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM email_services WHERE id = ? AND status = 1)", 
        id).Scan(&exists)
    if err != nil {
        log.Printf("检查记录是否存在失败: %v", err)
        utils.SendError(c, 500, "检查记录状态失败")
        return
    }
    if !exists {
        utils.SendError(c, 404, "指定的关联记录不存在")
        return
    }

    query := `
        UPDATE email_services SET username=?, password=?, phone=?, registration_date=?,
            subscription_type=?, subscription_expires=?, notes=?, updated_at=NOW()
        WHERE id=? AND status=1
    `
    result, err := database.DB.Exec(query, es.Username, es.Password, es.Phone, es.RegistrationDate,
        es.SubscriptionType, es.SubscriptionExpires, es.Notes, id)
    if err != nil {
        log.Printf("更新关联失败: %v", err)
        utils.SendError(c, 500, "更新关联失败")
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        utils.SendError(c, 404, "记录不存在或已被删除")
        return
    }

    utils.Success(c, "更新成功")
}

func DeleteEmailService(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil || id <= 0 {
        utils.SendError(c, 400, "ID参数无效")
        return
    }

    result, err := database.DB.Exec(
        "UPDATE email_services SET status=0, updated_at=NOW() WHERE id=? AND status=1", id)
    if err != nil {
        log.Printf("删除关联失败: %v", err)
        utils.SendError(c, 500, "删除关联失败")
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        utils.SendError(c, 404, "记录不存在或已被删除")
        return
    }

    utils.Success(c, "删除成功")
}
