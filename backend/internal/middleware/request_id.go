package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/util"
)

const (
	// HeaderRequestID 请求 ID 响应头。
	HeaderRequestID = "X-Request-Id"
)

// generateRequestID 生成 32 位十六进制请求 ID。
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}

// RequestID 请求追踪中间件：为每个请求注入 request_id。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = generateRequestID()
		}
		c.Set(util.CtxRequestID, rid)
		c.Header(HeaderRequestID, rid)
		c.Next()
	}
}
