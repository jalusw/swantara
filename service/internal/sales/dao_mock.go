package sales

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"gorm.io/gorm"
)

type SaleOrderDAOMock struct {
	dao.CRUDMock[SaleOrder]
	CreateWithLinesFunc func(ctx context.Context, order *SaleOrder, lines []*SaleOrderLine) (*SaleOrder, error)
	UpdateTxFunc        func(ctx context.Context, tx *gorm.DB, order *SaleOrder) (*SaleOrder, error)
}

func (m SaleOrderDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, order *SaleOrder) (*SaleOrder, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, order)
	}
	return order, nil
}

func (m SaleOrderDAOMock) CreateWithLines(ctx context.Context, order *SaleOrder, lines []*SaleOrderLine) (*SaleOrder, error) {
	if m.CreateWithLinesFunc != nil {
		return m.CreateWithLinesFunc(ctx, order, lines)
	}
	return order, nil
}

type SaleOrderLineDAOMock struct {
	dao.CRUDMock[SaleOrderLine]
	ListByOrderFunc  func(ctx context.Context, orderID uint64) ([]*SaleOrderLine, error)
	ReplaceLinesFunc func(ctx context.Context, orderID uint64, lines []*SaleOrderLine) error
	UpdateTxFunc     func(ctx context.Context, tx *gorm.DB, line *SaleOrderLine) (*SaleOrderLine, error)
}

func (m SaleOrderLineDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, line *SaleOrderLine) (*SaleOrderLine, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, line)
	}
	return line, nil
}

func (m SaleOrderLineDAOMock) ListByOrder(ctx context.Context, orderID uint64) ([]*SaleOrderLine, error) {
	if m.ListByOrderFunc != nil {
		return m.ListByOrderFunc(ctx, orderID)
	}
	return []*SaleOrderLine{}, nil
}

func (m SaleOrderLineDAOMock) ReplaceLines(ctx context.Context, orderID uint64, lines []*SaleOrderLine) error {
	if m.ReplaceLinesFunc != nil {
		return m.ReplaceLinesFunc(ctx, orderID, lines)
	}
	return nil
}

type SequenceDAOMock struct {
	ReserveFunc func(ctx context.Context, organizationID uint64, code string, now time.Time) (*sequence.Reservation, error)
}

func (m SequenceDAOMock) Reserve(ctx context.Context, organizationID uint64, code string, now time.Time) (*sequence.Reservation, error) {
	if m.ReserveFunc != nil {
		return m.ReserveFunc(ctx, organizationID, code, now)
	}
	return &sequence.Reservation{Value: 1, Number: "SO/00001"}, nil
}

func (m SequenceDAOMock) ReserveTx(ctx context.Context, tx *gorm.DB, organizationID uint64, code string, now time.Time) (*sequence.Reservation, error) {
	return m.Reserve(ctx, organizationID, code, now)
}
