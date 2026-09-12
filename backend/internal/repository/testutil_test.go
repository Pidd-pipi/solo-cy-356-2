package repository

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/communitygarden/server/internal/model"
)

var testDBCounter uint64

// testDSN 生成唯一的内存 SQLite DSN，避免跨测试共享数据。
func testDSN() string {
	n := atomic.AddUint64(&testDBCounter, 1)
	return fmt.Sprintf("file:memdb%d?mode=memory&cache=shared", n)
}

// newTestDB 创建内存 SQLite 测试库并完成迁移。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(testDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Plot{}, &model.PlantingPlan{}, &model.HarvestRecord{},
		&model.DiaryEntry{}, &model.DiaryComment{}, &model.CommunityPost{}, &model.CommunityComment{},
		&model.AuditLog{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

// seedUser 插入测试用户。
func seedUser(t *testing.T, db *gorm.DB, username, role string) *model.User {
	t.Helper()
	u := &model.User{Username: username, Password: "hash", Nickname: username, Role: role, Status: "active"}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u
}
