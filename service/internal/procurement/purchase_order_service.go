package procurement

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/kernel/state"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type ExpenseAccountEngine interface {
	ResolveExpenseAccount(ctx context.Context, variantID uint64) (uint64, error)
}

type PurchaseOrderService struct {
	orders         PurchaseOrderDAO
	lines          PurchaseOrderLineDAO
	requisitions   PurchaseRequestDAO
	linesDAO       PurchaseRequestLineDAO
	agreements     SupplyAgreementDAO
	agreementLines SupplyAgreementLineDAO
	creditMemos    PurchaseCreditMemoDAO
	debitMemos     PurchaseDebitMemoDAO
	batches        PaymentBatchDAO
	batchLines     PaymentBatchLineDAO
	sequences      sequence.Service
	configs        PurchaseConfigSource
	offers         OfferEngine
	contacts       contacts.ContactDAO
	suppliers      SupplierProfileLookup
	warehouses     inventory.WarehouseDAO
	locations      inventory.StockLocationDAO
	shipments      inventory.ShipmentDAO
	movements      inventory.StockMovementDAO
	resolver       inventory.ItemResolver
	expense        ExpenseAccountEngine
	taxes          dao.CRUD[reference.Tax]
	receive        ReceiveEngine
	approvals      ApprovalEngine
	bills          BillEngine
	openInvoices   OpenInvoiceLookup
	payments       OutboundPaymentEngine
	quality        QualityEngine
	converter      CurrencyConverter
	machine        state.Machine
}

func NewPurchaseOrderService(
	orders PurchaseOrderDAO,
	lines PurchaseOrderLineDAO,
	requisitions PurchaseRequestDAO,
	requisitionLines PurchaseRequestLineDAO,
	agreements SupplyAgreementDAO,
	agreementLines SupplyAgreementLineDAO,
	creditMemos PurchaseCreditMemoDAO,
	debitMemos PurchaseDebitMemoDAO,
	batches PaymentBatchDAO,
	batchLines PaymentBatchLineDAO,
	sequences sequence.Service,
	configs PurchaseConfigSource,
	offers OfferEngine,
	contacts contacts.ContactDAO,
	suppliers SupplierProfileLookup,
	warehouses inventory.WarehouseDAO,
	locations inventory.StockLocationDAO,
	shipments inventory.ShipmentDAO,
	movements inventory.StockMovementDAO,
	resolver inventory.ItemResolver,
	expense ExpenseAccountEngine,
	taxes dao.CRUD[reference.Tax],
	receive ReceiveEngine,
	approvals ApprovalEngine,
	bills BillEngine,
	openInvoices OpenInvoiceLookup,
	payments OutboundPaymentEngine,
	quality QualityEngine,
	converter CurrencyConverter,
) PurchaseOrderService {
	return PurchaseOrderService{
		orders:         orders,
		lines:          lines,
		requisitions:   requisitions,
		linesDAO:       requisitionLines,
		agreements:     agreements,
		agreementLines: agreementLines,
		creditMemos:    creditMemos,
		debitMemos:     debitMemos,
		batches:        batches,
		batchLines:     batchLines,
		sequences:      sequences,
		configs:        configs,
		offers:         offers,
		contacts:       contacts,
		suppliers:      suppliers,
		warehouses:     warehouses,
		locations:      locations,
		shipments:      shipments,
		movements:      movements,
		resolver:       resolver,
		expense:        expense,
		taxes:          taxes,
		receive:        receive,
		approvals:      approvals,
		bills:          bills,
		openInvoices:   openInvoices,
		payments:       payments,
		quality:        quality,
		converter:      converter,
		machine: state.NewMachine(
			state.Transition{From: model.Status(PurchaseOrderStateDraft), To: model.Status(PurchaseOrderStateSent)},
			state.Transition{From: model.Status(PurchaseOrderStateDraft), To: model.Status(PurchaseOrderStateCancelled)},
			state.Transition{From: model.Status(PurchaseOrderStateSent), To: model.Status(PurchaseOrderStateConfirmed)},
			state.Transition{From: model.Status(PurchaseOrderStateSent), To: model.Status(PurchaseOrderStateCancelled)},
			state.Transition{From: model.Status(PurchaseOrderStateConfirmed), To: model.Status(PurchaseOrderStateDone)},
			state.Transition{From: model.Status(PurchaseOrderStateConfirmed), To: model.Status(PurchaseOrderStateCancelled)},
		),
	}
}

func (s PurchaseOrderService) List(ctx context.Context, q *query.Query) (*query.Page[PurchaseOrder], error) {
	return s.orders.List(ctx, q)
}

func (s PurchaseOrderService) Find(ctx context.Context, id uint64) (*PurchaseOrder, error) {
	return s.orders.Find(ctx, id)
}

