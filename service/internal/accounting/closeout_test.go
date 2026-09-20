package accounting

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type chargeMock struct {
	fee      uint64
	interest uint64
}

func (m chargeMock) FeeAccountID(_ context.Context, _ uint64) (uint64, error) {
	return m.fee, nil
}

func (m chargeMock) InterestAccountID(_ context.Context, _ uint64) (uint64, error) {
	return m.interest, nil
}

type advanceMock struct {
	customer uint64
	supplier uint64
}

func (m advanceMock) CustomerAdvanceAccountID(_ context.Context, _ uint64) (uint64, error) {
	return m.customer, nil
}

func (m advanceMock) SupplierAdvanceAccountID(_ context.Context, _ uint64) (uint64, error) {
	return m.supplier, nil
}

func TestBankMatch_PostsFeeAndInterestMoves(t *testing.T) {
	ctx := context.Background()
	statement := &BankStatement{Base: model.Base{ID: 1}, JournalID: helper.Ptr(uint64(20)), State: BankStatementStateOpen}
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return statement, nil },
		},
	}
	line := &BankStatementLine{
		Base:           model.Base{ID: 30},
		Amount:         amount.FromFloat64(10000),
		FeeAmount:      amount.FromFloat64(100),
		InterestAmount: amount.FromFloat64(50),
		ContactID:      helper.Ptr(uint64(5)),
	}
	lines := BankStatementLineDAOMock{
		ListUnreconciledByStatementFunc: func(_ context.Context, _ uint64) ([]*BankStatementLine, error) {
			return []*BankStatementLine{line}, nil
		},
	}
	payments := PaymentDAOMock{
		ListPostedByContactFunc: func(_ context.Context, _ uint64) ([]*Payment, error) {
			return []*Payment{{Base: model.Base{ID: 90}, Amount: 10050, EntryID: helper.Ptr(uint64(60))}}, nil
		},
	}
	journals := dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 20}, OrganizationID: 10, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
		},
	}
	posted := 0
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
			posted++
			return &JournalEntry{Base: model.Base{ID: uint64(100 + posted)}}, nil
		},
	}
	svc := NewBankStatementService(statements, lines, payments, journals, TransactionerMock{}).WithPosting(poster, chargeMock{fee: 6100, interest: 7100})
	unreconciled, err := svc.Match(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(unreconciled) != 0 {
		t.Fatalf("unreconciled = %d, want 0", len(unreconciled))
	}
	if posted != 2 {
		t.Errorf("posted movements = %d, want 2 (fee + interest)", posted)
	}
	if line.FeeEntryID == nil || line.InterestEntryID == nil {
		t.Errorf("fee/interest entry ids not set: %+v", line)
	}
}

func TestBankMatch_SkipsPostingWithoutResolver(t *testing.T) {
	ctx := context.Background()
	statement := &BankStatement{Base: model.Base{ID: 1}, JournalID: helper.Ptr(uint64(20)), State: BankStatementStateOpen}
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return statement, nil },
		},
	}
	line := &BankStatementLine{Base: model.Base{ID: 30}, Amount: amount.FromFloat64(100), FeeAmount: amount.FromFloat64(10), ContactID: helper.Ptr(uint64(5))}
	lines := BankStatementLineDAOMock{
		ListUnreconciledByStatementFunc: func(_ context.Context, _ uint64) ([]*BankStatementLine, error) {
			return []*BankStatementLine{line}, nil
		},
	}
	payments := PaymentDAOMock{
		ListPostedByContactFunc: func(_ context.Context, _ uint64) ([]*Payment, error) {
			return []*Payment{{Base: model.Base{ID: 90}, Amount: 110, EntryID: helper.Ptr(uint64(60))}}, nil
		},
	}
	svc := NewBankStatementService(statements, lines, payments, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})
	unreconciled, err := svc.Match(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(unreconciled) != 0 {
		t.Errorf("unreconciled = %d, want 0 (gross 110 matches without posting)", len(unreconciled))
	}
	if line.FeeEntryID != nil {
		t.Errorf("fee entry should not post without resolver")
	}
}

