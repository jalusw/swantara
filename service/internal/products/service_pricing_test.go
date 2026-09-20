package products

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestProductService_CreateTemplate_CreatesTemplate(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: dao.CRUDMock[Item]{
			CreateFunc: func(_ context.Context, template *Item) (*Item, error) {
				template.ID = 1
				return template, nil
			},
		},
	}
	svc := NewProductService(templates, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	created, err := svc.CreateTemplate(ctx, &Item{Name: "T-Shirt"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("template id = %d, want 1", created.ID)
	}
}

func TestProductService_CreateTemplate_RejectsInvalidType(t *testing.T) {
	ctx := context.Background()

	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.CreateTemplate(ctx, &Item{Name: "T-Shirt", Type: "weird"})

	if helper.AssertError(t, err, true, ErrInvalidComputeType) {
		return
	}
}

func TestProductService_CreateTemplate_PropagatesCategoryLookupError(t *testing.T) {
	ctx := context.Background()

	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, categories, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.CreateTemplate(ctx, &Item{Name: "T-Shirt", CategoryID: ptr(uint64(5))})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_CreateVariant_CreatesVariant(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 1}}, nil
		}),
	}
	variants := ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[ItemVariant]{
			CreateFunc: func(_ context.Context, variant *ItemVariant) (*ItemVariant, error) {
				variant.ID = 1
				return variant, nil
			},
		},
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	created, err := svc.CreateVariant(ctx, &ItemVariant{ItemID: 1, Sku: ptr("TS-RED")})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("variant id = %d, want 1", created.ID)
	}
}

func TestProductService_CreateVariant_RejectsMissingTemplate(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, nil
		}),
	}
	svc := NewProductService(templates, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.CreateVariant(ctx, &ItemVariant{ItemID: 1})

	if helper.AssertError(t, err, true, ErrVariantTemplate) {
		return
	}
}

func TestProductService_CreateVariant_PropagatesTemplateLookupError(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewProductService(templates, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.CreateVariant(ctx, &ItemVariant{ItemID: 1})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_CreateVariant_PropagatesSkuLookupError(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 1}}, nil
		}),
	}
	variants := ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[ItemVariant]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*ItemVariant, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.CreateVariant(ctx, &ItemVariant{ItemID: 1, Sku: ptr("TS-RED")})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_GenerateVariants_RejectsMissingTemplate(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, nil
		}),
	}
	svc := NewProductService(templates, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.GenerateVariants(ctx, 1, []AttributeOption{{Name: "Color", Values: []string{"Red"}}})

	if helper.AssertError(t, err, true, ErrItemNotFound) {
		return
	}
}

func TestProductService_GenerateVariants_PropagatesTemplateLookupError(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewProductService(templates, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.GenerateVariants(ctx, 1, []AttributeOption{{Name: "Color", Values: []string{"Red"}}})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_CreateTemplateWithVariants_CreatesTemplateAndVariants(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: dao.CRUDMock[Item]{
			CreateFunc: func(_ context.Context, template *Item) (*Item, error) {
				template.ID = 1
				return template, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}}, nil
			},
		},
	}
	variants := ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[ItemVariant]{
			CreateFunc: func(_ context.Context, variant *ItemVariant) (*ItemVariant, error) {
				return variant, nil
			},
		},
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	created, generated, err := svc.CreateTemplateWithVariants(ctx, &Item{Name: "T-Shirt"}, []AttributeOption{
		{Name: "Color", Values: []string{"Red", "Blue"}},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 || len(generated) != 2 {
		t.Errorf("created = %+v, generated = %d, want template 1 with 2 variants", created, len(generated))
	}
}

func TestProductService_CreateTemplateWithVariants_RejectsInvalidTemplate(t *testing.T) {
	ctx := context.Background()

	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, _, err := svc.CreateTemplateWithVariants(ctx, &Item{Name: "T-Shirt", Type: "weird"}, nil)

	if helper.AssertError(t, err, true, ErrInvalidComputeType) {
		return
	}
}

func TestProductService_CreateTemplateWithVariants_RejectsEmptyMatrix(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: dao.CRUDMock[Item]{
			CreateFunc: func(_ context.Context, template *Item) (*Item, error) {
				template.ID = 1
				return template, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}}, nil
			},
		},
	}
	svc := NewProductService(templates, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, _, err := svc.CreateTemplateWithVariants(ctx, &Item{Name: "T-Shirt"}, nil)

	if helper.AssertError(t, err, true, ErrEmptyAttributeMatrix) {
		return
	}
}

