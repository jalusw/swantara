package inventory

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestInboundCostConfigSource_JournalID_ReadsNumberConfig(t *testing.T) {
	ctx := context.Background()
	configs := configLookupMock{
		listFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configInboundCostJournalID, OrganizationID: helper.Ptr(uint64(10)), Value: []byte(`42`)},
			}}, nil
		},
	}
	source := NewInboundCostConfigSource(configs)

	journalID, err := source.JournalID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if journalID != 42 {
		t.Errorf("journal id = %d, want 42", journalID)
	}
}

func TestInboundCostConfigSource_JournalID_AppliesGlobalConfig(t *testing.T) {
	ctx := context.Background()
	configs := configLookupMock{
		listFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configInboundCostJournalID, Value: []byte(`42`)},
			}}, nil
		},
	}
	source := NewInboundCostConfigSource(configs)

	journalID, err := source.JournalID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if journalID != 42 {
		t.Errorf("journal id = %d, want 42 from global config", journalID)
	}
}

func TestInboundCostConfigSource_JournalID_IgnoresOtherOrganization(t *testing.T) {
	ctx := context.Background()
	configs := configLookupMock{
		listFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configInboundCostJournalID, OrganizationID: helper.Ptr(uint64(20)), Value: []byte(`42`)},
			}}, nil
		},
	}
	source := NewInboundCostConfigSource(configs)

	journalID, err := source.JournalID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if journalID != 0 {
		t.Errorf("journal id = %d, want 0 for a different organization", journalID)
	}
}

func TestInboundCostConfigSource_JournalID_IgnoresDifferentKey(t *testing.T) {
	ctx := context.Background()
	configs := configLookupMock{
		listFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: "other.key", OrganizationID: helper.Ptr(uint64(10)), Value: []byte(`42`)},
			}}, nil
		},
	}
	source := NewInboundCostConfigSource(configs)

	journalID, err := source.JournalID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if journalID != 0 {
		t.Errorf("journal id = %d, want 0 for a different key", journalID)
	}
}

func TestInboundCostConfigSource_JournalID_EmptyValueLeavesZero(t *testing.T) {
	ctx := context.Background()
	configs := configLookupMock{
		listFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configInboundCostJournalID, OrganizationID: helper.Ptr(uint64(10))},
			}}, nil
		},
	}
	source := NewInboundCostConfigSource(configs)

	journalID, err := source.JournalID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if journalID != 0 {
		t.Errorf("journal id = %d, want 0 when the config value is empty", journalID)
	}
}

func TestInboundCostConfigSource_JournalID_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	configs := configLookupMock{
		listFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return nil, ErrWarehouseNotFound
		},
	}
	source := NewInboundCostConfigSource(configs)

	_, err := source.JournalID(ctx, 10)
	if helper.AssertError(t, err, true, ErrWarehouseNotFound) {
		return
	}
}

func TestInboundCostConfigSource_JournalID_PropagatesUnmarshalError(t *testing.T) {
	ctx := context.Background()
	configs := configLookupMock{
		listFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configInboundCostJournalID, OrganizationID: helper.Ptr(uint64(10)), Value: []byte(`not-json`)},
			}}, nil
		},
	}
	source := NewInboundCostConfigSource(configs)

	if _, err := source.JournalID(ctx, 10); err == nil {
		t.Fatal("expected error when config value is not valid JSON")
	}
}

type configLookupMock struct {
	listFunc func(ctx context.Context, q *query.Query) (*query.Page[reference.SystemConfig], error)
}

func (m configLookupMock) List(ctx context.Context, q *query.Query) (*query.Page[reference.SystemConfig], error) {
	return m.listFunc(ctx, q)
}
