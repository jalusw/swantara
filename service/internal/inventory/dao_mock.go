package inventory

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type TransactionerMock struct {
	RunFunc func(ctx context.Context, fn func(tx *gorm.DB) error) error
}

func (m TransactionerMock) Run(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, fn)
	}
	return fn(nil)
}

type WarehouseDAOMock struct {
	dao.CRUDMock[reference.Warehouse]
}

type StockLocationDAOMock struct {
	dao.CRUDMock[reference.StockLocation]
}

type ShipmentDAOMock struct {
	dao.CRUDMock[Shipment]
	CreateWithMovementsFunc   func(ctx context.Context, shipment *Shipment, movements []*StockMovement) (*Shipment, error)
	CreateWithMovementsTxFunc func(ctx context.Context, tx *gorm.DB, shipment *Shipment, movements []*StockMovement) (*Shipment, error)
	UpdateTxFunc              func(ctx context.Context, tx *gorm.DB, shipment *Shipment) (*Shipment, error)
}

func (m ShipmentDAOMock) CreateWithMovements(ctx context.Context, shipment *Shipment, movements []*StockMovement) (*Shipment, error) {
	if m.CreateWithMovementsFunc != nil {
		return m.CreateWithMovementsFunc(ctx, shipment, movements)
	}
	return shipment, nil
}

func (m ShipmentDAOMock) CreateWithMovementsTx(ctx context.Context, tx *gorm.DB, shipment *Shipment, movements []*StockMovement) (*Shipment, error) {
	if m.CreateWithMovementsTxFunc != nil {
		return m.CreateWithMovementsTxFunc(ctx, tx, shipment, movements)
	}
	return m.CreateWithMovements(ctx, shipment, movements)
}

func (m ShipmentDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, shipment *Shipment) (*Shipment, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, shipment)
	}
	return m.Update(ctx, shipment)
}

type StockMovementDAOMock struct {
	dao.CRUDMock[StockMovement]
	ListByShipmentFunc  func(ctx context.Context, shipmentID uint64) ([]*StockMovement, error)
	ListByOriginFunc    func(ctx context.Context, originType string, originID uint64) ([]*StockMovement, error)
	LedgerTotalsFunc    func(ctx context.Context, organizationID *uint64) ([]LedgerTotal, error)
	CreateTxFunc        func(ctx context.Context, tx *gorm.DB, movement *StockMovement) (*StockMovement, error)
	FindForUpdateTxFunc func(ctx context.Context, tx *gorm.DB, movementID uint64) (*StockMovement, error)
	ApplyTxFunc         func(ctx context.Context, tx *gorm.DB, movement *StockMovement) (*StockMovement, error)
	ApplyAllTxFunc      func(ctx context.Context, tx *gorm.DB, movements []*StockMovement) error
}

func (m StockMovementDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, movement *StockMovement) (*StockMovement, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, movement)
	}
	return movement, nil
}

func (m StockMovementDAOMock) ListByShipment(ctx context.Context, shipmentID uint64) ([]*StockMovement, error) {
	if m.ListByShipmentFunc != nil {
		return m.ListByShipmentFunc(ctx, shipmentID)
	}
	return []*StockMovement{}, nil
}

func (m StockMovementDAOMock) ListByOrigin(ctx context.Context, originType string, originID uint64) ([]*StockMovement, error) {
	if m.ListByOriginFunc != nil {
		return m.ListByOriginFunc(ctx, originType, originID)
	}
	return []*StockMovement{}, nil
}

func (m StockMovementDAOMock) LedgerTotals(ctx context.Context, organizationID *uint64) ([]LedgerTotal, error) {
	if m.LedgerTotalsFunc != nil {
		return m.LedgerTotalsFunc(ctx, organizationID)
	}
	return []LedgerTotal{}, nil
}

func (m StockMovementDAOMock) FindForUpdateTx(ctx context.Context, tx *gorm.DB, movementID uint64) (*StockMovement, error) {
	if m.FindForUpdateTxFunc != nil {
		return m.FindForUpdateTxFunc(ctx, tx, movementID)
	}
	if movement, err := m.Find(ctx, movementID); movement != nil || err != nil {
		return movement, err
	}
	return &StockMovement{Base: model.Base{ID: movementID}, State: MovementStateConfirmed}, nil
}

