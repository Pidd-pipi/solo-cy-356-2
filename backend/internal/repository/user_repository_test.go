package repository

import (
	"errors"
	"testing"

	"github.com/communitygarden/server/internal/util"
)

func TestUserRepository_CRUD(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	tests := []struct {
		name     string
		username string
		role     string
	}{
		{name: "create admin", username: "admin1", role: "admin"},
		{name: "create farmer", username: "farmer1", role: "farmer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := seedUser(t, db, tt.username, tt.role)
			if created.ID == 0 {
				t.Fatalf("expected non-zero id")
			}
			got, err := repo.FindByID(created.ID)
			if err != nil {
				t.Fatalf("FindByID: %v", err)
			}
			if got.Username != tt.username {
				t.Errorf("username = %s, want %s", got.Username, tt.username)
			}
			found, err := repo.FindByUsername(tt.username)
			if err != nil || found.ID != created.ID {
				t.Errorf("FindByUsername err=%v", err)
			}
		})
	}

	// 分页列表
	list, total, err := repo.List(util.PageQuery{Page: 1, PageSize: 10}, "", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Errorf("total=%d len=%d, want 2/2", total, len(list))
	}

	// 删除
	first := list[0]
	if err := repo.Delete(first.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByID(first.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestUserRepository_CountByRole(t *testing.T) {
	db := newTestDB(t)
	seedUser(t, db, "a", "admin")
	seedUser(t, db, "b", "farmer")
	seedUser(t, db, "c", "citizen")
	repo := NewUserRepository(db)
	counts, err := repo.CountByRole()
	if err != nil {
		t.Fatalf("CountByRole: %v", err)
	}
	if counts["admin"] != 1 || counts["farmer"] != 1 || counts["citizen"] != 1 {
		t.Errorf("counts = %v", counts)
	}
}
