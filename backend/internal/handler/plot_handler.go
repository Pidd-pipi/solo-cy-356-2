package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/middleware"
	"github.com/communitygarden/server/internal/service"
	"github.com/communitygarden/server/internal/util"
)

// PlotHandler 地块接口。
type PlotHandler struct {
	plotService *service.PlotService
	audit       middleware.AuditWriter
}

// NewPlotHandler 构造地块接口。
func NewPlotHandler(plotService *service.PlotService, audit middleware.AuditWriter) *PlotHandler {
	return &PlotHandler{plotService: plotService, audit: audit}
}

// Create 创建地块（管理员）。
func (h *PlotHandler) Create(c *gin.Context) {
	var req dto.CreatePlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	p, err := h.plotService.Create(&req, claims.Username)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToPlotOutDTO(p))
}

// Update 更新地块（管理员）。
func (h *PlotHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	var req dto.UpdatePlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	p, err := h.plotService.Update(uint(id), &req, claims.Username)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToPlotOutDTO(p))
}

// List 地块分页列表（公开）。
func (h *PlotHandler) List(c *gin.Context) {
	pq := util.ParsePageQuery(c)
	status := c.Query("status")
	plots, total, err := h.plotService.List(pq, status)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	list := make([]*dto.PlotOutDTO, 0, len(plots))
	for i := range plots {
		list = append(list, dto.ToPlotOutDTO(&plots[i]))
	}
	util.OK(c, util.PageResult{List: list, Total: total, Page: pq.Page, PageSize: pq.PageSize})
}

// Get 地块详情。
func (h *PlotHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	p, err := h.plotService.GetByID(uint(id))
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToPlotOutDTO(p))
}

// Adopt 认养地块（登录用户）。
func (h *PlotHandler) Adopt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	claims, _ := util.GetClaims(c)
	p, err := h.plotService.Adopt(uint(id), claims.UserID, claims.Role, claims.Username)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	_ = h.audit.Write(claims.UserID, claims.Username, claims.Role, "ADOPT_PLOT", "plot", strconv.FormatUint(uint64(id), 10),
		"用户认养地块 "+p.Code, c.ClientIP(), util.GetRequestID(c))
	util.OK(c, dto.ToPlotOutDTO(p))
}

// Release 释放地块（管理员或认养人）。
func (h *PlotHandler) Release(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	claims, _ := util.GetClaims(c)
	p, err := h.plotService.Release(uint(id), claims.UserID, claims.Role)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	_ = h.audit.Write(claims.UserID, claims.Username, claims.Role, "RELEASE_PLOT", "plot", strconv.FormatUint(uint64(id), 10),
		"释放地块 "+p.Code, c.ClientIP(), util.GetRequestID(c))
	util.OK(c, dto.ToPlotOutDTO(p))
}
