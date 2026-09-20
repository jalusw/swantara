package procurement

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/kernel/state"
)

type OrderCreator interface {
	Create(ctx context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error)
}

type SupplierQuoteRequestService struct {
	quote_requests SupplierQuoteRequestDAO
	rfqLines       SupplierQuoteRequestLineDAO
	quotes         SupplierQuoteDAO
	quoteLines     SupplierQuoteLineDAO
	requisitions   PurchaseRequestDAO
	reqLines       PurchaseRequestLineDAO
	sequences      sequence.Service
	contacts       contacts.ContactDAO
	orders         OrderCreator
	machine        state.Machine
	quoteMachine   state.Machine
}

func NewSupplierQuoteRequestService(
	quote_requests SupplierQuoteRequestDAO,
	rfqLines SupplierQuoteRequestLineDAO,
	quotes SupplierQuoteDAO,
	quoteLines SupplierQuoteLineDAO,
	requisitions PurchaseRequestDAO,
	requisitionLines PurchaseRequestLineDAO,
	sequences sequence.Service,
	contacts contacts.ContactDAO,
	orders OrderCreator,
) SupplierQuoteRequestService {
	return SupplierQuoteRequestService{
		quote_requests: quote_requests,
		rfqLines:       rfqLines,
		quotes:         quotes,
		quoteLines:     quoteLines,
		requisitions:   requisitions,
		reqLines:       requisitionLines,
		sequences:      sequences,
		contacts:       contacts,
		orders:         orders,
		machine: state.NewMachine(
			state.Transition{From: model.Status(QuoteRequestStateDraft), To: model.Status(QuoteRequestStateSent)},
			state.Transition{From: model.Status(QuoteRequestStateDraft), To: model.Status(QuoteRequestStateCancelled)},
			state.Transition{From: model.Status(QuoteRequestStateSent), To: model.Status(QuoteRequestStateDone)},
			state.Transition{From: model.Status(QuoteRequestStateSent), To: model.Status(QuoteRequestStateCancelled)},
		),
		quoteMachine: state.NewMachine(
			state.Transition{From: model.Status(SupplierQuoteStateSubmitted), To: model.Status(SupplierQuoteStateAccepted)},
			state.Transition{From: model.Status(SupplierQuoteStateSubmitted), To: model.Status(SupplierQuoteStateRejected)},
		),
	}
}

func (s SupplierQuoteRequestService) Create(ctx context.Context, quoteRequest *SupplierQuoteRequest, lines []*SupplierQuoteRequestLine) (*SupplierQuoteRequest, error) {
	if quoteRequest.OrganizationID == nil {
		return nil, ErrRFQNotFound
	}
	if len(lines) == 0 {
		return nil, ErrRFQNoLines
	}
	if err := s.validateRequester(ctx, quoteRequest.RequesterID); err != nil {
		return nil, err
	}
	for _, line := range lines {
		if !amount.FromFloat64(line.Qty).GreaterThan(amount.Zero()) {
			return nil, ErrRFQLineQty
		}
	}

	name, err := s.sequences.Next(ctx, *quoteRequest.OrganizationID, SequenceSupplierQuoteRequestCode)
	if err != nil {
		return nil, err
	}
	quoteRequest.Name = helper.Ptr(name)
	quoteRequest.State = QuoteRequestStateDraft
	if quoteRequest.CurrencyCode == nil {
		quoteRequest.CurrencyCode = helper.Ptr("IDR")
	}
	return s.quote_requests.CreateWithLines(ctx, quoteRequest, lines)
}

func (s SupplierQuoteRequestService) CreateFromRequest(ctx context.Context, requestID uint64, quoteRequest *SupplierQuoteRequest) (*SupplierQuoteRequest, error) {
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
	requisitionLines, err := s.reqLines.ListByRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if len(requisitionLines) == 0 {
		return nil, ErrRequisitionNoLines
	}

	lines := make([]*SupplierQuoteRequestLine, 0, len(requisitionLines))
	for _, requisitionLine := range requisitionLines {
		if requisitionLine.ItemID == nil {
			continue
		}
		lines = append(lines, &SupplierQuoteRequestLine{
			ItemID:      requisitionLine.ItemID,
			Description: requisitionLine.Description,
			Qty:         requisitionLine.Qty,
			UnitID:      requisitionLine.UnitID,
			NeededBy:    requisitionLine.NeededBy,
		})
	}
	if len(lines) == 0 {
		return nil, ErrRFQNoLines
	}
	if quoteRequest.OrganizationID == nil {
		quoteRequest.OrganizationID = requisition.OrganizationID
	}
	if quoteRequest.RequesterID == 0 {
		quoteRequest.RequesterID = requisition.RequesterID
	}
	return s.Create(ctx, quoteRequest, lines)
}

