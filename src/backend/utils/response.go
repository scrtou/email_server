package utils

import (
	"net/http"
	"strconv"
	"strings" // <-- 添加 strings 包导入

	"github.com/gin-gonic/gin"
	"email_server/models"
)

// SendSuccessResponse sends a standard success response.
   func SendSuccessResponse(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, models.Response{
    	Code:    http.StatusOK,
    	Message: "success",
    	Data:    data,
    })
   }

   // SendCreatedResponse sends a standard success response with HTTP 201 status.
   func SendCreatedResponse(c *gin.Context, data interface{}) {
    c.JSON(http.StatusCreated, models.Response{
        Code:    http.StatusCreated, // Use 201 for business code as well
        Message: "created",
        Data:    data,
    })
   }
   
   // SendErrorResponse sends a standard error response.
   func SendErrorResponse(c *gin.Context, statusCode int, message string) {
    c.JSON(statusCode, models.Response{
    	Code:    statusCode,
    	Message: message,
    })
   }
   
   // CreatePaginationMeta creates pagination metadata.
   func CreatePaginationMeta(currentPage, pageSize, totalItems int) map[string]interface{} {
    totalPages := 0
    if totalItems > 0 && pageSize > 0 {
    	totalPages = (totalItems + pageSize - 1) / pageSize
    }
    return map[string]interface{}{
    	"current_page": currentPage,
    	"page_size":    pageSize,
    	"total_items":  totalItems,
    	"total_pages":  totalPages,
    }
   }
   
   // SendSuccessResponseWithMeta sends a success response that includes metadata (e.g., for pagination).
   func SendSuccessResponseWithMeta(c *gin.Context, data interface{}, meta map[string]interface{}) {
    c.JSON(http.StatusOK, models.Response{
    	Code:    http.StatusOK,
    	Message: "success",
    	Data:    data,
    	Meta:    meta,
    })
   }
   
   // IsUniqueConstraintError checks if the error is a unique constraint violation.
   // This is a basic check for SQLite, a more robust solution might involve checking driver-specific error codes.
   func IsUniqueConstraintError(err error) bool {
    if err == nil {
    	return false
    }
    // SQLite unique constraint error message often contains "UNIQUE constraint failed"
    // or "constraint failed: UNIQUE constraint"
    // For other databases, this check would need to be different.
    // Example: For PostgreSQL, you might check for pq.ErrorCode("23505")
    // For MySQL, you might check for *mysql.MySQLError and then errNum == 1062
    
    // A common string for SQLite
    errMsg := err.Error()
    if strings.Contains(errMsg, "UNIQUE constraint failed") || strings.Contains(errMsg, "constraint failed: UNIQUE constraint") {
    	return true
    }
    // GORM might also wrap errors, so a more generic check might be needed if the above fails.
    // Sometimes the error might be like "Error 1062: Duplicate entry '...' for key '...'" for MySQL
    // or "ERROR: duplicate key value violates unique constraint "..." (SQLSTATE 23505)" for PostgreSQL
    return false
   }
   
   // Helper function to check if a string contains a substring (case-insensitive for flexibility if needed, but here case-sensitive)
   // func содержит(s, substr string) bool { // <-- 移除错误的函数
   // 	return strings.Contains(s, substr)
   // }
  
  // ExtractProviderFromEmail extracts the domain part as the provider from an email address.
  func ExtractProviderFromEmail(email string) string {
   parts := strings.Split(email, "@")
   if len(parts) == 2 {
   	return parts[1]
   }
   return "" // Return empty if not a valid email format or no domain part
  }

  // StringToUint converts a string to a uint.
  func StringToUint(s string) (uint, error) {
 val, err := strconv.ParseUint(s, 10, 32)
 if err != nil {
  return 0, err
 }
 return uint(val), nil
  }
