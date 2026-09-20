package inventory

import (
	"context"
	"errors"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WarehouseDAO interface {
	dao.CRUD[reference.Warehouse]
}

type warehouseDAO struct {
	dao.Base[reference.Warehouse]
}

func NewWarehouseDAO(db *gorm.DB) WarehouseDAO {
	return warehouseDAO{Base: dao.NewBase[reference.Warehouse](db)}
}

type StockLocationDAO interface {
	dao.CRUD[reference.StockLocation]
}

type stockLocationDAO struct {
	dao.Base[reference.StockLocation]
}

func NewStockLocationDAO(db *gorm.DB) StockLocationDAO {
	return stockLocationDAO{Base: dao.NewBase[reference.StockLocation](db)}
}

type ShipmentDAO interface {
	dao.CRUD[Shipment]
	CreateWithMovements(ctx context.Context, shipment *Shipment, movements []*StockMovement) (*Shipment, error)
	CreateWithMovementsTx(ctx context.Context, tx *gorm.DB, shipment *Shipment, movements []*StockMovement) (*Shipment, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, shipment *Shipment) (*Shipment, error)
}

type stockShipmentDAO struct {
	dao.Base[Shipment]
	db *gorm.DB
}

func NewShipmentDAO(db *gorm.DB) ShipmentDAO {
	return stockShipmentDAO{Base: dao.NewBase[Shipment](db), db: db}
}

func (d stockShipmentDAO) CreateWithMovements(ctx context.Context, shipment *Shipment, movements []*StockMovement) (*Shipment, error) {
	var created *Shipment
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		created, err = d.CreateWithMovementsTx(ctx, tx, shipment, movements)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (d stockShipmentDAO) CreateWithMovementsTx(ctx context.Context, tx *gorm.DB, shipment *Shipment, movements []*StockMovement) (*Shipment, error) {
	if err := tx.WithContext(ctx).Create(shipment).Error; err != nil {
		return nil, err
	}
	for _, movement := range movements {
		movement.ShipmentID = &shipment.ID
		if err := tx.WithContext(ctx).Create(movement).Error; err != nil {
			return nil, err
		}
	}
	return shipment, nil
}

func (d stockShipmentDAO) UpdateTx(ctx context.Context, tx *gorm.DB, shipment *Shipment) (*Shipment, error) {
	return shipment, tx.WithContext(ctx).Save(shipment).Error
}

type StockMovementDAO interface {
	dao.CRUD[StockMovement]
	ListByShipment(ctx context.Context, shipmentID uint64) ([]*StockMovement, error)
	ListByOrigin(ctx context.Context, originType string, originID uint64) ([]*StockMovement, error)
	LedgerTotals(ctx context.Context, organizationID *uint64) ([]LedgerTotal, error)
	CreateTx(ctx context.Context, tx *gorm.DB, movement *StockMovement) (*StockMovement, error)
	FindForUpdateTx(ctx context.Context, tx *gorm.DB, movementID uint64) (*StockMovement, error)
	ApplyTx(ctx context.Context, tx *gorm.DB, movement *StockMovement) (*StockMovement, error)
	ApplyAllTx(ctx context.Context, tx *gorm.DB, movements []*StockMovement) error
}

type LedgerTotal struct {
	ItemID     uint64
	LocationID uint64
	BatchID    *uint64
	Total      float64
}

type stockMovementDAO struct {
	dao.Base[StockMovement]
	db *gorm.DB
}

func NewStockMovementDAO(db *gorm.DB) StockMovementDAO {
	return stockMovementDAO{Base: dao.NewBase[StockMovement](db), db: db}
}

func (d stockMovementDAO) CreateTx(ctx context.Context, tx *gorm.DB, movement *StockMovement) (*StockMovement, error) {
	if err := tx.WithContext(ctx).Create(movement).Error; err != nil {
		return nil, err
	}
	return movement, nil
}

func (d stockMovementDAO) ListByShipment(ctx context.Context, shipmentID uint64) ([]*StockMovement, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "shipment_id", Operator: query.Equal, Value: shipmentID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d stockMovementDAO) ListByOrigin(ctx context.Context, originType string, originID uint64) ([]*StockMovement, error) {
	var movements []StockMovement
	if err := d.db.WithContext(ctx).
		Where("origin_type = ? AND origin_id = ? AND deleted_at IS NULL", originType, originID).
		Find(&movements).Error; err != nil {
		return nil, err
	}
	items := make([]*StockMovement, len(movements))
	for i := range movements {
		items[i] = &movements[i]
	}
	return items, nil
}

func (d stockMovementDAO) LedgerTotals(ctx context.Context, organizationID *uint64) ([]LedgerTotal, error) {
	rows := []LedgerTotal{}
	query := `
		SELECT item_id, location_id, batch_id, SUM(qty) AS total
		FROM (
			SELECT item_id, dst_location_id AS location_id, batch_id, qty
			FROM stock_movements WHERE state = ? AND deleted_at IS NULL
			UNION ALL
			SELECT item_id, src_location_id AS location_id, batch_id, -qty
			FROM stock_movements WHERE state = ? AND deleted_at IS NULL
		) AS legs
		GROUP BY item_id, location_id, batch_id`
	args := []any{MovementStateDone, MovementStateDone}
	if organizationID != nil {
		query = `
			SELECT item_id, location_id, batch_id, SUM(qty) AS total
			FROM (
				SELECT item_id, dst_location_id AS location_id, batch_id, qty, organization_id
				FROM stock_movements WHERE state = ? AND organization_id = ? AND deleted_at IS NULL
				UNION ALL
				SELECT item_id, src_location_id AS location_id, batch_id, -qty, organization_id
				FROM stock_movements WHERE state = ? AND organization_id = ? AND deleted_at IS NULL
			) AS legs
			GROUP BY item_id, location_id, batch_id`
		args = []any{MovementStateDone, *organizationID, MovementStateDone, *organizationID}
	}
	err := d.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error
	return rows, err
}

func (d stockMovementDAO) FindForUpdateTx(ctx context.Context, tx *gorm.DB, movementID uint64) (*StockMovement, error) {
	var movement StockMovement
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&movement, movementID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &movement, nil
}

func (d stockMovementDAO) ApplyTx(ctx context.Context, tx *gorm.DB, movement *StockMovement) (*StockMovement, error) {
	if err := tx.WithContext(ctx).Save(movement).Error; err != nil {
		return nil, err
	}
	if _, err := upsertQuantTx(tx, movement.OrganizationID, movement.ItemID, movement.SrcLocationID, movement.BatchID, -movement.Qty); err != nil {
		return nil, err
	}
	if _, err := upsertQuantTx(tx, movement.OrganizationID, movement.ItemID, movement.DstLocationID, movement.BatchID, movement.Qty); err != nil {
		return nil, err
	}
	return movement, nil
}

func (d stockMovementDAO) ApplyAllTx(ctx context.Context, tx *gorm.DB, movements []*StockMovement) error {
	for _, movement := range movements {
		if _, err := d.ApplyTx(ctx, tx, movement); err != nil {
			return err
		}
	}
	return nil
}

type StockBalanceDAO interface {
	dao.CRUD[StockBalance]
	FindByKey(ctx context.Context, itemID, locationID uint64, batchID *uint64) (*StockBalance, error)
	FindByKeyTx(ctx context.Context, tx *gorm.DB, itemID, locationID uint64, batchID *uint64, forUpdate bool) (*StockBalance, error)
	Upsert(ctx context.Context, organizationID *uint64, itemID, locationID uint64, batchID *uint64, delta float64) (*StockBalance, error)
	UpsertTx(ctx context.Context, tx *gorm.DB, organizationID *uint64, itemID, locationID uint64, batchID *uint64, delta float64) (*StockBalance, error)
	ListByItem(ctx context.Context, itemID uint64) ([]*StockBalance, error)
	ListAll(ctx context.Context, organizationID *uint64) ([]*StockBalance, error)
}

type stockQuantDAO struct {
	dao.Base[StockBalance]
	db *gorm.DB
}

func NewStockBalanceDAO(db *gorm.DB) StockBalanceDAO {
	return stockQuantDAO{Base: dao.NewBase[StockBalance](db), db: db}
}

func (d stockQuantDAO) FindByKey(ctx context.Context, itemID, locationID uint64, batchID *uint64) (*StockBalance, error) {
	return findByKeyQuant(d.db.WithContext(ctx), itemID, locationID, batchID, false)
}

func (d stockQuantDAO) FindByKeyTx(ctx context.Context, tx *gorm.DB, itemID, locationID uint64, batchID *uint64, forUpdate bool) (*StockBalance, error) {
	return findByKeyQuant(tx.WithContext(ctx), itemID, locationID, batchID, forUpdate)
}

func (d stockQuantDAO) Upsert(ctx context.Context, organizationID *uint64, itemID, locationID uint64, batchID *uint64, delta float64) (*StockBalance, error) {
	var updated *StockBalance
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		updated, err = upsertQuantTx(tx, organizationID, itemID, locationID, batchID, delta)
		return err
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (d stockQuantDAO) UpsertTx(ctx context.Context, tx *gorm.DB, organizationID *uint64, itemID, locationID uint64, batchID *uint64, delta float64) (*StockBalance, error) {
	return upsertQuantTx(tx, organizationID, itemID, locationID, batchID, delta)
}

func (d stockQuantDAO) ListByItem(ctx context.Context, itemID uint64) ([]*StockBalance, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "item_id", Operator: query.Equal, Value: itemID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d stockQuantDAO) ListAll(ctx context.Context, organizationID *uint64) ([]*StockBalance, error) {
	var quants []StockBalance
	query := d.db.WithContext(ctx)
	if organizationID != nil {
		query = query.Where("organization_id = ?", *organizationID)
	}
	if err := query.Find(&quants).Error; err != nil {
		return nil, err
	}
	items := make([]*StockBalance, len(quants))
	for i := range quants {
		items[i] = &quants[i]
	}
	return items, nil
}

func findByKeyQuant(tx *gorm.DB, itemID, locationID uint64, batchID *uint64, forUpdate bool) (*StockBalance, error) {
	var quant StockBalance
	query := tx.Where("item_id = ? AND location_id = ?", itemID, locationID)
	if batchID == nil {
		query = query.Where("batch_id IS NULL")
	} else {
		query = query.Where("batch_id = ?", *batchID)
	}
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := query.First(&quant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &quant, nil
}

func upsertQuantTx(tx *gorm.DB, organizationID *uint64, itemID, locationID uint64, batchID *uint64, delta float64) (*StockBalance, error) {
	quant, err := findByKeyQuant(tx, itemID, locationID, batchID, true)
	if err != nil {
		return nil, err
	}
	if quant == nil {
		quant = &StockBalance{
			OrganizationID: organizationID,
			ItemID:         itemID,
			LocationID:     locationID,
			BatchID:        batchID,
			Quantity:       delta,
		}
		return quant, tx.Create(quant).Error
	}
	quant.Quantity += delta
	return quant, tx.Save(quant).Error
}

type BatchDAO interface {
	dao.CRUD[Batch]
	ListByItem(ctx context.Context, itemID uint64) ([]*Batch, error)
	ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[Batch], error)
	FindInOrg(ctx context.Context, id, organizationID uint64) (*Batch, error)
}

type stockBatchDAO struct {
	dao.Base[Batch]
	db *gorm.DB
}

func NewBatchDAO(db *gorm.DB) BatchDAO {
	return stockBatchDAO{Base: dao.NewBase[Batch](db), db: db}
}

func (d stockBatchDAO) ListByItem(ctx context.Context, itemID uint64) ([]*Batch, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "item_id", Operator: query.Equal, Value: itemID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d stockBatchDAO) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[Batch], error) {
	var count int64
	var entities []Batch

	join := "JOIN item_variants ON item_variants.id = batchs.item_id " +
		"JOIN items ON items.id = item_variants.item_id " +
		"AND items.organization_id = ?"

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

	items := make([]*Batch, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}

	return &query.Page[Batch]{Items: items, Count: count}, nil
}

func (d stockBatchDAO) FindInOrg(ctx context.Context, id, organizationID uint64) (*Batch, error) {
	var batch Batch
	err := d.db.WithContext(ctx).Model(&Batch{}).
		Joins("JOIN item_variants ON item_variants.id = batchs.item_id "+
			"JOIN items ON items.id = item_variants.item_id "+
			"AND items.organization_id = ?", organizationID).
		Where("batchs.id = ?", id).
		Take(&batch).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

type StockHoldDAO interface {
	dao.CRUD[StockHold]
	ListByMovement(ctx context.Context, movementID uint64) ([]*StockHold, error)
	DeleteByMove(ctx context.Context, movementID uint64) error
	Reserve(ctx context.Context, quantID uint64, movementID *uint64, qty float64) (*StockHold, error)
	Release(ctx context.Context, reservationID uint64) error
	ReleaseByMovement(ctx context.Context, movementID uint64) error
	ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[StockHold], error)
	FindInOrg(ctx context.Context, id, organizationID uint64) (*StockHold, error)
	ReleaseByMovementInOrg(ctx context.Context, organizationID, movementID uint64) error
}

type stockHoldDAO struct {
	dao.Base[StockHold]
	db *gorm.DB
}

func NewStockHoldDAO(db *gorm.DB) StockHoldDAO {
	return stockHoldDAO{Base: dao.NewBase[StockHold](db), db: db}
}

func (d stockHoldDAO) ListByMovement(ctx context.Context, movementID uint64) ([]*StockHold, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "movement_id", Operator: query.Equal, Value: movementID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d stockHoldDAO) DeleteByMove(ctx context.Context, movementID uint64) error {
	var reservation StockHold
	return d.db.WithContext(ctx).Where("movement_id = ?", movementID).Delete(&reservation).Error
}

func (d stockHoldDAO) Reserve(ctx context.Context, quantID uint64, movementID *uint64, qty float64) (*StockHold, error) {
	reservation := &StockHold{BalanceID: quantID, MovementID: movementID, Qty: qty}
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&StockBalance{}).
			Where("id = ? AND quantity - reserved_qty >= ?", quantID, qty).
			Update("reserved_qty", gorm.Expr("reserved_qty + ?", qty))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrHoldConflict
		}
		return tx.Create(reservation).Error
	})
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

