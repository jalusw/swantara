package sales

import (
	"context"
	"log/slog"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/kernel/state"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type SaleOrderService struct {
	orders         SaleOrderDAO
	lines          SaleOrderLineDAO
	sequences      sequence.Service
	productSvc     products.ProductService
	price_books    products.PriceBookDAO
	taxes          dao.CRUD[reference.Tax]
	contacts       contacts.ContactDAO
	leads          crm.ProspectDAO
	stages         dao.CRUD[reference.PipelineStage]
	warehouses     inventory.WarehouseDAO
	locations      inventory.StockLocationDAO
	quants         inventory.StockBalanceDAO
	shipments      inventory.ShipmentDAO
	movements      inventory.StockMovementDAO
	reservations   inventory.StockHoldDAO
	reservationSvc inventory.HoldService
	ship           ShipEngine
	invoices       InvoiceEngine
	openInvoices   InvoiceLookup
	payments       PaymentEngine
	machine        state.Machine
}

func NewSaleOrderService(
	orders SaleOrderDAO,
	lines SaleOrderLineDAO,
	sequences sequence.Service,
	productSvc products.ProductService,
	price_books products.PriceBookDAO,
	taxes dao.CRUD[reference.Tax],
	contacts contacts.ContactDAO,
	leads crm.ProspectDAO,
	stages dao.CRUD[reference.PipelineStage],
	warehouses inventory.WarehouseDAO,
	locations inventory.StockLocationDAO,
	quants inventory.StockBalanceDAO,
	shipments inventory.ShipmentDAO,
	movements inventory.StockMovementDAO,
	reservations inventory.StockHoldDAO,
	reservationSvc inventory.HoldService,
	ship ShipEngine,
	invoices InvoiceEngine,
	openInvoices InvoiceLookup,
	payments PaymentEngine,
) SaleOrderService {
	return SaleOrderService{
		orders:         orders,
		lines:          lines,
		sequences:      sequences,
		productSvc:     productSvc,
		price_books:    price_books,
		taxes:          taxes,
		contacts:       contacts,
		leads:          leads,
		stages:         stages,
		warehouses:     warehouses,
		locations:      locations,
		quants:         quants,
		shipments:      shipments,
		movements:      movements,
		reservations:   reservations,
		reservationSvc: reservationSvc,
		ship:           ship,
		invoices:       invoices,
		openInvoices:   openInvoices,
		payments:       payments,
		machine: state.NewMachine(
			state.Transition{From: model.Status(OrderStateDraft), To: model.Status(OrderStateSent)},
			state.Transition{From: model.Status(OrderStateDraft), To: model.Status(OrderStateConfirmed)},
			state.Transition{From: model.Status(OrderStateDraft), To: model.Status(OrderStateCancelled)},
			state.Transition{From: model.Status(OrderStateSent), To: model.Status(OrderStateConfirmed)},
			state.Transition{From: model.Status(OrderStateSent), To: model.Status(OrderStateCancelled)},
			state.Transition{From: model.Status(OrderStateConfirmed), To: model.Status(OrderStateDone)},
			state.Transition{From: model.Status(OrderStateConfirmed), To: model.Status(OrderStateCancelled)},
		),
	}
}

