package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/service"
	"github.com/communitygarden/server/internal/util"
)

// PlantingPlanHandler 种植计划接口。
type PlantingPlanHandler struct {
	planService *service.PlantingPlanService
	harvestSvc  *service.HarvestRecordService
}

// NewPlantingPlanHandler 构造种植计划接口。
func NewPlantingPlanHandler(planService *service.PlantingPlanService, harvestSvc *service.HarvestRecordService) *PlantingPlanHandler {
	return &PlantingPlanHandler{planService: planService, harvestSvc: harvestSvc}
}

// Create 创建种植计划。
func (h *PlantingPlanHandler) Create(c *gin.Context) {
	var req dto.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	plan, err := h.planService.Create(&req, util.GetUserID(c))
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToPlanOutDTO(plan))
}

// Update 更新种植计划。
func (h *PlantingPlanHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	var req dto.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	plan, err := h.planService.Update(uint(id), claims.UserID, claims.Role, &req)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToPlanOutDTO(plan))
}

// List 种植计划分页列表。
func (h *PlantingPlanHandler) List(c *gin.Context) {
	pq := util.ParsePageQuery(c)
	claims, _ := util.GetClaims(c)
	var userID uint
	if claims.Role != string(constants.RoleAdmin) {
		userID = claims.UserID
	}
	status := c.Query("status")
	plans, total, err := h.planService.List(pq, userID, status)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	list := make([]*dto.PlanOutDTO, 0, len(plans))
	for i := range plans {
		list = append(list, dto.ToPlanOutDTO(&plans[i]))
	}
	util.OK(c, util.PageResult{List: list, Total: total, Page: pq.Page, PageSize: pq.PageSize})
}

// Get 种植计划详情。
func (h *PlantingPlanHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	plan, err := h.planService.GetByID(uint(id))
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToPlanOutDTO(plan))
}

// ChangeStatus 状态流转。
func (h *PlantingPlanHandler) ChangeStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	var req dto.ChangePlanStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	plan, err := h.planService.ChangeStatus(uint(id), claims.UserID, claims.Role, req.Status)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToPlanOutDTO(plan))
}

// Recommendations 季节作物推荐。
func (h *PlantingPlanHandler) Recommendations(c *gin.Context) {
	season := c.DefaultQuery("season", time.Now().Month().String())
	seasonMap := map[string]string{
		"January": "winter", "February": "winter", "March": "spring", "April": "spring",
		"May": "spring", "June": "summer", "July": "summer", "August": "summer",
		"September": "autumn", "October": "autumn", "November": "autumn", "December": "winter",
	}
	if s, ok := seasonMap[season]; ok {
		season = s
	}
	recs, err := h.planService.Recommendations(season)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, recs)
}

// Reminders 采摘提醒（近 7 天成熟）。
func (h *PlantingPlanHandler) Reminders(c *gin.Context) {
	claims, _ := util.GetClaims(c)
	plans, err := h.planService.Reminders(claims.UserID)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	list := make([]*dto.PlanOutDTO, 0, len(plans))
	for i := range plans {
		list = append(list, dto.ToPlanOutDTO(&plans[i]))
	}
	util.OK(c, list)
}

// AnnualStats 年度收成统计（复用 HarvestRecordService.AnnualStats）。
func (h *PlantingPlanHandler) AnnualStats(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	claims, _ := util.GetClaims(c)
	stats, err := h.harvestSvc.AnnualStats(claims.UserID, year)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, stats)
}
