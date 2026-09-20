package pos

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func readyPOSConfig() reference.POSConfig {
	organizationID := uint64(1)
	return reference.POSConfig{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		WarehouseID: helper.Ptr(uint64(4)), JournalID: helper.Ptr(uint64(2)), PriceBookID: helper.Ptr(uint64(3)),
	}
}

func sellReadyService(config *reference.POSConfig, configs dao.CRUDMock[reference.POSConfig]) POSService {
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1, State: SessionStateOpened}, nil
		},
	}}
	svc := testPOSService(configs, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
	svc.configs = dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return config, nil },
	}
	return svc
}

func withSellStock(svc POSService) POSService {
	svc.locations = inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "usage" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
			},
		},
	}
	svc.quants = inventory.StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return &inventory.StockBalance{Base: model.Base{ID: 3}, Quantity: 10}, nil
		},
	}
	return svc
}

func TestPOSService_Sell_RejectsNoLines(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{})

	_, err := svc.Sell(ctx, SellRequest{SessionID: 1, Payments: []SellPaymentRequest{{Method: "cash", Amount: 10}}})
	if helper.AssertError(t, err, true, ErrOrderNoLines) {
		return
	}
}

func TestPOSService_Sell_RejectsNoPayments(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{})

	_, err := svc.Sell(ctx, SellRequest{SessionID: 1, Lines: []SellLineRequest{{ItemID: 100, Qty: 2}}})
	if helper.AssertError(t, err, true, ErrPaymentMissing) {
		return
	}
}

func TestPOSService_Sell_RejectsClosedSession(t *testing.T) {
	ctx := context.Background()
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, State: SessionStateClosed}, nil
		},
	}}
	svc := testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrSessionState) {
		return
	}
}

func TestPOSService_Sell_RejectsConfigWithoutWarehouse(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	config.WarehouseID = nil
	svc := sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{})

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrConfigWarehouse) {
		return
	}
}

func TestPOSService_Sell_RejectsConfigWithoutJournal(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	config.JournalID = nil
	svc := sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{})

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrConfigJournal) {
		return
	}
}

func TestPOSService_Sell_RejectsConfigWithoutPriceBook(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	config.PriceBookID = nil
	svc := sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{})

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrConfigPriceBook) {
		return
	}
}

func TestPOSService_Sell_RejectsConfigWithoutOrganization(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	config.OrganizationID = nil
	svc := sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{})

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrConfigNotFound) {
		return
	}
}

func TestPOSService_Sell_RejectsInvalidPayment(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := withSellStock(sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{}))

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrPaymentTotal) {
		return
	}
}

func TestPOSService_Sell_RejectsProductZero(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := withSellStock(sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{}))

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 0, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrOrderProduct) {
		return
	}
}

func TestPOSService_Sell_RejectsInvalidQty(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := withSellStock(sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{}))

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: -1}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrOrderQty) {
		return
	}
}

func TestPOSService_Sell_RejectsInvalidDiscount(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := withSellStock(sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{}))

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2, DiscountPct: 120}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrOrderDiscount) {
		return
	}
}

func TestPOSService_Sell_RejectsUnknownTax(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := withSellStock(sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{}))
	svc.taxes = dao.CRUDMock[reference.Tax]{}

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2, TaxIDs: helper.Int64Array{9}}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrOrderTax) {
		return
	}
}

func TestPOSService_Sell_RejectsPurchaseTaxScope(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := withSellStock(sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{}))
	svc.taxes = dao.CRUDMock[reference.Tax]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
			return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopePurchase, Amount: helper.Ptr(10.0)}, nil
		},
	}

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2, TaxIDs: helper.Int64Array{9}}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrOrderTax) {
		return
	}
}

func TestPOSService_Sell_PropagatesStockLocationError(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{})

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrOrderNoStock) {
		return
	}
}

func TestPOSService_Sell_PropagatesCustomerLocationError(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{})
	svc.locations = inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "usage" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{}}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
			},
		},
	}

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, ErrOrderNoStock) {
		return
	}
}

