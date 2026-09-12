package router

import (
	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/middleware"
)

// registerPlantingPlans 种植计划路由。
func (r *Router) registerPlantingPlans(g *gin.RouterGroup) {
	plans := g.Group("/planting-plans")
	plans.Use(middleware.Auth(r.cfg, r.logger))
	{
		plans.GET("", r.planHandler.List)
		plans.POST("", r.planHandler.Create)
		plans.GET("/:id", r.planHandler.Get)
		plans.PUT("/:id", r.planHandler.Update)
		plans.POST("/:id/status", r.planHandler.ChangeStatus)
	}

	// 独立前缀避免与 :id 通配冲突
	crops := g.Group("/crops")
	crops.Use(middleware.Auth(r.cfg, r.logger))
	{
		crops.GET("/recommendations", r.planHandler.Recommendations)
	}

	reminders := g.Group("/reminders")
	reminders.Use(middleware.Auth(r.cfg, r.logger))
	{
		reminders.GET("/harvest", r.planHandler.Reminders)
	}
}
