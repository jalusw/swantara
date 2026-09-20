package interorganization_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/sales"
)

func TestInterorganizationService_CreateRule_RejectsMissingOrganizations(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	_, err := svc.CreateRule(context.Background(), interorganization.UpsertInterorganizationRuleRequest{FromOrganizationID: nil})
	if err != interorganization.ErrRuleRequired {
		t.Errorf("err = %v, want ErrRuleRequired", err)
	}
}

func TestInterorganizationService_CreateRule_RejectsMissingContacts(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	_, err := svc.CreateRule(context.Background(), interorganization.UpsertInterorganizationRuleRequest{
		FromOrganizationID: helper.Ptr(uint64(10)),
		ToOrganizationID:   helper.Ptr(uint64(20)),
		AutoMirror:         true,
	})
	if err != interorganization.ErrRuleContactsRequired {
		t.Errorf("err = %v, want ErrRuleContactsRequired", err)
	}
}

func TestInterorganizationService_CreateRule_CreatesRule(t *testing.T) {
	var created *interorganization.InterorganizationRule
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
					return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{}}, nil
				},
				CreateFunc: func(_ context.Context, rule *interorganization.InterorganizationRule) (*interorganization.InterorganizationRule, error) {
					created = rule
					return rule, nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	rule, err := svc.CreateRule(context.Background(), interorganization.UpsertInterorganizationRuleRequest{
		FromOrganizationID: helper.Ptr(uint64(10)),
		ToOrganizationID:   helper.Ptr(uint64(20)),
		AutoMirror:         true,
		SupplierContactID:  helper.Ptr(uint64(777)),
		CustomerContactID:  helper.Ptr(uint64(888)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created == nil || created != rule {
		t.Errorf("created rule not recorded")
	}
	if *created.FromOrganizationID != 10 || *created.ToOrganizationID != 20 || !created.AutoMirror {
		t.Errorf("rule = %+v, want org 10 to 20 mirrored", created)
	}
}

func TestInterorganizationService_CreateRule_PropagatesListError(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
					return nil, errors.New("db down")
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
	})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestInterorganizationService_UpdateRule_RejectsMissingID(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	_, err := svc.UpdateRule(context.Background(), interorganization.UpsertInterorganizationRuleRequest{})
	if err != interorganization.ErrRuleNotFound {
		t.Errorf("err = %v, want ErrRuleNotFound", err)
	}
}

func TestInterorganizationService_UpdateRule_ReturnsNotFound(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	_, err := svc.UpdateRule(context.Background(), interorganization.UpsertInterorganizationRuleRequest{ID: helper.Ptr(uint64(5))})
	if err != interorganization.ErrRuleNotFound {
		t.Errorf("err = %v, want ErrRuleNotFound", err)
	}
}

func TestInterorganizationService_UpdateRule_PropagatesFindError(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
					return nil, errors.New("db down")
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	_, err := svc.UpdateRule(context.Background(), interorganization.UpsertInterorganizationRuleRequest{ID: helper.Ptr(uint64(5))})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestInterorganizationService_UpdateRule_RejectsSameOrganization(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
					return &interorganization.InterorganizationRule{Base: baseID(5), FromOrganizationID: helper.Ptr(uint64(10)), ToOrganizationID: helper.Ptr(uint64(20))}, nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	_, err := svc.UpdateRule(context.Background(), interorganization.UpsertInterorganizationRuleRequest{
		ID:                 helper.Ptr(uint64(5)),
		FromOrganizationID: helper.Ptr(uint64(20)),
		ToOrganizationID:   helper.Ptr(uint64(20)),
	})
	if err != interorganization.ErrRuleSameOrganization {
		t.Errorf("err = %v, want ErrRuleSameOrganization", err)
	}
}

func TestInterorganizationService_UpdateRule_RejectsMissingContacts(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
					return &interorganization.InterorganizationRule{Base: baseID(5)}, nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	_, err := svc.UpdateRule(context.Background(), interorganization.UpsertInterorganizationRuleRequest{
		ID:         helper.Ptr(uint64(5)),
		AutoMirror: true,
	})
	if err != interorganization.ErrRuleContactsRequired {
		t.Errorf("err = %v, want ErrRuleContactsRequired", err)
	}
}

func TestInterorganizationService_UpdateRule_UpdatesProvidedFields(t *testing.T) {
	var updated *interorganization.InterorganizationRule
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
					return &interorganization.InterorganizationRule{Base: baseID(5), FromOrganizationID: helper.Ptr(uint64(10)), ToOrganizationID: helper.Ptr(uint64(20)), AutoMirror: false, SupplierContactID: helper.Ptr(uint64(1)), CustomerContactID: helper.Ptr(uint64(2))}, nil
				},
				UpdateFunc: func(_ context.Context, rule *interorganization.InterorganizationRule) (*interorganization.InterorganizationRule, error) {
					updated = rule
					return rule, nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	rule, err := svc.UpdateRule(context.Background(), interorganization.UpsertInterorganizationRuleRequest{
		ID:                 helper.Ptr(uint64(5)),
		FromOrganizationID: helper.Ptr(uint64(30)),
		ToOrganizationID:   helper.Ptr(uint64(40)),
		AutoMirror:         true,
		SupplierContactID:  helper.Ptr(uint64(777)),
		CustomerContactID:  helper.Ptr(uint64(888)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated == nil || updated != rule {
		t.Errorf("updated rule not recorded")
	}
	if *updated.FromOrganizationID != 30 || *updated.ToOrganizationID != 40 || !updated.AutoMirror || *updated.SupplierContactID != 777 || *updated.CustomerContactID != 888 {
		t.Errorf("rule = %+v, want org 30 to 40 mirrored with new contacts", updated)
	}
}

func TestInterorganizationService_DeleteRule_DeletesRule(t *testing.T) {
	var deletedID uint64
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				FindFunc: func(_ context.Context, id uint64) (*interorganization.InterorganizationRule, error) {
					return &interorganization.InterorganizationRule{Base: baseID(id)}, nil
				},
				DeleteFunc: func(_ context.Context, id uint64) error {
					deletedID = id
					return nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	err := svc.DeleteRule(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedID != 5 {
		t.Errorf("deleted id = %d, want 5", deletedID)
	}
}

func TestInterorganizationService_DeleteRule_ReturnsNotFound(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	err := svc.DeleteRule(context.Background(), 5)
	if err != interorganization.ErrRuleNotFound {
		t.Errorf("err = %v, want ErrRuleNotFound", err)
	}
}

func TestInterorganizationService_DeleteRule_PropagatesFindError(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
					return nil, errors.New("db down")
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	err := svc.DeleteRule(context.Background(), 5)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestInterorganizationService_ListRules_ListsRules(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
					return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: baseID(5), FromOrganizationID: helper.Ptr(uint64(10))}}}, nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	items, err := svc.ListRules(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || *items[0].FromOrganizationID != 10 {
		t.Errorf("items = %+v, want one rule for org 10", items)
	}
}

func TestInterorganizationService_ListRules_PropagatesListError(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
					return nil, errors.New("db down")
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	_, err := svc.ListRules(context.Background(), 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestInterorganizationService_ListTransactions_ListsTransactions(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
					return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(400), SourceOrganizationID: helper.Ptr(uint64(10))}}}, nil
				},
			},
		},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	items, err := svc.ListTransactions(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || *items[0].SourceOrganizationID != 10 {
		t.Errorf("items = %+v, want one transaction for org 10", items)
	}
}

func TestInterorganizationService_ListTransactions_PropagatesListError(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
					return nil, errors.New("db down")
				},
			},
		},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{},
		sales.SaleOrderLineDAOMock{},
	)

	_, err := svc.ListTransactions(context.Background(), 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestInterorganizationService_MirrorSaleOrder_PropagatesFindError(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return nil, errors.New("db down")
		}}},
		sales.SaleOrderLineDAOMock{},
	)

	_, _, err := svc.MirrorSaleOrder(context.Background(), interorganization.MirrorSaleOrderRequest{SaleOrderID: 30, ToOrganizationID: 20})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestInterorganizationService_MirrorSaleOrder_SourceNotFound(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{},
	)

	_, _, err := svc.MirrorSaleOrder(context.Background(), interorganization.MirrorSaleOrderRequest{SaleOrderID: 30, ToOrganizationID: 20})
	if err != interorganization.ErrMirrorSourceNotFound {
		t.Errorf("err = %v, want ErrMirrorSourceNotFound", err)
	}
}

