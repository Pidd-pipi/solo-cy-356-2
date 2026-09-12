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

// UserService 用户管理服务。
type UserService struct {
	userRepo repository.UserRepository
	logger   *slog.Logger
}

// NewUserService 构造用户服务。
func NewUserService(userRepo repository.UserRepository, logger *slog.Logger) *UserService {
	return &UserService{userRepo: userRepo, logger: logger}
}

// GetByID 查询用户详情。
func (s *UserService) GetByID(id uint) (*model.User, error) {
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("用户实体 id=%d 不存在", id))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return u, nil
}

// List 分页查询用户。
func (s *UserService) List(pq util.PageQuery, role, status string) ([]model.User, int64, error) {
	users, total, err := s.userRepo.List(pq, role, status)
	if err != nil {
		return nil, 0, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return users, total, nil
}

// UpdateProfile 更新用户资料。
func (s *UserService) UpdateProfile(id uint, req *dto.UpdateUserRequest) (*model.User, error) {
	u, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("用户实体 id=%d 不存在", id))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if req.Nickname != "" {
		u.Nickname = req.Nickname
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.Phone != "" {
		u.Phone = req.Phone
	}
	if err := s.userRepo.Update(u); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return u, nil
}

// ChangeRole 变更用户角色（RBAC：仅管理员）。
func (s *UserService) ChangeRole(operatorID, targetID uint, operatorRole, newRole string) (*model.User, error) {
	if operatorRole != string(constants.RoleAdmin) {
		return nil, util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf(constants.MsgPermissionDenied, util.RoleText(operatorRole), util.RoleText(string(constants.RoleAdmin))))
	}
	u, err := s.userRepo.FindByID(targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("用户实体 id=%d 不存在", targetID))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	old := u.Role
	u.Role = newRole
	if err := s.userRepo.Update(u); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogUserRoleChanged, "user_id", targetID, "operator", operatorID, "from", old, "to", newRole)
	return u, nil
}

// ChangeStatus 启用/禁用用户（仅管理员）。
func (s *UserService) ChangeStatus(operatorID, targetID uint, operatorRole, status string) (*model.User, error) {
	if operatorRole != string(constants.RoleAdmin) {
		return nil, util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf(constants.MsgPermissionDenied, util.RoleText(operatorRole), util.RoleText(string(constants.RoleAdmin))))
	}
	u, err := s.userRepo.FindByID(targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("用户实体 id=%d 不存在", targetID))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if operatorID == targetID && status == string(constants.UserStatusDisabled) {
		return nil, util.NewAppError(constants.CodeConflict, 409, "不能禁用当前登录的管理员账号")
	}
	u.Status = status
	if err := s.userRepo.Update(u); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogUserStatusChanged, "user_id", targetID, "operator", operatorID, "to", status)
	return u, nil
}
