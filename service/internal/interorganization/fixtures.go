package interorganization

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func DropshipLinkFixture(opts ...func(*DropshipLink) *DropshipLink) *DropshipLink {
	now := time.Now()
	dl := &DropshipLink{
		Base:                model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		SaleOrderLineID:     uint64(gofakeit.Number(1, 10000)),
		PurchaseOrderLineID: uint64(gofakeit.Number(1, 10000)),
		StockMovementID:     helper.Ptr(uint64(gofakeit.Number(1, 10000))),
	}
	for _, opt := range opts {
		opt(dl)
	}
	return dl
}

func InterorganizationRuleFixture(opts ...func(*InterorganizationRule) *InterorganizationRule) *InterorganizationRule {
	now := time.Now()
	r := &InterorganizationRule{
		Base:               model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		FromOrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		ToOrganizationID:   helper.Ptr(uint64(gofakeit.Number(1, 100))),
		AutoMirror:         gofakeit.Bool(),
		SupplierContactID:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		CustomerContactID:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func InterorganizationTransactionFixture(opts ...func(*InterorganizationTransaction) *InterorganizationTransaction) *InterorganizationTransaction {
	now := time.Now()
	tx := &InterorganizationTransaction{
		Base:                 model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		SourceOrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		SourceType:           "journal_entry",
		SourceID:             helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		MirrorOrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		MirrorType:           "journal_entry",
		MirrorID:             helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Amount:               gofakeit.Float64Range(1, 10000),
		State:                TransactionStateDone,
	}
	for _, opt := range opts {
		opt(tx)
	}
	return tx
}

func ConsolidationRunFixture(opts ...func(*ConsolidationRun) *ConsolidationRun) *ConsolidationRun {
	now := time.Now()
	cr := &ConsolidationRun{
		Base:                model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		GroupOrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		PeriodID:            helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		ReportingCurrency:   "USD",
		State:               RunStateDraft,
	}
	for _, opt := range opts {
		opt(cr)
	}
	return cr
}

func ConsolidationEliminationFixture(opts ...func(*ConsolidationElimination) *ConsolidationElimination) *ConsolidationElimination {
	now := time.Now()
	ce := &ConsolidationElimination{
		Base:                       model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ConsolidationRunID:         uint64(gofakeit.Number(1, 10000)),
		AccountID:                  uint64(gofakeit.Number(1, 10000)),
		CounterpartyOrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		Amount:                     gofakeit.Float64Range(1, 10000),
		Description:                gofakeit.Sentence(),
	}
	for _, opt := range opts {
		opt(ce)
	}
	return ce
}
