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
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type rateMock struct {
	rates map[string]float64
}

func (m rateMock) Rate(_ context.Context, currencyCode string, _ uint64, _ amount.RateType, _ time.Time) (amount.Amount, error) {
	if v, ok := m.rates[currencyCode]; ok {
		return amount.FromFloat64(v), nil
	}
	return amount.Amount{}, ErrRateNotFound
}

type fxMock struct {
	gain uint64
	loss uint64
}

func (m fxMock) FxGainAccountID(_ context.Context, _ uint64) (uint64, error) {
	return m.gain, nil
}

func (m fxMock) FxLossAccountID(_ context.Context, _ uint64) (uint64, error) {
	return m.loss, nil
}

type orgMock struct {
	base string
}

func (m orgMock) Find(_ context.Context, id uint64) (*reference.Organization, error) {
	return &reference.Organization{Base: model.Base{ID: id}, BaseCurrency: m.base}, nil
}

func TestStatementGross_FeeAddedBack(t *testing.T) {
	line := &BankStatementLine{
		Amount:         amount.FromFloat64(9900),
		FeeAmount:      amount.FromFloat64(100),
		InterestAmount: amount.Zero(),
	}
	if got := StatementGross(line).Float64(); got != 10000 {
		t.Errorf("gross = %v, want 10000", got)
	}
}

func TestStatementGross_InterestSubtracted(t *testing.T) {
	line := &BankStatementLine{
		Amount:         amount.FromFloat64(1050),
		InterestAmount: amount.FromFloat64(50),
	}
	if got := StatementGross(line).Float64(); got != 1000 {
		t.Errorf("gross = %v, want 1000", got)
	}
}

func TestBankMatch_FeeGrossMatchesPayment(t *testing.T) {
	ctx := context.Background()
	statement := &BankStatement{Base: model.Base{ID: 1}, State: BankStatementStateOpen}
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return statement, nil },
		},
	}
	line := &BankStatementLine{
		Base:      model.Base{ID: 30},
		Amount:    amount.FromFloat64(9900),
		FeeAmount: amount.FromFloat64(100),
		ContactID: helper.Ptr(uint64(5)),
	}
	lines := BankStatementLineDAOMock{
		ListUnreconciledByStatementFunc: func(_ context.Context, _ uint64) ([]*BankStatementLine, error) {
			return []*BankStatementLine{line}, nil
		},
	}
	payments := PaymentDAOMock{
		ListPostedByContactFunc: func(_ context.Context, _ uint64) ([]*Payment, error) {
			return []*Payment{{Base: model.Base{ID: 90}, Amount: 10000, EntryID: helper.Ptr(uint64(60))}}, nil
		},
	}
	svc := NewBankStatementService(statements, lines, payments, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})
	unreconciled, err := svc.Match(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(unreconciled) != 0 {
		t.Errorf("unreconciled = %d, want 0 (fee gross 10000 matches)", len(unreconciled))
	}
}

func TestBankMatch_CurrencyMismatchSkipsAmount(t *testing.T) {
	ctx := context.Background()
	statement := &BankStatement{Base: model.Base{ID: 1}, State: BankStatementStateOpen}
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return statement, nil },
		},
	}
	line := &BankStatementLine{
		Base:         model.Base{ID: 30},
		Amount:       amount.FromFloat64(100),
		CurrencyCode: helper.Ptr("USD"),
		ContactID:    helper.Ptr(uint64(5)),
	}
	lines := BankStatementLineDAOMock{
		ListUnreconciledByStatementFunc: func(_ context.Context, _ uint64) ([]*BankStatementLine, error) {
			return []*BankStatementLine{line}, nil
		},
	}
	payments := PaymentDAOMock{
		ListPostedByContactFunc: func(_ context.Context, _ uint64) ([]*Payment, error) {
			return []*Payment{{Base: model.Base{ID: 90}, Amount: 100, CurrencyCode: helper.Ptr("IDR"), EntryID: helper.Ptr(uint64(60))}}, nil
		},
	}
	svc := NewBankStatementService(statements, lines, payments, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})
	unreconciled, err := svc.Match(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(unreconciled) != 1 {
		t.Errorf("unreconciled = %d, want 1 (currency differs)", len(unreconciled))
	}
}

