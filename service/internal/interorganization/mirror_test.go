package interorganization_test

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/sales"
)

func newInterorganizationServiceForTest(
	rules interorganization.InterorganizationRuleDAO,
	trans interorganization.InterorganizationTransactionDAO,
	poCreate interorganization.PurchaseOrderCreator,
	soOrders sales.SaleOrderDAO,
	soLines sales.SaleOrderLineDAO,
) interorganization.InterorganizationService {
	return interorganization.NewInterorganizationService(rules, trans, poCreate, soOrders, soLines)
}

func TestInterorganizationService_CreateRule_RejectsSameOrganization(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	_, err := svc.CreateRule(context.Background(), interorganization.UpsertInterorganizationRuleRequest{
		FromOrganizationID: helper.Ptr(uint64(10)),
		ToOrganizationID:   helper.Ptr(uint64(10)),
	})
	if err != interorganization.ErrRuleSameOrganization {
		t.Errorf("err = %v, want ErrRuleSameOrganization", err)
	}
}

func TestInterorganizationService_CreateRule_RejectsDuplicatePair(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
					return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: baseID(5), FromOrganizationID: helper.Ptr(uint64(10)), ToOrganizationID: helper.Ptr(uint64(20))}}}, nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	_, err := svc.CreateRule(context.Background(), interorganization.UpsertInterorganizationRuleRequest{
		FromOrganizationID: helper.Ptr(uint64(10)),
		ToOrganizationID:   helper.Ptr(uint64(20)),
		AutoMirror:         true,
		SupplierContactID:  helper.Ptr(uint64(777)),
		CustomerContactID:  helper.Ptr(uint64(888)),
	})
	if err != interorganization.ErrRuleDuplicate {
		t.Errorf("err = %v, want ErrRuleDuplicate", err)
	}
}

func TestInterorganizationService_MirrorSaleOrder_CreatesMirrorPurchaseOrder(t *testing.T) {
	var mirroredOrder *procurement.PurchaseOrder
	var mirroredLines []*procurement.PurchaseOrderLine
	poCreate := purchaseOrderCreatorMock{
		CreateFunc: func(_ context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			mirroredOrder = order
			mirroredLines = lines
			return &procurement.PurchaseOrder{Base: baseID(200), OrganizationID: helper.Ptr(uint64(20)), SupplierID: order.SupplierID, AmountTotal: 200, State: procurement.PurchaseOrderStateDraft}, nil
		},
	}
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
					return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: baseID(5), FromOrganizationID: helper.Ptr(uint64(10)), ToOrganizationID: helper.Ptr(uint64(20)), AutoMirror: true, SupplierContactID: helper.Ptr(uint64(777))}}}, nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
				CreateFunc: func(_ context.Context, transaction *interorganization.InterorganizationTransaction) (*interorganization.InterorganizationTransaction, error) {
					return &interorganization.InterorganizationTransaction{Base: baseID(400), SourceOrganizationID: transaction.SourceOrganizationID, SourceType: transaction.SourceType, SourceID: transaction.SourceID, MirrorOrganizationID: transaction.MirrorOrganizationID, MirrorType: transaction.MirrorType, MirrorID: transaction.MirrorID, Amount: transaction.Amount, State: transaction.State}, nil
				},
			},
		},
		poCreate,
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: baseID(300), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, UnitPrice: 100}}, nil
		}},
	)

	po, transaction, err := svc.MirrorSaleOrder(context.Background(), interorganization.MirrorSaleOrderRequest{
		SaleOrderID:      30,
		ToOrganizationID: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mirroredOrder == nil || *mirroredOrder.OrganizationID != 20 || mirroredOrder.SupplierID != 777 {
		t.Errorf("mirrored order = %+v, want org 20 supplier 777", mirroredOrder)
	}
	if len(mirroredLines) != 1 || mirroredLines[0].QtyOrdered != 2 || mirroredLines[0].UnitPrice != 100 {
		t.Errorf("mirrored lines = %+v, want qty 2 price 100", mirroredLines)
	}
	if po.ID != 200 {
		t.Errorf("po id = %d, want 200", po.ID)
	}
	if transaction == nil || transaction.SourceType != "sale_order" || transaction.MirrorType != "purchase_order" || transaction.Amount != 200 {
		t.Errorf("transaction = %+v, want sale_order->purchase_order amount 200", transaction)
	}
}

func TestInterorganizationService_MirrorSaleOrder_NoRuleRejected(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{},
	)

	_, _, err := svc.MirrorSaleOrder(context.Background(), interorganization.MirrorSaleOrderRequest{
		SaleOrderID:      30,
		ToOrganizationID: 20,
	})
	if err != interorganization.ErrMirrorRuleMissing {
		t.Errorf("err = %v, want ErrMirrorRuleMissing", err)
	}
}
