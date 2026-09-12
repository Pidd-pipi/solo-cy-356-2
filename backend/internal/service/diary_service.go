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

// DiaryService 种植日记服务。
type DiaryService struct {
	diaryRepo repository.DiaryRepository
	planRepo  repository.PlantingPlanRepository
	logger    *slog.Logger
}

// NewDiaryService 构造种植日记服务。
func NewDiaryService(diaryRepo repository.DiaryRepository, planRepo repository.PlantingPlanRepository, logger *slog.Logger) *DiaryService {
	return &DiaryService{diaryRepo: diaryRepo, planRepo: planRepo, logger: logger}
}

// Create 发布种植日记（校验计划归属）。
func (s *DiaryService) Create(req *dto.CreateDiaryRequest, userID uint) (*model.DiaryEntry, error) {
	plan, err := s.planRepo.FindByID(req.PlanID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("种植计划实体 id=%d 不存在", req.PlanID))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if plan.UserID != userID {
		return nil, util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf("用户 id=%d 无权为他人种植计划 id=%d 写日记", userID, req.PlanID))
	}
	d := &model.DiaryEntry{
		PlanID:     req.PlanID,
		UserID:     userID,
		ActionType: req.ActionType,
		Title:      req.Title,
		Content:    req.Content,
		ImageURL:   req.ImageURL,
	}
	if err := s.diaryRepo.Create(d); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogDiaryCreated, "diary_id", d.ID, "plan_id", d.PlanID, "user_id", userID, "action", d.ActionType)
	return d, nil
}

// Update 更新种植日记（仅作者或管理员）。
func (s *DiaryService) Update(id, userID uint, role string, req *dto.UpdateDiaryRequest) (*model.DiaryEntry, error) {
	d, err := s.diaryRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("种植日记实体 id=%d 不存在", id))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if d.UserID != userID && role != string(constants.RoleAdmin) {
		return nil, util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf("角色 %s 无权修改他人种植日记 id=%d", util.RoleText(role), id))
	}
	if req.ActionType != nil {
		d.ActionType = *req.ActionType
	}
	if req.Title != nil {
		d.Title = *req.Title
	}
	if req.Content != nil {
		d.Content = *req.Content
	}
	if req.ImageURL != nil {
		d.ImageURL = *req.ImageURL
	}
	if err := s.diaryRepo.Update(d); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return d, nil
}

// Delete 删除种植日记（仅作者或管理员）。
func (s *DiaryService) Delete(id, userID uint, role string) error {
	d, err := s.diaryRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("种植日记实体 id=%d 不存在", id))
		}
		return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if d.UserID != userID && role != string(constants.RoleAdmin) {
		return util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf("角色 %s 无权删除他人种植日记 id=%d", util.RoleText(role), id))
	}
	if err := s.diaryRepo.Delete(id); err != nil {
		return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return nil
}

// List 分页查询种植日记。
func (s *DiaryService) List(pq util.PageQuery, userID, planID uint) ([]model.DiaryEntry, int64, error) {
	diaries, total, err := s.diaryRepo.List(pq, userID, planID)
	if err != nil {
		return nil, 0, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return diaries, total, nil
}

// GetByID 查询种植日记详情（含评论）。
func (s *DiaryService) GetByID(id uint) (*model.DiaryEntry, []model.DiaryComment, error) {
	d, err := s.diaryRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("种植日记实体 id=%d 不存在", id))
		}
		return nil, nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	comments, err := s.diaryRepo.ListComments(id)
	if err != nil {
		return nil, nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return d, comments, nil
}

// Like 点赞日记。
func (s *DiaryService) Like(id, userID uint) error {
	if _, err := s.diaryRepo.FindByID(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("种植日记实体 id=%d 不存在", id))
		}
		return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if err := s.diaryRepo.IncrementLike(id); err != nil {
		return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogDiaryLiked, "diary_id", id, "user_id", userID)
	return nil
}

// Comment 评论日记。
func (s *DiaryService) Comment(id, userID uint, content string) (*model.DiaryComment, error) {
	if _, err := s.diaryRepo.FindByID(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("种植日记实体 id=%d 不存在", id))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	c := &model.DiaryComment{DiaryID: id, UserID: userID, Content: content}
	if err := s.diaryRepo.CreateComment(c); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	withUser, err := s.diaryRepo.ListComments(id)
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	for i := range withUser {
		if withUser[i].ID == c.ID {
			s.logger.Info(constants.LogDiaryCommented, "diary_id", id, "comment_id", c.ID, "user_id", userID)
			return &withUser[i], nil
		}
	}
	return c, nil
}
