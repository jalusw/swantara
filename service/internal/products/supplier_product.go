package products

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type SupplierProduct struct {
	model.Base
	ItemID              uint64     `gorm:"not null" json:"item_id"`
	SupplierID          uint64     `gorm:"not null" json:"supplier_id"`
	SupplierSku         *string    `json:"supplier_sku"`
	SupplierProductName *string    `json:"supplier_product_name"`
	MinQty              float64    `gorm:"type:numeric(18,4);default:1" json:"min_qty"`
	Price               *float64   `gorm:"type:numeric(18,4)" json:"price"`
	CurrencyCode        *string    `gorm:"type:char(3)" json:"currency_code"`
	LeadTimeDays        *int       `json:"lead_time_days"`
	Priority            int        `gorm:"default:10" json:"priority"`
	ValidFrom           *time.Time `gorm:"type:date" json:"valid_from"`
	ValidTo             *time.Time `gorm:"type:date" json:"valid_to"`
}

const DefaultSupplierPriority = 10
