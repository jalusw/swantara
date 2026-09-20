package accounting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestTaxReturnService_Create_ComputesNetPayable(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	period := &TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 10, DateStart: &start, DateEnd: &end}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{
			FindFunc: func(_ context.Context, _ uint64) (*TaxPeriod, error) { return period, nil },
		},
	}
	perType := map[string]float64{InvoiceTypeCustomerInvoice: 110, InvoiceTypeSupplierBill: 40}
	invoiceTaxes := InvoiceTaxDAOMock{
		SumTaxByPeriodFunc: func(_ context.Context, _ uint64, invoiceType string, _, _ time.Time) (float64, error) {
			return perType[invoiceType], nil
		},
	}
	var created *TaxReturn
	returns := TaxReturnDAOMock{
		CRUDMock: dao.CRUDMock[TaxReturn]{
			CreateFunc: func(_ context.Context, taxReturn *TaxReturn) (*TaxReturn, error) {
				taxReturn.ID = 1
				created = taxReturn
				return taxReturn, nil
			},
		},
	}
	svc := NewTaxReturnService(returns, periods, invoiceTaxes, TransactionerMock{})

	taxReturn, err := svc.Create(ctx, CreateTaxReturnRequest{OrganizationID: 10, PeriodID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.OutputTax != 110 || created.InputTax != 40 || created.NetPayable != 70 {
		t.Errorf("tax return = %v/%v/%v, want 110/40/70", created.OutputTax, created.InputTax, created.NetPayable)
	}
	if created.State != TaxReturnStateDraft {
		t.Errorf("state = %s, want draft", created.State)
	}
	if taxReturn.ID != 1 {
		t.Errorf("id = %d, want 1", taxReturn.ID)
	}
}

func TestTaxReturnService_Create_RejectsDuplicate(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	period := &TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 10, DateStart: &start, DateEnd: &end}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{
			FindFunc: func(_ context.Context, _ uint64) (*TaxPeriod, error) { return period, nil },
		},
	}
	returns := TaxReturnDAOMock{
		FindByPeriodFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) {
			return &TaxReturn{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := NewTaxReturnService(returns, periods, InvoiceTaxDAOMock{}, TransactionerMock{})

	_, err := svc.Create(ctx, CreateTaxReturnRequest{OrganizationID: 10, PeriodID: 1})
	if helper.AssertError(t, err, true, ErrTaxReturnExists) {
		return
	}
}

func TestTaxReturnService_File_MarksFiled(t *testing.T) {
	ctx := context.Background()
	taxReturn := &TaxReturn{Base: model.Base{ID: 1}, State: TaxReturnStateDraft}
	returns := TaxReturnDAOMock{
		CRUDMock: dao.CRUDMock[TaxReturn]{
			FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) { return taxReturn, nil },
			UpdateFunc: func(_ context.Context, updated *TaxReturn) (*TaxReturn, error) {
				return updated, nil
			},
		},
	}
	svc := NewTaxReturnService(returns, taxPeriodDAOMock{}, InvoiceTaxDAOMock{}, TransactionerMock{})

	filed, err := svc.File(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filed.State != TaxReturnStateFiled || filed.FiledAt == nil {
		t.Errorf("filed = state %s filed_at %v, want filed/set", filed.State, filed.FiledAt)
	}
}

func TestTaxReturnService_Pay_MarksPaid(t *testing.T) {
	ctx := context.Background()
	taxReturn := &TaxReturn{Base: model.Base{ID: 1}, State: TaxReturnStateFiled}
	returns := TaxReturnDAOMock{
		CRUDMock: dao.CRUDMock[TaxReturn]{
			FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) { return taxReturn, nil },
			UpdateFunc: func(_ context.Context, updated *TaxReturn) (*TaxReturn, error) {
				return updated, nil
			},
		},
	}
	svc := NewTaxReturnService(returns, taxPeriodDAOMock{}, InvoiceTaxDAOMock{}, TransactionerMock{})

	paid, err := svc.Pay(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if paid.State != TaxReturnStatePaid {
		t.Errorf("state = %s, want paid", paid.State)
	}
}

func TestTaxReturnService_Pay_PostsSettlementMove(t *testing.T) {
	ctx := context.Background()
	taxReturn := &TaxReturn{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(10)), State: TaxReturnStateFiled, NetPayable: 70}
	returns := TaxReturnDAOMock{
		CRUDMock: dao.CRUDMock[TaxReturn]{
			FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) { return taxReturn, nil },
			UpdateFunc: func(_ context.Context, updated *TaxReturn) (*TaxReturn, error) {
				return updated, nil
			},
		},
	}
	var posted PostRequest
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
			posted = request
			return &JournalEntry{Base: model.Base{ID: 9}}, nil
		},
	}
	journals := dao.CRUDMock[reference.Journal]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Journal], error) {
			return &query.Page[reference.Journal]{Items: []*reference.Journal{
				{Base: model.Base{ID: 20}, OrganizationID: 10, Type: "bank", DefaultAccountID: helper.Ptr(uint64(1000))},
			}, Count: 1}, nil
		},
	}
	accounts := AccountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{
				{Base: model.Base{ID: 9000}, OrganizationID: 10, Type: AccountTypeTax, Active: true},
			}, Count: 1}, nil
		},
	}
	svc := NewTaxReturnService(returns, taxPeriodDAOMock{}, InvoiceTaxDAOMock{}, TransactionerMock{}).
		WithSettlement(poster, journals, accounts)

	paid, err := svc.Pay(ctx, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if paid.State != TaxReturnStatePaid {
		t.Errorf("state = %s, want paid", paid.State)
	}
	if len(posted.Lines) != 2 {
		t.Fatalf("settlement lines = %d, want 2", len(posted.Lines))
	}
	if posted.Lines[0].AccountID != 9000 || !posted.Lines[0].Debit.Equal(amount.FromFloat64(70)) {
		t.Errorf("debit leg = %d/%v, want 9000/70", posted.Lines[0].AccountID, posted.Lines[0].Debit)
	}
	if posted.Lines[1].AccountID != 1000 || !posted.Lines[1].Credit.Equal(amount.FromFloat64(70)) {
		t.Errorf("credit leg = %d/%v, want 1000/70", posted.Lines[1].AccountID, posted.Lines[1].Credit)
	}
	if posted.OriginType != OriginTypeTaxPayment || posted.OriginID != 3 {
		t.Errorf("origin = %s/%d, want tax_payment/3", posted.OriginType, posted.OriginID)
	}
}

