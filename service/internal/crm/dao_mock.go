package crm

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type ProspectDAOMock struct {
	dao.CRUDMock[Prospect]
	ListOpenFunc  func(ctx context.Context, organizationID *uint64) ([]*Prospect, error)
	CountWonFunc  func(ctx context.Context, organizationID *uint64) (int64, error)
	CountLostFunc func(ctx context.Context, organizationID *uint64) (int64, error)
}

func (m ProspectDAOMock) ListOpen(ctx context.Context, organizationID *uint64) ([]*Prospect, error) {
	if m.ListOpenFunc != nil {
		return m.ListOpenFunc(ctx, organizationID)
	}
	return []*Prospect{}, nil
}

func (m ProspectDAOMock) CountWon(ctx context.Context, organizationID *uint64) (int64, error) {
	if m.CountWonFunc != nil {
		return m.CountWonFunc(ctx, organizationID)
	}
	return 0, nil
}

func (m ProspectDAOMock) CountLost(ctx context.Context, organizationID *uint64) (int64, error) {
	if m.CountLostFunc != nil {
		return m.CountLostFunc(ctx, organizationID)
	}
	return 0, nil
}

type ProspectActivityDAOMock struct {
	dao.CRUDMock[ProspectActivity]
}

func (m ProspectActivityDAOMock) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[ProspectActivity], error) {
	return m.List(ctx, q)
}

func (m ProspectActivityDAOMock) FindInOrg(ctx context.Context, id, organizationID uint64) (*ProspectActivity, error) {
	return m.Find(ctx, id)
}
