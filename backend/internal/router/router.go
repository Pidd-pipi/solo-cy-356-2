package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/communitygarden/server/internal/config"
	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/handler"
	"github.com/communitygarden/server/internal/middleware"
	"github.com/communitygarden/server/internal/ws"
)

// Router 路由装配器。
type Router struct {
	cfg   *config.Config
	logger *slog.Logger
	rdb   *redis.Client

	authHandler     *handler.AuthHandler
	userHandler     *handler.UserHandler
	plotHandler     *handler.PlotHandler
	planHandler     *handler.PlantingPlanHandler
	harvestHandler  *handler.HarvestHandler
	diaryHandler    *handler.DiaryHandler
	communityHandler *handler.CommunityHandler
	auditHandler    *handler.AuditHandler
	statsHandler    *handler.StatsHandler

	auditService middleware.AuditWriter
	hub          *ws.Hub
}

// New 构造 Router。
func New(
	cfg *config.Config,
	logger *slog.Logger,
	rdb *redis.Client,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	plotHandler *handler.PlotHandler,
	planHandler *handler.PlantingPlanHandler,
	harvestHandler *handler.HarvestHandler,
	diaryHandler *handler.DiaryHandler,
	communityHandler *handler.CommunityHandler,
	auditHandler *handler.AuditHandler,
	statsHandler *handler.StatsHandler,
	auditService middleware.AuditWriter,
	hub *ws.Hub,
) *Router {
	return &Router{
		cfg: cfg, logger: logger, rdb: rdb,
		authHandler: authHandler, userHandler: userHandler, plotHandler: plotHandler,
		planHandler: planHandler, harvestHandler: harvestHandler, diaryHandler: diaryHandler,
		communityHandler: communityHandler, auditHandler: auditHandler, statsHandler: statsHandler,
		auditService: auditService, hub: hub,
	}
}

// Build 构建 Gin 引擎并注册全部路由。
func (r *Router) Build() *gin.Engine {
	if r.cfg.RunMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	// 全局中间件链
	engine.Use(middleware.RequestID())
	engine.Use(middleware.ErrorHandler(r.logger))
	engine.Use(middleware.RequestLog(r.logger))
	engine.Use(middleware.CORS())
	engine.Use(middleware.RateLimit(r.rdb, r.logger, 300, time.Minute))
	engine.Use(middleware.Audit(r.auditService, r.logger))

	// 健康检查
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": constants.CodeOK, "message": constants.MsgHealthOK, "data": gin.H{"status": "up", "time": time.Now().Format(time.RFC3339)}})
	})

	v1 := engine.Group("/api/v1")
	r.registerAuth(v1)
	r.registerUsers(v1)
	r.registerPlots(v1)
	r.registerPlantingPlans(v1)
	r.registerHarvests(v1)
	r.registerDiaries(v1)
	r.registerCommunity(v1)
	r.registerAudit(v1)
	r.registerStats(v1)
	r.registerWS(v1)
	return engine
}