func (s PurchaseOrderService) ListLines(ctx context.Context, orderID uint64) ([]*PurchaseOrderLine, error) {
	return s.lines.ListByOrder(ctx, orderID)
}

func (s PurchaseOrderService) Create(ctx context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error) {
	if order.OrganizationID == nil {
		return nil, ErrPurchaseOrderNotFound
	}
	if len(lines) == 0 {
		return nil, ErrPurchaseOrderNoLines
	}
	if err := s.validateVendor(ctx, order); err != nil {
		return nil, err
	}
	if err := s.validateWarehouse(ctx, order.OrganizationID, order.WarehouseID); err != nil {
		return nil, err
	}

	orderDate := order.OrderDate
	if orderDate == nil {
		now := time.Now()
		orderDate = &now
	}
	if err := s.priceLines(ctx, order, orderDate, lines); err != nil {
		return nil, err
	}

	name, err := s.sequences.Next(ctx, *order.OrganizationID, SequencePurchaseOrderCode)
	if err != nil {
		return nil, err
	}
	order.Name = helper.Ptr(name)
	order.State = PurchaseOrderStateDraft
	order.InvoiceStatus = PurchaseOrderInvoiceStatusNo
	order.ReceiptStatus = PurchaseOrderReceiptStatusPending
	if order.CurrencyCode == nil {
		order.CurrencyCode = helper.Ptr("IDR")
	}
	s.recomputeTotals(ctx, order, lines)
	return s.orders.CreateWithLines(ctx, order, lines)
}

