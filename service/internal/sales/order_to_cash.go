package sales

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type ShipEngine interface {
	Ship(ctx context.Context, moveID uint64, journalID uint64, date time.Time) (*inventory.CostLayer, error)
}

type InvoiceEngine interface {
	Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}

type InvoiceLookup interface {
	ListOpenByContact(ctx context.Context, contactID uint64) ([]*accounting.Invoice, error)
}

type PaymentEngine interface {
	Create(ctx context.Context, request accounting.CreatePaymentRequest) (*accounting.Payment, error)
}

func (s SaleOrderService) Deliver(ctx context.Context, orderID uint64, journalID uint64, date time.Time) (*SaleOrder, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.State != OrderStateConfirmed {
		return nil, ErrOrderState
	}
	if order.Name == nil {
		return nil, ErrOrderNotFound
	}

	lines, err := s.lines.ListByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrOrderNoLines
	}

	shipments, err := s.shipments.List(ctx, &query.Query{Filters: []query.Filter{{Field: "origin", Operator: query.Equal, Value: *order.Name}}})
	if err != nil {
		return nil, err
	}
	if len(shipments.Items) == 0 {
		return nil, ErrOrderShipmentNotFound
	}

	shippedByProduct := map[uint64]float64{}
	for _, shipment := range shipments.Items {
		if shipment.State != inventory.ShipmentStateAssigned {
			continue
		}
		movements, err := s.movements.ListByShipment(ctx, shipment.ID)
		if err != nil {
			return nil, err
		}
		for _, movement := range movements {
			if movement.State == inventory.MovementStateDone || movement.State == inventory.MovementStateCancelled {
				continue
			}
			if _, err := s.ship.Ship(ctx, movement.ID, journalID, date); err != nil {
				return nil, err
			}
			if err := s.reservations.ReleaseByMovement(ctx, movement.ID); err != nil {
				return nil, err
			}
			shippedByProduct[movement.ItemID] += movement.Qty
		}
		shipment.State = inventory.ShipmentStateDone
		if _, err := s.shipments.Update(ctx, shipment); err != nil {
			return nil, err
		}
	}

	if len(shippedByProduct) == 0 {
		return nil, ErrOrderNothingToDeliver
	}

	for _, line := range lines {
		if line.ItemID == nil {
			continue
		}
		if qty, ok := shippedByProduct[*line.ItemID]; ok {
			line.QtyDelivered += qty
			if _, err := s.lines.Update(ctx, line); err != nil {
				return nil, err
			}
		}
	}

	recomputeStatuses(order, lines)
	return s.orders.Update(ctx, order)
}

func (s SaleOrderService) CreateInvoice(ctx context.Context, orderID uint64, journalID uint64, date time.Time) (*accounting.Invoice, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.OrganizationID == nil {
		return nil, ErrOrderNotFound
	}

	lines, err := s.lines.ListByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrOrderNoLines
	}

	requests := make([]accounting.InvoiceLineRequest, 0, len(lines))
	toInvoice := map[uint64]float64{}
	for _, line := range lines {
		if line.ItemID == nil {
			continue
		}
		qty := line.QtyDelivered - line.QtyInvoiced
		if qty <= 0 {
			continue
		}
		accountID, err := s.productSvc.ResolveIncomeAccount(ctx, *line.ItemID)
		if err != nil {
			return nil, err
		}
		requests = append(requests, accounting.InvoiceLineRequest{
			ItemID:      line.ItemID,
			Description: helper.Deref(line.Description, ""),
			Qty:         qty,
			UnitID:      line.UnitID,
			UnitPrice:   line.UnitPrice,
			DiscountPct: line.DiscountPct,
			TaxIDs:      line.TaxIDs,
			AccountID:   accountID,
			DimensionID: line.DimensionID,
			SaleLineID:  helper.Ptr(line.ID),
		})
		toInvoice[line.ID] = qty
	}

	invoice, err := s.invoices.Create(ctx, accounting.CreateInvoiceRequest{
		OrganizationID: *order.OrganizationID,
		JournalID:      journalID,
		ContactID:      order.ContactID,
		Date:           date,
		Reference:      helper.Deref(order.Name, ""),
		Lines:          requests,
	})
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		if qty, ok := toInvoice[line.ID]; ok {
			line.QtyInvoiced += qty
			if _, err := s.lines.Update(ctx, line); err != nil {
				return nil, err
			}
		}
	}
	recomputeStatuses(order, lines)
	if _, err := s.orders.Update(ctx, order); err != nil {
		return nil, err
	}
	return invoice, nil
}

func (s SaleOrderService) CollectPayment(ctx context.Context, orderID uint64, journalID uint64, amount float64, date time.Time) (*accounting.Payment, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.OrganizationID == nil {
		return nil, ErrOrderNotFound
	}

	open, err := s.openInvoices.ListOpenByContact(ctx, order.ContactID)
	if err != nil {
		return nil, err
	}
	if len(open) == 0 {
		return nil, ErrOrderNoInvoices
	}

	invoiceIDs := make([]uint64, 0, len(open))
	for _, invoice := range open {
		invoiceIDs = append(invoiceIDs, invoice.ID)
	}

	return s.payments.Create(ctx, accounting.CreatePaymentRequest{
		OrganizationID: *order.OrganizationID,
		ContactID:      order.ContactID,
		JournalID:      journalID,
		Amount:         amount,
		Date:           date,
		Reference:      helper.Deref(order.Name, ""),
		InvoiceIDs:     invoiceIDs,
	})
}
