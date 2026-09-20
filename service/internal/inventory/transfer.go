package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type TransferLine struct {
	ItemID        uint64  `json:"item_id"`
	Qty           float64 `json:"qty"`
	BatchID       *uint64 `json:"batch_id"`
	SrcLocationID *uint64 `json:"src_location_id"`
	DstLocationID *uint64 `json:"dst_location_id"`
}

type pendingTransfer struct {
	movementID         uint64
	value              amount.Amount
	valuationAccountID uint64
}

type TransferService struct {
	transfers  WarehouseTransferDAO
	movements  StockMovementDAO
	locations  StockLocationDAO
	warehouses WarehouseDAO
	layers     CostLayerDAO
	ledger     LedgerService
	resolver   ItemResolver
	poster     Poster
	tx         db.Transactioner
}

func NewTransferService(
	transfers WarehouseTransferDAO,
	movements StockMovementDAO,
	locations StockLocationDAO,
	warehouses WarehouseDAO,
	layers CostLayerDAO,
	ledger LedgerService,
	resolver ItemResolver,
	poster Poster,
	tx db.Transactioner,
) TransferService {
	return TransferService{
		transfers:  transfers,
		movements:  movements,
		locations:  locations,
		warehouses: warehouses,
		layers:     layers,
		ledger:     ledger,
		resolver:   resolver,
		poster:     poster,
		tx:         tx,
	}
}

func (s TransferService) List(ctx context.Context, q *query.Query) (*query.Page[WarehouseTransfer], error) {
	return s.transfers.List(ctx, q)
}

func (s TransferService) Find(ctx context.Context, id uint64) (*WarehouseTransfer, error) {
	return s.transfers.Find(ctx, id)
}

func (s TransferService) Create(ctx context.Context, transfer *WarehouseTransfer, lines []TransferLine) (*WarehouseTransfer, error) {
	if transfer.SrcWarehouseID == transfer.DstWarehouseID {
		return nil, ErrSameWarehouse
	}
	if transfer.OrganizationID == nil {
		return nil, ErrOrganizationMissing
	}

	srcWarehouse, err := s.warehouses.Find(ctx, transfer.SrcWarehouseID)
	if err != nil {
		return nil, err
	}
	if srcWarehouse == nil || !helper.OwnedByOrg(srcWarehouse.OrganizationID, transfer.OrganizationID) {
		return nil, ErrWarehouseNotFound
	}
	dstWarehouse, err := s.warehouses.Find(ctx, transfer.DstWarehouseID)
	if err != nil {
		return nil, err
	}
	if dstWarehouse == nil || !helper.OwnedByOrg(dstWarehouse.OrganizationID, transfer.OrganizationID) {
		return nil, ErrWarehouseNotFound
	}

	transit, err := s.findTransit(ctx)
	if err != nil {
		return nil, err
	}
	if transit == nil {
		return nil, ErrTransitNotFound
	}

	name := transfer.Name

	outMovements := make([]*StockMovement, 0, len(lines))
	inMovements := make([]*StockMovement, 0, len(lines))
	var outSrc *uint64
	var inDst *uint64
	for _, line := range lines {
		if line.Qty <= 0 {
			return nil, ErrMovementQty
		}
		srcLocation, err := s.resolveLocation(ctx, line.SrcLocationID, transfer.OrganizationID, transfer.SrcWarehouseID)
		if err != nil {
			return nil, err
		}
		dstLocation, err := s.resolveLocation(ctx, line.DstLocationID, transfer.OrganizationID, transfer.DstWarehouseID)
		if err != nil {
			return nil, err
		}
		if outSrc == nil {
			outSrc = srcLocation
		}
		if inDst == nil {
			inDst = dstLocation
		}
		outMovements = append(outMovements, &StockMovement{
			OrganizationID: transfer.OrganizationID,
			ItemID:         line.ItemID,
			Qty:            line.Qty,
			BatchID:        line.BatchID,
			SrcLocationID:  *srcLocation,
			DstLocationID:  transit.ID,
			State:          MovementStateDraft,
			ScheduledDate:  transfer.ScheduledDate,
		})
		inMovements = append(inMovements, &StockMovement{
			OrganizationID: transfer.OrganizationID,
			ItemID:         line.ItemID,
			Qty:            line.Qty,
			BatchID:        line.BatchID,
			SrcLocationID:  transit.ID,
			DstLocationID:  *dstLocation,
			State:          MovementStateDraft,
			ScheduledDate:  transfer.ScheduledDate,
		})
	}

	outShipment := &Shipment{
		OrganizationID: transfer.OrganizationID,
		Name:           name,
		Type:           ShipmentTypeInternal,
		SrcLocationID:  outSrc,
		DstLocationID:  &transit.ID,
		State:          ShipmentStateDraft,
		ScheduledDate:  transfer.ScheduledDate,
	}
	inShipment := &Shipment{
		OrganizationID: transfer.OrganizationID,
		Name:           name,
		Type:           ShipmentTypeInternal,
		SrcLocationID:  &transit.ID,
		DstLocationID:  inDst,
		State:          ShipmentStateDraft,
		ScheduledDate:  transfer.ScheduledDate,
	}

	return s.transfers.CreateWithShipments(ctx, transfer, outShipment, outMovements, inShipment, inMovements)
}

