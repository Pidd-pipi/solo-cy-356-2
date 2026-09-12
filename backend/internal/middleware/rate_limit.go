package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/util"
)

// RateLimit Redis 固定窗口限流中间件；Redis 不可用时降级放行。
func RateLimit(rdb *redis.Client, logger *slog.Logger, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			c.Next()
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()
		key := fmt.Sprintf("ratelimit:%s:%s", c.ClientIP(), c.FullPath())
		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		if _, err := pipe.Exec(ctx); err != nil {
			logger.Warn(constants.LogRedisUnavailable, "err", err)
			c.Next()
			return
		}
		if incr.Val() > int64(limit) {
			logger.Warn(constants.LogRateLimitReached, "ip", c.ClientIP(), "path", c.FullPath(), "window", window.String())
			util.Fail(c, http.StatusTooManyRequests, constants.CodeRateLimited, constants.ErrorText[constants.CodeRateLimited])
			c.Abort()
			return
		}
		c.Next()
	}
}
