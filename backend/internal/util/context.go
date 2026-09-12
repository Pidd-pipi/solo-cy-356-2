package util

import (
	"github.com/gin-gonic/gin"
)

// 上下文键统一管理。
const (
	CtxRequestID = "ctx_request_id"
	CtxClaims    = "ctx_claims"
)

// GetRequestID 从上下文读取请求 ID。
func GetRequestID(c *gin.Context) string {
	v, ok := c.Get(CtxRequestID)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// GetClaims 从上下文读取 JWT 声明。
func GetClaims(c *gin.Context) (*Claims, bool) {
	v, ok := c.Get(CtxClaims)
	if !ok {
		return nil, false
	}
	claims, ok := v.(*Claims)
	return claims, ok
}

// GetUserID 从上下文读取当前用户 ID。
func GetUserID(c *gin.Context) uint {
	claims, ok := GetClaims(c)
	if !ok {
		return 0
	}
	return claims.UserID
}
