package dto

import (
	"github.com/communitygarden/server/internal/model"
)

// CreateDiaryRequest 发布种植日记。
type CreateDiaryRequest struct {
	PlanID     uint   `json:"plan_id" binding:"required,gt=0"`
	ActionType string `json:"action_type" binding:"required,oneof=sowing watering fertilizing pest_control harvest other"`
	Title      string `json:"title" binding:"required,max=128"`
	Content    string `json:"content" binding:"required,max=2000"`
	ImageURL   string `json:"image_url" binding:"omitempty,max=512"`
}

// UpdateDiaryRequest 更新种植日记。
type UpdateDiaryRequest struct {
	ActionType *string `json:"action_type" binding:"omitempty,oneof=sowing watering fertilizing pest_control harvest other"`
	Title      *string `json:"title" binding:"omitempty,max=128"`
	Content    *string `json:"content" binding:"omitempty,max=2000"`
	ImageURL   *string `json:"image_url" binding:"omitempty,max=512"`
}

// CommentRequest 评论请求。
type CommentRequest struct {
	Content string `json:"content" binding:"required,max=500"`
}

// DiaryOutDTO 种植日记输出。
type DiaryOutDTO struct {
	ID         uint             `json:"id"`
	PlanID     uint             `json:"plan_id"`
	PlanCode   string           `json:"plan_code"`
	UserID     uint             `json:"user_id"`
	Username   string           `json:"username"`
	Nickname   string           `json:"nickname"`
	ActionType string           `json:"action_type"`
	Title      string           `json:"title"`
	Content    string           `json:"content"`
	ImageURL   string           `json:"image_url"`
	LikeCount  int              `json:"like_count"`
	CreatedAt  string           `json:"created_at"`
	Comments   []DiaryCommentDTO `json:"comments,omitempty"`
}

// DiaryCommentDTO 日记评论输出。
type DiaryCommentDTO struct {
	ID        uint   `json:"id"`
	DiaryID   uint   `json:"diary_id"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// ToDiaryOutDTO 模型转 DTO。
func ToDiaryOutDTO(d *model.DiaryEntry) *DiaryOutDTO {
	dto := &DiaryOutDTO{
		ID:         d.ID,
		PlanID:     d.PlanID,
		UserID:     d.UserID,
		ActionType: d.ActionType,
		Title:      d.Title,
		Content:    d.Content,
		ImageURL:   d.ImageURL,
		LikeCount:  d.LikeCount,
		CreatedAt:  d.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if d.Plan != nil {
		dto.PlanCode = d.Plan.CropName
	}
	if d.User != nil {
		dto.Username = d.User.Username
		dto.Nickname = d.User.Nickname
	}
	return dto
}

// ToDiaryCommentDTO 评论模型转 DTO。
func ToDiaryCommentDTO(c *model.DiaryComment) *DiaryCommentDTO {
	dto := &DiaryCommentDTO{
		ID:        c.ID,
		DiaryID:   c.DiaryID,
		UserID:    c.UserID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if c.User != nil {
		dto.Username = c.User.Username
	}
	return dto
}
