package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code int    `json:"code"` // 业务码，0 表示成功
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

// Success 成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "ok", Data: data})
}

// SuccessWithMsg 成功响应（自定义提示信息）
func SuccessWithMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: msg})
}

// Error 失败响应（根据业务码映射 HTTP 状态码）
func Error(c *gin.Context, code int, msg string) {
	c.JSON(httpStatusOf(code), Response{Code: code, Msg: msg})
}

// httpStatusOf 根据业务码映射 HTTP 状态码，让日志能按级别区分
func httpStatusOf(code int) int {
	switch {
	case code == CodeSuccess:
		return http.StatusOK
	case code >= 40100 && code <= 40199:
		return http.StatusUnauthorized
	case code >= 400 && code <= 499:
		return http.StatusBadRequest
	case code >= 500:
		return http.StatusInternalServerError
	default:
		return http.StatusOK
	}
}
