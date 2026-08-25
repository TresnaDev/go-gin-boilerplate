package response

import "github.com/gin-gonic/gin"

type envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, envelope{Success: true, Message: message, Data: data})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, envelope{Success: false, Message: message})
}
