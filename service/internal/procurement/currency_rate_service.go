package procurement

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type CurrencyRateService struct {
	rates CurrencyRateDAO
}

func NewCurrencyRateService(rates CurrencyRateDAO) CurrencyRateService {
	return CurrencyRateService{rates: rates}
}

func (s CurrencyRateService) List(ctx context.Context, q *query.Query) (*query.Page[CurrencyRate], error) {
	return s.rates.List(ctx, q)
}

func (s CurrencyRateService) Find(ctx context.Context, id uint64) (*CurrencyRate, error) {
	return s.rates.Find(ctx, id)
}

func (s CurrencyRateService) Create(ctx context.Context, rate *CurrencyRate) (*CurrencyRate, error) {
	return s.rates.Create(ctx, rate)
}

func (s CurrencyRateService) Update(ctx context.Context, rate *CurrencyRate) (*CurrencyRate, error) {
	return s.rates.Update(ctx, rate)
}

func (s CurrencyRateService) Delete(ctx context.Context, id uint64) error {
	return s.rates.Delete(ctx, id)
}

func (s CurrencyRateService) FindByPair(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64) (*CurrencyRate, error) {
	return s.rates.FindByPair(ctx, fromCurrency, toCurrency, orgID)
}

func (s CurrencyRateService) FindByPairDate(ctx context.Context, fromCurrency, toCurrency string, orgID *uint64, rateDate time.Time) (*CurrencyRate, error) {
	return s.rates.FindByPairDate(ctx, fromCurrency, toCurrency, orgID, rateDate)
}
