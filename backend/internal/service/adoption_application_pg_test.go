package service

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/model"
)

// 本文件为 PostgreSQL 行锁级并发集成测试：SELECT ... FOR UPDATE 只有在
// PostgreSQL 上才生效，SQLite 无法复现真实行锁竞态。
// 默认跳过；设置 PG_TEST_DSN 后运行（会清空该库数据，务必使用专用测试库），例如：
//
//	PG_TEST_DSN="host=127.0.0.1 port=25432 user=communitygarden_user password=communitygarden_pwd dbname=communitygarden_it sslmode=disable" \
//	  go test ./internal/service/ -run TestConcurrencyPG -v -count=1

// newPGConcurrencyDB 连接 PG_TEST_DSN 指向的库并重建 schema。
func newPGConcurrencyDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("PG_TEST_DSN")
	if dsn == "" {
		t.Skip("PG_TEST_DSN 未设置，跳过 PostgreSQL 行锁级并发集成测试")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("connect pg: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(16)
	if err := db.AutoMigrate(
		&model.User{}, &model.Plot{}, &model.AdoptionApplication{}, &model.PlantingPlan{}, &model.HarvestRecord{},
		&model.DiaryEntry{}, &model.DiaryComment{}, &model.CommunityPost{}, &model.CommunityComment{},
		&model.AuditLog{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec("TRUNCATE adoption_applications, harvest_records, diary_comments, diary_entries, community_comments, community_posts, planting_plans, audit_logs, plots, users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return db
}

// TestConcurrencyPG_ApplyVsRejectPromotion 申请新地块与候补被驳回晋升并发（真实行锁）。
func TestConcurrencyPG_ApplyVsRejectPromotion(t *testing.T) {
	db := newPGConcurrencyDB(t)
	const rounds = 40
	for i := 0; i < rounds; i++ {
		svc, _ := newAppServices(t, db)
		applicant := newTestUser(t, db, fmt.Sprintf("pgc-u-%d", i), "citizen")
		other := newTestUser(t, db, fmt.Sprintf("pgc-x-%d", i), "citizen")
		admin := newTestUser(t, db, fmt.Sprintf("pgc-a-%d", i), "admin")
		plotA := newTestPlot(t, db, fmt.Sprintf("P-PGC-A-%d", i), "available", nil)
		plotB := newTestPlot(t, db, fmt.Sprintf("P-PGC-B-%d", i), "available", nil)
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
			t.Fatalf("round %d: 用户 %d 出现 %d 份待审核申请（不变式被并发打破）", i, applicant.ID, cnt)
		}
	}
}

// TestConcurrencyPG_ApplyVsWithdrawPromotion 申请新地块与候补被撤回晋升并发（真实行锁）。
func TestConcurrencyPG_ApplyVsWithdrawPromotion(t *testing.T) {
	db := newPGConcurrencyDB(t)
	const rounds = 40
	for i := 0; i < rounds; i++ {
		svc, _ := newAppServices(t, db)
		applicant := newTestUser(t, db, fmt.Sprintf("pgw-u-%d", i), "citizen")
		other := newTestUser(t, db, fmt.Sprintf("pgw-x-%d", i), "citizen")
		plotA := newTestPlot(t, db, fmt.Sprintf("P-PGW-A-%d", i), "available", nil)
		plotB := newTestPlot(t, db, fmt.Sprintf("P-PGW-B-%d", i), "available", nil)
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
			t.Fatalf("round %d: 用户 %d 出现 %d 份待审核申请（不变式被并发打破）", i, applicant.ID, cnt)
		}
	}
}

// TestConcurrencyPG_ApplyVsReleasePromotion 申请新地块与地块释放触发晋升并发（真实行锁）。
func TestConcurrencyPG_ApplyVsReleasePromotion(t *testing.T) {
	db := newPGConcurrencyDB(t)
	const rounds = 40
	for i := 0; i < rounds; i++ {
		appSvc, plotSvc := newAppServices(t, db)
		applicant := newTestUser(t, db, fmt.Sprintf("pgr-u-%d", i), "citizen")
		owner := newTestUser(t, db, fmt.Sprintf("pgr-x-%d", i), "citizen")
		admin := newTestUser(t, db, fmt.Sprintf("pgr-a-%d", i), "admin")
		plotA := newTestPlot(t, db, fmt.Sprintf("P-PGR-A-%d", i), "available", nil)
		plotB := newTestPlot(t, db, fmt.Sprintf("P-PGR-B-%d", i), "available", nil)
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
			t.Fatalf("round %d: 用户 %d 出现 %d 份待审核申请（不变式被并发打破）", i, applicant.ID, cnt)
		}
	}
}

// TestConcurrencyPG_DoubleApply 同一居民并发申请两块空闲地块（真实行锁）：恰好一份待审核。
func TestConcurrencyPG_DoubleApply(t *testing.T) {
	db := newPGConcurrencyDB(t)
	const rounds = 40
	for i := 0; i < rounds; i++ {
		svc, _ := newAppServices(t, db)
		user := newTestUser(t, db, fmt.Sprintf("pgd-u-%d", i), "citizen")
		plotB := newTestPlot(t, db, fmt.Sprintf("P-PGD-B-%d", i), "available", nil)
		plotC := newTestPlot(t, db, fmt.Sprintf("P-PGD-C-%d", i), "available", nil)

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
			t.Fatalf("round %d: 成功申请数=%d，期望恰好 1", i, succeeded)
		}
		if cnt := pendingCountOf(t, db, user.ID); cnt != 1 {
			t.Fatalf("round %d: 用户 %d 待审核申请数=%d，期望恰好 1", i, user.ID, cnt)
		}
	}
}

// TestConcurrencyPG_SamePlotApply 多居民并发申请同一空闲地块（真实行锁）：1 待审核 + 3 候补。
func TestConcurrencyPG_SamePlotApply(t *testing.T) {
	db := newPGConcurrencyDB(t)
	const rounds = 20
	for i := 0; i < rounds; i++ {
		svc, _ := newAppServices(t, db)
		plot := newTestPlot(t, db, fmt.Sprintf("P-PGS-%d", i), "available", nil)
		users := make([]*model.User, 4)
		for j := range users {
			users[j] = newTestUser(t, db, fmt.Sprintf("pgs-u-%d-%d", i, j), "citizen")
		}
		start := make(chan struct{})
		var wg sync.WaitGroup
		for _, u := range users {
			wg.Add(1)
			go func(u *model.User) {
				defer wg.Done()
				<-start
				_, _ = svc.Apply(u.ID, &dto.CreateApplicationRequest{PlotID: plot.ID}, u.Username, "citizen")
			}(u)
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
			t.Fatalf("round %d: 地块 %d 待审核=%d 候补=%d，期望 1 待审核 + 3 候补", i, plot.ID, pendingCnt, waitlistedCnt)
		}
	}
}
