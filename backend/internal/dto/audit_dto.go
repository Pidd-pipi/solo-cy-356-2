package dto

import (
	"github.com/communitygarden/server/internal/model"
)

// AuditLogOutDTO 审计日志输出。
type AuditLogOutDTO struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Detail       string `json:"detail"`
	IP           string `json:"ip"`
	RequestID    string `json:"request_id"`
	CreatedAt    string `json:"created_at"`
}

// ToAuditLogOutDTO 模型转 DTO。
func ToAuditLogOutDTO(a *model.AuditLog) *AuditLogOutDTO {
	return &AuditLogOutDTO{
		ID:           a.ID,
		UserID:       a.UserID,
		Username:     a.Username,
		Role:         a.Role,
		Action:       a.Action,
		ResourceType: a.ResourceType,
		ResourceID:   a.ResourceID,
		Detail:       a.Detail,
		IP:           a.IP,
		RequestID:    a.RequestID,
		CreatedAt:    a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
