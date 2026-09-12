package service

import (
	"testing"

	"github.com/communitygarden/server/internal/dto"
	"github.com/communitygarden/server/internal/repository"
)

func TestAuthService_RegisterAndLogin(t *testing.T) {
	db := newTestServiceDB(t)
	userRepo := repository.NewUserRepository(db)
	svc := NewAuthService(userRepo, testLogger(), "test-secret", 72)

	reg := &dto.RegisterRequest{Username: "newbie", Password: "pass123", Nickname: "新农友"}
	u, err := svc.Register(reg)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if u.Role != "citizen" {
		t.Errorf("default role=%s, want citizen", u.Role)
	}
	// 重复用户名
	if _, err := svc.Register(reg); err == nil {
		t.Fatalf("expected duplicate username error")
	}
	// 错误密码
	if _, err := svc.Login(&dto.LoginRequest{Username: "newbie", Password: "wrong"}); err == nil {
		t.Fatalf("expected login error")
	}
	// 正确登录
	resp, err := svc.Login(&dto.LoginRequest{Username: "newbie", Password: "pass123"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if resp.Token == "" || resp.User.Username != "newbie" {
		t.Errorf("login response invalid")
	}
}
