package service

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/model"
)

// 本文件为认养申请并发测试（随 go test ./... 默认执行）。
// 使用单连接内存 SQLite：多 goroutine 经连接池串行进入真实 service 事务代码，
// 屏障（chan）同时放行代替固定等待；断言均为与交错顺序无关的数据库最终状态。
// 行锁级（SELECT ... FOR UPDATE）竞态验证见 adoption_application_pg_test.go（PG_TEST_DSN 门控）。

// newConcurrencySQLiteDB 创建单连接内存 SQLite（并发调用经连接池串行化，避免 SQLite 表锁）。
func newConcurrencySQLiteDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(serviceTestDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(
		&model.User{}, &model.Plot{}, &model.AdoptionApplication{}, &model.PlantingPlan{}, &model.HarvestRecord{},
		&model.DiaryEntry{}, &model.DiaryComment{}, &model.CommunityPost{}, &model.CommunityComment{},
		&model.AuditLog{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// pendingCountOf 统计用户待审核申请数（不变式断言用）。
func pendingCountOf(t *testing.T, db *gorm.DB, userID uint) int64 {
	t.Helper()
	var cnt int64
	if err := db.Model(&model.AdoptionApplication{}).
		Where("user_id = ? AND status = ?", userID, string(constants.ApplicationPending)).
		Count(&cnt).Error; err != nil {
		t.Fatalf("count pending: %v", err)
	}
	return cnt
}

// TestConcurrency_SameUserDoubleApply 同一居民并发申请两块无待审地块：至多成功一份待审核。
func TestConcurrency_SameUserDoubleApply(t *testing.T) {
	db := newConcurrencySQLiteDB(t)
	const rounds = 10
	for i := 0; i < rounds; i++ {
		svc, _ := newAppServices(t, db)
		user := newTestUser(t, db, fmt.Sprintf("cu-%d", i), "citizen")
		plotB := newTestPlot(t, db, fmt.Sprintf("P-CU-B-%d", i), "available", nil)
		plotC := newTestPlot(t, db, fmt.Sprintf("P-CU-C-%d", i), "available", nil)

		var succeeded int64
		start := make(chan struct{})
		var wg sync.WaitGroup
		for _, plotID := range []uint{plotB.ID, plotC.ID} {
			wg.Add(1)
			go func(plotID uint) {
				defer wg.Done()
				<-start
				if _, err := svc.Apply(user.ID, &dto.CreateApplicationRequest{PlotID: plotID}, user.Username, "citizen"); err == nil {
					atomic.AddInt64(&succeeded, 1)
				}
			}(plotID)
		}
		close(start)
		wg.Wait()

		if succeeded != 1 {
			t.Errorf("round %d: 成功申请数=%d，期望恰好 1", i, succeeded)
		}
		if cnt := pendingCountOf(t, db, user.ID); cnt != 1 {
			t.Errorf("round %d: 待审核申请数=%d，期望恰好 1", i, cnt)
		}
	}
}

// TestConcurrency_SamePlotApply 多居民并发申请同一空闲地块：仅一份待审核，其余进候补。
func TestConcurrency_SamePlotApply(t *testing.T) {
	db := newConcurrencySQLiteDB(t)
	const rounds = 10
	for i := 0; i < rounds; i++ {
		svc, _ := newAppServices(t, db)
		plot := newTestPlot(t, db, fmt.Sprintf("P-CP-%d", i), "available", nil)
		start := make(chan struct{})
		var wg sync.WaitGroup
		for j := 0; j < 4; j++ {
			u := newTestUser(t, db, fmt.Sprintf("cp-%d-%d", i, j), "citizen")
			wg.Add(1)
			go func(uID uint, name string) {
				defer wg.Done()
				<-start
				_, _ = svc.Apply(uID, &dto.CreateApplicationRequest{PlotID: plot.ID}, name, "citizen")
			}(u.ID, u.Username)
		}
		close(start)
		wg.Wait()

		var pendingCnt, waitlistedCnt int64
		if err := db.Model(&model.AdoptionApplication{}).
			Where("plot_id = ? AND status = ?", plot.ID, string(constants.ApplicationPending)).
			Count(&pendingCnt).Error; err != nil {
			t.Fatalf("count: %v", err)
		}
		if err := db.Model(&model.AdoptionApplication{}).
			Where("plot_id = ? AND status = ?", plot.ID, string(constants.ApplicationWaitlisted)).
			Count(&waitlistedCnt).Error; err != nil {
			t.Fatalf("count: %v", err)
		}
		if pendingCnt != 1 || waitlistedCnt != 3 {
			t.Errorf("round %d: 地块待审核=%d 候补=%d，期望 1 待审核 + 3 候补", i, pendingCnt, waitlistedCnt)
		}
	}
}

// TestConcurrency_ApplyVsRejectPromotion 申请新地块与候补被驳回晋升并发：同一居民至多一份待审核。
func TestConcurrency_ApplyVsRejectPromotion(t *testing.T) {
	db := newConcurrencySQLiteDB(t)
	const rounds = 10
	for i := 0; i < rounds; i++ {
		svc, _ := newAppServices(t, db)
		applicant := newTestUser(t, db, fmt.Sprintf("cr-u-%d", i), "citizen")
		other := newTestUser(t, db, fmt.Sprintf("cr-x-%d", i), "citizen")
		admin := newTestUser(t, db, fmt.Sprintf("cr-a-%d", i), "admin")
		plotA := newTestPlot(t, db, fmt.Sprintf("P-CR-A-%d", i), "available", nil)
		plotB := newTestPlot(t, db, fmt.Sprintf("P-CR-B-%d", i), "available", nil)
		pendingApp := seedApplication(t, db, plotA.ID, other.ID, string(constants.ApplicationPending))
		seedApplication(t, db, plotA.ID, applicant.ID, string(constants.ApplicationWaitlisted))

		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			_, _ = svc.Apply(applicant.ID, &dto.CreateApplicationRequest{PlotID: plotB.ID}, applicant.Username, "citizen")
		}()
		go func() {
			defer wg.Done()
			<-start
			_, _ = svc.Review(pendingApp.ID, admin.ID, admin.Username, &dto.ReviewApplicationRequest{Action: "reject"})
		}()
		close(start)
		wg.Wait()

		if cnt := pendingCountOf(t, db, applicant.ID); cnt > 1 {
			t.Errorf("round %d: 用户出现 %d 份待审核申请（不变式被并发打破）", i, cnt)
		}
	}
}

// TestConcurrency_ApplyVsWithdrawPromotion 申请新地块与候补被撤回晋升并发：同一居民至多一份待审核。
func TestConcurrency_ApplyVsWithdrawPromotion(t *testing.T) {
	db := newConcurrencySQLiteDB(t)
	const rounds = 10
	for i := 0; i < rounds; i++ {
		svc, _ := newAppServices(t, db)
		applicant := newTestUser(t, db, fmt.Sprintf("cw-u-%d", i), "citizen")
		other := newTestUser(t, db, fmt.Sprintf("cw-x-%d", i), "citizen")
		plotA := newTestPlot(t, db, fmt.Sprintf("P-CW-A-%d", i), "available", nil)
		plotB := newTestPlot(t, db, fmt.Sprintf("P-CW-B-%d", i), "available", nil)
		pendingApp := seedApplication(t, db, plotA.ID, other.ID, string(constants.ApplicationPending))
		seedApplication(t, db, plotA.ID, applicant.ID, string(constants.ApplicationWaitlisted))

		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			_, _ = svc.Apply(applicant.ID, &dto.CreateApplicationRequest{PlotID: plotB.ID}, applicant.Username, "citizen")
		}()
		go func() {
			defer wg.Done()
			<-start
			_, _ = svc.Withdraw(pendingApp.ID, other.ID, "citizen")
		}()
		close(start)
		wg.Wait()

		if cnt := pendingCountOf(t, db, applicant.ID); cnt > 1 {
			t.Errorf("round %d: 用户出现 %d 份待审核申请（不变式被并发打破）", i, cnt)
		}
	}
}

// TestConcurrency_ApplyVsReleasePromotion 申请新地块与地块释放触发晋升并发：同一居民至多一份待审核。
func TestConcurrency_ApplyVsReleasePromotion(t *testing.T) {
	db := newConcurrencySQLiteDB(t)
	const rounds = 10
	for i := 0; i < rounds; i++ {
		appSvc, plotSvc := newAppServices(t, db)
		applicant := newTestUser(t, db, fmt.Sprintf("cr2-u-%d", i), "citizen")
		owner := newTestUser(t, db, fmt.Sprintf("cr2-x-%d", i), "citizen")
		admin := newTestUser(t, db, fmt.Sprintf("cr2-a-%d", i), "admin")
		plotA := newTestPlot(t, db, fmt.Sprintf("P-CR2-A-%d", i), "available", nil)
		plotB := newTestPlot(t, db, fmt.Sprintf("P-CR2-B-%d", i), "available", nil)
		// owner 的申请通过并认养 plotA，随后进入待释放；applicant 在 plotA 候补
		ownerApp := seedApplication(t, db, plotA.ID, owner.ID, string(constants.ApplicationPending))
		seedApplication(t, db, plotA.ID, applicant.ID, string(constants.ApplicationWaitlisted))
		if _, err := appSvc.Review(ownerApp.ID, admin.ID, admin.Username, &dto.ReviewApplicationRequest{Action: "approve"}); err != nil {
			t.Fatalf("round %d approve: %v", i, err)
		}
		if err := db.Model(&model.Plot{}).Where("id = ?", plotA.ID).Update("status", string(constants.PlotStatusHarvested)).Error; err != nil {
			t.Fatalf("round %d mark harvested: %v", i, err)
		}

		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			_, _ = appSvc.Apply(applicant.ID, &dto.CreateApplicationRequest{PlotID: plotB.ID}, applicant.Username, "citizen")
		}()
		go func() {
			defer wg.Done()
			<-start
			_, _ = plotSvc.Release(plotA.ID, owner.ID, "citizen")
		}()
		close(start)
		wg.Wait()

		if cnt := pendingCountOf(t, db, applicant.ID); cnt > 1 {
			t.Errorf("round %d: 用户出现 %d 份待审核申请（不变式被并发打破）", i, cnt)
		}
	}
}
