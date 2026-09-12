package model

import "time"

// PlantingPlan 种植计划实体（状态机：planned -> planting -> growing -> harvesting -> completed）。
type PlantingPlan struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	PlotID             uint       `gorm:"index;not null" json:"plot_id"`
	Plot               *Plot      `gorm:"foreignKey:PlotID" json:"plot"`
	UserID             uint       `gorm:"index;not null" json:"user_id"`
	User               *User      `gorm:"foreignKey:UserID" json:"user"`
	CropName           string     `gorm:"size:64;not null" json:"crop_name"`
	CropType           string     `gorm:"size:32;not null" json:"crop_type"`
	Season             string     `gorm:"size:32;not null" json:"season"`
	Status             string     `gorm:"size:32;not null;default:planned;index" json:"status"`
	PlantDate          *time.Time `json:"plant_date"`
	ExpectedHarvestDate *time.Time `json:"expected_harvest_date"`
	Notes              string     `gorm:"size:512" json:"notes"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
