package service

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/repository"
	"github.com/communitygarden/server/internal/util"
)

// CommunityService 农友社区服务。
type CommunityService struct {
	postRepo repository.CommunityRepository
	logger   *slog.Logger
}

// NewCommunityService 构造社区服务。
func NewCommunityService(postRepo repository.CommunityRepository, logger *slog.Logger) *CommunityService {
	return &CommunityService{postRepo: postRepo, logger: logger}
}

// Create 发布社区帖子。
func (s *CommunityService) Create(req *dto.CreatePostRequest, userID uint) (*model.CommunityPost, error) {
	p := &model.CommunityPost{
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		PostType:  req.PostType,
		Status:    string(constants.PostStatusPublished),
	}
	if err := s.postRepo.Create(p); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogPostCreated, "post_id", p.ID, "user_id", userID, "type", p.PostType)
	return p, nil
}

// Update 更新社区帖子（仅作者或管理员）。
func (s *CommunityService) Update(id, userID uint, role string, req *dto.UpdatePostRequest) (*model.CommunityPost, error) {
	p, err := s.postRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("社区帖子实体 id=%d 不存在", id))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if p.UserID != userID && role != string(constants.RoleAdmin) {
		return nil, util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf("角色 %s 无权修改他人社区帖子 id=%d", util.RoleText(role), id))
	}
	if req.Title != nil {
		p.Title = *req.Title
	}
	if req.Content != nil {
		p.Content = *req.Content
	}
	if req.PostType != nil {
		p.PostType = *req.PostType
	}
	if err := s.postRepo.Update(p); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return p, nil
}

// Remove 软删除帖子（作者或管理员，published -> removed）。
func (s *CommunityService) Remove(id, userID uint, role string) error {
	p, err := s.postRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("社区帖子实体 id=%d 不存在", id))
		}
		return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if p.UserID != userID && role != string(constants.RoleAdmin) {
		return util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf("角色 %s 无权删除他人社区帖子 id=%d", util.RoleText(role), id))
	}
	p.Status = string(constants.PostStatusRemoved)
	if err := s.postRepo.Update(p); err != nil {
		return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogPostRemoved, "post_id", id, "operator", userID, "role", role)
	return nil
}

// List 分页查询社区帖子（可按类型过滤）。
func (s *CommunityService) List(pq util.PageQuery, postType, status string) ([]model.CommunityPost, int64, error) {
	posts, total, err := s.postRepo.List(pq, postType, status)
	if err != nil {
		return nil, 0, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return posts, total, nil
}

// GetByID 查询帖子详情（含评论）。
func (s *CommunityService) GetByID(id uint) (*model.CommunityPost, []model.CommunityComment, error) {
	p, err := s.postRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("社区帖子实体 id=%d 不存在", id))
		}
		return nil, nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if p.Status == string(constants.PostStatusRemoved) {
		return nil, nil, util.NewAppError(constants.CodePostRemoved, 410, constants.ErrorText[constants.CodePostRemoved])
	}
	comments, err := s.postRepo.ListComments(id)
	if err != nil {
		return nil, nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return p, comments, nil
}

// Like 点赞帖子。
func (s *CommunityService) Like(id, userID uint) error {
	if _, err := s.postRepo.FindByID(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("社区帖子实体 id=%d 不存在", id))
		}
		return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if err := s.postRepo.IncrementLike(id); err != nil {
		return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogPostLiked, "post_id", id, "user_id", userID)
	return nil
}

// Comment 评论帖子（评论数 +1）。
func (s *CommunityService) Comment(id, userID uint, content string) (*model.CommunityComment, error) {
	p, err := s.postRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("社区帖子实体 id=%d 不存在", id))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if p.Status == string(constants.PostStatusRemoved) {
		return nil, util.NewAppError(constants.CodePostRemoved, 410, constants.ErrorText[constants.CodePostRemoved])
	}
	c := &model.CommunityComment{PostID: id, UserID: userID, Content: content}
	if err := s.postRepo.CreateComment(c); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	p.CommentCount++
	if err := s.postRepo.Update(p); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	withUser, err := s.postRepo.ListComments(id)
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	for i := range withUser {
		if withUser[i].ID == c.ID {
			s.logger.Info(constants.LogPostCommented, "post_id", id, "comment_id", c.ID, "user_id", userID)
			return &withUser[i], nil
		}
	}
	return c, nil
}

// CountByPostType 帖子类型统计（仪表盘复用）。
func (s *CommunityService) CountByPostType() (map[string]int64, error) {
	return s.postRepo.CountByPostType()
}
