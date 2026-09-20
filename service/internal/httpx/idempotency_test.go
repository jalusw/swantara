package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type fakeIdempotencyReader struct {
	entry     *IdempotencyEntry
	claimErr  error
	handled   int
	claimed   []*uint64
	completed []string
	released  []string
}

func (f *fakeIdempotencyReader) Claim(_ context.Context, organizationID *uint64, _, _ string) (*IdempotencyEntry, error) {
	f.claimed = append(f.claimed, organizationID)
	if f.claimErr != nil {
		return nil, f.claimErr
	}
	return f.entry, nil
}

func (f *fakeIdempotencyReader) Complete(_ context.Context, _ *uint64, key string, statusCode int, response []byte) error {
	f.completed = append(f.completed, key)
	f.entry = &IdempotencyEntry{StatusCode: statusCode, Response: response}
	return nil
}

func (f *fakeIdempotencyReader) Release(_ context.Context, _ *uint64, key string) error {
	f.released = append(f.released, key)
	return nil
}

func tenantApp(reader IdempotencyReader, handler fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(LocalOrganizationID, uint64(10))
		return c.Next()
	})
	app.Post("/post", IdempotencyGuard(reader, "post"), handler)
	return app
}

func idempotencyRequest(app *fiber.App) *http.Response {
	req := httptest.NewRequest("POST", "/post", nil)
	req.Header.Set("Idempotency-Key", "abc")
	resp, err := app.Test(req)
	if err != nil {
		panic(err)
	}
	return resp
}

func TestIdempotencyGuard_ReplaysStoredResponse(t *testing.T) {
	reader := &fakeIdempotencyReader{entry: &IdempotencyEntry{StatusCode: 200, Response: []byte(`{"replayed":true}`)}}
	app := tenantApp(reader, func(c fiber.Ctx) error {
		reader.handled++
		return c.SendString("original")
	})

	resp := idempotencyRequest(app)
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if body := readBody(t, resp); body != `{"replayed":true}` {
		t.Errorf("body = %s, want replayed response", body)
	}
	if reader.handled != 0 {
		t.Errorf("handler invoked %d times, want 0", reader.handled)
	}
	if len(reader.claimed) != 1 || reader.claimed[0] == nil || *reader.claimed[0] != 10 {
		t.Errorf("claim organizations = %v, want [10]", reader.claimed)
	}
}

func TestIdempotencyGuard_RecordsSuccessfulFirstCall(t *testing.T) {
	reader := &fakeIdempotencyReader{}
	app := tenantApp(reader, func(c fiber.Ctx) error {
		reader.handled++
		return c.SendString("created")
	})

	resp := idempotencyRequest(app)
	if resp.StatusCode != 200 || reader.handled != 1 {
		t.Errorf("status = %d handled = %d, want 200 and 1", resp.StatusCode, reader.handled)
	}
	if len(reader.completed) != 1 || reader.completed[0] != "abc" {
		t.Errorf("completed = %v, want ['abc']", reader.completed)
	}

	resp2 := idempotencyRequest(app)
	if body := readBody(t, resp2); body != "created" {
		t.Errorf("replay body = %s, want 'created'", body)
	}
	if reader.handled != 1 {
		t.Errorf("handler invoked %d times after replay, want 1", reader.handled)
	}
}

func TestIdempotencyGuard_ReleasesClaimOnHandlerError(t *testing.T) {
	reader := &fakeIdempotencyReader{}
	app := tenantApp(reader, func(c fiber.Ctx) error {
		reader.handled++
		return errors.New("boom")
	})

	resp := idempotencyRequest(app)
	if resp.StatusCode != 500 {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
	if len(reader.completed) != 0 {
		t.Errorf("completed = %v, want none", reader.completed)
	}
	if len(reader.released) != 1 || reader.released[0] != "abc" {
		t.Errorf("released = %v, want ['abc']", reader.released)
	}
}

func TestIdempotencyGuard_ReleasesClaimOnNonSuccessStatus(t *testing.T) {
	reader := &fakeIdempotencyReader{}
	app := tenantApp(reader, func(c fiber.Ctx) error {
		reader.handled++
		return c.Status(422).SendString("invalid")
	})

	resp := idempotencyRequest(app)
	if resp.StatusCode != 422 {
		t.Errorf("status = %d, want 422", resp.StatusCode)
	}
	if len(reader.completed) != 0 {
		t.Errorf("completed = %v, want none", reader.completed)
	}
	if len(reader.released) != 1 {
		t.Errorf("released = %v, want one release", reader.released)
	}
}

func TestIdempotencyGuard_InFlightKeyReturnsConflict(t *testing.T) {
	reader := &fakeIdempotencyReader{claimErr: ErrIdempotencyKeyInFlight}
	app := tenantApp(reader, func(c fiber.Ctx) error {
		reader.handled++
		return c.SendString("created")
	})

	resp := idempotencyRequest(app)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("status = %d, want 409", resp.StatusCode)
	}
	if reader.handled != 0 {
		t.Errorf("handler invoked %d times, want 0", reader.handled)
	}
}

func TestIdempotencyGuard_ReusedKeyReturnsUnprocessable(t *testing.T) {
	reader := &fakeIdempotencyReader{claimErr: ErrIdempotencyKeyReused}
	app := tenantApp(reader, func(c fiber.Ctx) error {
		reader.handled++
		return c.SendString("created")
	})

	resp := idempotencyRequest(app)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", resp.StatusCode)
	}
	if reader.handled != 0 {
		t.Errorf("handler invoked %d times, want 0", reader.handled)
	}
}

func TestIdempotencyGuard_PassesThroughWithoutOrganization(t *testing.T) {
	reader := &fakeIdempotencyReader{}
	app := fiber.New()
	app.Post("/post", IdempotencyGuard(reader, "post"), func(c fiber.Ctx) error {
		reader.handled++
		return c.SendString("created")
	})

	resp := idempotencyRequest(app)
	if resp.StatusCode != 200 || reader.handled != 1 {
		t.Errorf("status = %d handled = %d, want 200 and 1", resp.StatusCode, reader.handled)
	}
	if len(reader.claimed) != 0 || len(reader.completed) != 0 {
		t.Errorf("claims = %d completed = %d, want none", len(reader.claimed), len(reader.completed))
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}
	return string(body)
}
