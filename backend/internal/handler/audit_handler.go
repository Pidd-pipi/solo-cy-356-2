package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/service"
	"github.com/communitygarden/server/internal/util"
)

// AuditHandler 审计日志接口（仅管理员）。
type AuditHandler struct {
	auditService *service.AuditService
}

// NewAuditHandler 构造审计接口。
func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// List 审计日志分页列表。
func (h *AuditHandler) List(c *gin.Context) {
	pq := util.ParsePageQuery(c)
	action := c.Query("action")
	username := c.Query("username")
	logs, total, err := h.auditService.List(pq, action, username)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	list := make([]*dto.AuditLogOutDTO, 0, len(logs))
	for i := range logs {
		list = append(list, dto.ToAuditLogOutDTO(&logs[i]))
	}
	util.OK(c, util.PageResult{List: list, Total: total, Page: pq.Page, PageSize: pq.PageSize})
}
