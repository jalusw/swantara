package reference

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
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestFxRateService_Update_DefaultsRateTypeToSpot(t *testing.T) {
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

	updated, err := svc.Update(ctx, &FxRate{Base: model.Base{ID: 1}, CurrencyCode: "USD", Rate: 1.2, ValidFrom: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.RateType != "spot" {
		t.Errorf("rate type = %q, want spot", updated.RateType)
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRateService_Update_RejectsUnknownRateType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewFxRateService(dao.NewBase[FxRate](db), dao.NewBase[Currency](db))

	_, err := svc.Update(ctx, &FxRate{Rate: 1.2, RateType: "monthly", ValidFrom: time.Now()})
	if helper.AssertError(t, err, true, amount.ErrInvalidRate) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRateService_Update_RequiresValidFrom(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewFxRateService(dao.NewBase[FxRate](db), dao.NewBase[Currency](db))

	_, err := svc.Update(ctx, &FxRate{Rate: 1.2, RateType: "spot"})
	if helper.AssertError(t, err, true, ErrRateValidFromMissing) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRateService_Update_PropagatesCurrencyError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("USD", 1).
		WillReturnError(errors.New("db down"))

	svc := NewFxRateService(dao.NewBase[FxRate](db), dao.NewBase[Currency](db))

	_, err := svc.Update(ctx, &FxRate{Rate: 1.2, RateType: "spot", ValidFrom: time.Now(), CurrencyCode: "USD"})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRateService_Create_PropagatesCurrencyError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("USD", 1).
		WillReturnError(errors.New("db down"))

	svc := NewFxRateService(dao.NewBase[FxRate](db), dao.NewBase[Currency](db))

	_, err := svc.Create(ctx, &FxRate{CurrencyCode: "USD", Rate: 1.2, RateType: "spot", ValidFrom: time.Now()})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_CreateUnit_CreatesUnit(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at", "deleted_at"}).AddRow(1, "Weight", now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE name = $1 LIMIT $2`)).
		WithArgs("Piece", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "units"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	svc := NewUnitService(dao.NewBase[UnitGroup](db), dao.NewBase[Unit](db))

	created, err := svc.CreateUnit(ctx, &Unit{CategoryID: 1, Name: "Piece", Factor: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_CreateUnit_PropagatesCategoryError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))

	svc := NewUnitService(dao.NewBase[UnitGroup](db), dao.NewBase[Unit](db))

	_, err := svc.CreateUnit(ctx, &Unit{CategoryID: 1, Name: "Piece", Factor: 1})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_CreateUnit_PropagatesSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at", "deleted_at"}).AddRow(1, "Weight", now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE name = $1 LIMIT $2`)).
		WithArgs("Piece", 1).
		WillReturnError(errors.New("db down"))

	svc := NewUnitService(dao.NewBase[UnitGroup](db), dao.NewBase[Unit](db))

	_, err := svc.CreateUnit(ctx, &Unit{CategoryID: 1, Name: "Piece", Factor: 1})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_Convert_PropagatesFromFindError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))

	svc := NewUnitService(dao.NewBase[UnitGroup](db), dao.NewBase[Unit](db))

	_, err := svc.Convert(ctx, amount.FromFloat64(2), 1, 2)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_Convert_RejectsMissingFromUnit(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewUnitService(dao.NewBase[UnitGroup](db), dao.NewBase[Unit](db))

	_, err := svc.Convert(ctx, amount.FromFloat64(2), 1, 2)
	if helper.AssertError(t, err, true, ErrUnitNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_Convert_PropagatesToFindError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	uomColumns := []string{"id", "category_id", "name", "factor", "unit_type", "rounding", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).AddRow(1, 1, "kg", 1, "", 0.001, now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(2, 1).
		WillReturnError(errors.New("db down"))

	svc := NewUnitService(dao.NewBase[UnitGroup](db), dao.NewBase[Unit](db))

	_, err := svc.Convert(ctx, amount.FromFloat64(2), 1, 2)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_Convert_RejectsMissingToUnit(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	uomColumns := []string{"id", "category_id", "name", "factor", "unit_type", "rounding", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).AddRow(1, 1, "kg", 1, "", 0.001, now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewUnitService(dao.NewBase[UnitGroup](db), dao.NewBase[Unit](db))

	_, err := svc.Convert(ctx, amount.FromFloat64(2), 1, 2)
	if helper.AssertError(t, err, true, ErrUnitNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_Convert_RejectsNonPositiveFactor(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	uomColumns := []string{"id", "category_id", "name", "factor", "unit_type", "rounding", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).AddRow(1, 1, "kg", 0, "", 0.001, now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).AddRow(2, 1, "g", 0.001, "", 0.001, now, now, nil))

	svc := NewUnitService(dao.NewBase[UnitGroup](db), dao.NewBase[Unit](db))

	_, err := svc.Convert(ctx, amount.FromFloat64(2), 1, 2)
	if helper.AssertError(t, err, true, ErrInvalidFactor) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_Convert_RoundsToTwoWhenRoundingUnset(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	uomColumns := []string{"id", "category_id", "name", "factor", "unit_type", "rounding", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).AddRow(1, 1, "kg", 1, "", 0.001, now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).AddRow(2, 1, "g", 0.001, "", 0, now, now, nil))

	svc := NewUnitService(dao.NewBase[UnitGroup](db), dao.NewBase[Unit](db))

	converted, err := svc.Convert(ctx, amount.FromFloat64(2), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !converted.Equal(amount.FromFloat64(2000)) {
		t.Errorf("converted = %s, want 2000", converted)
	}

	query.AssertDBMockDone(t, mock)
}

func TestUnitService_Convert_ClampsNegativePrecision(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	uomColumns := []string{"id", "category_id", "name", "factor", "unit_type", "rounding", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).AddRow(1, 1, "kg", 1, "", 0.001, now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).AddRow(2, 1, "g", 0.001, "", 10, now, now, nil))

	svc := NewUnitService(dao.NewBase[UnitGroup](db), dao.NewBase[Unit](db))

	converted, err := svc.Convert(ctx, amount.FromFloat64(2), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !converted.Equal(amount.FromFloat64(2000)) {
		t.Errorf("converted = %s, want 2000", converted)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Create_CreatesTermWithLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_terms" WHERE organization_id = $1 AND name = $2`)).
		WithArgs(1, "Net 30").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE organization_id = $1 AND name = $2`)).
		WithArgs(1, "Net 30").
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "note", "code", "is_active", "template_key", "created_at", "updated_at", "deleted_at"}))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_terms"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_term_id"}))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_term_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(8))
	mock.ExpectCommit()

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	created, err := svc.Create(ctx, &PaymentTerm{OrganizationID: 1, Name: "Net 30"}, []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Create_RejectsInvalidLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	_, err := svc.Create(ctx, &PaymentTerm{Name: "Net 30"}, nil)
	if helper.AssertError(t, err, true, ErrInvalidTermLines) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Create_PropagatesTermCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_terms" WHERE organization_id = $1 AND name = $2`)).
		WithArgs(1, "Net 30").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE organization_id = $1 AND name = $2`)).
		WithArgs(1, "Net 30").
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "note", "code", "is_active", "template_key", "created_at", "updated_at", "deleted_at"}))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_terms"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	_, err := svc.Create(ctx, &PaymentTerm{OrganizationID: 1, Name: "Net 30"}, []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Create_PropagatesReplaceLinesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_terms" WHERE organization_id = $1 AND name = $2`)).
		WithArgs(1, "Net 30").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE organization_id = $1 AND name = $2`)).
		WithArgs(1, "Net 30").
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "note", "code", "is_active", "template_key", "created_at", "updated_at", "deleted_at"}))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_terms"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	_, err := svc.Create(ctx, &PaymentTerm{OrganizationID: 1, Name: "Net 30"}, []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Update_PropagatesFindError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE id = $1 ORDER BY "payment_terms"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	_, err := svc.Update(ctx, 1, &PaymentTerm{Name: "Net 30"}, []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Update_RejectsMissingTerm(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE id = $1 ORDER BY "payment_terms"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	_, err := svc.Update(ctx, 1, &PaymentTerm{Name: "Net 30"}, []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100}})
	if helper.AssertError(t, err, true, ErrTermNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Update_PropagatesTermUpdateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	termColumns := []string{"id", "organization_id", "name", "note", "code", "is_active", "template_key", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE id = $1 ORDER BY "payment_terms"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(termColumns).AddRow(1, 1, "Net 30", nil, nil, true, nil, now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_terms" WHERE organization_id = $1 AND name = $2`)).
		WithArgs(1, "Net 30").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE organization_id = $1 AND name = $2`)).
		WithArgs(1, "Net 30").
		WillReturnRows(sqlmock.NewRows(termColumns).AddRow(1, 1, "Net 30", nil, nil, true, nil, now, now, nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payment_terms" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	_, err := svc.Update(ctx, 1, &PaymentTerm{Name: "Net 30"}, []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Update_PropagatesReplaceLinesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	termColumns := []string{"id", "organization_id", "name", "note", "code", "is_active", "template_key", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE id = $1 ORDER BY "payment_terms"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(termColumns).AddRow(1, 1, "Net 30", nil, nil, true, nil, now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_terms" WHERE organization_id = $1 AND name = $2`)).
		WithArgs(1, "Net 30").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE organization_id = $1 AND name = $2`)).
		WithArgs(1, "Net 30").
		WillReturnRows(sqlmock.NewRows(termColumns).AddRow(1, 1, "Net 30", nil, nil, true, nil, now, now, nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payment_terms" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	_, err := svc.Update(ctx, 1, &PaymentTerm{Name: "Net 30"}, []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Splits_PropagatesFindError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE id = $1 ORDER BY "payment_terms"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	_, err := svc.Splits(ctx, 1, amount.FromFloat64(1000), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Splits_RejectsMissingTerm(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE id = $1 ORDER BY "payment_terms"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	_, err := svc.Splits(ctx, 1, amount.FromFloat64(1000), time.Now())
	if helper.AssertError(t, err, true, ErrTermNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Splits_PropagatesListLinesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	termColumns := []string{"id", "name", "note", "created_at", "updated_at", "deleted_at"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_terms" WHERE id = $1 ORDER BY "payment_terms"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(termColumns).AddRow(1, "Net 30", nil, now, now, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))

	svc := NewPaymentTermService(NewPaymentTermDAO(db))

	_, err := svc.Splits(ctx, 1, amount.FromFloat64(1000), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermService_Splits_RejectsPercentOverHundredWithBalance(t *testing.T) {
	ctx := context.Background()
	svc := NewPaymentTermService(PaymentTermDAOMock{
		CRUDMock: dao.CRUDMock[PaymentTerm]{
			FindFunc: func(ctx context.Context, id uint64) (*PaymentTerm, error) {
				return &PaymentTerm{Base: model.Base{ID: 1}}, nil
			},
		},
		ListLinesFunc: func(ctx context.Context, termID uint64) ([]*PaymentTermLine, error) {
			return []*PaymentTermLine{
				{Sequence: 10, ValueType: "percent", Value: 110},
				{Sequence: 20, ValueType: "balance"},
			}, nil
		},
	})

	_, err := svc.Splits(ctx, 1, amount.FromFloat64(1000), time.Now())
	if helper.AssertError(t, err, true, ErrInvalidTermLines) {
		return
	}
}

func TestDimensionService_Create_PropagatesSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE code = $1 LIMIT $2`)).
		WithArgs("MGMT", 1).
		WillReturnError(errors.New("db down"))

	svc := NewDimensionService(dao.NewBase[Dimension](db))

	_, err := svc.Create(ctx, &Dimension{Name: "Admin", Code: ptr("MGMT")})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDimensionService_Create_PropagatesParentFindError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	parentID := uint64(9)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(9, 1).
		WillReturnError(errors.New("db down"))

	svc := NewDimensionService(dao.NewBase[Dimension](db))

	_, err := svc.Create(ctx, &Dimension{Name: "Child", ParentID: &parentID})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDimensionService_Create_RejectsDuplicateCodeWithoutOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE code = $1 LIMIT $2`)).
		WithArgs("OPS", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "code", "kind", "parent_id", "active", "created_at", "updated_at", "deleted_at"}).
			AddRow(2, nil, "Ops", "OPS", nil, nil, true, now, now, nil))

	svc := NewDimensionService(dao.NewBase[Dimension](db))

	_, err := svc.Create(ctx, &Dimension{Name: "Ops", Code: ptr("OPS")})
	if helper.AssertError(t, err, true, ErrDuplicateCode) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDimensionService_Create_IgnoresDuplicateCodeInOtherOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	now := time.Now()
	organizationID := uint64(5)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE code = $1 LIMIT $2`)).
		WithArgs("OPS", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "code", "kind", "parent_id", "active", "created_at", "updated_at", "deleted_at"}).
			AddRow(2, nil, "Ops", "OPS", nil, nil, true, now, now, nil))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "dimensions"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	svc := NewDimensionService(dao.NewBase[Dimension](db))

	created, err := svc.Create(ctx, &Dimension{OrganizationID: &organizationID, Name: "Ops", Code: ptr("OPS")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountService_Update_RejectsUnknownType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewAccountService(dao.NewBase[Account](db))

	_, err := svc.Update(ctx, &Account{Base: model.Base{ID: 1}, Code: "9000", Name: "Mystery", Type: "magic"})
	if helper.AssertError(t, err, true, ErrInvalidAccountType) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountService_Update_UpdatesAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("1110", 1).
		WillReturnRows(sqlmock.NewRows([]string{"code"}))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewAccountService(dao.NewBase[Account](db))

	updated, err := svc.Update(ctx, &Account{Base: model.Base{ID: 1}, OrganizationID: 1, Code: "1110", Name: "Bank", Type: "bank"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != 1 {
		t.Errorf("id = %d, want 1", updated.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountService_Create_PropagatesSearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("1110", 1).
		WillReturnError(errors.New("db down"))

	svc := NewAccountService(dao.NewBase[Account](db))

	_, err := svc.Create(ctx, &Account{OrganizationID: 1, Code: "1110", Name: "Bank", Type: "bank"})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountService_Create_PropagatesParentFindError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("1200", 1).
		WillReturnRows(sqlmock.NewRows([]string{"code"}))

	parentID := uint64(9)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(9, 1).
		WillReturnError(errors.New("db down"))

	svc := NewAccountService(dao.NewBase[Account](db))

	_, err := svc.Create(ctx, &Account{OrganizationID: 1, Code: "1200", Name: "Child", Type: "asset", ParentID: &parentID})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalService_Update_RejectsUnknownType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewJournalService(dao.NewBase[Journal](db), dao.NewBase[Account](db))

	_, err := svc.Update(ctx, &Journal{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Rent", Type: "rent"})
	if helper.AssertError(t, err, true, ErrInvalidJournalType) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalService_Update_UpdatesJournal(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "journals" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewJournalService(dao.NewBase[Journal](db), dao.NewBase[Account](db))

	updated, err := svc.Update(ctx, &Journal{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Sales", Type: "sale"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != 1 {
		t.Errorf("id = %d, want 1", updated.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalService_Create_PropagatesAccountFindError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	defaultAccountID := uint64(5)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnError(errors.New("db down"))

	svc := NewJournalService(dao.NewBase[Journal](db), dao.NewBase[Account](db))

	_, err := svc.Create(ctx, &Journal{OrganizationID: 1, Name: "Sales", Type: "sale", DefaultAccountID: &defaultAccountID})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxService_Update_RejectsUnknownType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewTaxService(dao.NewBase[Tax](db), dao.NewBase[Account](db))

	taxAmount := 10.0
	_, err := svc.Update(ctx, &Tax{Base: model.Base{ID: 1}, Name: "VAT", Amount: &taxAmount, Type: "percentage", Scope: "sale"})
	if helper.AssertError(t, err, true, ErrInvalidTaxType) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxService_Update_UpdatesTax(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "taxes" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewTaxService(dao.NewBase[Tax](db), dao.NewBase[Account](db))

	taxAmount := 10.0
	updated, err := svc.Update(ctx, &Tax{Base: model.Base{ID: 1}, Name: "VAT", Amount: &taxAmount, Type: "percent", Scope: "sale"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != 1 {
		t.Errorf("id = %d, want 1", updated.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxService_Create_RejectsUnknownScope(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewTaxService(dao.NewBase[Tax](db), dao.NewBase[Account](db))

	taxAmount := 10.0
	_, err := svc.Create(ctx, &Tax{Name: "VAT", Amount: &taxAmount, Type: "percent", Scope: "local"})
	if helper.AssertError(t, err, true, ErrInvalidTaxScope) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxService_Create_PropagatesRefundAccountFindError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	refundAccountID := uint64(9)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(9, 1).
		WillReturnError(errors.New("db down"))

	svc := NewTaxService(dao.NewBase[Tax](db), dao.NewBase[Account](db))

	taxAmount := 10.0
	_, err := svc.Create(ctx, &Tax{Name: "VAT", Amount: &taxAmount, Type: "percent", Scope: "sale", RefundTaxAccountID: &refundAccountID})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxYearService_Update_RejectsInvalidRange(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	svc := NewTaxYearService(dao.NewBase[TaxYear](db))

	_, err := svc.Update(ctx, &TaxYear{Base: model.Base{ID: 1}, Name: "FY2026", DateStart: &start, DateEnd: &end})
	if helper.AssertError(t, err, true, ErrInvalidTaxYear) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxYearService_Update_UpdatesTaxYear(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tax_years" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewTaxYearService(dao.NewBase[TaxYear](db))

	updated, err := svc.Update(ctx, &TaxYear{Base: model.Base{ID: 1}, Name: "FY2026", DateStart: &start, DateEnd: &end})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != 1 {
		t.Errorf("id = %d, want 1", updated.ID)
	}

	query.AssertDBMockDone(t, mock)
}
