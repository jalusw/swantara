package products

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestProductService_CreateTemplate_RejectsUnknownCategory(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{}
	svc := NewProductService(templates, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return nil, nil
		},
	}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.CreateTemplate(ctx, &Item{Name: "T-Shirt", CategoryID: ptr(uint64(5))})
	if helper.AssertError(t, err, true, ErrItemCategory) {
		return
	}
}

func TestProductService_CreateTemplate_RejectsCategoryFromAnotherOrganization(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{}
	svc := NewProductService(templates, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return &reference.ItemCategory{Base: model.Base{ID: 5}, OrganizationID: ptr(uint64(20))}, nil
		},
	}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.CreateTemplate(ctx, &Item{Name: "T-Shirt", OrganizationID: ptr(uint64(10)), CategoryID: ptr(uint64(5))})
	if helper.AssertError(t, err, true, ErrItemCategoryOrganization) {
		return
	}
}

func TestProductService_ResolveIncomeAccount_RejectsCategoryFromAnotherOrganization(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*ItemVariant, error) {
				return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
			},
		},
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 5}, CategoryID: ptr(uint64(9)), OrganizationID: ptr(uint64(10))}, nil
		}),
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return &reference.ItemCategory{Base: model.Base{ID: 9}, OrganizationID: ptr(uint64(20)), IncomeAccountID: ptr(uint64(200))}, nil
		},
	}
	svc := NewProductService(templates, variants, categories, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveIncomeAccount(ctx, 1)
	if helper.AssertError(t, err, true, ErrItemCategoryOrganization) {
		return
	}
}

func TestProductService_ResolveExpenseAccount_RejectsCategoryFromAnotherOrganization(t *testing.T) {
	ctx := context.Background()

	variants := ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*ItemVariant, error) {
				return &ItemVariant{Base: model.Base{ID: 1}, ItemID: 5}, nil
			},
		},
	}
	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 5}, CategoryID: ptr(uint64(9)), OrganizationID: ptr(uint64(10))}, nil
		}),
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return &reference.ItemCategory{Base: model.Base{ID: 9}, OrganizationID: ptr(uint64(20)), ExpenseAccountID: ptr(uint64(300))}, nil
		},
	}
	svc := NewProductService(templates, variants, categories, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.ResolveExpenseAccount(ctx, 1)
	if helper.AssertError(t, err, true, ErrItemCategoryOrganization) {
		return
	}
}

func TestProductService_UpdateTemplate_CallsUpdate(t *testing.T) {
	ctx := context.Background()

	created := false
	templates := ItemDAOMock{
		CRUDMock: dao.CRUDMock[Item]{
			CreateFunc: func(_ context.Context, _ *Item) (*Item, error) {
				created = true
				return nil, nil
			},
			UpdateFunc: func(_ context.Context, template *Item) (*Item, error) {
				return template, nil
			},
		},
	}
	svc := NewProductService(templates, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	updated, err := svc.UpdateTemplate(ctx, &Item{Base: model.Base{ID: 7}, Name: "T-Shirt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created {
		t.Error("UpdateTemplate must not call Create")
	}
	if updated == nil || updated.ID != 7 {
		t.Errorf("updated = %v, want id 7", updated)
	}
}

func TestProductService_UpdateTemplate_RejectsUnknownCategory(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: dao.CRUDMock[Item]{
			UpdateFunc: func(_ context.Context, _ *Item) (*Item, error) {
				return nil, nil
			},
		},
	}
	svc := NewProductService(templates, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
			return nil, nil
		},
	}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.UpdateTemplate(ctx, &Item{Name: "T-Shirt", CategoryID: ptr(uint64(5))})
	if helper.AssertError(t, err, true, ErrItemCategory) {
		return
	}
}

func TestProductService_CreateVariant_RejectsDuplicateSku(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 1}}, nil
		}),
	}
	variants := ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[ItemVariant]{
			SearchFunc: func(_ context.Context, field string, value any) (*ItemVariant, error) {
				if field == "sku" && value == "TS-RED" {
					return &ItemVariant{Base: model.Base{ID: 9}, Sku: ptr("TS-RED")}, nil
				}
				return nil, nil
			},
		},
	}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.CreateVariant(ctx, &ItemVariant{ItemID: 1, Sku: ptr("TS-RED")})
	if helper.AssertError(t, err, true, ErrSkuTaken) {
		return
	}
}

