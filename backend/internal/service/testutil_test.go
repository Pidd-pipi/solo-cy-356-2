package service

import (
	"fmt"
	"log/slog"
	"sync/atomic"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/repository"
)

var serviceTestDBCounter uint64

// serviceTestDSN 生成唯一的内存 SQLite DSN。
func serviceTestDSN() string {
	n := atomic.AddUint64(&serviceTestDBCounter, 1)
	return fmt.Sprintf("file:memdb%d?mode=memory&cache=shared", n)
}

// newTestServiceDB 创建内存 SQLite 测试库。
func newTestServiceDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(serviceTestDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Plot{}, &model.AdoptionApplication{}, &model.PlantingPlan{}, &model.HarvestRecord{},
		&model.DiaryEntry{}, &model.DiaryComment{}, &model.CommunityPost{}, &model.CommunityComment{},
		&model.AuditLog{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(discardWriter{}, nil))
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

// newTestUser 插入测试用户。
func newTestUser(t *testing.T, db *gorm.DB, username, role string) *model.User {
	t.Helper()
	u := &model.User{Username: username, Password: "hash", Nickname: username, Role: role, Status: "active"}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

// newTestPlot 插入测试地块。
func newTestPlot(t *testing.T, db *gorm.DB, code, status string, adopterID *uint) *model.Plot {
	t.Helper()
	p := &model.Plot{Name: code, Code: code, Area: 10, SoilType: "loam", Sunlight: "full", Latitude: 31.0, Longitude: 121.0, Status: status, AdopterID: adopterID}
	if err := db.Create(p).Error; err != nil {
		t.Fatalf("create plot: %v", err)
	}
	return p
}

// adoptPlotDirect 直接将地块置为已认养（模拟申请审核通过后的状态，供测试前置准备使用）。
func adoptPlotDirect(t *testing.T, db *gorm.DB, plotID, userID uint) {
	t.Helper()
	if err := db.Model(&model.Plot{}).Where("id = ?", plotID).
		Updates(map[string]interface{}{"status": "adopted", "adopter_id": userID}).Error; err != nil {
		t.Fatalf("adopt plot direct: %v", err)
	}
}

func newPlotService(t *testing.T, db *gorm.DB) (*PlotService, repository.PlotRepository) {
	t.Helper()
	plotRepo := repository.NewPlotRepository(db)
	svc := NewPlotService(plotRepo, db, testLogger(), nil)
	return svc, plotRepo
}
