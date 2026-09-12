package repository

import (
	"testing"

	"github.com/communitygarden/server/internal/model"
	"github.com/communitygarden/server/internal/util"
)

func TestCommunityRepository_ListAndLike(t *testing.T) {
	db := newTestDB(t)
	repo := NewCommunityRepository(db)
	user := seedUser(t, db, "citizen", "citizen")
	posts := []model.CommunityPost{
		{UserID: user.ID, Title: "经验", Content: "a", PostType: "experience", Status: "published"},
		{UserID: user.ID, Title: "活动", Content: "b", PostType: "activity", Status: "published"},
	}
	for i := range posts {
		if err := repo.Create(&posts[i]); err != nil {
			t.Fatalf("create post: %v", err)
		}
	}
	if err := repo.IncrementLike(posts[0].ID); err != nil {
		t.Fatalf("IncrementLike: %v", err)
	}
	got, err := repo.FindByID(posts[0].ID)
	if err != nil || got.LikeCount != 1 {
		t.Errorf("like count=%d err=%v", got.LikeCount, err)
	}
	exp, total, err := repo.List(util.PageQuery{Page: 1, PageSize: 10}, "experience", "published")
	if err != nil || total != 1 || len(exp) != 1 {
		t.Errorf("list experience total=%d len=%d err=%v", total, len(exp), err)
	}
	counts, err := repo.CountByPostType()
	if err != nil || counts["experience"] != 1 || counts["activity"] != 1 {
		t.Errorf("CountByPostType=%v err=%v", counts, err)
	}
	comment := &model.CommunityComment{PostID: posts[0].ID, UserID: user.ID, Content: "赞"}
	if err := repo.CreateComment(comment); err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	comments, err := repo.ListComments(posts[0].ID)
	if err != nil || len(comments) != 1 {
		t.Errorf("comments len=%d err=%v", len(comments), err)
	}
}
