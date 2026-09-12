package repository

import (
	"testing"
	"time"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

func TestPlantingPlanRepository_ListAndCount(t *testing.T) {
	db := newTestDB(t)
	repo := NewPlantingPlanRepository(db)
	user := seedUser(t, db, "farmer", "farmer")
	plot := &model.Plot{Name: "P", Code: "P-100", Area: 10, SoilType: "loam", Sunlight: "full", Latitude: 31.0, Longitude: 121.0, Status: "adopted", AdopterID: &user.ID}
	if err := db.Create(plot).Error; err != nil {
		t.Fatalf("create plot: %v", err)
	}
	now := time.Now()
	plans := []struct {
		status string
		crop   string
	}{
		{"planned", "菠菜"}, {"growing", "番茄"}, {"completed", "白菜"},
	}
	for _, p := range plans {
		plan := &model.PlantingPlan{PlotID: plot.ID, UserID: user.ID, CropName: p.crop, CropType: "vegetable", Season: "spring", Status: p.status, PlantDate: &now, ExpectedHarvestDate: &now}
		if err := repo.Create(plan); err != nil {
			t.Fatalf("create plan: %v", err)
		}
	}
	list, total, err := repo.List(util.PageQuery{Page: 1, PageSize: 10}, user.ID, "")
	if err != nil || total != 3 || len(list) != 3 {
		t.Errorf("list: total=%d len=%d err=%v", total, len(list), err)
	}
	cnt, err := repo.CountByUser(user.ID)
	if err != nil || cnt != 3 {
		t.Errorf("CountByUser=%d err=%v", cnt, err)
	}
	active, err := repo.CountActiveByUser(user.ID)
	if err != nil || active != 2 {
		t.Errorf("CountActiveByUser=%d err=%v", active, err)
	}
	counts, err := repo.CountByStatus()
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	if counts["planned"] != 1 || counts["growing"] != 1 || counts["completed"] != 1 {
		t.Errorf("counts = %v", counts)
	}
}
