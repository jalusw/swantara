package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type JournalEntryDAOMock struct {
	dao.CRUDMock[JournalEntry]
	CreateWithLinesFunc   func(ctx context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error)
	CreateWithLinesTxFunc func(ctx context.Context, tx *gorm.DB, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error)
	FindByReversedFunc    func(ctx context.Context, moveID uint64) (*JournalEntry, error)
}

func (m JournalEntryDAOMock) CreateWithLines(ctx context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
	if m.CreateWithLinesFunc != nil {
		return m.CreateWithLinesFunc(ctx, entry, lines)
	}
	return entry, nil
}

func (m JournalEntryDAOMock) CreateWithLinesTx(ctx context.Context, tx *gorm.DB, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
	if m.CreateWithLinesTxFunc != nil {
		return m.CreateWithLinesTxFunc(ctx, tx, entry, lines)
	}
	return m.CreateWithLines(ctx, entry, lines)
}

func (m JournalEntryDAOMock) FindByReversed(ctx context.Context, moveID uint64) (*JournalEntry, error) {
	if m.FindByReversedFunc != nil {
		return m.FindByReversedFunc(ctx, moveID)
	}
	return nil, nil
}

type JournalLineDAOMock struct {
	dao.CRUDMock[JournalLine]
	ListByMovementFunc            func(ctx context.Context, moveID uint64) ([]*JournalLine, error)
	ListUnreconciledByAccountFunc func(ctx context.Context, accountID uint64) ([]*JournalLine, error)
	BalanceByOriginAndAccountFunc func(ctx context.Context, originType string, originID, accountID uint64) (float64, error)
	UpdateTxFunc                  func(ctx context.Context, tx *gorm.DB, line *JournalLine) (*JournalLine, error)
}

func (m JournalLineDAOMock) ListByMovement(ctx context.Context, moveID uint64) ([]*JournalLine, error) {
	if m.ListByMovementFunc != nil {
		return m.ListByMovementFunc(ctx, moveID)
	}
	return nil, nil
}

func (m JournalLineDAOMock) ListUnreconciledByAccount(ctx context.Context, accountID uint64) ([]*JournalLine, error) {
	if m.ListUnreconciledByAccountFunc != nil {
		return m.ListUnreconciledByAccountFunc(ctx, accountID)
	}
	return nil, nil
}

func (m JournalLineDAOMock) BalanceByOriginAndAccount(ctx context.Context, originType string, originID, accountID uint64) (float64, error) {
	if m.BalanceByOriginAndAccountFunc != nil {
		return m.BalanceByOriginAndAccountFunc(ctx, originType, originID, accountID)
	}
	return 0, nil
}

func (m JournalLineDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, line *JournalLine) (*JournalLine, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, line)
	}
	return line, nil
}
