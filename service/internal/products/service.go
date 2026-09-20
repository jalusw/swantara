package products

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type ProductService struct {
	templates   ItemDAO
	variants    ItemVariantDAO
	categories  dao.CRUD[reference.ItemCategory]
	price_books PriceBookDAO
	rules       PriceRuleDAO
}

func NewProductService(
	templates ItemDAO,
	variants ItemVariantDAO,
	categories dao.CRUD[reference.ItemCategory],
	price_books PriceBookDAO,
	rules PriceRuleDAO,
) ProductService {
	return ProductService{
		templates:   templates,
		variants:    variants,
		categories:  categories,
		price_books: price_books,
		rules:       rules,
	}
}

func (s ProductService) ListTemplates(ctx context.Context, q *query.Query) (*query.Page[Item], error) {
	return s.templates.List(ctx, q)
}

func (s ProductService) FindTemplate(ctx context.Context, id uint64) (*Item, error) {
	return s.templates.Find(ctx, id)
}

func (s ProductService) DeleteTemplate(ctx context.Context, id uint64) error {
	return s.templates.Delete(ctx, id)
}

func (s ProductService) ListVariantsByTemplate(ctx context.Context, templateID uint64) ([]*ItemVariant, error) {
	return s.variants.ListByTemplate(ctx, templateID)
}

func (s ProductService) ListPriceBooks(ctx context.Context, q *query.Query) (*query.Page[PriceBook], error) {
	return s.price_books.List(ctx, q)
}

func (s ProductService) FindPriceBook(ctx context.Context, id uint64) (*PriceBook, error) {
	return s.price_books.Find(ctx, id)
}

func (s ProductService) CreatePriceBook(ctx context.Context, price_book *PriceBook) (*PriceBook, error) {
	return s.price_books.Create(ctx, price_book)
}

func (s ProductService) UpdatePriceBook(ctx context.Context, price_book *PriceBook) (*PriceBook, error) {
	return s.price_books.Update(ctx, price_book)
}

func (s ProductService) DeletePriceBook(ctx context.Context, id uint64) error {
	return s.price_books.Delete(ctx, id)
}

func (s ProductService) ListRulesByPriceBook(ctx context.Context, price_bookID uint64) ([]*PriceRule, error) {
	return s.rules.ListByPriceBook(ctx, price_bookID)
}

func (s ProductService) CreateTemplate(ctx context.Context, template *Item) (*Item, error) {
	if err := s.validateTemplate(ctx, template); err != nil {
		return nil, err
	}
	return s.templates.Create(ctx, template)
}

func (s ProductService) UpdateTemplate(ctx context.Context, template *Item) (*Item, error) {
	if err := s.validateTemplate(ctx, template); err != nil {
		return nil, err
	}
	return s.templates.Update(ctx, template)
}

func (s ProductService) validateTemplate(ctx context.Context, template *Item) error {
	if template.CategoryID != nil {
		category, err := s.categories.Find(ctx, *template.CategoryID)
		if err != nil {
			return err
		}
		if category == nil {
			return ErrItemCategory
		}
		if template.OrganizationID != nil && category.OrganizationID != nil && *category.OrganizationID != *template.OrganizationID {
			return ErrItemCategoryOrganization
		}
	}

	if template.Type == "" {
		template.Type = "stockable"
	}
	if template.Tracking == "" {
		template.Tracking = "none"
	}
	if !validTemplateType(template.Type) {
		return ErrInvalidComputeType
	}
	return nil
}

func (s ProductService) CreateVariant(ctx context.Context, variant *ItemVariant) (*ItemVariant, error) {
	if err := s.ensureVariantTemplate(ctx, variant); err != nil {
		return nil, err
	}
	if err := s.ensureSkuUnique(ctx, variant.Sku, 0); err != nil {
		return nil, err
	}
	return s.variants.Create(ctx, variant)
}

func (s ProductService) GenerateVariants(ctx context.Context, templateID uint64, matrix []AttributeOption) ([]*ItemVariant, error) {
	template, err := s.templates.Find(ctx, templateID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, ErrItemNotFound
	}

	if len(matrix) == 0 {
		return nil, ErrEmptyAttributeMatrix
	}
	combinations := expandMatrix(matrix)
	if len(combinations) == 0 {
		return nil, ErrEmptyAttributeMatrix
	}

	variants := make([]*ItemVariant, 0, len(combinations))
	for _, combination := range combinations {
		attributes, err := marshalAttributes(combination)
		if err != nil {
			return nil, err
		}
		variant := &ItemVariant{
			ItemID:        templateID,
			AttributeJSON: attributes,
		}
		if err := s.ensureSkuUnique(ctx, variant.Sku, 0); err != nil {
			return nil, err
		}
		variants = append(variants, variant)
	}
	return variants, nil
}

