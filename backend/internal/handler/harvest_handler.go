package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/middleware"
	"github.com/communitygarden/server/internal/service"
	"github.com/communitygarden/server/internal/util"
)

// HarvestHandler 收成记录接口。
type HarvestHandler struct {
	harvestService *service.HarvestRecordService
	audit          middleware.AuditWriter
}

// NewHarvestHandler 构造收成记录接口。
func NewHarvestHandler(harvestService *service.HarvestRecordService, audit middleware.AuditWriter) *HarvestHandler {
	return &HarvestHandler{harvestService: harvestService, audit: audit}
}

// Create 记录收成。
func (h *HarvestHandler) Create(c *gin.Context) {
	var req dto.CreateHarvestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	record, err := h.harvestService.Create(&req, claims.UserID)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	_ = h.audit.Write(claims.UserID, claims.Username, claims.Role, "RECORD_HARVEST", "harvest", strconv.FormatUint(uint64(record.ID), 10),
		"记录收成 "+record.CropName+" "+util.WeightKgText(record.WeightKg), c.ClientIP(), util.GetRequestID(c))
	util.OK(c, dto.ToHarvestOutDTO(record))
}

// Update 更新收成记录。
func (h *HarvestHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	var req dto.UpdateHarvestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	record, err := h.harvestService.Update(uint(id), claims.UserID, claims.Role, &req)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToHarvestOutDTO(record))
}

// Delete 删除收成记录。
func (h *HarvestHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	claims, _ := util.GetClaims(c)
	if err := h.harvestService.Delete(uint(id), claims.UserID, claims.Role); err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, gin.H{"deleted": true})
}

// List 收成记录分页列表。
func (h *HarvestHandler) List(c *gin.Context) {
	pq := util.ParsePageQuery(c)
	claims, _ := util.GetClaims(c)
	var userID uint
	if claims.Role != string(constants.RoleAdmin) {
		userID = claims.UserID
	}
	records, total, err := h.harvestService.List(pq, userID)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	list := make([]*dto.HarvestOutDTO, 0, len(records))
	for i := range records {
		list = append(list, dto.ToHarvestOutDTO(&records[i]))
	}
	util.OK(c, util.PageResult{List: list, Total: total, Page: pq.Page, PageSize: pq.PageSize})
}

// AnnualStats 年度收成统计（复用 HarvestRecordService.AnnualStats）。
func (h *HarvestHandler) AnnualStats(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	claims, _ := util.GetClaims(c)
	stats, err := h.harvestService.AnnualStats(claims.UserID, year)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, stats)
}
