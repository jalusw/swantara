package inventory

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestInboundCostService_Create(t *testing.T) {
	ctx := context.Background()
	var createdLines []*InboundCostLine

	costsDefault := InboundCostDAOMock{}
	linesDefault := InboundCostLineDAOMock{}
	costsError := InboundCostDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *InboundCost) (*InboundCost, error) {
			return nil, ErrInboundCostNotFound
		},
	}
	linesError := InboundCostLineDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *InboundCostLine) (*InboundCostLine, error) {
			return nil, ErrInboundCostNoLines
		},
	}

	tests := []struct {
		name    string
		costs   InboundCostDAOMock
		lines   InboundCostLineDAOMock
		request CreateInboundCostRequest
		wantErr error
	}{
		{
			"persists cost and lines",
			InboundCostDAOMock{
				CreateTxFunc: func(_ context.Context, _ *gorm.DB, cost *InboundCost) (*InboundCost, error) {
					cost.ID = 9
					return cost, nil
				},
			},
			InboundCostLineDAOMock{
				CreateTxFunc: func(_ context.Context, _ *gorm.DB, line *InboundCostLine) (*InboundCostLine, error) {
					createdLines = append(createdLines, line)
					return line, nil
				},
			},
			CreateInboundCostRequest{
				OrganizationID:    10,
				Name:              "Freight",
				TargetShipmentIDs: helper.Int64Array{30},
				Lines: []CreateInboundCostLineRequest{
					{ItemID: 11, Description: "Freight cost", Amount: 100, SplitMethod: SplitMethodQuantity, AccountID: helper.Ptr(uint64(700))},
				},
			},
			nil,
		},
		{
			"no lines",
			costsDefault,
			linesDefault,
			CreateInboundCostRequest{OrganizationID: 10, TargetShipmentIDs: helper.Int64Array{30}, Lines: []CreateInboundCostLineRequest{}},
			ErrInboundCostNoLines,
		},
		{
			"no shipments",
			costsDefault,
			linesDefault,
			CreateInboundCostRequest{OrganizationID: 10, Lines: []CreateInboundCostLineRequest{{ItemID: 11}}},
			ErrInboundCostNoShipments,
		},
		{
			"propagates cost create error",
			costsError,
			linesDefault,
			CreateInboundCostRequest{OrganizationID: 10, TargetShipmentIDs: helper.Int64Array{30}, Lines: []CreateInboundCostLineRequest{{ItemID: 11}}},
			ErrInboundCostNotFound,
		},
		{
			"propagates line create error",
			costsDefault,
			linesError,
			CreateInboundCostRequest{OrganizationID: 10, TargetShipmentIDs: helper.Int64Array{30}, Lines: []CreateInboundCostLineRequest{{ItemID: 11}}},
			ErrInboundCostNoLines,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewInboundCostService(tt.costs, tt.lines, InboundCostAdjustmentDAOMock{}, StockMovementDAOMock{}, CostLayerDAOMock{}, ItemResolverMock{}, PosterMock{}, inboundCostConfigSourceMock{}, vendorBillLineLookupMock{}, TransactionerMock{})
			created, err := svc.Create(ctx, tt.request)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if tt.name == "persists cost and lines" {
				if created.ID != 9 || created.State != InboundCostStateDraft {
					t.Errorf("cost = %+v, want draft with id 9", created)
				}
				if len(createdLines) != 1 || createdLines[0].ItemID != 11 || createdLines[0].Amount != 100 {
					t.Errorf("lines = %+v, want one line for item 11", createdLines)
				}
			}
		})
	}
}

func TestInboundCostService_Get(t *testing.T) {
	ctx := context.Background()
	costs := InboundCostDAOMock{
		CRUDMock: dao.CRUDMock[InboundCost]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*InboundCost, error) {
				return &InboundCost{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(99)), State: InboundCostStateDraft}, nil
			},
		},
	}
	svc := NewInboundCostService(costs, InboundCostLineDAOMock{}, InboundCostAdjustmentDAOMock{}, StockMovementDAOMock{}, CostLayerDAOMock{}, ItemResolverMock{}, PosterMock{}, inboundCostConfigSourceMock{}, vendorBillLineLookupMock{}, TransactionerMock{})

	if helper.AssertError(t, mustGetInboundCost(ctx, svc), true, ErrInboundCostNotFound) {
		return
	}
}