func TestPaymentCrossCurrency_AdvancePostsBalanced(t *testing.T) {
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
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
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
		WithOrganizations(orgMock{base: "IDR"}).
		WithAdvanceAccounts(advanceMock{customer: 2200, supplier: 2300})
	if _, err := svc.Create(ctx, CreatePaymentRequest{
		OrganizationID: 10,
		ContactID:      5,
		JournalID:      20,
		Amount:         1720000,
		CurrencyCode:   helper.Ptr("IDR"),
		Date:           payDate,
		InvoiceIDs:     []uint64{11},
		AllowAdvance:   true,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	debits := amount.Zero()
	credits := amount.Zero()
	foundAdvance := false
	for _, line := range captured {
		debits = debits.Add(line.Debit)
		credits = credits.Add(line.Credit)
		if line.AccountID == 2200 && line.Credit.Equal(amount.FromFloat64(100000)) {
			foundAdvance = true
		}
	}
	if !amount.IsBalanced(debits, credits, 4) {
		t.Errorf("unbalanced debits %v credits %v", debits, credits)
	}
	if !foundAdvance {
		t.Errorf("lines = %+v, want customer advance 100000 credit", captured)
	}
}

func TestPaymentCrossCurrency_RejectsOverpayWithoutAdvance(t *testing.T) {
	ctx := context.Background()
	invDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	payDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	open := &Invoice{Base: model.Base{ID: 11}, State: InvoiceStatePosted, AmountResidual: amount.FromFloat64(100), CurrencyCode: helper.Ptr("USD"), InvoiceDate: &invDate}
	payments, invoices, accounts, journals := paymentTestFixtures(open)
	svc := NewPaymentService(payments, invoices, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{}).
		WithFxResolvers(rateStubForCross(), fxMock{gain: 9001, loss: 9002}).
		WithOrganizations(orgMock{base: "IDR"})
	_, err := svc.Create(ctx, CreatePaymentRequest{
		OrganizationID: 10, ContactID: 5, JournalID: 20,
		Amount: 1720000, CurrencyCode: helper.Ptr("IDR"), Date: payDate, InvoiceIDs: []uint64{11},
	})
	if err == nil || err != ErrOverAllocation {
		t.Errorf("err = %v, want ErrOverAllocation", err)
	}
}

func TestRollupTrialBalance_AggregatesParent(t *testing.T) {
	child := TrialBalanceLine{AccountID: 2, AccountCode: "1101", AccountName: "Cash drawer", AccountType: "cash", Debit: amount.FromFloat64(100), Balance: amount.FromFloat64(100)}
	accounts := []*reference.Account{
		{Base: model.Base{ID: 1}, Code: "1100", Name: "Cash", Type: "cash"},
		{Base: model.Base{ID: 2}, Code: "1101", Name: "Cash drawer", Type: "cash", ParentID: helper.Ptr(uint64(1))},
	}
	rolled := RollupTrialBalance([]TrialBalanceLine{child}, accounts)
	if len(rolled) != 2 {
		t.Fatalf("rolled = %d, want 2 (child + parent)", len(rolled))
	}
	var parent *TrialBalanceLine
	for i := range rolled {
		if rolled[i].AccountID == 1 {
			parent = &rolled[i]
		}
	}
	if parent == nil {
		t.Fatalf("parent missing in %+v", rolled)
	}
	if !parent.Debit.Equal(amount.FromFloat64(100)) {
		t.Errorf("parent debit = %v, want 100", parent.Debit)
	}
}

func TestIntegrityCheckPeriod_PassesConsistent(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	periods := TaxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{FindFunc: func(_ context.Context, id uint64) (*TaxPeriod, error) {
			return &TaxPeriod{Base: model.Base{ID: id}, OrganizationID: 10, DateStart: &start, DateEnd: &end, State: TaxPeriodStateOpen}, nil
		}},
	}
	tb := []TrialBalanceLine{
		{AccountID: 1, AccountCode: "1100", AccountName: "Cash", AccountType: "cash", Debit: amount.FromFloat64(12000), Balance: amount.FromFloat64(12000)},
		{AccountID: 2, AccountCode: "3100", AccountName: "Capital", AccountType: "equity", Credit: amount.FromFloat64(10000), Balance: amount.FromFloat64(-10000)},
		{AccountID: 3, AccountCode: "4100", AccountName: "Revenue", AccountType: "income", Credit: amount.FromFloat64(2000), Balance: amount.FromFloat64(-2000)},
	}
	trialDAO := TrialBalanceDAOMock{
		GenerateAsOfFunc: func(_ context.Context, _ uint64, _ string, _ bool) ([]TrialBalanceLine, error) {
			return tb, nil
		},
		GenerateByPeriodFunc: func(_ context.Context, _, _ uint64) ([]TrialBalanceLine, error) {
			return []TrialBalanceLine{
				{AccountID: 1, AccountCode: "1100", AccountName: "Cash", AccountType: "cash", Debit: amount.FromFloat64(2000), Balance: amount.FromFloat64(2000)},
				{AccountID: 3, AccountCode: "4100", AccountName: "Revenue", AccountType: "income", Credit: amount.FromFloat64(2000), Balance: amount.FromFloat64(-2000)},
			}, nil
		},
	}
	cashDAO := CashFlowDAOMock{
		GetOpeningCashFunc: func(_ context.Context, _ uint64, _ time.Time) (float64, error) { return 10000, nil },
		GenerateByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]CashFlowLine, error) {
			return []CashFlowLine{{AccountID: 1, Amount: 2000, CashFlowSection: CashFlowOperating}}, nil
		},
	}
	equityDAO := EquityDAOMock{
		OpeningEquityBalanceFunc: func(_ context.Context, _ uint64, _ time.Time) ([]EquityMovement, error) {
			return []EquityMovement{{AccountID: 2, Amount: 10000}}, nil
		},
		EquityBalancesByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]EquityMovement, error) {
			return []EquityMovement{}, nil
		},
		NetIncomeForPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) (float64, error) { return 2000, nil },
	}
	svc := NewReportIntegrityService(trialDAO, cashDAO, equityDAO, periods, AccountLookupMock{})
	report, err := svc.CheckPeriod(ctx, 10, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Checks) != 6 {
		t.Fatalf("checks = %d, want 6", len(report.Checks))
	}
	for _, c := range report.Checks {
		if c.Name == "trial_balance_balanced" && !c.Passed {
			t.Errorf("TB balanced check failed: %+v", c)
		}
		if c.Name == "cash_to_ledger" && !c.Passed {
			t.Errorf("cash to ledger failed: %+v", c)
		}
		if c.Name == "equity_income_to_ledger" && !c.Passed {
			t.Errorf("equity income check failed: %+v", c)
		}
	}
	_ = report.Passed
}

