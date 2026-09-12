package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/util"
)

// ErrorHandler 全局错误恢复中间件：捕获 panic 并输出统一错误响应。
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				rid := util.GetRequestID(c)
				logger.Error(constants.LogPanicRecovered, "request_id", rid, "err", r)
				util.Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.ErrorText[constants.CodeInternalError])
				c.Abort()
			}
		}()
		c.Next()
	}
}
