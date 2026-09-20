package handler

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type qualityScrapRouterMock struct {
	route func(ctx context.Context, organizationID, shipmentID, itemID, journalID, expenseAccountID uint64, date time.Time) error
}

func (m qualityScrapRouterMock) RouteToScrap(ctx context.Context, organizationID, shipmentID, itemID, journalID, expenseAccountID uint64, date time.Time) error {
	if m.route != nil {
		return m.route(ctx, organizationID, shipmentID, itemID, journalID, expenseAccountID, date)
	}
	return nil
}

func daoListFunc[E any](item *E) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[E], error) {
			return &query.Page[E]{Items: []*E{item}, Count: 1}, nil
		},
	}
}

func daoListErrFunc[E any](err error) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[E], error) {
			return nil, err
		},
	}
}

func daoFindFunc[E any](entity *E) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		FindFunc: func(_ context.Context, _ uint64) (*E, error) {
			return entity, nil
		},
	}
}

func daoFindErrFunc[E any](err error) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		FindFunc: func(_ context.Context, _ uint64) (*E, error) {
			return nil, err
		},
	}
}

func daoCreateFunc[E any](entity *E) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		CreateFunc: func(_ context.Context, _ *E) (*E, error) {
			return entity, nil
		},
	}
}

func daoCreateErrFunc[E any](err error) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{
		CreateFunc: func(_ context.Context, _ *E) (*E, error) {
			return nil, err
		},
	}
}