func (m StockMovementDAOMock) ApplyTx(ctx context.Context, tx *gorm.DB, movement *StockMovement) (*StockMovement, error) {
	if m.ApplyTxFunc != nil {
		return m.ApplyTxFunc(ctx, tx, movement)
	}
	if movement == nil {
		movement = &StockMovement{}
	}
	return movement, nil
}

func (m StockMovementDAOMock) ApplyAllTx(ctx context.Context, tx *gorm.DB, movements []*StockMovement) error {
	if m.ApplyAllTxFunc != nil {
		return m.ApplyAllTxFunc(ctx, tx, movements)
	}
	for _, movement := range movements {
		if _, err := m.ApplyTx(ctx, tx, movement); err != nil {
			return err
		}
	}
	return nil
}

type StockBalanceDAOMock struct {
	dao.CRUDMock[StockBalance]
	FindByKeyFunc  func(ctx context.Context, itemID, locationID uint64, batchID *uint64) (*StockBalance, error)
	UpsertFunc     func(ctx context.Context, organizationID *uint64, itemID, locationID uint64, batchID *uint64, delta float64) (*StockBalance, error)
	UpsertTxFunc   func(ctx context.Context, tx *gorm.DB, organizationID *uint64, itemID, locationID uint64, batchID *uint64, delta float64) (*StockBalance, error)
	ListByItemFunc func(ctx context.Context, itemID uint64) ([]*StockBalance, error)
	ListAllFunc    func(ctx context.Context, organizationID *uint64) ([]*StockBalance, error)
}

func (m StockBalanceDAOMock) FindByKey(ctx context.Context, itemID, locationID uint64, batchID *uint64) (*StockBalance, error) {
	if m.FindByKeyFunc != nil {
		return m.FindByKeyFunc(ctx, itemID, locationID, batchID)
	}
	return nil, nil
}

func (m StockBalanceDAOMock) FindByKeyTx(ctx context.Context, tx *gorm.DB, itemID, locationID uint64, batchID *uint64, forUpdate bool) (*StockBalance, error) {
	return m.FindByKey(ctx, itemID, locationID, batchID)
}

func (m StockBalanceDAOMock) Upsert(ctx context.Context, organizationID *uint64, itemID, locationID uint64, batchID *uint64, delta float64) (*StockBalance, error) {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(ctx, organizationID, itemID, locationID, batchID, delta)
	}
	return nil, nil
}

func (m StockBalanceDAOMock) UpsertTx(ctx context.Context, tx *gorm.DB, organizationID *uint64, itemID, locationID uint64, batchID *uint64, delta float64) (*StockBalance, error) {
	if m.UpsertTxFunc != nil {
		return m.UpsertTxFunc(ctx, tx, organizationID, itemID, locationID, batchID, delta)
	}
	return m.Upsert(ctx, organizationID, itemID, locationID, batchID, delta)
}

func (m StockBalanceDAOMock) ListByItem(ctx context.Context, itemID uint64) ([]*StockBalance, error) {
	if m.ListByItemFunc != nil {
		return m.ListByItemFunc(ctx, itemID)
	}
	return []*StockBalance{}, nil
}

func (m StockBalanceDAOMock) ListAll(ctx context.Context, organizationID *uint64) ([]*StockBalance, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(ctx, organizationID)
	}
	return []*StockBalance{}, nil
}

type BatchDAOMock struct {
	dao.CRUDMock[Batch]
	ListByItemFunc func(ctx context.Context, itemID uint64) ([]*Batch, error)
}

func (m BatchDAOMock) ListByItem(ctx context.Context, itemID uint64) ([]*Batch, error) {
	if m.ListByItemFunc != nil {
		return m.ListByItemFunc(ctx, itemID)
	}
	return []*Batch{}, nil
}

func (m BatchDAOMock) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[Batch], error) {
	return m.List(ctx, q)
}

