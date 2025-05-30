package handlers

import (
	"strconv"
	// "database/sql" // GORM handles this
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"email_server/database"
	"email_server/models"
	"email_server/utils"
)

// GetAllUsers 获取所有用户列表（管理员功能）
func GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	// status := c.Query("status") // Status field removed from User model
	keyword := c.Query("keyword")

	offset := (page - 1) * pageSize

	dbQuery := database.DB.Model(&models.User{})

	// Build base query for filtering
	// if status != "" { // Status field removed from User model
	// 	dbQuery = dbQuery.Where("status = ?", status)
	// }

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
// Note: This function is commented out because the 'Status' field has been removed from the 'User' model.
// If user status management is required, the 'User' model and this function will need to be updated accordingly.
/*
func UpdateUserStatus(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, 400, "用户ID错误")
		return
	}

	var req struct {
		Status int `json:"status" binding:"required,oneof=0 1"` // This Status field no longer exists in models.User
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, 400, "参数错误: "+err.Error())
		return
	}

	// The 'status' column does not exist in the 'users' table anymore.
	// result := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("status", req.Status)
	// if result.Error != nil {
	// 	utils.SendError(c, 500, "更新用户状态失败: "+result.Error.Error())
	// 	return
	// }
	// if result.RowsAffected == 0 {
	// 	utils.SendError(c, 404, "未找到用户或状态未改变")
	// 	return
	// }


	statusText := "启用" // This logic is now obsolete
	if req.Status == 0 {
		statusText = "禁用"
	}

	utils.SendSuccessResponse(c, "用户状态更新操作已停用 (用户模型已简化)") // Placeholder message
}
*/
