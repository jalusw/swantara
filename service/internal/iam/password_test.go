package iam

import (
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordService_HashAndVerify(t *testing.T) {
	cfg := &config.Config{}
	cfg.BcryptCost = bcrypt.DefaultCost
	cfg.Pepper = "lorem-ipsum"

	passwordSvc := NewPasswordService(cfg)
	password := "password"
	hash, err := passwordSvc.Hash(password)

	if err != nil {
		t.Error(err)
	}

	if err := passwordSvc.Verify(hash, password); err != nil {
		t.Error(err)
	}
}

func TestPasswordService_Verify_WrongPassword(t *testing.T) {
	cfg := &config.Config{}
	cfg.BcryptCost = bcrypt.MinCost
	cfg.Pepper = "lorem-ipsum"

	passwordSvc := NewPasswordService(cfg)
	hash, err := passwordSvc.Hash("correct-password")
	if err != nil {
		t.Error(err)
	}

	if err := passwordSvc.Verify(hash, "wrong-password"); err == nil {
		t.Error("expected error for wrong password, got nil")
	}
}
