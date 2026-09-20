package inventory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestTransferService_Create_RejectsSameWarehouse(t *testing.T) {
	ctx := context.Background()
	svc := NewTransferService(WarehouseTransferDAOMock{}, StockMovementDAOMock{}, StockLocationDAOMock{}, WarehouseDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Create(ctx, &WarehouseTransfer{SrcWarehouseID: 1, DstWarehouseID: 1}, nil)
	if helper.AssertError(t, err, true, ErrSameWarehouse) {
		return
	}
}

func transferLocationDAO(organizationID *uint64) StockLocationDAOMock {
	return StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: 10}, OrganizationID: organizationID}, nil
			},
		},
	}
}

func TestTransferService_Create_RejectsWithoutTransitLocation(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	warehouses := WarehouseDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return &reference.Warehouse{Name: "WH", OrganizationID: &organizationID}, nil
		}),
	}
	svc := NewTransferService(WarehouseTransferDAOMock{}, StockMovementDAOMock{}, StockLocationDAOMock{}, warehouses, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Create(ctx, &WarehouseTransfer{OrganizationID: &organizationID, SrcWarehouseID: 1, DstWarehouseID: 2}, nil)
	if helper.AssertError(t, err, true, ErrTransitNotFound) {
		return
	}
}

func TestTransferService_Create_CreatesOutboundAndInboundMoves(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	transfers := WarehouseTransferDAOMock{
		CreateWithShipmentsFunc: func(_ context.Context, transfer *WarehouseTransfer, outShipment *Shipment, outMovements []*StockMovement, inShipment *Shipment, inMovements []*StockMovement) (*WarehouseTransfer, error) {
			transfer.ID = 1
			return transfer, nil
		},
	}
	warehouses := WarehouseDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, id uint64) (*reference.Warehouse, error) {
			return &reference.Warehouse{Base: model.Base{ID: id}, Name: "WH", OrganizationID: &organizationID}, nil
		}),
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "usage" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, Usage: "transit"}}}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: &organizationID, Usage: "internal"}}}, nil
			},
		},
	}
	svc := NewTransferService(transfers, StockMovementDAOMock{}, locations, warehouses, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, locations, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	transfer, err := svc.Create(ctx, &WarehouseTransfer{
		OrganizationID: &organizationID, SrcWarehouseID: 1, DstWarehouseID: 2, Name: helper.Ptr("TRF-1"),
	}, []TransferLine{{ItemID: 100, Qty: 5}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transfer.ID != 1 {
		t.Errorf("transfer id = %d, want 1", transfer.ID)
	}
}

func TestTransferService_Send_PostsGoodsInTransit(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	outShipmentID := uint64(5)
	transfer := &WarehouseTransfer{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		SrcWarehouseID: 1, DstWarehouseID: 2, State: TransferStateDraft, OutShipmentID: &outShipmentID,
	}
	transfers := WarehouseTransferDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
		return transfer, nil
	})}
	outMove := &StockMovement{
		Base: model.Base{ID: 10}, OrganizationID: &organizationID,
		ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 90, State: MovementStateDraft,
	}
	movements := StockMovementDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) { return outMove, nil }),
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{outMove}, nil
		},
	}
	quants := StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
			return &StockBalance{ItemID: 100, Quantity: 5}, nil
		},
	}
	layers := CostLayerDAOMock{
		ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
			return []*CostLayer{
				{Base: model.Base{ID: 1}, RemainingQty: 5, RemainingValue: 100, UnitCost: helper.Ptr(20.0)},
			}, nil
		},
	}
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
		},
	}
	var posted []accounting.PostRequest
	poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
		posted = append(posted, request)
		return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
	}}
	ledger := NewLedgerService(movements, quants, transferLocationDAO(&organizationID), TransactionerMock{})
	svc := NewTransferService(transfers, movements, StockLocationDAOMock{}, WarehouseDAOMock{}, layers, ledger, resolver, poster, TransactionerMock{})

	if err := svc.Send(ctx, 1, 50, 1500, time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transfer.State != TransferStateInTransit {
		t.Errorf("state = %s, want in_transit", transfer.State)
	}
	if len(posted) != 1 {
		t.Fatalf("postings = %d, want 1", len(posted))
	}
	if got := posted[0].Lines[0]; got.AccountID != 1500 || got.Debit.Float64() != 100 {
		t.Errorf("line 0 = %+v, want goods in transit debit 100", got)
	}
	if got := posted[0].Lines[1]; got.AccountID != 1300 || got.Credit.Float64() != 100 {
		t.Errorf("line 1 = %+v, want inventory credit 100", got)
	}
}

