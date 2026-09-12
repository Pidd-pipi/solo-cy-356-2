package model

import "time"

// Plot 菜园地块实体（GIS 展示 + 认养）。
type Plot struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Code        string    `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Area        float64   `gorm:"not null" json:"area"`
	SoilType    string    `gorm:"size:32;not null" json:"soil_type"`
	Sunlight    string    `gorm:"size:32;not null" json:"sunlight"`
	Latitude    float64   `gorm:"not null" json:"latitude"`
	Longitude   float64   `gorm:"not null" json:"longitude"`
	Status      string    `gorm:"size:32;not null;default:available;index" json:"status"`
	AdopterID   *uint     `gorm:"index" json:"adopter_id"`
	Adopter     *User     `gorm:"foreignKey:AdopterID" json:"adopter"`
	Description string    `gorm:"size:512" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
