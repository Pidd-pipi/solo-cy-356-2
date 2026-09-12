package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/service"
	"github.com/communitygarden/server/internal/util"
)

// AuthHandler 认证接口。
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 构造认证接口。
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register 注册。
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	u, err := h.authService.Register(&req)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, dto.ToUserOutDTO(u))
}

// Login 登录。
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeValidationFailed, constants.ErrorText[constants.CodeValidationFailed]+": "+err.Error())
		return
	}
	resp, err := h.authService.Login(&req)
	if err != nil {
		util.FailWithAppError(c, err)
		return
	}
	util.OK(c, resp)
}
