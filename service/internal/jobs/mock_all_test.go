package jobs

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestJobDAOMock_All(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	bare := DAOMock{}
	if _, err := bare.List(ctx, q); err != nil {
		t.Errorf("List = %v", err)
	}
	if _, err := bare.Search(ctx, "f", 1); err != nil {
		t.Errorf("Search = %v", err)
	}
	if _, err := bare.Find(ctx, 1); err != nil {
		t.Errorf("Find = %v", err)
	}
	if err := bare.Delete(ctx, 1); err != nil {
		t.Errorf("Delete = %v", err)
	}
	if err := bare.HardDelete(ctx, 1); err != nil {
		t.Errorf("HardDelete = %v", err)
	}

	wired := DAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[JobRun], error) {
			return &query.Page[JobRun]{}, nil
		},
		SearchFunc:     func(_ context.Context, _ string, _ any) (*JobRun, error) { return &JobRun{}, nil },
		FindFunc:       func(_ context.Context, _ uint64) (*JobRun, error) { return &JobRun{}, nil },
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
	if err := wired.Delete(ctx, 1); err != nil {
		t.Errorf("Delete = %v", err)
	}
	if err := wired.HardDelete(ctx, 1); err != nil {
		t.Errorf("HardDelete = %v", err)
	}
}