func (s SaleOrderService) Create(ctx context.Context, order *SaleOrder, lines []*SaleOrderLine) (*SaleOrder, error) {
	if order.OrganizationID == nil {
		return nil, ErrOrderContact
	}
	if len(lines) == 0 {
		return nil, ErrOrderNoLines
	}
	if order.WarehouseID == nil {
		return nil, ErrOrderWarehouse
	}
	if order.PriceBookID == nil {
		return nil, ErrOrderPriceBook
	}
	if err := s.validateContact(ctx, order.OrganizationID, order.ContactID); err != nil {
		return nil, err
	}
	if err := s.validateWarehouse(ctx, order.OrganizationID, *order.WarehouseID); err != nil {
		return nil, err
	}
	if err := s.validatePriceBook(ctx, order.OrganizationID, *order.PriceBookID); err != nil {
		return nil, err
	}
	if order.ProspectID != nil {
		if err := s.validateWonLead(ctx, *order.OrganizationID, *order.ProspectID); err != nil {
			return nil, err
		}
	}

	date := time.Now()
	if order.OrderDate != nil {
		date = *order.OrderDate
	}

	untaxed := amount.Zero()
	tax := amount.Zero()
	for _, line := range lines {
		if line.ItemID == nil {
			return nil, ErrOrderVariant
		}
		if !amount.FromFloat64(line.QtyOrdered).GreaterThan(amount.Zero()) {
			return nil, ErrOrderQty
		}
		if line.DiscountPct < 0 || line.DiscountPct > 100 {
			return nil, ErrOrderDiscount
		}
		if err := s.validateLineProduct(ctx, order.OrganizationID, *line.ItemID); err != nil {
			return nil, err
		}
		resolved, err := s.productSvc.ResolvePrice(ctx, *order.PriceBookID, *line.ItemID, amount.FromFloat64(line.QtyOrdered), date)
		if err != nil {
			return nil, err
		}
		line.UnitPrice = resolved.Price.Float64()
		if err := s.resolveLineAmounts(ctx, order.OrganizationID, line); err != nil {
			return nil, err
		}
		untaxed = untaxed.Add(amount.FromFloat64(line.PriceSubtotal))
		tax = tax.Add(amount.FromFloat64(line.PriceTax))
	}

	name, err := s.sequences.Next(ctx, *order.OrganizationID, SequenceSaleOrderCode)
	if err != nil {
		return nil, err
	}
	order.Name = helper.Ptr(name)
	order.AmountUntaxed = untaxed.Round(4).Float64()
	order.AmountTax = tax.Round(4).Float64()
	order.AmountTotal = untaxed.Add(tax).Round(4).Float64()
	order.State = OrderStateDraft
	order.InvoiceStatus = InvoiceStatusNo
	order.DeliveryStatus = DeliveryStatusPending

	return s.orders.CreateWithLines(ctx, order, lines)
}

func (s SaleOrderService) UpdateDraft(ctx context.Context, orderID uint64, order *SaleOrder, lines []*SaleOrderLine) (*SaleOrder, error) {
	existing, err := s.orders.Find(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrOrderNotFound
	}
	if existing.State != OrderStateDraft {
		return nil, ErrOrderState
	}
	if len(lines) == 0 {
		return nil, ErrOrderNoLines
	}
	if order.WarehouseID == nil {
		return nil, ErrOrderWarehouse
	}
	if order.PriceBookID == nil {
		return nil, ErrOrderPriceBook
	}
	if err := s.validateContact(ctx, existing.OrganizationID, order.ContactID); err != nil {
		return nil, err
	}
	if err := s.validateWarehouse(ctx, existing.OrganizationID, *order.WarehouseID); err != nil {
		return nil, err
	}
	if err := s.validatePriceBook(ctx, existing.OrganizationID, *order.PriceBookID); err != nil {
		return nil, err
	}

	date := time.Now()
	if order.OrderDate != nil {
		date = *order.OrderDate
	}

	untaxed := amount.Zero()
	tax := amount.Zero()
	for _, line := range lines {
		if line.ItemID == nil {
			return nil, ErrOrderVariant
		}
		if !amount.FromFloat64(line.QtyOrdered).GreaterThan(amount.Zero()) {
			return nil, ErrOrderQty
		}
		if line.DiscountPct < 0 || line.DiscountPct > 100 {
			return nil, ErrOrderDiscount
		}
		if err := s.validateLineProduct(ctx, existing.OrganizationID, *line.ItemID); err != nil {
			return nil, err
		}
		resolved, err := s.productSvc.ResolvePrice(ctx, *order.PriceBookID, *line.ItemID, amount.FromFloat64(line.QtyOrdered), date)
		if err != nil {
			return nil, err
		}
		line.UnitPrice = resolved.Price.Float64()
		if err := s.resolveLineAmounts(ctx, existing.OrganizationID, line); err != nil {
			return nil, err
		}
		untaxed = untaxed.Add(amount.FromFloat64(line.PriceSubtotal))
		tax = tax.Add(amount.FromFloat64(line.PriceTax))
	}

	order.ID = existing.ID
	order.OrganizationID = existing.OrganizationID
	order.Name = existing.Name
	order.CurrencyCode = existing.CurrencyCode
	order.ProspectID = existing.ProspectID
	order.State = OrderStateDraft
	order.AmountUntaxed = untaxed.Round(4).Float64()
	order.AmountTax = tax.Round(4).Float64()
	order.AmountTotal = untaxed.Add(tax).Round(4).Float64()
	order.InvoiceStatus = InvoiceStatusNo
	order.DeliveryStatus = DeliveryStatusPending

	if err := s.lines.ReplaceLines(ctx, orderID, lines); err != nil {
		return nil, err
	}
	return s.orders.Update(ctx, order)
}

