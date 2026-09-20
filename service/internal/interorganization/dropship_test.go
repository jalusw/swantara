package interorganization_test

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/sales"
	"gorm.io/gorm"
)

func newDropShipServiceForTest(
	links interorganization.DropshipLinkDAO,
	poCreate interorganization.PurchaseOrderCreator,
	poOrders procurement.PurchaseOrderDAO,
	poLines procurement.PurchaseOrderLineDAO,
	soOrders sales.SaleOrderDAO,
	soLines sales.SaleOrderLineDAO,
	movements inventory.StockMovementDAO,
	locations inventory.StockLocationDAO,
	resolver inventory.ItemResolver,
	poster accounting.Poster,
	tx db.Transactioner,
) interorganization.DropShipService {
	return interorganization.NewDropShipService(links, poCreate, poOrders, poLines, soOrders, soLines, movements, locations, resolver, poster, tx)
}

func TestDropShipService_Create_MirrorsSaleOrderIntoPurchaseOrder(t *testing.T) {
	var createdPO *procurement.PurchaseOrder
	var createdLines []*procurement.PurchaseOrderLine
	links := interorganization.DropshipLinkDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.DropshipLink]{
			CreateFunc: func(_ context.Context, link *interorganization.DropshipLink) (*interorganization.DropshipLink, error) {
				return &interorganization.DropshipLink{Base: baseID(300), SaleOrderLineID: link.SaleOrderLineID, PurchaseOrderLineID: link.PurchaseOrderLineID}, nil
			},
		},
	}
	poCreate := purchaseOrderCreatorMock{
		CreateFunc: func(_ context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			createdPO = order
			createdLines = lines
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: order.OrganizationID, State: procurement.PurchaseOrderStateDraft}, nil
		},
	}
	poOrders := procurement.PurchaseOrderDAOMock{}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{
				{Base: baseID(210), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5},
				{Base: baseID(211), ItemID: helper.Ptr(uint64(101)), QtyOrdered: 3},
			}, nil
		},
	}
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed, CurrencyCode: helper.Ptr("IDR")}, nil
			},
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{
				{Base: baseID(300), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5, UnitPrice: 100},
				{Base: baseID(301), ItemID: helper.Ptr(uint64(101)), QtyOrdered: 3, UnitPrice: 50},
			}, nil
		},
	}
	movements := inventory.StockMovementDAOMock{}
	locations := inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: baseID(500), Usage: "customer"}}}, nil
			},
		},
	}
	svc := newDropShipServiceForTest(links, poCreate, poOrders, poLines, soOrders, soLines, movements, locations, inventory.ItemResolverMock{}, posterMock{}, txMock{})

	order, created, err := svc.Create(context.Background(), interorganization.CreateDropshipOrderRequest{
		OrganizationID: 10,
		SaleOrderID:    30,
		SupplierID:     700,
		Date:           time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if createdPO == nil || *createdPO.DestLocationID != 500 || createdPO.SupplierID != 700 {
		t.Errorf("po = %+v, want dest 500 supplier 700", createdPO)
	}
	if len(createdLines) != 2 || createdLines[0].QtyOrdered != 5 || createdLines[1].QtyOrdered != 3 {
		t.Errorf("lines = %+v, want 2 mirrored lines with qty 5 and 3", createdLines)
	}
	if order.ID != 200 {
		t.Errorf("order id = %d, want 200", order.ID)
	}
	if len(created) != 2 || created[0].SaleOrderLineID != 300 || created[0].PurchaseOrderLineID != 210 {
		t.Errorf("links = %+v, want line 300 linked to 210", created)
	}
}

func TestDropShipService_Create_CancelledSourceRejected(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateCancelled}, nil
		}}},
		sales.SaleOrderLineDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, _, err := svc.Create(context.Background(), interorganization.CreateDropshipOrderRequest{OrganizationID: 10, SaleOrderID: 30})
	if err != interorganization.ErrDropShipSourceNotActive {
		t.Errorf("err = %v, want ErrDropShipSourceNotActive", err)
	}
}

func TestDropShipService_Receive_AppliesMovesAndPostsCOGS(t *testing.T) {
	var appliedMovementIDs []uint64
	var postedDebit float64
	var soDelivered float64
	var poState string
	links := interorganization.DropshipLinkDAOMock{
		ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
			return []*interorganization.DropshipLink{{Base: baseID(400), SaleOrderLineID: 300, PurchaseOrderLineID: 210}}, nil
		},
	}
	poOrders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), DestLocationID: helper.Ptr(uint64(500)), State: procurement.PurchaseOrderStateConfirmed}, nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, order *procurement.PurchaseOrder) (*procurement.PurchaseOrder, error) {
			poState = order.State
			return order, nil
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5, UnitPrice: 100}}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, line *procurement.PurchaseOrderLine) (*procurement.PurchaseOrderLine, error) {
			return line, nil
		},
	}
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, order *sales.SaleOrder) (*sales.SaleOrder, error) {
			return order, nil
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
				return &sales.SaleOrderLine{Base: baseID(300), OrderID: 30, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}, nil
			},
		},
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: baseID(300), OrderID: 30, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, line *sales.SaleOrderLine) (*sales.SaleOrderLine, error) {
			soDelivered = line.QtyDelivered
			return line, nil
		},
	}
	movements := inventory.StockMovementDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
			return &inventory.StockMovement{Base: baseID(700), OrganizationID: movement.OrganizationID, ItemID: movement.ItemID, Qty: movement.Qty, SrcLocationID: movement.SrcLocationID, DstLocationID: movement.DstLocationID}, nil
		},
		FindForUpdateTxFunc: func(_ context.Context, _ *gorm.DB, moveID uint64) (*inventory.StockMovement, error) {
			return &inventory.StockMovement{Base: baseID(moveID), State: inventory.MovementStateConfirmed}, nil
		},
		ApplyAllTxFunc: func(_ context.Context, _ *gorm.DB, movements []*inventory.StockMovement) error {
			for _, movement := range movements {
				appliedMovementIDs = append(appliedMovementIDs, movement.ID)
			}
			return nil
		},
	}
	locations := inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: baseID(600), Usage: "supplier"}}}, nil
			},
		},
	}
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{CogsAccountID: 1, StockInputAccountID: 2}}, nil
		},
	}
	poster := posterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			postedDebit = request.Lines[0].Debit.Float64()
			return &accounting.JournalEntry{}, nil
		},
	}
	svc := newDropShipServiceForTest(links, purchaseOrderCreatorMock{}, poOrders, poLines, soOrders, soLines, movements, locations, resolver, poster, txMock{})

	order, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{
		OrganizationID:  10,
		PurchaseOrderID: 200,
		JournalID:       9,
		Date:            time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(appliedMovementIDs) != 1 || appliedMovementIDs[0] != 700 {
		t.Errorf("applied movements = %v, want [700]", appliedMovementIDs)
	}
	if postedDebit != 500 {
		t.Errorf("posted debit = %v, want 500", postedDebit)
	}
	if soDelivered != 5 {
		t.Errorf("so delivered = %v, want 5", soDelivered)
	}
	if order.State != procurement.PurchaseOrderStateDone {
		t.Errorf("po state = %s, want done", order.State)
	}
	if poState != procurement.PurchaseOrderStateDone {
		t.Errorf("po persisted state = %s, want done", poState)
	}
}
