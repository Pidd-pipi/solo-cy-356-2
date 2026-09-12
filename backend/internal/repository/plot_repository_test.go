package repository

import (
	"testing"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

func TestPlotRepository_ListFilter(t *testing.T) {
	db := newTestDB(t)
	repo := NewPlotRepository(db)
	user := seedUser(t, db, "farmer", "farmer")
	seed := func(code, status string, adopter *uint) {
		p := &model.Plot{Name: code, Code: code, Area: 10, SoilType: "loam", Sunlight: "full", Latitude: 31.0, Longitude: 121.0, Status: status, AdopterID: adopter}
		if err := repo.Create(p); err != nil {
			t.Fatalf("create plot: %v", err)
		}
	}
	uid := user.ID
	seed("P-001", "available", nil)
	seed("P-002", "adopted", &uid)
	seed("P-003", "harvested", &uid)

	all, total, err := repo.List(util.PageQuery{Page: 1, PageSize: 10}, "")
	if err != nil || total != 3 || len(all) != 3 {
		t.Errorf("list all: total=%d len=%d err=%v", total, len(all), err)
	}
	available, total, err := repo.List(util.PageQuery{Page: 1, PageSize: 10}, "available")
	if err != nil || total != 1 || len(available) != 1 {
		t.Errorf("list available: total=%d len=%d err=%v", total, len(available), err)
	}

	got, err := repo.FindByCode("P-002")
	if err != nil || got.ID == 0 {
		t.Errorf("FindByCode err=%v", err)
	}
	counts, err := repo.CountByStatus()
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	if counts["available"] != 1 || counts["adopted"] != 1 || counts["harvested"] != 1 {
		t.Errorf("counts = %v", counts)
	}
}