func (s PurchaseOrderService) CreateFromRequest(ctx context.Context, requestID uint64, order *PurchaseOrder) (*PurchaseOrder, error) {
	requisition, err := s.requisitions.Find(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if requisition == nil {
		return nil, ErrRequisitionNotFound
	}
	if requisition.State != RequestStateApproved {
		return nil, ErrRequisitionNotConvertible
	}
	requisitionLines, err := s.requisitionLines(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if len(requisitionLines) == 0 {
		return nil, ErrRequisitionNoLines
	}

	lines := make([]*PurchaseOrderLine, 0, len(requisitionLines))
	for _, requisitionLine := range requisitionLines {
		if requisitionLine.ItemID == nil {
			continue
		}
		lines = append(lines, &PurchaseOrderLine{
			ItemID:        requisitionLine.ItemID,
			Description:   requisitionLine.Description,
			QtyOrdered:    requisitionLine.Qty,
			UnitID:        requisitionLine.UnitID,
			QtyReceived:   0,
			QtyBilled:     0,
			DiscountPct:   0,
			PriceSubtotal: 0,
		})
	}
	if len(lines) == 0 {
		return nil, ErrPurchaseOrderNoLines
	}
	if order.OrganizationID == nil {
		order.OrganizationID = requisition.OrganizationID
	}
	return s.Create(ctx, order, lines)
}

func (s PurchaseOrderService) UpdateDraft(ctx context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error) {
	current, err := s.findOrder(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	if current.State != PurchaseOrderStateDraft {
		return nil, ErrPurchaseOrderState
	}
	if len(lines) == 0 {
		return nil, ErrPurchaseOrderNoLines
	}
	if err := s.validateVendor(ctx, order); err != nil {
		return nil, err
	}
	if err := s.validateWarehouse(ctx, current.OrganizationID, order.WarehouseID); err != nil {
		return nil, err
	}

	orderDate := order.OrderDate
	if orderDate == nil {
		now := time.Now()
		orderDate = &now
	}
	if err := s.priceLines(ctx, order, orderDate, lines); err != nil {
		return nil, err
	}

	order.Name = current.Name
	order.OrganizationID = current.OrganizationID
	order.State = PurchaseOrderStateDraft
	order.InvoiceStatus = current.InvoiceStatus
	order.ReceiptStatus = current.ReceiptStatus
	if order.CurrencyCode == nil {
		order.CurrencyCode = current.CurrencyCode
	}
	s.recomputeTotals(ctx, order, lines)
	if err := s.lines.ReplaceLines(ctx, order.ID, lines); err != nil {
		return nil, err
	}
	return s.orders.Update(ctx, order)
}

func (s PurchaseOrderService) Confirm(ctx context.Context, orderID uint64, byUserID uint64) (*PurchaseOrder, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(order.State), model.Status(PurchaseOrderStateSent)); err != nil {
		return nil, ErrPurchaseOrderState
	}
	if threshold := s.threshold(ctx, order); threshold > 0 && order.AmountTotal >= threshold {
		if !s.approved(ctx, order) {
			if err := s.requestApproval(ctx, order, byUserID); err != nil {
				return nil, err
			}
			return nil, ErrPurchaseOrderApprovalPending
		}
	}
	order.State = PurchaseOrderStateSent
	return s.orders.Update(ctx, order)
}

func (s PurchaseOrderService) Receive(ctx context.Context, orderID, journalID uint64, date time.Time) (*PurchaseOrder, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.State != PurchaseOrderStateSent && order.State != PurchaseOrderStateConfirmed {
		return nil, ErrPurchaseOrderState
	}
	if threshold := s.threshold(ctx, order); threshold > 0 && order.AmountTotal >= threshold && !s.approved(ctx, order) {
		return nil, ErrPurchaseOrderApprovalPending
	}
	if order.OrganizationID == nil {
		return nil, ErrPurchaseOrderNotFound
	}

	lines, err := s.lines.ListByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrPurchaseOrderNoLines
	}
	if order.State == PurchaseOrderStateSent {
		if err := s.machine.TryTransition(model.Status(order.State), model.Status(PurchaseOrderStateConfirmed)); err != nil {
			return nil, ErrPurchaseOrderState
		}
		order.State = PurchaseOrderStateConfirmed
		if _, err := s.orders.Update(ctx, order); err != nil {
			return nil, err
		}
	}

	shipment, err := s.findOrCreateShipment(ctx, order)
	if err != nil {
		return nil, err
	}

	received := false
	for _, line := range lines {
		if line.ItemID == nil {
			continue
		}
		remaining := amount.FromFloat64(line.QtyOrdered).Sub(amount.FromFloat64(line.QtyReceived))
		if !remaining.GreaterThan(amount.Zero()) {
			continue
		}
		movement, err := s.findOrCreateMovement(ctx, shipment, line)
		if err != nil {
			return nil, err
		}
		if movement.State == inventory.MovementStateDone || movement.State == inventory.MovementStateCancelled {
			continue
		}
		if _, err := s.receive.Receive(ctx, movement.ID, amount.FromFloat64(line.UnitPrice), journalID, date); err != nil {
			return nil, err
		}
		line.QtyReceived = amount.FromFloat64(line.QtyReceived).Add(remaining).Round(4).Float64()
		if _, err := s.lines.Update(ctx, line); err != nil {
			return nil, err
		}
		received = true
	}
	if !received {
		return nil, ErrPurchaseOrderNothingToReceive
	}

	itemIDs := make([]uint64, 0, len(lines))
	for _, line := range lines {
		if line.ItemID != nil {
			itemIDs = append(itemIDs, *line.ItemID)
		}
	}
	if len(itemIDs) > 0 {
		if _, err := s.quality.TriggerChecks(ctx, *order.OrganizationID, shipment.ID, itemIDs); err != nil {
			return nil, err
		}
	}

	s.recomputeStatuses(order, lines)
	if _, err := s.orders.Update(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s PurchaseOrderService) CreateSupplierBill(ctx context.Context, orderID, journalID uint64, date time.Time, override bool) (*accounting.Invoice, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.State != PurchaseOrderStateConfirmed && order.State != PurchaseOrderStateDone {
		return nil, ErrPurchaseOrderState
	}
	if order.OrganizationID == nil {
		return nil, ErrPurchaseOrderNotFound
	}

	lines, err := s.lines.ListByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrPurchaseOrderNoLines
	}

	shipment, err := s.shipmentFor(ctx, order)
	if err != nil {
		return nil, err
	}
	if shipment != nil && !override {
		failed, err := s.quality.HasFailedChecks(ctx, shipment.ID)
		if err != nil {
			return nil, err
		}
		if failed {
			return nil, ErrPurchaseOrderBillQualityBlocked
		}
	}

	requests := make([]accounting.InvoiceLineRequest, 0, len(lines))
	toBill := map[uint64]float64{}
	for _, line := range lines {
		qty := amount.FromFloat64(line.QtyReceived).Sub(amount.FromFloat64(line.QtyBilled))
		if !qty.GreaterThan(amount.Zero()) {
			continue
		}
		if line.ItemID == nil {
			continue
		}
		resolved, err := s.resolver.Resolve(ctx, *line.ItemID)
		if err != nil {
			return nil, err
		}
		accountID := resolved.StockInputAccountID
		if accountID == 0 {
			accountID, err = s.expense.ResolveExpenseAccount(ctx, *line.ItemID)
			if err != nil {
				return nil, err
			}
		}
		requests = append(requests, accounting.InvoiceLineRequest{
			ItemID:         line.ItemID,
			Description:    helper.Deref(line.Description, ""),
			Qty:            qty.Float64(),
			UnitID:         line.UnitID,
			UnitPrice:      line.UnitPrice,
			DiscountPct:    line.DiscountPct,
			TaxIDs:         line.TaxIDs,
			AccountID:      accountID,
			DimensionID:    line.DimensionID,
			PurchaseLineID: helper.Ptr(line.ID),
		})
		toBill[line.ID] = qty.Float64()
	}
	if len(requests) == 0 {
		return nil, ErrPurchaseOrderNothingToBill
	}

	invoice, err := s.bills.CreateSupplierBill(ctx, accounting.CreateSupplierBillRequest{
		OrganizationID: *order.OrganizationID,
		JournalID:      journalID,
		ContactID:      order.SupplierID,
		Date:           date,
		Reference:      helper.Deref(order.Name, ""),
		Lines:          requests,
	})
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		if qty, ok := toBill[line.ID]; ok {
			line.QtyBilled = amount.FromFloat64(line.QtyBilled).Add(amount.FromFloat64(qty)).Round(4).Float64()
			if _, err := s.lines.Update(ctx, line); err != nil {
				return nil, err
			}
		}
	}
	s.recomputeStatuses(order, lines)
	if _, err := s.orders.Update(ctx, order); err != nil {
		return nil, err
	}
	return invoice, nil
}

