package accounting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestBankStatementService_Create_StoresStatementWithLines(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)

	var createdLines []*BankStatementLine
	statements := BankStatementDAOMock{
		CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, statement *BankStatement, lines []*BankStatementLine) (*BankStatement, error) {
			statement.ID = 1
			for _, line := range lines {
				line.StatementID = statement.ID
			}
			createdLines = lines
			return statement, nil
		},
	}
	journals := dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 20}}, nil
		},
	}
	svc := NewBankStatementService(statements, BankStatementLineDAOMock{}, PaymentDAOMock{}, journals, TransactionerMock{})

	statement, err := svc.Create(ctx, CreateBankStatementRequest{
		OrganizationID: 10,
		JournalID:      20,
		Date:           date,
		BalanceStart:   1000,
		BalanceEnd:     1250,
		Lines: []BankStatementLineRequest{
			{Date: &date, Amount: 250, ContactID: helper.Ptr(uint64(5))},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statement.State != BankStatementStateOpen {
		t.Errorf("state = %s, want open", statement.State)
	}
	if len(createdLines) != 1 || createdLines[0].Amount.Float64() != 250 || createdLines[0].StatementID != 1 {
		t.Errorf("lines = %+v, want single 250 on statement 1", createdLines)
	}
}

func TestBankStatementService_Create_RejectsWithoutLines(t *testing.T) {
	ctx := context.Background()
	svc := NewBankStatementService(BankStatementDAOMock{}, BankStatementLineDAOMock{}, PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})

	_, err := svc.Create(ctx, CreateBankStatementRequest{OrganizationID: 10, JournalID: 20})
	if helper.AssertError(t, err, true, ErrStatementNoLines) {
		return
	}
}

func TestBankStatementService_Match_MatchesToPostedPayment(t *testing.T) {
	ctx := context.Background()
	statement := &BankStatement{Base: model.Base{ID: 1}, State: BankStatementStateOpen}
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return statement, nil },
		},
	}
	line := &BankStatementLine{Base: model.Base{ID: 30}, StatementID: 1, Amount: amount.FromFloat64(250), ContactID: helper.Ptr(uint64(5))}
	lines := BankStatementLineDAOMock{
		ListUnreconciledByStatementFunc: func(_ context.Context, _ uint64) ([]*BankStatementLine, error) {
			return []*BankStatementLine{line}, nil
		},
	}
	payments := PaymentDAOMock{
		ListPostedByContactFunc: func(_ context.Context, _ uint64) ([]*Payment, error) {
			return []*Payment{{Base: model.Base{ID: 90}, Amount: 250, EntryID: helper.Ptr(uint64(60))}}, nil
		},
	}
	svc := NewBankStatementService(statements, lines, payments, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})

	unreconciled, err := svc.Match(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(unreconciled) != 0 {
		t.Errorf("unreconciled = %d, want 0", len(unreconciled))
	}
	if !line.Reconciled || line.PaymentID == nil || *line.PaymentID != 90 {
		t.Errorf("line = reconciled %v payment %v, want true/90", line.Reconciled, line.PaymentID)
	}
}

func TestBankStatementService_Match_FlagsUnmatchedLine(t *testing.T) {
	ctx := context.Background()
	statement := &BankStatement{Base: model.Base{ID: 1}, State: BankStatementStateOpen}
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return statement, nil },
		},
	}
	line := &BankStatementLine{Base: model.Base{ID: 30}, Amount: amount.FromFloat64(250), ContactID: helper.Ptr(uint64(5))}
	lines := BankStatementLineDAOMock{
		ListUnreconciledByStatementFunc: func(_ context.Context, _ uint64) ([]*BankStatementLine, error) {
			return []*BankStatementLine{line}, nil
		},
	}
	payments := PaymentDAOMock{
		ListPostedByContactFunc: func(_ context.Context, _ uint64) ([]*Payment, error) {
			return []*Payment{{Base: model.Base{ID: 90}, Amount: 999, EntryID: helper.Ptr(uint64(60))}}, nil
		},
	}
	svc := NewBankStatementService(statements, lines, payments, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})

	unreconciled, err := svc.Match(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(unreconciled) != 1 || line.Reconciled {
		t.Errorf("unreconciled = %d line reconciled %v, want 1/false", len(unreconciled), line.Reconciled)
	}
}

func TestBankStatementService_Match_RejectsCancelledStatement(t *testing.T) {
	ctx := context.Background()
	statement := &BankStatement{Base: model.Base{ID: 1}, State: BankStatementStateCancelled}
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return statement, nil },
		},
	}
	svc := NewBankStatementService(statements, BankStatementLineDAOMock{}, PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})

	_, err := svc.Match(ctx, 1)
	if helper.AssertError(t, err, true, ErrStatementCancelled) {
		return
	}
}

