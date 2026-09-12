package repository

import (
	"testing"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

func TestDiaryRepository_LikeAndComment(t *testing.T) {
	db := newTestDB(t)
	repo := NewDiaryRepository(db)
	user := seedUser(t, db, "farmer", "farmer")
	plot := &model.Plot{Name: "P", Code: "P-300", Area: 10, SoilType: "loam", Sunlight: "full", Latitude: 31.0, Longitude: 121.0, Status: "adopted", AdopterID: &user.ID}
	if err := db.Create(plot).Error; err != nil {
		t.Fatalf("create plot: %v", err)
	}
	plan := &model.PlantingPlan{PlotID: plot.ID, UserID: user.ID, CropName: "菠菜", CropType: "vegetable", Season: "spring", Status: "growing"}
	if err := db.Create(plan).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}
	diary := &model.DiaryEntry{PlanID: plan.ID, UserID: user.ID, ActionType: "sowing", Title: "播种", Content: "第一周"}
	if err := repo.Create(diary); err != nil {
		t.Fatalf("create diary: %v", err)
	}
	if err := repo.IncrementLike(diary.ID); err != nil {
		t.Fatalf("IncrementLike: %v", err)
	}
	got, err := repo.FindByID(diary.ID)
	if err != nil || got.LikeCount != 1 {
		t.Errorf("like count=%d err=%v", got.LikeCount, err)
	}
	comment := &model.DiaryComment{DiaryID: diary.ID, UserID: user.ID, Content: "不错"}
	if err := repo.CreateComment(comment); err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	comments, err := repo.ListComments(diary.ID)
	if err != nil || len(comments) != 1 {
		t.Errorf("comments len=%d err=%v", len(comments), err)
	}
	list, total, err := repo.List(util.PageQuery{Page: 1, PageSize: 10}, user.ID, 0)
	if err != nil || total != 1 || len(list) != 1 {
		t.Errorf("List total=%d len=%d err=%v", total, len(list), err)
	}
}