func (s PurchaseOrderService) PaySupplierBill(ctx context.Context, orderID, journalID uint64, date time.Time) (*accounting.Payment, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.OrganizationID == nil {
		return nil, ErrPurchaseOrderNotFound
	}
	open, err := s.openInvoices.ListOpenByContact(ctx, order.SupplierID)
	if err != nil {
		return nil, err
	}
	if len(open) == 0 {
		return nil, ErrPurchaseOrderNoOpenBills
	}
	invoiceIDs := make([]uint64, 0, len(open))
	for _, invoice := range open {
		invoiceIDs = append(invoiceIDs, invoice.ID)
	}
	return s.payments.CreateOutbound(ctx, accounting.CreatePaymentRequest{
		OrganizationID: *order.OrganizationID,
		ContactID:      order.SupplierID,
		JournalID:      journalID,
		Amount:         order.AmountTotal,
		Date:           date,
		Reference:      helper.Deref(order.Name, ""),
		InvoiceIDs:     invoiceIDs,
	})
}

func (s PurchaseOrderService) Cancel(ctx context.Context, orderID uint64) (*PurchaseOrder, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(order.State), model.Status(PurchaseOrderStateCancelled)); err != nil {
		return nil, ErrPurchaseOrderState
	}
	order.State = PurchaseOrderStateCancelled
	return s.orders.Update(ctx, order)
}

func (s PurchaseOrderService) CreateVendorCreditMemo(ctx context.Context, orderID uint64, journalID uint64, date time.Time, amount float64, reason string) (*PurchaseCreditMemo, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.OrganizationID == nil {
		return nil, ErrPurchaseOrderNotFound
	}
	if order.State != PurchaseOrderStateConfirmed && order.State != PurchaseOrderStateDone {
		return nil, ErrPurchaseOrderState
	}

	invoice, err := s.bills.CreateCreditNote(ctx, accounting.CreateCreditNoteRequest{
		OrganizationID:    *order.OrganizationID,
		OriginalInvoiceID: 0,
		JournalID:         journalID,
		Date:              date,
		Reference:         helper.Deref(order.Name, ""),
	})
	if err != nil {
		return nil, err
	}

	memo := &PurchaseCreditMemo{
		OrderID:       orderID,
		JournalID:     journalID,
		ContactID:     order.SupplierID,
		Date:          &date,
		State:         CreditMemoStatePosted,
		AmountUntaxed: amount,
		AmountTotal:   amount,
		Reason:        &reason,
	}
	created, err := s.creditMemos.Create(ctx, memo)
	if err != nil {
		return nil, err
	}

	order.AmountTotal = amountFromFloat64(order.AmountTotal).Sub(amountFromFloat64(amount)).Float64()
	order.AmountUntaxed = amountFromFloat64(order.AmountUntaxed).Sub(amountFromFloat64(amount)).Float64()
	if _, err := s.orders.Update(ctx, order); err != nil {
		return nil, err
	}

	_ = invoice
	return created, nil
}

func (s PurchaseOrderService) CreateVendorDebitMemo(ctx context.Context, orderID uint64, journalID uint64, date time.Time, amount float64, reason string) (*PurchaseDebitMemo, error) {
	order, err := s.findOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.OrganizationID == nil {
		return nil, ErrPurchaseOrderNotFound
	}
	if order.State != PurchaseOrderStateConfirmed && order.State != PurchaseOrderStateDone {
		return nil, ErrPurchaseOrderState
	}

	invoice, err := s.bills.CreateSupplierBill(ctx, accounting.CreateSupplierBillRequest{
		OrganizationID: *order.OrganizationID,
		JournalID:      journalID,
		ContactID:      order.SupplierID,
		Date:           date,
		Reference:      helper.Deref(order.Name, ""),
		Lines: []accounting.InvoiceLineRequest{
			{AccountID: 1310, Qty: 1, UnitPrice: amount, Description: reason},
		},
	})
	if err != nil {
		return nil, err
	}

	memo := &PurchaseDebitMemo{
		OrderID:       orderID,
		JournalID:     journalID,
		ContactID:     order.SupplierID,
		Date:          &date,
		State:         DebitMemoStatePosted,
		AmountUntaxed: amount,
		AmountTotal:   amount,
		Reason:        &reason,
	}
	created, err := s.debitMemos.Create(ctx, memo)
	if err != nil {
		return nil, err
	}

	order.AmountTotal = amountFromFloat64(order.AmountTotal).Add(amountFromFloat64(amount)).Float64()
	order.AmountUntaxed = amountFromFloat64(order.AmountUntaxed).Add(amountFromFloat64(amount)).Float64()
	if _, err := s.orders.Update(ctx, order); err != nil {
		return nil, err
	}

	_ = invoice
	return created, nil
}