func (s ProductService) CreateTemplateWithVariants(ctx context.Context, template *Item, matrix []AttributeOption) (*Item, []*ItemVariant, error) {
	created, err := s.CreateTemplate(ctx, template)
	if err != nil {
		return nil, nil, err
	}
	variants, err := s.GenerateVariants(ctx, created.ID, matrix)
	if err != nil {
		return nil, nil, err
	}
	if len(variants) == 0 {
		return created, nil, nil
	}
	for _, variant := range variants {
		if _, err := s.variants.Create(ctx, variant); err != nil {
			return nil, nil, err
		}
	}
	return created, variants, nil
}

func (s ProductService) CreateRule(ctx context.Context, price_bookID uint64, rule *PriceRule) (*PriceRule, error) {
	price_book, err := s.price_books.Find(ctx, price_bookID)
	if err != nil {
		return nil, err
	}
	if price_book == nil {
		return nil, ErrPriceBookNotFound
	}
	if !validAppliesTo[rule.AppliesTo] {
		return nil, ErrInvalidAppliesTo
	}
	if !validComputeType[rule.ComputeType] {
		return nil, ErrInvalidComputeType
	}
	if err := validateRuleScope(rule); err != nil {
		return nil, err
	}
	if rule.DiscountPct != nil && (*rule.DiscountPct < 0 || *rule.DiscountPct > 100) {
		return nil, ErrInvalidRuleValue
	}
	if rule.FixedPrice != nil && *rule.FixedPrice < 0 {
		return nil, ErrInvalidRuleValue
	}
	rule.PriceBookID = price_bookID
	return s.rules.Create(ctx, rule)
}

func (s ProductService) ResolvePrice(ctx context.Context, price_bookID, variantID uint64, qty amount.Amount, date time.Time) (ResolvedPrice, error) {
	price_book, err := s.price_books.Find(ctx, price_bookID)
	if err != nil {
		return ResolvedPrice{}, err
	}
	if price_book == nil {
		return ResolvedPrice{}, ErrPriceBookNotFound
	}

	variant, err := s.variants.Find(ctx, variantID)
	if err != nil {
		return ResolvedPrice{}, err
	}
	if variant == nil {
		return ResolvedPrice{}, ErrVariantNotFound
	}

	template, err := s.templates.Find(ctx, variant.ItemID)
	if err != nil {
		return ResolvedPrice{}, err
	}
	if template == nil {
		return ResolvedPrice{}, ErrItemNotFound
	}

	rules, err := s.rules.ListByPriceBook(ctx, price_bookID)
	if err != nil {
		return ResolvedPrice{}, err
	}

	templateOfProduct := func(itemID uint64) (uint64, bool, error) {
		ruleVariant, err := s.variants.Find(ctx, itemID)
		if err != nil {
			return 0, false, err
		}
		if ruleVariant == nil {
			return 0, false, nil
		}
		return ruleVariant.ItemID, true, nil
	}

	best, err := selectBestRule(rules, variant, template, qty, date, templateOfProduct)
	if err != nil {
		return ResolvedPrice{}, err
	}
	base := amount.FromFloat64(template.ListPrice)
	if best == nil {
		return ResolvedPrice{Price: base, BasePrice: base, ItemID: template.ID}, nil
	}

	price, err := computeRulePrice(best, base)
	if err != nil {
		return ResolvedPrice{}, err
	}
	return ResolvedPrice{
		Price:       price,
		BasePrice:   base,
		ItemID:      template.ID,
		VariantID:   variant.ID,
		RuleID:      best.ID,
		AppliesTo:   best.AppliesTo,
		ComputeType: best.ComputeType,
		DiscountPct: best.DiscountPct,
	}, nil
}

func (s ProductService) ResolveIncomeAccount(ctx context.Context, variantID uint64) (uint64, error) {
	variant, err := s.variants.Find(ctx, variantID)
	if err != nil {
		return 0, err
	}
	if variant == nil {
		return 0, ErrVariantNotFound
	}
	template, err := s.templates.Find(ctx, variant.ItemID)
	if err != nil {
		return 0, err
	}
	if template == nil {
		return 0, ErrItemNotFound
	}
	if template.CategoryID == nil {
		return 0, ErrNoIncomeAccount
	}
	category, err := s.categories.Find(ctx, *template.CategoryID)
	if err != nil {
		return 0, err
	}
	if category == nil || category.IncomeAccountID == nil {
		return 0, ErrNoIncomeAccount
	}
	if template.OrganizationID != nil && category.OrganizationID != nil && *category.OrganizationID != *template.OrganizationID {
		return 0, ErrItemCategoryOrganization
	}
	return *category.IncomeAccountID, nil
}

