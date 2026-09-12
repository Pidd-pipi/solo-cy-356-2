package service

import (
	"testing"

	"gorm.io/gorm"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/repository"
)

// newAppServices 构造真实仓储的认养申请服务与地块服务（释放联动候补晋升）。
func newAppServices(t *testing.T, db *gorm.DB) (*AdoptionApplicationService, *PlotService) {
	t.Helper()
	appRepo := repository.NewAdoptionApplicationRepository(db)
	plotRepo := repository.NewPlotRepository(db)
	appSvc := NewAdoptionApplicationService(appRepo, plotRepo, db, testLogger())
	plotSvc := NewPlotService(plotRepo, db, testLogger(), appSvc)
	return appSvc, plotSvc
}

func seedApplication(t *testing.T, db *gorm.DB, plotID, userID uint, status string) *model.AdoptionApplication {
	t.Helper()
	a := &model.AdoptionApplication{PlotID: plotID, UserID: userID, Status: status, Reason: "测试申请"}
	if err := db.Create(a).Error; err != nil {
		t.Fatalf("seed application: %v", err)
	}
	return a
}

func TestApplicationService_Apply(t *testing.T) {
	db := newTestServiceDB(t)
	svc, _ := newAppServices(t, db)
	u1 := newTestUser(t, db, "apply-u1", "citizen")
	u2 := newTestUser(t, db, "apply-u2", "citizen")
	plot := newTestPlot(t, db, "P-APP", "available", nil)

	// 地块无待审申请：直接成为待审核
	a1, err := svc.Apply(u1.ID, &dto.CreateApplicationRequest{PlotID: plot.ID, Reason: "想种番茄"}, "apply-u1", "citizen")
	if err != nil {
		t.Fatalf("Apply first: %v", err)
	}
	if a1.Status != string(constants.ApplicationPending) {
		t.Errorf("first application status=%s, want pending", a1.Status)
	}

	// 地块已有待审申请：新申请进入候补队列
	a2, err := svc.Apply(u2.ID, &dto.CreateApplicationRequest{PlotID: plot.ID}, "apply-u2", "citizen")
	if err != nil {
		t.Fatalf("Apply second: %v", err)
	}
	if a2.Status != string(constants.ApplicationWaitlisted) {
		t.Errorf("second application status=%s, want waitlisted", a2.Status)
	}
}

func TestApplicationService_ApplyRules(t *testing.T) {
	db := newTestServiceDB(t)
	svc, _ := newAppServices(t, db)
	u1 := newTestUser(t, db, "rule-u1", "citizen")
	u2 := newTestUser(t, db, "rule-u2", "citizen")
	plotA := newTestPlot(t, db, "P-RULE-A", "available", nil)
	plotB := newTestPlot(t, db, "P-RULE-B", "available", nil)
	plotC := newTestPlot(t, db, "P-RULE-C", "available", nil)

	// u1 申请 plotA -> 待审核
	if _, err := svc.Apply(u1.ID, &dto.CreateApplicationRequest{PlotID: plotA.ID}, "rule-u1", "citizen"); err != nil {
		t.Fatalf("u1 apply plotA: %v", err)
	}
	// 同一地块重复提交被拦截
	if _, err := svc.Apply(u1.ID, &dto.CreateApplicationRequest{PlotID: plotA.ID}, "rule-u1", "citizen"); err == nil {
		t.Fatalf("expected duplicate application on same plot to be blocked")
	}
	// 已有待审核申请，再申请无待审的地块（会成为第二份待审核）被拦截
	if _, err := svc.Apply(u1.ID, &dto.CreateApplicationRequest{PlotID: plotB.ID}, "rule-u1", "citizen"); err == nil {
		t.Fatalf("expected second pending application to be blocked")
	}
	// u2 申请 plotA（已有待审）-> 候补
	a2, err := svc.Apply(u2.ID, &dto.CreateApplicationRequest{PlotID: plotA.ID}, "rule-u2", "citizen")
	if err != nil {
		t.Fatalf("u2 apply plotA: %v", err)
	}
	if a2.Status != string(constants.ApplicationWaitlisted) {
		t.Errorf("u2 plotA status=%s, want waitlisted", a2.Status)
	}
	// 候补中的居民可以申请其他空闲地块 -> 待审核
	a3, err := svc.Apply(u2.ID, &dto.CreateApplicationRequest{PlotID: plotB.ID}, "rule-u2", "citizen")
	if err != nil {
		t.Fatalf("waitlisted u2 apply plotB should be allowed: %v", err)
	}
	if a3.Status != string(constants.ApplicationPending) {
		t.Errorf("u2 plotB status=%s, want pending", a3.Status)
	}
	// u2 已持有待审核（plotB），再申请无待审的 plotC 被拦截
	if _, err := svc.Apply(u2.ID, &dto.CreateApplicationRequest{PlotID: plotC.ID}, "rule-u2", "citizen"); err == nil {
		t.Fatalf("expected second pending application on plotC to be blocked")
	}
	// 已持有待审核的居民申请已有待审的地块 -> 允许进入候补（不增加待审核数量）
	a4, err := svc.Apply(u1.ID, &dto.CreateApplicationRequest{PlotID: plotB.ID}, "rule-u1", "citizen")
	if err != nil {
		t.Fatalf("u1 with pending should be allowed to waitlist on plotB: %v", err)
	}
	if a4.Status != string(constants.ApplicationWaitlisted) {
		t.Errorf("u1 plotB status=%s, want waitlisted", a4.Status)
	}
}

