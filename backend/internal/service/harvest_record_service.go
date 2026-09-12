package service

import (
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/repository"
	"github.com/communitygarden/server/internal/util"
)

// HarvestRecordService 收成记录服务。
type HarvestRecordService struct {
	harvestRepo repository.HarvestRecordRepository
	planRepo    repository.PlantingPlanRepository
	db          *gorm.DB
	logger      *slog.Logger
}

// NewHarvestRecordService 构造收成记录服务。
func NewHarvestRecordService(harvestRepo repository.HarvestRecordRepository, planRepo repository.PlantingPlanRepository, db *gorm.DB, logger *slog.Logger) *HarvestRecordService {
	return &HarvestRecordService{harvestRepo: harvestRepo, planRepo: planRepo, db: db, logger: logger}
}

// Create 记录收成（事务：校验计划归属与成熟状态）。
func (s *HarvestRecordService) Create(req *dto.CreateHarvestRequest, userID uint) (*model.HarvestRecord, error) {
	var created *model.HarvestRecord
	err := s.db.Transaction(func(tx *gorm.DB) error {
		plan, err := s.planRepo.FindByID(req.PlanID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("种植计划实体 id=%d 不存在", req.PlanID))
			}
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		if plan.UserID != userID {
			return util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf("用户 id=%d 无权为他人种植计划 id=%d 记录收成", userID, req.PlanID))
		}
		if plan.Status != string(constants.PlanStatusGrowing) && plan.Status != string(constants.PlanStatusHarvesting) {
			return util.NewAppError(constants.CodeHarvestBeforeMature, 409, fmt.Sprintf("种植计划状态 %s 作物尚未成熟，不允许记录收成", util.PlanStatusText(plan.Status)))
		}
		harvestDate, err := dto.ParseHarvestDate(req.HarvestDate)
		if err != nil {
			return util.NewAppError(constants.CodeValidationFailed, 400, "harvest_date 字段格式必须为 yyyy-MM-dd")
		}
		h := &model.HarvestRecord{
			PlanID:      req.PlanID,
			UserID:      userID,
			CropName:    req.CropName,
			HarvestDate: *harvestDate,
			WeightKg:    req.WeightKg,
			Quality:     req.Quality,
			Notes:       req.Notes,
		}
		if err := s.harvestRepo.CreateWithTx(tx, h); err != nil {
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		created = h
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogHarvestRecorded, "harvest_id", created.ID, "plan_id", created.PlanID, "weight_kg", created.WeightKg, "quality", created.Quality)
	return created, nil
}

// Update 更新收成记录（仅记录本人或管理员）。
func (s *HarvestRecordService) Update(id, userID uint, role string, req *dto.UpdateHarvestRequest) (*model.HarvestRecord, error) {
	h, err := s.harvestRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("收成记录实体 id=%d 不存在", id))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if h.UserID != userID && role != string(constants.RoleAdmin) {
		return nil, util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf("角色 %s 无权修改他人收成记录 id=%d", util.RoleText(role), id))
	}
	if req.CropName != nil {
		h.CropName = *req.CropName
	}
	if req.HarvestDate != nil {
		t, err := dto.ParseHarvestDate(*req.HarvestDate)
		if err != nil {
			return nil, util.NewAppError(constants.CodeValidationFailed, 400, "harvest_date 字段格式必须为 yyyy-MM-dd")
		}
		h.HarvestDate = *t
	}
	if req.WeightKg != nil {
		h.WeightKg = *req.WeightKg
	}
	if req.Quality != nil {
		h.Quality = *req.Quality
	}
	if req.Notes != nil {
		h.Notes = *req.Notes
	}
	if err := s.harvestRepo.Update(h); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return h, nil
}

// Delete 删除收成记录（仅记录本人或管理员）。
func (s *HarvestRecordService) Delete(id, userID uint, role string) error {
	h, err := s.harvestRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("收成记录实体 id=%d 不存在", id))
		}
		return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if h.UserID != userID && role != string(constants.RoleAdmin) {
		return util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf("角色 %s 无权删除他人收成记录 id=%d", util.RoleText(role), id))
	}
	if err := s.harvestRepo.Delete(id); err != nil {
		return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return nil
}

// List 分页查询收成记录（按用户过滤）。
func (s *HarvestRecordService) List(pq util.PageQuery, userID uint) ([]model.HarvestRecord, int64, error) {
	records, total, err := s.harvestRepo.List(pq, userID)
	if err != nil {
		return nil, 0, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return records, total, nil
}

// AnnualStats 年度收成统计（被种植计划统计接口与收成统计接口复用）。
func (s *HarvestRecordService) AnnualStats(userID uint, year int) (*dto.AnnualStatsOutDTO, error) {
	if year <= 0 {
		return nil, util.NewAppError(constants.CodeValidationFailed, 400, "year 字段必须为正整数")
	}
	total, count, err := s.harvestRepo.SumWeightByYear(userID, year)
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	byCrop, err := s.harvestRepo.GroupByCropType(userID, year)
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	byQuality, err := s.harvestRepo.GroupByQuality(userID, year)
	if err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogHarvestStats, "user_id", userID, "year", year, "total_kg", total)
	return &dto.AnnualStatsOutDTO{
		Year:          year,
		TotalWeightKg: total,
		HarvestCount:  count,
		ByCropType:    byCrop,
		ByQuality:     byQuality,
	}, nil
}
