package service

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/repository"
	"github.com/communitygarden/server/internal/util"
)

// PlanStatusTransitions 种植计划状态机（服务层 + 前端按钮显隐 + 日志模板 + formatters 多处定义）。
var PlanStatusTransitions = map[constants.PlanStatus][]constants.PlanStatus{
	constants.PlanStatusPlanned:    {constants.PlanStatusPlanting},
	constants.PlanStatusPlanting:   {constants.PlanStatusGrowing},
	constants.PlanStatusGrowing:    {constants.PlanStatusHarvesting},
	constants.PlanStatusHarvesting: {constants.PlanStatusCompleted},
	constants.PlanStatusCompleted:  {},
}

// PlantingPlanService 种植计划服务。
type PlantingPlanService struct {
	planRepo  repository.PlantingPlanRepository
	plotRepo  repository.PlotRepository
	plotSvc   *PlotService
	db        *gorm.DB
	logger    *slog.Logger
}

// NewPlantingPlanService 构造种植计划服务。
func NewPlantingPlanService(planRepo repository.PlantingPlanRepository, plotRepo repository.PlotRepository, plotSvc *PlotService, db *gorm.DB, logger *slog.Logger) *PlantingPlanService {
	return &PlantingPlanService{planRepo: planRepo, plotRepo: plotRepo, plotSvc: plotSvc, db: db, logger: logger}
}

// Create 创建种植计划（事务：锁定地块、校验认养关系、季节推荐校验、生成收获时间线）。
func (s *PlantingPlanService) Create(req *dto.CreatePlanRequest, userID uint) (*model.PlantingPlan, error) {
	var created *model.PlantingPlan
	err := s.db.Transaction(func(tx *gorm.DB) error {
		plot, err := s.plotRepo.FindByIDForUpdate(tx, req.PlotID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("地块实体 id=%d 不存在", req.PlotID))
			}
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		if plot.Status != string(constants.PlotStatusAdopted) || plot.AdopterID == nil || *plot.AdopterID != userID {
			return util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf("地块 %s 未由用户 id=%d 认养，无法创建种植计划", plot.Code, userID))
		}
		season := constants.Season(req.Season)
		crops, ok := constants.SeasonCrops[season]
		if !ok || !containsStr(crops, req.CropName) {
			return util.NewAppError(constants.CodeCropNotInSeason, 400, fmt.Sprintf("作物 %s 不在 %s 推荐列表中", req.CropName, util.SeasonText(req.Season)))
		}
		var plantDateStr string
		if req.PlantDate != nil {
			plantDateStr = *req.PlantDate
		}
		plantDate, err := dto.ParseDate(plantDateStr)
		if err != nil {
			return util.NewAppError(constants.CodeValidationFailed, 400, "plant_date 字段格式必须为 yyyy-MM-dd")
		}
		now := time.Now()
		if plantDate == nil {
			plantDate = &now
		}
		harvest := util.NextHarvestDate(*plantDate, req.CropType)
		plan := &model.PlantingPlan{
			PlotID:              req.PlotID,
			UserID:              userID,
			CropName:            req.CropName,
			CropType:            req.CropType,
			Season:              req.Season,
			Status:              string(constants.PlanStatusPlanned),
			PlantDate:           plantDate,
			ExpectedHarvestDate: &harvest,
			Notes:               req.Notes,
		}
		if err := s.planRepo.CreateWithTx(tx, plan); err != nil {
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		created = plan
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogPlanCreated, "plan_id", created.ID, "plot_id", created.PlotID, "user_id", userID, "crop", created.CropName)
	return created, nil
}

// Update 更新种植计划（计划状态为 planned 时才允许修改）。
func (s *PlantingPlanService) Update(id, userID uint, role string, req *dto.UpdatePlanRequest) (*model.PlantingPlan, error) {
	plan, err := s.planRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("种植计划实体 id=%d 不存在", id))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if plan.UserID != userID && role != string(constants.RoleAdmin) {
		return nil, util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf("角色 %s 无权修改他人种植计划 id=%d", util.RoleText(role), id))
	}
	if plan.Status != string(constants.PlanStatusPlanned) {
		return nil, util.NewAppError(constants.CodePlanStateNotAllowed, 409, fmt.Sprintf("种植计划状态 %s 不允许编辑，仅 planned 可编辑", util.PlanStatusText(plan.Status)))
	}
	if req.CropName != nil {
		plan.CropName = *req.CropName
	}
	if req.CropType != nil {
		plan.CropType = *req.CropType
	}
	if req.Season != nil {
		plan.Season = *req.Season
	}
	if req.Notes != nil {
		plan.Notes = *req.Notes
	}
	if err := s.planRepo.Update(plan); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return plan, nil
}

