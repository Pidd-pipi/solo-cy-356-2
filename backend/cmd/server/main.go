package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/communitygarden/server/internal/config"
	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/database"
	"github.com/communitygarden/server/internal/handler"
	"github.com/communitygarden/server/internal/repository"
	"github.com/communitygarden/server/internal/router"
	"github.com/communitygarden/server/internal/service"
	"github.com/communitygarden/server/internal/util"
	"github.com/communitygarden/server/internal/ws"
)

func main() {
	logger := util.NewLogger("info")
	cfg, err := config.Load()
	if err != nil {
		logger.Error(constants.LogServerShutdown, "err", fmt.Errorf("load config: %w", err))
		os.Exit(1)
	}
	logger = util.NewLogger(cfg.LogLevel)

	db, err := database.Connect(cfg, logger)
	if err != nil {
		logger.Error(constants.LogServerShutdown, "err", fmt.Errorf("connect database: %w", err))
		os.Exit(1)
	}

	// Redis（限流中间件；不可用则降级放行）
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		logger.Warn(constants.LogRedisUnavailable, "err", err)
	} else {
		logger.Info(constants.LogRedisConnected, "host", cfg.RedisHost, "port", cfg.RedisPort)
	}
	cancel()

	// 仓储
	userRepo := repository.NewUserRepository(db)
	plotRepo := repository.NewPlotRepository(db)
	planRepo := repository.NewPlantingPlanRepository(db)
	harvestRepo := repository.NewHarvestRecordRepository(db)
	diaryRepo := repository.NewDiaryRepository(db)
	postRepo := repository.NewCommunityRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// 服务
	authService := service.NewAuthService(userRepo, logger, cfg.JWTSecret, cfg.JWTExpireHours)
	userService := service.NewUserService(userRepo, logger)
	plotService := service.NewPlotService(plotRepo, db, logger)
	auditService := service.NewAuditService(auditRepo, logger)
	planService := service.NewPlantingPlanService(planRepo, plotRepo, plotService, db, logger)
	harvestService := service.NewHarvestRecordService(harvestRepo, planRepo, db, logger)
	diaryService := service.NewDiaryService(diaryRepo, planRepo, logger)
	communityService := service.NewCommunityService(postRepo, logger)
	statsService := service.NewStatsService(userRepo, plotRepo, planRepo, harvestRepo, diaryRepo, postRepo, logger)

	// 处理器
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService, auditService)
	plotHandler := handler.NewPlotHandler(plotService, auditService)
	planHandler := handler.NewPlantingPlanHandler(planService, harvestService)
	harvestHandler := handler.NewHarvestHandler(harvestService, auditService)
	diaryHandler := handler.NewDiaryHandler(diaryService)
	communityHandler := handler.NewCommunityHandler(communityService)
	auditHandler := handler.NewAuditHandler(auditService)
	statsHandler := handler.NewStatsHandler(statsService)

	hub := ws.NewHub(logger, cfg.JWTSecret)

	appRouter := router.New(
		cfg, logger, rdb,
		authHandler, userHandler, plotHandler, planHandler, harvestHandler,
		diaryHandler, communityHandler, auditHandler, statsHandler,
		auditService, hub,
	)
	engine := appRouter.Build()

	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info(constants.LogServerStarted, "port", cfg.ServerPort, "env", cfg.RunMode)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(constants.LogServerShutdown, "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error(constants.LogServerShutdown, "err", err)
	}
	logger.Info(constants.LogServerShutdown, "err", nil)
}
