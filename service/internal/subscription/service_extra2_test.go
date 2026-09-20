package subscription

import (
	"context"
	"errors"
	"testing"
	"time"

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

func subInState(state string) *Subscription {
	orgID := uint64(10)
	return &Subscription{Base: model.Base{ID: 1}, OrganizationID: &orgID, State: state}
}

func statefulSubService(sub *Subscription, subErr error) SubscriptionService {
	svc := testSubscriptionService(SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[Subscription]{},
	}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	svc.subscriptions = SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*Subscription, error) { return sub, subErr },
			UpdateFunc: func(_ context.Context, s *Subscription) (*Subscription, error) { return s, nil },
		},
	}
	return svc
}

func TestSubscriptionService_Transitions(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("pause resume churn close", func(t *testing.T) {
		svc := statefulSubService(subInState(SubscriptionStateActive), nil)
		got, err := svc.Pause(ctx, 10, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.State != SubscriptionStatePaused {
			t.Errorf("state = %q", got.State)
		}

		svc = statefulSubService(subInState(SubscriptionStatePaused), nil)
		got, err = svc.Resume(ctx, 10, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.State != SubscriptionStateActive {
			t.Errorf("state = %q", got.State)
		}

		svc = statefulSubService(subInState(SubscriptionStatePaused), nil)
		got, err = svc.Churn(ctx, 10, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.State != SubscriptionStateChurned || got.DateEnd == nil {
			t.Errorf("churned = %+v", got)
		}

		svc = statefulSubService(subInState(SubscriptionStateActive), nil)
		got, err = svc.Close(ctx, 10, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.State != SubscriptionStateClosed {
			t.Errorf("state = %q", got.State)
		}
	})

	t.Run("transition guards", func(t *testing.T) {
		svc := statefulSubService(subInState(SubscriptionStateDraft), nil)
		_, err := svc.Pause(ctx, 10, 1)
		helper.AssertError(t, err, true, ErrSubscriptionState)

		svc = statefulSubService(subInState(SubscriptionStateActive), nil)
		_, err = svc.Resume(ctx, 10, 1)
		helper.AssertError(t, err, true, ErrSubscriptionState)

		svc = statefulSubService(subInState(SubscriptionStateDraft), nil)
		_, err = svc.Churn(ctx, 10, 1)
		helper.AssertError(t, err, true, ErrSubscriptionState)

		svc = statefulSubService(subInState(SubscriptionStateDraft), nil)
		_, err = svc.Close(ctx, 10, 1)
		helper.AssertError(t, err, true, ErrSubscriptionState)
	})

	t.Run("get errors propagate", func(t *testing.T) {
		svc := statefulSubService(nil, dbErr)
		_, err := svc.Pause(ctx, 10, 1)
		helper.AssertError(t, err, true, dbErr)

		svc = statefulSubService(nil, nil)
		_, err = svc.Resume(ctx, 10, 1)
		helper.AssertError(t, err, true, ErrSubscriptionNotFound)
	})
}

func TestSubscriptionService_BillDue_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	asOf := time.Now()

	t.Run("propagates list error", func(t *testing.T) {
		svc := testSubscriptionService(SubscriptionDAOMock{
			ListDueFunc: func(_ context.Context, _ time.Time) ([]*Subscription, error) { return nil, dbErr },
		}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
		_, err := svc.BillDue(ctx, asOf)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("propagates invoice error", func(t *testing.T) {
		svc := testSubscriptionService(SubscriptionDAOMock{
			ListDueFunc: func(_ context.Context, _ time.Time) ([]*Subscription, error) {
				return []*Subscription{{Base: model.Base{ID: 1}}}, nil
			},
			CRUDMock: dao.CRUDMock[Subscription]{
				FindFunc: func(_ context.Context, _ uint64) (*Subscription, error) { return nil, dbErr },
			},
		}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
		_, err := svc.BillDue(ctx, asOf)
		helper.AssertError(t, err, true, dbErr)
	})
}

func TestSubscriptionService_Validate(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("contact branches", func(t *testing.T) {
		svc := testSubscriptionService(SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
		if err := svc.validateContact(ctx, 10, 7); helper.AssertError(t, err, false, nil) {
			return
		}
		if err := svc.validatePriceBook(ctx, 10, 3); helper.AssertError(t, err, false, nil) {
			return
		}
	})

	t.Run("contact errors", func(t *testing.T) {
		svc := NewSubscriptionService(
			SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{},
			contacts.ContactDAOMock{
				CRUDMock: dao.CRUDMock[contacts.Contact]{
					FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) { return nil, dbErr },
				},
			}, testPriceBookDAO(), IncomeAccountResolverMock{}, InvoiceEngineMock{}, accounting.DeferralService{}, SubscriptionConfigSourceMock{}, TransactionerMock{},
		)
		helper.AssertError(t, svc.validateContact(ctx, 10, 7), true, dbErr)

		foreign := NewSubscriptionService(
			SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{},
			contacts.ContactDAOMock{
				CRUDMock: dao.CRUDMock[contacts.Contact]{
					FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
						return &contacts.Contact{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(99))}, nil
					},
				},
			}, testPriceBookDAO(), IncomeAccountResolverMock{}, InvoiceEngineMock{}, accounting.DeferralService{}, SubscriptionConfigSourceMock{}, TransactionerMock{},
		)
		helper.AssertError(t, foreign.validateContact(ctx, 10, 7), true, ErrSubscriptionContactOrganization)

		missing := NewSubscriptionService(
			SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{},
			contacts.ContactDAOMock{}, testPriceBookDAO(), IncomeAccountResolverMock{}, InvoiceEngineMock{}, accounting.DeferralService{}, SubscriptionConfigSourceMock{}, TransactionerMock{},
		)
		helper.AssertError(t, missing.validateContact(ctx, 10, 7), true, ErrSubscriptionContact)
	})

	t.Run("price_book errors", func(t *testing.T) {
		svc := NewSubscriptionService(
			SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{},
			testContactDAO(), products.PriceBookDAOMock{
				CRUDMock: dao.CRUDMock[products.PriceBook]{
					FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) { return nil, dbErr },
				},
			}, IncomeAccountResolverMock{}, InvoiceEngineMock{}, accounting.DeferralService{}, SubscriptionConfigSourceMock{}, TransactionerMock{},
		)
		helper.AssertError(t, svc.validatePriceBook(ctx, 10, 3), true, dbErr)

		missing := NewSubscriptionService(
			SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{},
			testContactDAO(), products.PriceBookDAOMock{}, IncomeAccountResolverMock{}, InvoiceEngineMock{}, accounting.DeferralService{}, SubscriptionConfigSourceMock{}, TransactionerMock{},
		)
		helper.AssertError(t, missing.validatePriceBook(ctx, 10, 3), true, ErrSubscriptionPriceBook)

		foreign := NewSubscriptionService(
			SubscriptionDAOMock{}, SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{},
			testContactDAO(), products.PriceBookDAOMock{
				CRUDMock: dao.CRUDMock[products.PriceBook]{
					FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
						return &products.PriceBook{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(99))}, nil
					},
				},
			}, IncomeAccountResolverMock{}, InvoiceEngineMock{}, accounting.DeferralService{}, SubscriptionConfigSourceMock{}, TransactionerMock{},
		)
		helper.AssertError(t, foreign.validatePriceBook(ctx, 10, 3), true, ErrSubscriptionPriceBookOrganization)
	})
}

func TestSubscriptionPlanService_CRUD(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	svc := NewSubscriptionPlanService(dao.CRUDMock[reference.SubscriptionPlan]{})
	if _, err := svc.List(ctx, q); err != nil {
		t.Errorf("List = %v", err)
	}
	if _, err := svc.Find(ctx, 1); err != nil {
		t.Errorf("Find = %v", err)
	}
	if _, err := svc.Create(ctx, &reference.SubscriptionPlan{}); err != nil {
		t.Errorf("Create = %v", err)
	}
	if _, err := svc.Update(ctx, &reference.SubscriptionPlan{}); err != nil {
		t.Errorf("Update = %v", err)
	}
	if err := svc.Delete(ctx, 1); err != nil {
		t.Errorf("Delete = %v", err)
	}
}

func TestSubscriptionFixtures(t *testing.T) {
	if SubscriptionFixture() == nil {
		t.Error("subscription = nil")
	}
	if SubscriptionLineFixture() == nil {
		t.Error("line = nil")
	}
}

func TestSubscriptionMock_All(t *testing.T) {
	ctx := context.Background()
	asOf := time.Now()

	t.Run("subscription dao", func(t *testing.T) {
		bare := SubscriptionDAOMock{}
		if _, err := bare.ListDue(ctx, asOf); err != nil {
			t.Errorf("ListDue = %v", err)
		}
		if _, err := bare.UpdateTx(ctx, nil, &Subscription{}); err != nil {
			t.Errorf("UpdateTx = %v", err)
		}
		wired := SubscriptionDAOMock{
			ListDueFunc: func(_ context.Context, _ time.Time) ([]*Subscription, error) {
				return []*Subscription{{}}, nil
			},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, s *Subscription) (*Subscription, error) { return s, nil },
		}
		_ = wired
	})
}
