package seeders

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"golang.org/x/crypto/bcrypt"
)

func brokenCostConfig() *config.Config {
	return &config.Config{BcryptCost: 100, Pepper: "x"}
}

func testPasswordService() iam.PasswordService {
	return iam.NewPasswordService(&config.Config{BcryptCost: bcrypt.MinCost, Pepper: "test"})
}

func TestSeedDemoUsers_Branches(t *testing.T) {
	t.Run("skips when admin exists", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		db.SkipDefaultTransaction = true
		svc, err := New(db, iam.PasswordService{})
		if err != nil {
			t.Fatal(err)
		}
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
		if err := svc.SeedDemoUsers(); err != nil {
			t.Errorf("SeedDemoUsers = %v", err)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates lookup error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		db.SkipDefaultTransaction = true
		svc, err := New(db, iam.PasswordService{})
		if err != nil {
			t.Fatal(err)
		}
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).WillReturnError(errors.New("db down"))
		if err := svc.SeedDemoUsers(); err == nil {
			t.Error("expected error, got nil")
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates hash error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		db.SkipDefaultTransaction = true
		svc, err := New(db, iam.NewPasswordService(brokenCostConfig()))
		if err != nil {
			t.Fatal(err)
		}
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		if err := svc.SeedDemoUsers(); err == nil {
			t.Error("expected error, got nil")
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates create error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		db.SkipDefaultTransaction = true
		svc, err := New(db, testPasswordService())
		if err != nil {
			t.Fatal(err)
		}
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).WillReturnError(errors.New("db down"))
		if err := svc.SeedDemoUsers(); err == nil {
			t.Error("expected error, got nil")
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates provision error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		db.SkipDefaultTransaction = true
		svc, err := New(db, testPasswordService())
		if err != nil {
			t.Fatal(err)
		}
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "members"`)).WillReturnError(errors.New("db down"))
		if err := svc.SeedDemoUsers(); err == nil {
			t.Error("expected error, got nil")
		}
		query.AssertDBMockDone(t, mock)
	})
}
