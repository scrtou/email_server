package utils

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "email_server/models"
)

func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, models.Response{
        Code:    200,
        Message: "success",
        Data:    data,
    })
}

func SendError(c *gin.Context, code int, message string) {
    c.JSON(code, models.Response{
        Code:    code,
        Message: message,
    })
}
