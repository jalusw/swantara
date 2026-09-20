package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type PostedMoveChecker struct {
	movements JournalEntryDAO
}

func NewPostedMoveChecker(movements JournalEntryDAO) PostedMoveChecker {
	return PostedMoveChecker{movements: movements}
}

func (c PostedMoveChecker) HasPosted(ctx context.Context, organizationID uint64) (bool, error) {
	page, err := c.movements.List(ctx, &query.Query{
		Filters: []query.Filter{
			{Field: "organization_id", Operator: query.Equal, Value: organizationID},
			{Field: "state", Operator: query.Equal, Value: EntryStatePosted},
		},
		Pagination: &query.Pagination{Page: 1, Size: 1},
	})
	if err != nil {
		return false, err
	}
	return page.Count > 0, nil
}
