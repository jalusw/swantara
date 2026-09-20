package subscription

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type ConfigLookupMock struct {
	ListFunc func(ctx context.Context, q *query.Query) (*query.Page[reference.SystemConfig], error)
}

func (m ConfigLookupMock) List(ctx context.Context, q *query.Query) (*query.Page[reference.SystemConfig], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{}, Count: 0}, nil
}