// ChangeStatus 状态流转（planned->planting->growing->harvesting->completed）。
// 流转到 completed 时在同一事务内将地块标记为待释放（harvested）。
func (s *PlantingPlanService) ChangeStatus(id, userID uint, role, target string) (*model.PlantingPlan, error) {
	plan, err := s.planRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("种植计划实体 id=%d 不存在", id))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if plan.UserID != userID && role != string(constants.RoleAdmin) {
		return nil, util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf("角色 %s 无权操作他人种植计划 id=%d", util.RoleText(role), id))
	}
	cur := constants.PlanStatus(plan.Status)
	next := constants.PlanStatus(target)
	if cur == constants.PlanStatusCompleted {
		return nil, util.NewAppError(constants.CodePlanAlreadyCompleted, 409, "种植计划已完成，无法再次流转")
	}
	allowed := PlanStatusTransitions[cur]
	if !containsPlanStatus(allowed, next) {
		return nil, util.NewAppError(constants.CodePlanStateNotAllowed, 409, fmt.Sprintf("种植计划状态不允许从 %s 流转到 %s", util.PlanStatusText(string(cur)), util.PlanStatusText(string(next))))
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		plan.Status = string(next)
		if err := s.planRepo.UpdateWithTx(tx, plan); err != nil {
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		if next == constants.PlanStatusCompleted {
			if err := s.plotSvc.MarkHarvested(tx, plan.PlotID); err != nil {
				return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
			}
			s.logger.Info(constants.LogPlanCompleted, "plan_id", plan.ID, "user_id", userID, "harvest_count", 0)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogPlanStatusChanged, "plan_id", plan.ID, "user_id", userID, "from", string(cur), "to", string(next))
	return plan, nil
}

// List 分页查询种植计划（按用户过滤）。
func (s *PlantingPlanService) List(pq util.PageQuery, userID uint, status string) ([]model.PlantingPlan, int64, error) {
	plans, total, err := s.planRepo.List(pq, userID, status)
	if err != nil {
		return nil, 0, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return plans, total, nil
}

// ListByUser 按用户查询（与 harvest 列表接口共用同族仓储方法）。
func (s *PlantingPlanService) ListByUser(userID uint, pq util.PageQuery) ([]model.PlantingPlan, int64, error) {
	return s.planRepo.ListByUser(userID, pq)
}

// GetByID 查询种植计划详情。
func (s *PlantingPlanService) GetByID(id uint) (*model.PlantingPlan, error) {
	plan, err := s.planRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("种植计划实体 id=%d 不存在", id))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return plan, nil
}

// Recommendations 季节作物推荐（收成预警时间线基础数据）。
func (s *PlantingPlanService) Recommendations(season string) ([]*dto.RecommendationOutDTO, error) {
	seasonEnum := constants.Season(season)
	if _, ok := constants.SeasonCrops[seasonEnum]; !ok {
		return nil, util.NewAppError(constants.CodeValidationFailed, 400, "season 字段必须是 spring/summer/autumn/winter 之一")
	}
	crops := constants.SeasonCrops[seasonEnum]
	out := make([]*dto.RecommendationOutDTO, 0, len(crops))
	for _, c := range crops {
		out = append(out, &dto.RecommendationOutDTO{
			Season:        season,
			Crops:         []string{c},
			HarvestInDays: 45,
		})
	}
	return out, nil
}

// Reminders 采摘提醒：expected_harvest_date 在未来 7 天内且状态未 completed。
func (s *PlantingPlanService) Reminders(userID uint) ([]model.PlantingPlan, error) {
	plans, _, err := s.planRepo.List(util.PageQuery{Page: 1, PageSize: 50}, userID, "")
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	out := make([]model.PlantingPlan, 0, len(plans))
	for _, p := range plans {
		if p.Status == string(constants.PlanStatusCompleted) || p.ExpectedHarvestDate == nil {
			continue
		}
		days := util.DaysUntil(*p.ExpectedHarvestDate)
		if days >= 0 && days <= 7 {
			out = append(out, p)
		}
	}
	return out, nil
}

// CountByStatus 计划状态统计（仪表盘复用）。
func (s *PlantingPlanService) CountByStatus() (map[string]int64, error) {
	return s.planRepo.CountByStatus()
}

func containsStr(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func containsPlanStatus(list []constants.PlanStatus, v constants.PlanStatus) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
