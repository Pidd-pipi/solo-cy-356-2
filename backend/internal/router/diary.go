package router

import (
	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/middleware"
)

// registerDiaries 种植日记路由。
func (r *Router) registerDiaries(g *gin.RouterGroup) {
	diaries := g.Group("/diaries")
	diaries.Use(middleware.Auth(r.cfg, r.logger))
	{
		diaries.GET("", r.diaryHandler.List)
		diaries.POST("", r.diaryHandler.Create)
		diaries.GET("/:id", r.diaryHandler.Get)
		diaries.PUT("/:id", r.diaryHandler.Update)
		diaries.DELETE("/:id", r.diaryHandler.Delete)
		diaries.POST("/:id/like", r.diaryHandler.Like)
		diaries.POST("/:id/comments", r.diaryHandler.Comment)
	}
}