func (d stockHoldDAO) Release(ctx context.Context, reservationID uint64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reservation StockHold
		if err := tx.First(&reservation, reservationID).Error; err != nil {
			return err
		}
		if err := tx.Model(&StockBalance{}).
			Where("id = ?", reservation.BalanceID).
			Update("reserved_qty", gorm.Expr("reserved_qty - ?", reservation.Qty)).Error; err != nil {
			return err
		}
		return tx.Delete(&StockHold{}, reservationID).Error
	})
}

func (d stockHoldDAO) ReleaseByMovement(ctx context.Context, movementID uint64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reservations []StockHold
		if err := tx.Where("movement_id = ?", movementID).Find(&reservations).Error; err != nil {
			return err
		}
		for _, reservation := range reservations {
			if err := tx.Model(&StockBalance{}).
				Where("id = ?", reservation.BalanceID).
				Update("reserved_qty", gorm.Expr("reserved_qty - ?", reservation.Qty)).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&StockHold{}, "movement_id = ?", movementID).Error
	})
}

func (d stockHoldDAO) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[StockHold], error) {
	var count int64
	var entities []StockHold

	countTx := d.db.WithContext(ctx).Model(&entities).
		Joins("JOIN stock_balances ON stock_balances.id = stock_holds.balance_id AND stock_balances.organization_id = ?", organizationID)
	if q != nil {
		countTx = query.ApplyFilters(countTx, q.Filters)
	}
	if err := countTx.Count(&count).Error; err != nil {
		return nil, err
	}

	tx := d.db.WithContext(ctx).Model(&entities).
		Joins("JOIN stock_balances ON stock_balances.id = stock_holds.balance_id AND stock_balances.organization_id = ?", organizationID)
	if q != nil {
		tx = query.ApplyQuery(tx, q)
	}
	if err := tx.Find(&entities).Error; err != nil {
		return nil, err
	}

	items := make([]*StockHold, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}

	return &query.Page[StockHold]{Items: items, Count: count}, nil
}