func TestInterorganizationService_MirrorSaleOrder_SourceNotPosted(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateDraft}, nil
		}}},
		sales.SaleOrderLineDAOMock{},
	)

	_, _, err := svc.MirrorSaleOrder(context.Background(), interorganization.MirrorSaleOrderRequest{SaleOrderID: 30, ToOrganizationID: 20})
	if err != interorganization.ErrMirrorSourceNotPosted {
		t.Errorf("err = %v, want ErrMirrorSourceNotPosted", err)
	}
}

func TestInterorganizationService_MirrorSaleOrder_PropagatesRuleListError(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
					return nil, errors.New("db down")
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{},
	)

	_, _, err := svc.MirrorSaleOrder(context.Background(), interorganization.MirrorSaleOrderRequest{SaleOrderID: 30, ToOrganizationID: 20})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestInterorganizationService_MirrorSaleOrder_RuleDisabled(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
					return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: baseID(5), FromOrganizationID: helper.Ptr(uint64(10)), ToOrganizationID: helper.Ptr(uint64(20)), AutoMirror: false, SupplierContactID: helper.Ptr(uint64(777))}}}, nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{},
	)

	_, _, err := svc.MirrorSaleOrder(context.Background(), interorganization.MirrorSaleOrderRequest{SaleOrderID: 30, ToOrganizationID: 20})
	if err != interorganization.ErrMirrorRuleDisabled {
		t.Errorf("err = %v, want ErrMirrorRuleDisabled", err)
	}
}