func TestApplicationService_ApplyPlotNotAvailable(t *testing.T) {
	db := newTestServiceDB(t)
	svc, _ := newAppServices(t, db)
	u1 := newTestUser(t, db, "na-u1", "citizen")
	owner := newTestUser(t, db, "na-owner", "farmer")
	plot := newTestPlot(t, db, "P-NA", "adopted", &owner.ID)

	if _, err := svc.Apply(u1.ID, &dto.CreateApplicationRequest{PlotID: plot.ID}, "na-u1", "citizen"); err == nil {
		t.Fatalf("expected error for non-available plot")
	}
	if _, err := svc.Apply(u1.ID, &dto.CreateApplicationRequest{PlotID: 99999}, "na-u1", "citizen"); err == nil {
		t.Fatalf("expected error for missing plot")
	}
}

func TestApplicationService_WithdrawPromotesWaitlisted(t *testing.T) {
	db := newTestServiceDB(t)
	svc, _ := newAppServices(t, db)
	u1 := newTestUser(t, db, "wd-u1", "citizen")
	u2 := newTestUser(t, db, "wd-u2", "citizen")
	u3 := newTestUser(t, db, "wd-u3", "citizen")
	plot := newTestPlot(t, db, "P-WD", "available", nil)

	a1 := seedApplication(t, db, plot.ID, u1.ID, string(constants.ApplicationPending))
	a2 := seedApplication(t, db, plot.ID, u2.ID, string(constants.ApplicationWaitlisted))
	a3 := seedApplication(t, db, plot.ID, u3.ID, string(constants.ApplicationWaitlisted))

	// 撤回待审核申请：最早候补（a2）晋升为待审核，a3 仍为候补
	got, err := svc.Withdraw(a1.ID, u1.ID, "citizen")
	if err != nil {
		t.Fatalf("Withdraw: %v", err)
	}
	if got.Status != string(constants.ApplicationWithdrawn) {
		t.Errorf("withdrawn status=%s", got.Status)
	}
	var promoted, stillWaiting model.AdoptionApplication
	if err := db.First(&promoted, a2.ID).Error; err != nil {
		t.Fatalf("reload a2: %v", err)
	}
	if promoted.Status != string(constants.ApplicationPending) {
		t.Errorf("a2 status=%s, want promoted to pending", promoted.Status)
	}
	if err := db.First(&stillWaiting, a3.ID).Error; err != nil {
		t.Fatalf("reload a3: %v", err)
	}
	if stillWaiting.Status != string(constants.ApplicationWaitlisted) {
		t.Errorf("a3 status=%s, want still waitlisted", stillWaiting.Status)
	}

	// 已撤回的申请不可再次撤回
	if _, err := svc.Withdraw(a1.ID, u1.ID, "citizen"); err == nil {
		t.Fatalf("expected error withdrawing twice")
	}
	// 他人不可撤回
	if _, err := svc.Withdraw(a3.ID, u2.ID, "citizen"); err == nil {
		t.Fatalf("expected forbidden error for non-owner withdraw")
	}
}

