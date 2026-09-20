package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/products"
)

type productAdapter struct {
	variants  products.ItemVariantDAO
	templates products.ItemDAO
}

func NewProductAdapter(variants products.ItemVariantDAO, templates products.ItemDAO) ItemResolver {
	return productAdapter{variants: variants, templates: templates}
}

func (a productAdapter) FindTemplate(ctx context.Context, itemID uint64) (*struct {
	Type       string
	CategoryID *uint64
}, error) {
	variant, err := a.variants.Find(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if variant != nil {
		tpl, err := a.templates.Find(ctx, variant.ItemID)
		if err != nil {
			return nil, err
		}
		if tpl != nil {
			return &struct {
				Type       string
				CategoryID *uint64
			}{Type: tpl.Type, CategoryID: tpl.CategoryID}, nil
		}
	}
	tpl, err := a.templates.Find(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if tpl != nil {
		return &struct {
			Type       string
			CategoryID *uint64
		}{Type: tpl.Type, CategoryID: tpl.CategoryID}, nil
	}
	return nil, nil
}
