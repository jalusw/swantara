package pos

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

var errBoom = errors.New("boom")

func TestPOSService_OpenSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		setup    func() (POSService, uint64, uint64, float64)
		wantErr  error
		validate func(t *testing.T, svc POSService, result *POSSession, err error)
	}{
		{
			name: "creates opened session",
			setup: func() (POSService, uint64, uint64, float64) {
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))}, nil
					},
				}
				sessions := POSSessionDAOMock{}
				svc := testPOSService(configs, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
				return svc, 1, 7, 100.0
			},
			validate: func(t *testing.T, _ POSService, session *POSSession, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if session.State != SessionStateOpened || session.ConfigID != 1 || session.CashierID != 7 || session.OpeningBalance != 100 {
					t.Errorf("session = %+v, want opened config 1 cashier 7 opening 100", session)
				}
				if session.OpenedAt == nil {
					t.Error("opened_at must be set")
				}
			},
		},
		{
			name: "rejects duplicate open session",
			setup: func() (POSService, uint64, uint64, float64) {
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))}, nil
					},
				}
				sessions := POSSessionDAOMock{
					CRUDMock: dao.CRUDMock[POSSession]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[POSSession], error) {
							return &query.Page[POSSession]{Items: []*POSSession{{Base: model.Base{ID: 2}, State: SessionStateOpened}}}, nil
						},
					},
				}
				svc := testPOSService(configs, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
				return svc, 1, 7, 100.0
			},
			wantErr: ErrSessionOpen,
		},
		{
			name: "rejects unknown config",
			setup: func() (POSService, uint64, uint64, float64) {
				svc := testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
				return svc, 99, 7, 100.0
			},
			wantErr: ErrConfigNotFound,
		},
		{
			name: "rejects non-member cashier",
			setup: func() (POSService, uint64, uint64, float64) {
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))}, nil
					},
				}
				svc := testPOSService(configs, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
				svc.members = iam.MemberDAOMock{
					FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
						return nil, nil
					},
				}
				return svc, 1, 7, 100.0
			},
			wantErr: ErrNoCashier,
		},
		{
			name: "propagates config error",
			setup: func() (POSService, uint64, uint64, float64) {
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return nil, errBoom
					},
				}
				svc := testPOSService(configs, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
				return svc, 1, 7, 100.0
			},
			wantErr: errBoom,
		},
		{
			name: "propagates member error",
			setup: func() (POSService, uint64, uint64, float64) {
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))}, nil
					},
				}
				svc := testPOSService(configs, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
				svc.members = iam.MemberDAOMock{
					FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
						return nil, errBoom
					},
				}
				return svc, 1, 7, 100.0
			},
			wantErr: errBoom,
		},
		{
			name: "propagates list error",
			setup: func() (POSService, uint64, uint64, float64) {
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return &reference.POSConfig{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))}, nil
					},
				}
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[POSSession], error) {
						return nil, errBoom
					},
				}}
				svc := testPOSService(configs, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
				return svc, 1, 7, 100.0
			},
			wantErr: errBoom,
		},
		{
			name: "rejects no cashier",
			setup: func() (POSService, uint64, uint64, float64) {
				svc := testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
				return svc, 1, 0, 100.0
			},
			wantErr: ErrNoCashier,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, configID, userID, balance := tt.setup()
			result, err := svc.OpenSession(ctx, configID, userID, balance)
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if tt.validate != nil {
				tt.validate(t, svc, result, err)
			}
		})
	}
}