func TestPOSService_Sell_PropagatesQuantError(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{})
	svc.locations = inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "usage" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
			},
		},
	}
	svc.quants = inventory.StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return nil, errors.New("boom")
		},
	}

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_Sell_PropagatesPostStockMovementsError(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := withSellStock(sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{}))
	svc.shipments = inventory.ShipmentDAOMock{
		CreateWithMovementsFunc: func(_ context.Context, _ *inventory.Shipment, _ []*inventory.StockMovement) (*inventory.Shipment, error) {
			return nil, errors.New("boom")
		},
	}

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_Sell_PropagatesShipError(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := withSellStock(sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{}))
	svc.shipments = inventory.ShipmentDAOMock{
		CreateWithMovementsFunc: func(_ context.Context, shipment *inventory.Shipment, movements []*inventory.StockMovement) (*inventory.Shipment, error) {
			shipment.ID = 20
			for i := range movements {
				movements[i].ID = uint64(30 + i)
			}
			return shipment, nil
		},
	}
	svc.valuer = ShipEngineMock{
		ShipFunc: func(_ context.Context, _ uint64, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
			return nil, errors.New("boom")
		},
	}

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_Sell_PropagatesShipmentUpdateError(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	svc := withSellStock(sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{}))
	svc.shipments = inventory.ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[inventory.Shipment]{
			UpdateFunc: func(_ context.Context, _ *inventory.Shipment) (*inventory.Shipment, error) {
				return nil, errors.New("boom")
			},
		},
		CreateWithMovementsFunc: func(_ context.Context, shipment *inventory.Shipment, movements []*inventory.StockMovement) (*inventory.Shipment, error) {
			shipment.ID = 20
			for i := range movements {
				movements[i].ID = uint64(30 + i)
			}
			return shipment, nil
		},
	}

	_, err := svc.Sell(ctx, SellRequest{
		SessionID: 1,
		Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
		Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
	})
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_OpenSession_PropagatesConfigError(t *testing.T) {
	ctx := context.Background()
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
			return nil, errors.New("boom")
		},
	}
	svc := testPOSService(configs, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})

	_, err := svc.OpenSession(ctx, 1, 7, 100)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_OpenSession_PropagatesMemberError(t *testing.T) {
	ctx := context.Background()
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
			return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))}, nil
		},
	}
	svc := testPOSService(configs, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
	svc.members = iam.MemberDAOMock{
		FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
			return nil, errors.New("boom")
		},
	}

	_, err := svc.OpenSession(ctx, 1, 7, 100)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_OpenSession_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
			return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))}, nil
		},
	}
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[POSSession], error) {
			return nil, errors.New("boom")
		},
	}}
	svc := testPOSService(configs, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})

	_, err := svc.OpenSession(ctx, 1, 7, 100)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_OpenSession_RejectsNoCashier(t *testing.T) {
	ctx := context.Background()
	svc := testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})

	_, err := svc.OpenSession(ctx, 1, 0, 100)
	if helper.AssertError(t, err, true, ErrNoCashier) {
		return
	}
}

func TestPOSService_StartClosing_PropagatesUpdateError(t *testing.T) {
	ctx := context.Background()
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, State: SessionStateOpened}, nil
		},
		UpdateFunc: func(_ context.Context, _ *POSSession) (*POSSession, error) {
			return nil, errors.New("boom")
		},
	}}
	svc := testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})

	_, err := svc.StartClosing(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_CloseSession_PropagatesPaymentTotalError(t *testing.T) {
	ctx := context.Background()
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, OpeningBalance: 100, State: SessionStateClosing}, nil
		},
	}}
	orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[POSOrder], error) {
			return nil, errors.New("boom")
		},
	}}
	svc := testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})

	_, err := svc.CloseSession(ctx, 1, 250)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_CloseSession_PropagatesUpdateError(t *testing.T) {
	ctx := context.Background()
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, OpeningBalance: 100, State: SessionStateClosing}, nil
		},
		UpdateFunc: func(_ context.Context, _ *POSSession) (*POSSession, error) {
			return nil, errors.New("boom")
		},
	}}
	svc := testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})

	_, err := svc.CloseSession(ctx, 1, 100)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_Refund_PropagatesLinesError(t *testing.T) {
	ctx := context.Background()
	orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
			return &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, State: OrderStateDone}, nil
		},
	}}
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
		},
	}}
	config := readyPOSConfig()
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	lines := POSOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
			return nil, errors.New("boom")
		},
	}
	svc := testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})

	_, err := svc.Refund(ctx, 10, 2, time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_Refund_RejectsNoLines(t *testing.T) {
	ctx := context.Background()
	orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
			return &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, State: OrderStateDone}, nil
		},
	}}
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
		},
	}}
	config := readyPOSConfig()
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	svc := testPOSService(configs, sessions, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})

	_, err := svc.Refund(ctx, 10, 2, time.Now())
	if helper.AssertError(t, err, true, ErrOrderNoLines) {
		return
	}
}

