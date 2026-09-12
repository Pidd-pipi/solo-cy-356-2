package router

import "github.com/gin-gonic/gin"

// registerWS 实时社区 WebSocket 路由（token 由 Hub 内通过 ?token= 校验）。
func (r *Router) registerWS(g *gin.RouterGroup) {
	g.GET("/ws/community", r.hub.Handle)
}
