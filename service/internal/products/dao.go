package products

import (
	"context"
	"errors"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type ItemDAO interface {
	dao.CRUD[Item]
	CreateWithVariants(ctx context.Context, template *Item, variants []*ItemVariant) (*Item, error)
}

type itemDAO struct {
	dao.Base[Item]
	db *gorm.DB
}

func NewItemDAO(db *gorm.DB) ItemDAO {
	return itemDAO{Base: dao.NewBase[Item](db), db: db}
}

func (d itemDAO) CreateWithVariants(ctx context.Context, template *Item, variants []*ItemVariant) (*Item, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(template).Error; err != nil {
			return err
		}
		for _, variant := range variants {
			variant.ItemID = template.ID
			if err := tx.Create(variant).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return template, nil
}

type ItemVariantDAO interface {
	dao.CRUD[ItemVariant]
	ListByTemplate(ctx context.Context, templateID uint64) ([]*ItemVariant, error)
	VariantExists(ctx context.Context, id uint64) (bool, error)
}

type itemVariantDAO struct {
	dao.Base[ItemVariant]
}

func NewItemVariantDAO(db *gorm.DB) ItemVariantDAO {
	return itemVariantDAO{Base: dao.NewBase[ItemVariant](db)}
}

func (d itemVariantDAO) ListByTemplate(ctx context.Context, templateID uint64) ([]*ItemVariant, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "item_id", Operator: query.Equal, Value: templateID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d itemVariantDAO) VariantExists(ctx context.Context, id uint64) (bool, error) {
	found, err := d.Find(ctx, id)
	if err != nil {
		return false, err
	}
	return found != nil, nil
}

type PriceBookDAO interface {
	dao.CRUD[PriceBook]
}

type price_bookDAO struct {
	dao.Base[PriceBook]
}

func NewPriceBookDAO(db *gorm.DB) PriceBookDAO {
	return price_bookDAO{Base: dao.NewBase[PriceBook](db)}
}

type PriceRuleDAO interface {
	dao.CRUD[PriceRule]
	ListByPriceBook(ctx context.Context, price_bookID uint64) ([]*PriceRule, error)
}

type price_bookRuleDAO struct {
	dao.Base[PriceRule]
}

func NewPriceRuleDAO(db *gorm.DB) PriceRuleDAO {
	return price_bookRuleDAO{Base: dao.NewBase[PriceRule](db)}
}

func (d price_bookRuleDAO) ListByPriceBook(ctx context.Context, price_bookID uint64) ([]*PriceRule, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "price_book_id", Operator: query.Equal, Value: price_bookID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type SupplierProductDAO interface {
	dao.CRUD[SupplierProduct]
	ListByItem(ctx context.Context, itemID uint64) ([]*SupplierProduct, error)
	ListBySupplier(ctx context.Context, supplierID uint64) ([]*SupplierProduct, error)
	ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[SupplierProduct], error)
	FindInOrg(ctx context.Context, id, organizationID uint64) (*SupplierProduct, error)
}

type supplierProductDAO struct {
	dao.Base[SupplierProduct]
	db *gorm.DB
}

func NewSupplierProductDAO(db *gorm.DB) SupplierProductDAO {
	return supplierProductDAO{Base: dao.NewBase[SupplierProduct](db), db: db}
}

func (d supplierProductDAO) ListByItem(ctx context.Context, itemID uint64) ([]*SupplierProduct, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "item_id", Operator: query.Equal, Value: itemID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d supplierProductDAO) ListBySupplier(ctx context.Context, supplierID uint64) ([]*SupplierProduct, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "supplier_id", Operator: query.Equal, Value: supplierID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d supplierProductDAO) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[SupplierProduct], error) {
	var count int64
	var entities []SupplierProduct

	join := "JOIN item_variants ON item_variants.id = supplier_products.item_id " +
		"JOIN items ON items.id = item_variants.item_id " +
		"AND items.organization_id = ?"

	countTx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID)
	if q != nil {
		countTx = query.ApplyFilters(countTx, q.Filters)
	}
	if err := countTx.Count(&count).Error; err != nil {
		return nil, err
	}

	tx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID)
	if q != nil {
		tx = query.ApplyQuery(tx, q)
	}
	if err := tx.Find(&entities).Error; err != nil {
		return nil, err
	}

	items := make([]*SupplierProduct, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}

	return &query.Page[SupplierProduct]{Items: items, Count: count}, nil
}

func (d supplierProductDAO) FindInOrg(ctx context.Context, id, organizationID uint64) (*SupplierProduct, error) {
	var offer SupplierProduct
	err := d.db.WithContext(ctx).Model(&SupplierProduct{}).
		Joins("JOIN item_variants ON item_variants.id = supplier_products.item_id "+
			"JOIN items ON items.id = item_variants.item_id "+
			"AND items.organization_id = ?", organizationID).
		Where("supplier_products.id = ?", id).
		Take(&offer).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &offer, nil
}
