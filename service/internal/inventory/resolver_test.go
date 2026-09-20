package inventory

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestItemResolver_Resolve_ResolvesCategoryAccounts(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return &products.Item{
					Base: model.Base{ID: 5}, CategoryID: helper.Ptr(uint64(9)), OrganizationID: helper.Ptr(uint64(10)),
					Tracking: "none", StandardCost: 25, Weight: 2.5, Volume: 3.5,
				}, nil
			},
		},
	}
	costMethod := "average"
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return &reference.ItemCategory{
				Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(10)), CostMethod: &costMethod,
				StockValuationAccountID: helper.Ptr(uint64(1300)),
				StockInputAccountID:     helper.Ptr(uint64(1310)),
				StockOutputAccountID:    helper.Ptr(uint64(1330)),
				CogsAccountID:           helper.Ptr(uint64(1320)),
			}, nil
		},
	}
	resolver := NewItemResolver(variants, templates, categories)

	resolved, err := resolver.Resolve(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.StockValuationAccountID != 1300 || resolved.StockInputAccountID != 1310 ||
		resolved.StockOutputAccountID != 1330 || resolved.CogsAccountID != 1320 {
		t.Errorf("accounts = %+v, want all category accounts resolved", resolved.StockAccounts)
	}
	if resolved.CostMethod != "average" || resolved.Tracking != "none" || resolved.Weight != 2.5 || resolved.Volume != 3.5 {
		t.Errorf("resolved = %+v, want category cost method and template fields", resolved)
	}
}

func TestItemResolver_Resolve_NilCategoryAccountsResolveToZero(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return &products.Item{Base: model.Base{ID: 5}, CategoryID: helper.Ptr(uint64(9))}, nil
			},
		},
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return &reference.ItemCategory{Base: model.Base{ID: 9}, StockValuationAccountID: helper.Ptr(uint64(1300))}, nil
		},
	}
	resolver := NewItemResolver(variants, templates, categories)

	resolved, err := resolver.Resolve(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.StockValuationAccountID != 1300 || resolved.StockInputAccountID != 0 || resolved.CogsAccountID != 0 {
		t.Errorf("accounts = %+v, want only valuation account set", resolved.StockAccounts)
	}
	if resolved.CostMethod != "fifo" {
		t.Errorf("cost method = %s, want default fifo", resolved.CostMethod)
	}
}

func TestItemResolver_Resolve_RejectsMissingVariant(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return nil, nil
			},
		},
	}
	resolver := NewItemResolver(variants, products.ItemDAOMock{}, dao.CRUDMock[reference.ItemCategory]{})

	_, err := resolver.Resolve(ctx, 1)
	if helper.AssertError(t, err, true, ErrVariantNotFound) {
		return
	}
}

func TestItemResolver_Resolve_RejectsMissingTemplate(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, nil
			},
		},
	}
	resolver := NewItemResolver(variants, templates, dao.CRUDMock[reference.ItemCategory]{})

	_, err := resolver.Resolve(ctx, 1)
	if helper.AssertError(t, err, true, ErrItemNotFound) {
		return
	}
}

func TestItemResolver_Resolve_SkipsMissingCategory(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return &products.Item{Base: model.Base{ID: 5}, CategoryID: helper.Ptr(uint64(9)), Tracking: "batch"}, nil
			},
		},
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return nil, nil
		},
	}
	resolver := NewItemResolver(variants, templates, categories)

	resolved, err := resolver.Resolve(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.Tracking != "batch" || resolved.StockValuationAccountID != 0 {
		t.Errorf("resolved = %+v, want template tracking and no accounts", resolved)
	}
}

func TestItemResolver_Resolve_PropagatesVariantFindError(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return nil, ErrVariantNotFound
			},
		},
	}
	resolver := NewItemResolver(variants, products.ItemDAOMock{}, dao.CRUDMock[reference.ItemCategory]{})

	_, err := resolver.Resolve(ctx, 1)
	if helper.AssertError(t, err, true, ErrVariantNotFound) {
		return
	}
}

func TestItemResolver_Resolve_PropagatesCategoryFindError(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return &products.Item{Base: model.Base{ID: 5}, CategoryID: helper.Ptr(uint64(9))}, nil
			},
		},
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return nil, ErrVariantNotFound
		},
	}
	resolver := NewItemResolver(variants, templates, categories)

	_, err := resolver.Resolve(ctx, 1)
	if helper.AssertError(t, err, true, ErrVariantNotFound) {
		return
	}
}

func TestItemResolver_Resolve_CategoryWithEmptyCostMethodKeepsDefault(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return &products.Item{Base: model.Base{ID: 5}, CategoryID: helper.Ptr(uint64(9))}, nil
			},
		},
	}
	costMethod := ""
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return &reference.ItemCategory{Base: model.Base{ID: 9}, CostMethod: &costMethod}, nil
		},
	}
	resolver := NewItemResolver(variants, templates, categories)

	resolved, err := resolver.Resolve(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.CostMethod != "fifo" {
		t.Errorf("cost method = %s, want default fifo for empty category cost method", resolved.CostMethod)
	}
}

func TestItemResolver_Resolve_RejectsCategoryFromAnotherOrganization(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return &products.Item{Base: model.Base{ID: 5}, CategoryID: helper.Ptr(uint64(9)), OrganizationID: helper.Ptr(uint64(10))}, nil
			},
		},
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return &reference.ItemCategory{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(20))}, nil
		},
	}
	resolver := NewItemResolver(variants, templates, categories)

	_, err := resolver.Resolve(ctx, 1)

	helper.AssertError(t, err, true, products.ErrItemCategoryOrganization)
}
