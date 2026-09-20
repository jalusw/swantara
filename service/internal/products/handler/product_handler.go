package handler

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/products"
)

type ProductHandler struct {
	svc products.ProductService
}

func NewProductHandler(svc products.ProductService) ProductHandler {
	return ProductHandler{svc: svc}
}

type ItemResponse struct {
	ID              uint64    `json:"id"`
	OrganizationID  *uint64   `json:"organization_id"`
	Name            string    `json:"name"`
	CategoryID      *uint64   `json:"category_id"`
	Type            string    `json:"type"`
	UnitID          *uint64   `json:"unit_id"`
	PurchaseUnitID  *uint64   `json:"purchase_unit_id"`
	ListPrice       float64   `json:"list_price"`
	StandardCost    float64   `json:"standard_cost"`
	IsPurchasable   bool      `json:"is_purchasable"`
	IsSellable      bool      `json:"is_sellable"`
	IsManufactured  bool      `json:"is_manufactured"`
	Tracking        string    `json:"tracking"`
	Weight          float64   `json:"weight"`
	Volume          float64   `json:"volume"`
	HsCode          *string   `json:"hs_code"`
	DescriptionSale *string   `json:"description_sale"`
	DescriptionPur  *string   `json:"description_purchase"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func newItemResponse(template *products.Item) ItemResponse {
	return ItemResponse{
		ID:              template.ID,
		OrganizationID:  template.OrganizationID,
		Name:            template.Name,
		CategoryID:      template.CategoryID,
		Type:            template.Type,
		UnitID:          template.UnitID,
		PurchaseUnitID:  template.PurchaseUnitID,
		ListPrice:       template.ListPrice,
		StandardCost:    template.StandardCost,
		IsPurchasable:   template.IsPurchasable,
		IsSellable:      template.IsSellable,
		IsManufactured:  template.IsManufactured,
		Tracking:        template.Tracking,
		Weight:          template.Weight,
		Volume:          template.Volume,
		HsCode:          template.HsCode,
		DescriptionSale: template.DescriptionSale,
		DescriptionPur:  template.DescriptionPur,
		Active:          template.Active,
		CreatedAt:       template.CreatedAt,
		UpdatedAt:       template.UpdatedAt,
	}
}

type ItemVariantResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ItemVariantResponse `json:"data"`
}
type ItemVariantResponse struct {
	ID            uint64          `json:"id"`
	ItemID        uint64          `json:"item_id"`
	Sku           *string         `json:"sku"`
	Barcode       *string         `json:"barcode"`
	AttributeJSON json.RawMessage `json:"attribute_json" swaggertype:"object"`
	ExtraCost     float64         `json:"extra_cost"`
	Active        bool            `json:"active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func newItemVariantResponse(variant *products.ItemVariant) ItemVariantResponse {
	return ItemVariantResponse{
		ID:            variant.ID,
		ItemID:        variant.ItemID,
		Sku:           variant.Sku,
		Barcode:       variant.Barcode,
		AttributeJSON: variant.AttributeJSON,
		ExtraCost:     variant.ExtraCost,
		Active:        variant.Active,
		CreatedAt:     variant.CreatedAt,
		UpdatedAt:     variant.UpdatedAt,
	}
}

func writeProductError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, products.ErrItemCategory):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item category does not exist.", nil)
	case errors.Is(err, products.ErrItemCategoryOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item category does not belong to this organization.", nil)
	case errors.Is(err, products.ErrItemNotFound):
		return httpx.CreateNotFoundResponse(c, "Item not found.")
	case errors.Is(err, products.ErrVariantNotFound):
		return httpx.CreateNotFoundResponse(c, "Item variant not found.")
	case errors.Is(err, products.ErrVariantTemplate):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item template does not exist.", nil)
	case errors.Is(err, products.ErrSkuTaken):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A item with this SKU already exists.", nil)
	case errors.Is(err, products.ErrEmptyAttributeMatrix):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Variant generation requires at least one attribute.", nil)
	default:
		httpx.RequestLog(c).Error("item write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save item.", err)
	}
}
