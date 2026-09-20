package interorganization

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/sales"
	"gorm.io/gorm"
)

type PurchaseOrderCreator interface {
	Create(ctx context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error)
}

type DropShipService struct {
	links     DropshipLinkDAO
	poCreate  PurchaseOrderCreator
	poOrders  procurement.PurchaseOrderDAO
	poLines   procurement.PurchaseOrderLineDAO
	soOrders  sales.SaleOrderDAO
	soLines   sales.SaleOrderLineDAO
	movements inventory.StockMovementDAO
	locations inventory.StockLocationDAO
	resolver  inventory.ItemResolver
	poster    accounting.Poster
	tx        db.Transactioner
}

func NewDropShipService(
	links DropshipLinkDAO,
	poCreate PurchaseOrderCreator,
	poOrders procurement.PurchaseOrderDAO,
	poLines procurement.PurchaseOrderLineDAO,
	soOrders sales.SaleOrderDAO,
	soLines sales.SaleOrderLineDAO,
	movements inventory.StockMovementDAO,
	locations inventory.StockLocationDAO,
	resolver inventory.ItemResolver,
	poster accounting.Poster,
	tx db.Transactioner,
) DropShipService {
	return DropShipService{
		links:     links,
		poCreate:  poCreate,
		poOrders:  poOrders,
		poLines:   poLines,
		soOrders:  soOrders,
		soLines:   soLines,
		movements: movements,
		locations: locations,
		resolver:  resolver,
		poster:    poster,
		tx:        tx,
	}
}

type CreateDropshipOrderRequest struct {
	OrganizationID uint64
	SaleOrderID    uint64
	SupplierID     uint64
	Date           time.Time
	DestLocationID *uint64
}

func (s DropShipService) Create(ctx context.Context, request CreateDropshipOrderRequest) (*procurement.PurchaseOrder, []*DropshipLink, error) {
	so, err := s.soOrders.Find(ctx, request.SaleOrderID)
	if err != nil {
		return nil, nil, err
	}
	if so == nil || so.OrganizationID == nil || *so.OrganizationID != request.OrganizationID {
		return nil, nil, ErrDropShipSourceNotFound
	}
	if so.State == sales.OrderStateCancelled {
		return nil, nil, ErrDropShipSourceNotActive
	}

	soLines, err := s.soLines.ListByOrder(ctx, so.ID)
	if err != nil {
		return nil, nil, err
	}
	source := make([]*sales.SaleOrderLine, 0, len(soLines))
	for _, line := range soLines {
		if line.ItemID == nil || line.QtyOrdered <= 0 {
			continue
		}
		source = append(source, line)
	}
	if len(source) == 0 {
		return nil, nil, ErrDropShipNoLines
	}

	dest := request.DestLocationID
	if dest == nil {
		location, err := s.locationByUsage(ctx, request.OrganizationID, "customer")
		if err != nil {
			return nil, nil, err
		}
		if location == nil {
			return nil, nil, ErrDropShipDestinationMiss
		}
		dest = &location.ID
	}

	poLines := make([]*procurement.PurchaseOrderLine, 0, len(source))
	for _, line := range source {
		poLines = append(poLines, &procurement.PurchaseOrderLine{
			ItemID:      line.ItemID,
			Description: line.Description,
			QtyOrdered:  line.QtyOrdered,
			UnitID:      line.UnitID,
			UnitPrice:   line.UnitPrice,
			DiscountPct: line.DiscountPct,
		})
	}

	po, err := s.poCreate.Create(ctx, &procurement.PurchaseOrder{
		OrganizationID: helper.Ptr(request.OrganizationID),
		SupplierID:     request.SupplierID,
		DestLocationID: dest,
		CurrencyCode:   so.CurrencyCode,
		PaymentTermID:  so.PaymentTermID,
		OrderDate:      &request.Date,
	}, poLines)
	if err != nil {
		return nil, nil, err
	}

	created, err := s.poLines.ListByOrder(ctx, po.ID)
	if err != nil {
		return nil, nil, err
	}

	links := make([]*DropshipLink, 0, len(created))
	for i := range created {
		link, err := s.links.Create(ctx, &DropshipLink{
			SaleOrderLineID:     source[i].ID,
			PurchaseOrderLineID: created[i].ID,
		})
		if err != nil {
			return nil, nil, err
		}
		links = append(links, link)
	}
	return po, links, nil
}

type ReceiveDropshipRequest struct {
	OrganizationID  uint64
	PurchaseOrderID uint64
	JournalID       uint64
	Date            time.Time
}

