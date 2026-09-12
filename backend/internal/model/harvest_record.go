package model

import "time"

// HarvestRecord 收成记录实体（采摘重量/品质 + 年度统计）。
type HarvestRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PlanID     uint      `gorm:"index;not null" json:"plan_id"`
	Plan       *PlantingPlan `gorm:"foreignKey:PlanID" json:"plan"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	User       *User     `gorm:"foreignKey:UserID" json:"user"`
	CropName   string    `gorm:"size:64;not null" json:"crop_name"`
	HarvestDate time.Time `gorm:"index;not null" json:"harvest_date"`
	WeightKg   float64   `gorm:"not null" json:"weight_kg"`
	Quality    string    `gorm:"size:32;not null" json:"quality"`
	Notes      string    `gorm:"size:512" json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