func TestTaxReturnService_Pay_RejectsUnfiled(t *testing.T) {
	ctx := context.Background()
	taxReturn := &TaxReturn{Base: model.Base{ID: 1}, State: TaxReturnStateDraft}
	returns := TaxReturnDAOMock{
		CRUDMock: dao.CRUDMock[TaxReturn]{
			FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) { return taxReturn, nil },
		},
	}
	svc := NewTaxReturnService(returns, taxPeriodDAOMock{}, InvoiceTaxDAOMock{}, TransactionerMock{})

	_, err := svc.Pay(ctx, 1)
	if helper.AssertError(t, err, true, ErrTaxReturnNotFiled) {
		return
	}
}

func TestTaxReturnService_Draft_MarksDraft(t *testing.T) {
	ctx := context.Background()
	taxReturn := &TaxReturn{Base: model.Base{ID: 1}, State: TaxReturnStateFiled}
	returns := TaxReturnDAOMock{
		CRUDMock: dao.CRUDMock[TaxReturn]{
			FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) { return taxReturn, nil },
			UpdateFunc: func(_ context.Context, updated *TaxReturn) (*TaxReturn, error) {
				return updated, nil
			},
		},
	}
	svc := NewTaxReturnService(returns, taxPeriodDAOMock{}, InvoiceTaxDAOMock{}, TransactionerMock{})

	drafted, err := svc.Draft(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if drafted.State != TaxReturnStateDraft {
		t.Errorf("state = %s, want draft", drafted.State)
	}
}

func TestTaxReturnService_Draft_RejectsPaid(t *testing.T) {
	ctx := context.Background()
	taxReturn := &TaxReturn{Base: model.Base{ID: 1}, State: TaxReturnStatePaid}
	returns := TaxReturnDAOMock{
		CRUDMock: dao.CRUDMock[TaxReturn]{
			FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) { return taxReturn, nil },
		},
	}
	svc := NewTaxReturnService(returns, taxPeriodDAOMock{}, InvoiceTaxDAOMock{}, TransactionerMock{})

	_, err := svc.Draft(ctx, 1)
	if helper.AssertError(t, err, true, ErrTaxReturnPaid) {
		return
	}
}

func TestTaxReturnService_Draft_NotFound(t *testing.T) {
	ctx := context.Background()
	returns := TaxReturnDAOMock{
		CRUDMock: dao.CRUDMock[TaxReturn]{
			FindFunc: func(_ context.Context, _ uint64) (*TaxReturn, error) { return nil, nil },
		},
	}
	svc := NewTaxReturnService(returns, taxPeriodDAOMock{}, InvoiceTaxDAOMock{}, TransactionerMock{})

	_, err := svc.Draft(ctx, 1)
	if helper.AssertError(t, err, true, ErrTaxReturnNotFound) {
		return
	}
}

type taxPeriodDAOMock struct {
	dao.CRUDMock[TaxPeriod]
}

func (m taxPeriodDAOMock) ListByOrganization(ctx context.Context, organizationID uint64) ([]*TaxPeriod, error) {
	return nil, nil
}

func (m taxPeriodDAOMock) FindByDate(ctx context.Context, organizationID uint64, date time.Time) (*TaxPeriod, error) {
	return nil, nil
}

func (m taxPeriodDAOMock) UpdateTx(ctx context.Context, _ *gorm.DB, period *TaxPeriod) (*TaxPeriod, error) {
	return m.Update(ctx, period)
}