type PaymentBatchRequest struct {
	JournalID uint64                    `json:"journal_id"`
	Date      time.Time                 `json:"date"`
	Orders    []PaymentBatchRequestLine `json:"orders"`
}

type PaymentBatchRequestLine struct {
	OrderID uint64  `json:"order_id"`
	Amount  float64 `json:"amount"`
}

func (s PurchaseOrderService) CreatePaymentBatch(ctx context.Context, organizationID uint64, supplierID uint64, request PaymentBatchRequest) (*PaymentBatch, []*PaymentBatchLine, error) {
	if len(request.Orders) == 0 {
		return nil, nil, ErrPaymentBatchNoOrders
	}

	var totalAmount float64
	batchLines := make([]*PaymentBatchLine, 0, len(request.Orders))

	for _, req := range request.Orders {
		order, err := s.findOrder(ctx, req.OrderID)
		if err != nil {
			return nil, nil, err
		}
		if order.SupplierID != supplierID {
			return nil, nil, ErrPaymentBatchMixedVendors
		}
		if order.State != PurchaseOrderStateConfirmed && order.State != PurchaseOrderStateDone {
			return nil, nil, ErrPaymentBatchInvalidOrderState
		}
		if req.Amount <= 0 {
			return nil, nil, ErrPaymentBatchInvalidAmount
		}
		totalAmount += req.Amount
		batchLines = append(batchLines, &PaymentBatchLine{
			OrderID: req.OrderID,
			Amount:  req.Amount,
		})
	}

	batch := &PaymentBatch{
		OrganizationID: &organizationID,
		JournalID:      request.JournalID,
		ContactID:      supplierID,
		Date:           &request.Date,
		State:          PaymentBatchStateDraft,
		TotalAmount:    totalAmount,
		PaymentCount:   len(request.Orders),
	}

	created, err := s.batches.CreateWithLines(ctx, batch, batchLines)
	if err != nil {
		return nil, nil, err
	}

	return created, batchLines, nil
}

func (s PurchaseOrderService) ConfirmBatch(ctx context.Context, batchID uint64) (*PaymentBatch, error) {
	batch, err := s.batches.Find(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, ErrPaymentBatchNotFound
	}
	if batch.State != PaymentBatchStateDraft {
		return nil, ErrPaymentBatchInvalidState
	}

	batchLines, err := s.batchLines.ListByBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}

	for _, line := range batchLines {
		order, err := s.findOrder(ctx, line.OrderID)
		if err != nil {
			return nil, err
		}
		order.InvoiceStatus = PurchaseOrderInvoiceStatusInvoiced
		if _, err := s.orders.Update(ctx, order); err != nil {
			return nil, err
		}
	}

	batch.State = PaymentBatchStatePosted
	return s.batches.Update(ctx, batch)
}

func (s PurchaseOrderService) GetPaymentBatch(ctx context.Context, batchID uint64) (*PaymentBatch, []*PaymentBatchLine, error) {
	batch, err := s.batches.Find(ctx, batchID)
	if err != nil {
		return nil, nil, err
	}
	if batch == nil {
		return nil, nil, ErrPaymentBatchNotFound
	}
	batchLines, err := s.batchLines.ListByBatch(ctx, batchID)
	if err != nil {
		return nil, nil, err
	}
	return batch, batchLines, nil
}

func (s PurchaseOrderService) ListPaymentBatches(ctx context.Context, q *query.Query) (*query.Page[PaymentBatch], error) {
	return s.batches.List(ctx, q)
}

