package service

import (
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/repository"
)

// 本文件为审核通过失败路径测试：通过 repository 接口注入写失败，
// 确定性验证"申请与地块状态同步"的多步写在任一步失败时整体回滚（随 go test ./... 执行）。

// failingAppRepo 包装真实申请仓储，在 UpdateWithTx 上注入失败。
type failingAppRepo struct {
	repository.AdoptionApplicationRepository
	failUpdate bool
}

func (r *failingAppRepo) UpdateWithTx(tx *gorm.DB, a *model.AdoptionApplication) error {
	if r.failUpdate {
		return errors.New("injected application update failure")
	}
	return r.AdoptionApplicationRepository.UpdateWithTx(tx, a)
}

// failingPlotRepo 包装真实地块仓储，在 UpdateWithTx 上注入失败。
type failingPlotRepo struct {
	repository.PlotRepository
	failUpdate bool
}

func (r *failingPlotRepo) UpdateWithTx(tx *gorm.DB, p *model.Plot) error {
	if r.failUpdate {
		return errors.New("injected plot update failure")
	}
	return r.PlotRepository.UpdateWithTx(tx, p)
}

// reloadAppAndPlot 重新读取申请与地块的数据库最终状态。
func reloadAppAndPlot(t *testing.T, db *gorm.DB, appID, plotID uint) (*model.AdoptionApplication, *model.Plot) {
	t.Helper()
	var app model.AdoptionApplication
	if err := db.First(&app, appID).Error; err != nil {
		t.Fatalf("reload application: %v", err)
	}
	var plot model.Plot
	if err := db.First(&plot, plotID).Error; err != nil {
		t.Fatalf("reload plot: %v", err)
	}
	return &app, &plot
}

// TestReviewApprove_RollbackOnApplicationUpdateFailure 第一步（申请置为通过）失败：
// 接口返回错误，申请保持待审核，地块保持空闲（无部分写入）。
func TestReviewApprove_RollbackOnApplicationUpdateFailure(t *testing.T) {
	db := newTestServiceDB(t)
	appRepo := repository.NewAdoptionApplicationRepository(db)
	plotRepo := repository.NewPlotRepository(db)
	svc := NewAdoptionApplicationService(&failingAppRepo{appRepo, true}, plotRepo, db, testLogger())

	admin := newTestUser(t, db, "fail-admin", "admin")
	applicant := newTestUser(t, db, "fail-u1", "citizen")
	plot := newTestPlot(t, db, "P-FAIL-APP", "available", nil)
	app := seedApplication(t, db, plot.ID, applicant.ID, string(constants.ApplicationPending))

	// 接口状态：返回错误
	if _, err := svc.Review(app.ID, admin.ID, "fail-admin", &dto.ReviewApplicationRequest{Action: "approve"}); err == nil {
		t.Fatalf("expected injected failure error")
	}
	// 数据库最终状态：申请仍待审核，地块仍空闲且无认养人
	gotApp, gotPlot := reloadAppAndPlot(t, db, app.ID, plot.ID)
	if gotApp.Status != string(constants.ApplicationPending) {
		t.Errorf("application status=%s after rollback, want pending", gotApp.Status)
	}
	if gotApp.ReviewedBy != nil || gotApp.ReviewedAt != nil {
		t.Errorf("review fields should be rolled back: by=%v at=%v", gotApp.ReviewedBy, gotApp.ReviewedAt)
	}
	if gotPlot.Status != string(constants.PlotStatusAvailable) || gotPlot.AdopterID != nil {
		t.Errorf("plot should stay available without adopter: status=%s adopter=%v", gotPlot.Status, gotPlot.AdopterID)
	}
}

// TestReviewApprove_RollbackOnPlotUpdateFailure 第二步（地块置为已认养）失败：
// 接口返回错误，已写入的申请状态一并回滚，两边都不留半成品。
func TestReviewApprove_RollbackOnPlotUpdateFailure(t *testing.T) {
	db := newTestServiceDB(t)
	appRepo := repository.NewAdoptionApplicationRepository(db)
	plotRepo := repository.NewPlotRepository(db)
	svc := NewAdoptionApplicationService(appRepo, &failingPlotRepo{plotRepo, true}, db, testLogger())

	admin := newTestUser(t, db, "fail-admin2", "admin")
	applicant := newTestUser(t, db, "fail-u2", "citizen")
	plot := newTestPlot(t, db, "P-FAIL-PLOT", "available", nil)
	app := seedApplication(t, db, plot.ID, applicant.ID, string(constants.ApplicationPending))

	// 接口状态：返回错误
	if _, err := svc.Review(app.ID, admin.ID, "fail-admin2", &dto.ReviewApplicationRequest{Action: "approve"}); err == nil {
		t.Fatalf("expected injected failure error")
	}
	// 数据库最终状态：申请仍待审核（第一步写入被回滚），地块仍空闲且无认养人
	gotApp, gotPlot := reloadAppAndPlot(t, db, app.ID, plot.ID)
	if gotApp.Status != string(constants.ApplicationPending) {
		t.Errorf("application status=%s after rollback, want pending（第一步写入必须一并回滚）", gotApp.Status)
	}
	if gotApp.ReviewedBy != nil || gotApp.ReviewedAt != nil {
		t.Errorf("review fields should be rolled back: by=%v at=%v", gotApp.ReviewedBy, gotApp.ReviewedAt)
	}
	if gotPlot.Status != string(constants.PlotStatusAvailable) || gotPlot.AdopterID != nil {
		t.Errorf("plot should stay available without adopter: status=%s adopter=%v", gotPlot.Status, gotPlot.AdopterID)
	}
}

// TestReviewApprove_RollbackWhenPlotUnavailable 审核时地块已被他人认养（业务失败）：
// 接口返回冲突错误，申请保持待审核，地块状态保持他人认养不变。
func TestReviewApprove_RollbackWhenPlotUnavailable(t *testing.T) {
	db := newTestServiceDB(t)
	svc, _ := newAppServices(t, db)
	admin := newTestUser(t, db, "na-admin", "admin")
	applicant := newTestUser(t, db, "na-u1", "citizen")
	owner := newTestUser(t, db, "na-owner", "farmer")
	plot := newTestPlot(t, db, "P-FAIL-NA", "available", nil)
	app := seedApplication(t, db, plot.ID, applicant.ID, string(constants.ApplicationPending))

	// 申请提交后地块被他人认养
	adoptPlotDirect(t, db, plot.ID, owner.ID)

	if _, err := svc.Review(app.ID, admin.ID, "na-admin", &dto.ReviewApplicationRequest{Action: "approve"}); err == nil {
		t.Fatalf("expected approve to fail when plot not available")
	}
	gotApp, gotPlot := reloadAppAndPlot(t, db, app.ID, plot.ID)
	if gotApp.Status != string(constants.ApplicationPending) {
		t.Errorf("application status=%s after rollback, want pending", gotApp.Status)
	}
	if gotPlot.Status != string(constants.PlotStatusAdopted) || gotPlot.AdopterID == nil || *gotPlot.AdopterID != owner.ID {
		t.Errorf("plot should remain adopted by owner: status=%s adopter=%v", gotPlot.Status, gotPlot.AdopterID)
	}
}
