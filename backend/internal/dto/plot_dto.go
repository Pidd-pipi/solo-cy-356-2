package dto

import (
	"github.com/communitygarden/server/internal/model"
)

// CreatePlotRequest 创建地块（管理员）。
type CreatePlotRequest struct {
	Name        string  `json:"name" binding:"required,max=128"`
	Code        string  `json:"code" binding:"required,max=32"`
	Area        float64 `json:"area" binding:"required,gt=0"`
	SoilType    string  `json:"soil_type" binding:"required,oneof=loam clay sand black"`
	Sunlight    string  `json:"sunlight" binding:"required,oneof=full partial shade"`
	Latitude    float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude   float64 `json:"longitude" binding:"required,min=-180,max=180"`
	Description string  `json:"description" binding:"omitempty,max=512"`
}

// UpdatePlotRequest 更新地块（管理员）。
type UpdatePlotRequest struct {
	Name        *string  `json:"name" binding:"omitempty,max=128"`
	Code        *string  `json:"code" binding:"omitempty,max=32"`
	Area        *float64 `json:"area" binding:"omitempty,gt=0"`
	SoilType    *string  `json:"soil_type" binding:"omitempty,oneof=loam clay sand black"`
	Sunlight    *string  `json:"sunlight" binding:"omitempty,oneof=full partial shade"`
	Latitude    *float64 `json:"latitude" binding:"omitempty,min=-90,max=90"`
	Longitude   *float64 `json:"longitude" binding:"omitempty,min=-180,max=180"`
	Description *string  `json:"description" binding:"omitempty,max=512"`
}

// PlotOutDTO 地块输出。
type PlotOutDTO struct {
	ID          uint        `json:"id"`
	Name        string      `json:"name"`
	Code        string      `json:"code"`
	Area        float64     `json:"area"`
	SoilType    string      `json:"soil_type"`
	Sunlight    string      `json:"sunlight"`
	Latitude    float64     `json:"latitude"`
	Longitude   float64     `json:"longitude"`
	Status      string      `json:"status"`
	AdopterID   *uint       `json:"adopter_id"`
	Adopter     *UserOutDTO `json:"adopter"`
	Description string      `json:"description"`
	CreatedAt   string      `json:"created_at"`
}

// ToPlotOutDTO 模型转 DTO。
func ToPlotOutDTO(p *model.Plot) *PlotOutDTO {
	dto := &PlotOutDTO{
		ID:          p.ID,
		Name:        p.Name,
		Code:        p.Code,
		Area:        p.Area,
		SoilType:    p.SoilType,
		Sunlight:    p.Sunlight,
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
		Status:      p.Status,
		AdopterID:   p.AdopterID,
		Description: p.Description,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if p.Adopter != nil {
		dto.Adopter = ToUserOutDTO(p.Adopter)
	}
	return dto
}
