package repository

import (
	"testing"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

func TestAuditRepository_ListFilter(t *testing.T) {
	db := newTestDB(t)
	repo := NewAuditRepository(db)
	for i := 0; i < 3; i++ {
		log := &model.AuditLog{UserID: 1, Username: "admin", Role: "admin", Action: "ADOPT_PLOT", ResourceType: "plot", ResourceID: "1", Detail: "认养", RequestID: "req"}
		if err := repo.Create(log); err != nil {
			t.Fatalf("create audit: %v", err)
		}
	}
	log := &model.AuditLog{UserID: 2, Username: "farmer", Role: "farmer", Action: "CREATE_POST", ResourceType: "community-posts", Detail: "发帖", RequestID: "req2"}
	if err := repo.Create(log); err != nil {
		t.Fatalf("create audit: %v", err)
	}
	list, total, err := repo.List(util.PageQuery{Page: 1, PageSize: 10}, "ADOPT_PLOT", "")
	if err != nil || total != 3 || len(list) != 3 {
		t.Errorf("list by action total=%d len=%d err=%v", total, len(list), err)
	}
	byUser, total, err := repo.List(util.PageQuery{Page: 1, PageSize: 10}, "", "farmer")
	if err != nil || total != 1 || len(byUser) != 1 {
		t.Errorf("list by user total=%d len=%d err=%v", total, len(byUser), err)
	}
	got, err := repo.FindByID(list[0].ID)
	if err != nil || got.Action != "ADOPT_PLOT" {
		t.Errorf("FindByID err=%v", err)
	}
}
