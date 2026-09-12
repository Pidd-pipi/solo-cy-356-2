package service

import (
	"log/slog"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/repository"
	"github.com/communitygarden/server/internal/util"
)

// AuditService 审计日志服务。
type AuditService struct {
	auditRepo repository.AuditRepository
	logger    *slog.Logger
}

// NewAuditService 构造审计服务。
func NewAuditService(auditRepo repository.AuditRepository, logger *slog.Logger) *AuditService {
	return &AuditService{auditRepo: auditRepo, logger: logger}
}

// Write 写入审计日志（service 埋点）。
func (s *AuditService) Write(userID uint, username, role, action, resourceType, resourceID, detail, ip, requestID string) error {
	log := &model.AuditLog{
		UserID:       userID,
		Username:     username,
		Role:         role,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Detail:       detail,
		IP:           ip,
		RequestID:    requestID,
	}
	if err := s.auditRepo.Create(log); err != nil {
		s.logger.Error(constants.LogInternalError, "err", err)
		return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogAuditWritten, "audit_id", log.ID, "action", action, "resource", resourceType, "request_id", requestID)
	return nil
}

// List 分页查询审计日志（管理员）。
func (s *AuditService) List(pq util.PageQuery, action, username string) ([]model.AuditLog, int64, error) {
	logs, total, err := s.auditRepo.List(pq, action, username)
	if err != nil {
		return nil, 0, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogAuditListed, "operator", username, "page", pq.Page, "page_size", pq.PageSize)
	return logs, total, nil
}