func (m BatchDAOMock) FindInOrg(ctx context.Context, id, organizationID uint64) (*Batch, error) {
	return m.Find(ctx, id)
}

type StockHoldDAOMock struct {
	dao.CRUDMock[StockHold]
	ListByMovementFunc    func(ctx context.Context, movementID uint64) ([]*StockHold, error)
	DeleteByMovementFunc  func(ctx context.Context, movementID uint64) error
	ReserveFunc           func(ctx context.Context, quantID uint64, movementID *uint64, qty float64) (*StockHold, error)
	ReleaseFunc           func(ctx context.Context, reservationID uint64) error
	ReleaseByMovementFunc func(ctx context.Context, movementID uint64) error
}

func (m StockHoldDAOMock) ListByMovement(ctx context.Context, movementID uint64) ([]*StockHold, error) {
	if m.ListByMovementFunc != nil {
		return m.ListByMovementFunc(ctx, movementID)
	}
	return []*StockHold{}, nil
}

func (m StockHoldDAOMock) DeleteByMove(ctx context.Context, movementID uint64) error {
	if m.DeleteByMovementFunc != nil {
		return m.DeleteByMovementFunc(ctx, movementID)
	}
	return nil
}

func (m StockHoldDAOMock) Reserve(ctx context.Context, quantID uint64, movementID *uint64, qty float64) (*StockHold, error) {
	if m.ReserveFunc != nil {
		return m.ReserveFunc(ctx, quantID, movementID, qty)
	}
	return &StockHold{Base: model.Base{ID: 1}, BalanceID: quantID, MovementID: movementID, Qty: qty}, nil
}

func (m StockHoldDAOMock) Release(ctx context.Context, reservationID uint64) error {
	if m.ReleaseFunc != nil {
		return m.ReleaseFunc(ctx, reservationID)
	}
	return nil
}

func (m StockHoldDAOMock) ReleaseByMovement(ctx context.Context, movementID uint64) error {
	if m.ReleaseByMovementFunc != nil {
		return m.ReleaseByMovementFunc(ctx, movementID)
	}
	return nil
}

func (m StockHoldDAOMock) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[StockHold], error) {
	return m.List(ctx, q)
}

func (m StockHoldDAOMock) FindInOrg(ctx context.Context, id, organizationID uint64) (*StockHold, error) {
	return m.Find(ctx, id)
}

func (m StockHoldDAOMock) ReleaseByMovementInOrg(ctx context.Context, organizationID, movementID uint64) error {
	return m.ReleaseByMovement(ctx, movementID)
}

type CostLayerDAOMock struct {
	dao.CRUDMock[CostLayer]
	ListOpenByItemFunc            func(ctx context.Context, itemID uint64) ([]*CostLayer, error)
	ListOpenByItemForUpdateTxFunc func(ctx context.Context, tx *gorm.DB, itemID uint64) ([]*CostLayer, error)
	ListOpenByItemInOrgFunc       func(ctx context.Context, itemID, organizationID uint64) ([]*CostLayer, error)
	ValueForItemFunc              func(ctx context.Context, itemID uint64) (float64, error)
	ListByMovementFunc            func(ctx context.Context, movementID uint64) ([]*CostLayer, error)
	CreateTxFunc                  func(ctx context.Context, tx *gorm.DB, layer *CostLayer) (*CostLayer, error)
	UpdateTxFunc                  func(ctx context.Context, tx *gorm.DB, layer *CostLayer) (*CostLayer, error)
}

func (m CostLayerDAOMock) ListOpenByItemInOrg(ctx context.Context, itemID, organizationID uint64) ([]*CostLayer, error) {
	if m.ListOpenByItemInOrgFunc != nil {
		return m.ListOpenByItemInOrgFunc(ctx, itemID, organizationID)
	}
	return nil, nil
}

func (m CostLayerDAOMock) ListOpenByItem(ctx context.Context, itemID uint64) ([]*CostLayer, error) {
	if m.ListOpenByItemFunc != nil {
		return m.ListOpenByItemFunc(ctx, itemID)
	}
	return []*CostLayer{}, nil
}

