package service

import (
	"testing"

	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/repository"
)

func TestCommunityService_CreateLikeCommentRemove(t *testing.T) {
	db := newTestServiceDB(t)
	postRepo := repository.NewCommunityRepository(db)
	svc := NewCommunityService(postRepo, testLogger())
	user := newTestUser(t, db, "citizen", "citizen")

	post, err := svc.Create(&dto.CreatePostRequest{Title: "活动", Content: "周末除草", PostType: "activity"}, user.ID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := svc.Like(post.ID, user.ID); err != nil {
		t.Fatalf("Like: %v", err)
	}
	comment, err := svc.Comment(post.ID, user.ID, "报名！")
	if err != nil {
		t.Fatalf("Comment: %v", err)
	}
	if comment.Content != "报名！" {
		t.Errorf("comment content=%s", comment.Content)
	}
	got, comments, err := svc.GetByID(post.ID)
	if err != nil || got.LikeCount != 1 || got.CommentCount != 1 || len(comments) != 1 {
		t.Errorf("GetByID invalid: like=%d cc=%d comments=%d err=%v", got.LikeCount, got.CommentCount, len(comments), err)
	}
	if err := svc.Remove(post.ID, user.ID, "citizen"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, _, err := svc.GetByID(post.ID); err == nil {
		t.Fatalf("expected removed post error")
	}
}