func TestPOSService_StartClosing(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		setup    func() POSService
		wantErr  error
		validate func(t *testing.T, result *POSSession)
	}{
		{
			name: "transitions to closing",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, State: SessionStateOpened}, nil
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			validate: func(t *testing.T, result *POSSession) {
				if result.State != SessionStateClosing {
					t.Errorf("state = %s, want closing", result.State)
				}
			},
		},
		{
			name: "rejects wrong state",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, State: SessionStateClosed}, nil
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			wantErr: ErrSessionState,
		},
		{
			name: "propagates find error",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return nil, errBoom
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			wantErr: errBoom,
		},
		{
			name: "propagates update error",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, State: SessionStateOpened}, nil
					},
					UpdateFunc: func(_ context.Context, _ *POSSession) (*POSSession, error) {
						return nil, errBoom
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			wantErr: errBoom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			result, err := svc.StartClosing(ctx, 1)
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestPOSService_CloseSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		setup    func() POSService
		wantErr  error
		validate func(t *testing.T, result *POSSession)
	}{
		{
			name: "reconciles and closes",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, OpeningBalance: 100, State: SessionStateClosing}, nil
					},
				}}
				orders := POSOrderDAOMock{
					CRUDMock: dao.CRUDMock[POSOrder]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[POSOrder], error) {
							return &query.Page[POSOrder]{Items: []*POSOrder{{Base: model.Base{ID: 5}}}}, nil
						},
					},
				}
				payments := POSPaymentDAOMock{
					SumBySessionFunc: func(_ context.Context, _ uint64) ([]PaymentMethodTotal, error) {
						return []PaymentMethodTotal{{Method: "cash", Total: 150}, {Method: "card", Total: 50}}, nil
					},
				}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, orders, POSOrderLineDAOMock{}, payments)
			},
			validate: func(t *testing.T, result *POSSession) {
				if result.State != SessionStateClosed || result.ClosedAt == nil {
					t.Errorf("session = %+v, want closed with closed_at", result)
				}
				if result.ClosingBalance == nil || *result.ClosingBalance != 250 {
					t.Errorf("closing_balance = %v, want 250 excluding card payments", result.ClosingBalance)
				}
			},
		},
		{
			name: "rejects mismatched balance",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, OpeningBalance: 100, State: SessionStateClosing}, nil
					},
				}}
				orders := POSOrderDAOMock{
					CRUDMock: dao.CRUDMock[POSOrder]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[POSOrder], error) {
							return &query.Page[POSOrder]{Items: []*POSOrder{{Base: model.Base{ID: 5}}}}, nil
						},
					},
				}
				payments := POSPaymentDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSPayment, error) {
						return []*POSPayment{{Base: model.Base{ID: 1}, Amount: 150}}, nil
					},
				}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, orders, POSOrderLineDAOMock{}, payments)
			},
			wantErr: ErrSessionReconciliation,
		},
		{
			name: "propagates payment total error",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, OpeningBalance: 100, State: SessionStateClosing}, nil
					},
				}}
				payments := POSPaymentDAOMock{
					SumBySessionFunc: func(_ context.Context, _ uint64) ([]PaymentMethodTotal, error) {
						return nil, errBoom
					},
				}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, payments)
			},
			wantErr: errBoom,
		},
		{
			name: "propagates update error",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, OpeningBalance: 100, State: SessionStateClosing}, nil
					},
					UpdateFunc: func(_ context.Context, _ *POSSession) (*POSSession, error) {
						return nil, errBoom
					},
				}}
				payments := POSPaymentDAOMock{
					SumBySessionFunc: func(_ context.Context, _ uint64) ([]PaymentMethodTotal, error) {
						return []PaymentMethodTotal{{Method: "cash", Total: 150}}, nil
					},
				}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, payments)
			},
			wantErr: errBoom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			result, err := svc.CloseSession(ctx, 1, 250)
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestPOSService_Sell(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	config := reference.POSConfig{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		WarehouseID: helper.Ptr(uint64(4)), JournalID: helper.Ptr(uint64(2)), PriceBookID: helper.Ptr(uint64(3)),
	}

	var posted *accounting.PostRequest
	var shiftPosted bool

	tests := []struct {
		name     string
		setup    func() POSService
		req      SellRequest
		wantErr  error
		validate func(t *testing.T, svc POSService, result *POSOrder)
	}{
		{
			name: "creates order posts stock and revenue",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1, State: SessionStateOpened}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
				}
				orders := POSOrderDAOMock{
					CreateWithLinesAndPaymentsFunc: func(_ context.Context, order *POSOrder, _ []*POSOrderLine, _ []*POSPayment) (*POSOrder, error) {
						order.ID = 10
						return order, nil
					},
				}
				locations := inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
							if q.Filters[0].Field == "usage" {
								return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
							}
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
						},
					},
				}
				quants := inventory.StockBalanceDAOMock{
					FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
						return &inventory.StockBalance{Base: model.Base{ID: 3}, Quantity: 10}, nil
					},
				}
				shipped := make([]uint64, 0)
				shipments := inventory.ShipmentDAOMock{
					CreateWithMovementsFunc: func(_ context.Context, shipment *inventory.Shipment, movements []*inventory.StockMovement) (*inventory.Shipment, error) {
						shipment.ID = 20
						for i := range movements {
							movements[i].ID = uint64(30 + i)
						}
						return shipment, nil
					},
				}
				movements := inventory.StockMovementDAOMock{}
				poster := inventory.PosterMock{}

				svc := testPOSService(configs, sessions, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
				svc.locations = locations
				svc.quants = quants
				svc.shipments = shipments
				svc.movements = movements
				svc.valuer = ShipEngineMock{
					ShipFunc: func(_ context.Context, moveID uint64, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
						shipped = append(shipped, moveID)
						return &inventory.CostLayer{MovementID: &moveID}, nil
					},
				}
				svc.poster = poster
				return svc
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2, TaxIDs: helper.Int64Array{9}}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 220}},
			},
			validate: func(t *testing.T, _ POSService, result *POSOrder) {
				if result.ID != 10 || result.State != OrderStateDone || result.AmountTotal.Float64() != 220 || result.AmountTax.Float64() != 20 {
					t.Errorf("order = %+v, want done total 220 tax 20", result)
				}
			},
		},
		{
			name: "rejects mismatched payment",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1, State: SessionStateOpened}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
				}
				locations := inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
							if q.Filters[0].Field == "usage" {
								return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
							}
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
						},
					},
				}
				quants := inventory.StockBalanceDAOMock{
					FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
						return &inventory.StockBalance{Base: model.Base{ID: 3}, Quantity: 10}, nil
					},
				}
				svc := testPOSService(configs, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
				svc.locations = locations
				svc.quants = quants
				return svc
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2, TaxIDs: helper.Int64Array{9}}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrPaymentTotal,
		},
		{
			name: "rejects insufficient stock",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1, State: SessionStateOpened}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
				}
				locations := inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
							if q.Filters[0].Field == "usage" {
								return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
							}
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
						},
					},
				}
				quants := inventory.StockBalanceDAOMock{
					FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
						return &inventory.StockBalance{Base: model.Base{ID: 3}, Quantity: 1}, nil
					},
				}
				svc := testPOSService(configs, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
				svc.locations = locations
				svc.quants = quants
				return svc
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrOrderStockUnavailable,
		},
		{
			name: "propagates revenue post error",
			setup: func() POSService {
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
				svc.poster = inventory.PosterMock{
					PostFunc: func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
						return nil, errBoom
					},
				}
				return svc
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: errBoom,
		},
		{
			name: "rejects no lines",
			setup: func() POSService {
				config := readyPOSConfig()
				return sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{})
			},
			req:     SellRequest{SessionID: 1, Payments: []SellPaymentRequest{{Method: "cash", Amount: 10}}},
			wantErr: ErrOrderNoLines,
		},
		{
			name: "rejects no payments",
			setup: func() POSService {
				config := readyPOSConfig()
				return sellReadyService(&config, dao.CRUDMock[reference.POSConfig]{})
			},
			req:     SellRequest{SessionID: 1, Lines: []SellLineRequest{{ItemID: 100, Qty: 2}}},
			wantErr: ErrPaymentMissing,
		},
		{
			name: "rejects closed session",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, State: SessionStateClosed}, nil
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrSessionState,
		},
		{
			name: "rejects config without warehouse",
			setup: func() POSService {
				cfg := readyPOSConfig()
				cfg.WarehouseID = nil
				return sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{})
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrConfigWarehouse,
		},
		{
			name: "rejects config without journal",
			setup: func() POSService {
				cfg := readyPOSConfig()
				cfg.JournalID = nil
				return sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{})
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrConfigJournal,
		},
		{
			name: "rejects config without price_book",
			setup: func() POSService {
				cfg := readyPOSConfig()
				cfg.PriceBookID = nil
				return sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{})
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrConfigPriceBook,
		},
		{
			name: "rejects config without organization",
			setup: func() POSService {
				cfg := readyPOSConfig()
				cfg.OrganizationID = nil
				return sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{})
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrConfigNotFound,
		},
		{
			name: "rejects invalid payment",
			setup: func() POSService {
				cfg := readyPOSConfig()
				return withSellStock(sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{}))
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "", Amount: 200}},
			},
			wantErr: ErrPaymentTotal,
		},
		{
			name: "rejects item zero",
			setup: func() POSService {
				cfg := readyPOSConfig()
				return withSellStock(sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{}))
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 0, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrOrderProduct,
		},
		{
			name: "rejects invalid qty",
			setup: func() POSService {
				cfg := readyPOSConfig()
				return withSellStock(sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{}))
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: -1}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrOrderQty,
		},
		{
			name: "rejects invalid discount",
			setup: func() POSService {
				cfg := readyPOSConfig()
				return withSellStock(sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{}))
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2, DiscountPct: 120}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrOrderDiscount,
		},
		{
			name: "rejects unknown tax",
			setup: func() POSService {
				cfg := readyPOSConfig()
				svc := withSellStock(sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{}))
				svc.taxes = dao.CRUDMock[reference.Tax]{}
				return svc
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2, TaxIDs: helper.Int64Array{9}}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrOrderTax,
		},
		{
			name: "rejects purchase tax scope",
			setup: func() POSService {
				cfg := readyPOSConfig()
				svc := withSellStock(sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{}))
				svc.taxes = dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopePurchase, Amount: helper.Ptr(10.0)}, nil
					},
				}
				return svc
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2, TaxIDs: helper.Int64Array{9}}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrOrderTax,
		},
		{
			name: "propagates stock location error",
			setup: func() POSService {
				cfg := readyPOSConfig()
				return sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{})
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrOrderNoStock,
		},
		{
			name: "propagates customer location error",
			setup: func() POSService {
				cfg := readyPOSConfig()
				svc := sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{})
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
				return svc
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: ErrOrderNoStock,
		},
		{
			name: "propagates quant error",
			setup: func() POSService {
				cfg := readyPOSConfig()
				svc := sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{})
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
						return nil, errBoom
					},
				}
				return svc
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: errBoom,
		},
		{
			name: "propagates post stock movements error",
			setup: func() POSService {
				cfg := readyPOSConfig()
				svc := withSellStock(sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{}))
				svc.shipments = inventory.ShipmentDAOMock{
					CreateWithMovementsFunc: func(_ context.Context, _ *inventory.Shipment, _ []*inventory.StockMovement) (*inventory.Shipment, error) {
						return nil, errBoom
					},
				}
				return svc
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: errBoom,
		},
		{
			name: "propagates ship error",
			setup: func() POSService {
				cfg := readyPOSConfig()
				svc := withSellStock(sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{}))
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
						return nil, errBoom
					},
				}
				return svc
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: errBoom,
		},
		{
			name: "propagates shipment update error",
			setup: func() POSService {
				cfg := readyPOSConfig()
				svc := withSellStock(sellReadyService(&cfg, dao.CRUDMock[reference.POSConfig]{}))
				svc.shipments = inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						UpdateFunc: func(_ context.Context, _ *inventory.Shipment) (*inventory.Shipment, error) {
							return nil, errBoom
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
				return svc
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 200}},
			},
			wantErr: errBoom,
		},
		{
			name: "posts payments to mapped accounts",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1, State: SessionStateOpened}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
				}
				orders := POSOrderDAOMock{
					CreateWithLinesAndPaymentsFunc: func(_ context.Context, order *POSOrder, _ []*POSOrderLine, _ []*POSPayment) (*POSOrder, error) {
						order.ID = 10
						return order, nil
					},
				}
				locations := inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
							if q.Filters[0].Field == "usage" {
								return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
							}
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
						},
					},
				}
				quants := inventory.StockBalanceDAOMock{
					FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
						return &inventory.StockBalance{Base: model.Base{ID: 3}, Quantity: 10}, nil
					},
				}
				shipments := inventory.ShipmentDAOMock{
					CreateWithMovementsFunc: func(_ context.Context, shipment *inventory.Shipment, movements []*inventory.StockMovement) (*inventory.Shipment, error) {
						shipment.ID = 20
						for i := range movements {
							movements[i].ID = uint64(30 + i)
						}
						return shipment, nil
					},
				}
				payMethods := PaymentAccountLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.POSPaymentAccount], error) {
						return &query.Page[reference.POSPaymentAccount]{Items: []*reference.POSPaymentAccount{
							{OrganizationID: organizationID, Method: "card", AccountID: 70},
						}}, nil
					},
				}
				svc := testPOSService(configs, sessions, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
				svc.payMethods = payMethods
				svc.locations = locations
				svc.quants = quants
				svc.shipments = shipments
				svc.valuer = ShipEngineMock{
					ShipFunc: func(_ context.Context, moveID uint64, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
						return &inventory.CostLayer{MovementID: &moveID}, nil
					},
				}
				svc.poster = inventory.PosterMock{
					PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
						if request.OriginType == originPOSOrder {
							posted = &request
						}
						return &accounting.JournalEntry{}, nil
					},
				}
				return svc
			},
			req: SellRequest{
				SessionID: 1,
				Lines:     []SellLineRequest{{ItemID: 100, Qty: 2}},
				Payments:  []SellPaymentRequest{{Method: "cash", Amount: 100}, {Method: "card", Amount: 100}},
			},
			validate: func(t *testing.T, _ POSService, result *POSOrder) {
				if posted == nil {
					t.Fatal("expected a posted revenue movement")
				}
				debits := map[uint64]float64{}
				for _, line := range posted.Lines {
					if line.Debit.Float64() > 0 {
						debits[line.AccountID] = line.Debit.Float64()
					}
				}
				if len(debits) != 2 || debits[40] != 100 || debits[70] != 100 {
					t.Errorf("debit lines = %v, want cash 100 on 40 and card 100 on 70", debits)
				}
			},
		},
		{
			name: "with contact generates invoice and shift",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1, State: SessionStateOpened}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &config, nil },
				}
				orders := POSOrderDAOMock{
					CreateWithLinesAndPaymentsFunc: func(_ context.Context, order *POSOrder, _ []*POSOrderLine, _ []*POSPayment) (*POSOrder, error) {
						order.ID = 10
						return order, nil
					},
				}
				orderLines := POSOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
						return []*POSOrderLine{{
							Base: model.Base{ID: 11}, ItemID: helper.Ptr(uint64(100)), Qty: 2,
							UnitPrice: amount.FromFloat64(100), TaxIDs: helper.Int64Array{9}, PriceSubtotal: amount.FromFloat64(200), PriceTax: amount.FromFloat64(20), PriceTotal: amount.FromFloat64(220),
						}}, nil
					},
				}
				locations := inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
							if q.Filters[0].Field == "usage" {
								return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
							}
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
						},
					},
				}
				quants := inventory.StockBalanceDAOMock{
					FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
						return &inventory.StockBalance{Base: model.Base{ID: 3}, Quantity: 10}, nil
					},
				}
				svc := testPOSService(configs, sessions, orders, orderLines, POSPaymentDAOMock{})
				svc.locations = locations
				svc.quants = quants
				svc.poster = inventory.PosterMock{
					PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
						if request.OriginType == "pos_order_invoice_shift" {
							shiftPosted = true
						}
						return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
					},
				}
				svc.accounts = AccountLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
						return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 50}, OrganizationID: organizationID, Type: "receivable", Active: true}}}, nil
					},
				}
				svc.invoice = InvoiceEngineMock{
					CreateFunc: func(_ context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
						return &accounting.Invoice{Base: model.Base{ID: 77}, ContactID: request.ContactID}, nil
					},
				}
				svc.journals = dao.CRUDMock[reference.Journal]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
						return &reference.Journal{Base: model.Base{ID: 2}, DefaultAccountID: helper.Ptr(uint64(40))}, nil
					},
				}
				return svc
			},
			req: SellRequest{
				SessionID: 1, ContactID: helper.Ptr(uint64(5)),
				Lines:    []SellLineRequest{{ItemID: 100, Qty: 2, TaxIDs: helper.Int64Array{9}}},
				Payments: []SellPaymentRequest{{Method: "cash", Amount: 220}},
			},
			validate: func(t *testing.T, _ POSService, result *POSOrder) {
				if result.InvoiceID == nil || *result.InvoiceID != 77 {
					t.Errorf("invoice_id = %v, want 77", result.InvoiceID)
				}
				if !shiftPosted {
					t.Error("revenue shift movement must be posted when contact invoiced")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			result, err := svc.Sell(ctx, tt.req)
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, svc, result)
			}
		})
	}
}

