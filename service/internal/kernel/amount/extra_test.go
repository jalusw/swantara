package amount

import (
	"context"
	"testing"
	"time"
)

func TestAmount_Scan_Branches(t *testing.T) {
	t.Run("nil zeroes", func(t *testing.T) {
		var a Amount
		if err := a.Scan(nil); err != nil {
			t.Fatalf("Scan = %v", err)
		}
		if !a.IsZero() {
			t.Errorf("amount = %v, want 0", a)
		}
	})

	t.Run("int64 scans", func(t *testing.T) {
		var a Amount
		if err := a.Scan(int64(42)); err != nil {
			t.Fatalf("Scan = %v", err)
		}
		if a.Int64() != 42 {
			t.Errorf("amount = %v, want 42", a)
		}
	})

	t.Run("invalid string errors", func(t *testing.T) {
		var a Amount
		if err := a.Scan("not-a-number"); err == nil {
			t.Error("expected error, got nil")
		}
		if err := a.Scan([]byte("also-bad")); err == nil {
			t.Error("expected error, got nil")
		}
		if err := a.Scan(true); err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestAmount_UnmarshalJSON_Branches(t *testing.T) {
	t.Run("null zeroes", func(t *testing.T) {
		var a Amount
		if err := a.UnmarshalJSON([]byte(`null`)); err != nil {
			t.Fatalf("UnmarshalJSON = %v", err)
		}
		if !a.IsZero() {
			t.Errorf("amount = %v, want 0", a)
		}
	})

	t.Run("invalid errors", func(t *testing.T) {
		var a Amount
		if err := a.UnmarshalJSON([]byte(`{`)); err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestRateSources_Missing(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	if _, err := (staticRateSource{}).Rate(ctx, "XXX", 1, RateSpot, now); err != nil {
		t.Errorf("static missing = %v", err)
	}
	if _, err := (dateRatedSource{}).Rate(ctx, "XXX", 1, RateSpot, now); err == nil {
		t.Error("expected error for missing dated rate, got nil")
	}
}