func (s SupplierQuoteRequestService) Send(ctx context.Context, rfqID uint64) (*SupplierQuoteRequest, error) {
	quoteRequest, err := s.findRFQ(ctx, rfqID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(quoteRequest.State), model.Status(QuoteRequestStateSent)); err != nil {
		return nil, ErrQuoteRequestState
	}
	quoteRequest.State = QuoteRequestStateSent
	return s.quote_requests.Update(ctx, quoteRequest)
}

func (s SupplierQuoteRequestService) Cancel(ctx context.Context, rfqID uint64) (*SupplierQuoteRequest, error) {
	quoteRequest, err := s.findRFQ(ctx, rfqID)
	if err != nil {
		return nil, err
	}
	if err := s.machine.TryTransition(model.Status(quoteRequest.State), model.Status(QuoteRequestStateCancelled)); err != nil {
		return nil, ErrQuoteRequestState
	}
	quoteRequest.State = QuoteRequestStateCancelled
	return s.quote_requests.Update(ctx, quoteRequest)
}

func (s SupplierQuoteRequestService) List(ctx context.Context, q *query.Query) (*query.Page[SupplierQuoteRequest], error) {
	return s.quote_requests.List(ctx, q)
}

func (s SupplierQuoteRequestService) Find(ctx context.Context, id uint64) (*SupplierQuoteRequest, error) {
	return s.quote_requests.Find(ctx, id)
}

func (s SupplierQuoteRequestService) ListLines(ctx context.Context, rfqID uint64) ([]*SupplierQuoteRequestLine, error) {
	return s.rfqLines.ListByRFQ(ctx, rfqID)
}

