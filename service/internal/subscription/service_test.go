package subscription

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestSubscriptionService_Create_ComputesMonthlyMRR(t *testing.T) {
	ctx := context.Background()
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	var created *Subscription
	subscriptions := SubscriptionDAOMock{CRUDMock: dao.CRUDMock[Subscription]{
		CreateFunc: func(_ context.Context, entity *Subscription) (*Subscription, error) {
			created = entity
			entity.ID = 1
			return entity, nil
		},
	}}
	lines := SubscriptionLineDAOMock{CRUDMock: dao.CRUDMock[SubscriptionLine]{
		CreateFunc: func(_ context.Context, entity *SubscriptionLine) (*SubscriptionLine, error) {
			entity.ID = 2
			return entity, nil
		},
	}}
	svc := testSubscriptionService(subscriptions, lines, plans)

	createdSub, err := svc.Create(ctx, CreateSubscriptionRequest{
		OrganizationID: 10,
		Name:           "Pro",
		ContactID:      7,
		PlanID:         1,
		PriceBookID:    3,
		CurrencyCode:   "USD",
		Lines: []CreateSubscriptionLineRequest{
			{ItemID: 100, Qty: 1, UnitPrice: 120, DiscountPct: 0},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createdSub.State != SubscriptionStateDraft {
		t.Errorf("state = %s, want draft", createdSub.State)
	}
	if createdSub.MRR != 120 {
		t.Errorf("mrr = %v, want 120", createdSub.MRR)
	}
	if created == nil || created.OrganizationID == nil || *created.OrganizationID != 10 {
		t.Errorf("created subscription = %+v, want organization 10", created)
	}
}

func TestSubscriptionService_Create_AnnualPlanSplitsMRR(t *testing.T) {
	ctx := context.Background()
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, RecurringInterval: "year", RecurringCount: 1}, nil
		},
	}
	subscriptions := SubscriptionDAOMock{}
	lines := SubscriptionLineDAOMock{}
	svc := testSubscriptionService(subscriptions, lines, plans)

	created, err := svc.Create(ctx, CreateSubscriptionRequest{
		OrganizationID: 10,
		Name:           "Pro Yearly",
		ContactID:      7,
		PlanID:         1,
		PriceBookID:    3,
		CurrencyCode:   "USD",
		Lines: []CreateSubscriptionLineRequest{
			{ItemID: 100, Qty: 1, UnitPrice: 1200, DiscountPct: 0},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.MRR != 100 {
		t.Errorf("mrr = %v, want 100", created.MRR)
	}
}

func TestSubscriptionService_Create_RequiresFields(t *testing.T) {
	ctx := context.Background()
	svc := testSubscriptionService(SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})

	_, err := svc.Create(ctx, CreateSubscriptionRequest{OrganizationID: 10, Name: "X", PlanID: 1, PriceBookID: 3, CurrencyCode: "USD"})
	if helper.AssertError(t, err, true, ErrSubscriptionContact) {
		return
	}
}

func TestSubscriptionService_Create_RejectsPlanFromAnotherOrganization(t *testing.T) {
	ctx := context.Background()
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(20)), RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	svc := testSubscriptionService(SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, plans)

	_, err := svc.Create(ctx, CreateSubscriptionRequest{
		OrganizationID: 10,
		Name:           "Pro",
		ContactID:      7,
		PlanID:         1,
		PriceBookID:    3,
		CurrencyCode:   "USD",
		Lines: []CreateSubscriptionLineRequest{
			{ItemID: 100, Qty: 1, UnitPrice: 120, DiscountPct: 0},
		},
	})
	if helper.AssertError(t, err, true, ErrSubscriptionPlanOrganization) {
		return
	}
}

func TestSubscriptionService_Create_RejectsProductFromAnotherOrganization(t *testing.T) {
	ctx := context.Background()
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	resolver := IncomeAccountResolverMock{
		ResolveVariantOrganizationFunc: func(_ context.Context, _ uint64) (*uint64, error) {
			return helper.Ptr(uint64(20)), nil
		},
	}
	svc := testSubscriptionServiceWithResolver(SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, plans, resolver)

	_, err := svc.Create(ctx, CreateSubscriptionRequest{
		OrganizationID: 10,
		Name:           "Pro",
		ContactID:      7,
		PlanID:         1,
		PriceBookID:    3,
		CurrencyCode:   "USD",
		Lines: []CreateSubscriptionLineRequest{
			{ItemID: 100, Qty: 1, UnitPrice: 120, DiscountPct: 0},
		},
	})
	if helper.AssertError(t, err, true, ErrSubscriptionLineProductOrganization) {
		return
	}
}

func TestSubscriptionService_Create_RejectsContactFromAnotherOrganization(t *testing.T) {
	ctx := context.Background()
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	svc := NewSubscriptionService(
		SubscriptionDAOMock{},
		SubscriptionLineDAOMock{},
		plans,
		contacts.ContactDAOMock{
			CRUDMock: dao.CRUDMock[contacts.Contact]{
				FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
					return &contacts.Contact{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(20))}, nil
				},
			},
		},
		testPriceBookDAO(),
		IncomeAccountResolverMock{},
		InvoiceEngineMock{},
		accounting.DeferralService{},
		SubscriptionConfigSourceMock{},
		TransactionerMock{},
	)

	_, err := svc.Create(ctx, CreateSubscriptionRequest{
		OrganizationID: 10,
		Name:           "Pro",
		ContactID:      7,
		PlanID:         1,
		PriceBookID:    3,
		CurrencyCode:   "USD",
		Lines: []CreateSubscriptionLineRequest{
			{ItemID: 100, Qty: 1, UnitPrice: 120, DiscountPct: 0},
		},
	})
	if helper.AssertError(t, err, true, ErrSubscriptionContactOrganization) {
		return
	}
}

