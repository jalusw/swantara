package seeders

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestSeedPaymentTermsForOrg_Errors(t *testing.T) {
	newSvc := func() (*Seeder, sqlmock.Sqlmock) {
		mockDB, mock, _ := sqlmock.New()
		mock.MatchExpectationsInOrder(false)
		db, _ := gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{})
		db.SkipDefaultTransaction = true
		svc, err := New(db, iam.PasswordService{})
		if err != nil {
			t.Fatal(err)
		}
		return svc, mock
	}
	seed, err := loadAccountingSeed()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("list lines error", func(t *testing.T) {
		svc, mock := newSvc()
		for range seed.PaymentTerms {
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_terms"`)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms"`)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}))
			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_terms"`)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		}
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines"`)).
			WillReturnError(errors.New("db down"))
		if err := svc.seedPaymentTermsForOrg(t.Context(), 10, seed); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("replace lines error", func(t *testing.T) {
		svc, mock := newSvc()
		for range seed.PaymentTerms {
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_terms"`)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms"`)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}))
			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_terms"`)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines"`)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines"`)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}))
		}
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "payment_term_lines"`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_term_lines"`)).
			WillReturnError(errors.New("db down"))
		if err := svc.seedPaymentTermsForOrg(t.Context(), 10, seed); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("skips terms with lines", func(t *testing.T) {
		svc, mock := newSvc()
		for range seed.PaymentTerms {
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_terms"`)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms"`)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines"`)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines"`)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		}
		if err := svc.seedPaymentTermsForOrg(t.Context(), 10, seed); err != nil {
			t.Errorf("seedPaymentTermsForOrg = %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet: %v", err)
		}
	})
}

func TestSeedFoundation_Progressive(t *testing.T) {
	newSvc := func() (*Seeder, sqlmock.Sqlmock) {
		mockDB, mock, _ := sqlmock.New()
		mock.MatchExpectationsInOrder(false)
		db, _ := gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{})
		db.SkipDefaultTransaction = true
		svc, err := New(db, iam.PasswordService{})
		if err != nil {
			t.Fatal(err)
		}
		return svc, mock
	}

	currencySearch := func(mock sqlmock.Sqlmock) {
		for range 155 {
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies"`)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(1, "IDR"))
		}
	}

	t.Run("fails at fx rates", func(t *testing.T) {
		svc, mock := newSvc()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		currencySearch(mock)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_rates"`)).
			WillReturnError(errors.New("db down"))
		if err := svc.SeedFoundation(); err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("fails at units", func(t *testing.T) {
		svc, mock := newSvc()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		currencySearch(mock)
		for range 4 {
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_rates"`)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates"`)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}))
			mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "fx_rates"`)).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		}
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "unit_groups"`)).
			WillReturnError(errors.New("db down"))
		if err := svc.SeedFoundation(); err == nil {
			t.Error("expected error, got nil")
		}
	})
}
