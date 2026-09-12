package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/util"
)

// RequestLog 请求日志中间件：统一记录 request_id/method/path/status/latency_ms。
func RequestLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := c.Writer.Status()
		latency := time.Since(start).Milliseconds()
		logger.Info(constants.LogRequestHandled,
			"request_id", util.GetRequestID(c),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency_ms", latency,
		)
	}
}