func TestPOSService_Refund_RejectsMissingConfig(t *testing.T) {
	ctx := context.Background()
	orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
			return &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, State: OrderStateDone}, nil
		},
	}}
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
		},
	}}
	svc := testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})

	_, err := svc.Refund(ctx, 10, 2, time.Now())
	if helper.AssertError(t, err, true, ErrConfigNotFound) {
		return
	}
}

func TestPOSService_Refund_PropagatesRestockMoveError(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	order := &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, AmountTotal: amount.FromFloat64(220), State: OrderStateDone}
	orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) { return order, nil },
	}}
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	lines := POSOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
			return []*POSOrderLine{{ItemID: helper.Ptr(uint64(100)), Qty: 2, PriceSubtotal: amount.FromFloat64(200), PriceTax: amount.FromFloat64(20), TaxIDs: helper.Int64Array{9}}}, nil
		},
	}
	svc := testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})
	svc.locations = inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "usage" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
			},
		},
	}
	svc.movements = inventory.StockMovementDAOMock{
		ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
			return nil, errors.New("boom")
		},
	}

	_, err := svc.Refund(ctx, 10, 2, time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_Refund_RejectsMissingOriginalMove(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	order := &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, AmountTotal: amount.FromFloat64(220), State: OrderStateDone}
	orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) { return order, nil },
	}}
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	lines := POSOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
			return []*POSOrderLine{{ItemID: helper.Ptr(uint64(100)), Qty: 2}}, nil
		},
	}
	svc := testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})
	svc.locations = inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "usage" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
			},
		},
	}
	svc.movements = inventory.StockMovementDAOMock{
		ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
			return []*inventory.StockMovement{{Base: model.Base{ID: 30}, ItemID: 999, State: inventory.MovementStateDone}}, nil
		},
	}

	_, err := svc.Refund(ctx, 10, 2, time.Now())
	if helper.AssertError(t, err, true, ErrOrderNoStock) {
		return
	}
}

func TestPOSService_Refund_PropagatesMoveCreateError(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	order := &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, AmountTotal: amount.FromFloat64(220), State: OrderStateDone}
	orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) { return order, nil },
	}}
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	lines := POSOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
			return []*POSOrderLine{{ItemID: helper.Ptr(uint64(100)), Qty: 2}}, nil
		},
	}
	svc := testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})
	svc.locations = inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "usage" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
			},
		},
	}
	svc.movements = inventory.StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[inventory.StockMovement]{
			CreateFunc: func(_ context.Context, _ *inventory.StockMovement) (*inventory.StockMovement, error) {
				return nil, errors.New("boom")
			},
		},
		ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
			return []*inventory.StockMovement{{Base: model.Base{ID: 30}, ItemID: 100, State: inventory.MovementStateDone}}, nil
		},
	}

	_, err := svc.Refund(ctx, 10, 2, time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_Refund_RejectsMissingUnitCost(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	order := &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, AmountTotal: amount.FromFloat64(220), State: OrderStateDone}
	orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) { return order, nil },
	}}
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	lines := POSOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
			return []*POSOrderLine{{ItemID: helper.Ptr(uint64(100)), Qty: 2}}, nil
		},
	}
	svc := testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})
	svc.locations = inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "usage" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
			},
		},
	}
	svc.movements = inventory.StockMovementDAOMock{
		ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
			return []*inventory.StockMovement{{Base: model.Base{ID: 30}, ItemID: 100, State: inventory.MovementStateDone}}, nil
		},
	}
	svc.layers = inventory.CostLayerDAOMock{
		ListByMovementFunc: func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{MovementID: helper.Ptr(uint64(30)), UnitCost: nil}}, nil
		},
	}

	_, err := svc.Refund(ctx, 10, 2, time.Now())
	if helper.AssertError(t, err, true, ErrOrderCost) {
		return
	}
}

