package products

import (
	"encoding/json"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type Item struct {
	model.Base
	OrganizationID  *uint64 `json:"organization_id"`
	Name            string  `gorm:"not null" json:"name"`
	CategoryID      *uint64 `json:"category_id"`
	Type            string  `gorm:"type:text" json:"type"`
	UnitID          *uint64 `json:"unit_id"`
	PurchaseUnitID  *uint64 `json:"purchase_unit_id"`
	ListPrice       float64 `gorm:"type:numeric(18,4)" json:"list_price"`
	StandardCost    float64 `gorm:"type:numeric(18,4)" json:"standard_cost"`
	IsPurchasable   bool    `gorm:"default:true" json:"is_purchasable"`
	IsSellable      bool    `gorm:"default:true" json:"is_sellable"`
	IsManufactured  bool    `gorm:"default:false" json:"is_manufactured"`
	Tracking        string  `gorm:"type:text" json:"tracking"`
	Weight          float64 `gorm:"type:numeric(18,4)" json:"weight"`
	Volume          float64 `gorm:"type:numeric(18,4)" json:"volume"`
	HsCode          *string `json:"hs_code"`
	DescriptionSale *string `json:"description_sale"`
	DescriptionPur  *string `gorm:"column:description_purchase" json:"description_purchase"`
	Active          bool    `gorm:"default:true" json:"active"`
}

type ItemVariant struct {
	model.Base
	ItemID        uint64          `gorm:"not null" json:"item_id"`
	Sku           *string         `gorm:"uniqueIndex" json:"sku"`
	Barcode       *string         `json:"barcode"`
	AttributeJSON json.RawMessage `gorm:"type:jsonb" json:"attribute_json" swaggertype:"object"`
	ExtraCost     float64         `gorm:"type:numeric(18,4);default:0" json:"extra_cost"`
	Active        bool            `gorm:"default:true" json:"active"`
}

type PriceBook struct {
	model.Base
	Name           string  `gorm:"not null" json:"name"`
	CurrencyCode   *string `gorm:"type:char(3)" json:"currency_code"`
	OrganizationID *uint64 `json:"organization_id"`
	Active         bool    `gorm:"default:true" json:"active"`
}

type PriceRule struct {
	model.Base
	PriceBookID uint64     `json:"price_book_id"`
	AppliesTo   string     `gorm:"type:text" json:"applies_to"`
	ItemID      *uint64    `json:"item_id"`
	CategoryID  *uint64    `json:"category_id"`
	MinQty      float64    `gorm:"type:numeric(18,4);default:0" json:"min_qty"`
	ComputeType string     `gorm:"type:text" json:"compute_type"`
	FixedPrice  *float64   `gorm:"type:numeric(18,4)" json:"fixed_price"`
	DiscountPct *float64   `gorm:"type:numeric(8,4)" json:"discount_pct"`
	DateStart   *time.Time `gorm:"type:date" json:"date_start"`
	DateEnd     *time.Time `gorm:"type:date" json:"date_end"`
}

const (
	AppliesToAll      = "all"
	AppliesToCategory = "category"
	AppliesToProduct  = "item"
	AppliesToVariant  = "variant"

	ComputeFixed   = "fixed"
	ComputePercent = "percent"
	ComputeFormula = "formula"
)

var validAppliesTo = map[string]bool{
	AppliesToAll:      true,
	AppliesToCategory: true,
	AppliesToProduct:  true,
	AppliesToVariant:  true,
}

var validComputeType = map[string]bool{
	ComputeFixed:   true,
	ComputePercent: true,
	ComputeFormula: true,
}
