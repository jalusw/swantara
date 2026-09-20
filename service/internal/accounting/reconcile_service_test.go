package accounting

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestReconcileService_Reconcile(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult))
		request   ReconcileRequest
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "fully reconciles lines when amounts match exactly",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				var createdFull *AccountFullReconcile
				var reconciledLines []*JournalLine
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
							switch id {
							case 30:
								return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
							case 31:
								return &JournalLine{Base: model.Base{ID: 31}, EntryID: 2, AccountID: 1200, Credit: amount.FromInt64(100)}, nil
							}
							return nil, nil
						},
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, line *JournalLine) (*JournalLine, error) {
						reconciledLines = append(reconciledLines, line)
						return line, nil
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				partials := AccountPartialReconcileDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, reconcile *AccountPartialReconcile) (*AccountPartialReconcile, error) {
						reconcile.ID = 50
						return reconcile, nil
					},
				}
				fulls := AccountFullReconcileDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, reconcile *AccountFullReconcile) (*AccountFullReconcile, error) {
						reconcile.ID = 70
						createdFull = reconcile
						return reconcile, nil
					},
				}
				svc := NewReconcileService(lines, partials, fulls, movements, TransactionerMock{})
				return &svc, func(t *testing.T, result *ReconcileResult) {
					if !result.Reconciled || result.Full == nil {
						t.Errorf("result = reconciled %v full %v, want true/full", result.Reconciled, result.Full)
					}
					if createdFull.Name == nil || *createdFull.Name != "Reconciliation 50" {
						t.Errorf("full name = %v, want Reconciliation 50", createdFull.Name)
					}
					if len(reconciledLines) != 2 || !reconciledLines[0].Reconciled || !reconciledLines[1].Reconciled {
						t.Errorf("reconciled lines = %+v, want both marked", reconciledLines)
					}
				}
			},
			request: ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
		},
		{
			name: "partial reconcile does not close lines",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				var partialAmounts []float64
				var updatedLines []*JournalLine
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
							switch id {
							case 30:
								return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(200)}, nil
							case 31:
								return &JournalLine{Base: model.Base{ID: 31}, EntryID: 2, AccountID: 1200, Credit: amount.FromInt64(200)}, nil
							}
							return nil, nil
						},
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, line *JournalLine) (*JournalLine, error) {
						updatedLines = append(updatedLines, line)
						return line, nil
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				partials := AccountPartialReconcileDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, reconcile *AccountPartialReconcile) (*AccountPartialReconcile, error) {
						reconcile.ID = 50
						partialAmounts = append(partialAmounts, reconcile.Amount)
						return reconcile, nil
					},
				}
				svc := NewReconcileService(lines, partials, AccountFullReconcileDAOMock{}, movements, TransactionerMock{})
				return &svc, func(t *testing.T, result *ReconcileResult) {
					if result.Reconciled || result.Full != nil {
						t.Errorf("result = reconciled %v full %v, want false/nil", result.Reconciled, result.Full)
					}
					if len(partialAmounts) != 1 || partialAmounts[0] != 50 {
						t.Errorf("partial amounts = %v, want [50]", partialAmounts)
					}
					if len(updatedLines) != 0 {
						t.Errorf("updated lines = %d, want 0", len(updatedLines))
					}
				}
			},
			request: ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 50},
		},
		{
			name: "rejects lines with different accounts",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
							if id == 30 {
								return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
							}
							return &JournalLine{Base: model.Base{ID: 31}, EntryID: 2, AccountID: 1201, Credit: amount.FromInt64(100)}, nil
						},
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, movements, TransactionerMock{})
				return &svc, nil
			},
			request:   ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 50},
			wantErr:   true,
			wantErrIs: ErrLinesDifferentAccount,
		},
		{
			name: "rejects amount exceeding remaining balance",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
							switch id {
							case 30:
								return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
							case 31:
								return &JournalLine{Base: model.Base{ID: 31}, EntryID: 2, AccountID: 1200, Credit: amount.FromInt64(100)}, nil
							}
							return nil, nil
						},
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, movements, TransactionerMock{})
				return &svc, nil
			},
			request:   ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 150},
			wantErr:   true,
			wantErrIs: ErrReconcileExceedsBalance,
		},
		{
			name: "rejects non-positive amount",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				svc := NewReconcileService(JournalLineDAOMock{}, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, JournalEntryDAOMock{}, TransactionerMock{})
				return &svc, nil
			},
			request:   ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 0},
			wantErr:   true,
			wantErrIs: ErrReconcileAmount,
		},
		{
			name: "rejects missing debit line",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalLine, error) { return nil, nil },
					},
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, JournalEntryDAOMock{}, TransactionerMock{})
				return &svc, nil
			},
			request:   ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
			wantErr:   true,
			wantErrIs: ErrLineNotFound,
		},
		{
			name: "rejects missing credit line",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
							if id == 30 {
								return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
							}
							return nil, nil
						},
					},
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, JournalEntryDAOMock{}, TransactionerMock{})
				return &svc, nil
			},
			request:   ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
			wantErr:   true,
			wantErrIs: ErrLineNotFound,
		},
		{
			name: "rejects debit entry outside organization",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalLine, error) {
							return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
						},
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 99}, nil
						},
					},
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, movements, TransactionerMock{})
				return &svc, nil
			},
			request:   ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
			wantErr:   true,
			wantErrIs: ErrLineNotInOrganization,
		},
		{
			name: "rejects credit entry outside organization",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				moveByID := map[uint64]*JournalEntry{
					1: {Base: model.Base{ID: 1}, OrganizationID: 10},
					2: {Base: model.Base{ID: 2}, OrganizationID: 99},
				}
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
							switch id {
							case 30:
								return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
							case 31:
								return &JournalLine{Base: model.Base{ID: 31}, EntryID: 2, AccountID: 1200, Credit: amount.FromInt64(100)}, nil
							}
							return nil, nil
						},
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, id uint64) (*JournalEntry, error) { return moveByID[id], nil },
					},
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, movements, TransactionerMock{})
				return &svc, nil
			},
			request:   ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
			wantErr:   true,
			wantErrIs: ErrLineNotInOrganization,
		},
		{
			name: "rejects invalid line without debit or credit",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalLine, error) {
							return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200}, nil
						},
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, movements, TransactionerMock{})
				return &svc, nil
			},
			request:   ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
			wantErr:   true,
			wantErrIs: ErrInvalidLine,
		},
		{
			name: "rejects over-reconciled line",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalLine, error) {
							return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
						},
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				partials := AccountPartialReconcileDAOMock{
					TotalByLineFunc: func(_ context.Context, _ uint64) (float64, error) { return 150, nil },
				}
				svc := NewReconcileService(lines, partials, AccountFullReconcileDAOMock{}, movements, TransactionerMock{})
				return &svc, nil
			},
			request: ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
			wantErr: true,
		},
		{
			name: "reconciles debit side only when amount equals debit remaining",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				var updatedLines []*JournalLine
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
							switch id {
							case 30:
								return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
							case 31:
								return &JournalLine{Base: model.Base{ID: 31}, EntryID: 2, AccountID: 1200, Credit: amount.FromInt64(200)}, nil
							}
							return nil, nil
						},
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, line *JournalLine) (*JournalLine, error) {
						updatedLines = append(updatedLines, line)
						return line, nil
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, movements, TransactionerMock{})
				return &svc, func(t *testing.T, result *ReconcileResult) {
					if result.Reconciled || result.Full != nil {
						t.Errorf("result = reconciled %v full %v, want false/nil", result.Reconciled, result.Full)
					}
					if len(updatedLines) != 1 || !updatedLines[0].Reconciled || updatedLines[0].ID != 30 {
						t.Errorf("updated lines = %+v, want debit line 30 reconciled only", updatedLines)
					}
				}
			},
			request: ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
		},
		{
			name: "reconciles credit side only when amount equals credit remaining",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				var updatedLines []*JournalLine
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
							switch id {
							case 30:
								return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(200)}, nil
							case 31:
								return &JournalLine{Base: model.Base{ID: 31}, EntryID: 2, AccountID: 1200, Credit: amount.FromInt64(100)}, nil
							}
							return nil, nil
						},
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, line *JournalLine) (*JournalLine, error) {
						updatedLines = append(updatedLines, line)
						return line, nil
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, movements, TransactionerMock{})
				return &svc, func(t *testing.T, result *ReconcileResult) {
					if result.Reconciled || result.Full != nil {
						t.Errorf("result = reconciled %v full %v, want false/nil", result.Reconciled, result.Full)
					}
					if len(updatedLines) != 1 || !updatedLines[0].Reconciled || updatedLines[0].ID != 31 {
						t.Errorf("updated lines = %+v, want credit line 31 reconciled only", updatedLines)
					}
				}
			},
			request: ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
		},
		{
			name: "propagates partial create error",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalLine, error) {
							return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
						},
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				partials := AccountPartialReconcileDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *AccountPartialReconcile) (*AccountPartialReconcile, error) {
						return nil, errors.New("insert failed")
					},
				}
				svc := NewReconcileService(lines, partials, AccountFullReconcileDAOMock{}, movements, TransactionerMock{})
				return &svc, nil
			},
			request: ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
			wantErr: true,
		},
		{
			name: "propagates line find error",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalLine, error) {
							return nil, errors.New("db down")
						},
					},
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, JournalEntryDAOMock{}, TransactionerMock{})
				return &svc, nil
			},
			request: ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
			wantErr: true,
		},
		{
			name: "propagates total by line error",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalLine, error) {
							return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
						},
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				partials := AccountPartialReconcileDAOMock{
					TotalByLineFunc: func(_ context.Context, _ uint64) (float64, error) { return 0, errors.New("db down") },
				}
				svc := NewReconcileService(lines, partials, AccountFullReconcileDAOMock{}, movements, TransactionerMock{})
				return &svc, nil
			},
			request: ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
			wantErr: true,
		},
		{
			name: "propagates transaction error",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalLine, error) {
							return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
						},
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				tx := TransactionerMock{
					RunFunc: func(_ context.Context, _ func(*gorm.DB) error) error { return errors.New("tx failed") },
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, movements, tx)
				return &svc, nil
			},
			request: ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
			wantErr: true,
		},
		{
			name: "full reconcile updates both lines and sets full reconcile id",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				var updatedLines []*JournalLine
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
							switch id {
							case 30:
								return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
							case 31:
								return &JournalLine{Base: model.Base{ID: 31}, EntryID: 2, AccountID: 1200, Credit: amount.FromInt64(100)}, nil
							}
							return nil, nil
						},
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, line *JournalLine) (*JournalLine, error) {
						updatedLines = append(updatedLines, line)
						return line, nil
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				partials := AccountPartialReconcileDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, reconcile *AccountPartialReconcile) (*AccountPartialReconcile, error) {
						reconcile.ID = 5
						return reconcile, nil
					},
				}
				fulls := AccountFullReconcileDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, reconcile *AccountFullReconcile) (*AccountFullReconcile, error) {
						reconcile.ID = 9
						return reconcile, nil
					},
				}
				svc := NewReconcileService(lines, partials, fulls, movements, TransactionerMock{})
				return &svc, func(t *testing.T, result *ReconcileResult) {
					if !result.Reconciled || result.Full == nil || result.Full.ID != 9 {
						t.Errorf("result = reconciled %v full %+v, want true/9", result.Reconciled, result.Full)
					}
					if result.Partial == nil || result.Partial.FullReconcileID == nil || *result.Partial.FullReconcileID != 9 {
						t.Errorf("partial = %+v, want full_reconcile_id 9", result.Partial)
					}
					if len(updatedLines) != 2 || !updatedLines[0].Reconciled || !updatedLines[1].Reconciled {
						t.Errorf("updated lines = %+v, want both reconciled", updatedLines)
					}
					if updatedLines[0].FullReconcileID == nil || *updatedLines[0].FullReconcileID != 9 {
						t.Errorf("debit line full_reconcile_id = %v, want 9", updatedLines[0].FullReconcileID)
					}
				}
			},
			request: ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
		},
		{
			name: "propagates full reconcile create error",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
							switch id {
							case 30:
								return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
							case 31:
								return &JournalLine{Base: model.Base{ID: 31}, EntryID: 2, AccountID: 1200, Credit: amount.FromInt64(100)}, nil
							}
							return nil, nil
						},
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				fulls := AccountFullReconcileDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *AccountFullReconcile) (*AccountFullReconcile, error) {
						return nil, errors.New("insert failed")
					},
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, fulls, movements, TransactionerMock{})
				return &svc, nil
			},
			request: ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
			wantErr: true,
		},
		{
			name: "propagates line update error",
			setup: func(t *testing.T) (*ReconcileService, func(t *testing.T, result *ReconcileResult)) {
				lines := JournalLineDAOMock{
					CRUDMock: dao.CRUDMock[JournalLine]{
						FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
							switch id {
							case 30:
								return &JournalLine{Base: model.Base{ID: 30}, EntryID: 1, AccountID: 1200, Debit: amount.FromInt64(100)}, nil
							case 31:
								return &JournalLine{Base: model.Base{ID: 31}, EntryID: 2, AccountID: 1200, Credit: amount.FromInt64(200)}, nil
							}
							return nil, nil
						},
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *JournalLine) (*JournalLine, error) {
						return nil, errors.New("update failed")
					},
				}
				movements := JournalEntryDAOMock{
					CRUDMock: dao.CRUDMock[JournalEntry]{
						FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
							return &JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
						},
					},
				}
				svc := NewReconcileService(lines, AccountPartialReconcileDAOMock{}, AccountFullReconcileDAOMock{}, movements, TransactionerMock{})
				return &svc, nil
			},
			request: ReconcileRequest{OrganizationID: 10, DebitLineID: 30, CreditLineID: 31, Amount: 100},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, verify := tt.setup(t)
			result, err := svc.Reconcile(context.Background(), tt.request)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if verify != nil {
				verify(t, result)
			}
		})
	}
}
