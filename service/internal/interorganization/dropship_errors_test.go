package interorganization_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
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

func TestDropShipService_Create_PropagatesFindError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return nil, errors.New("db down")
		}}},
		sales.SaleOrderLineDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, _, err := svc.Create(context.Background(), interorganization.CreateDropshipOrderRequest{SaleOrderID: 30})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestDropShipService_Create_SourceNotFound(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(99)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, _, err := svc.Create(context.Background(), interorganization.CreateDropshipOrderRequest{OrganizationID: 10, SaleOrderID: 30})
	if err != interorganization.ErrDropShipSourceNotFound {
		t.Errorf("err = %v, want ErrDropShipSourceNotFound", err)
	}
}

func TestDropShipService_Create_PropagatesLineError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return nil, errors.New("db down")
		}},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, _, err := svc.Create(context.Background(), interorganization.CreateDropshipOrderRequest{OrganizationID: 10, SaleOrderID: 30})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestDropShipService_Create_SkipsInvalidLinesAndReturnsNoLines(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: baseID(300), QtyOrdered: 0}, {Base: baseID(301), ItemID: nil}}, nil
		}},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, _, err := svc.Create(context.Background(), interorganization.CreateDropshipOrderRequest{OrganizationID: 10, SaleOrderID: 30})
	if err != interorganization.ErrDropShipNoLines {
		t.Errorf("err = %v, want ErrDropShipNoLines", err)
	}
}

func TestDropShipService_Create_PropagatesLocationError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: baseID(300), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}}, nil
		}},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return nil, errors.New("db down")
		}}},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, _, err := svc.Create(context.Background(), interorganization.CreateDropshipOrderRequest{OrganizationID: 10, SaleOrderID: 30})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestDropShipService_Create_DestinationMiss(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: baseID(300), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}}, nil
		}},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: baseID(500), Usage: "supplier"}}}, nil
		}}},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, _, err := svc.Create(context.Background(), interorganization.CreateDropshipOrderRequest{OrganizationID: 10, SaleOrderID: 30})
	if err != interorganization.ErrDropShipDestinationMiss {
		t.Errorf("err = %v, want ErrDropShipDestinationMiss", err)
	}
}

func TestDropShipService_Create_PropagatesPurchaseOrderCreateError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{
			CreateFunc: func(_ context.Context, _ *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
				return nil, errors.New("create failed")
			},
		},
		procurement.PurchaseOrderDAOMock{},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: baseID(300), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}}, nil
		}},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, _, err := svc.Create(context.Background(), interorganization.CreateDropshipOrderRequest{
		OrganizationID: 10,
		SaleOrderID:    30,
		DestLocationID: helper.Ptr(uint64(500)),
	})
	if err == nil || err.Error() != "create failed" {
		t.Errorf("err = %v, want create failed", err)
	}
}

func TestDropShipService_Create_PropagatesCreatedLineError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{},
		procurement.PurchaseOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return nil, errors.New("db down")
		}},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: baseID(300), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}}, nil
		}},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, _, err := svc.Create(context.Background(), interorganization.CreateDropshipOrderRequest{
		OrganizationID: 10,
		SaleOrderID:    30,
		DestLocationID: helper.Ptr(uint64(500)),
	})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestDropShipService_Create_PropagatesLinkCreateError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.DropshipLink]{
				CreateFunc: func(_ context.Context, _ *interorganization.DropshipLink) (*interorganization.DropshipLink, error) {
					return nil, errors.New("insert failed")
				},
			},
		},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{},
		procurement.PurchaseOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(100))}}, nil
		}},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: baseID(300), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}}, nil
		}},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, _, err := svc.Create(context.Background(), interorganization.CreateDropshipOrderRequest{
		OrganizationID: 10,
		SaleOrderID:    30,
		DestLocationID: helper.Ptr(uint64(500)),
	})
	if err == nil || err.Error() != "insert failed" {
		t.Errorf("err = %v, want insert failed", err)
	}
}

func TestDropShipService_Receive_PropagatesFindError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return nil, errors.New("db down")
		}}},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestDropShipService_Receive_OrderNotFound(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(99)), State: procurement.PurchaseOrderStateConfirmed}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err != interorganization.ErrDropShipOrderNotFound {
		t.Errorf("err = %v, want ErrDropShipOrderNotFound", err)
	}
}

func TestDropShipService_Receive_AlreadyReceivedWhenDone(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), State: procurement.PurchaseOrderStateDone}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err != interorganization.ErrDropShipAlreadyReceived {
		t.Errorf("err = %v, want ErrDropShipAlreadyReceived", err)
	}
}

func TestDropShipService_Receive_AlreadyReceivedWhenCancelled(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), State: procurement.PurchaseOrderStateCancelled}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err != interorganization.ErrDropShipAlreadyReceived {
		t.Errorf("err = %v, want ErrDropShipAlreadyReceived", err)
	}
}