func (s ProductService) ResolveExpenseAccount(ctx context.Context, variantID uint64) (uint64, error) {
	variant, err := s.variants.Find(ctx, variantID)
	if err != nil {
		return 0, err
	}
	if variant == nil {
		return 0, ErrVariantNotFound
	}
	template, err := s.templates.Find(ctx, variant.ItemID)
	if err != nil {
		return 0, err
	}
	if template == nil {
		return 0, ErrItemNotFound
	}
	if template.CategoryID == nil {
		return 0, ErrNoExpenseAccount
	}
	category, err := s.categories.Find(ctx, *template.CategoryID)
	if err != nil {
		return 0, err
	}
	if category == nil || category.ExpenseAccountID == nil {
		return 0, ErrNoExpenseAccount
	}
	if template.OrganizationID != nil && category.OrganizationID != nil && *category.OrganizationID != *template.OrganizationID {
		return 0, ErrItemCategoryOrganization
	}
	return *category.ExpenseAccountID, nil
}

func (s ProductService) ResolveVariantOrganization(ctx context.Context, variantID uint64) (*uint64, error) {
	variant, err := s.variants.Find(ctx, variantID)
	if err != nil {
		return nil, err
	}
	if variant == nil {
		return nil, ErrVariantNotFound
	}
	template, err := s.templates.Find(ctx, variant.ItemID)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, ErrItemNotFound
	}
	return template.OrganizationID, nil
}

func (s ProductService) ensureVariantTemplate(ctx context.Context, variant *ItemVariant) error {
	template, err := s.templates.Find(ctx, variant.ItemID)
	if err != nil {
		return err
	}
	if template == nil {
		return ErrVariantTemplate
	}
	return nil
}

func (s ProductService) ensureSkuUnique(ctx context.Context, sku *string, excludeID uint64) error {
	if sku == nil || *sku == "" {
		return nil
	}
	existing, err := s.variants.Search(ctx, "sku", *sku)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return ErrSkuTaken
	}
	return nil
}

func validTemplateType(value string) bool {
	switch value {
	case "stockable", "consumable", "service", "digital":
		return true
	default:
		return false
	}
}

func validateRuleScope(rule *PriceRule) error {
	switch rule.AppliesTo {
	case AppliesToAll:
		return nil
	case AppliesToCategory:
		if rule.CategoryID == nil {
			return ErrInvalidRuleScope
		}
		return nil
	case AppliesToProduct, AppliesToVariant:
		if rule.ItemID == nil {
			return ErrInvalidRuleScope
		}
		return nil
	default:
		return ErrInvalidAppliesTo
	}
}

type ResolvedPrice struct {
	Price       amount.Amount `json:"price"`
	BasePrice   amount.Amount `json:"base_price"`
	ItemID      uint64        `json:"item_id"`
	VariantID   uint64        `json:"variant_id"`
	RuleID      uint64        `json:"rule_id,omitempty"`
	AppliesTo   string        `json:"applies_to,omitempty"`
	ComputeType string        `json:"compute_type,omitempty"`
	DiscountPct *float64      `json:"discount_pct,omitempty"`
}

func computeRulePrice(rule *PriceRule, base amount.Amount) (amount.Amount, error) {
	switch rule.ComputeType {
	case ComputeFixed:
		if rule.FixedPrice == nil {
			return amount.Amount{}, ErrPriceUnavailable
		}
		return amount.FromFloat64(*rule.FixedPrice).Round(4), nil
	case ComputePercent:
		if rule.DiscountPct == nil {
			return amount.Amount{}, ErrPriceUnavailable
		}
		factor := amount.FromFloat64(1 - *rule.DiscountPct/100)
		return base.Mul(factor).Round(4), nil
	default:
		return amount.Amount{}, ErrPriceUnavailable
	}
}

type ItemCategoryService struct {
	categories dao.CRUD[reference.ItemCategory]
}

func NewItemCategoryService(categories dao.CRUD[reference.ItemCategory]) ItemCategoryService {
	return ItemCategoryService{categories: categories}
}

func (s ItemCategoryService) List(ctx context.Context, q *query.Query) (*query.Page[reference.ItemCategory], error) {
	return s.categories.List(ctx, q)
}

func (s ItemCategoryService) Find(ctx context.Context, id uint64) (*reference.ItemCategory, error) {
	return s.categories.Find(ctx, id)
}

func (s ItemCategoryService) Create(ctx context.Context, category *reference.ItemCategory) (*reference.ItemCategory, error) {
	return s.categories.Create(ctx, category)
}

func (s ItemCategoryService) Update(ctx context.Context, category *reference.ItemCategory) (*reference.ItemCategory, error) {
	return s.categories.Update(ctx, category)
}

func (s ItemCategoryService) Delete(ctx context.Context, id uint64) error {
	return s.categories.Delete(ctx, id)
}