func TestProductService_GenerateVariants_ExplodesMatrix(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 1}, Name: "T-Shirt"}, nil
		}),
	}
	variants := ItemVariantDAOMock{}
	svc := NewProductService(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	generated, err := svc.GenerateVariants(ctx, 1, []AttributeOption{
		{Name: "Color", Values: []string{"Red", "Blue"}},
		{Name: "Size", Values: []string{"S", "M", "L"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(generated) != 6 {
		t.Errorf("variants = %d, want 6", len(generated))
	}
}

func TestProductService_GenerateVariants_RejectsEmptyMatrix(t *testing.T) {
	ctx := context.Background()

	templates := ItemDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
			return &Item{Base: model.Base{ID: 1}}, nil
		}),
	}
	svc := NewProductService(templates, ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, PriceBookDAOMock{}, PriceRuleDAOMock{})

	_, err := svc.GenerateVariants(ctx, 1, nil)
	if helper.AssertError(t, err, true, ErrEmptyAttributeMatrix) {
		return
	}
}

func TestProductService_ResolvePrice_AppliesVariantPrecedence(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		rules []*PriceRule
		want  string
	}{
		{
			name: "variant rule beats item and default",
			rules: []*PriceRule{
				{Base: model.Base{ID: 1}, AppliesTo: AppliesToAll, ComputeType: ComputePercent, DiscountPct: ptr(10.0)},
				{Base: model.Base{ID: 2}, AppliesTo: AppliesToVariant, ItemID: ptr(uint64(100)), ComputeType: ComputeFixed, FixedPrice: ptr(25.0)},
			},
			want: "25",
		},
		{
			name: "item rule beats category and default",
			rules: []*PriceRule{
				{Base: model.Base{ID: 1}, AppliesTo: AppliesToAll, ComputeType: ComputePercent, DiscountPct: ptr(10.0)},
				{Base: model.Base{ID: 3}, AppliesTo: AppliesToProduct, ItemID: ptr(uint64(101)), ComputeType: ComputeFixed, FixedPrice: ptr(30.0)},
			},
			want: "30",
		},
		{
			name: "category rule beats default",
			rules: []*PriceRule{
				{Base: model.Base{ID: 1}, AppliesTo: AppliesToAll, ComputeType: ComputePercent, DiscountPct: ptr(10.0)},
				{Base: model.Base{ID: 4}, AppliesTo: AppliesToCategory, CategoryID: ptr(uint64(7)), ComputeType: ComputeFixed, FixedPrice: ptr(18.0)},
			},
			want: "18",
		},
		{
			name: "default rule applies when no scope matches",
			rules: []*PriceRule{
				{Base: model.Base{ID: 1}, AppliesTo: AppliesToAll, ComputeType: ComputePercent, DiscountPct: ptr(10.0)},
			},
			want: "18",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewProductService(
				ItemDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Item, error) {
						return &Item{Base: model.Base{ID: 1}, Name: "T-Shirt", ListPrice: 20}, nil
					}),
				},
				ItemVariantDAOMock{
					CRUDMock: dao.CRUDMock[ItemVariant]{
						FindFunc: func(_ context.Context, id uint64) (*ItemVariant, error) {
							switch id {
							case 100:
								return &ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
							case 101:
								return &ItemVariant{Base: model.Base{ID: 101}, ItemID: 1}, nil
							default:
								return &ItemVariant{Base: model.Base{ID: id}, ItemID: 1}, nil
							}
						},
						SearchFunc: func(_ context.Context, field string, value any) (*ItemVariant, error) {
							return nil, nil
						},
					},
				},
				dao.CRUDMock[reference.ItemCategory]{},
				PriceBookDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
						return &PriceBook{Base: model.Base{ID: 1}, Name: "Retail"}, nil
					}),
				},
				PriceRuleDAOMock{
					ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
						return tt.rules, nil
					},
				},
			)

			resolved, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := resolved.Price.String(); got != tt.want {
				t.Errorf("price = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestProductService_ResolvePrice_HonorsDateWindow(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

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
				return &PriceBook{Base: model.Base{ID: 1}, Name: "Promo"}, nil
			}),
		},
		PriceRuleDAOMock{
			ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
				return []*PriceRule{
					{Base: model.Base{ID: 1}, AppliesTo: AppliesToAll, ComputeType: ComputeFixed, FixedPrice: ptr(15.0), DateStart: &start, DateEnd: &end},
					{Base: model.Base{ID: 2}, AppliesTo: AppliesToAll, ComputeType: ComputePercent, DiscountPct: ptr(10.0)},
				}, nil
			},
		},
	)

	resolved, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resolved.Price.String(); got != "18" {
		t.Errorf("price = %s, want 18 (date outside promo window)", got)
	}
}