func TestDropShipService_Receive_NoLines(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), State: procurement.PurchaseOrderStateConfirmed}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err != interorganization.ErrDropShipNoLines {
		t.Errorf("err = %v, want ErrDropShipNoLines", err)
	}
}

func TestDropShipService_Receive_SourceNotFound(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{
			ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
				return []*interorganization.DropshipLink{{Base: baseID(400), SaleOrderLineID: 300, PurchaseOrderLineID: 210}}, nil
			},
		},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), State: procurement.PurchaseOrderStateConfirmed}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
		}},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
			return nil, nil
		}}},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err != interorganization.ErrDropShipSourceNotFound {
		t.Errorf("err = %v, want ErrDropShipSourceNotFound", err)
	}
}

func TestDropShipService_Receive_PropagatesSoLineFindError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{
			ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
				return []*interorganization.DropshipLink{{Base: baseID(400), SaleOrderLineID: 300, PurchaseOrderLineID: 210}}, nil
			},
		},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), State: procurement.PurchaseOrderStateConfirmed}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
		}},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
			return nil, errors.New("db down")
		}}},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestDropShipService_Receive_PropagatesSoFindError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{
			ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
				return []*interorganization.DropshipLink{{Base: baseID(400), SaleOrderLineID: 300, PurchaseOrderLineID: 210}}, nil
			},
		},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), State: procurement.PurchaseOrderStateConfirmed}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
		}},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return nil, errors.New("db down")
		}}},
		sales.SaleOrderLineDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
			return &sales.SaleOrderLine{Base: baseID(300), OrderID: 30, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}, nil
		}}},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestDropShipService_Receive_PropagatesAllSoLinesError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{
			ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
				return []*interorganization.DropshipLink{{Base: baseID(400), SaleOrderLineID: 300, PurchaseOrderLineID: 210}}, nil
			},
		},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), State: procurement.PurchaseOrderStateConfirmed}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
		}},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{
			CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{
				FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
					return &sales.SaleOrderLine{Base: baseID(300), OrderID: 30, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}, nil
				},
			},
			ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
				return nil, errors.New("db down")
			},
		},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestDropShipService_Receive_SupplierMissing(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{
			ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
				return []*interorganization.DropshipLink{{Base: baseID(400), SaleOrderLineID: 300, PurchaseOrderLineID: 210}}, nil
			},
		},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), DestLocationID: helper.Ptr(uint64(500)), State: procurement.PurchaseOrderStateConfirmed}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
		}},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
			return &sales.SaleOrderLine{Base: baseID(300), OrderID: 30, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}, nil
		}}},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: baseID(600), Usage: "customer"}}}, nil
		}}},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err != interorganization.ErrDropShipSupplierMissing {
		t.Errorf("err = %v, want ErrDropShipSupplierMissing", err)
	}
}

func TestDropShipService_Receive_DestinationMiss(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{
			ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
				return []*interorganization.DropshipLink{{Base: baseID(400), SaleOrderLineID: 300, PurchaseOrderLineID: 210}}, nil
			},
		},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), State: procurement.PurchaseOrderStateConfirmed}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
		}},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
			return &sales.SaleOrderLine{Base: baseID(300), OrderID: 30, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}, nil
		}}},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: baseID(600), Usage: "supplier"}}}, nil
		}}},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err != interorganization.ErrDropShipDestinationMiss {
		t.Errorf("err = %v, want ErrDropShipDestinationMiss", err)
	}
}

