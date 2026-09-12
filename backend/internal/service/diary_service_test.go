package service

import (
	"testing"

	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/repository"
)

func TestDiaryService_CreateLikeComment(t *testing.T) {
	db := newTestServiceDB(t)
	diaryRepo := repository.NewDiaryRepository(db)
	planRepo := repository.NewPlantingPlanRepository(db)
	svc := NewDiaryService(diaryRepo, planRepo, testLogger())

	user := newTestUser(t, db, "farmer", "farmer")
	plot := newTestPlot(t, db, "P-DIARY", "available", nil)
	plotSvc, _ := newPlotService(t, db)
	if _, err := plotSvc.Adopt(plot.ID, user.ID, "farmer", "farmer"); err != nil {
		t.Fatalf("adopt: %v", err)
	}
	planSvc := NewPlantingPlanService(planRepo, repository.NewPlotRepository(db), plotSvc, db, testLogger())
	plan, err := planSvc.Create(&dto.CreatePlanRequest{PlotID: plot.ID, CropName: "生菜", CropType: "vegetable", Season: "spring"}, user.ID)
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	diary, err := svc.Create(&dto.CreateDiaryRequest{PlanID: plan.ID, ActionType: "sowing", Title: "播种", Content: "催芽"}, user.ID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := svc.Like(diary.ID, user.ID); err != nil {
		t.Fatalf("Like: %v", err)
	}
	comment, err := svc.Comment(diary.ID, user.ID, "加油")
	if err != nil || comment.Content != "加油" {
		t.Errorf("Comment err=%v", err)
	}
	got, comments, err := svc.GetByID(diary.ID)
	if err != nil || got.LikeCount != 1 || len(comments) != 1 {
		t.Errorf("GetByID invalid: like=%d comments=%d err=%v", got.LikeCount, len(comments), err)
	}
}
