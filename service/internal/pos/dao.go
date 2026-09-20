package pos

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type POSSessionDAO interface {
	dao.CRUD[POSSession]
	ListInOrganization(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[POSSession], error)
}

type posSessionDAO struct {
	dao.Base[POSSession]
	db *gorm.DB
}

func NewPOSSessionDAO(db *gorm.DB) POSSessionDAO {
	return posSessionDAO{Base: dao.NewBase[POSSession](db), db: db}
}

func (d posSessionDAO) ListInOrganization(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[POSSession], error) {
	var count int64
	var entities []POSSession

	join := "JOIN pos_configs ON pos_configs.id = pos_sessions.config_id AND pos_configs.organization_id = ?"

	countTx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID)
	if q != nil {
		countTx = query.ApplyFilters(countTx, q.Filters)
	}
	if err := countTx.Count(&count).Error; err != nil {
		return nil, err
	}

	tx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID)
	if q != nil {
		tx = query.ApplyQuery(tx, q)
	}
	if err := tx.Find(&entities).Error; err != nil {
		return nil, err
	}

	items := make([]*POSSession, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return &query.Page[POSSession]{Items: items, Count: count}, nil
}

type POSOrderDAO interface {
	dao.CRUD[POSOrder]
	CreateWithLinesAndPayments(ctx context.Context, order *POSOrder, lines []*POSOrderLine, payments []*POSPayment) (*POSOrder, error)
	CreateWithLinesAndPaymentsTx(ctx context.Context, tx *gorm.DB, order *POSOrder, lines []*POSOrderLine, payments []*POSPayment) (*POSOrder, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, order *POSOrder) (*POSOrder, error)
	ListInOrganization(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[POSOrder], error)
}

type posOrderDAO struct {
	dao.Base[POSOrder]
	db *gorm.DB
}

func NewPOSOrderDAO(db *gorm.DB) POSOrderDAO {
	return posOrderDAO{Base: dao.NewBase[POSOrder](db), db: db}
}

func (d posOrderDAO) CreateWithLinesAndPayments(ctx context.Context, order *POSOrder, lines []*POSOrderLine, payments []*POSPayment) (*POSOrder, error) {
	var created *POSOrder
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		created, err = d.CreateWithLinesAndPaymentsTx(ctx, tx, order, lines, payments)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (d posOrderDAO) CreateWithLinesAndPaymentsTx(ctx context.Context, tx *gorm.DB, order *POSOrder, lines []*POSOrderLine, payments []*POSPayment) (*POSOrder, error) {
	if err := tx.WithContext(ctx).Create(order).Error; err != nil {
		return nil, err
	}
	for _, line := range lines {
		line.OrderID = order.ID
		if err := tx.WithContext(ctx).Create(line).Error; err != nil {
			return nil, err
		}
	}
	for _, payment := range payments {
		payment.OrderID = order.ID
		if err := tx.WithContext(ctx).Create(payment).Error; err != nil {
			return nil, err
		}
	}
	return order, nil
}

func (d posOrderDAO) UpdateTx(ctx context.Context, tx *gorm.DB, order *POSOrder) (*POSOrder, error) {
	return order, tx.WithContext(ctx).Save(order).Error
}

func (d posOrderDAO) ListInOrganization(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[POSOrder], error) {
	var count int64
	var entities []POSOrder

	join := "JOIN pos_sessions ON pos_sessions.id = pos_orders.session_id " +
		"JOIN pos_configs ON pos_configs.id = pos_sessions.config_id " +
		"AND pos_configs.organization_id = ?"

	countTx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID)
	if q != nil {
		countTx = query.ApplyFilters(countTx, q.Filters)
	}
	if err := countTx.Count(&count).Error; err != nil {
		return nil, err
	}

	tx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID)
	if q != nil {
		tx = query.ApplyQuery(tx, q)
	}
	if err := tx.Find(&entities).Error; err != nil {
		return nil, err
	}

	items := make([]*POSOrder, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return &query.Page[POSOrder]{Items: items, Count: count}, nil
}

type POSOrderLineDAO interface {
	dao.CRUD[POSOrderLine]
	ListByOrder(ctx context.Context, orderID uint64) ([]*POSOrderLine, error)
}

type posOrderLineDAO struct {
	dao.Base[POSOrderLine]
}

func NewPOSOrderLineDAO(db *gorm.DB) POSOrderLineDAO {
	return posOrderLineDAO{Base: dao.NewBase[POSOrderLine](db)}
}

func (d posOrderLineDAO) ListByOrder(ctx context.Context, orderID uint64) ([]*POSOrderLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "order_id", Operator: query.Equal, Value: orderID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type PaymentMethodTotal struct {
	Method string
	Total  float64
}

type POSPaymentDAO interface {
	dao.CRUD[POSPayment]
	ListByOrder(ctx context.Context, orderID uint64) ([]*POSPayment, error)
	SumBySession(ctx context.Context, sessionID uint64) ([]PaymentMethodTotal, error)
}

type posPaymentDAO struct {
	dao.Base[POSPayment]
	db *gorm.DB
}

func NewPOSPaymentDAO(db *gorm.DB) POSPaymentDAO {
	return posPaymentDAO{Base: dao.NewBase[POSPayment](db), db: db}
}

func (d posPaymentDAO) ListByOrder(ctx context.Context, orderID uint64) ([]*POSPayment, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "order_id", Operator: query.Equal, Value: orderID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d posPaymentDAO) SumBySession(ctx context.Context, sessionID uint64) ([]PaymentMethodTotal, error) {
	totals := []PaymentMethodTotal{}
	err := d.db.WithContext(ctx).
		Model(&POSPayment{}).
		Select("pos_payments.method AS method, SUM(pos_payments.amount) AS total").
		Joins("JOIN pos_orders ON pos_orders.id = pos_payments.order_id").
		Where("pos_orders.session_id = ? AND pos_payments.deleted_at IS NULL AND pos_orders.deleted_at IS NULL", sessionID).
		Group("pos_payments.method").
		Scan(&totals).Error
	return totals, err
}
