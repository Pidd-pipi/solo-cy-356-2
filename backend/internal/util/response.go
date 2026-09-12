package util

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
)

// Response 统一响应结构 { code, message, data }。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// OK 成功响应。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: constants.CodeOK, Message: constants.MsgSuccess, Data: data})
}

// Fail 失败响应（code + message）。
func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, Response{Code: code, Message: message, Data: nil})
}

// FailWithAppError 根据 AppError 输出失败响应。
func FailWithAppError(c *gin.Context, err error) {
	if ae, ok := err.(*AppError); ok {
		c.JSON(ae.HTTPStatus, Response{Code: ae.Code, Message: ae.Message, Data: nil})
		return
	}
	c.JSON(http.StatusInternalServerError, Response{Code: constants.CodeInternalError, Message: constants.ErrorText[constants.CodeInternalError], Data: nil})
}