func (s PurchaseOrderService) CreateFromAgreement(ctx context.Context, agreementID uint64, overrides *CreateFromAgreementOverrides) (*PurchaseOrder, error) {
	agreement, err := s.agreements.Find(ctx, agreementID)
	if err != nil {
		return nil, err
	}
	if agreement == nil {
		return nil, ErrAgreementNotFound
	}
	if agreement.State != SupplyAgreementStateActive {
		return nil, ErrAgreementNotActive
	}
	if agreement.EndDate != nil && agreement.EndDate.Before(time.Now()) {
		return nil, ErrAgreementExpired
	}

	agreementLines, err := s.agreementLines.ListByAgreement(ctx, agreementID)
	if err != nil {
		return nil, err
	}
	if len(agreementLines) == 0 {
		return nil, ErrAgreementNoLines
	}

	supplierID := agreement.SupplierID
	var warehouseID *uint64
	if overrides != nil {
		if overrides.WarehouseID != nil {
			warehouseID = overrides.WarehouseID
		}
		if overrides.SupplierID != nil {
			supplierID = *overrides.SupplierID
		}
	}

	order := &PurchaseOrder{
		OrganizationID: agreement.OrganizationID,
		SupplierID:     supplierID,
		WarehouseID:    warehouseID,
		State:          PurchaseOrderStateDraft,
		ReceiptStatus:  PurchaseOrderReceiptStatusPending,
		InvoiceStatus:  PurchaseOrderInvoiceStatusNo,
	}

	lines := make([]*PurchaseOrderLine, 0, len(agreementLines))
	var totalAmount float64
	for _, al := range agreementLines {
		qty := al.Qty
		if overrides != nil && overrides.QtyOverrides != nil {
			if q, ok := overrides.QtyOverrides[al.ID]; ok {
				qty = q
			}
		}
		if qty <= 0 {
			continue
		}
		lineAmount := qty * al.UnitPrice
		totalAmount += lineAmount
		lines = append(lines, &PurchaseOrderLine{
			ItemID:      al.ItemID,
			Description: al.Description,
			QtyOrdered:  qty,
			UnitPrice:   al.UnitPrice,
			UnitID:      al.UnitID,
		})
	}

	if len(lines) == 0 {
		return nil, ErrPurchaseOrderNothingToBill
	}

	if agreement.QtyLimit > 0 {
		newQty := amount.FromFloat64(agreement.ConsumedQty).Add(amount.FromFloat64(float64(len(agreementLines))))
		if newQty.GreaterThan(amount.FromFloat64(agreement.QtyLimit)) {
			return nil, ErrAgreementQtyExceeded
		}
	}
	if agreement.AmountLimit > 0 {
		newAmount := amount.FromFloat64(agreement.ConsumedAmount).Add(amount.FromFloat64(totalAmount))
		if newAmount.GreaterThan(amount.FromFloat64(agreement.AmountLimit)) {
			return nil, ErrAgreementAmountExceeded
		}
	}

	name, err := s.sequences.Next(ctx, *order.OrganizationID, SequencePurchaseOrderCode)
	if err != nil {
		return nil, err
	}
	order.Name = helper.Ptr(name)
	if agreement.CurrencyCode != nil {
		order.CurrencyCode = agreement.CurrencyCode
	}
	if order.CurrencyCode == nil {
		order.CurrencyCode = helper.Ptr("IDR")
	}

	order, err = s.orders.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		line.OrderID = order.ID
		if _, err := s.lines.Create(ctx, line); err != nil {
			return nil, err
		}
	}

	agreement.ConsumedQty = amount.FromFloat64(agreement.ConsumedQty).Add(amount.FromFloat64(float64(len(agreementLines)))).Float64()
	agreement.ConsumedAmount = amount.FromFloat64(agreement.ConsumedAmount).Add(amount.FromFloat64(totalAmount)).Float64()
	if _, err := s.agreements.Update(ctx, agreement); err != nil {
		return nil, err
	}

	return order, nil
}

type CreateFromAgreementOverrides struct {
	SupplierID   *uint64
	WarehouseID  *uint64
	QtyOverrides map[uint64]float64
}

func (s PurchaseOrderService) priceLines(ctx context.Context, order *PurchaseOrder, date *time.Time, lines []*PurchaseOrderLine) error {
	for _, line := range lines {
		if !amount.FromFloat64(line.QtyOrdered).GreaterThan(amount.Zero()) {
			return ErrPurchaseOrderLineQty
		}
		if line.DiscountPct < 0 || line.DiscountPct > 100 {
			return ErrPurchaseOrderLineDiscount
		}
		if line.ItemID == nil {
			if !amount.FromFloat64(line.UnitPrice).GreaterThan(amount.Zero()) {
				return ErrPurchaseOrderNoOffer
			}
			continue
		}
		if _, err := s.resolver.Resolve(ctx, *line.ItemID); err != nil {
			return err
		}
		if !amount.FromFloat64(line.UnitPrice).GreaterThan(amount.Zero()) {
			offer, err := s.offers.BestOfferForSupplier(ctx, *line.ItemID, order.SupplierID, amount.FromFloat64(line.QtyOrdered), *date)
			if err != nil {
				return ErrPurchaseOrderNoOffer
			}
			if offer == nil || offer.Price == nil {
				return ErrPurchaseOrderNoOffer
			}
			line.UnitPrice = *offer.Price
		}
	}
	return nil
}

