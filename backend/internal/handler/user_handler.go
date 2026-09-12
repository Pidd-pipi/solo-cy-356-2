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

// UserHandler 用户接口。
type UserHandler struct {
	userService *service.UserService
	audit       middleware.AuditWriter
}

// NewUserHandler 构造用户接口。
func NewUserHandler(userService *service.UserService, audit middleware.AuditWriter) *UserHandler {
	return &UserHandler{userService: userService, audit: audit}
}

// GetMe 当前用户信息。
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := util.GetUserID(c)
	u, err := h.userService.GetByID(userID)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToUserOutDTO(u))
}

// GetUser 查询用户详情。
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	u, err := h.userService.GetByID(uint(id))
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToUserOutDTO(u))
}

// ListUsers 用户分页列表（管理员）。
func (h *UserHandler) ListUsers(c *gin.Context) {
	pq := util.ParsePageQuery(c)
	role := c.Query("role")
	status := c.Query("status")
	users, total, err := h.userService.List(pq, role, status)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	list := make([]*dto.UserOutDTO, 0, len(users))
	for i := range users {
		list = append(list, dto.ToUserOutDTO(&users[i]))
	}
	util.OK(c, util.PageResult{List: list, Total: total, Page: pq.Page, PageSize: pq.PageSize})
}

// UpdateProfile 更新当前用户资料。
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	u, err := h.userService.UpdateProfile(util.GetUserID(c), &req)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToUserOutDTO(u))
}

// ChangeRole 变更用户角色（管理员）。
func (h *UserHandler) ChangeRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	var req dto.ChangeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	u, err := h.userService.ChangeRole(claims.UserID, uint(id), claims.Role, req.Role)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	_ = h.audit.Write(claims.UserID, claims.Username, claims.Role, "CHANGE_ROLE", "user", strconv.FormatUint(uint64(id), 10),
		"管理员变更用户角色为 "+util.RoleText(req.Role), c.ClientIP(), util.GetRequestID(c))
	util.OK(c, dto.ToUserOutDTO(u))
}

// ChangeStatus 启用/禁用用户（管理员）。
func (h *UserHandler) ChangeStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "路径参数 id 必须为正整数")
		return
	}
	var req dto.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	claims, _ := util.GetClaims(c)
	u, err := h.userService.ChangeStatus(claims.UserID, uint(id), claims.Role, req.Status)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToUserOutDTO(u))
}
