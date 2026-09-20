package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func deferralTestApp(t *testing.T, withTenant bool, svc accounting.DeferralService) *fiber.App {
	t.Helper()
	h := NewDeferralHandler(svc)
	return accountingTestApp(t, withTenant, func(api fiber.Router, guards httpx.RouteGuards) {
		h.Register(api, guards)
	})
}

func newDeferralService(schedules accounting.DeferredScheduleDAOMock, scheduleLines accounting.DeferredScheduleLineDAOMock) accounting.DeferralService {
	return accounting.NewDeferralService(schedules, scheduleLines, accounting.PosterMock{}, accounting.JournalResolverMock{}, accounting.TransactionerMock{})
}

func deferralScheduleSample() *accounting.DeferredSchedule {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return &accounting.DeferredSchedule{
		Base:                  model.Base{ID: 1},
		OrganizationID:        ptrUint64(10),
		Type:                  accounting.DeferredTypeDeferredRevenue,
		SourceType:            "invoice",
		SourceID:              5,
		TotalAmount:           1200,
		BalanceSheetAccountID: ptrUint64(100),
		PLAccountID:           ptrUint64(200),
		Method:                accounting.DeferredMethodLinear,
		DateStart:             &start,
		State:                 accounting.DeferredStateRunning,
	}
}

func TestDeferralHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		schedules := accounting.DeferredScheduleDAOMock{}
		schedules.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.DeferredSchedule], error) {
			return &query.Page[accounting.DeferredSchedule]{Items: []*accounting.DeferredSchedule{deferralScheduleSample()}}, nil
		}
		svc := newDeferralService(schedules, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/deferrals", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		svc := newDeferralService(accounting.DeferredScheduleDAOMock{}, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodGet, "/deferrals", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		schedules := accounting.DeferredScheduleDAOMock{}
		schedules.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.DeferredSchedule], error) {
			return nil, errors.New("boom")
		}
		svc := newDeferralService(schedules, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/deferrals", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestDeferralHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		schedules := accounting.DeferredScheduleDAOMock{}
		schedules.SearchFunc = func(_ context.Context, _ string, _ interface{}) (*accounting.DeferredSchedule, error) {
			return deferralScheduleSample(), nil
		}
		svc := newDeferralService(schedules, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/deferrals/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		svc := newDeferralService(accounting.DeferredScheduleDAOMock{}, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodGet, "/deferrals/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := newDeferralService(accounting.DeferredScheduleDAOMock{}, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/deferrals/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		schedules := accounting.DeferredScheduleDAOMock{}
		schedules.SearchFunc = func(_ context.Context, _ string, _ interface{}) (*accounting.DeferredSchedule, error) {
			return nil, nil
		}
		svc := newDeferralService(schedules, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/deferrals/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})
}

func TestDeferralHandlerListLines(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		schedules := accounting.DeferredScheduleDAOMock{}
		schedules.SearchFunc = func(_ context.Context, _ string, _ interface{}) (*accounting.DeferredSchedule, error) {
			return deferralScheduleSample(), nil
		}
		lines := accounting.DeferredScheduleLineDAOMock{}
		lines.ListByScheduleFunc = func(_ context.Context, _ uint64) ([]*accounting.DeferredScheduleLine, error) {
			start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
			return []*accounting.DeferredScheduleLine{{Base: model.Base{ID: 1}, ScheduleID: 1, Sequence: 1, RecognitionDate: &start, Amount: 100}}, nil
		}
		svc := newDeferralService(schedules, lines)
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/deferrals/1/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		svc := newDeferralService(accounting.DeferredScheduleDAOMock{}, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodGet, "/deferrals/1/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := newDeferralService(accounting.DeferredScheduleDAOMock{}, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/deferrals/abc/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("schedule not found", func(t *testing.T) {
		schedules := accounting.DeferredScheduleDAOMock{}
		schedules.SearchFunc = func(_ context.Context, _ string, _ interface{}) (*accounting.DeferredSchedule, error) {
			return nil, nil
		}
		svc := newDeferralService(schedules, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/deferrals/1/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("lines error", func(t *testing.T) {
		schedules := accounting.DeferredScheduleDAOMock{}
		schedules.SearchFunc = func(_ context.Context, _ string, _ interface{}) (*accounting.DeferredSchedule, error) {
			return deferralScheduleSample(), nil
		}
		lines := accounting.DeferredScheduleLineDAOMock{}
		lines.ListByScheduleFunc = func(_ context.Context, _ uint64) ([]*accounting.DeferredScheduleLine, error) {
			return nil, errors.New("boom")
		}
		svc := newDeferralService(schedules, lines)
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/deferrals/1/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestDeferralHandlerRecognize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		schedules := accounting.DeferredScheduleDAOMock{}
		schedules.ListRunningFunc = func(_ context.Context, _ *uint64) ([]*accounting.DeferredSchedule, error) {
			return []*accounting.DeferredSchedule{deferralScheduleSample()}, nil
		}
		lines := accounting.DeferredScheduleLineDAOMock{}
		start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
		lines.ListDueFunc = func(_ context.Context, _ uint64, _ time.Time) ([]*accounting.DeferredScheduleLine, error) {
			return []*accounting.DeferredScheduleLine{{Base: model.Base{ID: 1}, ScheduleID: 1, Sequence: 1, RecognitionDate: &start, Amount: 100}}, nil
		}
		svc := newDeferralService(schedules, lines)
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/deferrals/recognize", `{"as_of":"2026-03-01"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		svc := newDeferralService(accounting.DeferredScheduleDAOMock{}, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/deferrals/recognize", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		svc := newDeferralService(accounting.DeferredScheduleDAOMock{}, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodPost, "/deferrals/recognize", `{"as_of":"2026-03-01"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		svc := newDeferralService(accounting.DeferredScheduleDAOMock{}, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/deferrals/recognize", `{"as_of":"bogus"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestDeferralHandlerCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		schedules := accounting.DeferredScheduleDAOMock{}
		schedules.CreateTxFunc = func(_ context.Context, _ *gorm.DB, _ *accounting.DeferredSchedule) (*accounting.DeferredSchedule, error) {
			return deferralScheduleSample(), nil
		}
		lines := accounting.DeferredScheduleLineDAOMock{}
		lines.CreateTxFunc = func(_ context.Context, _ *gorm.DB, _ *accounting.DeferredScheduleLine) (*accounting.DeferredScheduleLine, error) {
			return &accounting.DeferredScheduleLine{Base: model.Base{ID: 1}}, nil
		}
		svc := newDeferralService(schedules, lines)
		app := deferralTestApp(t, true, svc)

		body := `{"type":"deferred_revenue","source_type":"invoice","source_id":5,"total_amount":1200,"balance_sheet_account_id":100,"pl_account_id":200,"method":"manual","date_start":"2026-01-01T00:00:00Z","lines":[{"recognition_date":"2026-02-01T00:00:00Z","amount":600}]}`
		resp, err := doRequest(app, http.MethodPost, "/deferrals", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		svc := newDeferralService(accounting.DeferredScheduleDAOMock{}, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, false, svc)

		body := `{"type":"deferred_revenue","source_type":"invoice","source_id":5,"total_amount":1200,"balance_sheet_account_id":100,"pl_account_id":200,"method":"manual","date_start":"2026-01-01T00:00:00Z","lines":[{"recognition_date":"2026-02-01T00:00:00Z","amount":600}]}`
		resp, err := doRequest(app, http.MethodPost, "/deferrals", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		svc := newDeferralService(accounting.DeferredScheduleDAOMock{}, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/deferrals", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		schedules := accounting.DeferredScheduleDAOMock{}
		svc := newDeferralService(schedules, accounting.DeferredScheduleLineDAOMock{})
		app := deferralTestApp(t, true, svc)

		body := `{"type":"bogus","source_type":"invoice","source_id":5,"total_amount":1200,"balance_sheet_account_id":100,"pl_account_id":200,"method":"manual","date_start":"2026-01-01T00:00:00Z","lines":[{"recognition_date":"2026-02-01T00:00:00Z","amount":600}]}`
		resp, err := doRequest(app, http.MethodPost, "/deferrals", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestWriteDeferralError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: accounting.ErrScheduleNotFound, status: http.StatusNotFound},
		{name: "type", err: accounting.ErrScheduleType, status: http.StatusUnprocessableEntity},
		{name: "method", err: accounting.ErrScheduleMethod, status: http.StatusUnprocessableEntity},
		{name: "amount", err: accounting.ErrScheduleAmount, status: http.StatusUnprocessableEntity},
		{name: "no account", err: accounting.ErrScheduleNoAccount, status: http.StatusUnprocessableEntity},
		{name: "invalid date", err: accounting.ErrScheduleInvalidDate, status: http.StatusUnprocessableEntity},
		{name: "no lines", err: accounting.ErrScheduleNoLines, status: http.StatusUnprocessableEntity},
		{name: "line amount", err: accounting.ErrScheduleLineAmount, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeDeferralError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
