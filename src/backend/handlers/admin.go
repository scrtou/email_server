package handlers

import (
    "database/sql"
    "strconv"

    "github.com/gin-gonic/gin"
    "email_server/database"
    "email_server/models"
    "email_server/utils"
)

// GetAllUsers 获取所有用户列表（管理员功能）
func GetAllUsers(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    status := c.Query("status")
    keyword := c.Query("keyword")

    offset := (page - 1) * pageSize

    query := `
        SELECT id, username, email, real_name, phone, role, status, 
               last_login, created_at, updated_at
        FROM users WHERE 1=1
    `
    countQuery := "SELECT COUNT(*) FROM users WHERE 1=1"
    args := []interface{}{}

    if status != "" {
        query += " AND status = ?"
        countQuery += " AND status = ?"
        args = append(args, status)
    }

    if keyword != "" {
        query += " AND (username LIKE ? OR email LIKE ? OR real_name LIKE ?)"
        countQuery += " AND (username LIKE ? OR email LIKE ? OR real_name LIKE ?)"
        likeKeyword := "%" + keyword + "%"
        args = append(args, likeKeyword, likeKeyword, likeKeyword)
    }

    // 获取总数
    var total int
    err := database.DB.QueryRow(countQuery, args...).Scan(&total)
    if err != nil {
        utils.SendError(c, 500, "查询用户总数失败")
        return
    }

    // 分页查询
    query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
    pageArgs := append(args, pageSize, offset)

    rows, err := database.DB.Query(query, pageArgs...)
    if err != nil {
        utils.SendError(c, 500, "查询用户列表失败")
        return
    }
    defer rows.Close()

    var users []*models.UserResponse
    for rows.Next() {
        user := &models.User{}
        var phone sql.NullString
        var lastLogin sql.NullTime
        
        err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.RealName,
            &phone, &user.Role, &user.Status, &lastLogin,
            &user.CreatedAt, &user.UpdatedAt)
        if err == nil {
            // 处理NULL值
            if phone.Valid {
                user.Phone = phone.String
            } else {
                user.Phone = ""
            }
            
            if lastLogin.Valid {
                user.LastLogin = &lastLogin.Time
            } else {
                user.LastLogin = nil
            }
            
            users = append(users, user.ToResponse())
        }
    }

    result := map[string]interface{}{
        "users":     users,
        "total":     total,
        "page":      page,
        "page_size": pageSize,
    }

    utils.Success(c, result)
}

// UpdateUserStatus 更新用户状态（管理员功能）
func UpdateUserStatus(c *gin.Context) {
    userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        utils.SendError(c, 400, "用户ID错误")
        return
    }

    var req struct {
        Status int `json:"status" binding:"required,oneof=0 1"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        utils.SendError(c, 400, "参数错误")
        return
    }

    _, err = database.DB.Exec("UPDATE users SET status = ?, updated_at = NOW() WHERE id = ?", 
        req.Status, userID)
    if err != nil {
        utils.SendError(c, 500, "更新用户状态失败")
        return
    }

    statusText := "启用"
    if req.Status == 0 {
        statusText = "禁用"
    }

    utils.Success(c, "用户已"+statusText)
}