func (d stockHoldDAO) FindInOrg(ctx context.Context, id, organizationID uint64) (*StockHold, error) {
	var reservation StockHold
	err := d.db.WithContext(ctx).Model(&StockHold{}).
		Joins("JOIN stock_balances ON stock_balances.id = stock_holds.balance_id AND stock_balances.organization_id = ?", organizationID).
		Where("stock_holds.id = ?", id).
		Take(&reservation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (d stockHoldDAO) ReleaseByMovementInOrg(ctx context.Context, organizationID, movementID uint64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reservations []StockHold
		if err := tx.Model(&StockHold{}).
			Joins("JOIN stock_balances ON stock_balances.id = stock_holds.balance_id AND stock_balances.organization_id = ?", organizationID).
			Where("stock_holds.movement_id = ?", movementID).Find(&reservations).Error; err != nil {
			return err
		}
		for _, reservation := range reservations {
			if err := tx.Model(&StockBalance{}).
				Where("id = ?", reservation.BalanceID).
				Update("reserved_qty", gorm.Expr("reserved_qty - ?", reservation.Qty)).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&StockHold{}, "movement_id = ?", movementID).Error
	})
}

type CostLayerDAO interface {
	dao.CRUD[CostLayer]
	ListOpenByItem(ctx context.Context, itemID uint64) ([]*CostLayer, error)
	ListOpenByItemForUpdateTx(ctx context.Context, tx *gorm.DB, itemID uint64) ([]*CostLayer, error)
	ListOpenByItemInOrg(ctx context.Context, itemID, organizationID uint64) ([]*CostLayer, error)
	ValueForItem(ctx context.Context, itemID uint64) (float64, error)
	ListByMovement(ctx context.Context, movementID uint64) ([]*CostLayer, error)
	CreateTx(ctx context.Context, tx *gorm.DB, layer *CostLayer) (*CostLayer, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, layer *CostLayer) (*CostLayer, error)
}

