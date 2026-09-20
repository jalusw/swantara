//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func newIntegrationAuthService(t *testing.T) iam.AuthService {
	t.Helper()
	return iam.NewAuthService(
		iam.NewUserDAO(testDB),
		iam.NewUserService(iam.NewUserDAO(testDB), iam.NewPasswordService(testCfg)),
		iam.NewUserSessionDAO(testDB),
		iam.NewPasswordService(testCfg),
		iam.NewTokenService(testCfg),
		iam.NewUserEmailVerificationDAO(testDB),
		iam.NewUserPasswordResetDAO(testDB),
		queue.NewQueueClient(testCfg),
		testCfg.AuthEmailVerifyTTLDuration,
		testCfg.AuthPasswordResetTTLDuration,
	)
}

func TestRegisterLoginIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)

	passwordSvc := iam.NewPasswordService(testCfg)
	authSvc := newIntegrationAuthService(t)

	email := gofakeit.Email()
	user, err := authSvc.Register(context.Background(), iam.RegisterData{
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Email:     email,
		Password:  "test-password",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected user to be persisted with an ID")
	}

	if err := passwordSvc.Verify(user.Password, "test-password"); err != nil {
		t.Error("stored password hash should verify")
	}

	result, err := authSvc.Login(context.Background(), email, "test-password", iam.SessionClientInfo{})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Error("expected access and refresh tokens")
	}
}

func TestRegisterDuplicateEmailIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)

	authSvc := newIntegrationAuthService(t)

	data := iam.RegisterData{
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Email:     gofakeit.Email(),
		Password:  "test-password",
	}

	if _, err := authSvc.Register(context.Background(), data); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	if _, err := authSvc.Register(context.Background(), data); !errors.Is(err, iam.ErrEmailRegistered) {
		t.Fatalf("expected ErrEmailRegistered, got %v", err)
	}
}

func TestLoginWrongPasswordIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)

	authSvc := newIntegrationAuthService(t)

	email := gofakeit.Email()
	if _, err := authSvc.Register(context.Background(), iam.RegisterData{
		FirstName: gofakeit.FirstName(),
		Email:     email,
		Password:  "test-password",
	}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if _, err := authSvc.Login(context.Background(), email, "wrong-password", iam.SessionClientInfo{}); !errors.Is(err, iam.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestEmailVerificationIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)

	authSvc := newIntegrationAuthService(t)
	user, err := authSvc.Register(context.Background(), iam.RegisterData{
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Email:     gofakeit.Email(),
		Password:  "test-password",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	rawToken := "raw-verification-token"
	verificationDAO := iam.NewUserEmailVerificationDAO(testDB)
	if _, err := verificationDAO.Create(context.Background(), &iam.UserEmailVerification{
		UserID:            user.ID,
		VerificationToken: helper.HashToken(rawToken),
		ExpiresAt:         time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("create verification failed: %v", err)
	}

	if err := authSvc.VerifyEmail(context.Background(), rawToken); err != nil {
		t.Fatalf("verify email failed: %v", err)
	}

	reloaded, err := iam.NewUserDAO(testDB).Find(context.Background(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.EmailVerifiedAt == nil {
		t.Error("expected EmailVerifiedAt to be set")
	}

	verification, err := verificationDAO.FindByToken(context.Background(), helper.HashToken(rawToken))
	if err != nil {
		t.Fatal(err)
	}
	if verification == nil || verification.UsedAt == nil {
		t.Error("expected verification to be marked used")
	}

	if err := authSvc.VerifyEmail(context.Background(), rawToken); !errors.Is(err, iam.ErrInvalidActionToken) {
		t.Fatalf("expected ErrInvalidActionToken on reuse, got %v", err)
	}
}

func TestPasswordResetIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)

	authSvc := newIntegrationAuthService(t)
	user, err := authSvc.Register(context.Background(), iam.RegisterData{
		FirstName: gofakeit.FirstName(),
		Email:     gofakeit.Email(),
		Password:  "test-password",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	rawToken := "raw-reset-token"
	resetDAO := iam.NewUserPasswordResetDAO(testDB)
	if _, err := resetDAO.Create(context.Background(), &iam.UserPasswordReset{
		UserID:         user.ID,
		ResetTokenHash: helper.HashToken(rawToken),
		ExpiresAt:      time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("create reset failed: %v", err)
	}

	if err := authSvc.ResetPassword(context.Background(), rawToken, "new-password"); err != nil {
		t.Fatalf("reset password failed: %v", err)
	}

	reloaded, err := iam.NewUserDAO(testDB).Find(context.Background(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := iam.NewPasswordService(testCfg).Verify(reloaded.Password, "new-password"); err != nil {
		t.Error("expected new password to verify")
	}

	reset, err := resetDAO.FindByTokenHash(context.Background(), helper.HashToken(rawToken))
	if err != nil {
		t.Fatal(err)
	}
	if reset == nil || reset.UsedAt == nil {
		t.Error("expected reset to be marked used")
	}

	if err := authSvc.ResetPassword(context.Background(), rawToken, "another-password"); !errors.Is(err, iam.ErrInvalidActionToken) {
		t.Fatalf("expected ErrInvalidActionToken on reuse, got %v", err)
	}
}