func TestInboundCostService_Post(t *testing.T) {
	ctx := context.Background()

	baseDraftCost := InboundCostDAOMock{
		CRUDMock: dao.CRUDMock[InboundCost]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*InboundCost, error) {
				return &InboundCost{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(10)), State: InboundCostStateDraft, TargetShipmentIDs: helper.Int64Array{30}}, nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, cost *InboundCost) (*InboundCost, error) {
			return cost, nil
		},
	}

	type postTestCase struct {
		name      string
		costs     InboundCostDAOMock
		lines     InboundCostLineDAOMock
		adjust    InboundCostAdjustmentDAOMock
		movements StockMovementDAOMock
		layers    CostLayerDAOMock
		resolver  ItemResolverMock
		poster    PosterMock
		config    inboundCostConfigSourceMock
		supplier  vendorBillLineLookupMock
		wantErr   error
	}

	var posted accounting.PostRequest
	var createdAdjustment *InboundCostAdjustment
	var updatedLayer *CostLayer
	var adjustments []*InboundCostAdjustment

	tests := []postTestCase{
		{
			"adjusts layers and posts dr stock cr clearing",
			baseDraftCost,
			InboundCostLineDAOMock{
				ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostLine, error) {
					return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11, Amount: 100, SplitMethod: SplitMethodQuantity, AccountID: helper.Ptr(uint64(700))}}, nil
				},
			},
			InboundCostAdjustmentDAOMock{
				CreateTxFunc: func(_ context.Context, _ *gorm.DB, adjustment *InboundCostAdjustment) (*InboundCostAdjustment, error) {
					createdAdjustment = adjustment
					return adjustment, nil
				},
			},
			StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
					return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 11, Qty: 10}}, nil
				},
			},
			CostLayerDAOMock{
				ListByMovementFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
					return []*CostLayer{{Base: model.Base{ID: 5}, RemainingQty: 10, Value: 100, RemainingValue: 100}}, nil
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, layer *CostLayer) (*CostLayer, error) {
					updatedLayer = layer
					return layer, nil
				},
			},
			ItemResolverMock{
				ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
					return ResolvedItem{StockAccounts: StockAccounts{StockValuationAccountID: 1100}}, nil
				},
			},
			PosterMock{
				PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = request
					return &accounting.JournalEntry{Base: model.Base{ID: 99}}, nil
				},
			},
			inboundCostConfigSourceMock{journal: 3},
			vendorBillLineLookupMock{},
			nil,
		},
		{
			"split by value absorbs rounding on last",
			InboundCostDAOMock{
				CRUDMock: dao.CRUDMock[InboundCost]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*InboundCost, error) {
						return &InboundCost{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(10)), State: InboundCostStateDraft, TargetShipmentIDs: helper.Int64Array{30}}, nil
					},
				},
			},
			InboundCostLineDAOMock{
				ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostLine, error) {
					return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11, Amount: 100, SplitMethod: SplitMethodValue, AccountID: helper.Ptr(uint64(700))}}, nil
				},
			},
			InboundCostAdjustmentDAOMock{
				CreateTxFunc: func(_ context.Context, _ *gorm.DB, adjustment *InboundCostAdjustment) (*InboundCostAdjustment, error) {
					adjustments = append(adjustments, adjustment)
					return adjustment, nil
				},
			},
			StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
					return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 11, Qty: 10}, {Base: model.Base{ID: 2}, ItemID: 11, Qty: 10}}, nil
				},
			},
			CostLayerDAOMock{
				ListByMovementFunc: func(_ context.Context, movementID uint64) ([]*CostLayer, error) {
					return []*CostLayer{{Base: model.Base{ID: 5 + movementID}, RemainingQty: 10, Value: 100, RemainingValue: 100}}, nil
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, layer *CostLayer) (*CostLayer, error) {
					return layer, nil
				},
			},
			ItemResolverMock{
				ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
					return ResolvedItem{StockAccounts: StockAccounts{StockValuationAccountID: 1100}}, nil
				},
			},
			PosterMock{},
			inboundCostConfigSourceMock{journal: 3},
			vendorBillLineLookupMock{},
			nil,
		},
		{
			"resolves clearing account from supplier bill line",
			baseDraftCost,
			InboundCostLineDAOMock{
				ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostLine, error) {
					return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11, Amount: 100, SplitMethod: SplitMethodQuantity, SupplierBillLineID: helper.Ptr(uint64(77))}}, nil
				},
			},
			InboundCostAdjustmentDAOMock{},
			StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
					return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 11, Qty: 10}}, nil
				},
			},
			CostLayerDAOMock{
				ListByMovementFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
					return []*CostLayer{{Base: model.Base{ID: 5}, RemainingQty: 10, Value: 100, RemainingValue: 100}}, nil
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, layer *CostLayer) (*CostLayer, error) {
					return layer, nil
				},
			},
			ItemResolverMock{
				ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
					return ResolvedItem{StockAccounts: StockAccounts{StockValuationAccountID: 1100}}, nil
				},
			},
			PosterMock{
				PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = request
					return &accounting.JournalEntry{Base: model.Base{ID: 99}}, nil
				},
			},
			inboundCostConfigSourceMock{journal: 3},
			vendorBillLineLookupMock{
				line: &accounting.InvoiceLine{Base: model.Base{ID: 77}, AccountID: helper.Ptr(uint64(700))},
			},
			nil,
		},
		{
			"posted state",
			InboundCostDAOMock{CRUDMock: dao.CRUDMock[InboundCost]{SearchFunc: func(_ context.Context, _ string, _ any) (*InboundCost, error) {
				return &InboundCost{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(10)), State: InboundCostStatePosted}, nil
			}}},
			InboundCostLineDAOMock{},
			InboundCostAdjustmentDAOMock{},
			StockMovementDAOMock{},
			CostLayerDAOMock{},
			ItemResolverMock{},
			PosterMock{},
			inboundCostConfigSourceMock{},
			vendorBillLineLookupMock{},
			ErrInboundCostState,
		},
		{
			"no lines",
			InboundCostDAOMock{CRUDMock: dao.CRUDMock[InboundCost]{SearchFunc: func(_ context.Context, _ string, _ any) (*InboundCost, error) {
				return &InboundCost{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(10)), State: InboundCostStateDraft}, nil
			}}},
			InboundCostLineDAOMock{},
			InboundCostAdjustmentDAOMock{},
			StockMovementDAOMock{},
			CostLayerDAOMock{},
			ItemResolverMock{},
			PosterMock{},
			inboundCostConfigSourceMock{},
			vendorBillLineLookupMock{},
			ErrInboundCostNoLines,
		},
		{
			"no journal",
			InboundCostDAOMock{CRUDMock: dao.CRUDMock[InboundCost]{SearchFunc: func(_ context.Context, _ string, _ any) (*InboundCost, error) {
				return &InboundCost{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(10)), State: InboundCostStateDraft, TargetShipmentIDs: helper.Int64Array{30}}, nil
			}}},
			InboundCostLineDAOMock{ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostLine, error) {
				return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11, AccountID: helper.Ptr(uint64(700))}}, nil
			}},
			InboundCostAdjustmentDAOMock{},
			StockMovementDAOMock{},
			CostLayerDAOMock{},
			ItemResolverMock{},
			PosterMock{},
			inboundCostConfigSourceMock{},
			vendorBillLineLookupMock{},
			ErrInboundCostNoJournal,
		},
		{
			"rejects without matching movements",
			baseDraftCost,
			InboundCostLineDAOMock{
				ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostLine, error) {
					return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11, Amount: 100, AccountID: helper.Ptr(uint64(700))}}, nil
				},
			},
			InboundCostAdjustmentDAOMock{},
			StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
					return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 99, Qty: 10}}, nil
				},
			},
			CostLayerDAOMock{},
			ItemResolverMock{},
			PosterMock{},
			inboundCostConfigSourceMock{journal: 3},
			vendorBillLineLookupMock{},
			ErrInboundCostNoMovements,
		},
		{
			"rejects line without clearing account",
			baseDraftCost,
			InboundCostLineDAOMock{
				ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostLine, error) {
					return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11, Amount: 100, SupplierBillLineID: helper.Ptr(uint64(77))}}, nil
				},
			},
			InboundCostAdjustmentDAOMock{},
			StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
					return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 11, Qty: 10}}, nil
				},
			},
			CostLayerDAOMock{},
			ItemResolverMock{},
			PosterMock{},
			inboundCostConfigSourceMock{journal: 3},
			vendorBillLineLookupMock{line: &accounting.InvoiceLine{}},
			ErrInboundCostNoAccount,
		},
		{
			"propagates supplier bill line error",
			baseDraftCost,
			InboundCostLineDAOMock{
				ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostLine, error) {
					return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11, Amount: 100, SupplierBillLineID: helper.Ptr(uint64(77))}}, nil
				},
			},
			InboundCostAdjustmentDAOMock{},
			StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
					return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 11, Qty: 10}}, nil
				},
			},
			CostLayerDAOMock{},
			ItemResolverMock{},
			PosterMock{},
			inboundCostConfigSourceMock{journal: 3},
			vendorBillLineLookupMock{
				line:    &accounting.InvoiceLine{},
				findErr: ErrInboundCostNotFound,
			},
			ErrInboundCostNotFound,
		},
		{
			"rejects when no open layer",
			baseDraftCost,
			InboundCostLineDAOMock{
				ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostLine, error) {
					return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11, Amount: 100, AccountID: helper.Ptr(uint64(700))}}, nil
				},
			},
			InboundCostAdjustmentDAOMock{},
			StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
					return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 11, Qty: 10}}, nil
				},
			},
			CostLayerDAOMock{
				ListByMovementFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
					return []*CostLayer{{Base: model.Base{ID: 5}, RemainingQty: 0}}, nil
				},
			},
			ItemResolverMock{
				ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
					return ResolvedItem{StockAccounts: StockAccounts{StockValuationAccountID: 1100}}, nil
				},
			},
			PosterMock{},
			inboundCostConfigSourceMock{journal: 3},
			vendorBillLineLookupMock{},
			ErrInboundCostValue,
		},
		{
			"rejects when shipment movements fail",
			baseDraftCost,
			InboundCostLineDAOMock{
				ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostLine, error) {
					return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11, Amount: 100, AccountID: helper.Ptr(uint64(700))}}, nil
				},
			},
			InboundCostAdjustmentDAOMock{},
			StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
					return nil, ErrMovementNotFound
				},
			},
			CostLayerDAOMock{},
			ItemResolverMock{},
			PosterMock{},
			inboundCostConfigSourceMock{journal: 3},
			vendorBillLineLookupMock{},
			ErrMovementNotFound,
		},
		{
			"propagates layer list error",
			baseDraftCost,
			InboundCostLineDAOMock{
				ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostLine, error) {
					return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11, Amount: 100, AccountID: helper.Ptr(uint64(700))}}, nil
				},
			},
			InboundCostAdjustmentDAOMock{},
			StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
					return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 11, Qty: 10}}, nil
				},
			},
			CostLayerDAOMock{
				ListByMovementFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
					return nil, ErrInboundCostValue
				},
			},
			ItemResolverMock{
				ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
					return ResolvedItem{StockAccounts: StockAccounts{StockValuationAccountID: 1100}}, nil
				},
			},
			PosterMock{},
			inboundCostConfigSourceMock{journal: 3},
			vendorBillLineLookupMock{},
			ErrInboundCostValue,
		},
		{
			"rejects missing valuation account",
			baseDraftCost,
			InboundCostLineDAOMock{
				ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostLine, error) {
					return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11, Amount: 100, AccountID: helper.Ptr(uint64(700))}}, nil
				},
			},
			InboundCostAdjustmentDAOMock{},
			StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
					return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 11, Qty: 10}}, nil
				},
			},
			CostLayerDAOMock{
				ListByMovementFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
					return []*CostLayer{{Base: model.Base{ID: 5}, RemainingQty: 10}}, nil
				},
			},
			ItemResolverMock{
				ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
					return ResolvedItem{}, nil
				},
			},
			PosterMock{},
			inboundCostConfigSourceMock{journal: 3},
			vendorBillLineLookupMock{},
			ErrValuationAccount,
		},
		{
			"propagates poster error",
			baseDraftCost,
			InboundCostLineDAOMock{
				ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostLine, error) {
					return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11, Amount: 100, SplitMethod: SplitMethodQuantity, AccountID: helper.Ptr(uint64(700))}}, nil
				},
			},
			InboundCostAdjustmentDAOMock{},
			StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
					return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 11, Qty: 10}}, nil
				},
			},
			CostLayerDAOMock{
				ListByMovementFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
					return []*CostLayer{{Base: model.Base{ID: 5}, RemainingQty: 10, Value: 100}}, nil
				},
			},
			ItemResolverMock{
				ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
					return ResolvedItem{StockAccounts: StockAccounts{StockValuationAccountID: 1100}}, nil
				},
			},
			PosterMock{
				PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return nil, ErrInboundCostNoJournal
				},
			},
			inboundCostConfigSourceMock{journal: 3},
			vendorBillLineLookupMock{},
			ErrInboundCostNoJournal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewInboundCostService(tt.costs, tt.lines, tt.adjust, tt.movements, tt.layers, tt.resolver, tt.poster, tt.config, tt.supplier, TransactionerMock{})
			postedCost, err := svc.Post(ctx, 10, 9)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			switch tt.name {
			case "adjusts layers and posts dr stock cr clearing":
				if postedCost.State != InboundCostStatePosted || postedCost.MovementID == nil || *postedCost.MovementID != 99 {
					t.Errorf("cost = %+v, want posted with movement 99", postedCost)
				}
				if createdAdjustment == nil || createdAdjustment.AdditionalCost != 100 || createdAdjustment.StockMovementID != 1 {
					t.Errorf("adjustment = %+v, want +100 on movement 1", createdAdjustment)
				}
				if updatedLayer == nil || updatedLayer.RemainingValue != 200 {
					t.Errorf("layer = %+v, want remaining value 200", updatedLayer)
				}
				if posted.JournalID != 3 || len(posted.Lines) != 2 {
					t.Fatalf("posted = %+v, want 2 lines on journal 3", posted)
				}
				if !posted.Lines[0].Credit.Equal(amount.FromFloat64(100)) || posted.Lines[0].AccountID != 700 {
					t.Errorf("clearing line = %+v, want Cr 700 of 100", posted.Lines[0])
				}
				if !posted.Lines[1].Debit.Equal(amount.FromFloat64(100)) || posted.Lines[1].AccountID != 1100 {
					t.Errorf("stock line = %+v, want Dr 1100 of 100", posted.Lines[1])
				}
			case "split by value absorbs rounding on last":
				if len(adjustments) != 2 {
					t.Fatalf("adjustments = %d, want 2", len(adjustments))
				}
				if adjustments[0].AdditionalCost != adjustments[1].AdditionalCost {
					t.Errorf("movements of equal value should split equally, got %v vs %v", adjustments[0].AdditionalCost, adjustments[1].AdditionalCost)
				}
				if total := adjustments[0].AdditionalCost + adjustments[1].AdditionalCost; total != 100 {
					t.Errorf("total adjusted = %v, want 100", total)
				}
			case "resolves clearing account from supplier bill line":
				if postedCost.State != InboundCostStatePosted {
					t.Errorf("cost = %+v, want posted", postedCost)
				}
				if len(posted.Lines) != 2 {
					t.Fatalf("posted = %+v, want 2 lines", posted.Lines)
				}
				if posted.Lines[0].AccountID != 700 || !posted.Lines[0].Credit.Equal(amount.FromFloat64(100)) {
					t.Errorf("clearing line = %+v, want Cr 700 of 100 resolved from the supplier bill line", posted.Lines[0])
				}
			}
		})
	}
}