func TestBankStatementService_Create_PropagatesJournalError(t *testing.T) {
	ctx := context.Background()
	journals := dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewBankStatementService(BankStatementDAOMock{}, BankStatementLineDAOMock{}, PaymentDAOMock{}, journals, TransactionerMock{})

	_, err := svc.Create(ctx, CreateBankStatementRequest{OrganizationID: 10, JournalID: 20, Lines: []BankStatementLineRequest{{Amount: 250}}})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestBankStatementService_Create_PropagatesCreateError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	statements := BankStatementDAOMock{
		CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, _ *BankStatement, _ []*BankStatementLine) (*BankStatement, error) {
			return nil, errors.New("insert failed")
		},
	}
	journals := dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 20}}, nil
		},
	}
	svc := NewBankStatementService(statements, BankStatementLineDAOMock{}, PaymentDAOMock{}, journals, TransactionerMock{})

	_, err := svc.Create(ctx, CreateBankStatementRequest{OrganizationID: 10, JournalID: 20, Date: date, Lines: []BankStatementLineRequest{{Amount: 250}}})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestBankStatementService_Match_RejectsMissingStatement(t *testing.T) {
	ctx := context.Background()
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return nil, nil },
		},
	}
	svc := NewBankStatementService(statements, BankStatementLineDAOMock{}, PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})

	_, err := svc.Match(ctx, 1)
	if helper.AssertError(t, err, true, ErrStatementNotFound) {
		return
	}
}

func TestBankStatementService_Match_PropagatesFindError(t *testing.T) {
	ctx := context.Background()
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewBankStatementService(statements, BankStatementLineDAOMock{}, PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})

	_, err := svc.Match(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestBankStatementService_Match_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	statement := &BankStatement{Base: model.Base{ID: 1}, State: BankStatementStateOpen}
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return statement, nil },
		},
	}
	lines := BankStatementLineDAOMock{
		ListUnreconciledByStatementFunc: func(_ context.Context, _ uint64) ([]*BankStatementLine, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewBankStatementService(statements, lines, PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})

	_, err := svc.Match(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestBankStatementService_Match_PropagatesPaymentListError(t *testing.T) {
	ctx := context.Background()
	statement := &BankStatement{Base: model.Base{ID: 1}, State: BankStatementStateOpen}
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return statement, nil },
		},
	}
	lines := BankStatementLineDAOMock{
		ListUnreconciledByStatementFunc: func(_ context.Context, _ uint64) ([]*BankStatementLine, error) {
			return []*BankStatementLine{{Base: model.Base{ID: 30}, Amount: amount.FromFloat64(250), ContactID: helper.Ptr(uint64(5))}}, nil
		},
	}
	payments := PaymentDAOMock{
		ListPostedByContactFunc: func(_ context.Context, _ uint64) ([]*Payment, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewBankStatementService(statements, lines, payments, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})

	_, err := svc.Match(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestBankStatementService_Match_PropagatesUpdateError(t *testing.T) {
	ctx := context.Background()
	statement := &BankStatement{Base: model.Base{ID: 1}, State: BankStatementStateOpen}
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return statement, nil },
		},
	}
	lines := BankStatementLineDAOMock{
		ListUnreconciledByStatementFunc: func(_ context.Context, _ uint64) ([]*BankStatementLine, error) {
			return []*BankStatementLine{{Base: model.Base{ID: 30}, Amount: amount.FromFloat64(250), ContactID: helper.Ptr(uint64(5))}}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *BankStatementLine) (*BankStatementLine, error) {
			return nil, errors.New("update failed")
		},
	}
	payments := PaymentDAOMock{
		ListPostedByContactFunc: func(_ context.Context, _ uint64) ([]*Payment, error) {
			return []*Payment{{Base: model.Base{ID: 90}, Amount: 250, EntryID: helper.Ptr(uint64(60))}}, nil
		},
	}
	svc := NewBankStatementService(statements, lines, payments, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})

	_, err := svc.Match(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestBankStatementService_Match_PaymentWithoutMoveIsUnmatched(t *testing.T) {
	ctx := context.Background()
	statement := &BankStatement{Base: model.Base{ID: 1}, State: BankStatementStateOpen}
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return statement, nil },
		},
	}
	line := &BankStatementLine{Base: model.Base{ID: 30}, Amount: amount.FromFloat64(250), ContactID: helper.Ptr(uint64(5))}
	lines := BankStatementLineDAOMock{
		ListUnreconciledByStatementFunc: func(_ context.Context, _ uint64) ([]*BankStatementLine, error) {
			return []*BankStatementLine{line}, nil
		},
	}
	payments := PaymentDAOMock{
		ListPostedByContactFunc: func(_ context.Context, _ uint64) ([]*Payment, error) {
			return []*Payment{{Base: model.Base{ID: 90}, Amount: 250}}, nil
		},
	}
	svc := NewBankStatementService(statements, lines, payments, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})

	unreconciled, err := svc.Match(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(unreconciled) != 1 || line.Reconciled {
		t.Errorf("unreconciled = %d line reconciled %v, want 1/false", len(unreconciled), line.Reconciled)
	}
}

func TestBankStatementService_Match_LineWithoutContactIsUnmatched(t *testing.T) {
	ctx := context.Background()
	statement := &BankStatement{Base: model.Base{ID: 1}, State: BankStatementStateOpen}
	statements := BankStatementDAOMock{
		CRUDMock: dao.CRUDMock[BankStatement]{
			FindFunc: func(_ context.Context, _ uint64) (*BankStatement, error) { return statement, nil },
		},
	}
	line := &BankStatementLine{Base: model.Base{ID: 30}, Amount: amount.FromFloat64(250)}
	lines := BankStatementLineDAOMock{
		ListUnreconciledByStatementFunc: func(_ context.Context, _ uint64) ([]*BankStatementLine, error) {
			return []*BankStatementLine{line}, nil
		},
	}
	svc := NewBankStatementService(statements, lines, PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{}, TransactionerMock{})

	unreconciled, err := svc.Match(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(unreconciled) != 1 {
		t.Errorf("unreconciled = %d, want 1", len(unreconciled))
	}
}
