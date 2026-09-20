package returns

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/sales"
	"gorm.io/gorm"
)

const (
	OriginSaleOrder     = "sale_order"
	OriginPurchaseOrder = "purchase_order"
	OriginRMA           = "rma"
)

type originOrderLookup struct {
	saleOrders     SaleOrderLookup
	purchaseOrders PurchaseOrderLookup
}

func NewOriginOrderLookup(saleOrders SaleOrderLookup, purchaseOrders PurchaseOrderLookup) OriginOrderLookup {
	return originOrderLookup{saleOrders: saleOrders, purchaseOrders: purchaseOrders}
}

func (l originOrderLookup) CustomerOrder(ctx context.Context, orderID uint64) (uint64, error) {
	order, err := l.saleOrders.Find(ctx, orderID)
	if err != nil {
		return 0, err
	}
	if order == nil || (order.State != sales.OrderStateConfirmed && order.State != sales.OrderStateDone) {
		return 0, ErrRMAOrder
	}
	return order.ContactID, nil
}

func (l originOrderLookup) VendorOrder(ctx context.Context, orderID uint64) (uint64, error) {
	order, err := l.purchaseOrders.Find(ctx, orderID)
	if err != nil {
		return 0, err
	}
	if order == nil || (order.State != procurement.PurchaseOrderStateConfirmed && order.State != procurement.PurchaseOrderStateDone) {
		return 0, ErrRMAOrder
	}
	return order.SupplierID, nil
}

type RMALineRequest struct {
	ItemID      uint64
	Qty         float64
	BatchID     *uint64
	Disposition string
}

type CreateRMARequest struct {
	OrganizationID  uint64
	Type            string
	ContactID       uint64
	OriginOrderType string
	OriginOrderID   uint64
	Reason          string
	Lines           []RMALineRequest
}

type RMAService struct {
	rmas           RMADAO
	lines          RMALineDAO
	orders         OriginOrderLookup
	movements      StockMovementLookup
	locations      StockLocationLookup
	layers         StockLayerLookup
	valuer         ReturnValuer
	invoices       InvoiceLookup
	credits        CreditNoteEngine
	saleReturns    SaleOrderReturnRecorder
	purchaseReturn PurchaseOrderReturnRecorder
	sequences      sequence.Service
	tx             db.Transactioner
}

func NewRMAService(
	rmas RMADAO,
	lines RMALineDAO,
	orders OriginOrderLookup,
	movements StockMovementLookup,
	locations StockLocationLookup,
	layers StockLayerLookup,
	valuer ReturnValuer,
	invoices InvoiceLookup,
	credits CreditNoteEngine,
	saleReturns SaleOrderReturnRecorder,
	purchaseReturn PurchaseOrderReturnRecorder,
	sequences sequence.Service,
	tx db.Transactioner,
) RMAService {
	return RMAService{
		rmas:           rmas,
		lines:          lines,
		orders:         orders,
		movements:      movements,
		locations:      locations,
		layers:         layers,
		valuer:         valuer,
		invoices:       invoices,
		credits:        credits,
		saleReturns:    saleReturns,
		purchaseReturn: purchaseReturn,
		sequences:      sequences,
		tx:             tx,
	}
}

func (s RMAService) List(ctx context.Context, q *query.Query) (*query.Page[RMA], error) {
	return s.rmas.List(ctx, q)
}

func (s RMAService) Find(ctx context.Context, id uint64) (*RMA, error) {
	return s.rmas.Find(ctx, id)
}

func (s RMAService) ListLines(ctx context.Context, rmaID uint64) ([]*RMALine, error) {
	return s.lines.ListByRMA(ctx, rmaID)
}

func (s RMAService) Create(ctx context.Context, request CreateRMARequest) (*RMA, error) {
	if request.Type != TypeCustomerReturn && request.Type != TypeVendorReturn {
		return nil, ErrRMAType
	}
	if len(request.Lines) == 0 {
		return nil, ErrRMALines
	}
	if err := s.validateLines(request.Lines); err != nil {
		return nil, err
	}
	if request.OriginOrderType != OriginSaleOrder && request.OriginOrderType != OriginPurchaseOrder {
		return nil, ErrRMAOrder
	}

	var contactID uint64
	var err error
	switch request.Type {
	case TypeCustomerReturn:
		contactID, err = s.orders.CustomerOrder(ctx, request.OriginOrderID)
	default:
		contactID, err = s.orders.VendorOrder(ctx, request.OriginOrderID)
	}
	if err != nil {
		return nil, err
	}
	if contactID != request.ContactID {
		return nil, ErrRMAContactMismatch
	}

	name, err := s.sequences.Next(ctx, request.OrganizationID, SequenceRMACode)
	if err != nil {
		return nil, err
	}

	rma := &RMA{
		OrganizationID:  &request.OrganizationID,
		Name:            &name,
		Type:            request.Type,
		ContactID:       request.ContactID,
		OriginOrderType: &request.OriginOrderType,
		OriginOrderID:   &request.OriginOrderID,
		Reason:          &request.Reason,
		State:           StateDraft,
	}
	lines := make([]*RMALine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = &RMALine{
			ItemID:      line.ItemID,
			Qty:         line.Qty,
			BatchID:     line.BatchID,
			Disposition: line.Disposition,
		}
	}

	var created *RMA
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		var err error
		created, err = s.rmas.CreateWithLinesTx(ctx, tx, rma, lines)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s RMAService) Confirm(ctx context.Context, rmaID uint64) (*RMA, error) {
	rma, err := s.rmas.Find(ctx, rmaID)
	if err != nil {
		return nil, err
	}
	if rma == nil {
		return nil, ErrRMANotFound
	}
	if rma.State != StateDraft {
		return nil, ErrRMAState
	}
	rma.State = StateConfirmed
	if _, err := s.rmas.Update(ctx, rma); err != nil {
		return nil, err
	}
	return rma, nil
}