func TestProductService_CreateTemplateWithVariants_PropagatesVariantCreateError(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: dao.CRUDMock[Item]{
			CreateFunc: func(_ context.Context, template *Item) (*Item, error) {
				template.ID = 1
				return template, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}}, nil
			},
		},
	}
	variants := ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[ItemVariant]{
			CreateFunc: func(_ context.Context, _ *ItemVariant) (*ItemVariant, error) {
				return nil, errors.New("insert failed")
			},
		},
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, _, err := svc.CreateTemplateWithVariants(ctx, &Item{Name: "T-Shirt"}, []AttributeOption{
		{Name: "Color", Values: []string{"Red"}},
	})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_CreateRule_CreatesRule(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return &PriceBook{Base: model.Base{ID: 1}, Name: "Retail"}, nil
		}),
	}
	rules := PriceRuleDAOMock{
		CRUDMock: dao.CRUDMock[PriceRule]{
			CreateFunc: func(_ context.Context, rule *PriceRule) (*PriceRule, error) {
				rule.ID = 1
				return rule, nil
			},
		},
	}
	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, rules)

	created, err := svc.CreateRule(ctx, 1, &PriceRule{AppliesTo: AppliesToAll, ComputeType: ComputePercent, DiscountPct: ptr(10.0)})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 || created.PriceBookID != 1 {
		t.Errorf("rule = %+v, want id 1 on price_book 1", created)
	}
}

func TestProductService_CreateRule_RejectsUnknownPriceBook(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return nil, nil
		}),
	}
	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, PriceRuleDAOMock{})

	_, err := svc.CreateRule(ctx, 1, &PriceRule{AppliesTo: AppliesToAll, ComputeType: ComputeFixed})

	if helper.AssertError(t, err, true, ErrPriceBookNotFound) {
		return
	}
}

func TestProductService_CreateRule_PropagatesPriceBookLookupError(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, PriceRuleDAOMock{})

	_, err := svc.CreateRule(ctx, 1, &PriceRule{AppliesTo: AppliesToAll, ComputeType: ComputeFixed})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_CreateRule_RejectsInvalidAppliesTo(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return &PriceBook{Base: model.Base{ID: 1}}, nil
		}),
	}
	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, PriceRuleDAOMock{})

	_, err := svc.CreateRule(ctx, 1, &PriceRule{AppliesTo: "bogus", ComputeType: ComputeFixed})

	if helper.AssertError(t, err, true, ErrInvalidAppliesTo) {
		return
	}
}

func TestProductService_CreateRule_RejectsInvalidComputeType(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return &PriceBook{Base: model.Base{ID: 1}}, nil
		}),
	}
	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, PriceRuleDAOMock{})

	_, err := svc.CreateRule(ctx, 1, &PriceRule{AppliesTo: AppliesToAll, ComputeType: "bogus"})

	if helper.AssertError(t, err, true, ErrInvalidComputeType) {
		return
	}
}

func TestProductService_CreateRule_RejectsCategoryWithoutScope(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return &PriceBook{Base: model.Base{ID: 1}}, nil
		}),
	}
	rules := PriceRuleDAOMock{}
	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, rules)

	_, err := svc.CreateRule(ctx, 1, &PriceRule{AppliesTo: AppliesToCategory, ComputeType: ComputeFixed})

	if helper.AssertError(t, err, true, ErrInvalidRuleScope) {
		return
	}
}

func TestProductService_CreateRule_AcceptsCategoryScope(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return &PriceBook{Base: model.Base{ID: 1}}, nil
		}),
	}
	rules := PriceRuleDAOMock{
		CRUDMock: dao.CRUDMock[PriceRule]{
			CreateFunc: func(_ context.Context, rule *PriceRule) (*PriceRule, error) {
				return rule, nil
			},
		},
	}
	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, rules)

	created, err := svc.CreateRule(ctx, 1, &PriceRule{AppliesTo: AppliesToCategory, CategoryID: ptr(uint64(7)), ComputeType: ComputeFixed, FixedPrice: ptr(15.0)})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.AppliesTo != AppliesToCategory {
		t.Errorf("rule = %+v, want category scope", created)
	}
}