func (s SupplierQuoteRequestService) ListQuotes(ctx context.Context, rfqID uint64) ([]*SupplierQuote, error) {
	page, err := s.quotes.List(ctx, &query.Query{Filters: []query.Filter{{Field: "quoteRequest_id", Operator: query.Equal, Value: rfqID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s SupplierQuoteRequestService) SubmitQuote(ctx context.Context, rfqID uint64, quote *SupplierQuote, lines []*SupplierQuoteLine) (*SupplierQuote, error) {
	quoteRequest, err := s.findRFQ(ctx, rfqID)
	if err != nil {
		return nil, err
	}
	if quoteRequest.State != QuoteRequestStateSent {
		return nil, ErrQuoteRequestState
	}
	if quote.SupplierID == 0 {
		return nil, ErrRFQQuoteVendor
	}
	supplier, err := s.contacts.Find(ctx, quote.SupplierID)
	if err != nil {
		return nil, err
	}
	if supplier == nil {
		return nil, ErrRFQQuoteVendor
	}
	if len(lines) == 0 {
		return nil, ErrRFQQuoteNoLines
	}

	rfqLines, err := s.rfqLines.ListByRFQ(ctx, rfqID)
	if err != nil {
		return nil, err
	}
	known := map[uint64]bool{}
	for _, rfqLine := range rfqLines {
		known[rfqLine.ID] = true
	}

	untaxed := amount.Zero()
	for _, line := range lines {
		if !amount.FromFloat64(line.Qty).GreaterThan(amount.Zero()) {
			return nil, ErrRFQQuoteLineQty
		}
		if !known[line.QuoteRequestLineID] {
			return nil, ErrRFQQuoteDuplicateLine
		}
		if line.UnitPrice < 0 {
			return nil, ErrRFQQuotePrice
		}
		subtotal := amount.FromFloat64(line.Qty).Mul(amount.FromFloat64(line.UnitPrice)).Round(4)
		line.PriceSubtotal = subtotal.Float64()
		untaxed = untaxed.Add(subtotal)
	}
	quote.QuoteRequestID = rfqID
	quote.State = SupplierQuoteStateSubmitted
	quote.AmountUntaxed = untaxed.Float64()
	quote.AmountTotal = untaxed.Float64()
	if quote.CurrencyCode == nil {
		quote.CurrencyCode = quoteRequest.CurrencyCode
	}
	if quote.QuoteDate == nil {
		now := time.Now()
		quote.QuoteDate = &now
	}
	return s.quotes.CreateWithLines(ctx, quote, lines)
}

func (s SupplierQuoteRequestService) AcceptQuote(ctx context.Context, quoteID uint64) (*SupplierQuoteRequest, error) {
	quote, err := s.findQuote(ctx, quoteID)
	if err != nil {
		return nil, err
	}
	if err := s.quoteMachine.TryTransition(model.Status(quote.State), model.Status(SupplierQuoteStateAccepted)); err != nil {
		return nil, ErrSupplierQuoteState
	}
	quoteRequest, err := s.findRFQ(ctx, quote.QuoteRequestID)
	if err != nil {
		return nil, err
	}
	if quoteRequest.State != QuoteRequestStateSent {
		return nil, ErrQuoteRequestState
	}

	quotesPage, err := s.quotes.List(ctx, &query.Query{Filters: []query.Filter{{Field: "quoteRequest_id", Operator: query.Equal, Value: quoteRequest.ID}}})
	if err != nil {
		return nil, err
	}
	for _, other := range quotesPage.Items {
		if other.ID == quote.ID {
			continue
		}
		if other.State == SupplierQuoteStateSubmitted {
			other.State = SupplierQuoteStateRejected
			if _, err := s.quotes.Update(ctx, other); err != nil {
				return nil, err
			}
		}
	}

	quote.State = SupplierQuoteStateAccepted
	if _, err := s.quotes.Update(ctx, quote); err != nil {
		return nil, err
	}

	quoteRequest.State = QuoteRequestStateDone
	return s.quote_requests.Update(ctx, quoteRequest)
}

func (s SupplierQuoteRequestService) CreatePurchaseOrder(ctx context.Context, rfqID uint64, order *PurchaseOrder) (*PurchaseOrder, error) {
	quoteRequest, err := s.findRFQ(ctx, rfqID)
	if err != nil {
		return nil, err
	}
	if quoteRequest.State != QuoteRequestStateDone {
		return nil, ErrRFQNotConvertible
	}

	quotesPage, err := s.quotes.List(ctx, &query.Query{Filters: []query.Filter{{Field: "quoteRequest_id", Operator: query.Equal, Value: rfqID}}})
	if err != nil {
		return nil, err
	}
	var accepted *SupplierQuote
	for _, quote := range quotesPage.Items {
		if quote.State == SupplierQuoteStateAccepted {
			accepted = quote
			break
		}
	}
	if accepted == nil {
		return nil, ErrRFQQuoteNoAccepted
	}

	quoteLines, err := s.quoteLines.ListByQuote(ctx, accepted.ID)
	if err != nil {
		return nil, err
	}
	if len(quoteLines) == 0 {
		return nil, ErrRFQQuoteNoLines
	}
	lines := make([]*PurchaseOrderLine, 0, len(quoteLines))
	for _, quoteLine := range quoteLines {
		lines = append(lines, &PurchaseOrderLine{
			ItemID:        quoteLine.ItemID,
			Description:   quoteLine.Description,
			QtyOrdered:    quoteLine.Qty,
			UnitID:        nil,
			QtyReceived:   0,
			QtyBilled:     0,
			UnitPrice:     quoteLine.UnitPrice,
			DiscountPct:   quoteLine.DiscountPct,
			PriceSubtotal: quoteLine.PriceSubtotal,
		})
	}
	if order.OrganizationID == nil {
		order.OrganizationID = quoteRequest.OrganizationID
	}
	if order.SupplierID == 0 {
		order.SupplierID = accepted.SupplierID
	}
	return s.orders.Create(ctx, order, lines)
}

func (s SupplierQuoteRequestService) validateRequester(ctx context.Context, requesterID uint64) error {
	if requesterID == 0 {
		return ErrRFQRequester
	}
	requester, err := s.contacts.Find(ctx, requesterID)
	if err != nil {
		return err
	}
	if requester == nil {
		return ErrRFQRequester
	}
	return nil
}

func (s SupplierQuoteRequestService) findRFQ(ctx context.Context, rfqID uint64) (*SupplierQuoteRequest, error) {
	quoteRequest, err := s.quote_requests.Find(ctx, rfqID)
	if err != nil {
		return nil, err
	}
	if quoteRequest == nil {
		return nil, ErrRFQNotFound
	}
	return quoteRequest, nil
}

func (s SupplierQuoteRequestService) findQuote(ctx context.Context, quoteID uint64) (*SupplierQuote, error) {
	quote, err := s.quotes.Find(ctx, quoteID)
	if err != nil {
		return nil, err
	}
	if quote == nil {
		return nil, ErrRFQQuoteNotFound
	}
	return quote, nil
}
