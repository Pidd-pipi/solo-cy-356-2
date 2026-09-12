package router

import (
	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/middleware"
)

// registerUsers 用户路由（RBAC）。
func (r *Router) registerUsers(g *gin.RouterGroup) {
	me := g.Group("/me")
	me.Use(middleware.Auth(r.cfg, r.logger))
	{
		me.GET("", r.userHandler.GetMe)
		me.PUT("", r.userHandler.UpdateProfile)
	}

	users := g.Group("/users")
	users.Use(middleware.Auth(r.cfg, r.logger))
	{
		users.GET("", middleware.RequireRoles(string(constants.RoleAdmin)), r.userHandler.ListUsers)
		users.GET("/:id", r.userHandler.GetUser)
		users.PUT("/:id/role", middleware.RequireRoles(string(constants.RoleAdmin)), r.userHandler.ChangeRole)
		users.PUT("/:id/status", middleware.RequireRoles(string(constants.RoleAdmin)), r.userHandler.ChangeStatus)
	}
}