func TestProductService_CreateRule_AcceptsProductScope(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return &PriceBook{Base: model.Base{ID: 1}}, nil
		}),
	}
	rules := PriceRuleDAOMock{
		CRUDMock: dao.CRUDMock[PriceRule]{
			CreateFunc: func(_ context.Context, rule *PriceRule) (*PriceRule, error) {
				return rule, nil
			},
		},
	}
	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, rules)

	created, err := svc.CreateRule(ctx, 1, &PriceRule{AppliesTo: AppliesToProduct, ItemID: ptr(uint64(9)), ComputeType: ComputeFixed, FixedPrice: ptr(15.0)})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.AppliesTo != AppliesToProduct {
		t.Errorf("rule = %+v, want item scope", created)
	}
}

func TestValidateRuleScope_RejectsUnknownScope(t *testing.T) {
	if err := validateRuleScope(&PriceRule{AppliesTo: "bogus"}); err != ErrInvalidAppliesTo {
		t.Errorf("err = %v, want ErrInvalidAppliesTo", err)
	}
}

func TestProductService_ResolvePrice_RejectsUnknownPriceBook(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return nil, nil
		}),
	}
	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, PriceRuleDAOMock{})

	_, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(1), time.Now())

	if helper.AssertError(t, err, true, ErrPriceBookNotFound) {
		return
	}
}

func TestProductService_ResolvePrice_PropagatesPriceBookLookupError(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewProductService(ItemDAOMock{}, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, PriceRuleDAOMock{})

	_, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(1), time.Now())

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_ResolvePrice_RejectsUnknownVariant(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return &PriceBook{Base: model.Base{ID: 1}}, nil
		}),
	}
	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return nil, nil
		}),
	}
	svc := NewProductService(ItemDAOMock{}, variants, dao.CRUDMock[reference.ItemCategory]{}, price_books, PriceRuleDAOMock{})

	_, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(1), time.Now())

	if helper.AssertError(t, err, true, ErrVariantNotFound) {
		return
	}
}

func TestProductService_ResolvePrice_PropagatesVariantLookupError(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return &PriceBook{Base: model.Base{ID: 1}}, nil
		}),
	}
	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewProductService(ItemDAOMock{}, variants, dao.CRUDMock[reference.ItemCategory]{}, price_books, PriceRuleDAOMock{})

	_, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(1), time.Now())

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_ResolvePrice_RejectsUnknownTemplate(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return &PriceBook{Base: model.Base{ID: 1}}, nil
		}),
	}
	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, nil
		}),
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, price_books, PriceRuleDAOMock{})

	_, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(1), time.Now())

	if helper.AssertError(t, err, true, ErrItemNotFound) {
		return
	}
}

func TestProductService_ResolvePrice_PropagatesTemplateLookupError(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return &PriceBook{Base: model.Base{ID: 1}}, nil
		}),
	}
	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, price_books, PriceRuleDAOMock{})

	_, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(1), time.Now())

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_ResolvePrice_PropagatesRuleListError(t *testing.T) {
	ctx := context.Background()

	price_books := PriceBookDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
			return &PriceBook{Base: model.Base{ID: 1}}, nil
		}),
	}
	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 1}, ListPrice: 20}, nil
		}),
	}
	rules := PriceRuleDAOMock{
		ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, price_books, rules)

	_, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(1), time.Now())

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_ResolvePrice_FallsBackToListPriceWhenNoRuleApplies(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	future := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	svc := NewProductService(
		ItemDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}, Name: "T-Shirt", ListPrice: 20}, nil
			}),
		},
		ItemVariantDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
				return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
			}),
		},
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}, Name: "Retail"}, nil
			}),
		},
		PriceRuleDAOMock{
			ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
				return []*PriceRule{
					{Base: model.Base{ID: 1}, AppliesTo: AppliesToAll, ComputeType: ComputeFixed, FixedPrice: ptr(15.0), DateStart: &future},
				}, nil
			},
		},
	)

	resolved, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resolved.Price.String(); got != "20" {
		t.Errorf("price = %s, want 20 (list price)", got)
	}
}

