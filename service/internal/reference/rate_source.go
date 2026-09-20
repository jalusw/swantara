package reference

import (
	"context"
	"errors"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

var ErrRateNotFound = errors.New("fx rate not found")

type FxRateSource struct {
	rates dao.Base[FxRate]
}

func NewFxRateSource(rates dao.Base[FxRate]) FxRateSource {
	return FxRateSource{rates: rates}
}

func (s FxRateSource) Rate(
	ctx context.Context,
	currencyCode string,
	organizationID uint64,
	rateType amount.RateType,
	date time.Time,
) (amount.Amount, error) {
	for _, scope := range []*uint64{&organizationID, nil} {
		rate, err := s.latest(ctx, currencyCode, scope, rateType, date)
		if err != nil {
			return amount.Amount{}, err
		}
		if rate != nil {
			return amount.FromFloat64(rate.Rate), nil
		}
	}
	return amount.Amount{}, ErrRateNotFound
}

func (s FxRateSource) latest(
	ctx context.Context,
	currencyCode string,
	organizationID *uint64,
	rateType amount.RateType,
	date time.Time,
) (*FxRate, error) {
	filters := []query.Filter{
		{Field: "currency_code", Operator: query.Equal, Value: currencyCode},
		{Field: "rate_type", Operator: query.Equal, Value: string(rateType)},
		{Field: "valid_from", Operator: query.LessEqual, Value: date},
	}
	if organizationID == nil {
		filters = append(filters, query.Filter{Field: "organization_id", Operator: query.IsNull})
	} else {
		filters = append(filters, query.Filter{Field: "organization_id", Operator: query.Equal, Value: *organizationID})
	}

	page, err := s.rates.List(ctx, &query.Query{
		Filters: filters,
		Sorts:   []query.Sort{{Field: "valid_from", Direction: query.Descending}},
		Pagination: &query.Pagination{
			Page: 1,
			Size: 1,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(page.Items) == 0 {
		return nil, nil
	}
	return page.Items[0], nil
}