func (s SaleOrderService) Send(ctx context.Context, orderID uint64) (*SaleOrder, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(order.State), model.Status(OrderStateSent)); err != nil {
		return nil, ErrOrderState
	}
	order.State = OrderStateSent
	return s.orders.Update(ctx, order)
}

func (s SaleOrderService) Confirm(ctx context.Context, orderID uint64) (*SaleOrder, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(order.State), model.Status(OrderStateConfirmed)); err != nil {
		return nil, ErrOrderState
	}
	if order.WarehouseID == nil {
		return nil, ErrOrderWarehouse
	}

	lines, err := s.lines.ListByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrOrderNoLines
	}

	stockLocationID, err := s.warehouseStockLocation(ctx, order.OrganizationID, *order.WarehouseID)
	if err != nil {
		return nil, err
	}
	customerLocationID, err := s.customerLocation(ctx, order.OrganizationID)
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		if line.ItemID == nil {
			continue
		}
		quant, err := s.quants.FindByKey(ctx, *line.ItemID, stockLocationID, nil)
		if err != nil {
			return nil, err
		}
		available := amount.Zero()
		if quant != nil {
			available = amount.FromFloat64(quant.Quantity).Sub(amount.FromFloat64(quant.ReservedQty))
		}
		if amount.FromFloat64(line.QtyOrdered).GreaterThan(available) {
			return nil, ErrOrderStockUnavailable
		}
	}

	movements := make([]*inventory.StockMovement, 0, len(lines))
	for _, line := range lines {
		if line.ItemID == nil {
			continue
		}
		movements = append(movements, &inventory.StockMovement{
			OrganizationID: order.OrganizationID,
			ItemID:         *line.ItemID,
			Qty:            line.QtyOrdered,
			UnitID:         line.UnitID,
			SrcLocationID:  stockLocationID,
			DstLocationID:  customerLocationID,
			State:          inventory.MovementStateConfirmed,
			OriginType:     helper.Ptr(accounting.OriginTypeSaleOrder),
			OriginID:       helper.Ptr(order.ID),
			ScheduledDate:  order.ExpectedDate,
		})
	}

	shipment := &inventory.Shipment{
		OrganizationID: order.OrganizationID,
		Name:           order.Name,
		Type:           inventory.ShipmentTypeOutgoing,
		ContactID:      helper.Ptr(order.ContactID),
		SrcLocationID:  helper.Ptr(stockLocationID),
		DstLocationID:  helper.Ptr(customerLocationID),
		State:          inventory.ShipmentStateConfirmed,
		Origin:         order.Name,
		ScheduledDate:  order.ExpectedDate,
	}

	created, err := s.shipments.CreateWithMovements(ctx, shipment, movements)
	if err != nil {
		return nil, err
	}

	reserved := make([]uint64, 0, len(movements))
	for _, movement := range movements {
		if _, err := s.reservationSvc.Reserve(ctx, *order.OrganizationID, movement.ItemID, stockLocationID, nil, amount.FromFloat64(movement.Qty), &movement.ID); err != nil {
			s.rollbackReservations(ctx, reserved)
			s.cancelShipment(ctx, created.ID)
			return nil, err
		}
		reserved = append(reserved, movement.ID)
		movement.State = inventory.MovementStateAssigned
		if _, err := s.movements.Update(ctx, movement); err != nil {
			s.rollbackReservations(ctx, reserved)
			s.cancelShipment(ctx, created.ID)
			return nil, err
		}
	}

	created.State = inventory.ShipmentStateAssigned
	if _, err := s.shipments.Update(ctx, created); err != nil {
		s.rollbackReservations(ctx, reserved)
		s.cancelShipment(ctx, created.ID)
		return nil, err
	}

	recomputeStatuses(order, lines)
	order.State = OrderStateConfirmed
	return s.orders.Update(ctx, order)
}