func (s PurchaseOrderService) recomputeTotals(ctx context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) {
	untaxed := amount.Zero()
	tax := amount.Zero()

	for _, line := range lines {
		qty := amount.FromFloat64(line.QtyOrdered)
		discount := amount.FromFloat64(1 - line.DiscountPct/100)
		subtotal := qty.Mul(amount.FromFloat64(line.UnitPrice)).Mul(discount).Round(4)
		line.PriceSubtotal = subtotal.Float64()
		untaxed = untaxed.Add(subtotal)
		for _, taxID := range line.TaxIDs {
			taxEntity, err := s.taxes.Find(ctx, uint64(taxID))
			if err != nil {
				continue
			}
			if taxEntity == nil || taxEntity.Amount == nil {
				continue
			}
			taxAmount, err := reference.TaxLineAmount(subtotal, qty, taxEntity)
			if err != nil {
				continue
			}
			tax = tax.Add(taxAmount)
		}
	}
	order.AmountUntaxed = untaxed.Float64()
	order.AmountTax = tax.Float64()
	order.AmountTotal = untaxed.Add(tax).Float64()
}

func (s PurchaseOrderService) recomputeStatuses(order *PurchaseOrder, lines []*PurchaseOrderLine) {
	total := amount.Zero()
	received := amount.Zero()
	billed := amount.Zero()
	returned := amount.Zero()
	for _, line := range lines {
		total = total.Add(amount.FromFloat64(line.QtyOrdered))
		received = received.Add(amount.FromFloat64(line.QtyReceived))
		billed = billed.Add(amount.FromFloat64(line.QtyBilled))
		returned = returned.Add(amount.FromFloat64(line.QtyReturns))
	}
	netReceived := received.Sub(returned)
	netBilled := billed.Sub(returned)
	switch {
	case total.IsZero():
		order.ReceiptStatus = PurchaseOrderReceiptStatusPending
	case netReceived.Equal(total):
		order.ReceiptStatus = PurchaseOrderReceiptStatusDone
	case netReceived.IsZero():
		order.ReceiptStatus = PurchaseOrderReceiptStatusPending
	default:
		order.ReceiptStatus = PurchaseOrderReceiptStatusPartial
	}
	switch {
	case netBilled.IsZero():
		order.InvoiceStatus = PurchaseOrderInvoiceStatusNo
	case netBilled.Equal(total):
		order.InvoiceStatus = PurchaseOrderInvoiceStatusInvoiced
	default:
		order.InvoiceStatus = PurchaseOrderInvoiceStatusToInvoice
	}
}

func (s PurchaseOrderService) RecordReturn(ctx context.Context, orderID, itemID uint64, qty float64) error {
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
		return ErrPurchaseOrderLineNotFound
	}
	s.recomputeStatuses(order, lines)
	_, err = s.orders.Update(ctx, order)
	return err
}

func (s PurchaseOrderService) findOrCreateShipment(ctx context.Context, order *PurchaseOrder) (*inventory.Shipment, error) {
	if order.Name == nil {
		return nil, ErrPurchaseOrderNotFound
	}
	shipment, err := s.shipmentFor(ctx, order)
	if err != nil {
		return nil, err
	}
	if shipment != nil {
		return shipment, nil
	}

	supplierLocation, err := s.supplierLocation(ctx, order.OrganizationID)
	if err != nil {
		return nil, err
	}
	destLocation, err := s.destLocation(ctx, order)
	if err != nil {
		return nil, err
	}

	lines, err := s.lines.ListByOrder(ctx, order.ID)
	if err != nil {
		return nil, err
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
			SrcLocationID:  supplierLocation,
			DstLocationID:  destLocation,
			State:          inventory.MovementStateConfirmed,
			OriginType:     helper.Ptr(ApprovalOwnerType),
			OriginID:       helper.Ptr(order.ID),
			ScheduledDate:  order.ExpectedDate,
		})
	}

	created, err := s.shipments.CreateWithMovements(ctx, &inventory.Shipment{
		OrganizationID: order.OrganizationID,
		Name:           order.Name,
		Type:           inventory.ShipmentTypeIncoming,
		ContactID:      helper.Ptr(order.SupplierID),
		SrcLocationID:  helper.Ptr(supplierLocation),
		DstLocationID:  helper.Ptr(destLocation),
		State:          inventory.ShipmentStateConfirmed,
		Origin:         order.Name,
		ScheduledDate:  order.ExpectedDate,
	}, movements)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s PurchaseOrderService) findOrCreateMovement(ctx context.Context, shipment *inventory.Shipment, line *PurchaseOrderLine) (*inventory.StockMovement, error) {
	if shipment.ID != 0 {
		movements, err := s.movements.ListByShipment(ctx, shipment.ID)
		if err != nil {
			return nil, err
		}
		for _, movement := range movements {
			if movement.ItemID == *line.ItemID {
				return movement, nil
			}
		}
	}
	return nil, ErrPurchaseOrderShipmentNotFound
}

