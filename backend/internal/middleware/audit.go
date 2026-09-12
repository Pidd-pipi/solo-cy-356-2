package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/util"
)

// AuditWriter 审计写入抽象（由 AuditService 实现）。
type AuditWriter interface {
	Write(userID uint, username, role, action, resourceType, resourceID, detail, ip, requestID string) error
}

// Audit 操作审计中间件：为写操作（POST/PUT/PATCH/DELETE）记录审计日志。
func Audit(writer AuditWriter, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
			start := time.Now()
			blw := &bodyLogWriter{ResponseWriter: c.Writer, status: 200}
			c.Writer = blw
			c.Next()
			claims, ok := util.GetClaims(c)
			var userID uint
			var username, role string
			if ok {
				userID = claims.UserID
				username = claims.Username
				role = claims.Role
			}
			if blw.status >= 200 && blw.status < 400 {
				action := method + " " + c.FullPath()
				resourceType := deriveResourceType(c.Request.URL.Path)
				if err := writer.Write(userID, username, role, action, resourceType, "", "", c.ClientIP(), util.GetRequestID(c)); err != nil {
					logger.Warn(constants.LogInternalError, "err", err)
				}
			}
			_ = start
			return
		}
		c.Next()
	}
}

// deriveResourceType 从路径提取资源类型（/api/v1/<resource>/...）。
func deriveResourceType(path string) string {
	const prefix = "/api/v1/"
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		rest := path[len(prefix):]
		for i := 0; i < len(rest); i++ {
			if rest[i] == '/' {
				return rest[:i]
			}
		}
		return rest
	}
	return path
}

type bodyLogWriter struct {
	gin.ResponseWriter
	status int
}

func (w *bodyLogWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
