package amount

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
)

func TestAmount_Round(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		dp        int32
		wantEqual string
	}{
		{name: "two decimals", input: "1.2345", dp: 2, wantEqual: "1.23"},
		{name: "half rounds up", input: "1.235", dp: 2, wantEqual: "1.24"},
		{name: "integer decimal places", input: "12.3456", dp: 0, wantEqual: "12"},
		{name: "negative value", input: "-3.14159", dp: 2, wantEqual: "-3.14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := FromString(tt.input)
			if err != nil {
				t.Fatalf("parse %q: %v", tt.input, err)
			}
			want, err := FromString(tt.wantEqual)
			if err != nil {
				t.Fatalf("parse %q: %v", tt.wantEqual, err)
			}
			if got := value.Round(tt.dp); !got.Equal(want) {
				t.Errorf("Round(%d) = %s, want %s", tt.dp, got, want)
			}
		})
	}
}

func TestAmount_ArithmeticIsExact(t *testing.T) {
	a, err := FromString("0.1")
	if err != nil {
		t.Fatal(err)
	}
	b, err := FromString("0.2")
	if err != nil {
		t.Fatal(err)
	}

	sum, err := FromString("0.3")
	if err != nil {
		t.Fatal(err)
	}
	if got := a.Add(b); !got.Equal(sum) {
		t.Errorf("0.1 + 0.2 = %s, want 0.3", got)
	}

	ten, _ := FromString("10")
	four, _ := FromString("4")
	twoPointFive, _ := FromString("2.5")
	if got, err := ten.Div(four); err != nil || !got.Equal(twoPointFive) {
		t.Errorf("10 / 4 = %v, %v; want 2.5", got, err)
	}

	onePointFive, _ := FromString("1.5")
	three, _ := FromString("3")
	if got := onePointFive.Mul(FromInt64(2)); !got.Equal(three) {
		t.Errorf("1.5 * 2 = %s, want 3", got)
	}
}

func TestAmount_SignHelpers(t *testing.T) {
	positive, _ := FromString("12.34")
	negative, _ := FromString("-12.34")
	zero := Zero()

	if !positive.IsPositive() || positive.IsNegative() || positive.IsZero() {
		t.Error("expected positive amount to only report positive")
	}
	if !negative.IsNegative() || negative.IsPositive() {
		t.Error("expected negative amount to only report negative")
	}
	if !zero.IsZero() || zero.Sign() != 0 {
		t.Error("expected zero amount to report zero")
	}
	if got := positive.Neg(); !got.Equal(negative) {
		t.Errorf("Neg() = %s, want -12.34", got)
	}
	if got := negative.Abs(); !got.Equal(positive) {
		t.Errorf("Abs() = %s, want 12.34", got)
	}
}

func TestAmount_SubNetAndBalance(t *testing.T) {
	total, _ := FromString("10.005")
	part, _ := FromString("0.005")
	if got := total.Sub(part); !got.Equal(FromInt64(10)) {
		t.Errorf("10.005 - 0.005 = %s, want 10", got)
	}

	items := []Amount{FromInt64(1), FromInt64(2), FromInt64(3)}
	if got := Net(items, 2); !got.Equal(FromInt64(6)) {
		t.Errorf("Net = %s, want 6", got)
	}

	debits, _ := FromString("100.005")
	credits, _ := FromString("100.01")
	if !IsBalanced(debits, credits, 2) {
		t.Error("expected debits and credits to balance at 2 decimals")
	}
	if IsBalanced(debits, credits, 3) {
		t.Error("expected debits and credits not to balance at 3 decimals")
	}
}

func TestAmount_DivisionByZero(t *testing.T) {
	_, err := FromInt64(1).Div(Zero())
	if helper.AssertError(t, err, true, ErrDivisionByZero) {
		return
	}
}