func TestPOSService_CreateInvoice(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(10)
	contactID := uint64(5)

	tests := []struct {
		name     string
		setup    func() POSService
		wantErr  error
		validate func(t *testing.T, result *accounting.Invoice)
	}{
		{
			name: "rejects order without contact",
			setup: func() POSService {
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
						return &POSOrder{Base: model.Base{ID: 10}}, nil
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			wantErr: ErrOrderNotFound,
		},
		{
			name: "rejects already invoiced",
			setup: func() POSService {
				invoiceID := uint64(77)
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
						return &POSOrder{Base: model.Base{ID: 10}, ContactID: helper.Ptr(uint64(5)), InvoiceID: &invoiceID}, nil
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			wantErr: ErrOrderInvoiced,
		},
		{
			name: "generates invoice for contact order",
			setup: func() POSService {
				orders := POSOrderDAOMock{
					CRUDMock: dao.CRUDMock[POSOrder]{
						FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
							return &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, ContactID: &contactID, AmountTotal: amount.FromFloat64(220), Name: helper.Ptr("POS/00001")}, nil
						},
						UpdateFunc: func(_ context.Context, o *POSOrder) (*POSOrder, error) { return o, nil },
					},
				}
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 3}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return &reference.POSConfig{Base: model.Base{ID: 3}, OrganizationID: &organizationID}, nil
					},
				}
				lines := POSOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
						return []*POSOrderLine{{
							Base: model.Base{ID: 11}, ItemID: helper.Ptr(uint64(100)), Qty: 2,
							UnitPrice: amount.FromFloat64(100), TaxIDs: helper.Int64Array{9}, PriceSubtotal: amount.FromFloat64(200), PriceTax: amount.FromFloat64(20), PriceTotal: amount.FromFloat64(220),
						}}, nil
					},
				}
				invoice := InvoiceEngineMock{
					CreateFunc: func(_ context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
						return &accounting.Invoice{Base: model.Base{ID: 77}, ContactID: request.ContactID}, nil
					},
				}
				accounts := AccountLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
						return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 50}, OrganizationID: organizationID, Type: "receivable", Active: true}}}, nil
					},
				}

				svc := testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})
				svc.invoice = invoice
				svc.accounts = accounts
				return svc
			},
			validate: func(t *testing.T, result *accounting.Invoice) {
				if result.ID != 77 {
					t.Errorf("invoice id = %d, want 77", result.ID)
				}
			},
		},
		{
			name: "rejects missing session",
			setup: func() POSService {
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
						return &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, ContactID: &contactID}, nil
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			wantErr: ErrSessionNotFound,
		},
		{
			name: "rejects missing config",
			setup: func() POSService {
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
						return &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, ContactID: &contactID}, nil
					},
				}}
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 3}, nil
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			wantErr: ErrConfigNotFound,
		},
		{
			name: "propagates invoice error",
			setup: func() POSService {
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
						return &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, ContactID: &contactID, AmountTotal: amount.FromFloat64(220)}, nil
					},
				}}
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 3}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return &reference.POSConfig{Base: model.Base{ID: 3}, OrganizationID: &organizationID}, nil
					},
				}
				lines := POSOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
						return []*POSOrderLine{{ItemID: helper.Ptr(uint64(100)), Qty: 2, UnitPrice: amount.FromFloat64(100), TaxIDs: helper.Int64Array{9}, PriceSubtotal: amount.FromFloat64(200), PriceTax: amount.FromFloat64(20), PriceTotal: amount.FromFloat64(220)}}, nil
					},
				}
				svc := testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})
				svc.invoice = InvoiceEngineMock{
					CreateFunc: func(_ context.Context, _ accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
						return nil, errBoom
					},
				}
				return svc
			},
			wantErr: errBoom,
		},
		{
			name: "rejects missing receivable account",
			setup: func() POSService {
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
						return &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, ContactID: &contactID, AmountTotal: amount.FromFloat64(220)}, nil
					},
				}}
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 3}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return &reference.POSConfig{Base: model.Base{ID: 3}, OrganizationID: &organizationID}, nil
					},
				}
				lines := POSOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
						return []*POSOrderLine{{ItemID: helper.Ptr(uint64(100)), Qty: 2, UnitPrice: amount.FromFloat64(100), TaxIDs: helper.Int64Array{9}, PriceSubtotal: amount.FromFloat64(200), PriceTax: amount.FromFloat64(20), PriceTotal: amount.FromFloat64(220)}}, nil
					},
				}
				svc := testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})
				svc.invoice = InvoiceEngineMock{
					CreateFunc: func(_ context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
						return &accounting.Invoice{Base: model.Base{ID: 77}, ContactID: request.ContactID}, nil
					},
				}
				return svc
			},
			wantErr: ErrConfigJournal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			result, err := svc.CreateInvoice(ctx, 10, 2, time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC))
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestPOSService_Refund(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	config := reference.POSConfig{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		WarehouseID: helper.Ptr(uint64(4)), JournalID: helper.Ptr(uint64(2)), PriceBookID: helper.Ptr(uint64(3)),
	}

	tests := []struct {
		name     string
		setup    func() POSService
		wantErr  error
		validate func(t *testing.T, result *POSOrder)
	}{
		{
			name: "restock posts reversal movement and sets refunded",
			setup: func() POSService {
				order := &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, AmountTotal: amount.FromFloat64(220), AmountTax: amount.FromFloat64(20), State: OrderStateDone, Name: helper.Ptr("POS/00001")}
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, o *POSOrder) (*POSOrder, error) {
						order = o
						return o, nil
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
				orderLines := POSOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
						return []*POSOrderLine{{
							ItemID: helper.Ptr(uint64(100)), Qty: 2, PriceSubtotal: amount.FromFloat64(200), PriceTax: amount.FromFloat64(20), TaxIDs: helper.Int64Array{9},
						}}, nil
					},
				}
				locations := inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
							if q.Filters[0].Field == "usage" {
								return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
							}
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
						},
					},
				}
				movements := inventory.StockMovementDAOMock{
					CRUDMock: dao.CRUDMock[inventory.StockMovement]{
						CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
							movement.ID = 40
							return movement, nil
						},
					},
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
						movement.ID = 40
						return movement, nil
					},
					ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{{
							Base: model.Base{ID: 30}, ItemID: 100, Qty: 2,
							SrcLocationID: 10, State: inventory.MovementStateDone,
						}}, nil
					},
				}
				layers := inventory.CostLayerDAOMock{
					ListByMovementFunc: func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
						return []*inventory.CostLayer{{MovementID: helper.Ptr(uint64(30)), UnitCost: helper.Ptr(100.0)}}, nil
					},
				}
				poster := inventory.PosterMock{
					PostFunc: func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
						return &accounting.JournalEntry{}, nil
					},
				}

				svc := testPOSService(configs, sessions, orders, orderLines, POSPaymentDAOMock{})
				svc.locations = locations
				svc.movements = movements
				svc.layers = layers
				svc.valuer = ShipEngineMock{}
				svc.poster = poster
				return svc
			},
			validate: func(t *testing.T, result *POSOrder) {
				if result.State != OrderStateRefunded {
					t.Errorf("state = %s, want refunded", result.State)
				}
			},
		},
		{
			name: "rejects non-done order",
			setup: func() POSService {
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
						return &POSOrder{Base: model.Base{ID: 10}, State: OrderStateRefunded}, nil
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			wantErr: ErrOrderState,
		},
		{
			name: "with invoice generates credit note",
			setup: func() POSService {
				invoiceID := uint64(77)
				order := &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, AmountTotal: amount.FromFloat64(220), AmountTax: amount.FromFloat64(20), State: OrderStateDone, Name: helper.Ptr("POS/00001"), InvoiceID: &invoiceID}
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, o *POSOrder) (*POSOrder, error) {
						order = o
						return o, nil
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
						return []*POSOrderLine{{
							ItemID: helper.Ptr(uint64(100)), Qty: 2, PriceSubtotal: amount.FromFloat64(200), PriceTax: amount.FromFloat64(20), TaxIDs: helper.Int64Array{9},
						}}, nil
					},
				}
				locations := inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
							if q.Filters[0].Field == "usage" {
								return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
							}
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
						},
					},
				}
				movements := inventory.StockMovementDAOMock{
					CRUDMock: dao.CRUDMock[inventory.StockMovement]{
						CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
							movement.ID = 40
							return movement, nil
						},
					},
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
						movement.ID = 40
						return movement, nil
					},
					ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{{
							Base: model.Base{ID: 30}, ItemID: 100, Qty: 2,
							SrcLocationID: 10, State: inventory.MovementStateDone,
						}}, nil
					},
				}
				layers := inventory.CostLayerDAOMock{
					ListByMovementFunc: func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
						return []*inventory.CostLayer{{MovementID: helper.Ptr(uint64(30)), UnitCost: helper.Ptr(100.0)}}, nil
					},
				}
				valuer := ShipEngineMock{
					RestockFunc: func(_ context.Context, moveID uint64, _ amount.Amount, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
						return &inventory.CostLayer{MovementID: &moveID}, nil
					},
				}
				invoice := InvoiceEngineMock{
					CreateCreditNoteFunc: func(_ context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
						return &accounting.Invoice{Base: model.Base{ID: 78}, ContactID: request.OriginalInvoiceID}, nil
					},
				}
				poster := inventory.PosterMock{
					PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
						return &accounting.JournalEntry{}, nil
					},
				}

				svc := testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})
				svc.locations = locations
				svc.movements = movements
				svc.layers = layers
				svc.valuer = valuer
				svc.invoice = invoice
				svc.poster = poster
				return svc
			},
			validate: func(t *testing.T, result *POSOrder) {
				if result.State != OrderStateRefunded {
					t.Errorf("state = %s, want refunded", result.State)
				}
			},
		},
		{
			name: "propagates lines error",
			setup: func() POSService {
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
				cfg := readyPOSConfig()
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &cfg, nil },
				}
				lines := POSOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*POSOrderLine, error) {
						return nil, errBoom
					},
				}
				return testPOSService(configs, sessions, orders, lines, POSPaymentDAOMock{})
			},
			wantErr: errBoom,
		},
		{
			name: "rejects no lines",
			setup: func() POSService {
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
				cfg := readyPOSConfig()
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &cfg, nil },
				}
				return testPOSService(configs, sessions, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			wantErr: ErrOrderNoLines,
		},
		{
			name: "rejects missing config",
			setup: func() POSService {
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
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			wantErr: ErrConfigNotFound,
		},
		{
			name: "propagates restock movement error",
			setup: func() POSService {
				cfg := readyPOSConfig()
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
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &cfg, nil },
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
						return nil, errBoom
					},
				}
				return svc
			},
			wantErr: errBoom,
		},
		{
			name: "rejects missing original movement",
			setup: func() POSService {
				cfg := readyPOSConfig()
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
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &cfg, nil },
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
				return svc
			},
			wantErr: ErrOrderNoStock,
		},
		{
			name: "propagates movement create error",
			setup: func() POSService {
				cfg := readyPOSConfig()
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
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &cfg, nil },
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
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *inventory.StockMovement) (*inventory.StockMovement, error) {
						return nil, errBoom
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
				return svc
			},
			wantErr: errBoom,
		},
		{
			name: "rejects missing unit cost",
			setup: func() POSService {
				cfg := readyPOSConfig()
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
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &cfg, nil },
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
				return svc
			},
			wantErr: ErrOrderCost,
		},
		{
			name: "propagates journal error",
			setup: func() POSService {
				cfg := readyPOSConfig()
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
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &cfg, nil },
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
						return nil, errBoom
					},
				}
				return svc
			},
			wantErr: errBoom,
		},
		{
			name: "rejects journal without default account",
			setup: func() POSService {
				cfg := readyPOSConfig()
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
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &cfg, nil },
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
				return svc
			},
			wantErr: ErrConfigJournal,
		},
		{
			name: "propagates update error",
			setup: func() POSService {
				cfg := readyPOSConfig()
				order := &POSOrder{Base: model.Base{ID: 10}, SessionID: 1, AmountTotal: amount.FromFloat64(220), State: OrderStateDone}
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, _ *POSOrder) (*POSOrder, error) {
						return nil, errBoom
					},
				}}
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 1}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) { return &cfg, nil },
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
				return svc
			},
			wantErr: errBoom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			result, err := svc.Refund(ctx, 10, 2, time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC))
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestPOSService_SessionPaymentBreakdown(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		setup    func() POSService
		wantErr  error
		validate func(t *testing.T, result map[string]float64)
	}{
		{
			name: "totals per method",
			setup: func() POSService {
				payments := POSPaymentDAOMock{
					SumBySessionFunc: func(_ context.Context, _ uint64) ([]PaymentMethodTotal, error) {
						return []PaymentMethodTotal{{Method: "cash", Total: 170}, {Method: "card", Total: 50}}, nil
					},
				}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, payments)
			},
			validate: func(t *testing.T, result map[string]float64) {
				if result["cash"] != 170 || result["card"] != 50 {
					t.Errorf("breakdown = %+v, want cash 170 card 50", result)
				}
			},
		},
		{
			name: "propagates sum error",
			setup: func() POSService {
				payments := POSPaymentDAOMock{
					SumBySessionFunc: func(_ context.Context, _ uint64) ([]PaymentMethodTotal, error) {
						return nil, errBoom
					},
				}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, payments)
			},
			wantErr: errBoom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			result, err := svc.SessionPaymentBreakdown(ctx, 1)
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestPOSService_ListSessions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		setup    func() POSService
		orgID    uint64
		wantErr  error
		validate func(t *testing.T, result *query.Page[POSSession])
	}{
		{
			name: "scopes to organization",
			setup: func() POSService {
				sessions := POSSessionDAOMock{
					ListInOrganizationFunc: func(_ context.Context, q *query.Query, organizationID uint64) (*query.Page[POSSession], error) {
						return &query.Page[POSSession]{Items: []*POSSession{{Base: model.Base{ID: 1}, State: SessionStateOpened}}}, nil
					},
				}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID: 10,
			validate: func(t *testing.T, result *query.Page[POSSession]) {
				if len(result.Items) != 1 {
					t.Errorf("items = %d, want 1", len(result.Items))
				}
			},
		},
		{
			name: "passes through without organization",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					ListFunc: func(_ context.Context, q *query.Query) (*query.Page[POSSession], error) {
						return &query.Page[POSSession]{Items: []*POSSession{{Base: model.Base{ID: 1}}}}, nil
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID: 0,
			validate: func(t *testing.T, result *query.Page[POSSession]) {
				if len(result.Items) != 1 {
					t.Errorf("items = %d, want 1", len(result.Items))
				}
			},
		},
		{
			name: "propagates list error",
			setup: func() POSService {
				sessions := POSSessionDAOMock{
					ListInOrganizationFunc: func(_ context.Context, _ *query.Query, _ uint64) (*query.Page[POSSession], error) {
						return nil, errBoom
					},
				}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID:   10,
			wantErr: errBoom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			result, err := svc.ListSessions(ctx, tt.orgID, &query.Query{})
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestPOSService_GetSession(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		setup    func() POSService
		orgID    uint64
		wantErr  error
		validate func(t *testing.T, result *POSSession)
	}{
		{
			name: "returns scoped session",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 3}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return &reference.POSConfig{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(10))}, nil
					},
				}
				return testPOSService(configs, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID: 10,
			validate: func(t *testing.T, result *POSSession) {
				if result.ID != 1 || result.ConfigID != 3 {
					t.Errorf("session = %+v, want id 1 config 3", result)
				}
			},
		},
		{
			name: "returns directly without organization",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, State: SessionStateOpened}, nil
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID: 0,
			validate: func(t *testing.T, result *POSSession) {
				if result.State != SessionStateOpened {
					t.Errorf("session = %+v, want opened", result)
				}
			},
		},
		{
			name: "rejects organization mismatch",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 3}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return &reference.POSConfig{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(99))}, nil
					},
				}
				return testPOSService(configs, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID:   10,
			wantErr: ErrSessionNotFound,
		},
		{
			name: "rejects missing config",
			setup: func() POSService {
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 3}, nil
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, sessions, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID:   10,
			wantErr: ErrSessionNotFound,
		},
		{
			name: "rejects missing session",
			setup: func() POSService {
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID:   10,
			wantErr: ErrSessionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			result, err := svc.GetSession(ctx, tt.orgID, 1)
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestPOSService_GetOrder(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		setup    func() POSService
		orgID    uint64
		wantErr  error
		validate func(t *testing.T, result *POSOrder)
	}{
		{
			name: "returns scoped order",
			setup: func() POSService {
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
						return &POSOrder{Base: model.Base{ID: 9}, SessionID: 1, State: OrderStateDone}, nil
					},
				}}
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 3}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return &reference.POSConfig{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(10))}, nil
					},
				}
				return testPOSService(configs, sessions, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID: 10,
			validate: func(t *testing.T, result *POSOrder) {
				if result.ID != 9 || result.State != OrderStateDone {
					t.Errorf("order = %+v, want done order 9", result)
				}
			},
		},
		{
			name: "rejects organization mismatch",
			setup: func() POSService {
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
						return &POSOrder{Base: model.Base{ID: 9}, SessionID: 1}, nil
					},
				}}
				sessions := POSSessionDAOMock{CRUDMock: dao.CRUDMock[POSSession]{
					FindFunc: func(_ context.Context, _ uint64) (*POSSession, error) {
						return &POSSession{Base: model.Base{ID: 1}, ConfigID: 3}, nil
					},
				}}
				configs := dao.CRUDMock[reference.POSConfig]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.POSConfig, error) {
						return &reference.POSConfig{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(99))}, nil
					},
				}
				return testPOSService(configs, sessions, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID:   10,
			wantErr: ErrOrderNotFound,
		},
		{
			name: "rejects missing order",
			setup: func() POSService {
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, POSOrderDAOMock{}, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID:   10,
			wantErr: ErrOrderNotFound,
		},
		{
			name: "propagates find error",
			setup: func() POSService {
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*POSOrder, error) {
						return nil, errBoom
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID:   10,
			wantErr: errBoom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			result, err := svc.GetOrder(ctx, tt.orgID, 9)
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestPOSService_ListOrders(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		setup    func() POSService
		orgID    uint64
		wantErr  error
		validate func(t *testing.T, result *query.Page[POSOrder])
	}{
		{
			name: "returns scoped orders",
			setup: func() POSService {
				orders := POSOrderDAOMock{
					ListInOrganizationFunc: func(_ context.Context, q *query.Query, organizationID uint64) (*query.Page[POSOrder], error) {
						return &query.Page[POSOrder]{Items: []*POSOrder{{Base: model.Base{ID: 9}, State: OrderStateDone}}}, nil
					},
				}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID: 10,
			validate: func(t *testing.T, result *query.Page[POSOrder]) {
				if len(result.Items) != 1 || result.Items[0].ID != 9 {
					t.Errorf("items = %+v, want one done order", result.Items)
				}
			},
		},
		{
			name: "passes through without organization",
			setup: func() POSService {
				orders := POSOrderDAOMock{CRUDMock: dao.CRUDMock[POSOrder]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[POSOrder], error) {
						return &query.Page[POSOrder]{}, nil
					},
				}}
				return testPOSService(dao.CRUDMock[reference.POSConfig]{}, POSSessionDAOMock{}, orders, POSOrderLineDAOMock{}, POSPaymentDAOMock{})
			},
			orgID: 0,
			validate: func(t *testing.T, _ *query.Page[POSOrder]) {
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			result, err := svc.ListOrders(ctx, tt.orgID, &query.Query{})
			if tt.wantErr != nil {
				helper.AssertError(t, err, true, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func testPOSService(configs dao.CRUDMock[reference.POSConfig], sessions POSSessionDAOMock, orders POSOrderDAOMock, lines POSOrderLineDAOMock, payments POSPaymentDAOMock) POSService {
	sequences := sequence.NewSequenceService(sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "POS/00001"}, nil
		},
	})
	svc := NewPOSService(
		configs,
		sessions,
		orders,
		lines,
		payments,
		testProductService(),
		dao.CRUDMock[reference.Tax]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
				return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, Amount: helper.Ptr(10.0), TaxAccountID: helper.Ptr(uint64(60))}, nil
			},
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Tax], error) {
				return &query.Page[reference.Tax]{Items: []*reference.Tax{{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, Amount: helper.Ptr(10.0), TaxAccountID: helper.Ptr(uint64(60))}}}, nil
			},
		},
		dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 2}, DefaultAccountID: helper.Ptr(uint64(40))}, nil
			},
		},
		AccountLookupMock{},
		PaymentAccountLookupMock{},
		inventory.StockLocationDAOMock{},
		inventory.StockBalanceDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.CostLayerDAOMock{},
		inventory.ShipmentDAOMock{},
		ShipEngineMock{},
		InvoiceEngineMock{},
		inventory.PosterMock{},
		iam.MemberDAOMock{
			FindByUserAndOrganizationFunc: func(_ context.Context, _, _ uint64) (*iam.Member, error) {
				return &iam.Member{Base: model.Base{ID: 1}}, nil
			},
		},
		sequences,
		TransactionerMock{},
	)
	return svc
}

func testProductService() products.ProductService {
	return products.NewProductService(
		products.ItemDAOMock{
			CRUDMock: dao.CRUDMock[products.Item]{
				FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
					return &products.Item{Base: model.Base{ID: 1}, ListPrice: 100, CategoryID: helper.Ptr(uint64(8))}, nil
				},
			},
		},
		products.ItemVariantDAOMock{
			CRUDMock: dao.CRUDMock[products.ItemVariant]{
				FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
					return &products.ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
				},
			},
		},
		dao.CRUDMock[reference.ItemCategory]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
				return &reference.ItemCategory{Base: model.Base{ID: 8}, IncomeAccountID: helper.Ptr(uint64(70))}, nil
			},
		},
		products.PriceBookDAOMock{
			CRUDMock: dao.CRUDMock[products.PriceBook]{
				FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
					return &products.PriceBook{Base: model.Base{ID: 3}, Name: "Retail"}, nil
				},
			},
		},
		products.PriceRuleDAOMock{},
	)
}
