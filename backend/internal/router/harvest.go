package router

import (
	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/middleware"
)

// registerHarvests 收成记录路由。
func (r *Router) registerHarvests(g *gin.RouterGroup) {
	harvests := g.Group("/harvests")
	harvests.Use(middleware.Auth(r.cfg, r.logger))
	{
		harvests.GET("", r.harvestHandler.List)
		harvests.POST("", r.harvestHandler.Create)
		harvests.PUT("/:id", r.harvestHandler.Update)
		harvests.DELETE("/:id", r.harvestHandler.Delete)
	}

	stats := g.Group("/stats")
	stats.Use(middleware.Auth(r.cfg, r.logger))
	{
		stats.GET("/annual", r.harvestHandler.AnnualStats)
	}
}
