package systemconfig

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type stubLookup struct {
	err   error
	items []*reference.SystemConfig
}

func (m stubLookup) List(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
	if m.err != nil {
		return nil, m.err
	}
	return &query.Page[reference.SystemConfig]{Items: m.items, Count: int64(len(m.items))}, nil
}

func TestRead(t *testing.T) {
	ctx := context.Background()
	orgID := uint64(10)

	t.Run("reads org value", func(t *testing.T) {
		lookup := stubLookup{items: []*reference.SystemConfig{
			{OrganizationID: &orgID, Key: "org.modules", Value: json.RawMessage(`{"sales":true}`)},
		}}
		var got map[string]bool
		if err := Read(ctx, lookup, 10, "org.modules", &got); helper.AssertError(t, err, false, nil) {
			return
		}
		if !got["sales"] {
			t.Errorf("value = %v", got)
		}
	})

	t.Run("reads global fallback", func(t *testing.T) {
		lookup := stubLookup{items: []*reference.SystemConfig{
			{Key: "platform.modules", Value: json.RawMessage(`{"pos":false}`)},
		}}
		var got map[string]bool
		if err := Read(ctx, lookup, 10, "platform.modules", &got); helper.AssertError(t, err, false, nil) {
			return
		}
		if got["pos"] {
			t.Errorf("value = %v", got)
		}
	})

	t.Run("skips foreign org", func(t *testing.T) {
		other := uint64(99)
		lookup := stubLookup{items: []*reference.SystemConfig{
			{OrganizationID: &other, Key: "org.modules", Value: json.RawMessage(`{"sales":true}`)},
		}}
		var got map[string]bool
		if err := Read(ctx, lookup, 10, "org.modules", &got); helper.AssertError(t, err, false, nil) {
			return
		}
		if len(got) != 0 {
			t.Errorf("value = %v, want empty", got)
		}
	})

	t.Run("empty value is no-op", func(t *testing.T) {
		lookup := stubLookup{items: []*reference.SystemConfig{
			{OrganizationID: &orgID, Key: "org.modules"},
		}}
		var got map[string]bool
		if err := Read(ctx, lookup, 10, "org.modules", &got); helper.AssertError(t, err, false, nil) {
			return
		}
	})

	t.Run("missing key is no-op", func(t *testing.T) {
		lookup := stubLookup{}
		var got map[string]bool
		if err := Read(ctx, lookup, 10, "missing", &got); helper.AssertError(t, err, false, nil) {
			return
		}
	})

	t.Run("propagates list error", func(t *testing.T) {
		lookup := stubLookup{err: errors.New("db down")}
		var got map[string]bool
		err := Read(ctx, lookup, 10, "org.modules", &got)
		helper.AssertError(t, err, true, lookup.err)
	})

	t.Run("propagates decode error", func(t *testing.T) {
		lookup := stubLookup{items: []*reference.SystemConfig{
			{OrganizationID: &orgID, Key: "org.modules", Value: json.RawMessage(`{`)},
		}}
		var got map[string]bool
		if err := Read(ctx, lookup, 10, "org.modules", &got); err == nil {
			t.Error("expected decode error, got nil")
		}
	})
}