func (s RMAService) Cancel(ctx context.Context, rmaID uint64) (*RMA, error) {
	rma, err := s.rmas.Find(ctx, rmaID)
	if err != nil {
		return nil, err
	}
	if rma == nil {
		return nil, ErrRMANotFound
	}
	if rma.State != StateDraft && rma.State != StateConfirmed {
		return nil, ErrRMAState
	}
	rma.State = StateCancelled
	if _, err := s.rmas.Update(ctx, rma); err != nil {
		return nil, err
	}
	return rma, nil
}

func (s RMAService) Receive(ctx context.Context, rmaID, journalID uint64, date time.Time) (*RMA, error) {
	rma, err := s.rmas.Find(ctx, rmaID)
	if err != nil {
		return nil, err
	}
	if rma == nil {
		return nil, ErrRMANotFound
	}
	if rma.State != StateConfirmed {
		return nil, ErrRMAState
	}
	if rma.OrganizationID == nil || rma.OriginOrderID == nil {
		return nil, ErrRMAOrganization
	}

	rmaLines, err := s.lines.ListByRMA(ctx, rmaID)
	if err != nil {
		return nil, err
	}
	if len(rmaLines) == 0 {
		return nil, ErrRMALines
	}

	originType := OriginSaleOrder
	if rma.Type == TypeVendorReturn {
		originType = OriginPurchaseOrder
	}
	originalMoves, err := s.movements.ListByOrigin(ctx, originType, *rma.OriginOrderID)
	if err != nil {
		return nil, err
	}

	for _, line := range rmaLines {
		original := findOriginalMove(originalMoves, line.ItemID)
		if original == nil {
			return nil, ErrRMAMoveNotFound
		}

		srcLocationID := original.DstLocationID
		dstLocationID := original.SrcLocationID
		if rma.Type == TypeCustomerReturn && line.Disposition == DispositionScrap {
			scrapLocation, err := s.scrapLocation(ctx)
			if err != nil {
				return nil, err
			}
			dstLocationID = scrapLocation
		}

		movement, err := s.movements.Create(ctx, &inventory.StockMovement{
			OrganizationID: rma.OrganizationID,
			ItemID:         line.ItemID,
			Qty:            line.Qty,
			UnitID:         original.UnitID,
			SrcLocationID:  srcLocationID,
			DstLocationID:  dstLocationID,
			BatchID:        line.BatchID,
			State:          inventory.MovementStateConfirmed,
			OriginType:     helper.Ptr(OriginRMA),
			OriginID:       &rma.ID,
			ScheduledDate:  &date,
		})
		if err != nil {
			return nil, err
		}
		line.StockMovementID = &movement.ID

		switch {
		case rma.Type == TypeVendorReturn:
			if _, err := s.valuer.ReturnToSupplier(ctx, movement.ID, journalID, date); err != nil {
				return nil, err
			}
		case line.Disposition == DispositionScrap:
			if err := s.applyScrapMove(ctx, movement.ID, date); err != nil {
				return nil, err
			}
		default:
			unitCost, err := s.originalUnitCost(ctx, original)
			if err != nil {
				return nil, err
			}
			if _, err := s.valuer.Restock(ctx, movement.ID, unitCost, journalID, date); err != nil {
				return nil, err
			}
		}

		if err := s.recordReturn(ctx, rma, line); err != nil {
			return nil, err
		}

		if _, err := s.lines.Update(ctx, line); err != nil {
			return nil, err
		}
	}

	rma.State = StateReceived
	if _, err := s.rmas.Update(ctx, rma); err != nil {
		return nil, err
	}
	return rma, nil
}

