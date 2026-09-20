package reference

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestFxRateService_Create_ValidatesRate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		rate     *FxRate
		wantErrV error
	}{
		{
			name:     "rejects non-positive rate",
			rate:     &FxRate{Rate: 0, RateType: "spot", ValidFrom: time.Now()},
			wantErrV: amount.ErrInvalidRate,
		},
		{
			name:     "rejects unknown rate type",
			rate:     &FxRate{Rate: 1.2, RateType: "monthly", ValidFrom: time.Now()},
			wantErrV: amount.ErrInvalidRate,
		},
		{
			name:     "requires valid_from",
			rate:     &FxRate{Rate: 1.2, RateType: "spot"},
			wantErrV: ErrRateValidFromMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			svc := NewFxRateService(dao.NewBase[FxRate](db), dao.NewBase[Currency](db))

			_, err := svc.Create(ctx, tt.rate)
			if helper.AssertError(t, err, true, tt.wantErrV) {
				return
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestFxRateService_Create_RejectsUnknownCurrency(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("XXX", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewFxRateService(dao.NewBase[FxRate](db), dao.NewBase[Currency](db))

	_, err := svc.Create(ctx, &FxRate{CurrencyCode: "XXX", Rate: 1.2, RateType: "spot", ValidFrom: time.Now()})
	if helper.AssertError(t, err, true, ErrCurrencyNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRateService_Create_CreatesWithSpotDefault(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("USD", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(1, "USD"))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "fx_rates"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "USD", sqlmock.AnyArg(), 1.2, "spot", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	svc := NewFxRateService(dao.NewBase[FxRate](db), dao.NewBase[Currency](db))

	created, err := svc.Create(ctx, &FxRate{CurrencyCode: "USD", Rate: 1.2, ValidFrom: time.Now()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.RateType != "spot" {
		t.Errorf("rate type = %q, want spot", created.RateType)
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_CreateUnit_Validates(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		unit     *Unit
		wantErrV error
	}{
		{
			name:     "rejects non-positive factor",
			unit:     &Unit{CategoryID: 1, Name: "Piece", Factor: 0},
			wantErrV: ErrInvalidFactor,
		},
		{
			name:     "rejects missing category",
			unit:     &Unit{CategoryID: 1, Name: "Piece", Factor: 1},
			wantErrV: ErrUnitGroupNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			svc := NewUnitService(
				dao.NewBase[UnitGroup](db),
				dao.NewBase[Unit](db),
			)

			if tt.wantErrV == ErrUnitGroupNotFound {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			}

			_, err := svc.CreateUnit(ctx, tt.unit)
			if helper.AssertError(t, err, true, tt.wantErrV) {
				return
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestUnitService_CreateUnit_RejectsDuplicateNameInCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE name = $1 LIMIT $2`)).
		WithArgs("Piece", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "name", "factor", "unit_type", "rounding", "created_at", "updated_at", "deleted_at"}).
			AddRow(2, 1, "Piece", 1, "", 0.01, time.Now(), time.Now(), nil))

	svc := NewUnitService(
		dao.NewBase[UnitGroup](db),
		dao.NewBase[Unit](db),
	)

	_, err := svc.CreateUnit(ctx, &Unit{CategoryID: 1, Name: "Piece", Factor: 1})
	if helper.AssertError(t, err, true, ErrUnitNameTaken) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_Convert_ConvertsAcrossFactors(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	uomColumns := []string{"id", "category_id", "name", "factor", "unit_type", "rounding", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).AddRow(1, 1, "kg", 1, "", 0.001, now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).AddRow(2, 1, "g", 0.001, "", 0.001, now, now, nil))

	svc := NewUnitService(
		dao.NewBase[UnitGroup](db),
		dao.NewBase[Unit](db),
	)

	converted, err := svc.Convert(ctx, amount.FromFloat64(2), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !converted.Equal(amount.FromFloat64(2000)) {
		t.Errorf("converted = %s, want 2000", converted)
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_Convert_RejectsCategoryMismatch(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	uomColumns := []string{"id", "category_id", "name", "factor", "unit_type", "rounding", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).AddRow(1, 1, "kg", 1, "", 0.001, now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).AddRow(2, 2, "h", 1, "", 0.001, now, now, nil))

	svc := NewUnitService(
		dao.NewBase[UnitGroup](db),
		dao.NewBase[Unit](db),
	)

	_, err := svc.Convert(ctx, amount.FromFloat64(2), 1, 2)
	if helper.AssertError(t, err, true, ErrUnitGroupMismatch) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Splits_ComputesPercentFixedBalance(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	termColumns := []string{"id", "name", "note", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE id = $1 ORDER BY "payment_terms"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(termColumns).AddRow(1, "30 60 90", nil, now, now, nil))

	lineColumns := []string{"id", "payment_term_id", "sequence", "value_type", "value", "days_after", "day_of_month", "discount_pct", "discount_days", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(lineColumns).
			AddRow(1, 1, 10, "percent", 50, 30, nil, nil, nil, now, now, nil).
			AddRow(2, 1, 20, "percent", 30, 60, nil, nil, nil, now, now, nil).
			AddRow(3, 1, 30, "balance", 0, 90, nil, nil, nil, now, now, nil))

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	date := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	splits, err := svc.Splits(ctx, 1, amount.FromFloat64(1000), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(splits) != 3 {
		t.Fatalf("splits = %d, want 3", len(splits))
	}
	if !splits[0].Amount.Equal(amount.FromFloat64(500)) {
		t.Errorf("split[0] = %s, want 500", splits[0].Amount)
	}
	if !splits[1].Amount.Equal(amount.FromFloat64(300)) {
		t.Errorf("split[1] = %s, want 300", splits[1].Amount)
	}
	if !splits[2].Amount.Equal(amount.FromFloat64(200)) {
		t.Errorf("split[2] = %s, want 200 (balance)", splits[2].Amount)
	}
	if splits[2].DueDate.Day() != 15 {
		t.Errorf("balance due day = %d, want 15 (90 days after)", splits[2].DueDate.Day())
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Splits_ValidatesLines(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		lines    []*PaymentTermLine
		wantErrV error
	}{
		{
			name:     "rejects empty lines",
			lines:    nil,
			wantErrV: ErrInvalidTermLines,
		},
		{
			name:     "rejects percent sum not equal to 100",
			lines:    []*PaymentTermLine{{ValueType: "percent", Value: 60}, {ValueType: "percent", Value: 30}},
			wantErrV: ErrInvalidTermLines,
		},
		{
			name:     "rejects multiple balance lines",
			lines:    []*PaymentTermLine{{ValueType: "percent", Value: 100}, {ValueType: "balance"}, {ValueType: "balance"}},
			wantErrV: ErrMultipleBalanceLines,
		},
		{
			name:     "rejects unknown value type",
			lines:    []*PaymentTermLine{{ValueType: "monthly", Value: 100}},
			wantErrV: ErrInvalidTermLines,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)

			now := time.Now()
			termColumns := []string{"id", "name", "note", "created_at", "updated_at", "deleted_at"}
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE id = $1 ORDER BY "payment_terms"."id" LIMIT $2`)).
				WithArgs(1, 1).
				WillReturnRows(sqlmock.NewRows(termColumns).AddRow(1, "term", nil, now, now, nil))

			lineColumns := []string{"id", "payment_term_id", "sequence", "value_type", "value", "days_after", "day_of_month", "discount_pct", "discount_days", "created_at", "updated_at", "deleted_at"}
			rows := sqlmock.NewRows(lineColumns)
			for i, line := range tt.lines {
				rows.AddRow(i+1, 1, (i+1)*10, line.ValueType, line.Value, 0, nil, nil, nil, now, now, nil)
			}
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
				WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(len(tt.lines)))
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines" WHERE payment_term_id = $1`)).
				WithArgs(1).
				WillReturnRows(rows)

			svc := NewPaymentTermService(NewPaymentTermDAO(db))

			_, err := svc.Splits(ctx, 1, amount.FromFloat64(1000), time.Now())
			if helper.AssertError(t, err, true, tt.wantErrV) {
				return
			}

			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestDimensionService_Create_RejectsDuplicateCode(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	organizationID := uint64(5)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE code = $1 LIMIT $2`)).
		WithArgs("MGMT", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "code", "kind", "parent_id", "active", "created_at", "updated_at", "deleted_at"}).
			AddRow(2, 5, "Management", "MGMT", "general", nil, true, now, now, nil))

	svc := NewDimensionService(dao.NewBase[Dimension](db))

	_, err := svc.Create(ctx, &Dimension{OrganizationID: &organizationID, Name: "Admin", Code: ptr("MGMT")})
	if helper.AssertError(t, err, true, ErrDuplicateCode) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDimensionService_Create_RejectsMissingParent(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	parentID := uint64(9)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(9, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewDimensionService(dao.NewBase[Dimension](db))

	_, err := svc.Create(ctx, &Dimension{Name: "Child", ParentID: &parentID})
	if helper.AssertError(t, err, true, ErrParentNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDimensionService_Create_CreatesAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "dimensions"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "Operations", nil, nil, nil, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	svc := NewDimensionService(dao.NewBase[Dimension](db))

	created, err := svc.Create(ctx, &Dimension{Name: "Operations"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func ptr[T any](v T) *T {
	return &v
}

func TestFxRateService_Update_ValidatesAndPersists(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("USD", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(1, "USD"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "fx_rates" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewFxRateService(dao.NewBase[FxRate](db), dao.NewBase[Currency](db))

	_, err := svc.Update(ctx, &FxRate{Base: model.Base{ID: 1}, CurrencyCode: "USD", Rate: 1.2, RateType: "spot", ValidFrom: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRateService_Update_RejectsInvalidRate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewFxRateService(dao.NewBase[FxRate](db), dao.NewBase[Currency](db))

	_, err := svc.Update(ctx, &FxRate{Rate: 0})
	if helper.AssertError(t, err, true, amount.ErrInvalidRate) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_UpdateUnit_RejectsDuplicateNameExcludingSelf(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at", "deleted_at"}).AddRow(1, "Weight", now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE name = $1 LIMIT $2`)).
		WithArgs("Gram", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "name", "factor"}).AddRow(2, 1, "Gram", 0.001))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "units" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewUnitService(dao.NewBase[UnitGroup](db), dao.NewBase[Unit](db))

	updated, err := svc.UpdateUnit(ctx, &Unit{Base: model.Base{ID: 2}, CategoryID: 1, Name: "Gram", Factor: 0.001})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != 2 {
		t.Errorf("id = %d, want 2", updated.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_UpdateUnit_RejectsNameTakenByOtherCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at", "deleted_at"}).AddRow(1, "Weight", now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE name = $1 LIMIT $2`)).
		WithArgs("Gram", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "category_id", "name", "factor"}).AddRow(3, 1, "Gram", 0.001))

	svc := NewUnitService(dao.NewBase[UnitGroup](db), dao.NewBase[Unit](db))

	_, err := svc.UpdateUnit(ctx, &Unit{Base: model.Base{ID: 2}, CategoryID: 1, Name: "Gram", Factor: 0.001})
	if helper.AssertError(t, err, true, ErrUnitNameTaken) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDimensionService_Update_RejectsDuplicateCodeExcludingSelf(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	organizationID := uint64(5)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE code = $1 LIMIT $2`)).
		WithArgs("MGMT", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "code", "kind", "parent_id", "active", "created_at", "updated_at", "deleted_at"}).
			AddRow(2, 5, "Management", "MGMT", "general", nil, true, now, now, nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "dimensions" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewDimensionService(dao.NewBase[Dimension](db))

	updated, err := svc.Update(ctx, &Dimension{Base: model.Base{ID: 2}, OrganizationID: &organizationID, Name: "Management", Code: ptr("MGMT")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != 2 {
		t.Errorf("id = %d, want 2", updated.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestDimensionService_Update_RejectsMissingParent(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	parentID := uint64(9)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(9, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewDimensionService(dao.NewBase[Dimension](db))

	_, err := svc.Update(ctx, &Dimension{Base: model.Base{ID: 2}, Name: "Child", ParentID: &parentID})
	if helper.AssertError(t, err, true, ErrParentNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Update_ReplacesLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE id = $1 ORDER BY "payment_terms"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "note", "code", "is_active", "template_key", "created_at", "updated_at", "deleted_at"}).AddRow(1, 1, "Net 30", nil, nil, true, nil, now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_terms" WHERE organization_id = $1 AND name = $2`)).
		WithArgs(1, "Net 30").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE organization_id = $1 AND name = $2`)).
		WithArgs(1, "Net 30").
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "note", "code", "is_active", "template_key", "created_at", "updated_at", "deleted_at"}).AddRow(1, 1, "Net 30", nil, nil, true, nil, now, now, nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payment_terms" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_term_id", "sequence", "value_type", "value"}).AddRow(7, 1, 10, "percent", 100))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "payment_term_lines"`)).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_term_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(8))
	mock.ExpectCommit()

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	created, err := svc.Update(ctx, 1, &PaymentTerm{Name: "Net 30"}, []*PaymentTermLine{
		{Sequence: 10, ValueType: "percent", Value: 100},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Update_RejectsInvalidLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	_, err := svc.Update(ctx, 1, &PaymentTerm{Name: "Net 30"}, nil)
	if helper.AssertError(t, err, true, ErrInvalidTermLines) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Delete_RemovesLinesAndTerm(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_term_id"}).AddRow(7, 1))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "payment_term_lines"`)).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "payment_terms"`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	if err := svc.Delete(ctx, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}