func TestTransferService_Send_PostFailure_DoesNotPersistState(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	outShipmentID := uint64(5)
	transfer := &WarehouseTransfer{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		SrcWarehouseID: 1, DstWarehouseID: 2, State: TransferStateDraft, OutShipmentID: &outShipmentID,
	}
	stateSaved := false
	transfers := WarehouseTransferDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
			return transfer, nil
		}),
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *WarehouseTransfer) (*WarehouseTransfer, error) {
			stateSaved = true
			return transfer, nil
		},
	}
	outMove := &StockMovement{
		Base: model.Base{ID: 10}, OrganizationID: &organizationID,
		ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 90, State: MovementStateDraft,
	}
	movements := StockMovementDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) { return outMove, nil }),
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{outMove}, nil
		},
	}
	quants := StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
			return &StockBalance{ItemID: 100, Quantity: 5}, nil
		},
	}
	layers := CostLayerDAOMock{
		ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
			return []*CostLayer{
				{Base: model.Base{ID: 1}, RemainingQty: 5, RemainingValue: 100, UnitCost: helper.Ptr(20.0)},
			}, nil
		},
	}
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
		},
	}
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return nil, errors.New("unbalanced posting")
		},
	}
	svc := NewTransferService(transfers, movements, StockLocationDAOMock{}, WarehouseDAOMock{}, layers, NewLedgerService(movements, quants, transferLocationDAO(&organizationID), TransactionerMock{}), resolver, poster, TransactionerMock{})

	err := svc.Send(ctx, 1, 50, 1500, time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("expected error when posting fails")
	}
	if stateSaved {
		t.Error("transfer state must not be persisted when posting fails inside the transaction")
	}
}

func TestTransferService_Send_RejectsInsufficientStock(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	outShipmentID := uint64(5)
	transfers := WarehouseTransferDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
		return &WarehouseTransfer{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: TransferStateDraft, OutShipmentID: &outShipmentID}, nil
	})}
	outMove := &StockMovement{Base: model.Base{ID: 10}, ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 90, State: MovementStateDraft}
	movements := StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{outMove}, nil
		},
	}
	quants := StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
			return &StockBalance{ItemID: 100, Quantity: 3}, nil
		},
	}
	svc := NewTransferService(transfers, movements, StockLocationDAOMock{}, WarehouseDAOMock{}, CostLayerDAOMock{}, NewLedgerService(movements, quants, transferLocationDAO(&organizationID), TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	err := svc.Send(ctx, 1, 50, 1500, time.Now())
	if helper.AssertError(t, err, true, ErrInsufficientStock) {
		return
	}
}

func TestTransferService_Receive_RejectsWhenNotInTransit(t *testing.T) {
	ctx := context.Background()
	transfers := WarehouseTransferDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
		return &WarehouseTransfer{Base: model.Base{ID: 1}, State: TransferStateDraft}, nil
	})}
	svc := NewTransferService(transfers, StockMovementDAOMock{}, StockLocationDAOMock{}, WarehouseDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	err := svc.Receive(ctx, 1, 50, 1500, time.Now())
	if helper.AssertError(t, err, true, ErrWarehouseTransferState) {
		return
	}
}

