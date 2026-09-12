package model

import "time"

// CommunityPost 农友社区帖子实体（经验/病虫害/食谱/活动）。
type CommunityPost struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	User         *User     `gorm:"foreignKey:UserID" json:"user"`
	Title        string    `gorm:"size:128;not null" json:"title"`
	Content      string    `gorm:"size:3000;not null" json:"content"`
	PostType     string    `gorm:"size:32;not null;index" json:"post_type"`
	Status       string    `gorm:"size:32;not null;default:published" json:"status"`
	LikeCount    int       `gorm:"not null;default:0" json:"like_count"`
	CommentCount int       `gorm:"not null;default:0" json:"comment_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CommunityComment 社区帖子评论。
type CommunityComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"index;not null" json:"post_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user"`
	Content   string    `gorm:"size:500;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
