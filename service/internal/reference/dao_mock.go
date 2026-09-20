package reference

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type AccountDAOMock struct {
	ListFunc   func(ctx context.Context, q *query.Query) (*query.Page[Account], error)
	CreateFunc func(ctx context.Context, entity *Account) (*Account, error)
}

func (m AccountDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[Account], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[Account]{}, nil
}

func (m AccountDAOMock) Create(ctx context.Context, entity *Account) (*Account, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

type TaxDAOMock struct {
	ListFunc   func(ctx context.Context, q *query.Query) (*query.Page[Tax], error)
	CreateFunc func(ctx context.Context, entity *Tax) (*Tax, error)
	UpdateFunc func(ctx context.Context, entity *Tax) (*Tax, error)
}

func (m TaxDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[Tax], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[Tax]{}, nil
}

func (m TaxDAOMock) Create(ctx context.Context, entity *Tax) (*Tax, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m TaxDAOMock) Update(ctx context.Context, entity *Tax) (*Tax, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

type WithholdingTaxDAOMock struct {
	ListFunc   func(ctx context.Context, q *query.Query) (*query.Page[WithholdingTax], error)
	CreateFunc func(ctx context.Context, entity *WithholdingTax) (*WithholdingTax, error)
}

func (m WithholdingTaxDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[WithholdingTax], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[WithholdingTax]{}, nil
}

func (m WithholdingTaxDAOMock) Create(ctx context.Context, entity *WithholdingTax) (*WithholdingTax, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}
