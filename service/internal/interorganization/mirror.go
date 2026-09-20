package interorganization

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/sales"
)

type InterorganizationService struct {
	rules       InterorganizationRuleDAO
	trans       InterorganizationTransactionDAO
	poCreate    PurchaseOrderCreator
	soOrders    sales.SaleOrderDAO
	soLines     sales.SaleOrderLineDAO
	invInvoices accounting.InvoiceDAO
	invLines    accounting.InvoiceLineDAO
	invBills    SupplierBillCreator
}

func NewInterorganizationService(
	rules InterorganizationRuleDAO,
	trans InterorganizationTransactionDAO,
	poCreate PurchaseOrderCreator,
	soOrders sales.SaleOrderDAO,
	soLines sales.SaleOrderLineDAO,
) InterorganizationService {
	return InterorganizationService{
		rules:    rules,
		trans:    trans,
		poCreate: poCreate,
		soOrders: soOrders,
		soLines:  soLines,
	}
}

type UpsertInterorganizationRuleRequest struct {
	ID                 *uint64
	FromOrganizationID *uint64
	ToOrganizationID   *uint64
	AutoMirror         bool
	SupplierContactID  *uint64
	CustomerContactID  *uint64
}

func (s InterorganizationService) CreateRule(ctx context.Context, request UpsertInterorganizationRuleRequest) (*InterorganizationRule, error) {
	if request.FromOrganizationID == nil || request.ToOrganizationID == nil {
		return nil, ErrRuleRequired
	}
	if *request.FromOrganizationID == *request.ToOrganizationID {
		return nil, ErrRuleSameOrganization
	}
	if request.AutoMirror && (request.SupplierContactID == nil || request.CustomerContactID == nil) {
		return nil, ErrRuleContactsRequired
	}
	existing, err := s.rules.List(ctx, &query.Query{Filters: []query.Filter{{Field: "from_organization_id", Operator: query.Equal, Value: *request.FromOrganizationID}}})
	if err != nil {
		return nil, err
	}
	for _, rule := range existing.Items {
		if rule.ToOrganizationID != nil && *rule.ToOrganizationID == *request.ToOrganizationID {
			return nil, ErrRuleDuplicate
		}
	}
	return s.rules.Create(ctx, &InterorganizationRule{
		FromOrganizationID: request.FromOrganizationID,
		ToOrganizationID:   request.ToOrganizationID,
		AutoMirror:         request.AutoMirror,
		SupplierContactID:  request.SupplierContactID,
		CustomerContactID:  request.CustomerContactID,
	})
}

func (s InterorganizationService) UpdateRule(ctx context.Context, request UpsertInterorganizationRuleRequest) (*InterorganizationRule, error) {
	if request.ID == nil {
		return nil, ErrRuleNotFound
	}
	rule, err := s.rules.Find(ctx, *request.ID)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, ErrRuleNotFound
	}
	if request.FromOrganizationID != nil && request.ToOrganizationID != nil &&
		*request.FromOrganizationID == *request.ToOrganizationID {
		return nil, ErrRuleSameOrganization
	}
	if request.AutoMirror && (request.SupplierContactID == nil || request.CustomerContactID == nil) {
		return nil, ErrRuleContactsRequired
	}
	if request.FromOrganizationID != nil {
		rule.FromOrganizationID = request.FromOrganizationID
	}
	if request.ToOrganizationID != nil {
		rule.ToOrganizationID = request.ToOrganizationID
	}
	rule.AutoMirror = request.AutoMirror
	if request.SupplierContactID != nil {
		rule.SupplierContactID = request.SupplierContactID
	}
	if request.CustomerContactID != nil {
		rule.CustomerContactID = request.CustomerContactID
	}
	return s.rules.Update(ctx, rule)
}

func (s InterorganizationService) FindRule(ctx context.Context, ruleID uint64) (*InterorganizationRule, error) {
	return s.rules.Find(ctx, ruleID)
}

func (s InterorganizationService) DeleteRule(ctx context.Context, ruleID uint64) error {
	rule, err := s.rules.Find(ctx, ruleID)
	if err != nil {
		return err
	}
	if rule == nil {
		return ErrRuleNotFound
	}
	return s.rules.Delete(ctx, ruleID)
}