func TestProductService_ResolvePrice_PropagatesProductRuleLookupError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	variants := ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[ItemVariant]{
			FindFunc: func(_ context.Context, id uint64) (*ItemVariant, error) {
				if id == 100 {
					return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
				}
				return nil, errors.New("db down")
			},
		},
	}
	rules := PriceRuleDAOMock{
		ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
			return []*PriceRule{
				{Base: model.Base{ID: 1}, AppliesTo: AppliesToProduct, ItemID: ptr(uint64(200)), ComputeType: ComputeFixed, FixedPrice: ptr(15.0)},
			}, nil
		},
	}
	svc := NewProductService(
		ItemDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}, Name: "T-Shirt", ListPrice: 20}, nil
			}),
		},
		variants,
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}, Name: "Retail"}, nil
			}),
		},
		rules,
	)

	_, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_ResolvePrice_IgnoresUnresolvableProductRule(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	variants := ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[ItemVariant]{
			FindFunc: func(_ context.Context, id uint64) (*ItemVariant, error) {
				if id == 100 {
					return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
				}
				return nil, nil
			},
		},
	}
	rules := PriceRuleDAOMock{
		ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
			return []*PriceRule{
				{Base: model.Base{ID: 1}, AppliesTo: AppliesToProduct, ItemID: ptr(uint64(200)), ComputeType: ComputeFixed, FixedPrice: ptr(15.0)},
			}, nil
		},
	}
	svc := NewProductService(
		ItemDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}, Name: "T-Shirt", ListPrice: 20}, nil
			}),
		},
		variants,
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}, Name: "Retail"}, nil
			}),
		},
		rules,
	)

	resolved, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resolved.Price.String(); got != "20" {
		t.Errorf("price = %s, want 20 (list price)", got)
	}
}

func TestProductService_ResolvePrice_RejectsFixedRuleWithoutPrice(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	svc := NewProductService(
		ItemDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}, ListPrice: 20}, nil
			}),
		},
		ItemVariantDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
				return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
			}),
		},
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}}, nil
			}),
		},
		PriceRuleDAOMock{
			ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
				return []*PriceRule{
					{Base: model.Base{ID: 1}, AppliesTo: AppliesToAll, ComputeType: ComputeFixed},
				}, nil
			},
		},
	)

	_, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)

	if helper.AssertError(t, err, true, ErrPriceUnavailable) {
		return
	}
}

func TestProductService_ResolvePrice_RejectsPercentRuleWithoutDiscount(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	svc := NewProductService(
		ItemDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}, ListPrice: 20}, nil
			}),
		},
		ItemVariantDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
				return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
			}),
		},
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}}, nil
			}),
		},
		PriceRuleDAOMock{
			ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
				return []*PriceRule{
					{Base: model.Base{ID: 1}, AppliesTo: AppliesToAll, ComputeType: ComputePercent},
				}, nil
			},
		},
	)

	_, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)

	if helper.AssertError(t, err, true, ErrPriceUnavailable) {
		return
	}
}

func TestProductService_ResolvePrice_RejectsFormulaCompute(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	svc := NewProductService(
		ItemDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}, ListPrice: 20}, nil
			}),
		},
		ItemVariantDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
				return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
			}),
		},
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}}, nil
			}),
		},
		PriceRuleDAOMock{
			ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
				return []*PriceRule{
					{Base: model.Base{ID: 1}, AppliesTo: AppliesToAll, ComputeType: ComputeFormula},
				}, nil
			},
		},
	)

	_, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)

	if helper.AssertError(t, err, true, ErrPriceUnavailable) {
		return
	}
}

func TestProductService_ResolvePrice_ExcludesRuleAfterWindow(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)

	svc := NewProductService(
		ItemDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}, ListPrice: 20}, nil
			}),
		},
		ItemVariantDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
				return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
			}),
		},
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}}, nil
			}),
		},
		PriceRuleDAOMock{
			ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
				return []*PriceRule{
					{Base: model.Base{ID: 1}, AppliesTo: AppliesToAll, ComputeType: ComputeFixed, FixedPrice: ptr(15.0), DateEnd: &end},
				}, nil
			},
		},
	)

	resolved, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resolved.Price.String(); got != "20" {
		t.Errorf("price = %s, want 20 (rule window ended)", got)
	}
}