type stockCostLayerDAO struct {
	dao.Base[CostLayer]
	db *gorm.DB
}

func NewCostLayerDAO(db *gorm.DB) CostLayerDAO {
	return stockCostLayerDAO{Base: dao.NewBase[CostLayer](db), db: db}
}

func (d stockCostLayerDAO) ListOpenByItem(ctx context.Context, itemID uint64) ([]*CostLayer, error) {
	var layers []CostLayer
	err := d.db.WithContext(ctx).
		Where("item_id = ? AND remaining_qty <> 0 AND deleted_at IS NULL", itemID).
		Order("id ASC").
		Find(&layers).Error
	if err != nil {
		return nil, err
	}
	items := make([]*CostLayer, len(layers))
	for i := range layers {
		items[i] = &layers[i]
	}
	return items, nil
}

func (d stockCostLayerDAO) ListOpenByItemForUpdateTx(ctx context.Context, tx *gorm.DB, itemID uint64) ([]*CostLayer, error) {
	var layers []CostLayer
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("item_id = ? AND remaining_qty <> 0 AND deleted_at IS NULL", itemID).
		Order("id ASC").
		Find(&layers).Error
	if err != nil {
		return nil, err
	}
	items := make([]*CostLayer, len(layers))
	for i := range layers {
		items[i] = &layers[i]
	}
	return items, nil
}

