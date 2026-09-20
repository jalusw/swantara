package systemconfig

import (
	"context"
	"encoding/json"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type Lookup interface {
	List(ctx context.Context, q *query.Query) (*query.Page[reference.SystemConfig], error)
}

func Read(ctx context.Context, configs Lookup, organizationID uint64, key string, target any) error {
	page, err := configs.List(ctx, &query.Query{Filters: []query.Filter{{Field: "key", Operator: query.Equal, Value: key}}})
	if err != nil {
		return err
	}
	for _, config := range page.Items {
		if config.Key == key && (config.OrganizationID == nil || *config.OrganizationID == organizationID) {
			if len(config.Value) == 0 {
				return nil
			}
			return json.Unmarshal(config.Value, target)
		}
	}
	return nil
}
