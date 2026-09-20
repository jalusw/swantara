package amount

import (
	"context"
	"errors"
	"time"
)

type RateType string

const (
	RateSpot    RateType = "spot"
	RateAverage RateType = "avg"
	RateClosing RateType = "closing"
)

var ErrInvalidRate = errors.New("fx rate must be greater than zero")

// RateSource resolves the applicable fx rate for a currency, organization, type and
// effective date. Implementations are expected to resolve by effective date,
// falling back to the latest rate valid on or before the given date.
type RateSource interface {
	Rate(ctx context.Context, currencyCode string, organizationID uint64, rateType RateType, date time.Time) (Amount, error)
}

// Converter translates amounts between currencies using a RateSource. Rates are
// expressed as one unit of foreign currency in terms of the base currency.
type Converter struct {
	baseCurrency string
	source       RateSource
}

func NewConverter(baseCurrency string, source RateSource) Converter {
	return Converter{baseCurrency: baseCurrency, source: source}
}

func (c Converter) Convert(
	ctx context.Context,
	value Amount,
	from, to string,
	organizationID uint64,
	rateType RateType,
	date time.Time,
) (Amount, error) {
	if from == to {
		return value, nil
	}

	if from == c.baseCurrency {
		rate, err := c.rate(ctx, to, organizationID, rateType, date)
		if err != nil {
			return Amount{}, err
		}
		return value.Div(rate)
	}

	fromRate, err := c.rate(ctx, from, organizationID, rateType, date)
	if err != nil {
		return Amount{}, err
	}
	baseValue := value.Mul(fromRate)

	if to == c.baseCurrency {
		return baseValue, nil
	}

	toRate, err := c.rate(ctx, to, organizationID, rateType, date)
	if err != nil {
		return Amount{}, err
	}
	return baseValue.Div(toRate)
}

func (c Converter) rate(
	ctx context.Context,
	currencyCode string,
	organizationID uint64,
	rateType RateType,
	date time.Time,
) (Amount, error) {
	rate, err := c.source.Rate(ctx, currencyCode, organizationID, rateType, date)
	if err != nil {
		return Amount{}, err
	}
	if rate.IsZero() || rate.IsNegative() {
		return Amount{}, ErrInvalidRate
	}
	return rate, nil
}