func (s DropShipService) Receive(ctx context.Context, request ReceiveDropshipRequest) (*procurement.PurchaseOrder, error) {
	po, err := s.poOrders.Find(ctx, request.PurchaseOrderID)
	if err != nil {
		return nil, err
	}
	if po == nil || po.OrganizationID == nil || *po.OrganizationID != request.OrganizationID {
		return nil, ErrDropShipOrderNotFound
	}
	if po.State == procurement.PurchaseOrderStateDone {
		return nil, ErrDropShipAlreadyReceived
	}
	if po.State == procurement.PurchaseOrderStateCancelled {
		return nil, ErrDropShipAlreadyReceived
	}

	lines, err := s.poLines.ListByOrder(ctx, po.ID)
	if err != nil {
		return nil, err
	}
	links, err := s.links.ListByPurchaseOrder(ctx, po.ID)
	if err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return nil, ErrDropShipNoLines
	}

	linksByLine := make(map[uint64]*DropshipLink, len(links))
	var so *sales.SaleOrder
	for _, link := range links {
		linksByLine[link.PurchaseOrderLineID] = link
		soLine, err := s.soLines.Find(ctx, link.SaleOrderLineID)
		if err != nil {
			return nil, err
		}
		if soLine == nil {
			continue
		}
		if so == nil {
			so, err = s.soOrders.Find(ctx, soLine.OrderID)
			if err != nil {
				return nil, err
			}
		}
	}
	if so == nil {
		return nil, ErrDropShipSourceNotFound
	}

	allSoLines, err := s.soLines.ListByOrder(ctx, so.ID)
	if err != nil {
		return nil, err
	}
	soLinesByID := make(map[uint64]*sales.SaleOrderLine, len(allSoLines))
	for _, line := range allSoLines {
		soLinesByID[line.ID] = line
	}

	supplier, err := s.locationByUsage(ctx, request.OrganizationID, "supplier")
	if err != nil {
		return nil, err
	}
	if supplier == nil {
		return nil, ErrDropShipSupplierMissing
	}
	if po.DestLocationID == nil {
		return nil, ErrDropShipDestinationMiss
	}
	dest := *po.DestLocationID

	var moveIDs []uint64
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		for _, line := range lines {
			link := linksByLine[line.ID]
			if link == nil || line.ItemID == nil {
				continue
			}
			qty := line.QtyOrdered - line.QtyReceived
			if qty <= 0 {
				continue
			}
			movement, err := s.movements.CreateTx(ctx, tx, &inventory.StockMovement{
				OrganizationID: po.OrganizationID,
				ItemID:         *line.ItemID,
				Qty:            qty,
				UnitID:         line.UnitID,
				SrcLocationID:  supplier.ID,
				DstLocationID:  dest,
				State:          inventory.MovementStateConfirmed,
				OriginType:     helper.Ptr(accounting.OriginTypePurchaseOrder),
				OriginID:       helper.Ptr(po.ID),
				ScheduledDate:  &request.Date,
			})
			if err != nil {
				return err
			}
			moveIDs = append(moveIDs, movement.ID)

			resolved, err := s.resolver.Resolve(ctx, *line.ItemID)
			if err != nil {
				return err
			}
			lineAmount := amount.FromFloat64(qty * line.UnitPrice)
			if _, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
				OrganizationID: *po.OrganizationID,
				JournalID:      request.JournalID,
				Date:           request.Date,
				Ref:            fmt.Sprintf("DRP/%d", po.ID),
				OriginType:     accounting.OriginTypeDropship,
				OriginID:       movement.ID,
				Description:    "Drop-ship receipt",
				Lines: []accounting.PostingLine{
					{AccountID: resolved.CogsAccountID, Name: "COGS", Debit: lineAmount},
					{AccountID: resolved.StockInputAccountID, Name: "Stock Input", Credit: lineAmount},
				},
			}); err != nil {
				return err
			}

			link.StockMovementID = &movement.ID
			if _, err := s.links.UpdateTx(ctx, tx, link); err != nil {
				return err
			}

			line.QtyReceived += qty
			if _, err := s.poLines.UpdateTx(ctx, tx, line); err != nil {
				return err
			}

			if soLine := soLinesByID[link.SaleOrderLineID]; soLine != nil {
				soLine.QtyDelivered += qty
				if _, err := s.soLines.UpdateTx(ctx, tx, soLine); err != nil {
					return err
				}
			}
		}
		if len(moveIDs) == 0 {
			return ErrDropShipNoLines
		}
		locked, err := inventory.FindApplicableMovementsTx(ctx, tx, s.movements, moveIDs)
		if err != nil {
			return err
		}
		for _, m := range locked {
			m.State = inventory.MovementStateDone
			m.DateDone = &request.Date
		}
		if err := s.movements.ApplyAllTx(ctx, tx, locked); err != nil {
			return err
		}

		so.DeliveryStatus = recomputeDeliveryStatus(allSoLines)
		if _, err := s.soOrders.UpdateTx(ctx, tx, so); err != nil {
			return err
		}
		po.State = procurement.PurchaseOrderStateDone
		po.ReceiptStatus = procurement.PurchaseOrderReceiptStatusDone
		if _, err := s.poOrders.UpdateTx(ctx, tx, po); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return po, nil
}

func (s DropShipService) locationByUsage(ctx context.Context, organizationID uint64, usage string) (*reference.StockLocation, error) {
	locations, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	for _, location := range locations.Items {
		if location.Usage == usage {
			return location, nil
		}
	}
	return nil, nil
}

func recomputeDeliveryStatus(lines []*sales.SaleOrderLine) string {
	ordered, delivered := amount.Zero(), amount.Zero()
	for _, line := range lines {
		ordered = ordered.Add(amount.FromFloat64(line.QtyOrdered))
		delivered = delivered.Add(amount.FromFloat64(line.QtyDelivered))
	}
	switch {
	case !delivered.IsPositive():
		return sales.DeliveryStatusPending
	case delivered.Equal(ordered):
		return sales.DeliveryStatusDone
	default:
		return sales.DeliveryStatusPartial
	}
}
