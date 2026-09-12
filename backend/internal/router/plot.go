package router

import (
	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/middleware"
)

// registerPlots 地块路由。
func (r *Router) registerPlots(g *gin.RouterGroup) {
	plots := g.Group("/plots")
	{
		plots.GET("", r.plotHandler.List)
		plots.GET("/:id", r.plotHandler.Get)
	}

	admin := plots.Group("")
	admin.Use(middleware.Auth(r.cfg, r.logger), middleware.RequireRoles(string(constants.RoleAdmin)))
	{
		admin.POST("", r.plotHandler.Create)
		admin.PUT("/:id", r.plotHandler.Update)
	}

	auth := plots.Group("")
	auth.Use(middleware.Auth(r.cfg, r.logger))
	{
		auth.POST("/:id/adopt", r.plotHandler.Adopt)
		auth.POST("/:id/release", r.plotHandler.Release)
	}
}