func (s RMAService) Refund(ctx context.Context, rmaID, journalID uint64, date time.Time, reference string) (*RMA, error) {
	rma, err := s.rmas.Find(ctx, rmaID)
	if err != nil {
		return nil, err
	}
	if rma == nil {
		return nil, ErrRMANotFound
	}
	if rma.State != StateReceived {
		return nil, ErrRMAState
	}
	if rma.OrganizationID == nil || rma.OriginOrderID == nil {
		return nil, ErrRMAOrganization
	}

	var originalInvoice *accounting.Invoice
	switch rma.Type {
	case TypeCustomerReturn:
		originalInvoice, err = s.invoices.FindBySaleOrder(ctx, *rma.OriginOrderID)
	default:
		originalInvoice, err = s.invoices.FindByPurchaseOrder(ctx, *rma.OriginOrderID)
	}
	if err != nil {
		return nil, err
	}
	if originalInvoice == nil {
		return nil, ErrRMAInvoice
	}

	request := accounting.CreateCreditNoteRequest{
		OrganizationID:    *rma.OrganizationID,
		OriginalInvoiceID: originalInvoice.ID,
		JournalID:         journalID,
		Date:              date,
		Reference:         reference,
	}
	var creditNote *accounting.Invoice
	if rma.Type == TypeCustomerReturn {
		creditNote, err = s.credits.CreateCreditNote(ctx, request)
	} else {
		creditNote, err = s.credits.CreateVendorCreditNote(ctx, request)
	}
	if err != nil {
		return nil, err
	}

	rmaLines, err := s.lines.ListByRMA(ctx, rmaID)
	if err != nil {
		return nil, err
	}
	for _, line := range rmaLines {
		line.CreditNoteID = &creditNote.ID
		if _, err := s.lines.Update(ctx, line); err != nil {
			return nil, err
		}
	}

	rma.State = StateRefunded
	if _, err := s.rmas.Update(ctx, rma); err != nil {
		return nil, err
	}
	return rma, nil
}

func (s RMAService) Done(ctx context.Context, rmaID uint64) (*RMA, error) {
	rma, err := s.rmas.Find(ctx, rmaID)
	if err != nil {
		return nil, err
	}
	if rma == nil {
		return nil, ErrRMANotFound
	}
	if rma.State != StateRefunded {
		return nil, ErrRMAState
	}
	rma.State = StateDone
	if _, err := s.rmas.Update(ctx, rma); err != nil {
		return nil, err
	}
	return rma, nil
}

func (s RMAService) validateLines(lines []RMALineRequest) error {
	for _, line := range lines {
		if line.ItemID == 0 {
			return ErrRMALineProduct
		}
		if line.Qty <= 0 {
			return ErrRMALineQty
		}
		switch line.Disposition {
		case DispositionRestock, DispositionScrap, DispositionRepair, DispositionReplace:
		default:
			return ErrRMADisposition
		}
	}
	return nil
}

func (s RMAService) scrapLocation(ctx context.Context) (uint64, error) {
	page, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "usage", Operator: query.Equal, Value: "scrap"}}})
	if err != nil {
		return 0, err
	}
	locations := page.Items
	if len(locations) == 0 {
		return 0, ErrRMAScrapLocation
	}
	return locations[0].ID, nil
}

func (s RMAService) originalUnitCost(ctx context.Context, original *inventory.StockMovement) (amount.Amount, error) {
	layers, err := s.layers.ListByMovement(ctx, original.ID)
	if err != nil {
		return amount.Amount{}, err
	}
	for _, layer := range layers {
		if layer.UnitCost != nil {
			return amount.FromFloat64(*layer.UnitCost), nil
		}
	}
	return amount.Amount{}, ErrRMACost
}

func findOriginalMove(movements []*inventory.StockMovement, itemID uint64) *inventory.StockMovement {
	for _, movement := range movements {
		if movement.ItemID == itemID && movement.State == inventory.MovementStateDone {
			return movement
		}
	}
	return nil
}

func (s RMAService) applyScrapMove(ctx context.Context, moveID uint64, date time.Time) error {
	return s.tx.Run(ctx, func(tx *gorm.DB) error {
		locked, err := inventory.FindApplicableMovementsTx(ctx, tx, s.movements, []uint64{moveID})
		if err != nil {
			return err
		}
		locked[0].State = inventory.MovementStateDone
		locked[0].DateDone = &date
		_, err = s.movements.ApplyTx(ctx, tx, locked[0])
		return err
	})
}

func (s RMAService) recordReturn(ctx context.Context, rma *RMA, line *RMALine) error {
	if rma.OriginOrderID == nil {
		return nil
	}
	if rma.Type == TypeVendorReturn {
		return s.purchaseReturn.RecordReturn(ctx, *rma.OriginOrderID, line.ItemID, line.Qty)
	}
	return s.saleReturns.RecordReturn(ctx, *rma.OriginOrderID, line.ItemID, line.Qty)
}