func (s SaleOrderService) Cancel(ctx context.Context, orderID uint64) (*SaleOrder, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(order.State), model.Status(OrderStateCancelled)); err != nil {
		return nil, ErrOrderState
	}
	if order.Name == nil {
		return nil, ErrOrderNotFound
	}

	shipments, err := s.shipments.List(ctx, &query.Query{Filters: []query.Filter{{Field: "origin", Operator: query.Equal, Value: *order.Name}}})
	if err != nil {
		return nil, err
	}
	for _, shipment := range shipments.Items {
		s.cancelShipment(ctx, shipment.ID)
	}

	order.State = OrderStateCancelled
	return s.orders.Update(ctx, order)
}

func (s SaleOrderService) Done(ctx context.Context, orderID uint64) (*SaleOrder, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(order.State), model.Status(OrderStateDone)); err != nil {
		return nil, ErrOrderState
	}
	order.State = OrderStateDone
	return s.orders.Update(ctx, order)
}

func (s SaleOrderService) ListLines(ctx context.Context, orderID uint64) ([]*SaleOrderLine, error) {
	return s.lines.ListByOrder(ctx, orderID)
}

func (s SaleOrderService) List(ctx context.Context, q *query.Query) (*query.Page[SaleOrder], error) {
	return s.orders.List(ctx, q)
}

func (s SaleOrderService) Find(ctx context.Context, id uint64) (*SaleOrder, error) {
	return s.findOrder(ctx, id)
}

func (s SaleOrderService) Delete(ctx context.Context, id uint64) error {
	return s.orders.Delete(ctx, id)
}

func (s SaleOrderService) RecomputeStatuses(ctx context.Context, orderID uint64) (*SaleOrder, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	lines, err := s.lines.ListByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	recomputeStatuses(order, lines)
	return s.orders.Update(ctx, order)
}

func (s SaleOrderService) RecordReturn(ctx context.Context, orderID, itemID uint64, qty float64) error {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return err
	}
	lines, err := s.lines.ListByOrder(ctx, orderID)
	if err != nil {
		return err
	}
	found := false
	for _, line := range lines {
		if line.ItemID != nil && *line.ItemID == itemID {
			line.QtyReturns = amount.FromFloat64(line.QtyReturns).Add(amount.FromFloat64(qty)).Round(4).Float64()
			if _, err := s.lines.Update(ctx, line); err != nil {
				return err
			}
			found = true
			break
		}
	}
	if !found {
		return ErrLineNotFound
	}
	recomputeStatuses(order, lines)
	_, err = s.orders.Update(ctx, order)
	return err
}

