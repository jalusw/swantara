package accounting

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestTaxPeriodService_Create_RejectsInvalidState(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	_, err := svc.Create(ctx, &TaxPeriod{OrganizationID: 1, TaxYearID: 1, Name: "Jan", DateStart: &start, DateEnd: &end, State: "sealed"})
	if helper.AssertError(t, err, true, ErrInvalidPeriodState) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Create_RejectsInvalidRange(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	_, err := svc.Create(ctx, &TaxPeriod{OrganizationID: 1, TaxYearID: 1, Name: "Feb", DateStart: &start, DateEnd: &end})
	if helper.AssertError(t, err, true, ErrInvalidPeriod) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Create_RejectsMissingTaxYear(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(99, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	_, err := svc.Create(ctx, &TaxPeriod{OrganizationID: 1, TaxYearID: 99, Name: "Jan", DateStart: &start, DateEnd: &end})
	if helper.AssertError(t, err, true, reference.ErrTaxYearNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Create_CreatesOpenPeriod(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	now := time.Now()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, "FY2026", &start, &end, nil, now, now, nil))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tax_periods"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 1, "Jan", &start, &end, "open", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	created, err := svc.Create(ctx, &TaxPeriod{OrganizationID: 1, TaxYearID: 1, Name: "Jan", DateStart: &start, DateEnd: &end})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}
	if created.State != TaxPeriodStateOpen {
		t.Errorf("state = %s, want open", created.State)
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Close_RejectsLocked(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE id = $1 ORDER BY "tax_periods"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_year_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, 1, "Jan", nil, nil, "locked", now, now, nil))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	_, err := svc.Close(ctx, 1)
	if helper.AssertError(t, err, true, ErrPeriodLocked) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Close_ClosesOpenPeriod(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE id = $1 ORDER BY "tax_periods"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_year_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, 1, "Jan", nil, nil, "open", now, now, nil))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tax_periods" SET`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 1, "Jan", nil, nil, "closed", sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	closed, err := svc.Close(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if closed.State != TaxPeriodStateClosed {
		t.Errorf("state = %s, want closed", closed.State)
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Lock_RejectsOpen(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE id = $1 ORDER BY "tax_periods"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_year_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, 1, "Jan", nil, nil, "open", now, now, nil))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	_, err := svc.Lock(ctx, 1)
	if helper.AssertError(t, err, true, ErrPeriodNotClosed) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_AssertOpen_RejectsClosed(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE organization_id = $1 AND date_start <= $2 AND date_end >= $3 AND deleted_at IS NULL ORDER BY date_start ASC`)).
		WithArgs(1, date, date, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_year_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, 1, "Jan", nil, nil, "closed", now, now, nil))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	err := svc.AssertOpen(ctx, 1, date)
	if helper.AssertError(t, err, true, ErrPeriodNotClosed) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_AssertOpen_AllowsOpenOrNoPeriod(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE organization_id = $1 AND date_start <= $2 AND date_end >= $3 AND deleted_at IS NULL ORDER BY date_start ASC`)).
		WithArgs(1, date, date, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_year_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, 1, "Jan", nil, nil, "open", now, now, nil))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	if err := svc.AssertOpen(ctx, 1, date); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPostingService_Post_RejectsClosedPeriod(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	periods := PeriodLookupMock{AssertOpenErr: ErrPeriodLocked}
	svc := NewPostingService(JournalEntryDAOMock{}).SetPeriods(periods)

	_, err := svc.Post(ctx, PostRequest{
		OrganizationID: 1,
		JournalID:      1,
		Date:           date,
		Lines: []PostingLine{
			{AccountID: 1300, Name: "Inventory", Debit: amount.FromFloat64(100)},
			{AccountID: 1350, Name: "Goods Received Not Billed", Credit: amount.FromFloat64(100)},
		},
	})
	if helper.AssertError(t, err, true, ErrPeriodLocked) {
		return
	}
}

