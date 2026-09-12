package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/service"
	"github.com/communitygarden/server/internal/util"
)

// StatsHandler 仪表盘统计接口。
type StatsHandler struct {
	statsService *service.StatsService
}

// NewStatsHandler 构造统计接口。
func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

// Overview 平台概览统计。
func (h *StatsHandler) Overview(c *gin.Context) {
	stats, err := h.statsService.Overview()
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, stats)
}
