package reporting

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func FxRevaluationFixture(opts ...func(*FxRevaluation) *FxRevaluation) *FxRevaluation {
	now := time.Now()
	r := &FxRevaluation{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: uint64(gofakeit.Number(1, 10000)),
		PeriodID:       uint64(gofakeit.Number(1, 10000)),
		Name:           helper.Ptr(gofakeit.AppName()),
		Date:           now,
		State:          FxRevaluationStatePosted,
		TotalGainLoss:  gofakeit.Float64Range(1, 10000),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func FxRevaluationLineFixture(opts ...func(*FxRevaluationLine) *FxRevaluationLine) *FxRevaluationLine {
	now := time.Now()
	l := &FxRevaluationLine{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		RevaluationID:  uint64(gofakeit.Number(1, 10000)),
		AccountID:      uint64(gofakeit.Number(1, 10000)),
		CurrencyCode:   "USD",
		ForeignBalance: gofakeit.Float64Range(1, 10000),
		BaseBalance:    gofakeit.Float64Range(1, 10000),
		ClosingRate:    gofakeit.Float64Range(0.5, 2.0),
		GainLoss:       gofakeit.Float64Range(1, 10000),
		EntryID:        helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Reversed:       false,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func AccrualFixture(opts ...func(*Accrual) *Accrual) *Accrual {
	now := time.Now()
	reversalDate := now.AddDate(0, 1, 0)
	a := &Accrual{
		Base:            model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID:  uint64(gofakeit.Number(1, 10000)),
		PeriodID:        uint64(gofakeit.Number(1, 10000)),
		Name:            helper.Ptr(gofakeit.AppName()),
		Description:     helper.Ptr(gofakeit.Paragraph(2, 3, 8, " ")),
		ReversalDate:    &reversalDate,
		State:           AccrualStatePosted,
		EntryID:         helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		ReversalEntryID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func AccrualLineFixture(opts ...func(*AccrualLine) *AccrualLine) *AccrualLine {
	now := time.Now()
	l := &AccrualLine{
		Base:      model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		AccrualID: uint64(gofakeit.Number(1, 10000)),
		AccountID: uint64(gofakeit.Number(1, 10000)),
		Name:      helper.Ptr(gofakeit.AppName()),
		Debit:     gofakeit.Float64Range(1, 10000),
		Credit:    0,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}
