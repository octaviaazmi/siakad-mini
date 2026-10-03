package utils

import "github.com/gin-gonic/gin"

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors,omitempty"`
}

func JSONSuccess(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, SuccessResponse{Success: true, Message: message, Data: data})
}

func JSONSuccessWithMeta(c *gin.Context, status int, message string, data, meta interface{}) {
	c.JSON(status, SuccessResponse{Success: true, Message: message, Data: data, Meta: meta})
}

func JSONError(c *gin.Context, status int, message string, errors interface{}) {
	c.JSON(status, ErrorResponse{Success: false, Message: message, Errors: errors})
}