func TestInboundCostService_List(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		costs   InboundCostDAOMock
		want    int
		wantID  uint64
		wantErr error
	}{
		{
			"returns costs scoped to organization",
			InboundCostDAOMock{
				CRUDMock: dao.CRUDMock[InboundCost]{
					ListFunc: func(_ context.Context, q *query.Query) (*query.Page[InboundCost], error) {
						if len(q.Filters) == 0 || q.Filters[0].Field != "organization_id" {
							t.Fatalf("filter = %+v, want organization_id", q.Filters)
						}
						return &query.Page[InboundCost]{Items: []*InboundCost{
							{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(10))},
						}}, nil
					},
				},
			},
			1, 9, nil,
		},
		{
			"propagates error",
			InboundCostDAOMock{
				CRUDMock: dao.CRUDMock[InboundCost]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[InboundCost], error) {
						return nil, ErrInboundCostNotFound
					},
				},
			},
			0, 0, ErrInboundCostNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewInboundCostService(tt.costs, InboundCostLineDAOMock{}, InboundCostAdjustmentDAOMock{}, StockMovementDAOMock{}, CostLayerDAOMock{}, ItemResolverMock{}, PosterMock{}, inboundCostConfigSourceMock{}, vendorBillLineLookupMock{}, TransactionerMock{})
			items, err := svc.List(ctx, 10)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if len(items) != tt.want || (tt.want > 0 && items[0].ID != tt.wantID) {
				t.Errorf("items = %+v, want %d items with id %d", items, tt.want, tt.wantID)
			}
		})
	}
}