func (s PurchaseOrderService) shipmentFor(ctx context.Context, order *PurchaseOrder) (*inventory.Shipment, error) {
	if order.Name == nil {
		return nil, ErrPurchaseOrderNotFound
	}
	shipments, err := s.shipments.List(ctx, &query.Query{Filters: []query.Filter{{Field: "origin", Operator: query.Equal, Value: *order.Name}}})
	if err != nil {
		return nil, err
	}
	for _, shipment := range shipments.Items {
		if shipment.Type == inventory.ShipmentTypeIncoming {
			return shipment, nil
		}
	}
	return nil, nil
}

func (s PurchaseOrderService) supplierLocation(ctx context.Context, organizationID *uint64) (uint64, error) {
	if organizationID == nil {
		return 0, ErrPurchaseOrderNotFound
	}
	locations, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "usage", Operator: query.Equal, Value: "supplier"}}})
	if err != nil {
		return 0, err
	}
	for _, location := range locations.Items {
		if location.OrganizationID == nil || *location.OrganizationID == *organizationID {
			return location.ID, nil
		}
	}
	return 0, ErrPurchaseOrderSupplierLocation
}

func (s PurchaseOrderService) validateWarehouse(ctx context.Context, organizationID *uint64, warehouseID *uint64) error {
	if warehouseID == nil {
		return nil
	}
	warehouse, err := s.warehouses.Find(ctx, *warehouseID)
	if err != nil {
		return err
	}
	if warehouse == nil || !helper.OwnedByOrg(warehouse.OrganizationID, organizationID) {
		return ErrPurchaseOrderWarehouse
	}
	return nil
}

func (s PurchaseOrderService) destLocation(ctx context.Context, order *PurchaseOrder) (uint64, error) {
	if order.DestLocationID != nil {
		location, err := s.locations.Find(ctx, *order.DestLocationID)
		if err != nil {
			return 0, err
		}
		if location == nil || !helper.OwnedByOrg(location.OrganizationID, order.OrganizationID) {
			return 0, ErrPurchaseOrderLocation
		}
		return location.ID, nil
	}
	if order.WarehouseID == nil {
		return 0, ErrPurchaseOrderLocation
	}
	locations, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "warehouse_id", Operator: query.Equal, Value: *order.WarehouseID}}})
	if err != nil {
		return 0, err
	}
	for _, location := range locations.Items {
		if location.Usage == "internal" && order.OrganizationID != nil && location.OrganizationID != nil && *location.OrganizationID == *order.OrganizationID {
			return location.ID, nil
		}
	}
	return 0, ErrPurchaseOrderLocation
}

func (s PurchaseOrderService) validateVendor(ctx context.Context, order *PurchaseOrder) error {
	contact, err := s.contacts.Find(ctx, order.SupplierID)
	if err != nil {
		return err
	}
	if contact == nil {
		return ErrPurchaseOrderVendor
	}
	supplier, err := s.suppliers.FindByContact(ctx, order.SupplierID)
	if err != nil {
		return err
	}
	if supplier == nil || !supplier.Active {
		return ErrPurchaseOrderVendorNotSupplier
	}
	return nil
}

func (s PurchaseOrderService) threshold(ctx context.Context, order *PurchaseOrder) float64 {
	if order.OrganizationID == nil {
		return 0
	}
	threshold, err := s.configs.ApprovalThreshold(ctx, *order.OrganizationID)
	if err != nil {
		return 0
	}
	return threshold
}

func (s PurchaseOrderService) approved(ctx context.Context, order *PurchaseOrder) bool {
	approved, err := s.approvals.IsApproved(ctx, ApprovalOwnerType, order.ID)
	return err == nil && approved
}

func (s PurchaseOrderService) requestApproval(ctx context.Context, order *PurchaseOrder, byUserID uint64) error {
	approverIDs, err := s.configs.ApproverIDs(ctx, *order.OrganizationID)
	if err != nil {
		return err
	}
	if len(approverIDs) == 0 {
		return ErrPurchaseOrderApprovalRequired
	}
	if _, err := s.approvals.Create(ctx, *order.OrganizationID, ApprovalOwnerType, order.ID, byUserID, approverIDs); err != nil {
		return err
	}
	return nil
}

func (s PurchaseOrderService) requisitionLines(ctx context.Context, requestID uint64) ([]*PurchaseRequestLine, error) {
	return s.linesDAO.ListByRequest(ctx, requestID)
}

func (s PurchaseOrderService) findOrder(ctx context.Context, orderID uint64) (*PurchaseOrder, error) {
	order, err := s.orders.Find(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrPurchaseOrderNotFound
	}
	return order, nil
}

func amountFromFloat64(f float64) amount.Amount {
	return amount.FromFloat64(f)
}