func TestSubscriptionService_Create_RejectsPriceBookFromAnotherOrganization(t *testing.T) {
	ctx := context.Background()
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	svc := NewSubscriptionService(
		SubscriptionDAOMock{},
		SubscriptionLineDAOMock{},
		plans,
		testContactDAO(),
		products.PriceBookDAOMock{
			CRUDMock: dao.CRUDMock[products.PriceBook]{
				FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
					return &products.PriceBook{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(20))}, nil
				},
			},
		},
		IncomeAccountResolverMock{},
		InvoiceEngineMock{},
		accounting.DeferralService{},
		SubscriptionConfigSourceMock{},
		TransactionerMock{},
	)

	_, err := svc.Create(ctx, CreateSubscriptionRequest{
		OrganizationID: 10,
		Name:           "Pro",
		ContactID:      7,
		PlanID:         1,
		PriceBookID:    3,
		CurrencyCode:   "USD",
		Lines: []CreateSubscriptionLineRequest{
			{ItemID: 100, Qty: 1, UnitPrice: 120, DiscountPct: 0},
		},
	})
	if helper.AssertError(t, err, true, ErrSubscriptionPriceBookOrganization) {
		return
	}
}

func TestSubscriptionService_Activate_SetsDatesAndState(t *testing.T) {
	ctx := context.Background()
	subscriptions := SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Subscription, error) {
				return &Subscription{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), State: SubscriptionStateDraft}, nil
			},
			UpdateFunc: func(_ context.Context, entity *Subscription) (*Subscription, error) {
				return entity, nil
			},
		},
	}
	svc := testSubscriptionService(subscriptions, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	svc.now = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

	active, err := svc.Activate(ctx, 10, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if active.State != SubscriptionStateActive {
		t.Errorf("state = %s, want active", active.State)
	}
	if active.DateStart == nil || active.NextInvoiceDate == nil {
		t.Error("date_start and next_invoice_date must be set on activation")
	}
}

func TestSubscriptionService_Activate_RejectsNonDraft(t *testing.T) {
	ctx := context.Background()
	subscriptions := SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Subscription, error) {
				return &Subscription{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), State: SubscriptionStateActive}, nil
			},
		},
	}
	svc := testSubscriptionService(subscriptions, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})

	_, err := svc.Activate(ctx, 10, 1)
	if helper.AssertError(t, err, true, ErrSubscriptionState) {
		return
	}
}

