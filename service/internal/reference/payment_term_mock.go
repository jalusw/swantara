package reference

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
)

type PaymentTermDAOMock struct {
	dao.CRUDMock[PaymentTerm]
	ListLinesFunc       func(ctx context.Context, termID uint64) ([]*PaymentTermLine, error)
	ReplaceLinesFunc    func(ctx context.Context, termID uint64, lines []*PaymentTermLine) error
	DeleteWithLinesFunc func(ctx context.Context, termID uint64) error
}

func (m PaymentTermDAOMock) ListLines(ctx context.Context, termID uint64) ([]*PaymentTermLine, error) {
	if m.ListLinesFunc != nil {
		return m.ListLinesFunc(ctx, termID)
	}
	return []*PaymentTermLine{}, nil
}

func (m PaymentTermDAOMock) ReplaceLines(ctx context.Context, termID uint64, lines []*PaymentTermLine) error {
	if m.ReplaceLinesFunc != nil {
		return m.ReplaceLinesFunc(ctx, termID, lines)
	}
	return nil
}

func (m PaymentTermDAOMock) DeleteWithLines(ctx context.Context, termID uint64) error {
	if m.DeleteWithLinesFunc != nil {
		return m.DeleteWithLinesFunc(ctx, termID)
	}
	return nil
}
