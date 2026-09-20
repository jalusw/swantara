package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/asset"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

func periodCloseTestSvcDefault() reporting.PeriodCloseService {
	fx := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	accruals := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	return periodCloseTestSvc(
		&reporting.PeriodCloserMock{},
		reporting.TaxPeriodFinderMock{},
		reporting.ConfigSourceMock{},
		reporting.AssetListReaderMock{},
		&reporting.DepreciationPosterMock{},
		fx,
		accruals,
		reporting.DeferralRecognizerMock{},
		reporting.KpiSummarizerMock{},
	)
}

func TestPeriodCloseHandler_Close_ClosesPeriod(t *testing.T) {
	svc := periodCloseTestSvcDefault()
	app := handlerTestApp(t, NewPeriodCloseHandler(svc).Register, true)

	body := `{"period_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/period-close/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPeriodCloseHandler_Close_RejectsValidation(t *testing.T) {
	svc := periodCloseTestSvcDefault()
	app := handlerTestApp(t, NewPeriodCloseHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/period-close/", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPeriodCloseHandler_Close_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := periodCloseTestSvcDefault()
	app := handlerTestApp(t, NewPeriodCloseHandler(svc).Register, false)

	resp, err := doRequest(app, http.MethodPost, "/period-close/", `{"period_id":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPeriodCloseHandler_Close_ReturnsNotFound(t *testing.T) {
	finder := reporting.TaxPeriodFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, nil
		},
	}
	fx := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	accruals := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	svc := periodCloseTestSvc(&reporting.PeriodCloserMock{}, finder, reporting.ConfigSourceMock{}, reporting.AssetListReaderMock{}, &reporting.DepreciationPosterMock{}, fx, accruals, reporting.DeferralRecognizerMock{}, reporting.KpiSummarizerMock{})
	app := handlerTestApp(t, NewPeriodCloseHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/period-close/", `{"period_id":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPeriodCloseHandler_Close_ReturnsUnprocessableWhenLocked(t *testing.T) {
	finder := reporting.TaxPeriodFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{State: accounting.TaxPeriodStateLocked}, nil
		},
	}
	fx := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	accruals := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	svc := periodCloseTestSvc(&reporting.PeriodCloserMock{}, finder, reporting.ConfigSourceMock{}, reporting.AssetListReaderMock{}, &reporting.DepreciationPosterMock{}, fx, accruals, reporting.DeferralRecognizerMock{}, reporting.KpiSummarizerMock{})
	app := handlerTestApp(t, NewPeriodCloseHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/period-close/", `{"period_id":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPeriodCloseHandler_Close_ReturnsUnprocessableWhenConfigMissing(t *testing.T) {
	config := reporting.ConfigSourceMock{
		JournalIDFn: func(_ context.Context, _ uint64) (uint64, error) {
			return 0, reporting.ErrConfigMissing
		},
	}
	fx := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, config, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	accruals := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, config)
	svc := periodCloseTestSvc(&reporting.PeriodCloserMock{}, reporting.TaxPeriodFinderMock{}, config, reporting.AssetListReaderMock{}, &reporting.DepreciationPosterMock{}, fx, accruals, reporting.DeferralRecognizerMock{}, reporting.KpiSummarizerMock{})
	app := handlerTestApp(t, NewPeriodCloseHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/period-close/", `{"period_id":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPeriodCloseHandler_Close_ReturnsServerError(t *testing.T) {
	finder := reporting.TaxPeriodFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, errors.New("db down")
		},
	}
	fx := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	accruals := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	svc := periodCloseTestSvc(&reporting.PeriodCloserMock{}, finder, reporting.ConfigSourceMock{}, reporting.AssetListReaderMock{}, &reporting.DepreciationPosterMock{}, fx, accruals, reporting.DeferralRecognizerMock{}, reporting.KpiSummarizerMock{})
	app := handlerTestApp(t, NewPeriodCloseHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/period-close/", `{"period_id":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPeriodCloseHandler_Close_ReturnsServerErrorOnAssets(t *testing.T) {
	assets := reporting.AssetListReaderMock{
		ListFn: func(_ context.Context, _ *query.Query) (*query.Page[asset.FixedAsset], error) {
			return nil, errors.New("db down")
		},
	}
	fx := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	accruals := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	svc := periodCloseTestSvc(&reporting.PeriodCloserMock{}, reporting.TaxPeriodFinderMock{}, reporting.ConfigSourceMock{}, assets, &reporting.DepreciationPosterMock{}, fx, accruals, reporting.DeferralRecognizerMock{}, reporting.KpiSummarizerMock{})
	app := handlerTestApp(t, NewPeriodCloseHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/period-close/", `{"period_id":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPeriodCloseHandler_WriteCloseError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "period not found", err: reporting.ErrPeriodNotFound, want: http.StatusNotFound},
		{name: "period locked", err: reporting.ErrPeriodLocked, want: http.StatusUnprocessableEntity},
		{name: "config missing", err: reporting.ErrConfigMissing, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeCloseError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