func TestSubscriptionService_PauseResumeChurnClose_Transitions(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name    string
		from    string
		action  func(svc SubscriptionService, ctx context.Context) error
		want    string
		wantErr error
	}{
		{name: "pause active", from: SubscriptionStateActive, action: func(svc SubscriptionService, ctx context.Context) error { _, err := svc.Pause(ctx, 10, 1); return err }, want: SubscriptionStatePaused},
		{name: "pause draft rejected", from: SubscriptionStateDraft, action: func(svc SubscriptionService, ctx context.Context) error { _, err := svc.Pause(ctx, 10, 1); return err }, wantErr: ErrSubscriptionState},
		{name: "resume paused", from: SubscriptionStatePaused, action: func(svc SubscriptionService, ctx context.Context) error { _, err := svc.Resume(ctx, 10, 1); return err }, want: SubscriptionStateActive},
		{name: "churn active", from: SubscriptionStateActive, action: func(svc SubscriptionService, ctx context.Context) error { _, err := svc.Churn(ctx, 10, 1); return err }, want: SubscriptionStateChurned},
		{name: "churn draft rejected", from: SubscriptionStateDraft, action: func(svc SubscriptionService, ctx context.Context) error { _, err := svc.Churn(ctx, 10, 1); return err }, wantErr: ErrSubscriptionState},
		{name: "close active", from: SubscriptionStateActive, action: func(svc SubscriptionService, ctx context.Context) error { _, err := svc.Close(ctx, 10, 1); return err }, want: SubscriptionStateClosed},
		{name: "close paused", from: SubscriptionStatePaused, action: func(svc SubscriptionService, ctx context.Context) error { _, err := svc.Close(ctx, 10, 1); return err }, want: SubscriptionStateClosed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			current := tc.from
			subscriptions := SubscriptionDAOMock{
				CRUDMock: dao.CRUDMock[Subscription]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*Subscription, error) {
						return &Subscription{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), State: current}, nil
					},
					UpdateFunc: func(_ context.Context, entity *Subscription) (*Subscription, error) {
						current = entity.State
						return entity, nil
					},
				},
			}
			svc := testSubscriptionService(subscriptions, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})

			err := tc.action(svc, ctx)
			if tc.wantErr != nil {
				if helper.AssertError(t, err, true, tc.wantErr) {
					return
				}
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			updated, _ := svc.Get(ctx, 10, 1)
			if updated.State != tc.want {
				t.Errorf("state = %s, want %s", updated.State, tc.want)
			}
		})
	}
}

func TestSubscriptionService_GenerateNextInvoice_CreatesInvoiceAndAdvancesDate(t *testing.T) {
	ctx := context.Background()
	dueDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	subscriptions := SubscriptionDAOMock{
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, entity *Subscription) (*Subscription, error) {
			return entity, nil
		},
		CRUDMock: dao.CRUDMock[Subscription]{
			FindFunc: func(_ context.Context, _ uint64) (*Subscription, error) {
				return &Subscription{
					Base:            model.Base{ID: 1},
					OrganizationID:  helper.Ptr(uint64(10)),
					ContactID:       helper.Ptr(uint64(7)),
					PlanID:          helper.Ptr(uint64(1)),
					PriceBookID:     helper.Ptr(uint64(3)),
					State:           SubscriptionStateActive,
					NextInvoiceDate: &dueDate,
				}, nil
			},
		},
	}
	lines := SubscriptionLineDAOMock{
		ListBySubscriptionFunc: func(_ context.Context, _ uint64) ([]*SubscriptionLine, error) {
			return []*SubscriptionLine{{Base: model.Base{ID: 2}, ItemID: helper.Ptr(uint64(100)), Qty: 1, UnitPrice: 120, DiscountPct: 0}}, nil
		},
	}
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	schedules := accounting.DeferredScheduleDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, entity *accounting.DeferredSchedule) (*accounting.DeferredSchedule, error) {
			entity.ID = 4
			return entity, nil
		},
	}
	scheduleLines := accounting.DeferredScheduleLineDAOMock{}
	var request accounting.CreateInvoiceRequest
	invoices := InvoiceEngineMock{
		CreateFunc: func(_ context.Context, req accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
			request = req
			return &accounting.Invoice{Base: model.Base{ID: 9}, ContactID: req.ContactID, AmountTotal: amount.FromFloat64(120)}, nil
		},
	}
	var updatedSubscription *Subscription
	subscriptions.UpdateTxFunc = func(_ context.Context, _ *gorm.DB, entity *Subscription) (*Subscription, error) {
		updatedSubscription = entity
		return entity, nil
	}
	svc := testSubscriptionServiceWithDeferrals(subscriptions, lines, plans, schedules, scheduleLines, invoices)

	invoice, err := svc.GenerateNextInvoice(ctx, 1, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if invoice.ID != 9 {
		t.Errorf("invoice id = %d, want 9", invoice.ID)
	}
	if request.DeferredAccount == nil || *request.DeferredAccount != 300 {
		t.Errorf("deferred account = %v, want 300", request.DeferredAccount)
	}
	if len(request.Lines) != 1 || request.Lines[0].AccountID != 200 {
		t.Errorf("invoice lines = %+v, want one line with account 200", request.Lines)
	}
	if updatedSubscription.NextInvoiceDate == nil || !updatedSubscription.NextInvoiceDate.Equal(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("next_invoice_date = %v, want 2026-02-01", updatedSubscription.NextInvoiceDate)
	}
}

