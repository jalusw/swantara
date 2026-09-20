package sales

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type SaleOrderDAO interface {
	dao.CRUD[SaleOrder]
	CreateWithLines(ctx context.Context, order *SaleOrder, lines []*SaleOrderLine) (*SaleOrder, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, order *SaleOrder) (*SaleOrder, error)
}

type saleOrderDAO struct {
	dao.Base[SaleOrder]
	db *gorm.DB
}

func NewSaleOrderDAO(db *gorm.DB) SaleOrderDAO {
	return saleOrderDAO{Base: dao.NewBase[SaleOrder](db), db: db}
}

func (d saleOrderDAO) CreateWithLines(ctx context.Context, order *SaleOrder, lines []*SaleOrderLine) (*SaleOrder, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for _, line := range lines {
			line.OrderID = order.ID
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (d saleOrderDAO) UpdateTx(ctx context.Context, tx *gorm.DB, order *SaleOrder) (*SaleOrder, error) {
	if err := tx.WithContext(ctx).Save(order).Error; err != nil {
		return nil, err
	}
	return order, nil
}

type SaleOrderLineDAO interface {
	dao.CRUD[SaleOrderLine]
	ListByOrder(ctx context.Context, orderID uint64) ([]*SaleOrderLine, error)
	ReplaceLines(ctx context.Context, orderID uint64, lines []*SaleOrderLine) error
	UpdateTx(ctx context.Context, tx *gorm.DB, line *SaleOrderLine) (*SaleOrderLine, error)
}

type saleOrderLineDAO struct {
	dao.Base[SaleOrderLine]
	db *gorm.DB
}

func NewSaleOrderLineDAO(db *gorm.DB) SaleOrderLineDAO {
	return saleOrderLineDAO{Base: dao.NewBase[SaleOrderLine](db), db: db}
}

func (d saleOrderLineDAO) ListByOrder(ctx context.Context, orderID uint64) ([]*SaleOrderLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "order_id", Operator: query.Equal, Value: orderID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d saleOrderLineDAO) UpdateTx(ctx context.Context, tx *gorm.DB, line *SaleOrderLine) (*SaleOrderLine, error) {
	if err := tx.WithContext(ctx).Save(line).Error; err != nil {
		return nil, err
	}
	return line, nil
}

func (d saleOrderLineDAO) ReplaceLines(ctx context.Context, orderID uint64, lines []*SaleOrderLine) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", orderID).Delete(&SaleOrderLine{}).Error; err != nil {
			return err
		}
		for _, line := range lines {
			line.OrderID = orderID
			line.ID = 0
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
