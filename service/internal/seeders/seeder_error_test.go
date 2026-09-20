package seeders

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestSeeder_SeedDemoUsers_ReturnsEmailLookupError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(".*").WillReturnError(errors.New("db down"))

	if err := seeder.SeedDemoUsers(); err == nil {
		t.Fatal("expected email lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_SeedDemoUsers_ReturnsUserLookupError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(".*").WillReturnError(errors.New("db down"))

	if err := seeder.SeedDemoUsers(); err == nil {
		t.Fatal("expected user lookup error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_SeedDemoUsers_ReturnsUserCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(".*").WillReturnError(errors.New("db down"))

	if err := seeder.SeedDemoUsers(); err == nil {
		t.Fatal("expected user create error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_SeedDemoUsers_ReturnsOrgSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(".*").WillReturnError(errors.New("db down"))

	if err := seeder.SeedDemoUsers(); err == nil {
		t.Fatal("expected organization search error")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeeder_SeedDemoUsers_ReturnsProvisionError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	seeder, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatalf("New error = %v", err)
	}

	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(".*").WillReturnError(errors.New("db down"))

	if err := seeder.SeedDemoUsers(); err == nil {
		t.Fatal("expected owner provisioning error")
	}
	query.AssertDBMockDone(t, mock)
}
