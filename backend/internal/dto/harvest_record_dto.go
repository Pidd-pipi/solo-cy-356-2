package dto

import (
	"time"

	"github.com/communitygarden/server/internal/model"
)

// CreateHarvestRequest 记录收成。
type CreateHarvestRequest struct {
	PlanID      uint    `json:"plan_id" binding:"required,gt=0"`
	CropName    string  `json:"crop_name" binding:"required,max=64"`
	HarvestDate string  `json:"harvest_date" binding:"required"`
	WeightKg    float64 `json:"weight_kg" binding:"required,gt=0"`
	Quality     string  `json:"quality" binding:"required,oneof=excellent good fair"`
	Notes       string  `json:"notes" binding:"omitempty,max=512"`
}

// UpdateHarvestRequest 更新收成记录。
type UpdateHarvestRequest struct {
	CropName    *string  `json:"crop_name" binding:"omitempty,max=64"`
	HarvestDate *string  `json:"harvest_date" binding:"omitempty"`
	WeightKg    *float64 `json:"weight_kg" binding:"omitempty,gt=0"`
	Quality     *string  `json:"quality" binding:"omitempty,oneof=excellent good fair"`
	Notes       *string  `json:"notes" binding:"omitempty,max=512"`
}

// HarvestOutDTO 收成记录输出。
type HarvestOutDTO struct {
	ID          uint    `json:"id"`
	PlanID      uint    `json:"plan_id"`
	PlanCode    string  `json:"plan_code"`
	CropName    string  `json:"crop_name"`
	UserID      uint    `json:"user_id"`
	Username    string  `json:"username"`
	HarvestDate string  `json:"harvest_date"`
	WeightKg    float64 `json:"weight_kg"`
	Quality     string  `json:"quality"`
	Notes       string  `json:"notes"`
	CreatedAt   string  `json:"created_at"`
}

// ToHarvestOutDTO 模型转 DTO。
func ToHarvestOutDTO(h *model.HarvestRecord) *HarvestOutDTO {
	dto := &HarvestOutDTO{
		ID:          h.ID,
		PlanID:      h.PlanID,
		CropName:    h.CropName,
		UserID:      h.UserID,
		HarvestDate: h.HarvestDate.Format("2006-01-02"),
		WeightKg:    h.WeightKg,
		Quality:     h.Quality,
		Notes:       h.Notes,
		CreatedAt:   h.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if h.Plan != nil {
		dto.PlanCode = h.Plan.CropName
	}
	if h.User != nil {
		dto.Username = h.User.Username
	}
	return dto
}

// ParseHarvestDate 解析收成日期。
func ParseHarvestDate(s string) (*time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