func (d stockCostLayerDAO) ListOpenByItemInOrg(ctx context.Context, itemID, organizationID uint64) ([]*CostLayer, error) {
	var layers []CostLayer
	err := d.db.WithContext(ctx).
		Joins("JOIN stock_movements ON stock_movements.id = cost_layers.movement_id").
		Where("cost_layers.item_id = ? AND stock_movements.organization_id = ? AND cost_layers.remaining_qty <> 0 AND cost_layers.deleted_at IS NULL", itemID, organizationID).
		Order("cost_layers.id ASC").
		Find(&layers).Error
	if err != nil {
		return nil, err
	}
	items := make([]*CostLayer, len(layers))
	for i := range layers {
		items[i] = &layers[i]
	}
	return items, nil
}

func (d stockCostLayerDAO) ValueForItem(ctx context.Context, itemID uint64) (float64, error) {
	var total float64
	err := d.db.WithContext(ctx).
		Model(&CostLayer{}).
		Where("item_id = ? AND deleted_at IS NULL", itemID).
		Select("COALESCE(SUM(remaining_value), 0)").
		Scan(&total).Error
	return total, err
}

func (d stockCostLayerDAO) ListByMovement(ctx context.Context, movementID uint64) ([]*CostLayer, error) {
	var layers []CostLayer
	if err := d.db.WithContext(ctx).
		Where("movement_id = ? AND deleted_at IS NULL", movementID).
		Order("id ASC").
		Find(&layers).Error; err != nil {
		return nil, err
	}
	items := make([]*CostLayer, len(layers))
	for i := range layers {
		items[i] = &layers[i]
	}
	return items, nil
}