func TestMoveImmutable_BlocksPostedUpdate(t *testing.T) {
	ctx := context.Background()
	d := journalEntryDAO{}
	_, err := d.Update(ctx, &JournalEntry{State: EntryStatePosted})
	if err == nil || err != ErrEntryImmutable {
		t.Errorf("err = %v, want ErrEntryImmutable", err)
	}
}

func TestMoveLineImmutable_BlocksAmountChange(t *testing.T) {
	ctx := context.Background()
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(`SELECT .* FROM "journal_lines".*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "entry_id", "debit", "credit", "account_id"}).AddRow(7, 60, "100", "0", 1100))
	mock.ExpectQuery(`SELECT .* FROM "journal_entrys".*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "state"}).AddRow(60, EntryStatePosted))
	d := NewJournalLineDAO(db)
	_, err := d.Update(ctx, &JournalLine{Base: model.Base{ID: 7}, EntryID: 60, AccountID: 1100, Debit: amount.FromFloat64(200)})
	if err == nil || err != ErrEntryImmutable {
		t.Errorf("err = %v, want ErrEntryImmutable", err)
	}
	query.AssertDBMockDone(t, mock)
}

func TestMoveLineImmutable_AllowsReconcileFlag(t *testing.T) {
	ctx := context.Background()
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM "journal_lines".*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "entry_id", "debit", "credit", "account_id", "reconciled"}).AddRow(7, 60, "100", "0", 1100, false))
	mock.ExpectQuery(`SELECT .* FROM "journal_entrys".*`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "state"}).AddRow(60, EntryStatePosted))
	mock.ExpectExec(`UPDATE "journal_lines".*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	d := NewJournalLineDAO(db)
	tx := db.Begin()
	updated, err := d.UpdateTx(ctx, tx, &JournalLine{Base: model.Base{ID: 7}, EntryID: 60, AccountID: 1100, Debit: amount.FromFloat64(100), Reconciled: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated.Reconciled {
		t.Errorf("reconciled not persisted")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}
	query.AssertDBMockDone(t, mock)
}
