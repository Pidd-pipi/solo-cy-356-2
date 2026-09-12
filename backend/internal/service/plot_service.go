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

// WaitlistPromoter 候补晋升抽象（由 AdoptionApplicationService 实现）。
// 地块释放后在同一事务内回调，将最早候补申请转为待审核。
type WaitlistPromoter interface {
	PromoteEarliestWaitlisted(tx *gorm.DB, plotID uint) (bool, error)
}

// PlotService 地块服务（认养使用事务 + SELECT FOR UPDATE）。
type PlotService struct {
	plotRepo repository.PlotRepository
	db       *gorm.DB
	logger   *slog.Logger
	promoter WaitlistPromoter
}

// NewPlotService 构造地块服务（promoter 可为 nil，表示释放后不晋升候补）。
func NewPlotService(plotRepo repository.PlotRepository, db *gorm.DB, logger *slog.Logger, promoter WaitlistPromoter) *PlotService {
	return &PlotService{plotRepo: plotRepo, db: db, logger: logger, promoter: promoter}
}

// GetByID 查询地块详情（被地块 handler 与种植计划 service 复用）。
func (s *PlotService) GetByID(id uint) (*model.Plot, error) {
	p, err := s.plotRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("地块实体 id=%d 不存在", id))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return p, nil
}

// Create 创建地块（管理员）。
func (s *PlotService) Create(req *dto.CreatePlotRequest, operator string) (*model.Plot, error) {
	if _, err := s.plotRepo.FindByCode(req.Code); err == nil {
		return nil, util.NewAppError(constants.CodeConflict, 409, fmt.Sprintf("地块编号 %s 已存在", req.Code))
	}
	p := &model.Plot{
		Name:        req.Name,
		Code:        req.Code,
		Area:        req.Area,
		SoilType:    req.SoilType,
		Sunlight:    req.Sunlight,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Status:      string(constants.PlotStatusAvailable),
		Description: req.Description,
	}
	if err := s.plotRepo.Create(p); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	s.logger.Info(constants.LogPlotCreated, "plot_id", p.ID, "code", p.Code, "operator", operator)
	return p, nil
}

// Update 更新地块（管理员）。
func (s *PlotService) Update(id uint, req *dto.UpdatePlotRequest, operator string) (*model.Plot, error) {
	p, err := s.plotRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("地块实体 id=%d 不存在", id))
		}
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Code != nil {
		p.Code = *req.Code
	}
	if req.Area != nil {
		p.Area = *req.Area
	}
	if req.SoilType != nil {
		p.SoilType = *req.SoilType
	}
	if req.Sunlight != nil {
		p.Sunlight = *req.Sunlight
	}
	if req.Latitude != nil {
		p.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		p.Longitude = *req.Longitude
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if err := s.plotRepo.Update(p); err != nil {
		return nil, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return p, nil
}

// List 分页查询地块（可过滤状态）。
func (s *PlotService) List(pq util.PageQuery, status string) ([]model.Plot, int64, error) {
	plots, total, err := s.plotRepo.List(pq, status)
	if err != nil {
		return nil, 0, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return plots, total, nil
}

// Release 释放地块（管理员或认养人，harvested -> available）。
func (s *PlotService) Release(plotID, operatorID uint, operatorRole string) (*model.Plot, error) {
	var released *model.Plot
	err := s.db.Transaction(func(tx *gorm.DB) error {
		plot, err := s.plotRepo.FindByIDForUpdate(tx, plotID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("地块实体 id=%d 不存在", plotID))
			}
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		if operatorRole != string(constants.RoleAdmin) && (plot.AdopterID == nil || *plot.AdopterID != operatorID) {
			return util.NewAppError(constants.CodeForbidden, 403, fmt.Sprintf("角色 %s 无权释放地块 %s", util.RoleText(operatorRole), plot.Code))
		}
		if plot.Status != string(constants.PlotStatusHarvested) {
			return util.NewAppError(constants.CodePlotNotAvailable, 409, fmt.Sprintf("地块 %s 当前状态为 %s，仅待释放状态可释放", plot.Code, util.PlotStatusText(plot.Status)))
		}
		plot.Status = string(constants.PlotStatusAvailable)
		plot.AdopterID = nil
		if err := s.plotRepo.UpdateWithTx(tx, plot); err != nil {
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		// 地块释放后，按申请时间将最早候补申请转为待审核（同一事务，失败整体回滚）
		if s.promoter != nil {
			if _, err := s.promoter.PromoteEarliestWaitlisted(tx, plotID); err != nil {
				return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
			}
		}
		released = plot
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogPlotReleased, "plot_id", released.ID, "code", released.Code, "operator", operatorID)
	return released, nil
}

// MarkHarvested 种植计划完成后将地块置为待释放（harvested）。
func (s *PlotService) MarkHarvested(tx *gorm.DB, plotID uint) error {
	plot, err := s.plotRepo.FindByIDForUpdate(tx, plotID)
	if err != nil {
		return err
	}
	plot.Status = string(constants.PlotStatusHarvested)
	return s.plotRepo.UpdateWithTx(tx, plot)
}

// CountByStatus 地块状态统计（仪表盘复用）。
func (s *PlotService) CountByStatus() (map[string]int64, error) {
	return s.plotRepo.CountByStatus()
}