func TestProductService_ResolvePrice_AppliesCategoryRule(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	svc := NewProductService(
		ItemDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}, ListPrice: 20, CategoryID: ptr(uint64(7))}, nil
			}),
		},
		ItemVariantDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
				return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
			}),
		},
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}}, nil
			}),
		},
		PriceRuleDAOMock{
			ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
				return []*PriceRule{
					{Base: model.Base{ID: 1}, AppliesTo: AppliesToCategory, CategoryID: ptr(uint64(7)), ComputeType: ComputeFixed, FixedPrice: ptr(15.0)},
				}, nil
			},
		},
	)

	resolved, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resolved.Price.String(); got != "15" {
		t.Errorf("price = %s, want 15", got)
	}
}

func TestProductService_ResolvePrice_IgnoresUnknownScope(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	svc := NewProductService(
		ItemDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}, ListPrice: 20}, nil
			}),
		},
		ItemVariantDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
				return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
			}),
		},
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}}, nil
			}),
		},
		PriceRuleDAOMock{
			ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
				return []*PriceRule{
					{Base: model.Base{ID: 1}, AppliesTo: "bogus", ComputeType: ComputeFixed, FixedPrice: ptr(15.0)},
				}, nil
			},
		},
	)

	resolved, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resolved.Price.String(); got != "20" {
		t.Errorf("price = %s, want 20 (unknown scope ignored)", got)
	}
}

func TestProductService_ResolvePrice_IgnoresProductRuleWithoutProduct(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	svc := NewProductService(
		ItemDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}, ListPrice: 20}, nil
			}),
		},
		ItemVariantDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
				return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
			}),
		},
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}}, nil
			}),
		},
		PriceRuleDAOMock{
			ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
				return []*PriceRule{
					{Base: model.Base{ID: 1}, AppliesTo: AppliesToProduct, ComputeType: ComputeFixed, FixedPrice: ptr(15.0)},
				}, nil
			},
		},
	)

	resolved, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resolved.Price.String(); got != "20" {
		t.Errorf("price = %s, want 20 (item rule without item ignored)", got)
	}
}

func TestProductService_ResolvePrice_ReturnsListPriceWithNoRules(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	svc := NewProductService(
		ItemDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
				return &Item{Base: model.Base{ID: 1}, ListPrice: 20}, nil
			}),
		},
		ItemVariantDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
				return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
			}),
		},
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}}, nil
			}),
		},
		PriceRuleDAOMock{},
	)

	resolved, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resolved.Price.String(); got != "20" {
		t.Errorf("price = %s, want 20 (list price)", got)
	}
}

func TestProductService_ResolveIncomeAccount_ReturnsAccount(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 5}, CategoryID: ptr(uint64(9)), OrganizationID: ptr(uint64(10))}, nil
		}),
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return &reference.ItemCategory{Base: model.Base{ID: 9}, OrganizationID: ptr(uint64(10)), IncomeAccountID: ptr(uint64(200))}, nil
		},
	}
	svc := NewProductService(templates, variants, categories, PriceBookDAOMock{}, PriceRuleDAOMock{})

	account, err := svc.ResolveIncomeAccount(ctx, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account != 200 {
		t.Errorf("account = %d, want 200", account)
	}
}

func TestProductService_ResolveIncomeAccount_RejectsVariantWithoutCategory(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 5}}, nil
		}),
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveIncomeAccount(ctx, 1)

	if helper.AssertError(t, err, true, ErrNoIncomeAccount) {
		return
	}
}

func TestProductService_ResolveIncomeAccount_RejectsCategoryWithoutAccount(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 5}, CategoryID: ptr(uint64(9))}, nil
		}),
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return &reference.ItemCategory{Base: model.Base{ID: 9}}, nil
		},
	}
	svc := NewProductService(templates, variants, categories, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveIncomeAccount(ctx, 1)

	if helper.AssertError(t, err, true, ErrNoIncomeAccount) {
		return
	}
}

func TestProductService_ResolveIncomeAccount_RejectsMissingVariant(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return nil, nil
		}),
	}
	svc := NewProductService(ItemDAOMock{}, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveIncomeAccount(ctx, 1)

	if helper.AssertError(t, err, true, ErrVariantNotFound) {
		return
	}
}

func TestProductService_ResolveIncomeAccount_PropagatesVariantError(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewProductService(ItemDAOMock{}, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveIncomeAccount(ctx, 1)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_ResolveIncomeAccount_RejectsMissingTemplate(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, nil
		}),
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveIncomeAccount(ctx, 1)

	if helper.AssertError(t, err, true, ErrItemNotFound) {
		return
	}
}