func TestDropShipService_Receive_PropagatesPoLinesError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), State: procurement.PurchaseOrderStateConfirmed}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return nil, errors.New("db down")
		}},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestDropShipService_Receive_PropagatesLinksError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{
			ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
				return nil, errors.New("db down")
			},
		},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), State: procurement.PurchaseOrderStateConfirmed}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(100))}}, nil
		}},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestDropShipService_Receive_PropagatesLocationError(t *testing.T) {
	svc := newDropShipServiceForTest(
		interorganization.DropshipLinkDAOMock{
			ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
				return []*interorganization.DropshipLink{{Base: baseID(400), SaleOrderLineID: 300, PurchaseOrderLineID: 210}}, nil
			},
		},
		purchaseOrderCreatorMock{},
		procurement.PurchaseOrderDAOMock{CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), DestLocationID: helper.Ptr(uint64(500)), State: procurement.PurchaseOrderStateConfirmed}, nil
		}}},
		procurement.PurchaseOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
		}},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{
			CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{
				FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
					return &sales.SaleOrderLine{Base: baseID(300), OrderID: 30, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}, nil
				},
			},
			ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
				return []*sales.SaleOrderLine{{Base: baseID(300), OrderID: 30, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
			},
		},
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return nil, errors.New("db down")
		}}},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{OrganizationID: 10, PurchaseOrderID: 200})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestDropShipService_Receive_PropagatesTxErrors(t *testing.T) {
	cases := []struct {
		name          string
		moveError     error
		resolverError error
		posterError   error
		linkError     error
		poLineError   error
		soLineError   error
		applyError    error
		soError       error
		poError       error
	}{
		{name: "movement create fails", moveError: errors.New("movement failed")},
		{name: "resolve fails", resolverError: errors.New("resolve failed")},
		{name: "post fails", posterError: errors.New("post failed")},
		{name: "link update fails", linkError: errors.New("link update failed")},
		{name: "po line update fails", poLineError: errors.New("po line update failed")},
		{name: "so line update fails", soLineError: errors.New("so line update failed")},
		{name: "apply fails", applyError: errors.New("apply failed")},
		{name: "so update fails", soError: errors.New("so update failed")},
		{name: "po update fails", poError: errors.New("po update failed")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newDropShipServiceForTest(
				interorganization.DropshipLinkDAOMock{
					ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
						return []*interorganization.DropshipLink{{Base: baseID(400), SaleOrderLineID: 300, PurchaseOrderLineID: 210}}, nil
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, link *interorganization.DropshipLink) (*interorganization.DropshipLink, error) {
						if tc.linkError != nil {
							return nil, tc.linkError
						}
						return link, nil
					},
				},
				purchaseOrderCreatorMock{},
				procurement.PurchaseOrderDAOMock{
					CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
						FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
							return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(10)), DestLocationID: helper.Ptr(uint64(500)), State: procurement.PurchaseOrderStateConfirmed}, nil
						},
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, order *procurement.PurchaseOrder) (*procurement.PurchaseOrder, error) {
						if tc.poError != nil {
							return nil, tc.poError
						}
						return order, nil
					},
				},
				procurement.PurchaseOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
						return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5, UnitPrice: 100}, {Base: baseID(211), QtyOrdered: 3}, {Base: baseID(212), ItemID: helper.Ptr(uint64(101)), QtyOrdered: 2}}, nil
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, line *procurement.PurchaseOrderLine) (*procurement.PurchaseOrderLine, error) {
						if tc.poLineError != nil {
							return nil, tc.poLineError
						}
						return line, nil
					},
				},
				sales.SaleOrderDAOMock{
					CRUDMock: dao.CRUDMock[sales.SaleOrder]{
						FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
							return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
						},
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, order *sales.SaleOrder) (*sales.SaleOrder, error) {
						if tc.soError != nil {
							return nil, tc.soError
						}
						return order, nil
					},
				},
				sales.SaleOrderLineDAOMock{
					CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{
						FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
							return &sales.SaleOrderLine{Base: baseID(300), OrderID: 30, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}, nil
						},
					},
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
						return []*sales.SaleOrderLine{{Base: baseID(300), OrderID: 30, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, line *sales.SaleOrderLine) (*sales.SaleOrderLine, error) {
						if tc.soLineError != nil {
							return nil, tc.soLineError
						}
						return line, nil
					},
				},
				inventory.StockMovementDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
						if tc.moveError != nil {
							return nil, tc.moveError
						}
						return &inventory.StockMovement{Base: baseID(700), ItemID: movement.ItemID, Qty: movement.Qty}, nil
					},
					FindForUpdateTxFunc: func(_ context.Context, _ *gorm.DB, moveID uint64) (*inventory.StockMovement, error) {
						return &inventory.StockMovement{Base: baseID(moveID), State: inventory.MovementStateConfirmed}, nil
					},
					ApplyAllTxFunc: func(_ context.Context, _ *gorm.DB, _ []*inventory.StockMovement) error {
						if tc.applyError != nil {
							return tc.applyError
						}
						return nil
					},
				},
				inventory.StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: baseID(600), Usage: "supplier"}}}, nil
				}}},
				inventory.ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
						if tc.resolverError != nil {
							return inventory.ResolvedItem{}, tc.resolverError
						}
						return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{CogsAccountID: 1, StockInputAccountID: 2}}, nil
					},
				},
				posterMock{
					PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
						if tc.posterError != nil {
							return nil, tc.posterError
						}
						return &accounting.JournalEntry{}, nil
					},
				},
				txMock{},
			)

			_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{
				OrganizationID:  10,
				PurchaseOrderID: 200,
				JournalID:       9,
				Date:            time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC),
			})
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestDropShipService_Receive_NoMovableLines(t *testing.T) {
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
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5, QtyReceived: 5}}, nil
		},
	}
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
			},
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
				return &sales.SaleOrderLine{Base: baseID(300), OrderID: 30, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}, nil
			},
		},
	}
	svc := newDropShipServiceForTest(
		links,
		purchaseOrderCreatorMock{},
		poOrders,
		poLines,
		soOrders,
		soLines,
		inventory.StockMovementDAOMock{},
		inventory.StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: baseID(600), Usage: "supplier"}}}, nil
		}}},
		inventory.ItemResolverMock{},
		posterMock{},
		txMock{},
	)

	_, err := svc.Receive(context.Background(), interorganization.ReceiveDropshipRequest{
		OrganizationID:  10,
		PurchaseOrderID: 200,
		JournalID:       9,
		Date:            time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC),
	})
	if err != interorganization.ErrDropShipNoLines {
		t.Errorf("err = %v, want ErrDropShipNoLines", err)
	}
}
