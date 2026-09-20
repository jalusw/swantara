package reference

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestAccountService_Create_RejectsUnknownType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewAccountService(dao.NewBase[Account](db))

	_, err := svc.Create(ctx, &Account{Code: "9000", Name: "Mystery", Type: "magic"})
	if helper.AssertError(t, err, true, ErrInvalidAccountType) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountService_Create_RejectsDuplicateCode(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("1110", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "code", "name", "type", "reconcilable", "currency_code", "parent_id", "active", "created_at", "updated_at", "deleted_at"}).
			AddRow(3, 1, "1110", "Bank", "bank", true, nil, nil, true, now, now, nil))

	svc := NewAccountService(dao.NewBase[Account](db))

	_, err := svc.Create(ctx, &Account{OrganizationID: 1, Code: "1110", Name: "Bank", Type: "bank"})
	if helper.AssertError(t, err, true, ErrDuplicateAccount) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountService_Create_RejectsMissingParent(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("1200", 1).
		WillReturnRows(sqlmock.NewRows([]string{"code"}))

	parentID := uint64(9)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(9, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewAccountService(dao.NewBase[Account](db))

	_, err := svc.Create(ctx, &Account{OrganizationID: 1, Code: "1200", Name: "Child", Type: "asset", ParentID: &parentID})
	if helper.AssertError(t, err, true, ErrInvalidParent) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountService_Create_CreatesAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("4100", 1).
		WillReturnRows(sqlmock.NewRows([]string{"code"}))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, "4100", "Sales", "income", false, nil, nil, true).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	svc := NewAccountService(dao.NewBase[Account](db))

	created, err := svc.Create(ctx, &Account{OrganizationID: 1, Code: "4100", Name: "Sales", Type: "income"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalService_Create_RejectsUnknownType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewJournalService(
		dao.NewBase[Journal](db),
		dao.NewBase[Account](db),
	)

	_, err := svc.Create(ctx, &Journal{OrganizationID: 1, Name: "Rent", Type: "rent"})
	if helper.AssertError(t, err, true, ErrInvalidJournalType) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalService_Create_RejectsMissingDefaultAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	defaultAccountID := uint64(5)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewJournalService(
		dao.NewBase[Journal](db),
		dao.NewBase[Account](db),
	)

	_, err := svc.Create(ctx, &Journal{OrganizationID: 1, Name: "Sales", Type: "sale", DefaultAccountID: &defaultAccountID})
	if helper.AssertError(t, err, true, ErrJournalDefaultAccount) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalService_Create_CreatesJournal(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	code := "SALE"
	currency := "USD"
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "journals"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, "Sales", "SALE", "sale", nil, &currency, nil, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	svc := NewJournalService(
		dao.NewBase[Journal](db),
		dao.NewBase[Account](db),
	)

	created, err := svc.Create(ctx, &Journal{OrganizationID: 1, Code: &code, Name: "Sales", Type: "sale", CurrencyCode: &currency})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxService_Create_RejectsUnknownType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewTaxService(
		dao.NewBase[Tax](db),
		dao.NewBase[Account](db),
	)

	amount := 10.0
	_, err := svc.Create(ctx, &Tax{Name: "VAT", Amount: &amount, Type: "percentage", Scope: "sale"})
	if helper.AssertError(t, err, true, ErrInvalidTaxType) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxService_Create_RejectsMissingAmount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewTaxService(
		dao.NewBase[Tax](db),
		dao.NewBase[Account](db),
	)

	_, err := svc.Create(ctx, &Tax{Name: "VAT", Type: "percent", Scope: "sale"})
	if helper.AssertError(t, err, true, ErrTaxAmountMissing) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxService_Create_RejectsMissingTaxAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	taxAccountID := uint64(7)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(7, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewTaxService(
		dao.NewBase[Tax](db),
		dao.NewBase[Account](db),
	)

	amount := 10.0
	_, err := svc.Create(ctx, &Tax{Name: "VAT", Amount: &amount, Type: "percent", Scope: "sale", TaxAccountID: &taxAccountID})
	if helper.AssertError(t, err, true, ErrInvalidTaxAccount) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxService_Create_CreatesTax(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "taxes"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "VAT", 10.0, "percent", "sale", false, nil, nil, true).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	svc := NewTaxService(
		dao.NewBase[Tax](db),
		dao.NewBase[Account](db),
	)

	amount := 10.0
	created, err := svc.Create(ctx, &Tax{Name: "VAT", Amount: &amount, Type: "percent", Scope: "sale"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxYearService_Create_RejectsInvalidRange(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	svc := NewTaxYearService(dao.NewBase[TaxYear](db))

	_, err := svc.Create(ctx, &TaxYear{Name: "FY2026", DateStart: &start, DateEnd: &end})
	if helper.AssertError(t, err, true, ErrInvalidTaxYear) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxYearService_Create_CreatesTaxYear(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tax_years"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "FY2026", start, end, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	svc := NewTaxYearService(dao.NewBase[TaxYear](db))

	created, err := svc.Create(ctx, &TaxYear{Name: "FY2026", DateStart: &start, DateEnd: &end})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}