func TestProductService_ResolveIncomeAccount_PropagatesTemplateError(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveIncomeAccount(ctx, 1)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_ResolveIncomeAccount_PropagatesCategoryError(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 5}, CategoryID: ptr(uint64(9))}, nil
		}),
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewProductService(templates, variants, categories, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveIncomeAccount(ctx, 1)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_ResolveExpenseAccount_ReturnsAccount(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 5}, CategoryID: ptr(uint64(9)), OrganizationID: ptr(uint64(10))}, nil
		}),
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return &reference.ItemCategory{Base: model.Base{ID: 9}, OrganizationID: ptr(uint64(10)), ExpenseAccountID: ptr(uint64(300))}, nil
		},
	}
	svc := NewProductService(templates, variants, categories, PriceBookDAOMock{}, PriceRuleDAOMock{})

	account, err := svc.ResolveExpenseAccount(ctx, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account != 300 {
		t.Errorf("account = %d, want 300", account)
	}
}

func TestProductService_ResolveExpenseAccount_RejectsVariantWithoutCategory(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 5}}, nil
		}),
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveExpenseAccount(ctx, 1)

	if helper.AssertError(t, err, true, ErrNoExpenseAccount) {
		return
	}
}

func TestProductService_ResolveExpenseAccount_RejectsCategoryWithoutAccount(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 5}, CategoryID: ptr(uint64(9))}, nil
		}),
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return &reference.ItemCategory{Base: model.Base{ID: 9}}, nil
		},
	}
	svc := NewProductService(templates, variants, categories, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveExpenseAccount(ctx, 1)

	if helper.AssertError(t, err, true, ErrNoExpenseAccount) {
		return
	}
}

func TestProductService_ResolveExpenseAccount_RejectsMissingVariant(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return nil, nil
		}),
	}
	svc := NewProductService(ItemDAOMock{}, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveExpenseAccount(ctx, 1)

	if helper.AssertError(t, err, true, ErrVariantNotFound) {
		return
	}
}

func TestProductService_ResolveExpenseAccount_PropagatesVariantError(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewProductService(ItemDAOMock{}, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveExpenseAccount(ctx, 1)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_ResolveExpenseAccount_RejectsMissingTemplate(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, nil
		}),
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveExpenseAccount(ctx, 1)

	if helper.AssertError(t, err, true, ErrItemNotFound) {
		return
	}
}

func TestProductService_ResolveExpenseAccount_PropagatesTemplateError(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveExpenseAccount(ctx, 1)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_ResolveExpenseAccount_PropagatesCategoryError(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 5}, CategoryID: ptr(uint64(9))}, nil
		}),
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewProductService(templates, variants, categories, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveExpenseAccount(ctx, 1)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_ResolveVariantOrganization_ReturnsOrganization(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 5}, OrganizationID: ptr(uint64(10))}, nil
		}),
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	org, err := svc.ResolveVariantOrganization(ctx, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org == nil || *org != 10 {
		t.Errorf("org = %v, want 10", org)
	}
}

func TestProductService_ResolveVariantOrganization_RejectsMissingVariant(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return nil, nil
		}),
	}
	svc := NewProductService(ItemDAOMock{}, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveVariantOrganization(ctx, 1)

	if helper.AssertError(t, err, true, ErrVariantNotFound) {
		return
	}
}

func TestProductService_ResolveVariantOrganization_PropagatesVariantError(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewProductService(ItemDAOMock{}, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveVariantOrganization(ctx, 1)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestProductService_ResolveVariantOrganization_RejectsMissingTemplate(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, nil
		}),
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveVariantOrganization(ctx, 1)

	if helper.AssertError(t, err, true, ErrItemNotFound) {
		return
	}
}

func TestProductService_ResolveVariantOrganization_PropagatesTemplateError(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*ItemVariant, error) {
			return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
		}),
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return nil, errors.New("db down")
		}),
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveVariantOrganization(ctx, 1)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestExpandMatrix_SkipsAttributesWithoutValues(t *testing.T) {
	result := expandMatrix([]AttributeOption{
		{Name: "Color", Values: []string{"Red"}},
		{Name: "Size", Values: nil},
	})

	if len(result) != 1 || result[0]["Color"] != "Red" {
		t.Errorf("result = %+v, want one combination with Red", result)
	}
}