func TestInboundCostService_ListLines(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		costs   InboundCostDAOMock
		lines   InboundCostLineDAOMock
		want    int
		wantPID uint64
		wantErr error
	}{
		{
			"returns lines after ownership check",
			InboundCostDAOMock{
				CRUDMock: dao.CRUDMock[InboundCost]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*InboundCost, error) {
						return &InboundCost{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(10))}, nil
					},
				},
			},
			InboundCostLineDAOMock{
				ListByInboundCostFunc: func(_ context.Context, costID uint64) ([]*InboundCostLine, error) {
					if costID != 9 {
						t.Fatalf("cost id = %d, want 9", costID)
					}
					return []*InboundCostLine{{Base: model.Base{ID: 1}, ItemID: 11}}, nil
				},
			},
			1, 11, nil,
		},
		{
			"rejects other organization",
			InboundCostDAOMock{
				CRUDMock: dao.CRUDMock[InboundCost]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*InboundCost, error) {
						return &InboundCost{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(99))}, nil
					},
				},
			},
			InboundCostLineDAOMock{},
			0, 0, ErrInboundCostNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewInboundCostService(tt.costs, tt.lines, InboundCostAdjustmentDAOMock{}, StockMovementDAOMock{}, CostLayerDAOMock{}, ItemResolverMock{}, PosterMock{}, inboundCostConfigSourceMock{}, vendorBillLineLookupMock{}, TransactionerMock{})
			items, err := svc.ListLines(ctx, 10, 9)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if len(items) != tt.want || (tt.want > 0 && items[0].ItemID != tt.wantPID) {
				t.Errorf("lines = %+v, want %d lines with item %d", items, tt.want, tt.wantPID)
			}
		})
	}
}

