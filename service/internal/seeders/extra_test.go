package seeders

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestLoadStandardSeeds(t *testing.T) {
	if _, err := loadSeedData[accountingSeed]("data/does-not-exist.json"); err == nil {
		t.Error("expected error for missing file, got nil")
	}
	if got := capitalize("view"); got != "View" {
		t.Errorf("capitalize = %q", got)
	}
	if got := capitalize(""); got != "" {
		t.Errorf("capitalize empty = %q", got)
	}
}

func TestSeedPermissions_Flow(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	svc, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatal(err)
	}

	seed, err := loadIamSeed()
	if err != nil {
		t.Fatal(err)
	}

	for range seed.permissions {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "permissions"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "permissions"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}

	if err := svc.SeedPermissions(); err != nil {
		t.Fatalf("SeedPermissions = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeedPermissions_SkipsExisting(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	svc, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatal(err)
	}

	seed, err := loadIamSeed()
	if err != nil {
		t.Fatal(err)
	}

	for range seed.permissions {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "permissions"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	}

	if err := svc.SeedPermissions(); err != nil {
		t.Fatalf("SeedPermissions = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeedPermissions_ToleratesErrors(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	svc, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatal(err)
	}

	seed, err := loadIamSeed()
	if err != nil {
		t.Fatal(err)
	}

	for i := range seed.permissions {
		if i%2 == 0 {
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "permissions"`)).
				WillReturnError(errors.New("db down"))
		} else {
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "permissions"`)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}))
			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "permissions"`)).
				WillReturnError(errors.New("db down"))
		}
	}

	if err := svc.SeedPermissions(); err != nil {
		t.Fatalf("SeedPermissions = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}