func (d stockCostLayerDAO) CreateTx(ctx context.Context, tx *gorm.DB, layer *CostLayer) (*CostLayer, error) {
	return layer, tx.Create(layer).Error
}

func (d stockCostLayerDAO) UpdateTx(ctx context.Context, tx *gorm.DB, layer *CostLayer) (*CostLayer, error) {
	return layer, tx.Save(layer).Error
}

type ReorderRuleDAO interface {
	dao.CRUD[ReorderRule]
	ListActive(ctx context.Context) ([]*ReorderRule, error)
	ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[ReorderRule], error)
	ListActiveInOrg(ctx context.Context, organizationID uint64) ([]*ReorderRule, error)
	FindInOrg(ctx context.Context, id, organizationID uint64) (*ReorderRule, error)
}

type reorderRuleDAO struct {
	dao.Base[ReorderRule]
	db *gorm.DB
}

func NewReorderRuleDAO(db *gorm.DB) ReorderRuleDAO {
	return reorderRuleDAO{Base: dao.NewBase[ReorderRule](db), db: db}
}

func (d reorderRuleDAO) ListActive(ctx context.Context) ([]*ReorderRule, error) {
	var rules []ReorderRule
	if err := d.db.WithContext(ctx).Where("active = ? AND deleted_at IS NULL", true).Find(&rules).Error; err != nil {
		return nil, err
	}
	items := make([]*ReorderRule, len(rules))
	for i := range rules {
		items[i] = &rules[i]
	}
	return items, nil
}

func (d reorderRuleDAO) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[ReorderRule], error) {
	var count int64
	var entities []ReorderRule

	join := "JOIN item_variants ON item_variants.id = reorder_rules.item_id " +
		"JOIN items ON items.id = item_variants.item_id " +
		"AND items.organization_id = ?"

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

	items := make([]*ReorderRule, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}

	return &query.Page[ReorderRule]{Items: items, Count: count}, nil
}

