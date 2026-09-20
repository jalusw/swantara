package dao

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type CRUDMock[E any] struct {
	ListFunc       func(ctx context.Context, q *query.Query) (*query.Page[E], error)
	SearchFunc     func(ctx context.Context, field string, value any) (*E, error)
	FindFunc       func(ctx context.Context, id uint64) (*E, error)
	CreateFunc     func(ctx context.Context, entity *E) (*E, error)
	UpdateFunc     func(ctx context.Context, entity *E) (*E, error)
	DeleteFunc     func(ctx context.Context, id uint64) error
	HardDeleteFunc func(ctx context.Context, id uint64) error
}

func (m CRUDMock[E]) List(ctx context.Context, q *query.Query) (*query.Page[E], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[E]{Items: []*E{}, Count: 0}, nil
}

func (m CRUDMock[E]) Search(ctx context.Context, field string, value any) (*E, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, field, value)
	}
	return nil, nil
}

func (m CRUDMock[E]) Find(ctx context.Context, id uint64) (*E, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m CRUDMock[E]) Create(ctx context.Context, entity *E) (*E, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m CRUDMock[E]) Update(ctx context.Context, entity *E) (*E, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

func (m CRUDMock[E]) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m CRUDMock[E]) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}
