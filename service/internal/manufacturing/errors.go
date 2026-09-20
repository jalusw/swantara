package manufacturing

import (
	"errors"

	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

var (
	ErrRecipeNotFound    = errors.New("recipe not found")
	ErrRecipeItem        = errors.New("recipe item variant does not exist")
	ErrInvalidRecipeType = errors.New("invalid recipe type")
	ErrRecipeComponent   = errors.New("recipe line component variant does not exist")
	ErrRecipeLineQty     = errors.New("recipe line quantity must be greater than zero")
	ErrRecipeLineScrap   = errors.New("recipe line scrap percentage cannot be negative")
	ErrRecipeQty         = errors.New("recipe quantity cannot be negative")
	ErrRecipeOutputQty   = errors.New("recipe output quantity must be greater than zero")
	ErrRecipeNoLines     = errors.New("recipe has no lines")

	ErrProductionOrderNotFound     = errors.New("production order not found")
	ErrProductionOrderState        = errors.New("production order state does not allow this operation")
	ErrProductionOrderItem         = errors.New("production order item variant does not exist")
	ErrProductionOrderRecipe       = errors.New("production order recipe not found")
	ErrProductionOrderQty          = errors.New("production order quantity must be greater than zero")
	ErrProductionOrderOrganization = errors.New("production order requires an organization")
	ErrProductionOrderSequence     = errors.New("production order document sequence not configured")
	ErrProductionOrderLocation     = errors.New("production order location does not exist")
	ErrProductionOrderReservation  = errors.New("failed to reserve production order components")
	ErrConsumedMaterial            = errors.New("production order has no components")

	ErrShopTaskNotFound   = errors.New("shop task not found")
	ErrShopTaskState      = errors.New("shop task state does not allow this operation")
	ErrShopTaskQuantity   = errors.New("shop task quantity must be greater than zero")
	ErrShopTaskLabor      = errors.New("shop task labor hours must be greater than zero")
	ErrShopTaskStep       = errors.New("shop task routing operation does not exist")
	ErrShopTaskWorkCenter = errors.New("shop task work center does not exist")
	ErrShopTaskMismatch   = errors.New("shop task does not belong to the production order")
	ErrProductionLocation = errors.New("production location does not exist")
	ErrConsumeComponent   = errors.New("component does not belong to the production order")
	ErrConsumeQuantity    = errors.New("consume quantity exceeds the remaining component quantity")
	ErrConsumeMovement    = errors.New("failed to create consumption movement")
	ErrProduceQuantity    = errors.New("produce quantity exceeds the remaining quantity to produce")
	ErrProduceMovement    = errors.New("failed to create production movement")
	ErrWIPAccount         = errors.New("work in progress account is required")
	ErrLaborAccount       = errors.New("applied labor account is required")
	ErrVarianceAccount    = errors.New("production variance account is required")

	ErrPlanningRunNotRequired = errors.New("planning run horizon must be greater than zero")
	ErrPlanningItem           = errors.New("planning planned order item variant does not exist")
	ErrPlanningNoRecipe       = errors.New("planning production planned order requires a recipe")
	ErrPlanningPlannedType    = errors.New("planning planned order type is not supported")
	ErrPlanningPlannedState   = errors.New("planning planned order is already confirmed")
	ErrPlanningSupplier       = errors.New("planning purchase planned order requires a supplier")
	ErrPlanningSrcWarehouse   = errors.New("planning transfer planned order requires source and destination warehouse")

	ErrOutsideProcessingNotFound         = errors.New("outside processing order not found")
	ErrOutsideProcessingDuplicate        = errors.New("outside processing order already exists for this production order")
	ErrOutsideProcessingState            = errors.New("outside processing order state does not allow this operation")
	ErrOutsideProcessingNotSubcontracted = errors.New("production order recipe is not a subcontract recipe")
	ErrOutsideProcessingSupplier         = errors.New("outside processing order requires an active supplier")
	ErrOutsideProcessingOrderState       = errors.New("production order state does not allow this operation")
	ErrOutsideProcessingPurchaseOrder    = errors.New("subcontract purchase order could not be created")
	ErrOutsideProcessingSupplierLocation = errors.New("supplier location does not exist")
	ErrOutsideProcessingComponents       = errors.New("outside processing order has no components to send")
	ErrOutsideProcessingOperation        = errors.New("subcontract operation amount could not be determined")

	ErrHoldOverflow      = inventory.ErrHoldOverflow
	ErrBalanceNotFound   = inventory.ErrBalanceNotFound
	ErrMovementNotFound  = inventory.ErrMovementNotFound
	ErrMovementState     = inventory.ErrMovementState
	ErrInsufficientStock = inventory.ErrInsufficientStock
	ErrNotConsumption    = inventory.ErrNotConsumption
	ErrNotProduction     = inventory.ErrNotProduction
	ErrValuationAccount  = inventory.ErrValuationAccount
	ErrNegativeCost      = inventory.ErrNegativeCost
	ErrBatchRequired     = inventory.ErrBatchRequired
)
