package reference

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestReferenceMock_Fallbacks(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	t.Run("account dao", func(t *testing.T) {
		bare := AccountDAOMock{}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Create(ctx, &Account{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		wired := AccountDAOMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Account], error) {
				return &query.Page[Account]{}, nil
			},
			CreateFunc: func(_ context.Context, e *Account) (*Account, error) { return e, nil },
		}
		if _, err := wired.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := wired.Create(ctx, &Account{}); err != nil {
			t.Errorf("Create = %v", err)
		}
	})

	t.Run("tax dao", func(t *testing.T) {
		bare := TaxDAOMock{}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Create(ctx, &Tax{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := bare.Update(ctx, &Tax{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		wired := TaxDAOMock{
			ListFunc:   func(_ context.Context, _ *query.Query) (*query.Page[Tax], error) { return &query.Page[Tax]{}, nil },
			CreateFunc: func(_ context.Context, e *Tax) (*Tax, error) { return e, nil },
			UpdateFunc: func(_ context.Context, e *Tax) (*Tax, error) { return e, nil },
		}
		if _, err := wired.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := wired.Create(ctx, &Tax{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := wired.Update(ctx, &Tax{}); err != nil {
			t.Errorf("Update = %v", err)
		}
	})

	t.Run("withholding dao", func(t *testing.T) {
		bare := WithholdingTaxDAOMock{}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Create(ctx, &WithholdingTax{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		wired := WithholdingTaxDAOMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[WithholdingTax], error) {
				return &query.Page[WithholdingTax]{}, nil
			},
			CreateFunc: func(_ context.Context, e *WithholdingTax) (*WithholdingTax, error) { return e, nil },
		}
		if _, err := wired.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := wired.Create(ctx, &WithholdingTax{}); err != nil {
			t.Errorf("Create = %v", err)
		}
	})
}

func TestReferenceFixtures(t *testing.T) {
	if CurrencyFixture() == nil {
		t.Error("currency = nil")
	}
	if OrganizationFixture() == nil {
		t.Error("org = nil")
	}
	if AccountFixture() == nil {
		t.Error("account = nil")
	}
	if JournalFixture() == nil {
		t.Error("journal = nil")
	}
	if TaxFixture() == nil {
		t.Error("tax = nil")
	}
	if WarehouseFixture() == nil {
		t.Error("warehouse = nil")
	}
	if StockLocationFixture() == nil {
		t.Error("location = nil")
	}
	if UnitFixture() == nil {
		t.Error("unit = nil")
	}
	if PaymentTermFixture() == nil {
		t.Error("term = nil")
	}
}

func TestReferenceModels_Tables(t *testing.T) {
	tables := map[string]string{
		(PipelineStage{}).TableName():     "pipeline_stages",
		(POSConfig{}).TableName():         "pos_configs",
		(POSPaymentAccount{}).TableName(): "pos_payment_accounts",
		(WithholdingTax{}).TableName():    "withholding_taxes",
	}
	for got, want := range tables {
		if got != want {
			t.Errorf("table = %q, want %q", got, want)
		}
	}
}

func validTermLines() []*PaymentTermLine {
	return []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100, DaysAfter: 30}}
}

func TestPaymentTermService_Create_Branches(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(terms PaymentTermDAOMock) PaymentTermService {
		return NewPaymentTermService(terms)
	}

	t.Run("creates term", func(t *testing.T) {
		svc := newSvc(PaymentTermDAOMock{})
		got, err := svc.Create(ctx, &PaymentTerm{OrganizationID: 10, Name: "Net 30"}, validTermLines())
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.Name != "Net 30" {
			t.Errorf("name = %q", got.Name)
		}
	})

	t.Run("rejects duplicate name", func(t *testing.T) {
		svc := newSvc(PaymentTermDAOMock{
			CRUDMock: dao.CRUDMock[PaymentTerm]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[PaymentTerm], error) {
					return &query.Page[PaymentTerm]{Items: []*PaymentTerm{{Base: model.Base{ID: 9}, Name: "Net 30"}}, Count: 1}, nil
				},
			},
		})
		_, err := svc.Create(ctx, &PaymentTerm{OrganizationID: 10, Name: "Net 30"}, validTermLines())
		helper.AssertError(t, err, true, ErrDuplicateTermName)
	})

	t.Run("propagates list error", func(t *testing.T) {
		svc := newSvc(PaymentTermDAOMock{
			CRUDMock: dao.CRUDMock[PaymentTerm]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[PaymentTerm], error) {
					return nil, dbErr
				},
			},
		})
		_, err := svc.Create(ctx, &PaymentTerm{OrganizationID: 10, Name: "Net 30"}, validTermLines())
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("propagates create error", func(t *testing.T) {
		svc := newSvc(PaymentTermDAOMock{
			CRUDMock: dao.CRUDMock[PaymentTerm]{
				CreateFunc: func(_ context.Context, _ *PaymentTerm) (*PaymentTerm, error) { return nil, dbErr },
			},
		})
		_, err := svc.Create(ctx, &PaymentTerm{OrganizationID: 10, Name: "Net 30"}, validTermLines())
		helper.AssertError(t, err, true, dbErr)
	})
}

func TestPaymentTermService_ValidateInOrg(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(term *PaymentTerm, findErr error) PaymentTermService {
		return NewPaymentTermService(PaymentTermDAOMock{
			CRUDMock: dao.CRUDMock[PaymentTerm]{
				FindFunc: func(_ context.Context, _ uint64) (*PaymentTerm, error) { return term, findErr },
			},
		})
	}

	t.Run("validates membership", func(t *testing.T) {
		svc := newSvc(&PaymentTerm{Base: model.Base{ID: 1}, OrganizationID: 10}, nil)
		if err := svc.ValidateTermInOrganization(ctx, 1, 10); helper.AssertError(t, err, false, nil) {
			return
		}
	})

	t.Run("rejects foreign and missing", func(t *testing.T) {
		svc := newSvc(&PaymentTerm{Base: model.Base{ID: 1}, OrganizationID: 11}, nil)
		helper.AssertError(t, svc.ValidateTermInOrganization(ctx, 1, 10), true, ErrPaymentTermNotInOrg)

		svc = newSvc(nil, nil)
		helper.AssertError(t, svc.ValidateTermInOrganization(ctx, 1, 10), true, ErrPaymentTermNotInOrg)
	})

	t.Run("propagates error", func(t *testing.T) {
		svc := newSvc(nil, dbErr)
		helper.AssertError(t, svc.ValidateTermInOrganization(ctx, 1, 10), true, dbErr)
	})
}

func TestSeedDefaultPaymentTerms(t *testing.T) {
	ctx := context.Background()

	t.Run("seeds defaults", func(t *testing.T) {
		created := 0
		svc := PaymentTermDAOMock{
			CRUDMock: dao.CRUDMock[PaymentTerm]{
				CreateFunc: func(_ context.Context, term *PaymentTerm) (*PaymentTerm, error) {
					term.ID = uint64(created + 1)
					created++
					return term, nil
				},
			},
			ReplaceLinesFunc: func(_ context.Context, _ uint64, _ []*PaymentTermLine) error { return nil },
		}
		if err := SeedDefaultPaymentTerms(ctx, svc, 10); err != nil {
			t.Fatalf("seed = %v", err)
		}
		if created == 0 {
			t.Error("expected terms to be created")
		}
	})

	t.Run("propagates create error", func(t *testing.T) {
		dbErr := errors.New("db down")
		svc := PaymentTermDAOMock{
			CRUDMock: dao.CRUDMock[PaymentTerm]{
				CreateFunc: func(_ context.Context, _ *PaymentTerm) (*PaymentTerm, error) { return nil, dbErr },
			},
		}
		helper.AssertError(t, SeedDefaultPaymentTerms(ctx, svc, 10), true, dbErr)
	})
}
