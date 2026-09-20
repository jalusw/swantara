package inventory

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type StockAccounts struct {
	StockValuationAccountID uint64
	StockInputAccountID     uint64
	StockOutputAccountID    uint64
	CogsAccountID           uint64
}

type ResolvedItem struct {
	StockAccounts
	Tracking       string
	CostMethod     string
	StandardCost   float64
	Weight         float64
	Volume         float64
	OrganizationID *uint64
}

type ItemResolver interface {
	Resolve(ctx context.Context, variantID uint64) (ResolvedItem, error)
}

type productResolver struct {
	variants   products.ItemVariantDAO
	templates  products.ItemDAO
	categories dao.CRUD[reference.ItemCategory]
}

func NewItemResolver(
	variants products.ItemVariantDAO,
	templates products.ItemDAO,
	categories dao.CRUD[reference.ItemCategory],
) ItemResolver {
	return productResolver{variants: variants, templates: templates, categories: categories}
}

func (r productResolver) Resolve(ctx context.Context, variantID uint64) (ResolvedItem, error) {
	variant, err := r.variants.Find(ctx, variantID)
	if err != nil {
		return ResolvedItem{}, err
	}
	if variant == nil {
		return ResolvedItem{}, ErrVariantNotFound
	}

	template, err := r.templates.Find(ctx, variant.ItemID)
	if err != nil {
		return ResolvedItem{}, err
	}
	if template == nil {
		return ResolvedItem{}, ErrItemNotFound
	}

	accounts := StockAccounts{}
	costMethod := "fifo"
	if template.CategoryID != nil {
		category, err := r.categories.Find(ctx, *template.CategoryID)
		if err != nil {
			return ResolvedItem{}, err
		}
		if category != nil {
			if template.OrganizationID != nil && category.OrganizationID != nil && *category.OrganizationID != *template.OrganizationID {
				return ResolvedItem{}, products.ErrItemCategoryOrganization
			}
			accounts = StockAccounts{
				StockValuationAccountID: derefUint(category.StockValuationAccountID),
				StockInputAccountID:     derefUint(category.StockInputAccountID),
				StockOutputAccountID:    derefUint(category.StockOutputAccountID),
				CogsAccountID:           derefUint(category.CogsAccountID),
			}
			if category.CostMethod != nil && *category.CostMethod != "" {
				costMethod = *category.CostMethod
			}
		}
	}
	return ResolvedItem{StockAccounts: accounts, Tracking: template.Tracking, CostMethod: costMethod, StandardCost: template.StandardCost, Weight: template.Weight, Volume: template.Volume, OrganizationID: template.OrganizationID}, nil
}

func derefUint(value *uint64) uint64 {
	if value == nil {
		return 0
	}
	return *value
}
