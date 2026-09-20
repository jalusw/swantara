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
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func invoiceableSub() *Subscription {
	orgID := uint64(10)
	next := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	return &Subscription{
		Base:            model.Base{ID: 1},
		OrganizationID:  &orgID,
		ContactID:       helper.Ptr(uint64(20)),
		PlanID:          helper.Ptr(uint64(3)),
		PriceBookID:     helper.Ptr(uint64(4)),
		State:           SubscriptionStateActive,
		NextInvoiceDate: &next,
	}
}

func invoiceTestService(sub *Subscription, subErr error, plan *reference.SubscriptionPlan, planErr error, lines []*SubscriptionLine, linesErr error) SubscriptionService {
	schedules := accounting.DeferredScheduleDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, entity *accounting.DeferredSchedule) (*accounting.DeferredSchedule, error) {
			entity.ID = 4
			return entity, nil
		},
	}
	svc := testSubscriptionServiceWithDeferrals(SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[Subscription]{
			FindFunc: func(_ context.Context, _ uint64) (*Subscription, error) { return sub, subErr },
		},
	}, SubscriptionLineDAOMock{
		ListBySubscriptionFunc: func(_ context.Context, _ uint64) ([]*SubscriptionLine, error) { return lines, linesErr },
	}, dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) { return plan, planErr },
	}, schedules, accounting.DeferredScheduleLineDAOMock{}, InvoiceEngineMock{})
	return svc
}

func TestSubscriptionService_GenerateNextInvoice(t *testing.T) {
	ctx := context.Background()
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	dbErr := errors.New("db down")
	plan := &reference.SubscriptionPlan{Base: model.Base{ID: 3}, RecurringInterval: "month", RecurringCount: 1}
	lines := []*SubscriptionLine{{Base: model.Base{ID: 5}, ItemID: helper.Ptr(uint64(100)), Qty: 1, UnitPrice: 50}}

	t.Run("generates invoice", func(t *testing.T) {
		svc := invoiceTestService(invoiceableSub(), nil, plan, nil, lines, nil)
		got, err := svc.GenerateNextInvoice(ctx, 1, asOf)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got == nil {
			t.Error("invoice = nil")
		}
	})

	cases := []struct {
		name     string
		sub      *Subscription
		subErr   error
		plan     *reference.SubscriptionPlan
		planErr  error
		lines    []*SubscriptionLine
		linesErr error
		wantErr  error
	}{
		{name: "lookup error", subErr: dbErr, wantErr: dbErr},
		{name: "missing", wantErr: ErrSubscriptionNotFound},
		{name: "bad state", sub: func() *Subscription { s := invoiceableSub(); s.State = SubscriptionStateDraft; return s }(), plan: plan, lines: lines, wantErr: ErrSubscriptionState},
		{name: "not due", sub: func() *Subscription {
			s := invoiceableSub()
			future := asOf.AddDate(0, 1, 0)
			s.NextInvoiceDate = &future
			return s
		}(), plan: plan, lines: lines, wantErr: ErrSubscriptionNotDue},
		{name: "missing ids", sub: &Subscription{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), State: SubscriptionStateActive, NextInvoiceDate: &asOf}, plan: plan, lines: lines, wantErr: ErrSubscriptionNotFound},
		{name: "plan error", sub: invoiceableSub(), planErr: dbErr, lines: lines, wantErr: dbErr},
		{name: "plan missing", sub: invoiceableSub(), lines: lines, wantErr: ErrSubscriptionPlanNotFound},
		{name: "lines error", sub: invoiceableSub(), plan: plan, linesErr: dbErr, wantErr: dbErr},
		{name: "no lines", sub: invoiceableSub(), plan: plan, wantErr: ErrSubscriptionNoLines},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc := invoiceTestService(tt.sub, tt.subErr, tt.plan, tt.planErr, tt.lines, tt.linesErr)
			_, err := svc.GenerateNextInvoice(ctx, 1, asOf)
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestSubscriptionService_Transitions_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("resume lookup error", func(t *testing.T) {
		svc := statefulSubService(nil, dbErr)
		_, err := svc.Resume(ctx, 10, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("churn update error", func(t *testing.T) {
		svc := testSubscriptionService(SubscriptionDAOMock{
			CRUDMock: dao.CRUDMock[Subscription]{
				SearchFunc: func(_ context.Context, _ string, _ any) (*Subscription, error) {
					return subInState(SubscriptionStateActive), nil
				},
				UpdateFunc: func(_ context.Context, _ *Subscription) (*Subscription, error) { return nil, dbErr },
			},
		}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
		_, err := svc.Churn(ctx, 10, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("close update error", func(t *testing.T) {
		svc := testSubscriptionService(SubscriptionDAOMock{
			CRUDMock: dao.CRUDMock[Subscription]{
				SearchFunc: func(_ context.Context, _ string, _ any) (*Subscription, error) {
					return subInState(SubscriptionStateActive), nil
				},
				UpdateFunc: func(_ context.Context, _ *Subscription) (*Subscription, error) { return nil, dbErr },
			},
		}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
		_, err := svc.Close(ctx, 10, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("pause update error", func(t *testing.T) {
		svc := testSubscriptionService(SubscriptionDAOMock{
			CRUDMock: dao.CRUDMock[Subscription]{
				SearchFunc: func(_ context.Context, _ string, _ any) (*Subscription, error) {
					return subInState(SubscriptionStateActive), nil
				},
				UpdateFunc: func(_ context.Context, _ *Subscription) (*Subscription, error) { return nil, dbErr },
			},
		}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
		_, err := svc.Pause(ctx, 10, 1)
		helper.AssertError(t, err, true, dbErr)
	})
}

var _ = accounting.Invoice{}
