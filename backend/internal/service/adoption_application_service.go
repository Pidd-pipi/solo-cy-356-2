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

// AdoptionApplicationService 认养申请服务（审核流 + 候补队列，多步写操作全部走事务）。
type AdoptionApplicationService struct {
	appRepo  repository.AdoptionApplicationRepository
	plotRepo repository.PlotRepository
	db       *gorm.DB
	logger   *slog.Logger
}

// NewAdoptionApplicationService 构造认养申请服务。
func NewAdoptionApplicationService(appRepo repository.AdoptionApplicationRepository, plotRepo repository.PlotRepository, db *gorm.DB, logger *slog.Logger) *AdoptionApplicationService {
	return &AdoptionApplicationService{appRepo: appRepo, plotRepo: plotRepo, db: db, logger: logger}
}

// Apply 提交认养申请。
// 规则：同一居民在同一地块仅允许一份进行中（待审核/候补中）的申请；同一居民同一时间仅允许一份
// 待审核申请（候补不限，候补中的居民可继续申请其他空闲地块）；地块已有待审申请时新申请进入候补队列。
// 并发控制：事务内按 地块行 -> 用户行 的顺序加锁（与晋升/审核/释放的加锁顺序一致，避免 AB-BA 死锁），
// 使同一用户的并发申请、申请与驳回/撤回/释放触发的候补晋升两两串行，保证待审核唯一性。
func (s *AdoptionApplicationService) Apply(userID uint, req *dto.CreateApplicationRequest, username, role string) (*model.AdoptionApplication, error) {
	var created *model.AdoptionApplication
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 先锁地块行：与同地块的并发申请、审核通过、释放后晋升串行化
		plot, err := s.plotRepo.FindByIDForUpdate(tx, req.PlotID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("地块实体 id=%d 不存在", req.PlotID))
			}
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		if plot.Status != string(constants.PlotStatusAvailable) {
			return util.NewAppError(constants.CodePlotNotAvailable, 409,
				fmt.Sprintf("地块 %s 当前状态为 %s，不可申请认养", plot.Code, util.PlotStatusText(plot.Status)))
		}
		// 再锁申请人用户行：与同一用户的并发申请、针对该用户的候补晋升串行化
		if err := s.appRepo.LockApplicant(tx, userID); err != nil {
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		// 同一居民在同一地块仅允许一份进行中的申请（重复提交拦截）
		samePlot, err := s.appRepo.CountActiveByUserOnPlot(tx, userID, req.PlotID)
		if err != nil {
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		if samePlot > 0 {
			return util.NewAppError(constants.CodeApplicationExists, 409,
				fmt.Sprintf("用户 %s（角色 %s）在地块 %s 已存在待审核或候补中的申请，请勿重复提交", username, util.RoleText(role), plot.Code))
		}
		pendingOnPlot, err := s.appRepo.CountPendingByPlot(tx, req.PlotID)
		if err != nil {
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		status := string(constants.ApplicationWaitlisted)
		if pendingOnPlot == 0 {
			// 新申请将成为待审核：同一居民同一时间仅允许一份待审核申请
			pendingByUser, err := s.appRepo.CountPendingByUser(tx, userID)
			if err != nil {
				return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
			}
			if pendingByUser > 0 {
				return util.NewAppError(constants.CodeApplicationExists, 409,
					fmt.Sprintf("用户 %s（角色 %s）已存在待审核的认养申请，同一居民同一时间仅允许一份待审核申请", username, util.RoleText(role)))
			}
			status = string(constants.ApplicationPending)
		}
		app := &model.AdoptionApplication{
			PlotID: req.PlotID,
			UserID: userID,
			Status: status,
			Reason: req.Reason,
		}
		if err := s.appRepo.CreateWithTx(tx, app); err != nil {
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		created = app
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogApplicationSubmitted, "application_id", created.ID, "plot_id", created.PlotID, "user_id", userID, "status", created.Status)
	return s.appRepo.FindByID(created.ID)
}

// ListMine 我的申请分页列表（进度查询）。
func (s *AdoptionApplicationService) ListMine(pq util.PageQuery, userID uint) ([]model.AdoptionApplication, int64, error) {
	apps, total, err := s.appRepo.ListByUser(pq, userID)
	if err != nil {
		return nil, 0, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return apps, total, nil
}

// List 全部申请分页列表（管理员审核入口，可按状态过滤）。
func (s *AdoptionApplicationService) List(pq util.PageQuery, status string) ([]model.AdoptionApplication, int64, error) {
	apps, total, err := s.appRepo.List(pq, status)
	if err != nil {
		return nil, 0, util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
	}
	return apps, total, nil
}

// Withdraw 撤回申请（本人或管理员）：撤回待审核申请后，按申请时间晋升该地块最早候补申请为待审核。
func (s *AdoptionApplicationService) Withdraw(appID, operatorID uint, operatorRole string) (*model.AdoptionApplication, error) {
	var withdrawn *model.AdoptionApplication
	err := s.db.Transaction(func(tx *gorm.DB) error {
		app, err := s.appRepo.FindByIDForUpdate(tx, appID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("认养申请实体 id=%d 不存在", appID))
			}
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		if operatorRole != string(constants.RoleAdmin) && app.UserID != operatorID {
			return util.NewAppError(constants.CodeForbidden, 403,
				fmt.Sprintf("角色 %s 无权撤回用户 id=%d 的认养申请", util.RoleText(operatorRole), app.UserID))
		}
		if app.Status != string(constants.ApplicationPending) && app.Status != string(constants.ApplicationWaitlisted) {
			return util.NewAppError(constants.CodeApplicationNotActionable, 409,
				fmt.Sprintf("认养申请 id=%d 当前状态为 %s，仅待审核/候补中可撤回", app.ID, util.ApplicationStatusText(app.Status)))
		}
		wasPending := app.Status == string(constants.ApplicationPending)
		app.Status = string(constants.ApplicationWithdrawn)
		if err := s.appRepo.UpdateWithTx(tx, app); err != nil {
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		if wasPending {
			if _, err := s.PromoteEarliestWaitlisted(tx, app.PlotID); err != nil {
				return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
			}
		}
		withdrawn = app
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogApplicationWithdrawn, "application_id", withdrawn.ID, "user_id", withdrawn.UserID, "operator_role", operatorRole)
	return s.appRepo.FindByID(withdrawn.ID)
}

// Review 审核认养申请（管理员，仅待审核状态可审核）。
// 通过：同一事务内同步申请状态与地块状态（approved + adopted），任一步失败整体回滚。
// 驳回：申请置为已驳回，并按申请时间晋升该地块最早候补申请为待审核。
func (s *AdoptionApplicationService) Review(appID, reviewerID uint, reviewerName string, req *dto.ReviewApplicationRequest) (*model.AdoptionApplication, error) {
	var reviewed *model.AdoptionApplication
	err := s.db.Transaction(func(tx *gorm.DB) error {
		app, err := s.appRepo.FindByIDForUpdate(tx, appID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("认养申请实体 id=%d 不存在", appID))
			}
			return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
		}
		if app.Status != string(constants.ApplicationPending) {
			return util.NewAppError(constants.CodeApplicationNotActionable, 409,
				fmt.Sprintf("认养申请 id=%d 当前状态为 %s，仅待审核状态可审核", app.ID, util.ApplicationStatusText(app.Status)))
		}
		now := time.Now()
		app.ReviewedBy = &reviewerID
		app.ReviewedAt = &now
		app.ReviewNote = req.Note
		if req.Action == string(constants.ReviewActionApprove) {
			plot, err := s.plotRepo.FindByIDForUpdate(tx, app.PlotID)
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return util.NewAppError(constants.CodeNotFound, 404, fmt.Sprintf("地块实体 id=%d 不存在", app.PlotID))
				}
				return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
			}
			if plot.Status != string(constants.PlotStatusAvailable) {
				return util.NewAppError(constants.CodePlotNotAvailable, 409,
					fmt.Sprintf("地块 %s 当前状态为 %s，无法通过认养申请", plot.Code, util.PlotStatusText(plot.Status)))
			}
			app.Status = string(constants.ApplicationApproved)
			if err := s.appRepo.UpdateWithTx(tx, app); err != nil {
				return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
			}
			plot.Status = string(constants.PlotStatusAdopted)
			plot.AdopterID = &app.UserID
			if err := s.plotRepo.UpdateWithTx(tx, plot); err != nil {
				return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
			}
		} else {
			app.Status = string(constants.ApplicationRejected)
			if err := s.appRepo.UpdateWithTx(tx, app); err != nil {
				return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
			}
			if _, err := s.PromoteEarliestWaitlisted(tx, app.PlotID); err != nil {
				return util.NewAppError(constants.CodeInternalError, 500, constants.ErrorText[constants.CodeInternalError]).Wrap(err)
			}
		}
		reviewed = app
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogApplicationReviewed, "application_id", reviewed.ID, "reviewer", reviewerName, "action", req.Action)
	return s.appRepo.FindByID(reviewed.ID)
}

// PromoteEarliestWaitlisted 将地块最早的可晋升候补申请转为待审核。
// 驳回、撤回待审申请、地块释放后调用；无合格候补时返回 promoted=false。
// 实现 WaitlistPromoter 接口，供 PlotService 释放地块时在同一事务内回调。
// 并发控制：先锁地块行（与同地块并发 Apply/审核通过串行），再按申请时间升序逐个锁定
// 候选人用户行并复核其待审核申请数（与该用户并发 Apply 串行），跳过已持有待审核
// 申请的居民，保证并发下同一居民也不会出现两份待审核申请。
func (s *AdoptionApplicationService) PromoteEarliestWaitlisted(tx *gorm.DB, plotID uint) (bool, error) {
	if _, err := s.plotRepo.FindByIDForUpdate(tx, plotID); err != nil {
		return false, err
	}
	candidates, err := s.appRepo.ListWaitlistedForUpdate(tx, plotID)
	if err != nil {
		return false, err
	}
	for i := range candidates {
		c := &candidates[i]
		if err := s.appRepo.LockApplicant(tx, c.UserID); err != nil {
			return false, err
		}
		pending, err := s.appRepo.CountPendingByUser(tx, c.UserID)
		if err != nil {
			return false, err
		}
		if pending > 0 {
			continue // 已持有待审核申请的居民跳过，不产生第二份待审核
		}
		c.Status = string(constants.ApplicationPending)
		if err := s.appRepo.UpdateWithTx(tx, c); err != nil {
			return false, err
		}
		s.logger.Info(constants.LogApplicationPromoted, "application_id", c.ID, "plot_id", c.PlotID, "user_id", c.UserID)
		return true, nil
	}
	return false, nil
}