func TestInboundCostService_ListAdjustments(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		costs    InboundCostDAOMock
		adjust   InboundCostAdjustmentDAOMock
		want     int
		wantCost float64
		wantErr  error
	}{
		{
			"returns adjustments",
			InboundCostDAOMock{
				CRUDMock: dao.CRUDMock[InboundCost]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*InboundCost, error) {
						return &InboundCost{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(10))}, nil
					},
				},
			},
			InboundCostAdjustmentDAOMock{
				ListByInboundCostFunc: func(_ context.Context, costID uint64) ([]*InboundCostAdjustment, error) {
					return []*InboundCostAdjustment{{Base: model.Base{ID: 1}, StockMovementID: 5, AdditionalCost: 25}}, nil
				},
			},
			1, 25, nil,
		},
		{
			"propagates list error",
			InboundCostDAOMock{
				CRUDMock: dao.CRUDMock[InboundCost]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*InboundCost, error) {
						return &InboundCost{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(10))}, nil
					},
				},
			},
			InboundCostAdjustmentDAOMock{
				ListByInboundCostFunc: func(_ context.Context, _ uint64) ([]*InboundCostAdjustment, error) {
					return nil, ErrInboundCostNotFound
				},
			},
			0, 0, ErrInboundCostNotFound,
		},
		{
			"rejects missing cost",
			InboundCostDAOMock{
				CRUDMock: dao.CRUDMock[InboundCost]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*InboundCost, error) {
						return nil, nil
					},
				},
			},
			InboundCostAdjustmentDAOMock{},
			0, 0, ErrInboundCostNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewInboundCostService(tt.costs, InboundCostLineDAOMock{}, tt.adjust, StockMovementDAOMock{}, CostLayerDAOMock{}, ItemResolverMock{}, PosterMock{}, inboundCostConfigSourceMock{}, vendorBillLineLookupMock{}, TransactionerMock{})
			items, err := svc.ListAdjustments(ctx, 10, 9)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if len(items) != tt.want || (tt.want > 0 && items[0].AdditionalCost != tt.wantCost) {
				t.Errorf("adjustments = %+v, want %d adjustments of %v", items, tt.want, tt.wantCost)
			}
		})
	}
}