func TestSubscriptionService_GenerateNextInvoice_NotDue(t *testing.T) {
	ctx := context.Background()
	futureDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	subscriptions := SubscriptionDAOMock{CRUDMock: dao.CRUDMock[Subscription]{
		FindFunc: func(_ context.Context, _ uint64) (*Subscription, error) {
			return &Subscription{
				Base:            model.Base{ID: 1},
				OrganizationID:  helper.Ptr(uint64(10)),
				ContactID:       helper.Ptr(uint64(7)),
				PlanID:          helper.Ptr(uint64(1)),
				PriceBookID:     helper.Ptr(uint64(3)),
				State:           SubscriptionStateActive,
				NextInvoiceDate: &futureDate,
			}, nil
		},
	}}
	svc := testSubscriptionService(subscriptions, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})

	_, err := svc.GenerateNextInvoice(ctx, 1, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
	if helper.AssertError(t, err, true, ErrSubscriptionNotDue) {
		return
	}
}

func TestSubscriptionService_GenerateNextInvoice_RequiresJournalConfig(t *testing.T) {
	ctx := context.Background()
	dueDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	subscriptions := SubscriptionDAOMock{CRUDMock: dao.CRUDMock[Subscription]{
		FindFunc: func(_ context.Context, _ uint64) (*Subscription, error) {
			return &Subscription{
				Base:            model.Base{ID: 1},
				OrganizationID:  helper.Ptr(uint64(10)),
				ContactID:       helper.Ptr(uint64(7)),
				PlanID:          helper.Ptr(uint64(1)),
				PriceBookID:     helper.Ptr(uint64(3)),
				State:           SubscriptionStateActive,
				NextInvoiceDate: &dueDate,
			}, nil
		},
	}}
	lines := SubscriptionLineDAOMock{
		ListBySubscriptionFunc: func(_ context.Context, _ uint64) ([]*SubscriptionLine, error) {
			return []*SubscriptionLine{{Base: model.Base{ID: 2}, ItemID: helper.Ptr(uint64(100)), Qty: 1, UnitPrice: 120}}, nil
		},
	}
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	configs := SubscriptionConfigSourceMock{
		JournalIDFunc: func(_ context.Context, _ uint64) (uint64, error) { return 0, nil },
	}
	svc := testSubscriptionServiceWithConfigSource(subscriptions, lines, plans, configs)

	_, err := svc.GenerateNextInvoice(ctx, 1, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
	if helper.AssertError(t, err, true, ErrSubscriptionConfig) {
		return
	}
}

func TestSubscriptionService_BillDue_IsIdempotent(t *testing.T) {
	ctx := context.Background()
	asOf := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	calls := 0
	subscriptions := SubscriptionDAOMock{
		ListDueFunc: func(_ context.Context, _ time.Time) ([]*Subscription, error) {
			if calls > 0 {
				return []*Subscription{}, nil
			}
			return []*Subscription{{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))}}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, entity *Subscription) (*Subscription, error) {
			return entity, nil
		},
		CRUDMock: dao.CRUDMock[Subscription]{
			FindFunc: func(_ context.Context, _ uint64) (*Subscription, error) {
				return &Subscription{
					Base:            model.Base{ID: 1},
					OrganizationID:  helper.Ptr(uint64(10)),
					ContactID:       helper.Ptr(uint64(7)),
					PlanID:          helper.Ptr(uint64(1)),
					PriceBookID:     helper.Ptr(uint64(3)),
					State:           SubscriptionStateActive,
					NextInvoiceDate: &dueDate,
				}, nil
			},
		},
	}
	lines := SubscriptionLineDAOMock{
		ListBySubscriptionFunc: func(_ context.Context, _ uint64) ([]*SubscriptionLine, error) {
			return []*SubscriptionLine{{Base: model.Base{ID: 2}, ItemID: helper.Ptr(uint64(100)), Qty: 1, UnitPrice: 120}}, nil
		},
	}
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	invoiceCount := 0
	invoices := InvoiceEngineMock{
		CreateFunc: func(_ context.Context, _ accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
			invoiceCount++
			return &accounting.Invoice{Base: model.Base{ID: 9}}, nil
		},
	}
	svc := testSubscriptionServiceWithDeferrals(subscriptions, lines, plans, accounting.DeferredScheduleDAOMock{}, accounting.DeferredScheduleLineDAOMock{}, invoices)

	count, err := svc.BillDue(ctx, asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 || invoiceCount != 1 {
		t.Errorf("first run count = %d invoices = %d, want 1 and 1", count, invoiceCount)
	}
	calls++
	count, err = svc.BillDue(ctx, asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 || invoiceCount != 1 {
		t.Errorf("second run count = %d invoices = %d, want 0 and 1", count, invoiceCount)
	}
}

func TestSubscriptionService_RecognizeRevenue_DelegatesToDeferralService(t *testing.T) {
	ctx := context.Background()
	asOf := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	schedules := accounting.DeferredScheduleDAOMock{
		ListRunningFunc: func(_ context.Context, _ *uint64) ([]*accounting.DeferredSchedule, error) {
			return []*accounting.DeferredSchedule{{
				Base:                  model.Base{ID: 4},
				OrganizationID:        helper.Ptr(uint64(10)),
				TotalAmount:           240,
				BalanceSheetAccountID: helper.Ptr(uint64(300)),
				PLAccountID:           helper.Ptr(uint64(200)),
				SourceID:              2,
				State:                 accounting.DeferredStateRunning,
			}}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, entity *accounting.DeferredSchedule) (*accounting.DeferredSchedule, error) {
			return entity, nil
		},
	}
	scheduleLines := accounting.DeferredScheduleLineDAOMock{
		ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*accounting.DeferredScheduleLine, error) {
			return []*accounting.DeferredScheduleLine{
				{Base: model.Base{ID: 11}, Sequence: 1, Amount: 120, RecognitionDate: helper.Ptr(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))},
			}, nil
		},
	}
	deferrals := accounting.NewDeferralService(schedules, scheduleLines, PosterMock{}, SubscriptionConfigSourceMock{}, TransactionerMock{})

	posted, err := deferrals.RecognizeDue(ctx, nil, asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if posted != 1 {
		t.Errorf("posted = %d, want 1", posted)
	}
}

