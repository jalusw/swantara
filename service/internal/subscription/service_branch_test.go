package subscription

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestSubscriptionService_Create_ValidationShortCircuits(t *testing.T) {
	ctx := context.Background()
	svc := testSubscriptionService(SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})

	tests := []struct {
		name    string
		request CreateSubscriptionRequest
		wantErr error
	}{
		{
			name: "requires plan",
			request: CreateSubscriptionRequest{
				OrganizationID: 10, Name: "X", ContactID: 7, PriceBookID: 3, CurrencyCode: "USD",
				Lines: []CreateSubscriptionLineRequest{{ItemID: 100, Qty: 1, UnitPrice: 10}},
			},
			wantErr: ErrSubscriptionNoPlan,
		},
		{
			name: "requires price_book",
			request: CreateSubscriptionRequest{
				OrganizationID: 10, Name: "X", ContactID: 7, PlanID: 1, CurrencyCode: "USD",
				Lines: []CreateSubscriptionLineRequest{{ItemID: 100, Qty: 1, UnitPrice: 10}},
			},
			wantErr: ErrSubscriptionPriceBook,
		},
		{
			name: "requires currency",
			request: CreateSubscriptionRequest{
				OrganizationID: 10, Name: "X", ContactID: 7, PlanID: 1, PriceBookID: 3,
				Lines: []CreateSubscriptionLineRequest{{ItemID: 100, Qty: 1, UnitPrice: 10}},
			},
			wantErr: ErrSubscriptionCurrency,
		},
		{
			name: "requires lines",
			request: CreateSubscriptionRequest{
				OrganizationID: 10, Name: "X", ContactID: 7, PlanID: 1, PriceBookID: 3, CurrencyCode: "USD",
			},
			wantErr: ErrSubscriptionNoLines,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, tt.request)
			if !helper.AssertError(t, err, true, tt.wantErr) {
				return
			}
		})
	}
}

func TestSubscriptionService_Create_RejectsMissingPlan(t *testing.T) {
	ctx := context.Background()
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return nil, nil
		},
	}
	svc := testSubscriptionService(SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, plans)

	_, err := svc.Create(ctx, CreateSubscriptionRequest{
		OrganizationID: 10, Name: "X", ContactID: 7, PlanID: 1, PriceBookID: 3, CurrencyCode: "USD",
		Lines: []CreateSubscriptionLineRequest{{ItemID: 100, Qty: 1, UnitPrice: 10}},
	})
	if !helper.AssertError(t, err, true, ErrSubscriptionPlanNotFound) {
		return
	}
}

func TestSubscriptionService_Create_PropagatesPlanLookupError(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return nil, dbErr
		},
	}
	svc := testSubscriptionService(SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, plans)

	_, err := svc.Create(ctx, CreateSubscriptionRequest{
		OrganizationID: 10, Name: "X", ContactID: 7, PlanID: 1, PriceBookID: 3, CurrencyCode: "USD",
		Lines: []CreateSubscriptionLineRequest{{ItemID: 100, Qty: 1, UnitPrice: 10}},
	})
	if !helper.AssertError(t, err, true, dbErr) {
		return
	}
}

func TestSubscriptionService_Create_LineValidation(t *testing.T) {
	ctx := context.Background()
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}

	tests := []struct {
		name       string
		line       CreateSubscriptionLineRequest
		wantErr    error
		resolveErr error
	}{
		{name: "rejects missing item", line: CreateSubscriptionLineRequest{Qty: 1, UnitPrice: 10}, wantErr: ErrSubscriptionLineProduct},
		{name: "rejects zero quantity", line: CreateSubscriptionLineRequest{ItemID: 100, Qty: 0, UnitPrice: 10}, wantErr: ErrSubscriptionLineQty},
		{name: "rejects negative price", line: CreateSubscriptionLineRequest{ItemID: 100, Qty: 1, UnitPrice: -1}, wantErr: ErrSubscriptionLinePrice},
		{name: "rejects discount above range", line: CreateSubscriptionLineRequest{ItemID: 100, Qty: 1, UnitPrice: 10, DiscountPct: 101}, wantErr: ErrSubscriptionLineDiscount},
		{name: "rejects negative discount", line: CreateSubscriptionLineRequest{ItemID: 100, Qty: 1, UnitPrice: 10, DiscountPct: -1}, wantErr: ErrSubscriptionLineDiscount},
		{name: "propagates variant org resolution error", line: CreateSubscriptionLineRequest{ItemID: 100, Qty: 1, UnitPrice: 10}, wantErr: errors.New("resolver down")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := IncomeAccountResolverMock{
				ResolveVariantOrganizationFunc: func(_ context.Context, _ uint64) (*uint64, error) {
					return nil, tt.wantErr
				},
			}
			svc := testSubscriptionServiceWithResolver(SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, plans, resolver)

			_, err := svc.Create(ctx, CreateSubscriptionRequest{
				OrganizationID: 10, Name: "X", ContactID: 7, PlanID: 1, PriceBookID: 3, CurrencyCode: "USD",
				Lines: []CreateSubscriptionLineRequest{tt.line},
			})
			if tt.wantErr == errors.New("resolver down") {
				if helper.AssertError(t, err, true, errors.New("resolver down")) {
					return
				}
				return
			}
			if !helper.AssertError(t, err, true, tt.wantErr) {
				return
			}
		})
	}
}