func TestInboundCostService_moveBasis(t *testing.T) {
	svc := NewInboundCostService(InboundCostDAOMock{}, InboundCostLineDAOMock{}, InboundCostAdjustmentDAOMock{}, StockMovementDAOMock{}, CostLayerDAOMock{}, ItemResolverMock{}, PosterMock{}, inboundCostConfigSourceMock{}, vendorBillLineLookupMock{}, TransactionerMock{})
	movement := &StockMovement{Qty: 4}
	layer := &CostLayer{Value: 60}

	tests := []struct {
		name     string
		method   string
		resolved ResolvedItem
		layer    *CostLayer
		want     float64
		wantErr  error
	}{
		{"weight", SplitMethodWeight, ResolvedItem{Weight: 2}, nil, 8, nil},
		{"volume", SplitMethodVolume, ResolvedItem{Volume: 3}, nil, 12, nil},
		{"value", SplitMethodValue, ResolvedItem{}, layer, 60, nil},
		{"equal", SplitMethodEqual, ResolvedItem{}, nil, 1, nil},
		{"unknown", "unknown", ResolvedItem{}, nil, 0, ErrInboundCostSplit},
		{"value requires layer", SplitMethodValue, ResolvedItem{}, nil, 0, ErrInboundCostValue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.moveBasis(tt.method, movement, tt.layer, tt.resolved)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if got.Float64() != tt.want {
				t.Errorf("moveBasis(%s) = %v, want %v", tt.method, got.Float64(), tt.want)
			}
		})
	}
}

func TestInboundCostService_inboundCostLineName(t *testing.T) {
	tests := []struct {
		name        string
		line        *InboundCostLine
		description string
	}{
		{"falls back when no description", &InboundCostLine{}, "Landed cost"},
		{"uses description", &InboundCostLine{Description: helper.Ptr("Ocean freight")}, "Ocean freight"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inboundCostLineName(tt.line); got != tt.description {
				t.Errorf("inboundCostLineName() = %s, want %s", got, tt.description)
			}
		})
	}
}

func mustGetInboundCost(ctx context.Context, svc InboundCostService) (err error) {
	_, err = svc.Get(ctx, 10, 9)
	return err
}

type inboundCostConfigSourceMock struct {
	journal uint64
}

func (m inboundCostConfigSourceMock) JournalID(ctx context.Context, organizationID uint64) (uint64, error) {
	return m.journal, nil
}

type vendorBillLineLookupMock struct {
	line    *accounting.InvoiceLine
	findErr error
}

func (m vendorBillLineLookupMock) Find(ctx context.Context, id uint64) (*accounting.InvoiceLine, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.line, nil
}
