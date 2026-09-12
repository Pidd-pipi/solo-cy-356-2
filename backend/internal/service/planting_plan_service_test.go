package service

import (
	"testing"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/repository"
)

func TestPlanStatusTransitions_Table(t *testing.T) {
	tests := []struct {
		from constants.PlanStatus
		to   constants.PlanStatus
		ok   bool
	}{
		{constants.PlanStatusPlanned, constants.PlanStatusPlanting, true},
		{constants.PlanStatusPlanting, constants.PlanStatusGrowing, true},
		{constants.PlanStatusGrowing, constants.PlanStatusHarvesting, true},
		{constants.PlanStatusHarvesting, constants.PlanStatusCompleted, true},
		{constants.PlanStatusPlanned, constants.PlanStatusGrowing, false},
		{constants.PlanStatusCompleted, constants.PlanStatusPlanned, false},
		{constants.PlanStatusGrowing, constants.PlanStatusPlanned, false},
	}
	for _, tt := range tests {
		allowed := PlanStatusTransitions[tt.from]
		found := false
		for _, s := range allowed {
			if s == tt.to {
				found = true
			}
		}
		if found != tt.ok {
			t.Errorf("transition %s->%s ok=%v, want %v", tt.from, tt.to, found, tt.ok)
		}
	}
}

func TestPlantingPlanService_CreateAndFlow(t *testing.T) {
	db := newTestServiceDB(t)
	planRepo := repository.NewPlantingPlanRepository(db)
	plotRepo := repository.NewPlotRepository(db)
	plotSvc, _ := newPlotService(t, db)
	svc := NewPlantingPlanService(planRepo, plotRepo, plotSvc, db, testLogger())

	user := newTestUser(t, db, "farmer", "farmer")
	plot := newTestPlot(t, db, "P-PLAN", "available", nil)
	if _, err := plotSvc.Adopt(plot.ID, user.ID, "farmer", "farmer"); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	// 季节不匹配的作物应报错
	req := &dto.CreatePlanRequest{PlotID: plot.ID, CropName: "西瓜", CropType: "fruit", Season: "winter"}
	if _, err := svc.Create(req, user.ID); err == nil {
		t.Fatalf("expected crop not in season error")
	}

	req = &dto.CreatePlanRequest{PlotID: plot.ID, CropName: "菠菜", CropType: "vegetable", Season: "spring"}
	plan, err := svc.Create(req, user.ID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if plan.Status != string(constants.PlanStatusPlanned) || plan.ExpectedHarvestDate == nil {
		t.Errorf("plan invalid: status=%s", plan.Status)
	}

	// 状态流转到 completed 后地块应变为 harvested
	steps := []string{"planting", "growing", "harvesting", "completed"}
	for _, s := range steps {
		if _, err := svc.ChangeStatus(plan.ID, user.ID, "farmer", s); err != nil {
			t.Fatalf("ChangeStatus(%s): %v", s, err)
		}
	}
	if _, err := svc.ChangeStatus(plan.ID, user.ID, "farmer", "planting"); err == nil {
		t.Fatalf("expected completed plan transition error")
	}
	plotAfter, err := plotSvc.GetByID(plot.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if plotAfter.Status != string(constants.PlotStatusHarvested) {
		t.Errorf("plot status=%s, want harvested", plotAfter.Status)
	}
}

func TestPlantingPlanService_Recommendations(t *testing.T) {
	db := newTestServiceDB(t)
	planRepo := repository.NewPlantingPlanRepository(db)
	plotRepo := repository.NewPlotRepository(db)
	plotSvc, _ := newPlotService(t, db)
	svc := NewPlantingPlanService(planRepo, plotRepo, plotSvc, db, testLogger())
	recs, err := svc.Recommendations("spring")
	if err != nil {
		t.Fatalf("Recommendations: %v", err)
	}
	if len(recs) == 0 {
		t.Fatalf("expected recommendations")
	}
	if _, err := svc.Recommendations("bogus"); err == nil {
		t.Fatalf("expected validation error for bad season")
	}
}
