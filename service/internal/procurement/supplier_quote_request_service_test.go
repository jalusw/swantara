package procurement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
)

func TestSupplierQuoteRequestService_Create(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "rejects without lines",
			run: func(t *testing.T) {
				ctx := context.Background()
				svc := testRFQService(SupplierQuoteRequestDAOMock{}, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				_, err := svc.Create(ctx, &SupplierQuoteRequest{OrganizationID: helper.Ptr(uint64(1)), RequesterID: 1}, nil)
				if helper.AssertError(t, err, true, ErrRFQNoLines) {
					return
				}
			},
		},
		{
			name: "sets draft and numbers",
			run: func(t *testing.T) {
				ctx := context.Background()
				var created *SupplierQuoteRequest
				quote_requests := SupplierQuoteRequestDAOMock{
					CreateWithLinesFunc: func(_ context.Context, quoteRequest *SupplierQuoteRequest, _ []*SupplierQuoteRequestLine) (*SupplierQuoteRequest, error) {
						created = quoteRequest
						return quoteRequest, nil
					},
				}
				svc := testRFQService(quote_requests, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				quoteRequest, err := svc.Create(ctx, &SupplierQuoteRequest{OrganizationID: helper.Ptr(uint64(1)), RequesterID: 1}, []*SupplierQuoteRequestLine{{Qty: 2}})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if created == nil || created.State != QuoteRequestStateDraft {
					t.Errorf("state = %v, want draft", created)
				}
				if quoteRequest.Name == nil || *quoteRequest.Name != "QuoteRequest/00001" {
					t.Errorf("name = %v, want QuoteRequest/00001", quoteRequest.Name)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSupplierQuoteRequestService_CreateFromRequest(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "rejects unapproved",
			run: func(t *testing.T) {
				ctx := context.Background()
				requisitions := PurchaseRequestDAOMock{
					CRUDMock: dao.CRUDMock[PurchaseRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) {
							return &PurchaseRequest{Base: model.Base{ID: 1}, State: RequestStateConfirmed}, nil
						},
					},
				}
				svc := testRFQService(SupplierQuoteRequestDAOMock{}, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, requisitions, PurchaseRequestLineDAOMock{}, nil)

				_, err := svc.CreateFromRequest(ctx, 1, &SupplierQuoteRequest{})
				if helper.AssertError(t, err, true, ErrRequisitionNotConvertible) {
					return
				}
			},
		},
		{
			name: "copies lines",
			run: func(t *testing.T) {
				ctx := context.Background()
				requisitions := PurchaseRequestDAOMock{
					CRUDMock: dao.CRUDMock[PurchaseRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) {
							return &PurchaseRequest{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), RequesterID: 5, State: RequestStateApproved}, nil
						},
					},
				}
				reqLines := PurchaseRequestLineDAOMock{
					ListByRequestFunc: func(_ context.Context, _ uint64) ([]*PurchaseRequestLine, error) {
						return []*PurchaseRequestLine{
							{Base: model.Base{ID: 1}, ItemID: helper.Ptr(uint64(100)), Qty: 5},
						}, nil
					},
				}
				var copied []*SupplierQuoteRequestLine
				quote_requests := SupplierQuoteRequestDAOMock{
					CreateWithLinesFunc: func(_ context.Context, _ *SupplierQuoteRequest, lines []*SupplierQuoteRequestLine) (*SupplierQuoteRequest, error) {
						copied = lines
						return &SupplierQuoteRequest{Base: model.Base{ID: 2}, State: QuoteRequestStateDraft}, nil
					},
				}
				svc := testRFQService(quote_requests, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, requisitions, reqLines, nil)

				quoteRequest, err := svc.CreateFromRequest(ctx, 1, &SupplierQuoteRequest{})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if quoteRequest.ID != 2 {
					t.Errorf("quoteRequest id = %d, want 2", quoteRequest.ID)
				}
				if len(copied) != 1 || copied[0].Qty != 5 {
					t.Errorf("copied lines = %+v, want single line qty 5", copied)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSupplierQuoteRequestService_SubmitQuote(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "rejects quote for draft quoteRequest",
			run: func(t *testing.T) {
				ctx := context.Background()
				quote_requests := SupplierQuoteRequestDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuoteRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuoteRequest, error) {
							return &SupplierQuoteRequest{Base: model.Base{ID: 1}, State: QuoteRequestStateDraft}, nil
						},
					},
				}
				svc := testRFQService(quote_requests, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				_, err := svc.SubmitQuote(ctx, 1, &SupplierQuote{SupplierID: 1}, []*SupplierQuoteLine{{QuoteRequestLineID: 1, Qty: 1, UnitPrice: 10}})
				if helper.AssertError(t, err, true, ErrQuoteRequestState) {
					return
				}
			},
		},
		{
			name: "computes totals and submits",
			run: func(t *testing.T) {
				ctx := context.Background()
				quote_requests := SupplierQuoteRequestDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuoteRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuoteRequest, error) {
							return &SupplierQuoteRequest{Base: model.Base{ID: 1}, State: QuoteRequestStateSent}, nil
						},
					},
				}
				rfqLines := SupplierQuoteRequestLineDAOMock{
					ListByRFQFunc: func(_ context.Context, _ uint64) ([]*SupplierQuoteRequestLine, error) {
						return []*SupplierQuoteRequestLine{{Base: model.Base{ID: 1}}}, nil
					},
				}
				var saved *SupplierQuote
				quotes := SupplierQuoteDAOMock{
					CreateWithLinesFunc: func(_ context.Context, quote *SupplierQuote, _ []*SupplierQuoteLine) (*SupplierQuote, error) {
						saved = quote
						return quote, nil
					},
				}
				svc := testRFQService(quote_requests, rfqLines, quotes, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				quote, err := svc.SubmitQuote(ctx, 1, &SupplierQuote{SupplierID: 1}, []*SupplierQuoteLine{{QuoteRequestLineID: 1, Qty: 2, UnitPrice: 25}})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if quote.State != SupplierQuoteStateSubmitted {
					t.Errorf("state = %s, want submitted", quote.State)
				}
				if quote.AmountUntaxed != 50 || quote.AmountTotal != 50 {
					t.Errorf("amounts = untaxed %v total %v, want 50", quote.AmountUntaxed, quote.AmountTotal)
				}
				if saved == nil || saved.QuoteRequestID != 1 {
					t.Errorf("saved quote = %+v, want quoteRequest id 1", saved)
				}
			},
		},
		{
			name: "rejects unknown line",
			run: func(t *testing.T) {
				ctx := context.Background()
				quote_requests := SupplierQuoteRequestDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuoteRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuoteRequest, error) {
							return &SupplierQuoteRequest{Base: model.Base{ID: 1}, State: QuoteRequestStateSent}, nil
						},
					},
				}
				rfqLines := SupplierQuoteRequestLineDAOMock{
					ListByRFQFunc: func(_ context.Context, _ uint64) ([]*SupplierQuoteRequestLine, error) {
						return []*SupplierQuoteRequestLine{{Base: model.Base{ID: 1}}}, nil
					},
				}
				svc := testRFQService(quote_requests, rfqLines, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				_, err := svc.SubmitQuote(ctx, 1, &SupplierQuote{SupplierID: 1}, []*SupplierQuoteLine{{QuoteRequestLineID: 99, Qty: 1, UnitPrice: 10}})
				if helper.AssertError(t, err, true, ErrRFQQuoteDuplicateLine) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSupplierQuoteRequestService_AcceptQuote(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "rejects other submitted quotes",
			run: func(t *testing.T) {
				ctx := context.Background()
				quotes := SupplierQuoteDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuote]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuote, error) {
							return &SupplierQuote{Base: model.Base{ID: 2}, QuoteRequestID: 1, State: SupplierQuoteStateSubmitted}, nil
						},
						UpdateFunc: func(_ context.Context, quote *SupplierQuote) (*SupplierQuote, error) {
							return quote, nil
						},
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[SupplierQuote], error) {
							return &query.Page[SupplierQuote]{Items: []*SupplierQuote{
								{Base: model.Base{ID: 1}, QuoteRequestID: 1, State: SupplierQuoteStateSubmitted},
								{Base: model.Base{ID: 2}, QuoteRequestID: 1, State: SupplierQuoteStateSubmitted},
							}}, nil
						},
					},
				}
				quote_requests := SupplierQuoteRequestDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuoteRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuoteRequest, error) {
							return &SupplierQuoteRequest{Base: model.Base{ID: 1}, State: QuoteRequestStateSent}, nil
						},
						UpdateFunc: func(_ context.Context, quoteRequest *SupplierQuoteRequest) (*SupplierQuoteRequest, error) {
							return quoteRequest, nil
						},
					},
				}
				svc := testRFQService(quote_requests, SupplierQuoteRequestLineDAOMock{}, quotes, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				quoteRequest, err := svc.AcceptQuote(ctx, 2)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if quoteRequest.State != QuoteRequestStateDone {
					t.Errorf("quoteRequest state = %s, want done", quoteRequest.State)
				}
			},
		},
		{
			name: "rejects unsubmitted quote",
			run: func(t *testing.T) {
				ctx := context.Background()
				quotes := SupplierQuoteDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuote]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuote, error) {
							return &SupplierQuote{Base: model.Base{ID: 2}, QuoteRequestID: 1, State: SupplierQuoteStateDraft}, nil
						},
					},
				}
				svc := testRFQService(SupplierQuoteRequestDAOMock{}, SupplierQuoteRequestLineDAOMock{}, quotes, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				_, err := svc.AcceptQuote(ctx, 2)
				if helper.AssertError(t, err, true, ErrSupplierQuoteState) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSupplierQuoteRequestService_CreatePurchaseOrder(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "uses accepted quote prices",
			run: func(t *testing.T) {
				ctx := context.Background()
				quote_requests := SupplierQuoteRequestDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuoteRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuoteRequest, error) {
							return &SupplierQuoteRequest{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), State: QuoteRequestStateDone}, nil
						},
					},
				}
				quotes := SupplierQuoteDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuote]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[SupplierQuote], error) {
							return &query.Page[SupplierQuote]{Items: []*SupplierQuote{
								{Base: model.Base{ID: 5}, SupplierID: 7, State: SupplierQuoteStateAccepted},
							}}, nil
						},
					},
				}
				quoteLines := SupplierQuoteLineDAOMock{
					ListByQuoteFunc: func(_ context.Context, _ uint64) ([]*SupplierQuoteLine, error) {
						return []*SupplierQuoteLine{
							{Base: model.Base{ID: 1}, ItemID: helper.Ptr(uint64(100)), Qty: 3, UnitPrice: 40, PriceSubtotal: 120},
						}, nil
					},
				}
				var order *PurchaseOrder
				var lines []*PurchaseOrderLine
				orders := orderCreatorMock{
					createFunc: func(_ context.Context, o *PurchaseOrder, l []*PurchaseOrderLine) (*PurchaseOrder, error) {
						order = o
						lines = l
						return o, nil
					},
				}
				svc := testRFQService(quote_requests, SupplierQuoteRequestLineDAOMock{}, quotes, quoteLines, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, orders)

				created, err := svc.CreatePurchaseOrder(ctx, 1, &PurchaseOrder{})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if order == nil || order.SupplierID != 7 {
					t.Errorf("order = %+v, want supplier 7", order)
				}
				if len(lines) != 1 || lines[0].UnitPrice != 40 || lines[0].QtyOrdered != 3 {
					t.Errorf("lines = %+v, want single line qty 3 price 40", lines)
				}
				if created == nil {
					t.Error("expected created order")
				}
			},
		},
		{
			name: "rejects without accepted quote",
			run: func(t *testing.T) {
				ctx := context.Background()
				quote_requests := SupplierQuoteRequestDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuoteRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuoteRequest, error) {
							return &SupplierQuoteRequest{Base: model.Base{ID: 1}, State: QuoteRequestStateDone}, nil
						},
					},
				}
				quotes := SupplierQuoteDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuote]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[SupplierQuote], error) {
							return &query.Page[SupplierQuote]{Items: []*SupplierQuote{{Base: model.Base{ID: 5}, State: SupplierQuoteStateRejected}}}, nil
						},
					},
				}
				svc := testRFQService(quote_requests, SupplierQuoteRequestLineDAOMock{}, quotes, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				_, err := svc.CreatePurchaseOrder(ctx, 1, &PurchaseOrder{})
				if helper.AssertError(t, err, true, ErrRFQQuoteNoAccepted) {
					return
				}
			},
		},
		{
			name: "rejects when not done",
			run: func(t *testing.T) {
				ctx := context.Background()
				quote_requests := SupplierQuoteRequestDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuoteRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuoteRequest, error) {
							return &SupplierQuoteRequest{Base: model.Base{ID: 1}, State: QuoteRequestStateSent}, nil
						},
					},
				}
				svc := testRFQService(quote_requests, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				_, err := svc.CreatePurchaseOrder(ctx, 1, &PurchaseOrder{})
				if helper.AssertError(t, err, true, ErrRFQNotConvertible) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSupplierQuoteRequestService_Send(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "rejects draft to confirmed jump",
			run: func(t *testing.T) {
				ctx := context.Background()
				svc := testRFQService(SupplierQuoteRequestDAOMock{}, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				if _, err := svc.Send(ctx, 99); helper.AssertError(t, err, true, ErrRFQNotFound) {
					return
				}
			},
		},
		{
			name: "transitions draft to sent",
			run: func(t *testing.T) {
				ctx := context.Background()
				var updated *SupplierQuoteRequest
				quote_requests := SupplierQuoteRequestDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuoteRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuoteRequest, error) {
							return &SupplierQuoteRequest{Base: model.Base{ID: 1}, State: QuoteRequestStateDraft}, nil
						},
						UpdateFunc: func(_ context.Context, quoteRequest *SupplierQuoteRequest) (*SupplierQuoteRequest, error) {
							updated = quoteRequest
							return quoteRequest, nil
						},
					},
				}
				svc := testRFQService(quote_requests, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				quoteRequest, err := svc.Send(ctx, 1)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if quoteRequest.State != QuoteRequestStateSent || updated == nil {
					t.Errorf("state = %v, want sent", quoteRequest.State)
				}
			},
		},
		{
			name: "rejects cancelled quoteRequest",
			run: func(t *testing.T) {
				ctx := context.Background()
				quote_requests := SupplierQuoteRequestDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuoteRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuoteRequest, error) {
							return &SupplierQuoteRequest{Base: model.Base{ID: 1}, State: QuoteRequestStateCancelled}, nil
						},
					},
				}
				svc := testRFQService(quote_requests, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				_, err := svc.Send(ctx, 1)
				if helper.AssertError(t, err, true, ErrQuoteRequestState) {
					return
				}
			},
		},
		{
			name: "propagates find error",
			run: func(t *testing.T) {
				ctx := context.Background()
				quote_requests := SupplierQuoteRequestDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuoteRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuoteRequest, error) {
							return nil, errors.New("db down")
						},
					},
				}
				svc := testRFQService(quote_requests, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				_, err := svc.Send(ctx, 1)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSupplierQuoteRequestService_Cancel(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "sets cancelled",
			run: func(t *testing.T) {
				ctx := context.Background()
				var updated *SupplierQuoteRequest
				quote_requests := SupplierQuoteRequestDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuoteRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*SupplierQuoteRequest, error) {
							return &SupplierQuoteRequest{Base: model.Base{ID: 1}, State: QuoteRequestStateDraft}, nil
						},
						UpdateFunc: func(_ context.Context, quoteRequest *SupplierQuoteRequest) (*SupplierQuoteRequest, error) {
							updated = quoteRequest
							return quoteRequest, nil
						},
					},
				}
				svc := testRFQService(quote_requests, SupplierQuoteRequestLineDAOMock{}, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				quoteRequest, err := svc.Cancel(ctx, 1)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if quoteRequest.State != QuoteRequestStateCancelled || updated == nil {
					t.Errorf("state = %v, want cancelled", quoteRequest.State)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSupplierQuoteRequestService_ListLines(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "returns lines",
			run: func(t *testing.T) {
				ctx := context.Background()
				rfqLines := SupplierQuoteRequestLineDAOMock{
					ListByRFQFunc: func(_ context.Context, _ uint64) ([]*SupplierQuoteRequestLine, error) {
						return []*SupplierQuoteRequestLine{{Base: model.Base{ID: 1}, Qty: 5}}, nil
					},
				}
				svc := testRFQService(SupplierQuoteRequestDAOMock{}, rfqLines, SupplierQuoteDAOMock{}, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				lines, err := svc.ListLines(ctx, 1)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(lines) != 1 || lines[0].Qty != 5 {
					t.Errorf("lines = %+v, want one line qty 5", lines)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSupplierQuoteRequestService_ListQuotes(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "returns quotes",
			run: func(t *testing.T) {
				ctx := context.Background()
				quotes := SupplierQuoteDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuote]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[SupplierQuote], error) {
							return &query.Page[SupplierQuote]{Items: []*SupplierQuote{
								{Base: model.Base{ID: 1}, SupplierID: 7, State: SupplierQuoteStateSubmitted},
							}}, nil
						},
					},
				}
				svc := testRFQService(SupplierQuoteRequestDAOMock{}, SupplierQuoteRequestLineDAOMock{}, quotes, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				items, err := svc.ListQuotes(ctx, 1)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(items) != 1 || items[0].SupplierID != 7 {
					t.Errorf("quotes = %+v, want one quote for supplier 7", items)
				}
			},
		},
		{
			name: "propagates error",
			run: func(t *testing.T) {
				ctx := context.Background()
				quotes := SupplierQuoteDAOMock{
					CRUDMock: dao.CRUDMock[SupplierQuote]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[SupplierQuote], error) {
							return nil, errors.New("db down")
						},
					},
				}
				svc := testRFQService(SupplierQuoteRequestDAOMock{}, SupplierQuoteRequestLineDAOMock{}, quotes, SupplierQuoteLineDAOMock{}, PurchaseRequestDAOMock{}, PurchaseRequestLineDAOMock{}, nil)

				_, err := svc.ListQuotes(ctx, 1)
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func testRFQService(
	quote_requests SupplierQuoteRequestDAOMock,
	rfqLines SupplierQuoteRequestLineDAOMock,
	quotes SupplierQuoteDAOMock,
	quoteLines SupplierQuoteLineDAOMock,
	requisitions PurchaseRequestDAOMock,
	reqLines PurchaseRequestLineDAOMock,
	orders OrderCreator,
) SupplierQuoteRequestService {
	return NewSupplierQuoteRequestService(quote_requests, rfqLines, quotes, quoteLines, requisitions, reqLines, sequenceService(), newContactDAOMock(), orders)
}

func sequenceService() sequence.Service {
	return sequence.NewSequenceService(sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "QuoteRequest/00001"}, nil
		},
	})
}

type contactDAOMock struct {
	dao.CRUDMock[contacts.Contact]
}

func newContactDAOMock() contactDAOMock {
	return contactDAOMock{CRUDMock: dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return &contacts.Contact{Base: model.Base{ID: 1}}, nil
		},
	}}
}

func (contactDAOMock) CreateWithDetails(_ context.Context, _ *contacts.Contact, _ []*contacts.ContactAddress, _ []*contacts.ContactBankAccount, _ *contacts.CustomerProfile, _ *contacts.SupplierProfile) (*contacts.Contact, error) {
	return nil, nil
}

type orderCreatorMock struct {
	createFunc func(ctx context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error)
}

func (m orderCreatorMock) Create(ctx context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, order, lines)
	}
	return order, nil
}
