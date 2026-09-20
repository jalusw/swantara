package seeders

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestLoadSeedData_Mismatch(t *testing.T) {
	if _, err := loadSeedData[string]("data/iam_seed.json"); err == nil {
		t.Error("expected unmarshal error, got nil")
	}
}

func TestSeedDemoUsers_OrgError(t *testing.T) {
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
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).WillReturnError(errors.New("db down"))

	if err := svc.SeedDemoUsers(); err == nil {
		t.Error("expected error, got nil")
	}
	query.AssertDBMockDone(t, mock)
}

func TestSeedPaymentTermsForOrg_Success(t *testing.T) {
	mockDB, mock, _ := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	db, _ := gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{})
	db.SkipDefaultTransaction = true
	svc, err := New(db, iam.PasswordService{})
	if err != nil {
		t.Fatal(err)
	}
	seed, err := loadAccountingSeed()
	if err != nil {
		t.Fatal(err)
	}

	emptyIDs := func() *sqlmock.Rows { return sqlmock.NewRows([]string{"id"}) }
	idRow := func() *sqlmock.Rows { return sqlmock.NewRows([]string{"id"}).AddRow(1) }
	for range seed.PaymentTerms {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_terms"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms"`)).WillReturnRows(emptyIDs())
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_terms"`)).WillReturnRows(idRow())
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines"`)).WillReturnRows(emptyIDs())
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines"`)).WillReturnRows(emptyIDs())
	}
	for range []int{1, 1, 1, 1, 1} {
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_term_lines"`)).WillReturnRows(idRow())
	}

	if err := svc.seedPaymentTermsForOrg(t.Context(), 10, seed); err != nil {
		t.Errorf("seedPaymentTermsForOrg = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}
