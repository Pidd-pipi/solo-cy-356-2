package model

import "time"

// AuditLog 操作审计日志实体。
type AuditLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index" json:"user_id"`
	Username     string    `gorm:"size:64;index" json:"username"`
	Role         string    `gorm:"size:32" json:"role"`
	Action       string    `gorm:"size:64;index;not null" json:"action"`
	ResourceType string    `gorm:"size:64" json:"resource_type"`
	ResourceID   string    `gorm:"size:64" json:"resource_id"`
	Detail       string    `gorm:"size:1000" json:"detail"`
	IP           string    `gorm:"size:64" json:"ip"`
	RequestID    string    `gorm:"size:64;index" json:"request_id"`
	CreatedAt    time.Time `json:"created_at"`
}