func TestTransferService_Receive_PostsInventoryInbound(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	inShipmentID := uint64(6)
	transfer := &WarehouseTransfer{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		SrcWarehouseID: 1, DstWarehouseID: 2, State: TransferStateInTransit, InShipmentID: &inShipmentID,
	}
	transfers := WarehouseTransferDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
		return transfer, nil
	})}
	inMove := &StockMovement{
		Base: model.Base{ID: 11}, OrganizationID: &organizationID,
		ItemID: 100, Qty: 5, SrcLocationID: 90, DstLocationID: 10, State: MovementStateDraft,
	}
	movements := StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{inMove}, nil
		},
	}
	layers := CostLayerDAOMock{
		ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
			return []*CostLayer{
				{Base: model.Base{ID: 1}, RemainingQty: 5, RemainingValue: 100, UnitCost: helper.Ptr(20.0)},
			}, nil
		},
	}
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
		},
	}
	var posted []accounting.PostRequest
	poster := PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
		posted = append(posted, request)
		return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
	}}
	svc := NewTransferService(transfers, movements, StockLocationDAOMock{}, WarehouseDAOMock{}, layers, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), resolver, poster, TransactionerMock{})

	if err := svc.Receive(ctx, 1, 50, 1500, time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transfer.State != TransferStateReceived {
		t.Errorf("state = %s, want received", transfer.State)
	}
	if len(posted) != 1 {
		t.Fatalf("postings = %d, want 1", len(posted))
	}
	if got := posted[0].Lines[0]; got.AccountID != 1300 || got.Debit.Float64() != 100 {
		t.Errorf("line 0 = %+v, want inventory debit 100 on account 1300", got)
	}
	if got := posted[0].Lines[1]; got.AccountID != 1500 || got.Credit.Float64() != 100 {
		t.Errorf("line 1 = %+v, want transit credit 100 on account 1500", got)
	}
}

func TestTransferService_Receive_SkipsPostingWhenNoValue(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	inShipmentID := uint64(6)
	transfer := &WarehouseTransfer{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		State: TransferStateInTransit, InShipmentID: &inShipmentID,
	}
	transfers := WarehouseTransferDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
		return transfer, nil
	})}
	movements := StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{{Base: model.Base{ID: 11}, ItemID: 100, Qty: 5}}, nil
		},
	}
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
		},
	}
	posted := false
	poster := PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
		posted = true
		return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
	}}
	svc := NewTransferService(transfers, movements, StockLocationDAOMock{}, WarehouseDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), resolver, poster, TransactionerMock{})

	if err := svc.Receive(ctx, 1, 50, 1500, time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transfer.State != TransferStateReceived {
		t.Errorf("state = %s, want received", transfer.State)
	}
	if posted {
		t.Error("no posting expected when the movement has no valuation value")
	}
}

func TestTransferService_Receive_NotFound(t *testing.T) {
	ctx := context.Background()
	transfers := WarehouseTransferDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
		return nil, nil
	})}
	svc := NewTransferService(transfers, StockMovementDAOMock{}, StockLocationDAOMock{}, WarehouseDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	err := svc.Receive(ctx, 1, 50, 1500, time.Now())
	if helper.AssertError(t, err, true, ErrWarehouseTransferNotFound) {
		return
	}
}

func TestTransferService_Receive_RejectsMissingInShipment(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	transfers := WarehouseTransferDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
		return &WarehouseTransfer{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: TransferStateInTransit}, nil
	})}
	svc := NewTransferService(transfers, StockMovementDAOMock{}, StockLocationDAOMock{}, WarehouseDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	err := svc.Receive(ctx, 1, 50, 1500, time.Now())
	if helper.AssertError(t, err, true, ErrWarehouseTransferState) {
		return
	}
}

func TestTransferService_Receive_RejectsMissingTransitAccount(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	inShipmentID := uint64(6)
	transfers := WarehouseTransferDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
		return &WarehouseTransfer{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: TransferStateInTransit, InShipmentID: &inShipmentID}, nil
	})}
	svc := NewTransferService(transfers, StockMovementDAOMock{}, StockLocationDAOMock{}, WarehouseDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	err := svc.Receive(ctx, 1, 50, 0, time.Now())
	if helper.AssertError(t, err, true, ErrTransitAccount) {
		return
	}
}

