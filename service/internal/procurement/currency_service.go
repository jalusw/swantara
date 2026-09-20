package procurement

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type currencyConverter struct {
	rates CurrencyRateDAO
}

func NewCurrencyConverter(rates CurrencyRateDAO) CurrencyConverter {
	return currencyConverter{rates: rates}
}

func (c currencyConverter) Convert(ctx context.Context, amt amount.Amount, fromCurrency, toCurrency string, orgID *uint64, rateDate time.Time) (amount.Amount, error) {
	if fromCurrency == toCurrency {
		return amt, nil
	}
	rate, err := c.rates.FindByPairDate(ctx, fromCurrency, toCurrency, orgID, rateDate)
	if err != nil {
		return amount.Zero(), err
	}
	if rate == nil {
		reverse, err := c.rates.FindByPairDate(ctx, toCurrency, fromCurrency, orgID, rateDate)
		if err != nil {
			return amount.Zero(), err
		}
		if reverse == nil {
			return amount.Zero(), ErrCurrencyRateNotFound
		}
		if reverse.Rate == 0 {
			return amount.Zero(), ErrCurrencyRateInvalid
		}
		converted, err := amt.Div(amount.FromFloat64(reverse.Rate))
		if err != nil {
			return amount.Zero(), err
		}
		return converted.Round(4), nil
	}
	if rate.Rate == 0 {
		return amount.Zero(), ErrCurrencyRateInvalid
	}
	return amt.Mul(amount.FromFloat64(rate.Rate)).Round(4), nil
}

func (c currencyConverter) HasRate(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64) (bool, error) {
	if fromCurrency == toCurrency {
		return true, nil
	}
	rate, err := c.rates.FindByPair(ctx, fromCurrency, toCurrency, orgID)
	if err != nil {
		return false, err
	}
	if rate != nil {
		return true, nil
	}
	reverse, err := c.rates.FindByPair(ctx, toCurrency, fromCurrency, orgID)
	if err != nil {
		return false, err
	}
	return reverse != nil, nil
}