func (m CostLayerDAOMock) ListOpenByItemForUpdateTx(ctx context.Context, _ *gorm.DB, itemID uint64) ([]*CostLayer, error) {
	if m.ListOpenByItemForUpdateTxFunc != nil {
		return m.ListOpenByItemForUpdateTxFunc(ctx, nil, itemID)
	}
	return m.ListOpenByItem(ctx, itemID)
}

func (m CostLayerDAOMock) ValueForItem(ctx context.Context, itemID uint64) (float64, error) {
	if m.ValueForItemFunc != nil {
		return m.ValueForItemFunc(ctx, itemID)
	}
	return 0, nil
}

func (m CostLayerDAOMock) ListByMovement(ctx context.Context, movementID uint64) ([]*CostLayer, error) {
	if m.ListByMovementFunc != nil {
		return m.ListByMovementFunc(ctx, movementID)
	}
	return []*CostLayer{}, nil
}

func (m CostLayerDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, layer *CostLayer) (*CostLayer, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, layer)
	}
	return m.Create(ctx, layer)
}

func (m CostLayerDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, layer *CostLayer) (*CostLayer, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, layer)
	}
	return m.Update(ctx, layer)
}

type ReorderRuleDAOMock struct {
	dao.CRUDMock[ReorderRule]
	ListActiveFunc func(ctx context.Context) ([]*ReorderRule, error)
}

func (m ReorderRuleDAOMock) ListActive(ctx context.Context) ([]*ReorderRule, error) {
	if m.ListActiveFunc != nil {
		return m.ListActiveFunc(ctx)
	}
	return []*ReorderRule{}, nil
}

func (m ReorderRuleDAOMock) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[ReorderRule], error) {
	return m.List(ctx, q)
}

func (m ReorderRuleDAOMock) ListActiveInOrg(ctx context.Context, organizationID uint64) ([]*ReorderRule, error) {
	return m.ListActive(ctx)
}

func (m ReorderRuleDAOMock) FindInOrg(ctx context.Context, id, organizationID uint64) (*ReorderRule, error) {
	return m.Find(ctx, id)
}

type StockCountDAOMock struct {
	dao.CRUDMock[StockCount]
	CreateWithLinesFunc func(ctx context.Context, count *StockCount, lines []*StockCountLine) (*StockCount, error)
	UpdateTxFunc        func(ctx context.Context, tx *gorm.DB, count *StockCount) (*StockCount, error)
}

func (m StockCountDAOMock) CreateWithLines(ctx context.Context, count *StockCount, lines []*StockCountLine) (*StockCount, error) {
	if m.CreateWithLinesFunc != nil {
		return m.CreateWithLinesFunc(ctx, count, lines)
	}
	return count, nil
}

func (m StockCountDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, count *StockCount) (*StockCount, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, count)
	}
	return m.Update(ctx, count)
}

type StockCountLineDAOMock struct {
	dao.CRUDMock[StockCountLine]
	ListByCountFunc func(ctx context.Context, stockCountID uint64) ([]*StockCountLine, error)
}

func (m StockCountLineDAOMock) ListByCount(ctx context.Context, stockCountID uint64) ([]*StockCountLine, error) {
	if m.ListByCountFunc != nil {
		return m.ListByCountFunc(ctx, stockCountID)
	}
	return []*StockCountLine{}, nil
}

type WarehouseTransferDAOMock struct {
	dao.CRUDMock[WarehouseTransfer]
	CreateWithShipmentsFunc func(ctx context.Context, transfer *WarehouseTransfer, outShipment *Shipment, outMovements []*StockMovement, inShipment *Shipment, inMovements []*StockMovement) (*WarehouseTransfer, error)
	UpdateTxFunc            func(ctx context.Context, tx *gorm.DB, transfer *WarehouseTransfer) (*WarehouseTransfer, error)
}

func (m WarehouseTransferDAOMock) CreateWithShipments(ctx context.Context, transfer *WarehouseTransfer, outShipment *Shipment, outMovements []*StockMovement, inShipment *Shipment, inMovements []*StockMovement) (*WarehouseTransfer, error) {
	if m.CreateWithShipmentsFunc != nil {
		return m.CreateWithShipmentsFunc(ctx, transfer, outShipment, outMovements, inShipment, inMovements)
	}
	return transfer, nil
}

