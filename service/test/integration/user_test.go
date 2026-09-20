//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestUserCreateAndUpdateIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)

	userDAO := iam.NewUserDAO(testDB)
	userSvc := iam.NewUserService(userDAO, iam.NewPasswordService(testCfg))

	email := gofakeit.Email()
	user, err := userSvc.Create(context.Background(), &iam.User{
		Username:  gofakeit.Username(),
		FirstName: gofakeit.FirstName(),
		Email:     email,
		Password:  "test-password",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected user to be persisted")
	}

	updated, err := userSvc.Update(context.Background(), user.ID, &iam.User{
		Username:  gofakeit.Username(),
		FirstName: "Updated",
		Email:     email,
	}, nil)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if updated.FirstName != "Updated" {
		t.Errorf("expected updated first name, got %q", updated.FirstName)
	}

	loaded, err := userDAO.Find(context.Background(), user.ID)
	if err != nil || loaded == nil {
		t.Fatalf("expected updated user to be findable: %v", err)
	}

	if err := iam.NewPasswordService(testCfg).Verify(loaded.Password, "test-password"); err != nil {
		t.Error("expected password hash to verify")
	}
}

func TestUserCreateDuplicateEmailIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)

	userSvc := iam.NewUserService(iam.NewUserDAO(testDB), iam.NewPasswordService(testCfg))

	email := gofakeit.Email()
	if _, err := userSvc.Create(context.Background(), &iam.User{
		Username:  gofakeit.Username(),
		FirstName: gofakeit.FirstName(),
		Email:     email,
		Password:  "test-password",
	}); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err := userSvc.Create(context.Background(), &iam.User{
		Username:  gofakeit.Username(),
		FirstName: gofakeit.FirstName(),
		Email:     email,
		Password:  "test-password",
	})
	if !errors.Is(err, iam.ErrEmailRegistered) {
		t.Fatalf("expected ErrEmailRegistered, got %v", err)
	}
}
