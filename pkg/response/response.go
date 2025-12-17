package response

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code int         `json:"code" example:"0"`
	Msg  string      `json:"msg" example:"操作成功"`
	Data interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code: 0,
		Msg:  "操作成功",
		Data: data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, msg string) {
	c.JSON(code, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// SuccessWithMsg 成功响应带自定义消息
func SuccessWithMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(200, Response{
		Code: 0,
		Msg:  msg,
		Data: data,
	})
}

// Errorf 格式化错误响应
func Errorf(c *gin.Context, code int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	Error(c, code, msg)
}

// SuccessWithMsgf 格式化成功消息响应
func SuccessWithMsgf(c *gin.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	SuccessWithMsg(c, msg, nil)
}

// SuccessWithDataAndMsgf 格式化成功消息带数据响应
func SuccessWithDataAndMsgf(c *gin.Context, data interface{}, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	SuccessWithMsg(c, msg, data)
}
