package dto

import (
	"github.com/communitygarden/server/internal/model"
)

// CreatePostRequest 发布社区帖子。
type CreatePostRequest struct {
	Title    string `json:"title" binding:"required,max=128"`
	Content  string `json:"content" binding:"required,max=3000"`
	PostType string `json:"post_type" binding:"required,oneof=experience pest recipe activity"`
}

// UpdatePostRequest 更新社区帖子。
type UpdatePostRequest struct {
	Title    *string `json:"title" binding:"omitempty,max=128"`
	Content  *string `json:"content" binding:"omitempty,max=3000"`
	PostType *string `json:"post_type" binding:"omitempty,oneof=experience pest recipe activity"`
}

// CommunityCommentDTO 社区评论输出。
type CommunityCommentDTO struct {
	ID        uint   `json:"id"`
	PostID    uint   `json:"post_id"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// PostOutDTO 社区帖子输出。
type PostOutDTO struct {
	ID           uint                  `json:"id"`
	UserID       uint                  `json:"user_id"`
	Username     string                `json:"username"`
	Nickname     string                `json:"nickname"`
	Title        string                `json:"title"`
	Content      string                `json:"content"`
	PostType     string                `json:"post_type"`
	Status       string                `json:"status"`
	LikeCount    int                   `json:"like_count"`
	CommentCount int                   `json:"comment_count"`
	CreatedAt    string                `json:"created_at"`
	Comments     []CommunityCommentDTO `json:"comments,omitempty"`
}

// ToPostOutDTO 模型转 DTO。
func ToPostOutDTO(p *model.CommunityPost) *PostOutDTO {
	dto := &PostOutDTO{
		ID:           p.ID,
		UserID:       p.UserID,
		Title:        p.Title,
		Content:      p.Content,
		PostType:     p.PostType,
		Status:       p.Status,
		LikeCount:    p.LikeCount,
		CommentCount: p.CommentCount,
		CreatedAt:    p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if p.User != nil {
		dto.Username = p.User.Username
		dto.Nickname = p.User.Nickname
	}
	return dto
}

// ToCommunityCommentDTO 评论模型转 DTO。
func ToCommunityCommentDTO(c *model.CommunityComment) *CommunityCommentDTO {
	dto := &CommunityCommentDTO{
		ID:        c.ID,
		PostID:    c.PostID,
		UserID:    c.UserID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if c.User != nil {
		dto.Username = c.User.Username
	}
	return dto
}
