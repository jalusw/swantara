package subscription

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestSubscriptionConfigSource_JournalID_ReturnsValue(t *testing.T) {
	ctx := context.Background()
	source := NewSubscriptionConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), Key: configSubscriptionJournalID, Value: []byte(`7`)},
			}}, nil
		},
	})

	value, err := source.JournalID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 7 {
		t.Errorf("value = %d, want 7", value)
	}
}

func TestSubscriptionConfigSource_JournalID_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	source := NewSubscriptionConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return nil, errors.New("db down")
		},
	})

	_, err := source.JournalID(ctx, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestSubscriptionConfigSource_JournalID_ReturnsZeroWhenNotConfigured(t *testing.T) {
	ctx := context.Background()
	source := NewSubscriptionConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(20)), Key: configSubscriptionJournalID, Value: []byte(`7`)},
				{Base: model.Base{ID: 2}, OrganizationID: helper.Ptr(uint64(10)), Key: configSubscriptionDeferredRevenueAcct, Value: []byte(`300`)},
			}}, nil
		},
	})

	value, err := source.JournalID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 0 {
		t.Errorf("value = %d, want 0", value)
	}
}

func TestSubscriptionConfigSource_JournalID_ReturnsZeroForEmptyValue(t *testing.T) {
	ctx := context.Background()
	source := NewSubscriptionConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), Key: configSubscriptionJournalID},
			}}, nil
		},
	})

	value, err := source.JournalID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 0 {
		t.Errorf("value = %d, want 0", value)
	}
}

func TestSubscriptionConfigSource_DeferredRevenueAccountID_ReturnsValue(t *testing.T) {
	ctx := context.Background()
	source := NewSubscriptionConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Base: model.Base{ID: 2}, Key: configSubscriptionDeferredRevenueAcct, Value: []byte(`300`)},
			}}, nil
		},
	})

	value, err := source.DeferredRevenueAccountID(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 300 {
		t.Errorf("value = %d, want 300", value)
	}
}

func TestSubscriptionConfigSource_DeferredRevenueAccountID_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	source := NewSubscriptionConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return nil, errors.New("db down")
		},
	})

	_, err := source.DeferredRevenueAccountID(ctx, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}
