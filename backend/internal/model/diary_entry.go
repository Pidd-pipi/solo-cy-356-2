package model

import "time"

// DiaryEntry 种植日记实体（图文记录 + 点赞 + 评论）。
type DiaryEntry struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PlanID    uint      `gorm:"index;not null" json:"plan_id"`
	Plan      *PlantingPlan `gorm:"foreignKey:PlanID" json:"plan"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user"`
	ActionType string   `gorm:"size:32;not null" json:"action_type"`
	Title     string    `gorm:"size:128;not null" json:"title"`
	Content   string    `gorm:"size:2000;not null" json:"content"`
	ImageURL  string    `gorm:"size:512" json:"image_url"`
	LikeCount int       `gorm:"not null;default:0" json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DiaryComment 种植日记评论。
type DiaryComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DiaryID   uint      `gorm:"index;not null" json:"diary_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user"`
	Content   string    `gorm:"size:500;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
