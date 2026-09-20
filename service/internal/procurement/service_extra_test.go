package procurement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestCurrencyConverter_Convert_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("reverse lookup error propagates", func(t *testing.T) {
		calls := 0
		rates := &CurrencyRateDAOMock{
			FindByPairDateFunc: func(_ context.Context, _, _ string, _ *uint64, _ time.Time) (*CurrencyRate, error) {
				calls++
				if calls == 1 {
					return nil, nil
				}
				return nil, dbErr
			},
		}
		_, err := NewCurrencyConverter(rates).Convert(ctx, amount.FromFloat64(100), "USD", "IDR", nil, time.Now())
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("direct zero rate invalid", func(t *testing.T) {
		rates := &CurrencyRateDAOMock{
			FindByPairDateFunc: func(_ context.Context, _, _ string, _ *uint64, _ time.Time) (*CurrencyRate, error) {
				return &CurrencyRate{Rate: 0}, nil
			},
		}
		_, err := NewCurrencyConverter(rates).Convert(ctx, amount.FromFloat64(100), "USD", "IDR", nil, time.Now())
		helper.AssertError(t, err, true, ErrCurrencyRateInvalid)
	})

	t.Run("reverse zero rate invalid", func(t *testing.T) {
		calls := 0
		rates := &CurrencyRateDAOMock{
			FindByPairDateFunc: func(_ context.Context, _, _ string, _ *uint64, _ time.Time) (*CurrencyRate, error) {
				calls++
				if calls == 1 {
					return nil, nil
				}
				return &CurrencyRate{Rate: 0}, nil
			},
		}
		_, err := NewCurrencyConverter(rates).Convert(ctx, amount.FromFloat64(100), "USD", "IDR", nil, time.Now())
		helper.AssertError(t, err, true, ErrCurrencyRateInvalid)
	})
}

func TestCurrencyConverter_HasRate_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("direct lookup error propagates", func(t *testing.T) {
		rates := &CurrencyRateDAOMock{
			FindByPairFunc: func(_ context.Context, _, _ string, _ *uint64) (*CurrencyRate, error) {
				return nil, dbErr
			},
		}
		_, err := NewCurrencyConverter(rates).HasRate(ctx, "USD", "IDR", nil)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("reverse lookup error propagates", func(t *testing.T) {
		calls := 0
		rates := &CurrencyRateDAOMock{
			FindByPairFunc: func(_ context.Context, _, _ string, _ *uint64) (*CurrencyRate, error) {
				calls++
				if calls == 1 {
					return nil, nil
				}
				return nil, dbErr
			},
		}
		_, err := NewCurrencyConverter(rates).HasRate(ctx, "USD", "IDR", nil)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("reverse rate found", func(t *testing.T) {
		calls := 0
		rates := &CurrencyRateDAOMock{
			FindByPairFunc: func(_ context.Context, _, _ string, _ *uint64) (*CurrencyRate, error) {
				calls++
				if calls == 1 {
					return nil, nil
				}
				return &CurrencyRate{Rate: 16000}, nil
			},
		}
		ok, err := NewCurrencyConverter(rates).HasRate(ctx, "USD", "IDR", nil)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if !ok {
			t.Error("expected true for reverse rate")
		}
	})
}

func TestAgreementService_Find_List(t *testing.T) {
	ctx := context.Background()

	t.Run("find delegates and maps missing", func(t *testing.T) {
		svc, agreements, _ := testAgreementService()
		agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) {
			return &SupplyAgreement{SupplierID: 10}, nil
		}
		got, err := svc.Find(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.SupplierID != 10 {
			t.Errorf("supplier = %d, want 10", got.SupplierID)
		}

		agreements.FindFunc = func(_ context.Context, _ uint64) (*SupplyAgreement, error) {
			return nil, errors.New("db down")
		}
		_, err = svc.Find(ctx, 1)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("list delegates", func(t *testing.T) {
		svc, agreements, _ := testAgreementService()
		agreements.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[SupplyAgreement], error) {
			return &query.Page[SupplyAgreement]{Items: []*SupplyAgreement{{SupplierID: 10}}, Count: 1}, nil
		}
		page, err := svc.List(ctx, &query.Query{})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if page.Count != 1 {
			t.Errorf("count = %d, want 1", page.Count)
		}
	})
}

func TestSupplierScorecardService_Find_List(t *testing.T) {
	ctx := context.Background()

	newSvc := func(mock SupplierScorecardDAOMock) SupplierScorecardService {
		return NewSupplierScorecardService(mock)
	}

	t.Run("find maps missing and propagates error", func(t *testing.T) {
		mock := SupplierScorecardDAOMock{}
		mock.FindFunc = func(_ context.Context, _ uint64) (*SupplierScorecard, error) {
			return nil, nil
		}
		_, err := newSvc(mock).Find(ctx, 1)
		helper.AssertError(t, err, true, ErrScorecardNotFound)

		mock.FindFunc = func(_ context.Context, _ uint64) (*SupplierScorecard, error) {
			return &SupplierScorecard{SupplierID: 20}, nil
		}
		got, err := newSvc(mock).Find(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.SupplierID != 20 {
			t.Errorf("supplier = %d, want 20", got.SupplierID)
		}

		mock.FindFunc = func(_ context.Context, _ uint64) (*SupplierScorecard, error) {
			return nil, errors.New("db down")
		}
		_, err = newSvc(mock).Find(ctx, 1)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("list delegates and propagates error", func(t *testing.T) {
		mock := SupplierScorecardDAOMock{}
		mock.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[SupplierScorecard], error) {
			return &query.Page[SupplierScorecard]{Items: []*SupplierScorecard{{SupplierID: 20}}, Count: 1}, nil
		}
		page, err := newSvc(mock).List(ctx, &query.Query{})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if page.Count != 1 {
			t.Errorf("count = %d, want 1", page.Count)
		}

		mock.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[SupplierScorecard], error) {
			return nil, errors.New("db down")
		}
		_, err = newSvc(mock).List(ctx, &query.Query{})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("list by supplier propagates error", func(t *testing.T) {
		mock := SupplierScorecardDAOMock{}
		mock.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[SupplierScorecard], error) {
			return nil, errors.New("db down")
		}
		_, err := newSvc(mock).ListBySupplier(ctx, 20)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
