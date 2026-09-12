package dto

import (
	"github.com/communitygarden/server/internal/model"
)

// UserOutDTO 用户对外输出。
type UserOutDTO struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// UpdateUserRequest 用户资料更新。
type UpdateUserRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,max=64"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,max=32"`
}

// ChangeRoleRequest 角色变更（RBAC，仅管理员）。
type ChangeRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin farmer citizen"`
}

// ChangeStatusRequest 用户状态变更（仅管理员）。
type ChangeStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active disabled"`
}

// ToUserOutDTO 模型转 DTO。
func ToUserOutDTO(u *model.User) *UserOutDTO {
	return &UserOutDTO{
		ID:        u.ID,
		Username:  u.Username,
		Nickname:  u.Nickname,
		Email:     u.Email,
		Phone:     u.Phone,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