func TestConverter_SameCurrency(t *testing.T) {
	converter := NewConverter("IDR", staticRateSource{})

	value := FromInt64(100)
	got, err := converter.Convert(context.Background(), value, "IDR", "IDR", 1, RateSpot, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(value) {
		t.Errorf("Convert = %s, want 100", got)
	}
}

func TestConverter_BaseToForeign(t *testing.T) {
	converter := NewConverter("IDR", staticRateSource{"USD": FromInt64(15450)})

	got, err := converter.Convert(context.Background(), FromInt64(1545000), "IDR", "USD", 1, RateSpot, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(FromInt64(100)) {
		t.Errorf("Convert = %s, want 100", got)
	}
}

func TestConverter_ForeignToBase(t *testing.T) {
	converter := NewConverter("IDR", staticRateSource{"USD": FromInt64(15450)})

	got, err := converter.Convert(context.Background(), FromInt64(100), "USD", "IDR", 1, RateSpot, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(FromInt64(1545000)) {
		t.Errorf("Convert = %s, want 1545000", got)
	}
}

func TestConverter_Triangulation(t *testing.T) {
	converter := NewConverter("IDR", staticRateSource{
		"EUR": FromInt64(16250),
		"USD": FromInt64(15450),
	})

	got, err := converter.Convert(context.Background(), FromInt64(100), "EUR", "USD", 1, RateSpot, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want, _ := FromString("105.177993527508090615")
	if !got.Equal(want) {
		t.Errorf("Convert = %s, want %s", got, want)
	}
}

func TestConverter_ResolvesByDate(t *testing.T) {
	converter := NewConverter("IDR", dateRatedSource{
		"USD": {
			"2026-01-01": FromInt64(15000),
			"2026-06-01": FromInt64(16000),
		},
	})

	january, _ := time.Parse("2006-01-02", "2026-01-01")
	june, _ := time.Parse("2006-01-02", "2026-06-01")

	janGot, err := converter.Convert(context.Background(), FromInt64(1), "USD", "IDR", 1, RateSpot, january)
	if err != nil {
		t.Fatalf("january: %v", err)
	}
	junGot, err := converter.Convert(context.Background(), FromInt64(1), "USD", "IDR", 1, RateSpot, june)
	if err != nil {
		t.Fatalf("june: %v", err)
	}

	if !janGot.Equal(FromInt64(15000)) || !junGot.Equal(FromInt64(16000)) {
		t.Errorf("expected 15000 in january and 16000 in june, got %s / %s", janGot, junGot)
	}
}

func TestConverter_RejectsZeroRate(t *testing.T) {
	converter := NewConverter("IDR", staticRateSource{"USD": Zero()})

	_, fromZero := converter.Convert(context.Background(), FromInt64(1), "USD", "IDR", 1, RateSpot, time.Now())
	if helper.AssertError(t, fromZero, true, ErrInvalidRate) {
		return
	}

	_, toZero := converter.Convert(context.Background(), FromInt64(1), "IDR", "USD", 1, RateSpot, time.Now())
	if helper.AssertError(t, toZero, true, ErrInvalidRate) {
		return
	}
}

func TestConverter_SourceErrorPropagates(t *testing.T) {
	sentinel := errors.New("rate source unavailable")
	converter := NewConverter("IDR", failingRateSource{err: sentinel})

	_, err := converter.Convert(context.Background(), FromInt64(1), "USD", "IDR", 1, RateSpot, time.Now())
	if helper.AssertError(t, err, true, sentinel) {
		return
	}
}

func TestAmount_FromFloat64AndFloat64(t *testing.T) {
	value := FromFloat64(12.34)
	if got := value.Float64(); got != 12.34 {
		t.Errorf("Float64 = %v, want 12.34", got)
	}
}

func TestAmount_FromStringRejectsInvalid(t *testing.T) {
	_, err := FromString("not-a-number")
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestAmount_ComparisonOperators(t *testing.T) {
	small := FromInt64(1)
	big := FromInt64(2)

	if !big.GreaterThan(small) {
		t.Error("expected 2 to be greater than 1")
	}
	if big.GreaterThan(big) {
		t.Error("expected 2 not to be greater than itself")
	}
	if !small.LessThan(big) {
		t.Error("expected 1 to be less than 2")
	}
	if small.LessThan(small) {
		t.Error("expected 1 not to be less than itself")
	}
}

func TestAmount_String(t *testing.T) {
	value, err := FromString("123.4500")
	if err != nil {
		t.Fatal(err)
	}
	if got := value.String(); got != "123.45" {
		t.Errorf("String = %q, want 123.45", got)
	}
}

func TestAmount_JSONRoundTrip(t *testing.T) {
	original, err := FromString("99.99")
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(encoded) != `99.99` {
		t.Errorf("marshal = %s, want 99.99", encoded)
	}

	var decoded Amount
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !decoded.Equal(original) {
		t.Errorf("round trip = %s, want 99.99", decoded)
	}
}

func TestAmount_UnmarshalRejectsInvalid(t *testing.T) {
	var decoded Amount
	err := json.Unmarshal([]byte(`"oops"`), &decoded)
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestAmountValueScanRoundTrip(t *testing.T) {
	original := FromFloat64(1234.5678)

	stored, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}
	stringValue, ok := stored.(string)
	if !ok {
		t.Fatalf("Value() = %T, want string", stored)
	}

	var restored Amount
	if err := restored.Scan(stringValue); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if !restored.Equal(original) {
		t.Errorf("round trip = %s, want %s", restored, original)
	}
}

func TestAmountScanHandlesDriversAndNull(t *testing.T) {
	var fromBytes Amount
	if err := fromBytes.Scan([]byte("99.9999")); err != nil {
		t.Fatalf("Scan([]byte) error = %v", err)
	}
	if !fromBytes.Equal(FromFloat64(99.9999)) {
		t.Errorf("from bytes = %s", fromBytes)
	}

	var fromFloat Amount
	if err := fromFloat.Scan(float64(7.25)); err != nil {
		t.Fatalf("Scan(float64) error = %v", err)
	}
	if !fromFloat.Equal(FromFloat64(7.25)) {
		t.Errorf("from float = %s", fromFloat)
	}

	var fromNull Amount
	if err := fromNull.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if !fromNull.IsZero() {
		t.Errorf("from null = %s, want zero", fromNull)
	}

	var invalid Amount
	if err := invalid.Scan("not-a-number"); err == nil {
		t.Error("expected error for non-numeric string")
	}
}

func TestAmountJSONNumberShape(t *testing.T) {
	data, err := json.Marshal(FromFloat64(1500.5))
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}
	if string(data) != "1500.5" {
		t.Errorf("marshaled = %s, want bare number 1500.5", data)
	}
}
