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

// AdoptionApplicationHandler 认养申请接口。
type AdoptionApplicationHandler struct {
	appService *service.AdoptionApplicationService
	audit      middleware.AuditWriter
}

// NewAdoptionApplicationHandler 构造认养申请接口。
func NewAdoptionApplicationHandler(appService *service.AdoptionApplicationService, audit middleware.AuditWriter) *AdoptionApplicationHandler {
	return &AdoptionApplicationHandler{appService: appService, audit: audit}
}

// Apply 提交认养申请（登录用户）。
func (h *AdoptionApplicationHandler) Apply(c *gin.Context) {
	var req dto.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	app, err := h.appService.Apply(claims.UserID, &req, claims.Username, claims.Role)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	_ = h.audit.Write(claims.UserID, claims.Username, claims.Role, "APPLY_ADOPTION", "adoption_application", strconv.FormatUint(uint64(app.ID), 10),
		"提交地块认养申请 plot_id="+strconv.FormatUint(uint64(app.PlotID), 10), c.ClientIP(), util.GetRequestID(c))
	util.OK(c, dto.ToApplicationOutDTO(app))
}

// ListMine 我的认养申请列表（进度查询）。
func (h *AdoptionApplicationHandler) ListMine(c *gin.Context) {
	pq := util.ParsePageQuery(c)
	claims, _ := util.GetClaims(c)
	apps, total, err := h.appService.ListMine(pq, claims.UserID)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	list := make([]*dto.ApplicationOutDTO, 0, len(apps))
	for i := range apps {
		list = append(list, dto.ToApplicationOutDTO(&apps[i]))
	}
	util.OK(c, util.PageResult{List: list, Total: total, Page: pq.Page, PageSize: pq.PageSize})
}

// List 全部认养申请列表（管理员审核入口，?status= 过滤）。
func (h *AdoptionApplicationHandler) List(c *gin.Context) {
	pq := util.ParsePageQuery(c)
	status := c.Query("status")
	apps, total, err := h.appService.List(pq, status)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	list := make([]*dto.ApplicationOutDTO, 0, len(apps))
	for i := range apps {
		list = append(list, dto.ToApplicationOutDTO(&apps[i]))
	}
	util.OK(c, util.PageResult{List: list, Total: total, Page: pq.Page, PageSize: pq.PageSize})
}

// Withdraw 撤回认养申请（本人或管理员）。
func (h *AdoptionApplicationHandler) Withdraw(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	claims, _ := util.GetClaims(c)
	app, err := h.appService.Withdraw(uint(id), claims.UserID, claims.Role)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	_ = h.audit.Write(claims.UserID, claims.Username, claims.Role, "WITHDRAW_APPLICATION", "adoption_application", strconv.FormatUint(uint64(id), 10),
		"撤回认养申请 plot_id="+strconv.FormatUint(uint64(app.PlotID), 10), c.ClientIP(), util.GetRequestID(c))
	util.OK(c, dto.ToApplicationOutDTO(app))
}

// Review 审核认养申请（管理员，approve/reject）。
func (h *AdoptionApplicationHandler) Review(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	var req dto.ReviewApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	app, err := h.appService.Review(uint(id), claims.UserID, claims.Username, &req)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	_ = h.audit.Write(claims.UserID, claims.Username, claims.Role, "REVIEW_APPLICATION", "adoption_application", strconv.FormatUint(uint64(id), 10),
		"审核认养申请 action="+req.Action, c.ClientIP(), util.GetRequestID(c))
	util.OK(c, dto.ToApplicationOutDTO(app))
}