func recomputeStatuses(order *SaleOrder, lines []*SaleOrderLine) {
	ordered, delivered, invoiced, returned := amount.Zero(), amount.Zero(), amount.Zero(), amount.Zero()
	for _, line := range lines {
		ordered = ordered.Add(amount.FromFloat64(line.QtyOrdered))
		delivered = delivered.Add(amount.FromFloat64(line.QtyDelivered))
		invoiced = invoiced.Add(amount.FromFloat64(line.QtyInvoiced))
		returned = returned.Add(amount.FromFloat64(line.QtyReturns))
	}

	netDelivered := delivered.Sub(returned)
	switch {
	case !netDelivered.IsPositive():
		order.DeliveryStatus = DeliveryStatusPending
	case netDelivered.Equal(ordered):
		order.DeliveryStatus = DeliveryStatusDone
	default:
		order.DeliveryStatus = DeliveryStatusPartial
	}

	netInvoiced := invoiced.Sub(returned)
	switch {
	case !netInvoiced.IsPositive():
		order.InvoiceStatus = InvoiceStatusNo
	case netInvoiced.Equal(ordered):
		order.InvoiceStatus = InvoiceStatusInvoiced
	default:
		order.InvoiceStatus = InvoiceStatusToInvoice
	}
}

func (s SaleOrderService) findOrder(ctx context.Context, orderID uint64) (*SaleOrder, error) {
	order, err := s.orders.Find(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s SaleOrderService) validateContact(ctx context.Context, organizationID *uint64, contactID uint64) error {
	contact, err := s.contacts.Find(ctx, contactID)
	if err != nil {
		return err
	}
	if contact == nil {
		return ErrOrderContact
	}
	if organizationID != nil && contact.OrganizationID != nil && *contact.OrganizationID != *organizationID {
		return ErrOrderContactOrganization
	}
	return nil
}

func (s SaleOrderService) validateWarehouse(ctx context.Context, organizationID *uint64, warehouseID uint64) error {
	warehouse, err := s.warehouses.Find(ctx, warehouseID)
	if err != nil {
		return err
	}
	if warehouse == nil {
		return ErrOrderWarehouse
	}
	if !helper.OwnedByOrg(warehouse.OrganizationID, organizationID) {
		return ErrOrderWarehouseOrganization
	}
	return nil
}

func (s SaleOrderService) validatePriceBook(ctx context.Context, organizationID *uint64, price_bookID uint64) error {
	price_book, err := s.price_books.Find(ctx, price_bookID)
	if err != nil {
		return err
	}
	if price_book == nil {
		return ErrOrderPriceBook
	}
	if organizationID != nil && price_book.OrganizationID != nil && *price_book.OrganizationID != *organizationID {
		return ErrOrderPriceBookOrganization
	}
	return nil
}

func (s SaleOrderService) validateLineProduct(ctx context.Context, organizationID *uint64, variantID uint64) error {
	variantOrg, err := s.productSvc.ResolveVariantOrganization(ctx, variantID)
	if err != nil {
		return err
	}
	if organizationID != nil && variantOrg != nil && *variantOrg != *organizationID {
		return ErrOrderVariantOrganization
	}
	return nil
}

func (s SaleOrderService) validateWonLead(ctx context.Context, organizationID, prospectID uint64) error {
	lead, err := s.leads.Find(ctx, prospectID)
	if err != nil {
		return err
	}
	if lead == nil {
		return ErrOrderLead
	}
	if lead.OrganizationID != nil && *lead.OrganizationID != organizationID {
		return ErrOrderLeadOrganization
	}
	if lead.Type != crm.ProspectKindOpportunity || lead.StageID == nil {
		return ErrOrderLeadNotWon
	}
	stage, err := s.stages.Find(ctx, *lead.StageID)
	if err != nil {
		return err
	}
	if stage == nil || !stage.IsWon {
		return ErrOrderLeadNotWon
	}
	if lead.OrganizationID != nil && stage.OrganizationID != nil && *stage.OrganizationID != *lead.OrganizationID {
		return ErrOrderLeadNotWon
	}
	return nil
}

func (s SaleOrderService) resolveLineAmounts(ctx context.Context, organizationID *uint64, line *SaleOrderLine) error {
	qty := amount.FromFloat64(line.QtyOrdered)
	discount := amount.FromFloat64(1 - line.DiscountPct/100)
	subtotal := qty.Mul(amount.FromFloat64(line.UnitPrice)).Mul(discount).Round(4)
	lineTax := amount.Zero()

	for _, taxID := range line.TaxIDs {
		tax, err := s.taxes.Find(ctx, uint64(taxID))
		if err != nil {
			return err
		}
		if tax == nil {
			return ErrOrderTax
		}
		if organizationID != nil && tax.OrganizationID != nil && *tax.OrganizationID != *organizationID {
			return ErrOrderTaxOrganization
		}
		if tax.Scope != reference.TaxScopeSale && tax.Scope != reference.TaxScopeNone {
			return ErrOrderTaxInvalid
		}
		taxAmount, err := reference.TaxLineAmount(subtotal, qty, tax)
		if err != nil {
			return err
		}
		lineTax = lineTax.Add(taxAmount)
	}

	line.PriceSubtotal = subtotal.Float64()
	line.PriceTax = lineTax.Round(4).Float64()
	line.PriceTotal = subtotal.Add(lineTax).Round(4).Float64()
	return nil
}

func (s SaleOrderService) warehouseStockLocation(ctx context.Context, organizationID *uint64, warehouseID uint64) (uint64, error) {
	locations, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "warehouse_id", Operator: query.Equal, Value: warehouseID}}})
	if err != nil {
		return 0, err
	}
	for _, location := range locations.Items {
		if location.Usage == "internal" && organizationID != nil && location.OrganizationID != nil && *location.OrganizationID == *organizationID {
			return location.ID, nil
		}
	}
	return 0, ErrOrderLocation
}

