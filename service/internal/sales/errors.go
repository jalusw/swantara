package sales

import (
	"errors"

	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/products"
)

var (
	ErrOrderNotFound              = errors.New("sale order not found")
	ErrLineNotFound               = errors.New("sale order line not found")
	ErrOrderState                 = errors.New("sale order is not in the required state")
	ErrOrderNoLines               = errors.New("sale order must have at least one line")
	ErrOrderQty                   = errors.New("sale order line quantity must be positive")
	ErrOrderContact               = errors.New("sale order contact does not exist")
	ErrOrderPriceBook             = errors.New("sale order price_book does not exist")
	ErrOrderWarehouse             = errors.New("sale order warehouse does not exist")
	ErrOrderContactOrganization   = errors.New("sale order contact does not belong to the organization")
	ErrOrderPriceBookOrganization = errors.New("sale order price_book does not belong to the organization")
	ErrOrderWarehouseOrganization = errors.New("sale order warehouse does not belong to the organization")
	ErrOrderVariant               = errors.New("sale order item variant does not exist")
	ErrOrderVariantOrganization   = errors.New("sale order line item does not belong to the organization")
	ErrOrderLead                  = errors.New("sale order source opportunity does not exist")
	ErrOrderLeadNotWon            = errors.New("sale order source opportunity is not won")
	ErrOrderLeadOrganization      = errors.New("sale order source opportunity does not belong to the organization")
	ErrOrderTax                   = errors.New("sale order tax does not exist")
	ErrOrderTaxInvalid            = errors.New("sale order tax is not a sale tax")
	ErrOrderTaxOrganization       = errors.New("sale order tax does not belong to the organization")
	ErrOrderCurrency              = errors.New("sale order currency does not exist")
	ErrOrderLocation              = errors.New("sale order warehouse stock location not found")
	ErrOrderCustomerLocation      = errors.New("sale order customer location not found")
	ErrOrderStockUnavailable      = errors.New("sale order line quantity is not available to reserve")
	ErrOrderDiscount              = errors.New("sale order line discount must be between 0 and 100")
	ErrOrderSequenceNotFound      = errors.New("sale order document sequence not configured")
	ErrOrderShipmentNotFound      = errors.New("sale order outgoing shipment not found")
	ErrOrderNothingToDeliver      = errors.New("sale order has nothing ready to deliver")
	ErrOrderNoInvoices            = errors.New("sale order has no open invoices to pay")

	ErrBalanceNotFound = inventory.ErrBalanceNotFound
	ErrHoldOverflow    = inventory.ErrHoldOverflow
	ErrNoIncomeAccount = products.ErrNoIncomeAccount
)
