package dto

import (
	"time"

	"github.com/communitygarden/server/internal/model"
)

// CreatePlanRequest 创建种植计划。
type CreatePlanRequest struct {
	PlotID   uint   `json:"plot_id" binding:"required,gt=0"`
	CropName string `json:"crop_name" binding:"required,max=64"`
	CropType string `json:"crop_type" binding:"required,oneof=vegetable fruit herb"`
	Season   string `json:"season" binding:"required,oneof=spring summer autumn winter"`
	PlantDate *string `json:"plant_date" binding:"omitempty"`
	Notes    string `json:"notes" binding:"omitempty,max=512"`
}

// UpdatePlanRequest 更新种植计划。
type UpdatePlanRequest struct {
	CropName *string `json:"crop_name" binding:"omitempty,max=64"`
	CropType *string `json:"crop_type" binding:"omitempty,oneof=vegetable fruit herb"`
	Season   *string `json:"season" binding:"omitempty,oneof=spring summer autumn winter"`
	Notes    *string `json:"notes" binding:"omitempty,max=512"`
}

// ChangePlanStatusRequest 状态流转。
type ChangePlanStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=planned planting growing harvesting completed"`
}

// PlanOutDTO 种植计划输出。
type PlanOutDTO struct {
	ID                  uint        `json:"id"`
	PlotID              uint        `json:"plot_id"`
	PlotCode            string      `json:"plot_code"`
	PlotName            string      `json:"plot_name"`
	UserID              uint        `json:"user_id"`
	Username            string      `json:"username"`
	CropName            string      `json:"crop_name"`
	CropType            string      `json:"crop_type"`
	Season              string      `json:"season"`
	Status              string      `json:"status"`
	PlantDate           *string     `json:"plant_date"`
	ExpectedHarvestDate *string     `json:"expected_harvest_date"`
	Notes               string      `json:"notes"`
	CreatedAt           string      `json:"created_at"`
}

// ToPlanOutDTO 模型转 DTO。
func ToPlanOutDTO(p *model.PlantingPlan) *PlanOutDTO {
	dto := &PlanOutDTO{
		ID:        p.ID,
		PlotID:    p.PlotID,
		UserID:    p.UserID,
		CropName:  p.CropName,
		CropType:  p.CropType,
		Season:    p.Season,
		Status:    p.Status,
		Notes:     p.Notes,
		CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if p.Plot != nil {
		dto.PlotCode = p.Plot.Code
		dto.PlotName = p.Plot.Name
	}
	if p.User != nil {
		dto.Username = p.User.Username
	}
	if p.PlantDate != nil {
		s := p.PlantDate.Format("2006-01-02")
		dto.PlantDate = &s
	}
	if p.ExpectedHarvestDate != nil {
		s := p.ExpectedHarvestDate.Format("2006-01-02")
		dto.ExpectedHarvestDate = &s
	}
	return dto
}

// RecommendationOutDTO 季节作物推荐输出。
type RecommendationOutDTO struct {
	Season   string   `json:"season"`
	Crops    []string `json:"crops"`
	HarvestInDays int  `json:"harvest_in_days"`
}

// AnnualStatsOutDTO 年度收成统计输出。
type AnnualStatsOutDTO struct {
	Year          int              `json:"year"`
	TotalWeightKg float64          `json:"total_weight_kg"`
	HarvestCount  int              `json:"harvest_count"`
	ByCropType    map[string]float64 `json:"by_crop_type"`
	ByQuality     map[string]int   `json:"by_quality"`
}

// ParseDate 解析 yyyy-MM-dd 日期字符串。
func ParseDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
