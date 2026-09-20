package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func journalEntryTestApp(t *testing.T, withTenant bool, movements accounting.JournalEntryDAOMock, lines accounting.JournalLineDAOMock) *fiber.App {
	t.Helper()
	poster := accounting.NewPostingService(movements).WithLines(lines)
	h := NewJournalEntryHandler(poster)
	return accountingTestApp(t, withTenant, h.Register)
}

func journalEntryTestAppWithReverser(t *testing.T, movements accounting.JournalEntryDAOMock, lines accounting.JournalLineDAOMock) *fiber.App {
	t.Helper()
	poster := accounting.NewPostingService(movements).WithLines(lines).SetReverser(accounting.NewReversalEngine(movements, lines))
	h := NewJournalEntryHandler(poster)
	return accountingTestApp(t, true, h.Register)
}

func TestJournalEntryHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.JournalEntry], error) {
			return &query.Page[accounting.JournalEntry]{Items: []*accounting.JournalEntry{sampleJournalEntry()}, Count: 1}, nil
		}
		app := journalEntryTestApp(t, true, movements, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/journal-entries/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.JournalEntry], error) {
			return &query.Page[accounting.JournalEntry]{Items: []*accounting.JournalEntry{sampleJournalEntry()}, Count: 1}, nil
		}
		app := journalEntryTestApp(t, true, movements, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/journal-entries/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := journalEntryTestApp(t, true, accounting.JournalEntryDAOMock{}, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/journal-entries/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := journalEntryTestApp(t, false, accounting.JournalEntryDAOMock{}, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/journal-entries/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.JournalEntry], error) {
			return nil, errors.New("boom")
		}
		app := journalEntryTestApp(t, true, movements, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/journal-entries/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestJournalEntryHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return sampleJournalEntry(), nil
		}
		lines := accounting.JournalLineDAOMock{}
		lines.ListByMovementFunc = func(_ context.Context, _ uint64) ([]*accounting.JournalLine, error) {
			return []*accounting.JournalLine{sampleJournalLine()}, nil
		}
		app := journalEntryTestApp(t, true, movements, lines)

		resp, err := doRequest(app, http.MethodGet, "/journal-entries/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := journalEntryTestApp(t, true, accounting.JournalEntryDAOMock{}, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/journal-entries/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return nil, nil
		}
		app := journalEntryTestApp(t, true, movements, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/journal-entries/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("tenant mismatch", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 99}, nil
		}
		app := journalEntryTestApp(t, true, movements, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/journal-entries/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return nil, errors.New("boom")
		}
		app := journalEntryTestApp(t, true, movements, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/journal-entries/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("lines error", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return sampleJournalEntry(), nil
		}
		lines := accounting.JournalLineDAOMock{}
		lines.ListByMovementFunc = func(_ context.Context, _ uint64) ([]*accounting.JournalLine, error) {
			return nil, errors.New("boom")
		}
		app := journalEntryTestApp(t, true, movements, lines)

		resp, err := doRequest(app, http.MethodGet, "/journal-entries/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestJournalEntryHandlerPost(t *testing.T) {
	body := `{"journal_id":1,"date":"2026-01-15","ref":"REF","lines":[{"account_id":100,"debit":100},{"account_id":200,"credit":100}]}`

	t.Run("success", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.CreateWithLinesFunc = func(_ context.Context, entry *accounting.JournalEntry, _ []*accounting.JournalLine) (*accounting.JournalEntry, error) {
			entry.ID = 1
			return entry, nil
		}
		app := journalEntryTestApp(t, true, movements, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/journal-entries", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := journalEntryTestApp(t, true, accounting.JournalEntryDAOMock{}, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/journal-entries", `{"journal_id":1}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		app := journalEntryTestApp(t, true, accounting.JournalEntryDAOMock{}, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/journal-entries", `{"journal_id":1,"date":"bogus","lines":[{"account_id":100,"debit":100},{"account_id":200,"credit":100}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := journalEntryTestApp(t, false, accounting.JournalEntryDAOMock{}, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/journal-entries", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unbalanced", func(t *testing.T) {
		app := journalEntryTestApp(t, true, accounting.JournalEntryDAOMock{}, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/journal-entries", `{"journal_id":1,"date":"2026-01-15","lines":[{"account_id":100,"debit":50},{"account_id":200,"credit":100}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("origin incomplete", func(t *testing.T) {
		app := journalEntryTestApp(t, true, accounting.JournalEntryDAOMock{}, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/journal-entries", `{"journal_id":1,"date":"2026-01-15","origin_id":5,"lines":[{"account_id":100,"debit":100},{"account_id":200,"credit":100}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestJournalEntryHandlerReverse(t *testing.T) {
	body := `{"journal_id":2,"date":"2026-02-01","description":"reversal"}`

	t.Run("success", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return sampleJournalEntry(), nil
		}
		lines := accounting.JournalLineDAOMock{}
		lines.ListByMovementFunc = func(_ context.Context, _ uint64) ([]*accounting.JournalLine, error) {
			return []*accounting.JournalLine{sampleJournalLine()}, nil
		}
		app := journalEntryTestAppWithReverser(t, movements, lines)

		resp, err := doRequest(app, http.MethodPost, "/journal-entries/1/reverse", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := journalEntryTestAppWithReverser(t, accounting.JournalEntryDAOMock{}, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/journal-entries/abc/reverse", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := journalEntryTestAppWithReverser(t, accounting.JournalEntryDAOMock{}, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/journal-entries/1/reverse", `{"journal_id":2}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return nil, nil
		}
		app := journalEntryTestAppWithReverser(t, movements, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/journal-entries/1/reverse", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("lookup error", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return nil, errors.New("boom")
		}
		app := journalEntryTestAppWithReverser(t, movements, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/journal-entries/1/reverse", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return sampleJournalEntry(), nil
		}
		app := journalEntryTestAppWithReverser(t, movements, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/journal-entries/1/reverse", `{"journal_id":2,"description":"x","date":"bogus"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("reverser missing", func(t *testing.T) {
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return sampleJournalEntry(), nil
		}
		app := journalEntryTestApp(t, true, movements, accounting.JournalLineDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/journal-entries/1/reverse", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWriteJournalEntryError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: accounting.ErrEntryNotFound, status: http.StatusNotFound},
		{name: "not posted", err: accounting.ErrEntryNotPosted, status: http.StatusConflict},
		{name: "already reversed", err: accounting.ErrEntryReversed, status: http.StatusConflict},
		{name: "period locked", err: accounting.ErrPeriodLocked, status: http.StatusConflict},
		{name: "reverser missing", err: accounting.ErrReverserMissing, status: http.StatusInternalServerError},
		{name: "no lines", err: accounting.ErrNoLines, status: http.StatusUnprocessableEntity},
		{name: "invalid line", err: accounting.ErrInvalidLine, status: http.StatusUnprocessableEntity},
		{name: "unbalanced", err: accounting.ErrUnbalanced, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeJournalEntryError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
