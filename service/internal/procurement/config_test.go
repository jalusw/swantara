package procurement

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestSystemConfigSource_ApprovalThreshold_ReturnsValue(t *testing.T) {
	ctx := context.Background()
	source := NewSystemConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Key: configApprovalThreshold, Value: []byte(`1000`)},
			}}, nil
		},
	})

	value, err := source.ApprovalThreshold(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 1000 {
		t.Errorf("value = %v, want 1000", value)
	}
}

func TestSystemConfigSource_ApprovalThreshold_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	source := NewSystemConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return nil, errors.New("db down")
		},
	})

	_, err := source.ApprovalThreshold(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestSystemConfigSource_ApproverIDs_ReturnsValue(t *testing.T) {
	ctx := context.Background()
	source := NewSystemConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Base: model.Base{ID: 2}, OrganizationID: helper.Ptr(uint64(1)), Key: configApproverIDs, Value: []byte(`[7,8]`)},
			}}, nil
		},
	})

	value, err := source.ApproverIDs(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(value) != 2 || value[0] != 7 || value[1] != 8 {
		t.Errorf("value = %v, want [7 8]", value)
	}
}

func TestSystemConfigSource_ApproverIDs_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	source := NewSystemConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return nil, errors.New("db down")
		},
	})

	_, err := source.ApproverIDs(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestSystemConfigSource_Read_SkipsEmptyValue(t *testing.T) {
	ctx := context.Background()
	source := NewSystemConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Key: configApprovalThreshold},
			}}, nil
		},
	})

	value, err := source.ApprovalThreshold(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 0 {
		t.Errorf("value = %v, want 0", value)
	}
}

func TestSystemConfigSource_Read_IgnoresNonMatchingConfig(t *testing.T) {
	ctx := context.Background()
	source := NewSystemConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(2)), Key: configApprovalThreshold, Value: []byte(`1000`)},
			}}, nil
		},
	})

	value, err := source.ApprovalThreshold(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 0 {
		t.Errorf("value = %v, want 0", value)
	}
}

func TestSystemConfigSource_Read_IgnoresWrongKey(t *testing.T) {
	ctx := context.Background()
	source := NewSystemConfigSource(ConfigLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Key: configApproverIDs, Value: []byte(`[7]`)},
			}}, nil
		},
	})

	value, err := source.ApprovalThreshold(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 0 {
		t.Errorf("value = %v, want 0", value)
	}
}