func (d reorderRuleDAO) ListActiveInOrg(ctx context.Context, organizationID uint64) ([]*ReorderRule, error) {
	var rules []ReorderRule
	err := d.db.WithContext(ctx).Model(&ReorderRule{}).
		Joins("JOIN item_variants ON item_variants.id = reorder_rules.item_id "+
			"JOIN items ON items.id = item_variants.item_id "+
			"AND items.organization_id = ?", organizationID).
		Where("reorder_rules.active = ? AND reorder_rules.deleted_at IS NULL", true).
		Find(&rules).Error
	if err != nil {
		return nil, err
	}
	items := make([]*ReorderRule, len(rules))
	for i := range rules {
		items[i] = &rules[i]
	}
	return items, nil
}

func (d reorderRuleDAO) FindInOrg(ctx context.Context, id, organizationID uint64) (*ReorderRule, error) {
	var rule ReorderRule
	err := d.db.WithContext(ctx).Model(&ReorderRule{}).
		Joins("JOIN item_variants ON item_variants.id = reorder_rules.item_id "+
			"JOIN items ON items.id = item_variants.item_id "+
			"AND items.organization_id = ?", organizationID).
		Where("reorder_rules.id = ?", id).
		Take(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

type StockCountDAO interface {
	dao.CRUD[StockCount]
	CreateWithLines(ctx context.Context, count *StockCount, lines []*StockCountLine) (*StockCount, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, count *StockCount) (*StockCount, error)
}

type inventoryCountDAO struct {
	dao.Base[StockCount]
	db *gorm.DB
}

func NewStockCountDAO(db *gorm.DB) StockCountDAO {
	return inventoryCountDAO{Base: dao.NewBase[StockCount](db), db: db}
}

func (d inventoryCountDAO) CreateWithLines(ctx context.Context, count *StockCount, lines []*StockCountLine) (*StockCount, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(count).Error; err != nil {
			return err
		}
		for _, line := range lines {
			line.StockCountID = count.ID
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return count, nil
}

func (d inventoryCountDAO) UpdateTx(ctx context.Context, tx *gorm.DB, count *StockCount) (*StockCount, error) {
	return count, tx.Save(count).Error
}

type StockCountLineDAO interface {
	dao.CRUD[StockCountLine]
	ListByCount(ctx context.Context, stockCountID uint64) ([]*StockCountLine, error)
}

type inventoryCountLineDAO struct {
	dao.Base[StockCountLine]
}

func NewStockCountLineDAO(db *gorm.DB) StockCountLineDAO {
	return inventoryCountLineDAO{Base: dao.NewBase[StockCountLine](db)}
}

func (d inventoryCountLineDAO) ListByCount(ctx context.Context, stockCountID uint64) ([]*StockCountLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "stock_count_id", Operator: query.Equal, Value: stockCountID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type WarehouseTransferDAO interface {
	dao.CRUD[WarehouseTransfer]
	CreateWithShipments(ctx context.Context, transfer *WarehouseTransfer, outShipment *Shipment, outMovements []*StockMovement, inShipment *Shipment, inMovements []*StockMovement) (*WarehouseTransfer, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, transfer *WarehouseTransfer) (*WarehouseTransfer, error)
}

type transferOrderDAO struct {
	dao.Base[WarehouseTransfer]
	db *gorm.DB
}

func NewWarehouseTransferDAO(db *gorm.DB) WarehouseTransferDAO {
	return transferOrderDAO{Base: dao.NewBase[WarehouseTransfer](db), db: db}
}

func (d transferOrderDAO) CreateWithShipments(ctx context.Context, transfer *WarehouseTransfer, outShipment *Shipment, outMovements []*StockMovement, inShipment *Shipment, inMovements []*StockMovement) (*WarehouseTransfer, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(outShipment).Error; err != nil {
			return err
		}
		for _, movement := range outMovements {
			movement.ShipmentID = &outShipment.ID
			if err := tx.Create(movement).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(inShipment).Error; err != nil {
			return err
		}
		for _, movement := range inMovements {
			movement.ShipmentID = &inShipment.ID
			if err := tx.Create(movement).Error; err != nil {
				return err
			}
		}
		transfer.OutShipmentID = &outShipment.ID
		transfer.InShipmentID = &inShipment.ID
		if err := tx.Create(transfer).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return transfer, nil
}

func (d transferOrderDAO) UpdateTx(ctx context.Context, tx *gorm.DB, transfer *WarehouseTransfer) (*WarehouseTransfer, error) {
	return transfer, tx.Save(transfer).Error
}

type InboundCostDAO interface {
	dao.CRUD[InboundCost]
	CreateTx(ctx context.Context, tx *gorm.DB, cost *InboundCost) (*InboundCost, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, cost *InboundCost) (*InboundCost, error)
}

type inboundCostDAO struct {
	dao.Base[InboundCost]
	db *gorm.DB
}

func NewInboundCostDAO(db *gorm.DB) InboundCostDAO {
	return inboundCostDAO{Base: dao.NewBase[InboundCost](db), db: db}
}

func (d inboundCostDAO) CreateTx(ctx context.Context, tx *gorm.DB, cost *InboundCost) (*InboundCost, error) {
	return cost, tx.Create(cost).Error
}

func (d inboundCostDAO) UpdateTx(ctx context.Context, tx *gorm.DB, cost *InboundCost) (*InboundCost, error) {
	return cost, tx.Save(cost).Error
}

type InboundCostLineDAO interface {
	dao.CRUD[InboundCostLine]
	ListByInboundCost(ctx context.Context, costID uint64) ([]*InboundCostLine, error)
	CreateTx(ctx context.Context, tx *gorm.DB, line *InboundCostLine) (*InboundCostLine, error)
}

type inboundCostLineDAO struct {
	dao.Base[InboundCostLine]
	db *gorm.DB
}

func NewInboundCostLineDAO(db *gorm.DB) InboundCostLineDAO {
	return inboundCostLineDAO{Base: dao.NewBase[InboundCostLine](db), db: db}
}

func (d inboundCostLineDAO) ListByInboundCost(ctx context.Context, costID uint64) ([]*InboundCostLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "inbound_cost_id", Operator: query.Equal, Value: costID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d inboundCostLineDAO) CreateTx(ctx context.Context, tx *gorm.DB, line *InboundCostLine) (*InboundCostLine, error) {
	return line, tx.Create(line).Error
}

type InboundCostAdjustmentDAO interface {
	dao.CRUD[InboundCostAdjustment]
	ListByInboundCost(ctx context.Context, costID uint64) ([]*InboundCostAdjustment, error)
	CreateTx(ctx context.Context, tx *gorm.DB, adjustment *InboundCostAdjustment) (*InboundCostAdjustment, error)
}

type inboundCostAdjustmentDAO struct {
	dao.Base[InboundCostAdjustment]
	db *gorm.DB
}

func NewInboundCostAdjustmentDAO(db *gorm.DB) InboundCostAdjustmentDAO {
	return inboundCostAdjustmentDAO{Base: dao.NewBase[InboundCostAdjustment](db), db: db}
}

func (d inboundCostAdjustmentDAO) ListByInboundCost(ctx context.Context, costID uint64) ([]*InboundCostAdjustment, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "inbound_cost_id", Operator: query.Equal, Value: costID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d inboundCostAdjustmentDAO) CreateTx(ctx context.Context, tx *gorm.DB, adjustment *InboundCostAdjustment) (*InboundCostAdjustment, error) {
	return adjustment, tx.Create(adjustment).Error
}