func (s TransferService) Send(ctx context.Context, transferID uint64, journalID, transitAccountID uint64, date time.Time) error {
	transfer, err := s.transfers.Find(ctx, transferID)
	if err != nil {
		return err
	}
	if transfer == nil {
		return ErrWarehouseTransferNotFound
	}
	if transfer.State != TransferStateDraft && transfer.State != TransferStateSent {
		return ErrWarehouseTransferState
	}
	if transfer.OutShipmentID == nil {
		return ErrWarehouseTransferState
	}
	if transitAccountID == 0 {
		return ErrTransitAccount
	}
	if transfer.OrganizationID == nil {
		return ErrOrganizationMissing
	}

	outMovements, err := s.movements.ListByShipment(ctx, *transfer.OutShipmentID)
	if err != nil {
		return err
	}
	pending := make([]pendingTransfer, 0, len(outMovements))
	movementIDs := make([]uint64, 0, len(outMovements))
	for _, movement := range outMovements {
		onHand, err := s.ledger.OnHand(ctx, transfer.OrganizationID, movement.ItemID, movement.SrcLocationID)
		if err != nil {
			return err
		}
		if onHand < movement.Qty {
			return ErrInsufficientStock
		}

		value, resolved, err := s.transferValue(ctx, movement)
		if err != nil {
			return err
		}
		pending = append(pending, pendingTransfer{movementID: movement.ID, value: value, valuationAccountID: resolved.StockValuationAccountID})
		movementIDs = append(movementIDs, movement.ID)
	}

	if err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		locked, err := FindApplicableMovementsTx(ctx, tx, s.movements, movementIDs)
		if err != nil {
			return err
		}
		for _, m := range locked {
			m.State = MovementStateDone
			m.DateDone = &date
		}
		if err := s.movements.ApplyAllTx(ctx, tx, locked); err != nil {
			return err
		}
		for _, p := range pending {
			if !p.value.IsZero() {
				if _, err := s.postTransferGLTx(ctx, tx, transfer, p.movementID, journalID, transitAccountID, p.valuationAccountID, p.value, date, true); err != nil {
					return err
				}
			}
		}

		transfer.State = TransferStateInTransit
		_, err = s.transfers.UpdateTx(ctx, tx, transfer)
		return err
	}); err != nil {
		return err
	}
	return nil
}

func (s TransferService) Receive(ctx context.Context, transferID uint64, journalID, transitAccountID uint64, date time.Time) error {
	transfer, err := s.transfers.Find(ctx, transferID)
	if err != nil {
		return err
	}
	if transfer == nil {
		return ErrWarehouseTransferNotFound
	}
	if transfer.State != TransferStateInTransit {
		return ErrWarehouseTransferState
	}
	if transfer.InShipmentID == nil {
		return ErrWarehouseTransferState
	}
	if transitAccountID == 0 {
		return ErrTransitAccount
	}
	if transfer.OrganizationID == nil {
		return ErrOrganizationMissing
	}

	inMovements, err := s.movements.ListByShipment(ctx, *transfer.InShipmentID)
	if err != nil {
		return err
	}
	pending := make([]pendingTransfer, 0, len(inMovements))
	movementIDs := make([]uint64, 0, len(inMovements))
	for _, movement := range inMovements {
		value, resolved, err := s.transferValue(ctx, movement)
		if err != nil {
			return err
		}
		pending = append(pending, pendingTransfer{movementID: movement.ID, value: value, valuationAccountID: resolved.StockValuationAccountID})
		movementIDs = append(movementIDs, movement.ID)
	}

	if err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		locked, err := FindApplicableMovementsTx(ctx, tx, s.movements, movementIDs)
		if err != nil {
			return err
		}
		for _, m := range locked {
			m.State = MovementStateDone
			m.DateDone = &date
		}
		if err := s.movements.ApplyAllTx(ctx, tx, locked); err != nil {
			return err
		}
		for _, p := range pending {
			if !p.value.IsZero() {
				if _, err := s.postTransferGLTx(ctx, tx, transfer, p.movementID, journalID, transitAccountID, p.valuationAccountID, p.value, date, false); err != nil {
					return err
				}
			}
		}

		transfer.State = TransferStateReceived
		_, err = s.transfers.UpdateTx(ctx, tx, transfer)
		return err
	}); err != nil {
		return err
	}
	return nil
}

