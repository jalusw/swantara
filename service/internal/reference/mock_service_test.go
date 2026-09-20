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

func TestPaymentTermService_Create_UsesMock(t *testing.T) {
	ctx := context.Background()
	svc := NewPaymentTermService(PaymentTermDAOMock{
		CRUDMock: dao.CRUDMock[PaymentTerm]{
			ListFunc: func(ctx context.Context, q *query.Query) (*query.Page[PaymentTerm], error) {
				return &query.Page[PaymentTerm]{Items: []*PaymentTerm{}, Count: 0}, nil
			},
			CreateFunc: func(ctx context.Context, term *PaymentTerm) (*PaymentTerm, error) {
				return &PaymentTerm{Base: model.Base{ID: 1}, Name: term.Name, OrganizationID: term.OrganizationID}, nil
			},
		},
	})

	created, err := svc.Create(ctx, &PaymentTerm{OrganizationID: 1, Name: "Net 30"}, []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}
}

func TestPaymentTermService_Create_UsesMockReplaceLines(t *testing.T) {
	ctx := context.Background()
	svc := NewPaymentTermService(PaymentTermDAOMock{
		CRUDMock: dao.CRUDMock[PaymentTerm]{
			ListFunc: func(ctx context.Context, q *query.Query) (*query.Page[PaymentTerm], error) {
				return &query.Page[PaymentTerm]{Items: []*PaymentTerm{}, Count: 0}, nil
			},
			CreateFunc: func(ctx context.Context, term *PaymentTerm) (*PaymentTerm, error) {
				return &PaymentTerm{Base: model.Base{ID: 1}, Name: term.Name, OrganizationID: term.OrganizationID}, nil
			},
		},
		ReplaceLinesFunc: func(ctx context.Context, termID uint64, lines []*PaymentTermLine) error {
			return nil
		},
	})

	created, err := svc.Create(ctx, &PaymentTerm{OrganizationID: 1, Name: "Net 30"}, []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}
}

func TestPaymentTermService_Delete_UsesMock(t *testing.T) {
	ctx := context.Background()
	svc := NewPaymentTermService(PaymentTermDAOMock{})

	if err := svc.Delete(ctx, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPaymentTermService_Delete_UsesMockWithFunc(t *testing.T) {
	ctx := context.Background()
	svc := NewPaymentTermService(PaymentTermDAOMock{
		DeleteWithLinesFunc: func(ctx context.Context, termID uint64) error {
			return nil
		},
	})

	if err := svc.Delete(ctx, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPaymentTermService_Splits_UsesMockDefaultListLines(t *testing.T) {
	ctx := context.Background()
	svc := NewPaymentTermService(PaymentTermDAOMock{
		CRUDMock: dao.CRUDMock[PaymentTerm]{
			FindFunc: func(ctx context.Context, id uint64) (*PaymentTerm, error) {
				return &PaymentTerm{Base: model.Base{ID: 1}}, nil
			},
		},
	})

	_, err := svc.Splits(ctx, 1, amount.FromFloat64(1000), time.Now())
	if helper.AssertError(t, err, true, ErrInvalidTermLines) {
		return
	}
}

func TestPaymentTermService_Splits_ComputesFixedLineWithDayOfMonth(t *testing.T) {
	ctx := context.Background()
	svc := NewPaymentTermService(PaymentTermDAOMock{
		CRUDMock: dao.CRUDMock[PaymentTerm]{
			FindFunc: func(ctx context.Context, id uint64) (*PaymentTerm, error) {
				return &PaymentTerm{Base: model.Base{ID: 1}}, nil
			},
		},
		ListLinesFunc: func(ctx context.Context, termID uint64) ([]*PaymentTermLine, error) {
			return []*PaymentTermLine{
				{Sequence: 10, ValueType: "percent", Value: 100},
				{Sequence: 20, ValueType: "fixed", Value: 250, DayOfMonth: helper.Ptr(5)},
			}, nil
		},
	})

	date := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	splits, err := svc.Splits(ctx, 1, amount.FromFloat64(1000), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(splits) != 2 {
		t.Fatalf("splits = %d, want 2", len(splits))
	}
	if !splits[1].Amount.Equal(amount.FromFloat64(250)) {
		t.Errorf("split amount = %s, want 250", splits[1].Amount)
	}
	if splits[1].DueDate.Day() != 5 || splits[1].DueDate.Month() != time.February {
		t.Errorf("due date = %s, want February 5", splits[1].DueDate)
	}
}

func TestCarrierService_Update_RejectsEmptyName(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	svc := NewCarrierService(dao.NewBase[Carrier](db), VariantExistenceMock{})

	_, err := svc.Update(ctx, &Carrier{Name: "  "})
	if helper.AssertError(t, err, true, ErrCarrierName) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestCarrierService_Update_PropagatesVariantError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	deliveryItemID := uint64(7)

	svc := NewCarrierService(dao.NewBase[Carrier](db), VariantExistenceMock{
		VariantExistsFunc: func(ctx context.Context, id uint64) (bool, error) {
			return false, errors.New("db down")
		},
	})

	_, err := svc.Update(ctx, &Carrier{Name: "DHL", DeliveryItemID: &deliveryItemID})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestCarrierService_Update_AllowsNilVariantsChecker(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	deliveryItemID := uint64(7)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "carriers" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewCarrierService(dao.NewBase[Carrier](db), nil)

	updated, err := svc.Update(ctx, &Carrier{Base: model.Base{ID: 1}, Name: "DHL", DeliveryItemID: &deliveryItemID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != 1 {
		t.Errorf("id = %d, want 1", updated.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestCarrierService_Update_DefaultVariantCheckPasses(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	deliveryItemID := uint64(7)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "carriers" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	svc := NewCarrierService(dao.NewBase[Carrier](db), VariantExistenceMock{})

	updated, err := svc.Update(ctx, &Carrier{Base: model.Base{ID: 1}, Name: "DHL", DeliveryItemID: &deliveryItemID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != 1 {
		t.Errorf("id = %d, want 1", updated.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRateSource_PropagatesListError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	date := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_rates" WHERE currency_code = $1 AND rate_type = $2 AND valid_from <= $3 AND organization_id = $4`)).
		WithArgs("USD", "spot", date, 5).
		WillReturnError(errors.New("db down"))

	source := NewFxRateSource(dao.NewBase[FxRate](db))

	_, err := source.Rate(context.Background(), "USD", 5, amount.RateSpot, date)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