func (s SaleOrderService) customerLocation(ctx context.Context, organizationID *uint64) (uint64, error) {
	locations, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "usage", Operator: query.Equal, Value: "customer"}}})
	if err != nil {
		return 0, err
	}
	for _, location := range locations.Items {
		if organizationID != nil && location.OrganizationID != nil && *location.OrganizationID == *organizationID {
			return location.ID, nil
		}
	}
	return 0, ErrOrderCustomerLocation
}

func (s SaleOrderService) rollbackReservations(ctx context.Context, moveIDs []uint64) {
	for _, moveID := range moveIDs {
		if err := s.reservations.ReleaseByMovement(ctx, moveID); err != nil {
			slog.Error("sale order reservation rollback failed", "movement_id", moveID, "error", err)
		}
	}
}

func (s SaleOrderService) cancelShipment(ctx context.Context, shipmentID uint64) {
	movements, err := s.movements.ListByShipment(ctx, shipmentID)
	if err != nil {
		slog.Error("sale order shipment cancel failed to list movements", "shipment_id", shipmentID, "error", err)
		return
	}
	for _, movement := range movements {
		if err := s.reservations.ReleaseByMovement(ctx, movement.ID); err != nil {
			slog.Error("sale order shipment cancel failed to release reservation", "movement_id", movement.ID, "error", err)
		}
		movement.State = inventory.MovementStateCancelled
		if _, err := s.movements.Update(ctx, movement); err != nil {
			slog.Error("sale order shipment cancel failed to update movement", "movement_id", movement.ID, "error", err)
		}
	}
	shipment, err := s.shipments.Find(ctx, shipmentID)
	if err != nil {
		slog.Error("sale order shipment cancel failed to find shipment", "shipment_id", shipmentID, "error", err)
		return
	}
	if shipment == nil {
		return
	}
	shipment.State = inventory.ShipmentStateCancelled
	if _, err := s.shipments.Update(ctx, shipment); err != nil {
		slog.Error("sale order shipment cancel failed to update shipment", "shipment_id", shipmentID, "error", err)
	}
}
