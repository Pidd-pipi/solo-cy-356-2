package router

import (
	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/middleware"
)

// registerStats 仪表盘统计路由。
func (r *Router) registerStats(g *gin.RouterGroup) {
	dashboard := g.Group("/dashboard")
	dashboard.Use(middleware.Auth(r.cfg, r.logger))
	{
		dashboard.GET("/stats", r.statsHandler.Overview)
	}
}