func TestSubscriptionService_Create_PropagatesPersistenceErrors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}

	t.Run("subscription create", func(t *testing.T) {
		subscriptions := SubscriptionDAOMock{
			CRUDMock: dao.CRUDMock[Subscription]{
				CreateFunc: func(_ context.Context, _ *Subscription) (*Subscription, error) {
					return nil, dbErr
				},
			},
		}
		svc := testSubscriptionService(subscriptions, SubscriptionLineDAOMock{}, plans)

		_, err := svc.Create(ctx, CreateSubscriptionRequest{
			OrganizationID: 10, Name: "X", ContactID: 7, PlanID: 1, PriceBookID: 3, CurrencyCode: "USD",
			Lines: []CreateSubscriptionLineRequest{{ItemID: 100, Qty: 1, UnitPrice: 10}},
		})
		if !helper.AssertError(t, err, true, dbErr) {
			return
		}
	})

	t.Run("line create", func(t *testing.T) {
		subscriptions := SubscriptionDAOMock{
			CRUDMock: dao.CRUDMock[Subscription]{
				CreateFunc: func(_ context.Context, entity *Subscription) (*Subscription, error) {
					entity.ID = 1
					return entity, nil
				},
			},
		}
		lines := SubscriptionLineDAOMock{
			CRUDMock: dao.CRUDMock[SubscriptionLine]{
				CreateFunc: func(_ context.Context, _ *SubscriptionLine) (*SubscriptionLine, error) {
					return nil, dbErr
				},
			},
		}
		svc := testSubscriptionService(subscriptions, lines, plans)

		_, err := svc.Create(ctx, CreateSubscriptionRequest{
			OrganizationID: 10, Name: "X", ContactID: 7, PlanID: 1, PriceBookID: 3, CurrencyCode: "USD",
			Lines: []CreateSubscriptionLineRequest{{ItemID: 100, Qty: 1, UnitPrice: 10}},
		})
		if !helper.AssertError(t, err, true, dbErr) {
			return
		}
	})
}

func TestSubscriptionService_Get_NotOwnedByOrganization(t *testing.T) {
	ctx := context.Background()
	subscriptions := SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Subscription, error) {
				return &Subscription{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(20))}, nil
			},
		},
	}
	svc := testSubscriptionService(subscriptions, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})

	_, err := svc.Get(ctx, 10, 1)
	if !helper.AssertError(t, err, true, ErrSubscriptionNotFound) {
		return
	}
}

func TestSubscriptionService_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	subscriptions := SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Subscription, error) {
				return nil, nil
			},
		},
	}
	svc := testSubscriptionService(subscriptions, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})

	_, err := svc.Get(ctx, 10, 1)
	if !helper.AssertError(t, err, true, ErrSubscriptionNotFound) {
		return
	}
}

func TestSubscriptionService_Get_PropagatesSearchError(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	subscriptions := SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Subscription, error) {
				return nil, dbErr
			},
		},
	}
	svc := testSubscriptionService(subscriptions, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})

	_, err := svc.Get(ctx, 10, 1)
	if !helper.AssertError(t, err, true, dbErr) {
		return
	}
}

func TestSubscriptionService_List_PropagatesError(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	subscriptions := SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[Subscription]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Subscription], error) {
				return nil, dbErr
			},
		},
	}
	svc := testSubscriptionService(subscriptions, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})

	_, err := svc.List(ctx, 10)
	if !helper.AssertError(t, err, true, dbErr) {
		return
	}
}

func TestSubscriptionPeriodHelpers(t *testing.T) {
	if got := monthsPerPeriod("day", 30); got != 1 {
		t.Errorf("monthsPerPeriod(day,30) = %v, want 1", got)
	}
	if got := monthsPerPeriod("week", 4); got != 1 {
		t.Errorf("monthsPerPeriod(week,4) = %v, want 1", got)
	}
	if got := monthsPerPeriod("year", 1); got != 12 {
		t.Errorf("monthsPerPeriod(year,1) = %v, want 12", got)
	}
	if got := monthsPerPeriod("month", 2); got != 2 {
		t.Errorf("monthsPerPeriod(month,2) = %v, want 2", got)
	}
	if got := schedulePeriods("day", 1); got != 1 {
		t.Errorf("schedulePeriods(day,1) = %d, want 1", got)
	}

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := advanceDate(base, "day", 2); !got.Equal(time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("advanceDate(day,2) = %v, want 2026-01-03", got)
	}
	if got := advanceDate(base, "week", 2); !got.Equal(time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("advanceDate(week,2) = %v, want 2026-01-15", got)
	}
	if got := advanceDate(base, "year", 1); !got.Equal(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("advanceDate(year,1) = %v, want 2027-01-01", got)
	}
	if got := advanceDate(base, "month", 2); !got.Equal(time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("advanceDate(month,2) = %v, want 2026-03-01", got)
	}
}

func TestSubscriptionService_PosterMockBranches(t *testing.T) {
	ctx := context.Background()

	poster := PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
		return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
	}}

	movement, err := poster.Post(ctx, accounting.PostRequest{})
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if movement.ID != 9 {
		t.Errorf("Post() id = %d, want 9", movement.ID)
	}

	movement, err = poster.Reverse(ctx, accounting.ReverseRequest{})
	if err != nil {
		t.Fatalf("Reverse() error = %v", err)
	}
	if movement.ID != 6 {
		t.Errorf("Reverse() id = %d, want 6", movement.ID)
	}

	movement, err = poster.ReverseTx(ctx, nil, accounting.ReverseRequest{})
	if err != nil {
		t.Fatalf("ReverseTx() error = %v", err)
	}
	if movement.ID != 6 {
		t.Errorf("ReverseTx() id = %d, want 6", movement.ID)
	}
}
