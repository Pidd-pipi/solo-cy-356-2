package model

import "time"

// AdoptionApplication 菜园地块认养申请实体（审核流 + 候补队列）。
type AdoptionApplication struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	PlotID     uint       `gorm:"not null;index" json:"plot_id"`
	Plot       *Plot      `gorm:"foreignKey:PlotID" json:"plot"`
	UserID     uint       `gorm:"not null;index" json:"user_id"`
	User       *User      `gorm:"foreignKey:UserID" json:"user"`
	Status     string     `gorm:"size:32;not null;default:pending;index" json:"status"`
	Reason     string     `gorm:"size:512" json:"reason"`
	ReviewNote string     `gorm:"size:512" json:"review_note"`
	ReviewedBy *uint      `gorm:"index" json:"reviewed_by"`
	ReviewedAt *time.Time `json:"reviewed_at"`
	CreatedAt  time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
