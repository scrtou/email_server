package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"email_server/database"
	"email_server/models"
	"email_server/utils"
)

// GetAllUsers 获取所有用户列表（管理员功能）
func GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	role := c.Query("role")
	keyword := c.Query("keyword")

	offset := (page - 1) * pageSize

	dbQuery := database.DB.Model(&models.User{})

	// Build base query for filtering
	if status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			dbQuery = dbQuery.Where("status = ?", statusInt)
		}
	}

	if role != "" {
		dbQuery = dbQuery.Where("role = ?", role)
	}

	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		dbQuery = dbQuery.Where("username LIKE ? OR email LIKE ?", likeKeyword, likeKeyword)
	}

	// 获取总数
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		utils.SendErrorResponse(c, 500, "查询用户总数失败: "+err.Error())
		return
	}

	// 分页查询
	var usersDB []*models.User
	if err := dbQuery.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&usersDB).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			usersDB = []*models.User{} // Return empty list if no records found
		} else {
			utils.SendErrorResponse(c, 500, "查询用户列表失败: "+err.Error())
			return
		}
	}

	userResponses := make([]*models.UserResponse, len(usersDB))
	for i, u := range usersDB {
		userResponses[i] = u.ToResponse()
	}

	result := map[string]interface{}{
		"users":     userResponses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}

	utils.SendSuccessResponse(c, result)
}

// UpdateUserStatus 更新用户状态（管理员功能）
func UpdateUserStatus(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, 400, "无效的用户ID")
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, 400, "参数错误")
		return
	}

	// 验证状态值
	if req.Status != models.StatusActive && req.Status != models.StatusBanned {
		utils.SendErrorResponse(c, 400, "无效的状态值")
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := database.DB.First(&user, uint(userID)).Error; err != nil {
		if err.Error() == "record not found" {
			utils.SendErrorResponse(c, 404, "用户不存在")
		} else {
			utils.SendErrorResponse(c, 500, "系统错误")
		}
		return
	}

	// 不能修改自己的状态
	currentUserID, _ := c.Get("user_id")
	if currentUserID.(int64) == int64(userID) {
		utils.SendErrorResponse(c, 400, "不能修改自己的状态")
		return
	}

	// 更新用户状态
	if err := database.DB.Model(&user).Update("status", req.Status).Error; err != nil {
		utils.SendErrorResponse(c, 500, "更新失败")
		return
	}

	statusText := "激活"
	if req.Status == models.StatusBanned {
		statusText = "封禁"
	}

	utils.SendSuccessResponse(c, gin.H{
		"message":     "用户状态更新成功",
		"status":      req.Status,
		"status_text": statusText,
	})
}

// UpdateUserRole 更新用户角色（管理员功能）
func UpdateUserRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, 400, "无效的用户ID")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, 400, "参数错误")
		return
	}

	// 验证角色值
	if req.Role != models.RoleAdmin && req.Role != models.RoleUser {
		utils.SendErrorResponse(c, 400, "无效的角色值")
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := database.DB.First(&user, uint(userID)).Error; err != nil {
		if err.Error() == "record not found" {
			utils.SendErrorResponse(c, 404, "用户不存在")
		} else {
			utils.SendErrorResponse(c, 500, "系统错误")
		}
		return
	}

	// 不能修改自己的角色
	currentUserID, _ := c.Get("user_id")
	if currentUserID.(int64) == int64(userID) {
		utils.SendErrorResponse(c, 400, "不能修改自己的角色")
		return
	}

	// 更新用户角色
	if err := database.DB.Model(&user).Update("role", req.Role).Error; err != nil {
		utils.SendErrorResponse(c, 500, "更新失败")
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"message": "用户角色更新成功",
		"role":    req.Role,
	})
}
