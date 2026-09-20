package audit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestService_AppendFillsActorAndTimestamp(t *testing.T) {
	ctx := model.ContextWithActor(context.Background(), 42)
	var captured *Log

	svc := NewAuditLogService(LogDAOMock{
		CreateFunc: func(_ context.Context, entry *Log) (*Log, error) {
			captured = entry
			return entry, nil
		},
	})

	err := svc.Append(ctx, &Log{EntityTable: "sale_orders", RecordID: 7, Action: ActionUpdate, Diff: json.RawMessage(`{"state":"draft"}`)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured.ChangedBy != 42 {
		t.Errorf("expected changed_by 42, got %d", captured.ChangedBy)
	}
	if captured.ChangedAt.IsZero() {
		t.Error("expected changed_at to be filled")
	}
}

func TestService_AppendPreservesExplicitValues(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var captured *Log

	svc := NewAuditLogService(LogDAOMock{
		CreateFunc: func(_ context.Context, entry *Log) (*Log, error) {
			captured = entry
			return entry, nil
		},
	})

	err := svc.Append(context.Background(), &Log{
		EntityTable: "audit_logs_override",
		RecordID:    1,
		Action:      ActionInsert,
		ChangedBy:   9,
		ChangedAt:   at,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured.ChangedBy != 9 || !captured.ChangedAt.Equal(at) {
		t.Errorf("expected explicit values preserved, got %#v", captured)
	}
}

func TestService_AppendPropagatesCreateError(t *testing.T) {
	sentinel := errors.New("db error")
	svc := NewAuditLogService(LogDAOMock{
		CreateFunc: func(_ context.Context, _ *Log) (*Log, error) {
			return nil, sentinel
		},
	})

	err := svc.Append(context.Background(), &Log{EntityTable: "sale_orders", RecordID: 1, Action: ActionInsert})
	if helper.AssertError(t, err, true, sentinel) {
		return
	}
}

func TestRecord_MarshalsDiff(t *testing.T) {
	var captured *Log

	svc := NewAuditLogService(LogDAOMock{
		CreateFunc: func(_ context.Context, entry *Log) (*Log, error) {
			captured = entry
			return entry, nil
		},
	})

	err := Record(context.Background(), svc, "sale_orders", 21, ActionInsert, map[string]any{"code": "SO/0042"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured.EntityTable != "sale_orders" || captured.RecordID != 21 || captured.Action != ActionInsert {
		t.Errorf("unexpected entry: %#v", captured)
	}
	var diff map[string]any
	if err := json.Unmarshal(captured.Diff, &diff); err != nil {
		t.Fatalf("invalid diff json: %v", err)
	}
	if diff["code"] != "SO/0042" {
		t.Errorf("unexpected diff: %#v", diff)
	}
}
