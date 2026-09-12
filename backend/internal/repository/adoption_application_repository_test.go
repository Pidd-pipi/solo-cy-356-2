package repository

import (
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

func seedPlotForApp(t *testing.T, db *gorm.DB, code string) *model.Plot {
	t.Helper()
	p := &model.Plot{Name: code, Code: code, Area: 10, SoilType: "loam", Sunlight: "full", Latitude: 31.0, Longitude: 121.0, Status: "available"}
	if err := db.Create(p).Error; err != nil {
		t.Fatalf("seed plot: %v", err)
	}
	return p
}

func seedApp(t *testing.T, db *gorm.DB, plotID, userID uint, status string) *model.AdoptionApplication {
	t.Helper()
	a := &model.AdoptionApplication{PlotID: plotID, UserID: userID, Status: status, Reason: "测试"}
	if err := db.Create(a).Error; err != nil {
		t.Fatalf("seed application: %v", err)
	}
	return a
}

func TestAdoptionApplicationRepository_Counts(t *testing.T) {
	db := newTestDB(t)
	repo := NewAdoptionApplicationRepository(db)
	u1 := seedUser(t, db, "app-u1", "citizen")
	u2 := seedUser(t, db, "app-u2", "citizen")
	plot := seedPlotForApp(t, db, "P-REPO-1")
	plot2 := seedPlotForApp(t, db, "P-REPO-1B")

	seedApp(t, db, plot.ID, u1.ID, string(constants.ApplicationPending))
	seedApp(t, db, plot.ID, u2.ID, string(constants.ApplicationWaitlisted))
	seedApp(t, db, plot2.ID, u2.ID, string(constants.ApplicationPending))

	err := db.Transaction(func(tx *gorm.DB) error {
		n, err := repo.CountActiveByUserOnPlot(tx, u1.ID, plot.ID)
		if err != nil || n != 1 {
			t.Errorf("CountActiveByUserOnPlot u1/plot = %d, want 1 (err=%v)", n, err)
		}
		n, err = repo.CountActiveByUserOnPlot(tx, u1.ID, plot2.ID)
		if err != nil || n != 0 {
			t.Errorf("CountActiveByUserOnPlot u1/plot2 = %d, want 0 (err=%v)", n, err)
		}
		n, err = repo.CountPendingByUser(tx, u2.ID)
		if err != nil || n != 1 {
			t.Errorf("CountPendingByUser u2 = %d, want 1（候补不计入） (err=%v)", n, err)
		}
		pending, err := repo.CountPendingByPlot(tx, plot.ID)
		if err != nil || pending != 1 {
			t.Errorf("CountPendingByPlot = %d, want 1 (err=%v)", pending, err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
}

func TestAdoptionApplicationRepository_FindEarliestEligibleWaitlisted(t *testing.T) {
	db := newTestDB(t)
	repo := NewAdoptionApplicationRepository(db)
	u1 := seedUser(t, db, "wl-u1", "citizen")
	u2 := seedUser(t, db, "wl-u2", "citizen")
	u3 := seedUser(t, db, "wl-u3", "citizen")
	plot := seedPlotForApp(t, db, "P-REPO-2")
	otherPlot := seedPlotForApp(t, db, "P-REPO-2B")

	older := seedApp(t, db, plot.ID, u1.ID, string(constants.ApplicationWaitlisted))
	// 手工构造更早/更晚的申请时间，验证按申请时间升序
	later := seedApp(t, db, plot.ID, u2.ID, string(constants.ApplicationWaitlisted))
	seedApp(t, db, plot.ID, u3.ID, string(constants.ApplicationPending))
	db.Model(&model.AdoptionApplication{}).Where("id = ?", older.ID).Update("created_at", time.Now().Add(-2*time.Hour))
	db.Model(&model.AdoptionApplication{}).Where("id = ?", later.ID).Update("created_at", time.Now().Add(-1*time.Hour))

	err := db.Transaction(func(tx *gorm.DB) error {
		got, err := repo.FindEarliestEligibleWaitlistedForUpdate(tx, plot.ID)
		if err != nil {
			t.Fatalf("FindEarliestEligibleWaitlistedForUpdate: %v", err)
		}
		if got.ID != older.ID {
			t.Errorf("earliest eligible waitlisted = id %d, want %d（最早申请时间）", got.ID, older.ID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}

	// u1 在别的地块持有待审核申请后，其候补应被跳过，晋升 u2
	seedApp(t, db, otherPlot.ID, u1.ID, string(constants.ApplicationPending))
	err = db.Transaction(func(tx *gorm.DB) error {
		got, err := repo.FindEarliestEligibleWaitlistedForUpdate(tx, plot.ID)
		if err != nil {
			t.Fatalf("FindEarliestEligibleWaitlistedForUpdate with pending holder: %v", err)
		}
		if got.ID != later.ID {
			t.Errorf("earliest eligible waitlisted = id %d, want %d（跳过已有待审核申请的居民）", got.ID, later.ID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}

	// 无候补时返回 ErrNotFound
	emptyPlot := seedPlotForApp(t, db, "P-REPO-3")
	err = db.Transaction(func(tx *gorm.DB) error {
		if _, err := repo.FindEarliestEligibleWaitlistedForUpdate(tx, emptyPlot.ID); err != ErrNotFound {
			t.Errorf("want ErrNotFound, got %v", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
}

func TestAdoptionApplicationRepository_ListAndListByUser(t *testing.T) {
	db := newTestDB(t)
	repo := NewAdoptionApplicationRepository(db)
	u1 := seedUser(t, db, "list-u1", "citizen")
	u2 := seedUser(t, db, "list-u2", "citizen")
	plot := seedPlotForApp(t, db, "P-REPO-4")

	seedApp(t, db, plot.ID, u1.ID, string(constants.ApplicationPending))
	seedApp(t, db, plot.ID, u2.ID, string(constants.ApplicationWaitlisted))
	seedApp(t, db, plot.ID, u1.ID, string(constants.ApplicationWithdrawn))

	mine, total, err := repo.ListByUser(util.PageQuery{Page: 1, PageSize: 10}, u1.ID)
	if err != nil || total != 2 || len(mine) != 2 {
		t.Errorf("ListByUser u1: total=%d len=%d err=%v", total, len(mine), err)
	}
	if mine[0].Plot == nil || mine[0].Plot.Code != "P-REPO-4" {
		t.Errorf("ListByUser should preload Plot: %+v", mine[0].Plot)
	}

	all, total, err := repo.List(util.PageQuery{Page: 1, PageSize: 10}, "")
	if err != nil || total != 3 {
		t.Errorf("List all: total=%d err=%v", total, err)
	}
	pending, total, err := repo.List(util.PageQuery{Page: 1, PageSize: 10}, string(constants.ApplicationPending))
	if err != nil || total != 1 || len(pending) != 1 {
		t.Errorf("List pending: total=%d len=%d err=%v", total, len(pending), err)
	}
	_ = all
}