func TestSubscriptionService_Metrics_ComputesMRRArrChurnLTV(t *testing.T) {
	ctx := context.Background()
	subscriptions := SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[Subscription]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Subscription], error) {
				return &query.Page[Subscription]{Items: []*Subscription{
					{State: SubscriptionStateActive, MRR: 100},
					{State: SubscriptionStatePaused, MRR: 50},
					{State: SubscriptionStateChurned, MRR: 0},
				}}, nil
			},
		},
	}
	svc := testSubscriptionService(subscriptions, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})

	metrics, err := svc.Metrics(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.MRR != 150 {
		t.Errorf("mrr = %v, want 150", metrics.MRR)
	}
	if metrics.ARR != 1800 {
		t.Errorf("arr = %v, want 1800", metrics.ARR)
	}
	if metrics.Churned != 1 {
		t.Errorf("churned = %d, want 1", metrics.Churned)
	}
}

func TestSubscriptionService_List_ScopesToOrganization(t *testing.T) {
	ctx := context.Background()
	var queryField string
	var queryValue any
	subscriptions := SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[Subscription]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[Subscription], error) {
				if len(q.Filters) > 0 {
					queryField = q.Filters[0].Field
					queryValue = q.Filters[0].Value
				}
				return &query.Page[Subscription]{Items: []*Subscription{{Base: model.Base{ID: 1}}}}, nil
			},
		},
	}
	svc := testSubscriptionService(subscriptions, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})

	page, err := svc.List(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page) != 1 {
		t.Errorf("items = %d, want 1", len(page))
	}
	if queryField != "organization_id" || queryValue != uint64(10) {
		t.Errorf("list by %s = %v, want organization_id = 10", queryField, queryValue)
	}
}

