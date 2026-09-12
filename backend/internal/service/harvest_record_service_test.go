package service

import (
	"testing"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/repository"
)

func TestHarvestRecordService_CreateAndStats(t *testing.T) {
	db := newTestServiceDB(t)
	harvestRepo := repository.NewHarvestRecordRepository(db)
	planRepo := repository.NewPlantingPlanRepository(db)
	svc := NewHarvestRecordService(harvestRepo, planRepo, db, testLogger())

	user := newTestUser(t, db, "farmer", "farmer")
	plot := newTestPlot(t, db, "P-HV", "available", nil)
	plotSvc, _ := newPlotService(t, db)
	if _, err := plotSvc.Adopt(plot.ID, user.ID, "farmer", "farmer"); err != nil {
		t.Fatalf("adopt: %v", err)
	}
	planSvc := NewPlantingPlanService(planRepo, repository.NewPlotRepository(db), plotSvc, db, testLogger())
	plan, err := planSvc.Create(&dto.CreatePlanRequest{PlotID: plot.ID, CropName: "番茄", CropType: "vegetable", Season: "summer"}, user.ID)
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	// planned 状态不能记录收成
	if _, err := svc.Create(&dto.CreateHarvestRequest{PlanID: plan.ID, CropName: "番茄", HarvestDate: "2026-06-01", WeightKg: 5, Quality: "good"}, user.ID); err == nil {
		t.Fatalf("expected immature error")
	}
	if _, err := planSvc.ChangeStatus(plan.ID, user.ID, "farmer", "planting"); err != nil {
		t.Fatalf("change status: %v", err)
	}
	if _, err := planSvc.ChangeStatus(plan.ID, user.ID, "farmer", "growing"); err != nil {
		t.Fatalf("change status: %v", err)
	}
	rec, err := svc.Create(&dto.CreateHarvestRequest{PlanID: plan.ID, CropName: "番茄", HarvestDate: "2026-06-01", WeightKg: 5, Quality: "good"}, user.ID)
	if err != nil {
		t.Fatalf("Create harvest: %v", err)
	}
	if rec.Quality != string(constants.QualityGood) {
		t.Errorf("quality=%s", rec.Quality)
	}
	stats, err := svc.AnnualStats(user.ID, 2026)
	if err != nil || stats.TotalWeightKg != 5 || stats.HarvestCount != 1 {
		t.Errorf("stats=%+v err=%v", stats, err)
	}
}