func (s TransferService) transferValue(ctx context.Context, movement *StockMovement) (amount.Amount, ResolvedItem, error) {
	resolved, err := s.resolver.Resolve(ctx, movement.ItemID)
	if err != nil {
		return amount.Amount{}, ResolvedItem{}, err
	}
	if resolved.StockValuationAccountID == 0 {
		return amount.Amount{}, ResolvedItem{}, ErrValuationAccount
	}
	unitCost, err := s.currentUnitCost(ctx, movement.ItemID)
	if err != nil {
		return amount.Amount{}, ResolvedItem{}, err
	}
	value := amount.FromFloat64(movement.Qty).Mul(unitCost).Round(4)
	return value, resolved, nil
}

func (s TransferService) postTransferGLTx(ctx context.Context, tx *gorm.DB, transfer *WarehouseTransfer, movementID uint64, journalID, transitAccountID, valuationAccountID uint64, value amount.Amount, date time.Time, outgoing bool) (*accounting.JournalEntry, error) {
	return s.poster.PostTx(ctx, tx, transferPostRequest(transfer, movementID, journalID, transitAccountID, valuationAccountID, value, date, outgoing))
}

func transferPostRequest(transfer *WarehouseTransfer, movementID uint64, journalID, transitAccountID, valuationAccountID uint64, value amount.Amount, date time.Time, outgoing bool) accounting.PostRequest {
	var lines []accounting.PostingLine
	if outgoing {
		lines = []accounting.PostingLine{
			{AccountID: transitAccountID, Name: "Goods in Transit", Debit: value},
			{AccountID: valuationAccountID, Name: "Inventory", Credit: value},
		}
	} else {
		lines = []accounting.PostingLine{
			{AccountID: valuationAccountID, Name: "Inventory", Debit: value},
			{AccountID: transitAccountID, Name: "Goods in Transit", Credit: value},
		}
	}
	return accounting.PostRequest{
		OrganizationID: *transfer.OrganizationID,
		JournalID:      journalID,
		Date:           date,
		Ref:            fmt.Sprintf("TRF/%d", transfer.ID),
		OriginType:     accounting.OriginTypeStockMovement,
		OriginID:       movementID,
		Description:    "Internal transfer",
		Lines:          lines,
	}
}

func (s TransferService) findTransit(ctx context.Context) (*reference.StockLocation, error) {
	page, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "usage", Operator: query.Equal, Value: "transit"}}})
	if err != nil {
		return nil, err
	}
	locations := page.Items
	if len(locations) == 0 {
		return nil, nil
	}
	return locations[0], nil
}

func (s TransferService) resolveLocation(ctx context.Context, explicit *uint64, organizationID *uint64, warehouseID uint64) (*uint64, error) {
	if explicit != nil {
		location, err := s.locations.Find(ctx, *explicit)
		if err != nil {
			return nil, err
		}
		if location == nil || !helper.OwnedByOrg(location.OrganizationID, organizationID) {
			return nil, ErrLocationNotFound
		}
		return explicit, nil
	}
	page, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "warehouse_id", Operator: query.Equal, Value: warehouseID}}})
	if err != nil {
		return nil, err
	}
	locations := page.Items
	for _, location := range locations {
		if location.Usage == "internal" && helper.OwnedByOrg(location.OrganizationID, organizationID) {
			return helper.Ptr(location.ID), nil
		}
	}
	return nil, ErrLocationRequired
}

func (s TransferService) currentUnitCost(ctx context.Context, itemID uint64) (amount.Amount, error) {
	openLayers, err := s.layers.ListOpenByItem(ctx, itemID)
	if err != nil {
		return amount.Amount{}, err
	}
	return currentUnitCost(openLayers), nil
}