func (m WarehouseTransferDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, transfer *WarehouseTransfer) (*WarehouseTransfer, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, transfer)
	}
	return m.Update(ctx, transfer)
}

type ItemResolverMock struct {
	ResolveFunc func(ctx context.Context, variantID uint64) (ResolvedItem, error)
}

func (m ItemResolverMock) Resolve(ctx context.Context, variantID uint64) (ResolvedItem, error) {
	if m.ResolveFunc != nil {
		return m.ResolveFunc(ctx, variantID)
	}
	return ResolvedItem{Tracking: "none"}, nil
}

type PosterMock struct {
	PostFunc      func(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error)
	PostTxFunc    func(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error)
	ReverseFunc   func(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
	ReverseTxFunc func(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
}

func (m PosterMock) Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostFunc != nil {
		return m.PostFunc(ctx, request)
	}
	return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
}

func (m PosterMock) PostTx(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostTxFunc != nil {
		return m.PostTxFunc(ctx, tx, request)
	}
	return m.Post(ctx, request)
}

func (m PosterMock) Reverse(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	if m.ReverseFunc != nil {
		return m.ReverseFunc(ctx, request)
	}
	return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
}

func (m PosterMock) ReverseTx(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	if m.ReverseTxFunc != nil {
		return m.ReverseTxFunc(ctx, tx, request)
	}
	return m.Reverse(ctx, request)
}

type InboundCostDAOMock struct {
	dao.CRUDMock[InboundCost]
	CreateTxFunc func(ctx context.Context, tx *gorm.DB, cost *InboundCost) (*InboundCost, error)
	UpdateTxFunc func(ctx context.Context, tx *gorm.DB, cost *InboundCost) (*InboundCost, error)
}

func (m InboundCostDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, cost *InboundCost) (*InboundCost, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, cost)
	}
	return cost, nil
}

func (m InboundCostDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, cost *InboundCost) (*InboundCost, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, cost)
	}
	return cost, nil
}

type InboundCostLineDAOMock struct {
	dao.CRUDMock[InboundCostLine]
	ListByInboundCostFunc func(ctx context.Context, costID uint64) ([]*InboundCostLine, error)
	CreateTxFunc          func(ctx context.Context, tx *gorm.DB, line *InboundCostLine) (*InboundCostLine, error)
}

func (m InboundCostLineDAOMock) ListByInboundCost(ctx context.Context, costID uint64) ([]*InboundCostLine, error) {
	if m.ListByInboundCostFunc != nil {
		return m.ListByInboundCostFunc(ctx, costID)
	}
	return []*InboundCostLine{}, nil
}

func (m InboundCostLineDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, line *InboundCostLine) (*InboundCostLine, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, line)
	}
	return line, nil
}

type InboundCostAdjustmentDAOMock struct {
	dao.CRUDMock[InboundCostAdjustment]
	ListByInboundCostFunc func(ctx context.Context, costID uint64) ([]*InboundCostAdjustment, error)
	CreateTxFunc          func(ctx context.Context, tx *gorm.DB, adjustment *InboundCostAdjustment) (*InboundCostAdjustment, error)
}

func (m InboundCostAdjustmentDAOMock) ListByInboundCost(ctx context.Context, costID uint64) ([]*InboundCostAdjustment, error) {
	if m.ListByInboundCostFunc != nil {
		return m.ListByInboundCostFunc(ctx, costID)
	}
	return []*InboundCostAdjustment{}, nil
}

func (m InboundCostAdjustmentDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, adjustment *InboundCostAdjustment) (*InboundCostAdjustment, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, adjustment)
	}
	return adjustment, nil
}

type InboundCostConfigSourceMock struct {
	JournalIDFunc func(ctx context.Context, organizationID uint64) (uint64, error)
}

func (m InboundCostConfigSourceMock) JournalID(ctx context.Context, organizationID uint64) (uint64, error) {
	if m.JournalIDFunc != nil {
		return m.JournalIDFunc(ctx, organizationID)
	}
	return 0, nil
}
