package dto

import (
	"github.com/communitygarden/server/internal/model"
)

// CreateApplicationRequest 提交认养申请（登录用户）。
type CreateApplicationRequest struct {
	PlotID uint   `json:"plot_id" binding:"required,gt=0"`
	Reason string `json:"reason" binding:"omitempty,max=512"`
}

// ReviewApplicationRequest 审核认养申请（管理员）。
type ReviewApplicationRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Note   string `json:"note" binding:"omitempty,max=512"`
}

// ApplicationOutDTO 认养申请输出。
type ApplicationOutDTO struct {
	ID         uint        `json:"id"`
	PlotID     uint        `json:"plot_id"`
	Plot       *PlotOutDTO `json:"plot"`
	UserID     uint        `json:"user_id"`
	User       *UserOutDTO `json:"user"`
	Status     string      `json:"status"`
	Reason     string      `json:"reason"`
	ReviewNote string      `json:"review_note"`
	ReviewedBy *uint       `json:"reviewed_by"`
	ReviewedAt string      `json:"reviewed_at"`
	CreatedAt  string      `json:"created_at"`
}

// ToApplicationOutDTO 模型转 DTO。
func ToApplicationOutDTO(a *model.AdoptionApplication) *ApplicationOutDTO {
	out := &ApplicationOutDTO{
		ID:         a.ID,
		PlotID:     a.PlotID,
		UserID:     a.UserID,
		Status:     a.Status,
		Reason:     a.Reason,
		ReviewNote: a.ReviewNote,
		ReviewedBy: a.ReviewedBy,
		CreatedAt:  a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if a.ReviewedAt != nil {
		out.ReviewedAt = a.ReviewedAt.Format("2006-01-02 15:04:05")
	}
	if a.Plot != nil {
		out.Plot = ToPlotOutDTO(a.Plot)
	}
	if a.User != nil {
		out.User = ToUserOutDTO(a.User)
	}
	return out
}