func TestApplicationService_PromotionSkipsUserWithPending(t *testing.T) {
	db := newTestServiceDB(t)
	svc, _ := newAppServices(t, db)
	u1 := newTestUser(t, db, "skip-u1", "citizen")
	u2 := newTestUser(t, db, "skip-u2", "citizen")
	u3 := newTestUser(t, db, "skip-u3", "citizen")
	plotA := newTestPlot(t, db, "P-SKIP-A", "available", nil)
	plotB := newTestPlot(t, db, "P-SKIP-B", "available", nil)

	// plotA：u1 待审核，u2/u3 依次候补；u2 在 plotB 已持有待审核
	a1 := seedApplication(t, db, plotA.ID, u1.ID, string(constants.ApplicationPending))
	a2 := seedApplication(t, db, plotA.ID, u2.ID, string(constants.ApplicationWaitlisted))
	a3 := seedApplication(t, db, plotA.ID, u3.ID, string(constants.ApplicationWaitlisted))
	seedApplication(t, db, plotB.ID, u2.ID, string(constants.ApplicationPending))

	// u1 撤回后，最早候补 u2 已有待审核申请（plotB），应跳过并晋升 u3
	if _, err := svc.Withdraw(a1.ID, u1.ID, "citizen"); err != nil {
		t.Fatalf("Withdraw: %v", err)
	}
	var skipped, promoted model.AdoptionApplication
	if err := db.First(&skipped, a2.ID).Error; err != nil {
		t.Fatalf("reload a2: %v", err)
	}
	if skipped.Status != string(constants.ApplicationWaitlisted) {
		t.Errorf("a2 status=%s, want still waitlisted（持有者已有待审核申请应被跳过）", skipped.Status)
	}
	if err := db.First(&promoted, a3.ID).Error; err != nil {
		t.Fatalf("reload a3: %v", err)
	}
	if promoted.Status != string(constants.ApplicationPending) {
		t.Errorf("a3 status=%s, want promoted to pending", promoted.Status)
	}
}

func TestApplicationService_ReviewApproveSyncsPlot(t *testing.T) {
	db := newTestServiceDB(t)
	svc, _ := newAppServices(t, db)
	admin := newTestUser(t, db, "review-admin", "admin")
	applicant := newTestUser(t, db, "review-u1", "citizen")
	plot := newTestPlot(t, db, "P-APR", "available", nil)
	app := seedApplication(t, db, plot.ID, applicant.ID, string(constants.ApplicationPending))

	got, err := svc.Review(app.ID, admin.ID, "review-admin", &dto.ReviewApplicationRequest{Action: "approve", Note: "欢迎认养"})
	if err != nil {
		t.Fatalf("Review approve: %v", err)
	}
	if got.Status != string(constants.ApplicationApproved) {
		t.Errorf("application status=%s, want approved", got.Status)
	}
	if got.ReviewedBy == nil || *got.ReviewedBy != admin.ID || got.ReviewedAt == nil {
		t.Errorf("review fields missing: by=%v at=%v", got.ReviewedBy, got.ReviewedAt)
	}
	// 地块状态同步：adopted + 认养人指向申请人
	var updated model.Plot
	if err := db.First(&updated, plot.ID).Error; err != nil {
		t.Fatalf("reload plot: %v", err)
	}
	if updated.Status != string(constants.PlotStatusAdopted) || updated.AdopterID == nil || *updated.AdopterID != applicant.ID {
		t.Errorf("plot not synced: status=%s adopter=%v", updated.Status, updated.AdopterID)
	}
}

