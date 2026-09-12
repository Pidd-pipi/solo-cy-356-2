package service

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/model"
)

// 本文件为 PostgreSQL 并发集成测试：真实行锁（SELECT ... FOR UPDATE）只有在
// PostgreSQL 上才生效，SQLite 单测无法复现并发竞态。
// 默认跳过；设置 PG_TEST_DSN 后运行，例如：
//
//	PG_TEST_DSN="host=127.0.0.1 port=25432 user=communitygarden_user password=communitygarden_pwd dbname=communitygarden_it sslmode=disable" \
//	  go test ./internal/service/ -run TestConcurrency -v -count=1

// newConcurrencyTestDB 连接 PG_TEST_DSN 指向的库并重建 schema（会清空数据，专用测试库）。
func newConcurrencyTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("PG_TEST_DSN")
	if dsn == "" {
		t.Skip("PG_TEST_DSN 未设置，跳过 PostgreSQL 并发集成测试")
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

// countPendingByUser 统计用户待审核申请数（不变式断言用）。
func countPendingByUser(t *testing.T, db *gorm.DB, userID uint) int64 {
	t.Helper()
	var cnt int64
	if err := db.Model(&model.AdoptionApplication{}).
		Where("user_id = ? AND status = ?", userID, string(constants.ApplicationPending)).
		Count(&cnt).Error; err != nil {
		t.Fatalf("count pending: %v", err)
	}
	return cnt
}

// TestConcurrency_ApplyVsPromotion 居民申请新地块与其候补被晋升同时发生：
// 同一居民也只能保留一份待审核申请。
func TestConcurrency_ApplyVsPromotion(t *testing.T) {
	db := newConcurrencyTestDB(t)
	const rounds = 40
	for i := 0; i < rounds; i++ {
		svc, _ := newAppServices(t, db)
		applicant := newTestUser(t, db, fmt.Sprintf("cc-u-%d", i), "citizen")
		other := newTestUser(t, db, fmt.Sprintf("cc-x-%d", i), "citizen")
		admin := newTestUser(t, db, fmt.Sprintf("cc-a-%d", i), "admin")
		plotA := newTestPlot(t, db, fmt.Sprintf("P-CCA-%d", i), "available", nil)
		plotB := newTestPlot(t, db, fmt.Sprintf("P-CCB-%d", i), "available", nil)
		// other 在 plotA 待审核；applicant 在 plotA 候补
		pendingApp := seedApplication(t, db, plotA.ID, other.ID, string(constants.ApplicationPending))
		seedApplication(t, db, plotA.ID, applicant.ID, string(constants.ApplicationWaitlisted))

		// 并发：applicant 申请 plotB（会成为待审核） + admin 驳回 other 的申请（晋升 applicant 的候补）
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

		if cnt := countPendingByUser(t, db, applicant.ID); cnt > 1 {
			t.Fatalf("round %d: 用户 %d 出现 %d 份待审核申请（不变式被并发打破）", i, applicant.ID, cnt)
		}
	}
}

// TestConcurrency_ApplyVsWithdraw 居民申请新地块与其待审申请被他人撤回触发晋升同时发生：
// 同一居民也只能保留一份待审核申请。
func TestConcurrency_ApplyVsWithdraw(t *testing.T) {
	db := newConcurrencyTestDB(t)
	const rounds = 40
	for i := 0; i < rounds; i++ {
		svc, _ := newAppServices(t, db)
		applicant := newTestUser(t, db, fmt.Sprintf("cw-u-%d", i), "citizen")
		other := newTestUser(t, db, fmt.Sprintf("cw-x-%d", i), "citizen")
		plotA := newTestPlot(t, db, fmt.Sprintf("P-CWA-%d", i), "available", nil)
		plotB := newTestPlot(t, db, fmt.Sprintf("P-CWB-%d", i), "available", nil)
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

		if cnt := countPendingByUser(t, db, applicant.ID); cnt > 1 {
			t.Fatalf("round %d: 用户 %d 出现 %d 份待审核申请（不变式被并发打破）", i, applicant.ID, cnt)
		}
	}
}

// TestConcurrency_DoubleApply 同一居民并发申请两块无待审的空闲地块：至多成功一份待审核。
func TestConcurrency_DoubleApply(t *testing.T) {
	db := newConcurrencyTestDB(t)
	const rounds = 40
	for i := 0; i < rounds; i++ {
		svc, _ := newAppServices(t, db)
		user := newTestUser(t, db, fmt.Sprintf("cd-u-%d", i), "citizen")
		plotB := newTestPlot(t, db, fmt.Sprintf("P-CDB-%d", i), "available", nil)
		plotC := newTestPlot(t, db, fmt.Sprintf("P-CDC-%d", i), "available", nil)

		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			_, _ = svc.Apply(user.ID, &dto.CreateApplicationRequest{PlotID: plotB.ID}, user.Username, "citizen")
		}()
		go func() {
			defer wg.Done()
			<-start
			_, _ = svc.Apply(user.ID, &dto.CreateApplicationRequest{PlotID: plotC.ID}, user.Username, "citizen")
		}()
		close(start)
		wg.Wait()

		if cnt := countPendingByUser(t, db, user.ID); cnt != 1 {
			t.Fatalf("round %d: 用户 %d 待审核申请数=%d，期望恰好 1", i, user.ID, cnt)
		}
	}
}

// TestConcurrency_SamePlotApply 多居民并发申请同一空闲地块：地块仅产生一份待审核，其余进候补。
func TestConcurrency_SamePlotApply(t *testing.T) {
	db := newConcurrencyTestDB(t)
	const rounds = 20
	for i := 0; i < rounds; i++ {
		svc, _ := newAppServices(t, db)
		plot := newTestPlot(t, db, fmt.Sprintf("P-CS-%d", i), "available", nil)
		users := make([]*model.User, 4)
		for j := range users {
			users[j] = newTestUser(t, db, fmt.Sprintf("cs-u-%d-%d", i, j), "citizen")
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