func paymentTestFixtures(open *Invoice) (PaymentDAO, InvoiceDAO, AccountLookupMock, dao.CRUDMock[reference.Journal]) {
	payments := PaymentDAOMock{
		CreateWithAllocationsTxFunc: func(_ context.Context, _ *gorm.DB, payment *Payment, allocations []*PaymentAllocation) (*Payment, error) {
			payment.ID = 1
			return payment, nil
		},
	}
	invoices := InvoiceDAOMock{
		CRUDMock: dao.CRUDMock[Invoice]{
			FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) { return open, nil },
		},
	}
	accounts := AccountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
		},
	}
	journals := dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
		},
	}
	return payments, invoices, accounts, journals
}

func TestPaymentCrossCurrency_InboundPostsGainBalanced(t *testing.T) {
	ctx := context.Background()
	invDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	payDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	open := &Invoice{Base: model.Base{ID: 11}, State: InvoiceStatePosted, PaymentState: PaymentStateNotPaid, AmountResidual: amount.FromFloat64(100), CurrencyCode: helper.Ptr("USD"), InvoiceDate: &invDate}
	payments, invoices, accounts, journals := paymentTestFixtures(open)
	var captured []PostingLine
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
			captured = request.Lines
			return &JournalEntry{Base: model.Base{ID: 60}}, nil
		},
	}
	svc := NewPaymentService(payments, invoices, poster, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{}).
		WithFxResolvers(rateMock{rates: map[string]float64{"USD": 16200}}, fxMock{gain: 9001, loss: 9002}).
		WithOrganizations(orgMock{base: "IDR"})
	_ = svc
	svcCross := NewPaymentService(payments, invoices, poster, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{}).
		WithFxResolvers(rateStubForCross(), fxMock{gain: 9001, loss: 9002}).
		WithOrganizations(orgMock{base: "IDR"})
	payment, err := svcCross.Create(ctx, CreatePaymentRequest{
		OrganizationID: 10,
		ContactID:      5,
		JournalID:      20,
		Amount:         1620000,
		CurrencyCode:   helper.Ptr("IDR"),
		Date:           payDate,
		InvoiceIDs:     []uint64{11},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.Type != PaymentTypeInbound {
		t.Errorf("type = %s, want inbound", payment.Type)
	}
	if len(captured) != 3 {
		t.Fatalf("lines = %d, want 3 (bank/ar/gain)", len(captured))
	}
	debits := amount.Zero()
	credits := amount.Zero()
	for _, line := range captured {
		debits = debits.Add(line.Debit)
		credits = credits.Add(line.Credit)
	}
	if !amount.IsBalanced(debits, credits, 4) {
		t.Errorf("unbalanced debits %v credits %v", debits, credits)
	}
	foundGain := false
	for _, line := range captured {
		if line.AccountID == 9001 && line.Credit.Equal(amount.FromFloat64(20000)) {
			foundGain = true
		}
	}
	if !foundGain {
		t.Errorf("lines = %+v, want FX gain 20000 credit", captured)
	}
}

type datedRateMock struct {
	invRate float64
	payRate float64
	invDate time.Time
}

func (m datedRateMock) Rate(_ context.Context, _ string, _ uint64, _ amount.RateType, date time.Time) (amount.Amount, error) {
	if date.Equal(m.invDate) || date.Before(m.invDate.Add(24*time.Hour)) && date.After(m.invDate.Add(-24*time.Hour)) {
		return amount.FromFloat64(m.invRate), nil
	}
	return amount.FromFloat64(m.payRate), nil
}

func rateStubForCross() CurrencyRateResolver {
	return datedRateMock{
		invRate: 16000,
		payRate: 16200,
		invDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestPaymentCrossCurrency_OutboundPostsLossBalanced(t *testing.T) {
	ctx := context.Background()
	invDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	payDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	open := &Invoice{Base: model.Base{ID: 11}, State: InvoiceStatePosted, PaymentState: PaymentStateNotPaid, AmountResidual: amount.FromFloat64(100), CurrencyCode: helper.Ptr("USD"), InvoiceDate: &invDate}
	payments := PaymentDAOMock{
		CreateWithAllocationsTxFunc: func(_ context.Context, _ *gorm.DB, payment *Payment, allocations []*PaymentAllocation) (*Payment, error) {
			payment.ID = 1
			return payment, nil
		},
	}
	invoices := InvoiceDAOMock{
		CRUDMock: dao.CRUDMock[Invoice]{
			FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) { return open, nil },
		},
	}
	accounts := AccountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 2100}, OrganizationID: 10, Active: true}}}, nil
		},
	}
	journals := dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
		},
	}
	var captured []PostingLine
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
			captured = request.Lines
			return &JournalEntry{Base: model.Base{ID: 60}}, nil
		},
	}
	svc := NewPaymentService(payments, invoices, poster, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{}).
		WithFxResolvers(rateStubForCross(), fxMock{gain: 9001, loss: 9002}).
		WithOrganizations(orgMock{base: "IDR"})
	if _, err := svc.CreateOutbound(ctx, CreatePaymentRequest{
		OrganizationID: 10,
		ContactID:      5,
		JournalID:      20,
		Amount:         1620000,
		CurrencyCode:   helper.Ptr("IDR"),
		Date:           payDate,
		InvoiceIDs:     []uint64{11},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured) != 3 {
		t.Fatalf("lines = %d, want 3 (payable/bank/loss)", len(captured))
	}
	debits := amount.Zero()
	credits := amount.Zero()
	for _, line := range captured {
		debits = debits.Add(line.Debit)
		credits = credits.Add(line.Credit)
	}
	if !amount.IsBalanced(debits, credits, 4) {
		t.Errorf("unbalanced debits %v credits %v", debits, credits)
	}
	foundLoss := false
	for _, line := range captured {
		if line.AccountID == 9002 && line.Debit.Equal(amount.FromFloat64(20000)) {
			foundLoss = true
		}
	}
	if !foundLoss {
		t.Errorf("lines = %+v, want FX loss 20000 debit", captured)
	}
}