func TestPOSService_Refund_PropagatesJournalError(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	order := &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, AmountTotal: amount.FromFloat64(220), State: OrderStateDone}
	orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) { return order, nil },
	}}
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	lines := POSOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
			return []*POSOrderLine{{ItemID: helper.Ptr(uint64(100)), Qty: 2}}, nil
		},
	}
	svc := testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})
	svc.locations = inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "usage" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
			},
		},
	}
	svc.movements = inventory.StockMovementDAOMock{
		ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
			return []*inventory.StockMovement{{Base: model.Base{ID: 30}, ItemID: 100, State: inventory.MovementStateDone}}, nil
		},
	}
	svc.layers = inventory.CostLayerDAOMock{
		ListByMovementFunc: func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{MovementID: helper.Ptr(uint64(30)), UnitCost: helper.Ptr(100.0)}}, nil
		},
	}
	svc.journals = dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return nil, errors.New("boom")
		},
	}

	_, err := svc.Refund(ctx, 10, 2, time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_Refund_RejectsJournalWithoutDefaultAccount(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	order := &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, AmountTotal: amount.FromFloat64(220), State: OrderStateDone}
	orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) { return order, nil },
	}}
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	lines := POSOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
			return []*POSOrderLine{{ItemID: helper.Ptr(uint64(100)), Qty: 2}}, nil
		},
	}
	svc := testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})
	svc.locations = inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "usage" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
			},
		},
	}
	svc.movements = inventory.StockMovementDAOMock{
		ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
			return []*inventory.StockMovement{{Base: model.Base{ID: 30}, ItemID: 100, State: inventory.MovementStateDone}}, nil
		},
	}
	svc.layers = inventory.CostLayerDAOMock{
		ListByMovementFunc: func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{MovementID: helper.Ptr(uint64(30)), UnitCost: helper.Ptr(100.0)}}, nil
		},
	}
	svc.journals = dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 2}}, nil
		},
	}

	_, err := svc.Refund(ctx, 10, 2, time.Now())
	if helper.AssertError(t, err, true, ErrConfigJournal) {
		return
	}
}

func TestPOSService_Refund_PropagatesUpdateError(t *testing.T) {
	ctx := context.Background()
	config := readyPOSConfig()
	order := &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, AmountTotal: amount.FromFloat64(220), State: OrderStateDone}
	orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
		FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) { return order, nil },
		UpdateFunc: func(_ context.Context, _ *POSOrder) (*POSOrder, error) {
			return nil, errors.New("boom")
		},
	}}
	sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
		FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
			return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
		},
	}}
	configs := dao.CRUDMock[reference.POSConfig]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
	}
	lines := POSOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
			return []*POSOrderLine{{ItemID: helper.Ptr(uint64(100)), Qty: 2, PriceSubtotal: amount.FromFloat64(200), PriceTax: amount.FromFloat64(20), TaxIDs: helper.Int64Array{9}}}, nil
		},
	}
	svc := testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})
	svc.locations = inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "usage" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
			},
		},
	}
	svc.movements = inventory.StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[inventory.StockMovement]{
			CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 40
				return movement, nil
			},
		},
		ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
			return []*inventory.StockMovement{{Base: model.Base{ID: 30}, ItemID: 100, State: inventory.MovementStateDone}}, nil
		},
	}
	svc.layers = inventory.CostLayerDAOMock{
		ListByMovementFunc: func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{MovementID: helper.Ptr(uint64(30)), UnitCost: helper.Ptr(100.0)}}, nil
		},
	}
	svc.valuer = ShipEngineMock{
		RestockFunc: func(_ context.Context, moveID uint64, _ amount.Amount, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
			return &inventory.CostLayer{MovementID: &moveID}, nil
		},
	}

	_, err := svc.Refund(ctx, 10, 2, time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPOSService_SessionPaymentBreakdown_PropagatesSumError(t *testing.T) {
	ctx := context.Background()
	payments := POSPaymentDAOMock{
		SumBySessionFunc: func(_ context.Context, _ uint64) ([]PaymentMethodTotal, error) {
			return nil, errors.New("boom")
		},
	}
	svc := testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, payments)

	_, err := svc.SessionPaymentBreakdown(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}
