package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/repository"
	"github.com/communitygarden/server/internal/util"
)

// AuthService 认证服务。
type AuthService struct {
	userRepo repository.UserRepository
	logger   *slog.Logger
	jwtSecret string
	expireHours int
}

// NewAuthService 构造认证服务。
func NewAuthService(userRepo repository.UserRepository, logger *slog.Logger, jwtSecret string, expireHours int) *AuthService {
	return &AuthService{userRepo: userRepo, logger: logger, jwtSecret: jwtSecret, expireHours: expireHours}
}

// Register 注册用户（默认角色 citizen）。
func (s *AuthService) Register(req *dto.RegisterRequest) (*model.User, error) {
	if _, err := s.userRepo.FindByUsername(req.Username); err == nil {
		msg := fmt.Sprintf("%s：用户名 %s 已被占用", constants.ErrorText[constants.CodeDuplicateUsername], req.Username)
		s.logger.Warn(constants.LogUserRegistered, "username", req.Username, "err", "duplicate")
		return nil, util.NewAppError(constants.CodeDuplicateUsername, 409, msg)
	}
	hash, err := util.HashPassword(req.Password)
	if err != nil {
		s.logger.Error(constants.LogInternalError, "err", err)
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError])
	}
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}
	u := &model.User{
		Username: req.Username,
		Password: hash,
		Nickname: nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Role:     string(constants.RoleCitizen),
		Status:   string(constants.UserStatusActive),
	}
	if err := s.userRepo.Create(u); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogUserRegistered, "user_id", u.ID, "username", u.Username, "role", u.Role)
	return u, nil
}

// Login 登录并签发 JWT。
func (s *AuthService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	u, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeInvalidCredentials, 401, constants.ErrorText[constants.CodeInvalidCredentials])
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if !util.CheckPassword(u.Password, req.Password) {
		s.logger.Warn(constants.LogAuthTokenInvalid, "username", req.Username, "err", "bad password")
		return nil, util.NewAppError(constants.CodeInvalidCredentials, 401, constants.ErrorText[constants.CodeInvalidCredentials])
	}
	if u.Status != string(constants.UserStatusActive) {
		return nil, util.NewAppError(constants.CodeUserDisabled, 403, fmt.Sprintf("%s：用户名 %s 角色 %s", constants.ErrorText[constants.CodeUserDisabled], u.Username, util.RoleText(u.Role)))
	}
	token, err := util.GenerateToken(s.jwtSecret, s.expireHours, u.ID, u.Username, u.Role)
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogAuthTokenIssued, "user_id", u.ID, "username", u.Username, "role", u.Role)
	return &dto.LoginResponse{Token: token, User: dto.ToUserOutDTO(u), ExpiresIn: s.expireHours * 3600}, nil
}
