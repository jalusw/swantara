package dao

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type mockEntity struct {
	ID uint64
}

func (m mockEntity) GetID() uint64 { return m.ID }

func TestCRUDMock_All(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	bare := CRUDMock[mockEntity]{}
	if _, err := bare.List(ctx, q); err != nil {
		t.Errorf("List = %v", err)
	}
	if _, err := bare.Search(ctx, "f", 1); err != nil {
		t.Errorf("Search = %v", err)
	}
	if _, err := bare.Find(ctx, 1); err != nil {
		t.Errorf("Find = %v", err)
	}
	if _, err := bare.Create(ctx, &mockEntity{}); err != nil {
		t.Errorf("Create = %v", err)
	}
	if _, err := bare.Update(ctx, &mockEntity{}); err != nil {
		t.Errorf("Update = %v", err)
	}
	if err := bare.Delete(ctx, 1); err != nil {
		t.Errorf("Delete = %v", err)
	}
	if err := bare.HardDelete(ctx, 1); err != nil {
		t.Errorf("HardDelete = %v", err)
	}

	wired := CRUDMock[mockEntity]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[mockEntity], error) {
			return &query.Page[mockEntity]{}, nil
		},
		SearchFunc:     func(_ context.Context, _ string, _ any) (*mockEntity, error) { return &mockEntity{}, nil },
		FindFunc:       func(_ context.Context, _ uint64) (*mockEntity, error) { return &mockEntity{}, nil },
		CreateFunc:     func(_ context.Context, e *mockEntity) (*mockEntity, error) { return e, nil },
		UpdateFunc:     func(_ context.Context, e *mockEntity) (*mockEntity, error) { return e, nil },
		DeleteFunc:     func(_ context.Context, _ uint64) error { return nil },
		HardDeleteFunc: func(_ context.Context, _ uint64) error { return nil },
	}
	if _, err := wired.List(ctx, q); err != nil {
		t.Errorf("List = %v", err)
	}
	if _, err := wired.Search(ctx, "f", 1); err != nil {
		t.Errorf("Search = %v", err)
	}
	if _, err := wired.Find(ctx, 1); err != nil {
		t.Errorf("Find = %v", err)
	}
	if _, err := wired.Create(ctx, &mockEntity{}); err != nil {
		t.Errorf("Create = %v", err)
	}
	if _, err := wired.Update(ctx, &mockEntity{}); err != nil {
		t.Errorf("Update = %v", err)
	}
	if err := wired.Delete(ctx, 1); err != nil {
		t.Errorf("Delete = %v", err)
	}
	if err := wired.HardDelete(ctx, 1); err != nil {
		t.Errorf("HardDelete = %v", err)
	}
}
