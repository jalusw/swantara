package audit

import (
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestDiff_InsertCapturesWholeRecord(t *testing.T) {
	entity := &saleOrder{
		Base:        model.Base{ID: 7},
		Code:        "SO/00042",
		State:       "draft",
		AmountTotal: "1000.00",
	}

	diff := Diff(nil, entity)

	if diff["Code"] != "SO/00042" || diff["State"] != "draft" || diff["AmountTotal"] != "1000.00" {
		t.Errorf("expected full record in diff, got %#v", diff)
	}
	base, ok := diff["Base"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested base diff, got %#v", diff["Base"])
	}
	if base["id"] != uint64(7) {
		t.Errorf("expected id 7 in nested base, got %#v", base["id"])
	}
}

func TestDiff_DeleteCapturesWholeRecord(t *testing.T) {
	entity := &saleOrder{Base: model.Base{ID: 3}, Code: "SO/00003", State: "cancelled"}

	diff := Diff(entity, nil)

	if diff["Code"] != "SO/00003" || diff["State"] != "cancelled" {
		t.Errorf("expected full record in diff, got %#v", diff)
	}
}

func TestDiff_UpdateCapturesOnlyChangedFields(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	before := &saleOrder{Base: model.Base{ID: 9, CreatedAt: now, UpdatedAt: now}, Code: "SO/00009", State: "draft", AmountTotal: "100.00"}
	after := &saleOrder{Base: model.Base{ID: 9, CreatedAt: now, UpdatedAt: now}, Code: "SO/00009", State: "confirmed", AmountTotal: "100.00"}

	diff := Diff(before, after)

	if len(diff) != 1 {
		t.Fatalf("expected only state to change, got %#v", diff)
	}
	if diff["State"] != "confirmed" {
		t.Errorf("expected state confirmed, got %#v", diff["State"])
	}
}

func TestDiff_EqualTimestampsAreIgnored(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	before := &saleOrder{Base: model.Base{ID: 9, CreatedAt: now, UpdatedAt: now}, State: "draft"}
	after := &saleOrder{Base: model.Base{ID: 9, CreatedAt: now, UpdatedAt: now}, State: "confirmed"}

	diff := Diff(before, after)

	if diff["Base"] != nil {
		t.Errorf("expected equal timestamps to be ignored, got %#v", diff["Base"])
	}
	if diff["State"] != "confirmed" {
		t.Errorf("expected state confirmed, got %#v", diff["State"])
	}
}

func TestDiff_ChangedTimestampsAreCaptured(t *testing.T) {
	before := &saleOrder{Base: model.Base{ID: 9, UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, State: "draft"}
	after := &saleOrder{Base: model.Base{ID: 9, UpdatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}, State: "draft"}

	diff := Diff(before, after)

	base, ok := diff["Base"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested base diff, got %#v", diff["Base"])
	}
	if base["updated_at"] == nil {
		t.Error("expected changed updated_at to be captured")
	}
}

func TestDiff_NilBeforeAndNilAfter(t *testing.T) {
	if diff := Diff(nil, nil); len(diff) != 0 {
		t.Errorf("expected empty diff, got %#v", diff)
	}
}

func TestDiff_RedactsSensitiveColumns(t *testing.T) {
	before := &secretRecord{Base: model.Base{ID: 1}, Token: "old-token", State: "draft"}
	after := &secretRecord{Base: model.Base{ID: 1}, Token: "new-token", State: "confirmed"}

	diff := Diff(before, after)

	if diff["token"] != RedactedValue {
		t.Errorf("expected redacted token, got %#v", diff["token"])
	}
	if diff["state"] != "confirmed" {
		t.Errorf("expected state captured, got %#v", diff["state"])
	}
}

func TestDiff_RedactsSensitiveColumnsOnInsert(t *testing.T) {
	entity := &secretRecord{Base: model.Base{ID: 1}, Token: "issued-token", State: "pending"}

	diff := Diff(nil, entity)

	if diff["token"] != RedactedValue {
		t.Errorf("expected redacted token on insert, got %#v", diff["token"])
	}
}

func TestDiff_RedactsSensitiveColumnsOnDelete(t *testing.T) {
	entity := &secretRecord{Base: model.Base{ID: 1}, Token: "revoked-token", State: "cancelled"}

	diff := Diff(entity, nil)

	if diff["token"] != RedactedValue {
		t.Errorf("expected redacted token on delete, got %#v", diff["token"])
	}
}
