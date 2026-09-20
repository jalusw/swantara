package reference

import (
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

func TestTaxLineAmount_ReturnsZeroWhenNoAmount(t *testing.T) {
	got, err := TaxLineAmount(amount.FromFloat64(200), amount.FromInt64(1), &Tax{Type: TaxTypePercent})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(amount.Zero()) {
		t.Errorf("tax = %s, want zero", got)
	}
}

func TestTaxLineAmount_ComputesPercent(t *testing.T) {
	taxAmount := 10.0

	got, err := TaxLineAmount(amount.FromFloat64(200), amount.FromInt64(1), &Tax{Amount: &taxAmount, Type: TaxTypePercent})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(amount.FromFloat64(20)) {
		t.Errorf("tax = %s, want 20", got)
	}
}

func TestTaxLineAmount_ComputesFixed(t *testing.T) {
	taxAmount := 5.0

	got, err := TaxLineAmount(amount.FromFloat64(200), amount.FromInt64(3), &Tax{Amount: &taxAmount, Type: TaxTypeFixed})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(amount.FromFloat64(15)) {
		t.Errorf("tax = %s, want 15", got)
	}
}

func TestTaxLineAmount_ReturnsZeroForGroupType(t *testing.T) {
	taxAmount := 10.0

	got, err := TaxLineAmount(amount.FromFloat64(200), amount.FromInt64(1), &Tax{Amount: &taxAmount, Type: TaxTypeGroup})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(amount.Zero()) {
		t.Errorf("tax = %s, want zero", got)
	}
}
