package procurement

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type ScorecardWeights struct {
	Quality  float64
	Delivery float64
	Price    float64
}

var DefaultWeights = ScorecardWeights{
	Quality:  0.4,
	Delivery: 0.3,
	Price:    0.3,
}

type ScorecardUpdater interface {
	UpdateScorecard(ctx context.Context, organizationID *uint64, supplierID uint64, orderDoneAt time.Time) error
}

type SupplierScorecardService struct {
	scorecards SupplierScorecardDAO
	weights    ScorecardWeights
}

func NewSupplierScorecardService(scorecards SupplierScorecardDAO) SupplierScorecardService {
	return SupplierScorecardService{
		scorecards: scorecards,
		weights:    DefaultWeights,
	}
}

func (s SupplierScorecardService) Find(ctx context.Context, id uint64) (*SupplierScorecard, error) {
	sc, err := s.scorecards.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		return nil, ErrScorecardNotFound
	}
	return sc, nil
}

func (s SupplierScorecardService) List(ctx context.Context, q *query.Query) (*query.Page[SupplierScorecard], error) {
	return s.scorecards.List(ctx, q)
}

func (s SupplierScorecardService) ListBySupplier(ctx context.Context, supplierID uint64) ([]*SupplierScorecard, error) {
	page, err := s.scorecards.List(ctx, &query.Query{
		Filters: []query.Filter{{Field: "supplier_id", Operator: query.Equal, Value: supplierID}},
		Sorts:   []query.Sort{{Field: "period_start", Direction: query.Descending}},
	})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s SupplierScorecardService) ComputeScores(totalOrders, onTimeDeliveries, qualityFailures int, priceVariance float64) (quality, delivery, price, overall float64) {
	if totalOrders == 0 {
		return 0, 0, 0, 0
	}
	quality = 100 - (float64(qualityFailures) / float64(totalOrders) * 100)
	delivery = float64(onTimeDeliveries) / float64(totalOrders) * 100
	if priceVariance <= 0 {
		price = 100
	} else if priceVariance >= 20 {
		price = 0
	} else {
		price = 100 - (priceVariance / 20 * 100)
	}
	overall = quality*s.weights.Quality + delivery*s.weights.Delivery + price*s.weights.Price
	return
}

func (s SupplierScorecardService) UpdateScorecard(ctx context.Context, organizationID *uint64, supplierID uint64, orderDoneAt time.Time) error {
	periodStart := time.Date(orderDoneAt.Year(), orderDoneAt.Month(), 1, 0, 0, 0, 0, orderDoneAt.Location())
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	existing, err := s.scorecards.FindByVendorAndPeriod(ctx, supplierID, &periodStart, &periodEnd)
	if err != nil {
		return err
	}

	if existing == nil {
		existing = &SupplierScorecard{
			OrganizationID: organizationID,
			SupplierID:     supplierID,
			PeriodStart:    &periodStart,
			PeriodEnd:      &periodEnd,
		}
		existing.TotalOrders++
		existing.OnTimeDeliveries++
		quality, delivery, price, overall := s.ComputeScores(
			existing.TotalOrders, existing.OnTimeDeliveries, existing.QualityFailures, 0,
		)
		existing.QualityScore = quality
		existing.DeliveryScore = delivery
		existing.PriceScore = price
		existing.OverallScore = overall
		_, err = s.scorecards.Create(ctx, existing)
		return err
	}

	existing.TotalOrders++
	existing.OnTimeDeliveries++
	quality, delivery, price, overall := s.ComputeScores(
		existing.TotalOrders, existing.OnTimeDeliveries, existing.QualityFailures, 0,
	)
	existing.QualityScore = quality
	existing.DeliveryScore = delivery
	existing.PriceScore = price
	existing.OverallScore = overall
	_, err = s.scorecards.Update(ctx, existing)
	return err
}

func (s SupplierScorecardService) RecordQualityFailure(ctx context.Context, organizationID *uint64, supplierID uint64, orderDoneAt time.Time) error {
	periodStart := time.Date(orderDoneAt.Year(), orderDoneAt.Month(), 1, 0, 0, 0, 0, orderDoneAt.Location())
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	existing, err := s.scorecards.FindByVendorAndPeriod(ctx, supplierID, &periodStart, &periodEnd)
	if err != nil {
		return err
	}
	if existing == nil {
		existing = &SupplierScorecard{
			OrganizationID:  organizationID,
			SupplierID:      supplierID,
			PeriodStart:     &periodStart,
			PeriodEnd:       &periodEnd,
			TotalOrders:     1,
			QualityFailures: 1,
		}
	} else {
		existing.TotalOrders++
		existing.QualityFailures++
	}
	quality, delivery, price, overall := s.ComputeScores(
		existing.TotalOrders, existing.OnTimeDeliveries, existing.QualityFailures, 0,
	)
	existing.QualityScore = quality
	existing.DeliveryScore = delivery
	existing.PriceScore = price
	existing.OverallScore = overall
	if existing.ID == 0 {
		_, err = s.scorecards.Create(ctx, existing)
	} else {
		_, err = s.scorecards.Update(ctx, existing)
	}
	return err
}
