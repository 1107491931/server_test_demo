package utils

import (
	"common/errs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errs.SUCCESS,
		Message: errs.GetMsg(errs.SUCCESS),
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errs.SUCCESS,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应 (基础方法)
func Error(c *gin.Context, httpCode int, errCode int, message string) {
	// 强制返回 HTTP 200 OK，前端根据 code 判断
	c.JSON(http.StatusOK, Response{
		Code:    errCode,
		Message: message,
		Data:    map[string]interface{}{},
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, errs.INVALID_PARAMS, message)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, errs.UNAUTHORIZED, message)
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, errs.NOT_FOUND, message)
}

// InternalServerError 500错误
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, errs.SERVER_ERROR, message)
}

// BusinessError 业务逻辑错误 (返回200状态码，但包含错误码)
func BusinessError(c *gin.Context, errCode int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    errCode,
		Message: message,
		Data:    map[string]interface{}{},
	})
}
