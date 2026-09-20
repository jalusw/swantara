package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

func TestAccrualHandler_Create_CreatesAccrual(t *testing.T) {
	accruals := reporting.AccrualDAOMock{
		CreateFn: func(_ context.Context, entity *reporting.Accrual) (*reporting.Accrual, error) {
			entity.ID = 1
			return entity, nil
		},
	}
	svc := accrualTestSvc(accruals, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	body := `{"period_id":1,"name":"Rent","date":"2026-08-15","reversal_date":"2026-09-01","lines":[{"account_id":10,"name":"Rent","debit":1000},{"account_id":20,"name":"Payable","credit":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/accruals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestAccrualHandler_Create_RejectsValidation(t *testing.T) {
	svc := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/accruals/", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAccrualHandler_Create_RejectsInvalidDate(t *testing.T) {
	svc := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	body := `{"period_id":1,"name":"Rent","date":"bogus","lines":[{"account_id":10,"debit":1000},{"account_id":20,"credit":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/accruals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAccrualHandler_Create_RejectsInvalidReversalDate(t *testing.T) {
	svc := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	body := `{"period_id":1,"name":"Rent","reversal_date":"bogus","lines":[{"account_id":10,"debit":1000},{"account_id":20,"credit":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/accruals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAccrualHandler_Create_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, false)

	body := `{"period_id":1,"name":"Rent","lines":[{"account_id":10,"debit":1000},{"account_id":20,"credit":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/accruals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAccrualHandler_Create_RejectsUnbalancedLines(t *testing.T) {
	svc := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	body := `{"period_id":1,"name":"Rent","lines":[{"account_id":10,"debit":1000},{"account_id":20,"credit":500}]}`
	resp, err := doRequest(app, http.MethodPost, "/accruals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAccrualHandler_Create_ReturnsUnprocessableWhenConfigMissing(t *testing.T) {
	config := reporting.ConfigSourceMock{
		JournalIDFn: func(_ context.Context, _ uint64) (uint64, error) {
			return 0, reporting.ErrConfigMissing
		},
	}
	svc := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, config)
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	body := `{"period_id":1,"name":"Rent","lines":[{"account_id":10,"debit":1000},{"account_id":20,"credit":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/accruals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAccrualHandler_Create_ReturnsServerError(t *testing.T) {
	poster := &reporting.PosterMock{
		PostFn: func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return nil, errors.New("db down")
		},
	}
	svc := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, poster, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	body := `{"period_id":1,"name":"Rent","lines":[{"account_id":10,"debit":1000},{"account_id":20,"credit":1000}]}`
	resp, err := doRequest(app, http.MethodPost, "/accruals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAccrualHandler_ReverseDue_ReversesDueAccruals(t *testing.T) {
	accruals := reporting.AccrualDAOMock{
		ListByOrganizationFn: func(_ context.Context, _ uint64) ([]*reporting.Accrual, error) {
			return []*reporting.Accrual{samplePostedAccrual()}, nil
		},
	}
	svc := accrualTestSvc(accruals, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	body := `{"as_of":"2026-09-01"}`
	resp, err := doRequest(app, http.MethodPost, "/accruals/reverse-due", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAccrualHandler_ReverseDue_RejectsValidation(t *testing.T) {
	svc := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/accruals/reverse-due", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAccrualHandler_ReverseDue_RejectsInvalidDate(t *testing.T) {
	svc := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/accruals/reverse-due", `{"as_of":"bogus"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAccrualHandler_ReverseDue_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, false)

	resp, err := doRequest(app, http.MethodPost, "/accruals/reverse-due", `{"as_of":"2026-09-01"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAccrualHandler_ReverseDue_ReturnsServerError(t *testing.T) {
	accruals := reporting.AccrualDAOMock{
		ListByOrganizationFn: func(_ context.Context, _ uint64) ([]*reporting.Accrual, error) {
			return nil, errors.New("db down")
		},
	}
	svc := accrualTestSvc(accruals, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodPost, "/accruals/reverse-due", `{"as_of":"2026-09-01"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAccrualHandler_List_ReturnsAccruals(t *testing.T) {
	accruals := reporting.AccrualDAOMock{
		ListByOrganizationFn: func(_ context.Context, _ uint64) ([]*reporting.Accrual, error) {
			return []*reporting.Accrual{{State: reporting.AccrualStatePosted}}, nil
		},
	}
	svc := accrualTestSvc(accruals, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/accruals/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAccrualHandler_List_ReturnsServerError(t *testing.T) {
	accruals := reporting.AccrualDAOMock{
		ListByOrganizationFn: func(_ context.Context, _ uint64) ([]*reporting.Accrual, error) {
			return nil, errors.New("db down")
		},
	}
	svc := accrualTestSvc(accruals, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/accruals/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAccrualHandler_List_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := accrualTestSvc(reporting.AccrualDAOMock{}, reporting.AccrualLineDAOMock{}, &reporting.PosterMock{}, reporting.ConfigSourceMock{})
	app := handlerTestApp(t, NewAccrualHandler(svc).Register, false)

	resp, err := doRequest(app, http.MethodGet, "/accruals/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAccrualHandler_WriteAccrualError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "unbalanced accrual", err: reporting.ErrUnbalancedAccrual, want: http.StatusUnprocessableEntity},
		{name: "config missing", err: reporting.ErrConfigMissing, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeAccrualError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