func TestApplicationService_ReviewApproveRollbackWhenPlotUnavailable(t *testing.T) {
	db := newTestServiceDB(t)
	svc, _ := newAppServices(t, db)
	admin := newTestUser(t, db, "rb-admin", "admin")
	applicant := newTestUser(t, db, "rb-u1", "citizen")
	owner := newTestUser(t, db, "rb-owner", "farmer")
	plot := newTestPlot(t, db, "P-RB", "available", nil)
	app := seedApplication(t, db, plot.ID, applicant.ID, string(constants.ApplicationPending))

	// 申请提交后地块被他人直接认养：审核通过必须失败且申请保持待审核（回滚）
	plot.Status = string(constants.PlotStatusAdopted)
	plot.AdopterID = &owner.ID
	if err := db.Save(plot).Error; err != nil {
		t.Fatalf("update plot: %v", err)
	}
	if _, err := svc.Review(app.ID, admin.ID, "rb-admin", &dto.ReviewApplicationRequest{Action: "approve"}); err == nil {
		t.Fatalf("expected approve to fail when plot not available")
	}
	var reloaded model.AdoptionApplication
	if err := db.First(&reloaded, app.ID).Error; err != nil {
		t.Fatalf("reload application: %v", err)
	}
	if reloaded.Status != string(constants.ApplicationPending) {
		t.Errorf("application status=%s after rollback, want pending", reloaded.Status)
	}
}

func TestApplicationService_ReviewRejectPromotesWaitlisted(t *testing.T) {
	db := newTestServiceDB(t)
	svc, _ := newAppServices(t, db)
	admin := newTestUser(t, db, "rej-admin", "admin")
	u1 := newTestUser(t, db, "rej-u1", "citizen")
	u2 := newTestUser(t, db, "rej-u2", "citizen")
	plot := newTestPlot(t, db, "P-REJ", "available", nil)
	a1 := seedApplication(t, db, plot.ID, u1.ID, string(constants.ApplicationPending))
	a2 := seedApplication(t, db, plot.ID, u2.ID, string(constants.ApplicationWaitlisted))

	got, err := svc.Review(a1.ID, admin.ID, "rej-admin", &dto.ReviewApplicationRequest{Action: "reject", Note: "信息不完整"})
	if err != nil {
		t.Fatalf("Review reject: %v", err)
	}
	if got.Status != string(constants.ApplicationRejected) || got.ReviewNote != "信息不完整" {
		t.Errorf("reject result invalid: status=%s note=%s", got.Status, got.ReviewNote)
	}
	var promoted model.AdoptionApplication
	if err := db.First(&promoted, a2.ID).Error; err != nil {
		t.Fatalf("reload a2: %v", err)
	}
	if promoted.Status != string(constants.ApplicationPending) {
		t.Errorf("a2 status=%s, want promoted to pending after reject", promoted.Status)
	}

	// 非待审核状态不可审核
	if _, err := svc.Review(a1.ID, admin.ID, "rej-admin", &dto.ReviewApplicationRequest{Action: "approve"}); err == nil {
		t.Fatalf("expected error reviewing non-pending application")
	}
}

func TestApplicationService_ReleasePromotesWaitlisted(t *testing.T) {
	db := newTestServiceDB(t)
	appSvc, plotSvc := newAppServices(t, db)
	admin := newTestUser(t, db, "rel-admin", "admin")
	u1 := newTestUser(t, db, "rel-u1", "citizen")
	u2 := newTestUser(t, db, "rel-u2", "citizen")
	plot := newTestPlot(t, db, "P-RELS", "available", nil)

	// u1 申请（待审核）→ 审核通过认养 → 计划完成置为待释放 → u2 申请进入候补 → 释放地块 → u2 晋升待审核
	a1 := seedApplication(t, db, plot.ID, u1.ID, string(constants.ApplicationPending))
	if _, err := appSvc.Review(a1.ID, admin.ID, "rel-admin", &dto.ReviewApplicationRequest{Action: "approve"}); err != nil {
		t.Fatalf("approve: %v", err)
	}
	var p model.Plot
	if err := db.First(&p, plot.ID).Error; err != nil {
		t.Fatalf("reload plot: %v", err)
	}
	p.Status = string(constants.PlotStatusHarvested)
	if err := db.Save(&p).Error; err != nil {
		t.Fatalf("mark harvested: %v", err)
	}
	a2 := seedApplication(t, db, plot.ID, u2.ID, string(constants.ApplicationWaitlisted))

	if _, err := plotSvc.Release(plot.ID, u1.ID, "citizen"); err != nil {
		t.Fatalf("Release: %v", err)
	}
	var promoted model.AdoptionApplication
	if err := db.First(&promoted, a2.ID).Error; err != nil {
		t.Fatalf("reload a2: %v", err)
	}
	if promoted.Status != string(constants.ApplicationPending) {
		t.Errorf("a2 status=%s, want promoted to pending after plot release", promoted.Status)
	}
}
