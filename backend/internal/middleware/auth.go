package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/config"
	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/util"
)

// Auth JWT 认证中间件：解析 Bearer Token 并注入用户声明。
func Auth(cfg *config.Config, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.MsgLoginRequired)
			c.Abort()
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := util.ParseToken(cfg.JWTSecret, token)
		if err != nil {
			logger.Warn(constants.LogAuthTokenInvalid, "request_id", util.GetRequestID(c), "err", err)
			util.Fail(c, http.StatusUnauthorized, constants.CodeUnauthorized, constants.ErrorText[constants.CodeUnauthorized])
			c.Abort()
			return
		}
		c.Set(util.CtxClaims, claims)
		c.Next()
	}
}
