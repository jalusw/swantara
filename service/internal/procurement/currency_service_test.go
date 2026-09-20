package procurement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

func TestCurrencyConverter_Convert_SameCurrency(t *testing.T) {
	ctx := context.Background()
	rates := &CurrencyRateDAOMock{}
	conv := NewCurrencyConverter(rates)

	result, err := conv.Convert(ctx, amount.FromFloat64(100), "IDR", "IDR", nil, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Float64() != 100 {
		t.Errorf("expected 100, got %f", result.Float64())
	}
}

func TestCurrencyConverter_Convert_DirectRate(t *testing.T) {
	ctx := context.Background()
	rates := &CurrencyRateDAOMock{
		FindByPairDateFunc: func(_ context.Context, from, to string, _ *uint64, _ time.Time) (*CurrencyRate, error) {
			if from == "USD" && to == "IDR" {
				return &CurrencyRate{Rate: 15000}, nil
			}
			return nil, nil
		},
	}
	conv := NewCurrencyConverter(rates)

	result, err := conv.Convert(ctx, amount.FromFloat64(1), "USD", "IDR", nil, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Float64() != 15000 {
		t.Errorf("expected 15000, got %f", result.Float64())
	}
}

func TestCurrencyConverter_Convert_ReverseRate(t *testing.T) {
	ctx := context.Background()
	rates := &CurrencyRateDAOMock{
		FindByPairDateFunc: func(_ context.Context, from, to string, _ *uint64, _ time.Time) (*CurrencyRate, error) {
			if from == "IDR" && to == "USD" {
				return nil, nil
			}
			if from == "USD" && to == "IDR" {
				return &CurrencyRate{Rate: 15000}, nil
			}
			return nil, nil
		},
	}
	conv := NewCurrencyConverter(rates)

	result, err := conv.Convert(ctx, amount.FromFloat64(15000), "IDR", "USD", nil, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Float64() != 1 {
		t.Errorf("expected 1, got %f", result.Float64())
	}
}

func TestCurrencyConverter_Convert_NoRateFound(t *testing.T) {
	ctx := context.Background()
	rates := &CurrencyRateDAOMock{
		FindByPairDateFunc: func(_ context.Context, _, _ string, _ *uint64, _ time.Time) (*CurrencyRate, error) {
			return nil, nil
		},
	}
	conv := NewCurrencyConverter(rates)

	_, err := conv.Convert(ctx, amount.FromFloat64(100), "USD", "EUR", nil, time.Now())
	if !errors.Is(err, ErrCurrencyRateNotFound) {
		t.Errorf("expected ErrCurrencyRateNotFound, got %v", err)
	}
}

func TestCurrencyConverter_Convert_ZeroRate(t *testing.T) {
	ctx := context.Background()
	rates := &CurrencyRateDAOMock{
		FindByPairDateFunc: func(_ context.Context, from, to string, _ *uint64, _ time.Time) (*CurrencyRate, error) {
			if from == "USD" && to == "IDR" {
				return nil, nil
			}
			if from == "IDR" && to == "USD" {
				return &CurrencyRate{Rate: 0}, nil
			}
			return nil, nil
		},
	}
	conv := NewCurrencyConverter(rates)

	_, err := conv.Convert(ctx, amount.FromFloat64(15000), "IDR", "USD", nil, time.Now())
	if !errors.Is(err, ErrCurrencyRateInvalid) {
		t.Errorf("expected ErrCurrencyRateInvalid, got %v", err)
	}
}

func TestCurrencyConverter_Convert_PropagatesError(t *testing.T) {
	ctx := context.Background()
	rates := &CurrencyRateDAOMock{
		FindByPairDateFunc: func(_ context.Context, _, _ string, _ *uint64, _ time.Time) (*CurrencyRate, error) {
			return nil, errors.New("db down")
		},
	}
	conv := NewCurrencyConverter(rates)

	_, err := conv.Convert(ctx, amount.FromFloat64(100), "USD", "IDR", nil, time.Now())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCurrencyConverter_HasRate_SameCurrency(t *testing.T) {
	ctx := context.Background()
	rates := &CurrencyRateDAOMock{}
	conv := NewCurrencyConverter(rates)

	has, err := conv.HasRate(ctx, "IDR", "IDR", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has {
		t.Error("expected true for same currency")
	}
}

func TestCurrencyConverter_HasRate_DirectRate(t *testing.T) {
	ctx := context.Background()
	rates := &CurrencyRateDAOMock{
		FindByPairFunc: func(_ context.Context, from, to string, _ *uint64) (*CurrencyRate, error) {
			if from == "USD" && to == "IDR" {
				return &CurrencyRate{Rate: 15000}, nil
			}
			return nil, nil
		},
	}
	conv := NewCurrencyConverter(rates)

	has, err := conv.HasRate(ctx, "USD", "IDR", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has {
		t.Error("expected true")
	}
}

func TestCurrencyConverter_HasRate_ReverseRate(t *testing.T) {
	ctx := context.Background()
	rates := &CurrencyRateDAOMock{
		FindByPairFunc: func(_ context.Context, from, to string, _ *uint64) (*CurrencyRate, error) {
			if from == "IDR" && to == "USD" {
				return nil, nil
			}
			if from == "USD" && to == "IDR" {
				return &CurrencyRate{Rate: 15000}, nil
			}
			return nil, nil
		},
	}
	conv := NewCurrencyConverter(rates)

	has, err := conv.HasRate(ctx, "IDR", "USD", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has {
		t.Error("expected true for reverse rate")
	}
}

func TestCurrencyConverter_HasRate_NoRate(t *testing.T) {
	ctx := context.Background()
	rates := &CurrencyRateDAOMock{
		FindByPairFunc: func(_ context.Context, _, _ string, _ *uint64) (*CurrencyRate, error) {
			return nil, nil
		},
	}
	conv := NewCurrencyConverter(rates)

	has, err := conv.HasRate(ctx, "USD", "EUR", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if has {
		t.Error("expected false")
	}
}

func TestCurrencyConverter_Convert_WithOrgID(t *testing.T) {
	ctx := context.Background()
	orgID := helper.Ptr(uint64(1))
	rates := &CurrencyRateDAOMock{
		FindByPairDateFunc: func(_ context.Context, from, to string, org *uint64, _ time.Time) (*CurrencyRate, error) {
			if from == "USD" && to == "IDR" && org != nil && *org == 1 {
				return &CurrencyRate{Rate: 14500}, nil
			}
			return nil, nil
		},
	}
	conv := NewCurrencyConverter(rates)

	result, err := conv.Convert(ctx, amount.FromFloat64(1), "USD", "IDR", orgID, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Float64() != 14500 {
		t.Errorf("expected 14500, got %f", result.Float64())
	}
}
