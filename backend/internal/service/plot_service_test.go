package service

import (
	"fmt"
	"testing"

	"github.com/communitygarden/server/internal/constants"
	"github.com/communitygarden/server/internal/util"
)

func TestPlotService_Release(t *testing.T) {
	db := newTestServiceDB(t)
	svc, _ := newPlotService(t, db)
	owner := newTestUser(t, db, "farmer", "farmer")
	other := newTestUser(t, db, "citizen2", "citizen")
	uid := owner.ID
	plot := newTestPlot(t, db, "P-REL", "harvested", &uid)

	// 非认养人不能释放
	if _, err := svc.Release(plot.ID, other.ID, "citizen"); err == nil {
		t.Fatalf("expected forbidden error for non-owner")
	}
	// 认养人可释放
	got, err := svc.Release(plot.ID, owner.ID, "farmer")
	if err != nil {
		t.Fatalf("Release: %v", err)
	}
	if got.Status != string(constants.PlotStatusAvailable) || got.AdopterID != nil {
		t.Errorf("release result invalid: status=%s adopter=%v", got.Status, got.AdopterID)
	}
}

func TestPlotService_AdoptUsesPageQuery(t *testing.T) {
	// 验证 List 分页复用
	db := newTestServiceDB(t)
	svc, _ := newPlotService(t, db)
	for i := 0; i < 3; i++ {
		newTestPlot(t, db, fmt.Sprintf("P-LIST-%d", i), "available", nil)
	}
	plots, total, err := svc.List(util.PageQuery{Page: 1, PageSize: 2}, "")
	if err != nil || total != 3 {
		t.Errorf("List total=%d err=%v", total, err)
	}
	if len(plots) != 2 {
		t.Errorf("List len=%d, want 2", len(plots))
	}
}