func testSubscriptionService(subscriptions SubscriptionDAOMock, lines SubscriptionLineDAOMock, plans dao.CRUDMock[reference.SubscriptionPlan]) SubscriptionService {
	return NewSubscriptionService(
		subscriptions,
		lines,
		plans,
		testContactDAO(),
		testPriceBookDAO(),
		IncomeAccountResolverMock{},
		InvoiceEngineMock{},
		accounting.DeferralService{},
		SubscriptionConfigSourceMock{},
		TransactionerMock{},
	)
}

func testSubscriptionServiceWithResolver(subscriptions SubscriptionDAOMock, lines SubscriptionLineDAOMock, plans dao.CRUDMock[reference.SubscriptionPlan], resolver IncomeAccountResolverMock) SubscriptionService {
	return NewSubscriptionService(
		subscriptions,
		lines,
		plans,
		testContactDAO(),
		testPriceBookDAO(),
		resolver,
		InvoiceEngineMock{},
		accounting.DeferralService{},
		SubscriptionConfigSourceMock{},
		TransactionerMock{},
	)
}

func testSubscriptionServiceWithDeferrals(subscriptions SubscriptionDAOMock, lines SubscriptionLineDAOMock, plans dao.CRUDMock[reference.SubscriptionPlan], schedules accounting.DeferredScheduleDAOMock, scheduleLines accounting.DeferredScheduleLineDAOMock, invoices InvoiceEngineMock) SubscriptionService {
	deferrals := accounting.NewDeferralService(
		schedules,
		scheduleLines,
		PosterMock{},
		SubscriptionConfigSourceMock{},
		TransactionerMock{},
	)
	return NewSubscriptionService(
		subscriptions,
		lines,
		plans,
		testContactDAO(),
		testPriceBookDAO(),
		IncomeAccountResolverMock{},
		invoices,
		deferrals,
		SubscriptionConfigSourceMock{},
		TransactionerMock{},
	)
}

func testSubscriptionServiceWithConfigSource(subscriptions SubscriptionDAOMock, lines SubscriptionLineDAOMock, plans dao.CRUDMock[reference.SubscriptionPlan], configs SubscriptionConfigSourceMock) SubscriptionService {
	return NewSubscriptionService(
		subscriptions,
		lines,
		plans,
		testContactDAO(),
		testPriceBookDAO(),
		IncomeAccountResolverMock{},
		InvoiceEngineMock{},
		accounting.DeferralService{},
		configs,
		TransactionerMock{},
	)
}

func testContactDAO() contacts.ContactDAOMock {
	return contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return &contacts.Contact{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10))}, nil
			},
		},
	}
}

func testPriceBookDAO() products.PriceBookDAOMock {
	return products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return &products.PriceBook{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(10))}, nil
			},
		},
	}
}
