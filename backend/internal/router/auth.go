package router

import "github.com/gin-gonic/gin"

// registerAuth 认证路由。
func (r *Router) registerAuth(g *gin.RouterGroup) {
	auth := g.Group("/auth")
	{
		auth.POST("/register", r.authHandler.Register)
		auth.POST("/login", r.authHandler.Login)
	}
}