func TestTransferService_Receive_RejectsMissingOrganization(t *testing.T) {
	ctx := context.Background()
	inShipmentID := uint64(6)
	transfers := WarehouseTransferDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
		return &WarehouseTransfer{Base: model.Base{ID: 1}, State: TransferStateInTransit, InShipmentID: &inShipmentID}, nil
	})}
	svc := NewTransferService(transfers, StockMovementDAOMock{}, StockLocationDAOMock{}, WarehouseDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	err := svc.Receive(ctx, 1, 50, 1500, time.Now())
	if helper.AssertError(t, err, true, ErrOrganizationMissing) {
		return
	}
}

func TestTransferService_Receive_RejectsWhenListMovesFails(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	inShipmentID := uint64(6)
	transfers := WarehouseTransferDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
		return &WarehouseTransfer{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: TransferStateInTransit, InShipmentID: &inShipmentID}, nil
	})}
	movements := StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return nil, ErrMovementNotFound
		},
	}
	svc := NewTransferService(transfers, movements, StockLocationDAOMock{}, WarehouseDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	err := svc.Receive(ctx, 1, 50, 1500, time.Now())
	if helper.AssertError(t, err, true, ErrMovementNotFound) {
		return
	}
}

func TestTransferService_Receive_RejectsMissingValuationAccount(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	inShipmentID := uint64(6)
	transfers := WarehouseTransferDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
		return &WarehouseTransfer{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: TransferStateInTransit, InShipmentID: &inShipmentID}, nil
	})}
	movements := StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{{Base: model.Base{ID: 11}, ItemID: 100, Qty: 5}}, nil
		},
	}
	svc := NewTransferService(transfers, movements, StockLocationDAOMock{}, WarehouseDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	err := svc.Receive(ctx, 1, 50, 1500, time.Now())
	if helper.AssertError(t, err, true, ErrValuationAccount) {
		return
	}
}

func TestTransferService_Receive_PostFailure_DoesNotPersistState(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	inShipmentID := uint64(6)
	transfer := &WarehouseTransfer{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		State: TransferStateInTransit, InShipmentID: &inShipmentID,
	}
	stateSaved := false
	transfers := WarehouseTransferDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
			return transfer, nil
		}),
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *WarehouseTransfer) (*WarehouseTransfer, error) {
			stateSaved = true
			return transfer, nil
		},
	}
	movements := StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{{Base: model.Base{ID: 11}, ItemID: 100, Qty: 5}}, nil
		},
	}
	layers := CostLayerDAOMock{
		ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
			return []*CostLayer{{Base: model.Base{ID: 1}, RemainingQty: 5, RemainingValue: 100, UnitCost: helper.Ptr(20.0)}}, nil
		},
	}
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
		},
	}
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return nil, errors.New("unbalanced posting")
		},
	}
	svc := NewTransferService(transfers, movements, StockLocationDAOMock{}, WarehouseDAOMock{}, layers, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), resolver, poster, TransactionerMock{})

	err := svc.Receive(ctx, 1, 50, 1500, time.Now())
	if err == nil {
		t.Fatal("expected error when posting fails")
	}
	if stateSaved {
		t.Error("transfer state must not be persisted when posting fails inside the transaction")
	}
}

func TestTransferService_Receive_RejectsWhenApplyFails(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	inShipmentID := uint64(6)
	transfers := WarehouseTransferDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*WarehouseTransfer, error) {
		return &WarehouseTransfer{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: TransferStateInTransit, InShipmentID: &inShipmentID}, nil
	})}
	movements := StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{{Base: model.Base{ID: 11}, ItemID: 100, Qty: 5}}, nil
		},
		FindForUpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ uint64) (*StockMovement, error) {
			return &StockMovement{Base: model.Base{ID: 11}, ItemID: 100, Qty: 5, State: MovementStateDone}, nil
		},
	}
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
		},
	}
	svc := NewTransferService(transfers, movements, StockLocationDAOMock{}, WarehouseDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), resolver, PosterMock{}, TransactionerMock{})

	err := svc.Receive(ctx, 1, 50, 1500, time.Now())
	if helper.AssertError(t, err, true, ErrMovementState) {
		return
	}
}

