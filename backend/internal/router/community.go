package router

import (
	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/middleware"
)

// registerCommunity 农友社区路由。
func (r *Router) registerCommunity(g *gin.RouterGroup) {
	posts := g.Group("/community-posts")
	posts.Use(middleware.Auth(r.cfg, r.logger))
	{
		posts.GET("", r.communityHandler.List)
		posts.POST("", r.communityHandler.Create)
		posts.GET("/:id", r.communityHandler.Get)
		posts.PUT("/:id", r.communityHandler.Update)
		posts.DELETE("/:id", r.communityHandler.Remove)
		posts.POST("/:id/like", r.communityHandler.Like)
		posts.POST("/:id/comments", r.communityHandler.Comment)
	}
}
