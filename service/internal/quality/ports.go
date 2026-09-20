package quality

import (
	"context"
	"time"
)

type ScrapRouter interface {
	RouteToScrap(ctx context.Context, organizationID, shipmentID uint64, itemID uint64, journalID, expenseAccountID uint64, date time.Time) error
}