func TestInterorganizationService_MirrorSaleOrder_ContactsRequired(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
					return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: baseID(5), FromOrganizationID: helper.Ptr(uint64(10)), ToOrganizationID: helper.Ptr(uint64(20)), AutoMirror: true}}}, nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{},
	)

	_, _, err := svc.MirrorSaleOrder(context.Background(), interorganization.MirrorSaleOrderRequest{SaleOrderID: 30, ToOrganizationID: 20})
	if err != interorganization.ErrRuleContactsRequired {
		t.Errorf("err = %v, want ErrRuleContactsRequired", err)
	}
}

func TestInterorganizationService_MirrorSaleOrder_PropagatesLineError(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
					return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: baseID(5), FromOrganizationID: helper.Ptr(uint64(10)), ToOrganizationID: helper.Ptr(uint64(20)), AutoMirror: true, SupplierContactID: helper.Ptr(uint64(777))}}}, nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return nil, errors.New("db down")
		}},
	)

	_, _, err := svc.MirrorSaleOrder(context.Background(), interorganization.MirrorSaleOrderRequest{SaleOrderID: 30, ToOrganizationID: 20})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestInterorganizationService_MirrorSaleOrder_SkipsInvalidLinesAndReturnsNoLines(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
					return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: baseID(5), FromOrganizationID: helper.Ptr(uint64(10)), ToOrganizationID: helper.Ptr(uint64(20)), AutoMirror: true, SupplierContactID: helper.Ptr(uint64(777))}}}, nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: baseID(300), QtyOrdered: 0}}, nil
		}},
	)

	_, _, err := svc.MirrorSaleOrder(context.Background(), interorganization.MirrorSaleOrderRequest{SaleOrderID: 30, ToOrganizationID: 20})
	if err != interorganization.ErrMirrorSourceNoLines {
		t.Errorf("err = %v, want ErrMirrorSourceNoLines", err)
	}
}

func TestInterorganizationService_MirrorSaleOrder_PropagatesCreateError(t *testing.T) {
	svc := newInterorganizationServiceForTest(
		interorganization.InterorganizationRuleDAOMock{
			CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
					return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: baseID(5), FromOrganizationID: helper.Ptr(uint64(10)), ToOrganizationID: helper.Ptr(uint64(20)), AutoMirror: true, SupplierContactID: helper.Ptr(uint64(777))}}}, nil
				},
			},
		},
		interorganization.InterorganizationTransactionDAOMock{},
		purchaseOrderCreatorMock{
			CreateFunc: func(_ context.Context, _ *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
				return nil, errors.New("create failed")
			},
		},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: baseID(300), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}}, nil
		}},
	)

	_, _, err := svc.MirrorSaleOrder(context.Background(), interorganization.MirrorSaleOrderRequest{SaleOrderID: 30, ToOrganizationID: 20})
	if err == nil || err.Error() != "create failed" {
		t.Errorf("err = %v, want create failed", err)
	}
}

func TestInterorganizationService_MirrorSaleOrder_PropagatesTransactionError(t *testing.T) {
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
				CreateFunc: func(_ context.Context, _ *interorganization.InterorganizationTransaction) (*interorganization.InterorganizationTransaction, error) {
					return nil, errors.New("insert failed")
				},
			},
		},
		purchaseOrderCreatorMock{},
		sales.SaleOrderDAOMock{CRUDMock: dao.CRUDMock[sales.SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: baseID(30), OrganizationID: helper.Ptr(uint64(10)), State: sales.OrderStateConfirmed}, nil
		}}},
		sales.SaleOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: baseID(300), ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}}, nil
		}},
	)

	_, _, err := svc.MirrorSaleOrder(context.Background(), interorganization.MirrorSaleOrderRequest{SaleOrderID: 30, ToOrganizationID: 20})
	if err == nil || err.Error() != "insert failed" {
		t.Errorf("err = %v, want insert failed", err)
	}
}
