package router

import (
	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/middleware"
)

// registerAdoptionApplications 认养申请路由。
func (r *Router) registerAdoptionApplications(g *gin.RouterGroup) {
	apps := g.Group("/applications")
	apps.Use(middleware.Auth(r.cfg, r.logger))
	{
		apps.POST("", r.applicationHandler.Apply)
		apps.GET("/mine", r.applicationHandler.ListMine)
		apps.POST("/:id/withdraw", r.applicationHandler.Withdraw)
		apps.GET("", middleware.RequireRoles(string(constants.RoleAdmin)), r.applicationHandler.List)
		apps.POST("/:id/review", middleware.RequireRoles(string(constants.RoleAdmin)), r.applicationHandler.Review)
	}
}
