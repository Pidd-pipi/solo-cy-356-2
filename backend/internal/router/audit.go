package router

import (
	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/middleware"
)

// registerAudit 审计日志路由（仅管理员）。
func (r *Router) registerAudit(g *gin.RouterGroup) {
	logs := g.Group("/audit-logs")
	logs.Use(middleware.Auth(r.cfg, r.logger), middleware.RequireRoles(string(constants.RoleAdmin)))
	{
		logs.GET("", r.auditHandler.List)
	}
}