func TestTransferService_Create_UsesExplicitLineLocations(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	transfers := WarehouseTransferDAOMock{
		CreateWithShipmentsFunc: func(_ context.Context, transfer *WarehouseTransfer, outShipment *Shipment, outMovements []*StockMovement, inShipment *Shipment, inMovements []*StockMovement) (*WarehouseTransfer, error) {
			if len(outMovements) != 1 || len(inMovements) != 1 {
				t.Errorf("movements = out %d in %d, want 1/1", len(outMovements), len(inMovements))
			}
			if outMovements[0].SrcLocationID != 55 || outMovements[0].DstLocationID != 90 {
				t.Errorf("out movement locations = %d->%d, want 55->90 (transit)", outMovements[0].SrcLocationID, outMovements[0].DstLocationID)
			}
			if inMovements[0].SrcLocationID != 90 || inMovements[0].DstLocationID != 56 {
				t.Errorf("in movement locations = %d->%d, want 90 (transit)->56", inMovements[0].SrcLocationID, inMovements[0].DstLocationID)
			}
			return transfer, nil
		},
	}
	warehouses := WarehouseDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, id uint64) (*reference.Warehouse, error) {
			return &reference.Warehouse{Base: model.Base{ID: id}, Name: "WH", OrganizationID: &organizationID}, nil
		}),
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, Usage: "transit"}}}, nil
			},
			FindFunc: func(_ context.Context, id uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: id}, OrganizationID: &organizationID, Usage: "internal"}, nil
			},
		},
	}
	svc := NewTransferService(transfers, StockMovementDAOMock{}, locations, warehouses, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, locations, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Create(ctx, &WarehouseTransfer{
		OrganizationID: &organizationID, SrcWarehouseID: 1, DstWarehouseID: 2, Name: helper.Ptr("TRF-2"),
	}, []TransferLine{{ItemID: 100, Qty: 5, SrcLocationID: helper.Ptr(uint64(55)), DstLocationID: helper.Ptr(uint64(56))}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransferService_Create_RejectsForeignExplicitLineLocation(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	foreignOrg := uint64(2)
	warehouses := WarehouseDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, id uint64) (*reference.Warehouse, error) {
			return &reference.Warehouse{Base: model.Base{ID: id}, Name: "WH", OrganizationID: &organizationID}, nil
		}),
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, Usage: "transit"}}}, nil
			},
			FindFunc: func(_ context.Context, id uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: id}, OrganizationID: &foreignOrg, Usage: "internal"}, nil
			},
		},
	}
	svc := NewTransferService(WarehouseTransferDAOMock{}, StockMovementDAOMock{}, locations, warehouses, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, locations, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Create(ctx, &WarehouseTransfer{
		OrganizationID: &organizationID, SrcWarehouseID: 1, DstWarehouseID: 2,
	}, []TransferLine{{ItemID: 100, Qty: 5, SrcLocationID: helper.Ptr(uint64(55))}})
	if helper.AssertError(t, err, true, ErrLocationNotFound) {
		return
	}
}

func TestTransferService_Create_RejectsForeignWarehouse(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	foreignOrg := uint64(2)
	warehouses := WarehouseDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return &reference.Warehouse{Name: "WH", OrganizationID: &foreignOrg}, nil
		}),
	}
	svc := NewTransferService(WarehouseTransferDAOMock{}, StockMovementDAOMock{}, StockLocationDAOMock{}, warehouses, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Create(ctx, &WarehouseTransfer{OrganizationID: &organizationID, SrcWarehouseID: 1, DstWarehouseID: 2}, nil)
	if helper.AssertError(t, err, true, ErrWarehouseNotFound) {
		return
	}
}
