package products

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type ItemDAOMock struct {
	dao.CRUDMock[Item]
	CreateWithVariantsFunc func(ctx context.Context, template *Item, variants []*ItemVariant) (*Item, error)
}

func (m ItemDAOMock) CreateWithVariants(ctx context.Context, template *Item, variants []*ItemVariant) (*Item, error) {
	if m.CreateWithVariantsFunc != nil {
		return m.CreateWithVariantsFunc(ctx, template, variants)
	}
	return template, nil
}

type ItemVariantDAOMock struct {
	dao.CRUDMock[ItemVariant]
	ListByTemplateFunc func(ctx context.Context, templateID uint64) ([]*ItemVariant, error)
	VariantExistsFunc  func(ctx context.Context, id uint64) (bool, error)
}

func (m ItemVariantDAOMock) ListByTemplate(ctx context.Context, templateID uint64) ([]*ItemVariant, error) {
	if m.ListByTemplateFunc != nil {
		return m.ListByTemplateFunc(ctx, templateID)
	}
	return []*ItemVariant{}, nil
}

func (m ItemVariantDAOMock) VariantExists(ctx context.Context, id uint64) (bool, error) {
	if m.VariantExistsFunc != nil {
		return m.VariantExistsFunc(ctx, id)
	}
	return true, nil
}

type PriceBookDAOMock struct {
	dao.CRUDMock[PriceBook]
}

type PriceRuleDAOMock struct {
	dao.CRUDMock[PriceRule]
	ListByPriceBookFunc func(ctx context.Context, price_bookID uint64) ([]*PriceRule, error)
}

func (m PriceRuleDAOMock) ListByPriceBook(ctx context.Context, price_bookID uint64) ([]*PriceRule, error) {
	if m.ListByPriceBookFunc != nil {
		return m.ListByPriceBookFunc(ctx, price_bookID)
	}
	return []*PriceRule{}, nil
}

type SupplierProductDAOMock struct {
	dao.CRUDMock[SupplierProduct]
	ListByItemFunc     func(ctx context.Context, itemID uint64) ([]*SupplierProduct, error)
	ListBySupplierFunc func(ctx context.Context, supplierID uint64) ([]*SupplierProduct, error)
}

func (m SupplierProductDAOMock) ListByItem(ctx context.Context, itemID uint64) ([]*SupplierProduct, error) {
	if m.ListByItemFunc != nil {
		return m.ListByItemFunc(ctx, itemID)
	}
	return []*SupplierProduct{}, nil
}

func (m SupplierProductDAOMock) ListBySupplier(ctx context.Context, supplierID uint64) ([]*SupplierProduct, error) {
	if m.ListBySupplierFunc != nil {
		return m.ListBySupplierFunc(ctx, supplierID)
	}
	return []*SupplierProduct{}, nil
}

func (m SupplierProductDAOMock) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[SupplierProduct], error) {
	return m.List(ctx, q)
}

func (m SupplierProductDAOMock) FindInOrg(ctx context.Context, id, organizationID uint64) (*SupplierProduct, error) {
	return m.Find(ctx, id)
}
