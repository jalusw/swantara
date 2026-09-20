package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

func TestFxRevaluationHandler_Revalue_PostsRevaluation(t *testing.T) {
	reports := reporting.ReportDAOMock{
		OpenForeignPositionsFn: func(_ context.Context, _ uint64) ([]reporting.ForeignPositionRow, error) {
			return sampleForeignPosition(), nil
		},
	}
	rates := reporting.RateResolverMock{
		RateFn: func(_ context.Context, _ string, _ uint64, _ amount.RateType, _ time.Time) (amount.Amount, error) {
			return amount.FromInt64(15000), nil
		},
	}
	orgs := reporting.OrgReaderMock{
		FindFn: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{BaseCurrency: "IDR"}, nil
		},
	}
	svc := fxRevaluationTestSvc(reports, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, rates, orgs)
	app := handlerTestApp(t, NewFxRevaluationHandler(svc).Register, true)

	body := `{"period_id":1,"date":"2026-08-31"}`
	resp, err := doRequest(app, http.MethodPost, "/fx-revaluations/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestFxRevaluationHandler_Revalue_ReturnsEmptyWhenNoPositions(t *testing.T) {
	orgs := reporting.OrgReaderMock{
		FindFn: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{BaseCurrency: "IDR"}, nil
		},
	}
	svc := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, orgs)
	app := handlerTestApp(t, NewFxRevaluationHandler(svc).Register, true)

	body := `{"period_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/fx-revaluations/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestFxRevaluationHandler_Revalue_RejectsValidation(t *testing.T) {
	svc := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	app := handlerTestApp(t, NewFxRevaluationHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/fx-revaluations/", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestFxRevaluationHandler_Revalue_RejectsInvalidDate(t *testing.T) {
	svc := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	app := handlerTestApp(t, NewFxRevaluationHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/fx-revaluations/", `{"period_id":1,"date":"bogus"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestFxRevaluationHandler_Revalue_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	app := handlerTestApp(t, NewFxRevaluationHandler(svc).Register, false)

	resp, err := doRequest(app, http.MethodPost, "/fx-revaluations/", `{"period_id":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestFxRevaluationHandler_Revalue_ReturnsUnprocessableWhenConfigMissing(t *testing.T) {
	config := reporting.ConfigSourceMock{
		JournalIDFn: func(_ context.Context, _ uint64) (uint64, error) {
			return 0, reporting.ErrConfigMissing
		},
	}
	svc := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, config, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	app := handlerTestApp(t, NewFxRevaluationHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/fx-revaluations/", `{"period_id":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestFxRevaluationHandler_Revalue_ReturnsUnprocessableWhenRateMissing(t *testing.T) {
	reports := reporting.ReportDAOMock{
		OpenForeignPositionsFn: func(_ context.Context, _ uint64) ([]reporting.ForeignPositionRow, error) {
			return sampleForeignPosition(), nil
		},
	}
	rates := reporting.RateResolverMock{
		RateFn: func(_ context.Context, _ string, _ uint64, _ amount.RateType, _ time.Time) (amount.Amount, error) {
			return amount.Amount{}, errors.New("no rate")
		},
	}
	orgs := reporting.OrgReaderMock{
		FindFn: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{BaseCurrency: "IDR"}, nil
		},
	}
	svc := fxRevaluationTestSvc(reports, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, rates, orgs)
	app := handlerTestApp(t, NewFxRevaluationHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/fx-revaluations/", `{"period_id":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestFxRevaluationHandler_Revalue_ReturnsUnprocessableWhenBaseCurrencyMissing(t *testing.T) {
	orgs := reporting.OrgReaderMock{
		FindFn: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return nil, nil
		},
	}
	svc := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, orgs)
	app := handlerTestApp(t, NewFxRevaluationHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/fx-revaluations/", `{"period_id":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestFxRevaluationHandler_Revalue_ReturnsServerError(t *testing.T) {
	reports := reporting.ReportDAOMock{
		OpenForeignPositionsFn: func(_ context.Context, _ uint64) ([]reporting.ForeignPositionRow, error) {
			return nil, errors.New("db down")
		},
	}
	svc := fxRevaluationTestSvc(reports, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	app := handlerTestApp(t, NewFxRevaluationHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/fx-revaluations/", `{"period_id":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestFxRevaluationHandler_List_ReturnsRuns(t *testing.T) {
	runs := reporting.FxRevaluationDAOMock{
		ListByOrganizationFn: func(_ context.Context, _ uint64) ([]*reporting.FxRevaluation, error) {
			return []*reporting.FxRevaluation{{State: reporting.FxRevaluationStatePosted}}, nil
		},
	}
	svc := fxRevaluationTestSvc(reporting.ReportDAOMock{}, runs, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	app := handlerTestApp(t, NewFxRevaluationHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/fx-revaluations/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestFxRevaluationHandler_List_ReturnsServerError(t *testing.T) {
	runs := reporting.FxRevaluationDAOMock{
		ListByOrganizationFn: func(_ context.Context, _ uint64) ([]*reporting.FxRevaluation, error) {
			return nil, errors.New("db down")
		},
	}
	svc := fxRevaluationTestSvc(reporting.ReportDAOMock{}, runs, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	app := handlerTestApp(t, NewFxRevaluationHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/fx-revaluations/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestFxRevaluationHandler_List_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := fxRevaluationTestSvc(reporting.ReportDAOMock{}, reporting.FxRevaluationDAOMock{}, reporting.FxRevaluationLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{}, reporting.RateResolverMock{}, reporting.OrgReaderMock{})
	app := handlerTestApp(t, NewFxRevaluationHandler(svc).Register, false)

	resp, err := doRequest(app, http.MethodGet, "/fx-revaluations/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestFxRevaluationHandler_WriteRevaluationError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "config missing", err: reporting.ErrConfigMissing, want: http.StatusUnprocessableEntity},
		{name: "closing rate missing", err: reporting.ErrClosingRateMissing, want: http.StatusUnprocessableEntity},
		{name: "base currency missing", err: reporting.ErrBaseCurrencyMissing, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeRevaluationError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