func (s InterorganizationService) ListRules(ctx context.Context, organizationID uint64) ([]*InterorganizationRule, error) {
	page, err := s.rules.List(ctx, &query.Query{Filters: []query.Filter{{Field: "from_organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type MirrorSaleOrderRequest struct {
	SaleOrderID      uint64
	ToOrganizationID uint64
}

func (s InterorganizationService) MirrorSaleOrder(ctx context.Context, request MirrorSaleOrderRequest) (*procurement.PurchaseOrder, *InterorganizationTransaction, error) {
	so, err := s.soOrders.Find(ctx, request.SaleOrderID)
	if err != nil {
		return nil, nil, err
	}
	if so == nil || so.OrganizationID == nil {
		return nil, nil, ErrMirrorSourceNotFound
	}
	if so.State == sales.OrderStateDraft || so.State == sales.OrderStateCancelled {
		return nil, nil, ErrMirrorSourceNotPosted
	}

	rule, err := s.ruleFor(ctx, *so.OrganizationID, request.ToOrganizationID)
	if err != nil {
		return nil, nil, err
	}
	if rule == nil {
		return nil, nil, ErrMirrorRuleMissing
	}
	if !rule.AutoMirror {
		return nil, nil, ErrMirrorRuleDisabled
	}
	if rule.SupplierContactID == nil {
		return nil, nil, ErrRuleContactsRequired
	}

	soLines, err := s.soLines.ListByOrder(ctx, so.ID)
	if err != nil {
		return nil, nil, err
	}
	source := make([]*sales.SaleOrderLine, 0, len(soLines))
	for _, line := range soLines {
		if line.ItemID == nil || line.QtyOrdered <= 0 {
			continue
		}
		source = append(source, line)
	}
	if len(source) == 0 {
		return nil, nil, ErrMirrorSourceNoLines
	}

	poLines := make([]*procurement.PurchaseOrderLine, 0, len(source))
	for _, line := range source {
		poLines = append(poLines, &procurement.PurchaseOrderLine{
			ItemID:      line.ItemID,
			Description: line.Description,
			QtyOrdered:  line.QtyOrdered,
			UnitID:      line.UnitID,
			UnitPrice:   line.UnitPrice,
			DiscountPct: line.DiscountPct,
		})
	}

	po, err := s.poCreate.Create(ctx, &procurement.PurchaseOrder{
		OrganizationID: helper.Ptr(request.ToOrganizationID),
		SupplierID:     *rule.SupplierContactID,
		CurrencyCode:   so.CurrencyCode,
		PaymentTermID:  so.PaymentTermID,
		OrderDate:      so.OrderDate,
	}, poLines)
	if err != nil {
		return nil, nil, err
	}

	transaction, err := s.trans.Create(ctx, &InterorganizationTransaction{
		SourceOrganizationID: so.OrganizationID,
		SourceType:           "sale_order",
		SourceID:             helper.Ptr(so.ID),
		MirrorOrganizationID: helper.Ptr(request.ToOrganizationID),
		MirrorType:           "purchase_order",
		MirrorID:             helper.Ptr(po.ID),
		Amount:               po.AmountTotal,
		State:                TransactionStateDone,
	})
	if err != nil {
		return nil, nil, err
	}
	return po, transaction, nil
}

func (s InterorganizationService) ListTransactions(ctx context.Context, organizationID uint64) ([]*InterorganizationTransaction, error) {
	page, err := s.trans.List(ctx, &query.Query{Filters: []query.Filter{{Field: "source_organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s InterorganizationService) ruleFor(ctx context.Context, fromOrganizationID, toOrganizationID uint64) (*InterorganizationRule, error) {
	page, err := s.rules.List(ctx, &query.Query{Filters: []query.Filter{{Field: "from_organization_id", Operator: query.Equal, Value: fromOrganizationID}}})
	if err != nil {
		return nil, err
	}
	for _, rule := range page.Items {
		if rule.ToOrganizationID != nil && *rule.ToOrganizationID == toOrganizationID {
			return rule, nil
		}
	}
	return nil, nil
}
