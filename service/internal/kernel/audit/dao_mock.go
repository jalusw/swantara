package audit

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type LogDAOMock struct {
	ListFunc       func(ctx context.Context, q *query.Query) (*query.Page[Log], error)
	FindFunc       func(ctx context.Context, id uint64) (*Log, error)
	CreateFunc     func(ctx context.Context, entity *Log) (*Log, error)
	UpdateFunc     func(ctx context.Context, entity *Log) (*Log, error)
	DeleteFunc     func(ctx context.Context, id uint64) error
	HardDeleteFunc func(ctx context.Context, id uint64) error
	SearchFunc     func(ctx context.Context, field string, value any) (*Log, error)
}

func (m LogDAOMock) Search(ctx context.Context, field string, value any) (*Log, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, field, value)
	}
	return nil, nil
}

func (m LogDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[Log], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[Log]{Items: []*Log{}, Count: 0}, nil
}

func (m LogDAOMock) Find(ctx context.Context, id uint64) (*Log, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m LogDAOMock) Create(ctx context.Context, entity *Log) (*Log, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m LogDAOMock) Update(ctx context.Context, entity *Log) (*Log, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

func (m LogDAOMock) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m LogDAOMock) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}