func TestProductService_ResolvePrice_AppliesThroughEndDay(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 7, 31, 9, 0, 0, 0, time.UTC)

	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

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
				return &PriceBook{Base: model.Base{ID: 1}, Name: "Promo"}, nil
			}),
		},
		PriceRuleDAOMock{
			ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
				return []*PriceRule{
					{Base: model.Base{ID: 1}, AppliesTo: AppliesToAll, ComputeType: ComputeFixed, FixedPrice: ptr(15.0), DateStart: &start, DateEnd: &end},
				}, nil
			},
		},
	)

	resolved, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resolved.Price.String(); got != "15" {
		t.Errorf("price = %s, want 15 (rule still applies on its end day)", got)
	}
}

func TestProductService_ResolvePrice_HonorsMinQty(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

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
				return &PriceBook{Base: model.Base{ID: 1}, Name: "Wholesale"}, nil
			}),
		},
		PriceRuleDAOMock{
			ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*PriceRule, error) {
				return []*PriceRule{
					{Base: model.Base{ID: 1}, AppliesTo: AppliesToAll, MinQty: 10, ComputeType: ComputeFixed, FixedPrice: ptr(12.0)},
					{Base: model.Base{ID: 2}, AppliesTo: AppliesToAll, ComputeType: ComputePercent, DiscountPct: ptr(10.0)},
				}, nil
			},
		},
	)

	resolved, err := svc.ResolvePrice(ctx, 1, 100, amount.FromFloat64(5), date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resolved.Price.String(); got != "18" {
		t.Errorf("price = %s, want 18 (min qty 10 not met)", got)
	}
}

func TestProductService_CreateRule_RejectsInvalidScope(t *testing.T) {
	ctx := context.Background()

	svc := NewProductService(
		ItemDAOMock{},
		ItemVariantDAOMock{},
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}, Name: "Retail"}, nil
			}),
		},
		PriceRuleDAOMock{},
	)

	_, err := svc.CreateRule(ctx, 1, &PriceRule{AppliesTo: AppliesToVariant, ComputeType: ComputeFixed})
	if helper.AssertError(t, err, true, ErrInvalidRuleScope) {
		return
	}
}

func TestProductService_CreateRule_RejectsOutOfRangeDiscount(t *testing.T) {
	ctx := context.Background()

	svc := NewProductService(
		ItemDAOMock{},
		ItemVariantDAOMock{},
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}, Name: "Retail"}, nil
			}),
		},
		PriceRuleDAOMock{},
	)

	_, err := svc.CreateRule(ctx, 1, &PriceRule{AppliesTo: AppliesToAll, ComputeType: ComputePercent, DiscountPct: ptr(120.0)})
	if helper.AssertError(t, err, true, ErrInvalidRuleValue) {
		return
	}
}

func TestProductService_CreateRule_RejectsNegativeFixedPrice(t *testing.T) {
	ctx := context.Background()

	svc := NewProductService(
		ItemDAOMock{},
		ItemVariantDAOMock{},
		dao.CRUDMock[reference.ItemCategory]{},
		PriceBookDAOMock{
			CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*PriceBook, error) {
				return &PriceBook{Base: model.Base{ID: 1}, Name: "Retail"}, nil
			}),
		},
		PriceRuleDAOMock{},
	)

	_, err := svc.CreateRule(ctx, 1, &PriceRule{AppliesTo: AppliesToAll, ComputeType: ComputeFixed, FixedPrice: ptr(-5.0)})
	if helper.AssertError(t, err, true, ErrInvalidRuleValue) {
		return
	}
}

func daoCRUDFind[E any](fn func(ctx context.Context, id uint64) (*E, error)) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{FindFunc: fn}
}

func ptr[T any](v T) *T {
	return &v
}
