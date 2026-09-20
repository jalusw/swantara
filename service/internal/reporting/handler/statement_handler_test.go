package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

func TestStatementHandler_ProfitAndLoss_ReturnsStatement(t *testing.T) {
	reports := reporting.ReportDAOMock{
		StatementBalancesFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]reporting.AccountBalanceRow, error) {
			return []reporting.AccountBalanceRow{
				{AccountID: 1, Code: "4000", AccountType: "income", Balance: -500},
				{AccountID: 2, Code: "5000", AccountType: "cogs", Balance: 200},
				{AccountID: 3, Code: "6000", AccountType: "expense", Balance: 100},
			}, nil
		},
	}
	svc := statementTestSvc(reports, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/profit-and-loss?period_id=1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestStatementHandler_ProfitAndLoss_RejectsInvalidPeriod(t *testing.T) {
	svc := statementTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/profit-and-loss?period_id=abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestStatementHandler_ProfitAndLoss_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := statementTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, false)

	resp, err := doRequest(app, http.MethodGet, "/reports/profit-and-loss?period_id=1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestStatementHandler_ProfitAndLoss_ReturnsNotFound(t *testing.T) {
	periods := reporting.TaxPeriodFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, nil
		},
	}
	svc := statementTestSvc(reporting.ReportDAOMock{}, periods, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/profit-and-loss?period_id=1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestStatementHandler_ProfitAndLoss_ReturnsServerError(t *testing.T) {
	reports := reporting.ReportDAOMock{
		StatementBalancesFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]reporting.AccountBalanceRow, error) {
			return nil, errors.New("db down")
		},
	}
	svc := statementTestSvc(reports, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/profit-and-loss?period_id=1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestStatementHandler_BalanceSheet_ReturnsSheet(t *testing.T) {
	reports := reporting.ReportDAOMock{
		StatementBalancesFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]reporting.AccountBalanceRow, error) {
			return []reporting.AccountBalanceRow{
				{AccountID: 1, Code: "1000", AccountType: "cash", Balance: 500},
				{AccountID: 2, Code: "2000", AccountType: "payable", Balance: 200},
				{AccountID: 3, Code: "3000", AccountType: "equity", Balance: 300},
			}, nil
		},
	}
	svc := statementTestSvc(reports, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/balance-sheet?as_of=2026-08-01", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestStatementHandler_BalanceSheet_RejectsInvalidAsOf(t *testing.T) {
	svc := statementTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/balance-sheet?as_of=bogus", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestStatementHandler_BalanceSheet_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := statementTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, false)

	resp, err := doRequest(app, http.MethodGet, "/reports/balance-sheet", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestStatementHandler_BalanceSheet_ReturnsServerError(t *testing.T) {
	reports := reporting.ReportDAOMock{
		StatementBalancesFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]reporting.AccountBalanceRow, error) {
			return nil, errors.New("db down")
		},
	}
	svc := statementTestSvc(reports, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/balance-sheet", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestStatementHandler_CashFlow_ReturnsFlow(t *testing.T) {
	reports := reporting.ReportDAOMock{
		CashFlowFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]reporting.CashFlowSectionRow, error) {
			return []reporting.CashFlowSectionRow{{Section: "operating", Amount: 100}}, nil
		},
		StatementBalancesFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]reporting.AccountBalanceRow, error) {
			return []reporting.AccountBalanceRow{{AccountID: 1, AccountType: "cash", Balance: 50}}, nil
		},
	}
	svc := statementTestSvc(reports, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/cash-flow?start=2026-08-01&end=2026-08-31", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestStatementHandler_CashFlow_RejectsMissingDates(t *testing.T) {
	svc := statementTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/cash-flow", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestStatementHandler_CashFlow_RejectsInvalidStart(t *testing.T) {
	svc := statementTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/cash-flow?start=bogus&end=2026-08-31", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestStatementHandler_CashFlow_RejectsInvalidEnd(t *testing.T) {
	svc := statementTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/cash-flow?start=2026-08-01&end=bogus", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestStatementHandler_CashFlow_RejectsEndBeforeStart(t *testing.T) {
	svc := statementTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/cash-flow?start=2026-08-31&end=2026-08-01", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestStatementHandler_CashFlow_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := statementTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, false)

	resp, err := doRequest(app, http.MethodGet, "/reports/cash-flow?start=2026-08-01&end=2026-08-31", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestStatementHandler_CashFlow_ReturnsServerError(t *testing.T) {
	reports := reporting.ReportDAOMock{
		CashFlowFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]reporting.CashFlowSectionRow, error) {
			return nil, errors.New("db down")
		},
	}
	svc := statementTestSvc(reports, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/cash-flow?start=2026-08-01&end=2026-08-31", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestStatementHandler_YearEndRoll_RollsYearEnd(t *testing.T) {
	years := reporting.TaxYearFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*reference.TaxYear, error) {
			return sampleTaxYear(), nil
		},
	}
	periodByDate := reporting.TaxPeriodByDateFinderMock{
		FindByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{Base: model.Base{ID: 1}, State: accounting.TaxPeriodStateOpen}, nil
		},
	}
	reports := reporting.ReportDAOMock{
		HasYearEndCloseFn: func(_ context.Context, _, _ uint64) (bool, error) {
			return false, nil
		},
		StatementBalancesFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]reporting.AccountBalanceRow, error) {
			return []reporting.AccountBalanceRow{{AccountID: 1, Code: "4000", AccountType: "income", Balance: -500}}, nil
		},
	}
	svc := statementTestSvc(reports, reporting.TaxPeriodFinderMock{}, periodByDate, years, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	body := `{"tax_year_id":1,"journal_id":2,"retained_earnings_account_id":3}`
	resp, err := doRequest(app, http.MethodPost, "/period-close/year-end-roll", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestStatementHandler_YearEndRoll_RejectsValidation(t *testing.T) {
	svc := statementTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/period-close/year-end-roll", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestStatementHandler_YearEndRoll_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := statementTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, reporting.TaxYearFinderMock{}, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, false)

	body := `{"tax_year_id":1,"journal_id":2,"retained_earnings_account_id":3}`
	resp, err := doRequest(app, http.MethodPost, "/period-close/year-end-roll", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestStatementHandler_YearEndRoll_ReturnsNotFound(t *testing.T) {
	years := reporting.TaxYearFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*reference.TaxYear, error) {
			return nil, nil
		},
	}
	svc := statementTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.TaxPeriodByDateFinderMock{}, years, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	body := `{"tax_year_id":1,"journal_id":2,"retained_earnings_account_id":3}`
	resp, err := doRequest(app, http.MethodPost, "/period-close/year-end-roll", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestStatementHandler_YearEndRoll_RejectsAlreadyClosed(t *testing.T) {
	years := reporting.TaxYearFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*reference.TaxYear, error) {
			return sampleTaxYear(), nil
		},
	}
	periodByDate := reporting.TaxPeriodByDateFinderMock{
		FindByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{Base: model.Base{ID: 1}, State: accounting.TaxPeriodStateOpen}, nil
		},
	}
	reports := reporting.ReportDAOMock{
		HasYearEndCloseFn: func(_ context.Context, _, _ uint64) (bool, error) {
			return true, nil
		},
	}
	svc := statementTestSvc(reports, reporting.TaxPeriodFinderMock{}, periodByDate, years, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	body := `{"tax_year_id":1,"journal_id":2,"retained_earnings_account_id":3}`
	resp, err := doRequest(app, http.MethodPost, "/period-close/year-end-roll", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestStatementHandler_YearEndRoll_ReturnsServerError(t *testing.T) {
	years := reporting.TaxYearFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*reference.TaxYear, error) {
			return sampleTaxYear(), nil
		},
	}
	periodByDate := reporting.TaxPeriodByDateFinderMock{
		FindByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{Base: model.Base{ID: 1}, State: accounting.TaxPeriodStateOpen}, nil
		},
	}
	reports := reporting.ReportDAOMock{
		HasYearEndCloseFn: func(_ context.Context, _, _ uint64) (bool, error) {
			return false, nil
		},
		StatementBalancesFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]reporting.AccountBalanceRow, error) {
			return nil, errors.New("db down")
		},
	}
	svc := statementTestSvc(reports, reporting.TaxPeriodFinderMock{}, periodByDate, years, reporting.StatementPosterMock{})
	app := handlerTestApp(t, NewStatementHandler(svc).Register, true)

	body := `{"tax_year_id":1,"journal_id":2,"retained_earnings_account_id":3}`
	resp, err := doRequest(app, http.MethodPost, "/period-close/year-end-roll", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestStatementHandler_WriteStatementError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "period not found", err: reporting.ErrPeriodNotFound, want: http.StatusNotFound},
		{name: "year end already closed", err: reporting.ErrYearEndAlreadyClosed, want: http.StatusUnprocessableEntity},
		{name: "period locked", err: reporting.ErrPeriodLocked, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeStatementError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