func TestPostingService_Post_PeriodOpenAndBalanced(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	periods := PeriodLookupMock{}
	var captured *JournalEntry
	movements := JournalEntryDAOMock{
		CreateWithLinesFunc: func(_ context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
			captured = entry
			entry.ID = 1
			return entry, nil
		},
	}
	svc := NewPostingService(movements).SetPeriods(periods)

	entry, err := svc.Post(ctx, PostRequest{
		OrganizationID: 1,
		JournalID:      1,
		Date:           date,
		Lines: []PostingLine{
			{AccountID: 1300, Name: "Inventory", Debit: amount.FromFloat64(100)},
			{AccountID: 1350, Name: "Goods Received Not Billed", Credit: amount.FromFloat64(100)},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ID != 1 || captured.State != EntryStatePosted {
		t.Errorf("entry id/state = %d/%s, want 1/posted", entry.ID, captured.State)
	}
}

func TestTaxPeriodService_Open_OpensClosedPeriod(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE id = $1 ORDER BY "tax_periods"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_year_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, 1, "Jan", nil, nil, "closed", now, now, nil))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tax_periods" SET`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 1, "Jan", nil, nil, "open", sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	opened, err := svc.Open(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opened.State != TaxPeriodStateOpen {
		t.Errorf("state = %s, want open", opened.State)
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Open_RejectsLocked(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE id = $1 ORDER BY "tax_periods"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_year_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, 1, "Jan", nil, nil, "locked", now, now, nil))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	_, err := svc.Open(ctx, 1)
	if helper.AssertError(t, err, true, ErrPeriodLocked) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_StateTransitions_NotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE id = $1 ORDER BY "tax_periods"."id" LIMIT $2`)).
		WithArgs(99, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	_, err := svc.Open(ctx, 99)
	if helper.AssertError(t, err, true, ErrPeriodNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Close_NotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE id = $1 ORDER BY "tax_periods"."id" LIMIT $2`)).
		WithArgs(99, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	_, err := svc.Close(ctx, 99)
	if helper.AssertError(t, err, true, ErrPeriodNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Lock_NotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE id = $1 ORDER BY "tax_periods"."id" LIMIT $2`)).
		WithArgs(99, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	_, err := svc.Lock(ctx, 99)
	if helper.AssertError(t, err, true, ErrPeriodNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Lock_LocksClosedPeriod(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE id = $1 ORDER BY "tax_periods"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_year_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, 1, "Jan", nil, nil, "closed", now, now, nil))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tax_periods" SET`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), 1, 1, "Jan", nil, nil, "locked", sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	locked, err := svc.Lock(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if locked.State != TaxPeriodStateLocked {
		t.Errorf("state = %s, want locked", locked.State)
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_AssertOpen_NoPeriod(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE organization_id = $1 AND date_start <= $2 AND date_end >= $3 AND deleted_at IS NULL ORDER BY date_start ASC`)).
		WithArgs(1, date, date, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	err := svc.AssertOpen(ctx, 1, date)
	if helper.AssertError(t, err, true, ErrPeriodNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodDAO_ListByOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "tax_periods" WHERE organization_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE organization_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_year_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 7, 1, "Jan", nil, nil, "open", now, now, nil).
			AddRow(2, 7, 1, "Feb", nil, nil, "closed", now, now, nil))

	periods := NewTaxPeriodDAO(db)

	items, err := periods.ListByOrganization(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].Name != "Jan" || items[1].State != TaxPeriodStateClosed {
		t.Errorf("items = %+v, want Jan/Feb", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodDAO_ListByOrganization_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "tax_periods" WHERE organization_id = $1`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	periods := NewTaxPeriodDAO(db)

	_, err := periods.ListByOrganization(ctx, 7)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

type PeriodLookupMock struct {
	AssertOpenErr error
}

func (m PeriodLookupMock) AssertOpen(ctx context.Context, organizationID uint64, date time.Time) error {
	return m.AssertOpenErr
}

func TestTaxPeriodService_Close_PropagatesUpdateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE id = $1 ORDER BY "tax_periods"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_year_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, 1, "Jan", nil, nil, "open", now, now, nil))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tax_periods" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	_, err := svc.Close(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Lock_PropagatesUpdateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE id = $1 ORDER BY "tax_periods"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_year_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, 1, "Jan", nil, nil, "closed", now, now, nil))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tax_periods" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	_, err := svc.Lock(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Open_PropagatesUpdateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE id = $1 ORDER BY "tax_periods"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_year_id", "name", "date_start", "date_end", "state", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, 1, "Jan", nil, nil, "closed", now, now, nil))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tax_periods" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	_, err := svc.Open(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_AssertOpen_PropagatesFindByDateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE organization_id = $1 AND date_start <= $2 AND date_end >= $3 AND deleted_at IS NULL ORDER BY date_start ASC`)).
		WithArgs(1, date, date, 1).
		WillReturnError(errors.New("db down"))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	err := svc.AssertOpen(ctx, 1, date)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Create_PropagatesTaxYearError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	_, err := svc.Create(ctx, &TaxPeriod{OrganizationID: 1, TaxYearID: 1, Name: "Jan", DateStart: &start, DateEnd: &end})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodService_Create_RejectsYearOfAnotherOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id"}).AddRow(1, 99))

	svc := NewTaxPeriodService(NewTaxPeriodDAO(db), dao.NewBase[reference.TaxYear](db))

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	_, err := svc.Create(ctx, &TaxPeriod{OrganizationID: 1, TaxYearID: 1, Name: "Jan", DateStart: &start, DateEnd: &end})
	if helper.AssertError(t, err, true, reference.ErrTaxYearNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxPeriodDAO_FindByDate_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_periods" WHERE organization_id = $1 AND date_start <= $2 AND date_end >= $3 AND deleted_at IS NULL ORDER BY date_start ASC`)).
		WithArgs(1, date, date, 1).
		WillReturnError(errors.New("db down"))

	periods := NewTaxPeriodDAO(db)

	_, err := periods.FindByDate(ctx, 1, date)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