func TestInvoiceFunctionalConversion(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var captured []PostingLine
	invoices := InvoiceDAOMock{
		CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, lines []*InvoiceLine, taxes []*InvoiceTax) (*Invoice, error) {
			invoice.ID = 1
			return invoice, nil
		},
	}
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
			captured = request.Lines
			return &JournalEntry{Base: model.Base{ID: 50}}, nil
		},
	}
	accounts := AccountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
		},
	}
	taxes := dao.CRUDMock[reference.Tax]{}
	sequences := sequence.NewSequenceService(SequenceDAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "INV/00001"}, nil
		},
	})
	svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, poster, accounts, taxes, sequences, TransactionerMock{}).
		WithCurrencyResolvers(orgMock{base: "IDR"}, rateMock{rates: map[string]float64{"USD": 16000}})
	invoice, err := svc.Create(ctx, CreateInvoiceRequest{
		OrganizationID: 10,
		JournalID:      20,
		ContactID:      5,
		Date:           date,
		CurrencyCode:   helper.Ptr("USD"),
		Lines:          []InvoiceLineRequest{{Description: "Widget", Qty: 1, UnitPrice: 100, AccountID: 4100}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if invoice.AmountTotal.Float64() != 100 {
		t.Errorf("foreign total = %v, want 100", invoice.AmountTotal)
	}
	if len(captured) != 2 {
		t.Fatalf("lines = %d, want 2", len(captured))
	}
	if !captured[0].Debit.Equal(amount.FromFloat64(1600000)) {
		t.Errorf("AR debit = %v, want 1600000 functional", captured[0].Debit)
	}
	if captured[0].AmountCurrency != 100 {
		t.Errorf("amount currency = %v, want 100", captured[0].AmountCurrency)
	}
}

func TestTrialBalanceBalancedFlag(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	periods := TaxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{FindFunc: func(_ context.Context, id uint64) (*TaxPeriod, error) {
			return &TaxPeriod{Base: model.Base{ID: id}, OrganizationID: 10, DateStart: &start, DateEnd: &end, State: TaxPeriodStateOpen}, nil
		}},
	}
	balancedDAO := TrialBalanceDAOMock{
		GenerateByPeriodFunc: func(_ context.Context, _, _ uint64) ([]TrialBalanceLine, error) {
			return []TrialBalanceLine{
				{AccountID: 1, Debit: amount.FromFloat64(5000), Credit: amount.Zero()},
				{AccountID: 2, Debit: amount.Zero(), Credit: amount.FromFloat64(5000)},
			}, nil
		},
	}
	svc := NewTrialBalanceService(balancedDAO, periods)
	tb, err := svc.Generate(ctx, 10, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tb.Balanced {
		t.Errorf("balanced = false, want true")
	}
	if !tb.Difference.IsZero() {
		t.Errorf("difference = %v, want 0", tb.Difference)
	}
	unbalancedDAO := TrialBalanceDAOMock{
		GenerateByPeriodFunc: func(_ context.Context, _, _ uint64) ([]TrialBalanceLine, error) {
			return []TrialBalanceLine{
				{AccountID: 1, Debit: amount.FromFloat64(5000), Credit: amount.Zero()},
				{AccountID: 2, Debit: amount.Zero(), Credit: amount.FromFloat64(3000)},
			}, nil
		},
	}
	svc = NewTrialBalanceService(unbalancedDAO, periods)
	tb, err = svc.Generate(ctx, 10, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tb.Balanced {
		t.Errorf("balanced = true, want false")
	}
}

func TestEquityBalancedFlag(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	periods := TaxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{FindFunc: func(_ context.Context, id uint64) (*TaxPeriod, error) {
			return &TaxPeriod{Base: model.Base{ID: id}, OrganizationID: 10, DateStart: &start, DateEnd: &end, State: TaxPeriodStateOpen}, nil
		}},
	}
	equityDAO := EquityDAOMock{
		OpeningEquityBalanceFunc: func(_ context.Context, _ uint64, _ time.Time) ([]EquityMovement, error) {
			return []EquityMovement{{AccountID: 100, Amount: 50000}}, nil
		},
		EquityBalancesByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]EquityMovement, error) {
			return []EquityMovement{{AccountID: 100, Amount: 10000}}, nil
		},
		NetIncomeForPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) (float64, error) {
			return 2500, nil
		},
	}
	svc := NewEquityService(equityDAO, periods)
	eq, err := svc.Generate(ctx, 10, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eq.Balanced {
		t.Errorf("balanced = false, want true (60000 = 50000+10000)")
	}
}

func TestCashFlowBalancedFlag(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	daoMock := CashFlowDAOMock{
		GetOpeningCashFunc: func(_ context.Context, _ uint64, _ time.Time) (float64, error) {
			return 10000, nil
		},
		GenerateByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]CashFlowLine, error) {
			return []CashFlowLine{{AccountID: 1, Amount: 2000, CashFlowSection: CashFlowOperating}}, nil
		},
	}
	svc := NewCashFlowService(daoMock)
	cf, err := svc.Generate(ctx, 10, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cf.Balanced {
		t.Errorf("balanced = false, want true")
	}
	if cf.ClosingCash != 12000 {
		t.Errorf("closing = %v, want 12000", cf.ClosingCash)
	}
}
