package service

import (
	"log/slog"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/repository"
	"github.com/communitygarden/server/internal/util"
)

// DashboardStats 仪表盘统计输出。
type DashboardStats struct {
	UsersByRole      map[string]int64 `json:"users_by_role"`
	PlotsByStatus    map[string]int64 `json:"plots_by_status"`
	PlansByStatus    map[string]int64 `json:"plans_by_status"`
	PostsByType      map[string]int64 `json:"posts_by_type"`
	TotalDiaries     int64            `json:"total_diaries"`
	TotalHarvests    int64            `json:"total_harvests"`
}

// StatsService 统计服务（聚合多个仓储，供仪表盘使用）。
type StatsService struct {
	userRepo    repository.UserRepository
	plotRepo    repository.PlotRepository
	planRepo    repository.PlantingPlanRepository
	harvestRepo repository.HarvestRecordRepository
	diaryRepo   repository.DiaryRepository
	postRepo    repository.CommunityRepository
	logger      *slog.Logger
}

// NewStatsService 构造统计服务。
func NewStatsService(
	userRepo repository.UserRepository,
	plotRepo repository.PlotRepository,
	planRepo repository.PlantingPlanRepository,
	harvestRepo repository.HarvestRecordRepository,
	diaryRepo repository.DiaryRepository,
	postRepo repository.CommunityRepository,
	logger *slog.Logger,
) *StatsService {
	return &StatsService{userRepo: userRepo, plotRepo: plotRepo, planRepo: planRepo, harvestRepo: harvestRepo, diaryRepo: diaryRepo, postRepo: postRepo, logger: logger}
}

// Overview 汇总仪表盘统计。
func (s *StatsService) Overview() (*DashboardStats, error) {
	usersByRole, err := s.userRepo.CountByRole()
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	plotsByStatus, err := s.plotRepo.CountByStatus()
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	plansByStatus, err := s.planRepo.CountByStatus()
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	postsByType, err := s.postRepo.CountByPostType()
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	totalDiaries, err := s.diaryRepo.CountByUser(0)
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	totalHarvests, err := s.harvestRepo.CountByUser(0)
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return &DashboardStats{
		UsersByRole:   usersByRole,
		PlotsByStatus: plotsByStatus,
		PlansByStatus: plansByStatus,
		PostsByType:   postsByType,
		TotalDiaries:  totalDiaries,
		TotalHarvests: totalHarvests,
	}, nil
}
