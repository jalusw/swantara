package inventory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func transferTestService(warehouses WarehouseDAOMock) TransferService {
	return transferTestServiceWithLocations(warehouses, StockLocationDAOMock{})
}

func transferTestServiceWithLocations(warehouses WarehouseDAOMock, locations StockLocationDAOMock) TransferService {
	return NewTransferService(WarehouseTransferDAOMock{}, StockMovementDAOMock{}, locations, warehouses, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})
}

func TestTransferService_Create_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	orgID := uint64(10)

	warehouses := func(srcErr error, src *reference.Warehouse, dstErr error, dst *reference.Warehouse) WarehouseDAOMock {
		calls := 0
		return WarehouseDAOMock{
			CRUDMock: dao.CRUDMock[reference.Warehouse]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
					calls++
					if calls == 1 {
						return src, srcErr
					}
					return dst, dstErr
				},
			},
		}
	}
	owned := &reference.Warehouse{Base: model.Base{ID: 1}, OrganizationID: &orgID}
	transfer := func() *WarehouseTransfer {
		return &WarehouseTransfer{OrganizationID: &orgID, SrcWarehouseID: 1, DstWarehouseID: 2}
	}

	t.Run("missing org", func(t *testing.T) {
		svc := transferTestService(WarehouseDAOMock{})
		_, err := svc.Create(ctx, &WarehouseTransfer{SrcWarehouseID: 1, DstWarehouseID: 2}, nil)
		helper.AssertError(t, err, true, ErrOrganizationMissing)
	})

	t.Run("warehouse errors", func(t *testing.T) {
		svc := transferTestService(warehouses(dbErr, nil, nil, nil))
		_, err := svc.Create(ctx, transfer(), nil)
		helper.AssertError(t, err, true, dbErr)

		svc = transferTestService(warehouses(nil, nil, nil, nil))
		_, err = svc.Create(ctx, transfer(), nil)
		helper.AssertError(t, err, true, ErrWarehouseNotFound)

		svc = transferTestService(warehouses(nil, owned, dbErr, nil))
		_, err = svc.Create(ctx, transfer(), nil)
		helper.AssertError(t, err, true, dbErr)

		svc = transferTestService(warehouses(nil, owned, nil, nil))
		_, err = svc.Create(ctx, transfer(), nil)
		helper.AssertError(t, err, true, ErrWarehouseNotFound)
	})
}

func TestTransferService_Send_Errors(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	dbErr := errors.New("db down")
	orgID := uint64(10)
	shipmentID := uint64(5)

	newSvc := func(transfer *WarehouseTransfer, transferErr error, movements []*StockMovement, movesErr error) TransferService {
		transfers := WarehouseTransferDAOMock{
			CRUDMock: dao.CRUDMock[WarehouseTransfer]{
				FindFunc: func(_ context.Context, _ uint64) (*WarehouseTransfer, error) { return transfer, transferErr },
			},
		}
		moveMock := StockMovementDAOMock{
			ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) { return movements, movesErr },
		}
		return NewTransferService(transfers, moveMock, StockLocationDAOMock{}, WarehouseDAOMock{}, CostLayerDAOMock{}, NewLedgerService(moveMock, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})
	}
	draft := func() *WarehouseTransfer {
		return &WarehouseTransfer{Base: model.Base{ID: 1}, OrganizationID: &orgID, SrcWarehouseID: 1, DstWarehouseID: 2, State: TransferStateDraft, OutShipmentID: &shipmentID}
	}

	cases := []struct {
		name      string
		transfer  *WarehouseTransfer
		transferE error
		transit   uint64
		movements []*StockMovement
		movesErr  error
		wantErr   error
	}{
		{name: "lookup error", transferE: dbErr, transit: 1, wantErr: dbErr},
		{name: "missing", transit: 1, wantErr: ErrWarehouseTransferNotFound},
		{name: "bad state", transfer: &WarehouseTransfer{Base: model.Base{ID: 1}, State: TransferStateInTransit}, transit: 1, wantErr: ErrWarehouseTransferState},
		{name: "no shipment", transfer: &WarehouseTransfer{Base: model.Base{ID: 1}, State: TransferStateDraft}, transit: 1, wantErr: ErrWarehouseTransferState},
		{name: "no transit account", transfer: draft(), wantErr: ErrTransitAccount},
		{name: "no org", transfer: &WarehouseTransfer{Base: model.Base{ID: 1}, State: TransferStateDraft, OutShipmentID: &shipmentID}, transit: 1, wantErr: ErrOrganizationMissing},
		{name: "movements error", transfer: draft(), transit: 1, movesErr: dbErr, wantErr: dbErr},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc(tt.transfer, tt.transferE, tt.movements, tt.movesErr)
			helper.AssertError(t, svc.Send(ctx, 1, 500, tt.transit, now), true, tt.wantErr)
		})
	}
}

func TestTransferService_Create_Lines(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	orgID := uint64(10)
	owned := &reference.Warehouse{Base: model.Base{ID: 1}, OrganizationID: &orgID}
	transfer := func() *WarehouseTransfer {
		return &WarehouseTransfer{OrganizationID: &orgID, SrcWarehouseID: 1, DstWarehouseID: 2}
	}
	warehouses := func(srcErr error, src *reference.Warehouse, dstErr error, dst *reference.Warehouse) WarehouseDAOMock {
		calls := 0
		return WarehouseDAOMock{
			CRUDMock: dao.CRUDMock[reference.Warehouse]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
					calls++
					if calls == 1 {
						return src, srcErr
					}
					return dst, dstErr
				},
			},
		}
	}

	t.Run("line qty and resolve errors", func(t *testing.T) {
		locations := StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 50}, Usage: "transit"}}}, nil
				},
			},
		}
		wh := warehouses(nil, owned, nil, owned)
		svc := transferTestServiceWithLocations(wh, locations)
		_, err := svc.Create(ctx, transfer(), []TransferLine{{Qty: 0}})
		helper.AssertError(t, err, true, ErrMovementQty)

		badLocations := StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 50}, Usage: "transit"}}}, nil
				},
				FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
					return nil, dbErr
				},
			},
		}
		svc = transferTestServiceWithLocations(wh, badLocations)
		_, err = svc.Create(ctx, transfer(), []TransferLine{{Qty: 5, SrcLocationID: helper.Ptr(uint64(51))}})
		helper.AssertError(t, err, true, dbErr)
	})
}
