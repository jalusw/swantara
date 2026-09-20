package returns

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type RMADAOMock struct {
	dao.CRUDMock[RMA]
	CreateWithLinesTxFunc func(ctx context.Context, tx *gorm.DB, rma *RMA, lines []*RMALine) (*RMA, error)
	UpdateTxFunc          func(ctx context.Context, tx *gorm.DB, rma *RMA) (*RMA, error)
}

func rmaDAOMock(find func(ctx context.Context, id uint64) (*RMA, error)) RMADAOMock {
	return RMADAOMock{CRUDMock: dao.CRUDMock[RMA]{FindFunc: find}}
}

func (m RMADAOMock) CreateWithLinesTx(ctx context.Context, tx *gorm.DB, rma *RMA, lines []*RMALine) (*RMA, error) {
	if m.CreateWithLinesTxFunc != nil {
		return m.CreateWithLinesTxFunc(ctx, tx, rma, lines)
	}
	return rma, nil
}

func (m RMADAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, rma *RMA) (*RMA, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, rma)
	}
	return rma, nil
}

type RMALineDAOMock struct {
	dao.CRUDMock[RMALine]
	ListByRMAFunc func(ctx context.Context, rmaID uint64) ([]*RMALine, error)
	UpdateTxFunc  func(ctx context.Context, tx *gorm.DB, line *RMALine) (*RMALine, error)
}

func (m RMALineDAOMock) ListByRMA(ctx context.Context, rmaID uint64) ([]*RMALine, error) {
	if m.ListByRMAFunc != nil {
		return m.ListByRMAFunc(ctx, rmaID)
	}
	return []*RMALine{}, nil
}

func (m RMALineDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, line *RMALine) (*RMALine, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, line)
	}
	return line, nil
}
