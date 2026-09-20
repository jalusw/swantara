package quality

import (
	"errors"

	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

var (
	ErrQualityPointNotFound   = errors.New("quality point not found")
	ErrQualityCheckNotFound   = errors.New("quality check not found")
	ErrQualityCheckDone       = errors.New("quality check already resolved")
	ErrQualityCheckNoValue    = errors.New("measured value required for measurement check")
	ErrQualityCheckNoProduct  = errors.New("quality check requires a item")
	ErrQualityAlertNotFound   = errors.New("quality alert not found")
	ErrQualityAlertState      = errors.New("quality alert is not in the required state")
	ErrQualityNoChecksCreated = errors.New("no quality checks were created")
	ErrQualityScrapRouter     = errors.New("quality scrap router is not configured")

	ErrScrapLocation     = inventory.ErrScrapLocation
	ErrScrapAccount      = inventory.ErrScrapAccount
	ErrMovementNotFound  = inventory.ErrMovementNotFound
	ErrInsufficientStock = inventory.ErrInsufficientStock
)
