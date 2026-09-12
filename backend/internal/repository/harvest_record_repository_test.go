package repository

import (
	"testing"
	"time"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

func TestHarvestRecordRepository_Stats(t *testing.T) {
	db := newTestDB(t)
	repo := NewHarvestRecordRepository(db)
	user := seedUser(t, db, "farmer", "farmer")
	plot := &model.Plot{Name: "P", Code: "P-200", Area: 10, SoilType: "loam", Sunlight: "full", Latitude: 31.0, Longitude: 121.0, Status: "adopted", AdopterID: &user.ID}
	if err := db.Create(plot).Error; err != nil {
		t.Fatalf("create plot: %v", err)
	}
	now := time.Now()
	plan := &model.PlantingPlan{PlotID: plot.ID, UserID: user.ID, CropName: "番茄", CropType: "vegetable", Season: "summer", Status: "harvesting", PlantDate: &now}
	if err := db.Create(plan).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}
	year := now.Year()
	records := []model.HarvestRecord{
		{PlanID: plan.ID, UserID: user.ID, CropName: "番茄", HarvestDate: time.Date(year, 6, 1, 0, 0, 0, 0, time.Local), WeightKg: 12.5, Quality: "excellent"},
		{PlanID: plan.ID, UserID: user.ID, CropName: "番茄", HarvestDate: time.Date(year, 6, 15, 0, 0, 0, 0, time.Local), WeightKg: 8.0, Quality: "good"},
	}
	for i := range records {
		if err := repo.Create(&records[i]); err != nil {
			t.Fatalf("create harvest: %v", err)
		}
	}
	total, cnt, err := repo.SumWeightByYear(user.ID, year)
	if err != nil || cnt != 2 {
		t.Errorf("SumWeightByYear total=%.2f cnt=%d err=%v", total, cnt, err)
	}
	if total < 20.4 || total > 20.6 {
		t.Errorf("total = %v, want ~20.5", total)
	}
	byCrop, err := repo.GroupByCropType(user.ID, year)
	if err != nil || byCrop["vegetable"] < 20 {
		t.Errorf("GroupByCropType=%v err=%v", byCrop, err)
	}
	byQuality, err := repo.GroupByQuality(user.ID, year)
	if err != nil || byQuality["excellent"] != 1 || byQuality["good"] != 1 {
		t.Errorf("GroupByQuality=%v err=%v", byQuality, err)
	}
	list, totalN, err := repo.List(util.PageQuery{Page: 1, PageSize: 10}, user.ID)
	if err != nil || totalN != 2 || len(list) != 2 {
		t.Errorf("List total=%d len=%d err=%v", totalN, len(list), err)
	}
}
