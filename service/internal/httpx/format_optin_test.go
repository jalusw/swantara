package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type optInRow struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type optInCreateRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func readOptBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(buf)
}

func TestOptIn_ResponseDefaultsToJSON(t *testing.T) {
	app := fiber.New()
	app.Get("/item", func(c fiber.Ctx) error {
		return CreateSuccessResponse(c, "ok", optInRow{ID: 1, Name: "Ada"})
	})
	req := httptest.NewRequest(http.MethodGet, "/item", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body := readOptBody(t, resp)
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		t.Errorf("content-type = %q, want json", resp.Header.Get(fiber.HeaderContentType))
	}
	if !strings.Contains(body, `"success":true`) {
		t.Errorf("body = %q, want json envelope", body)
	}
}

func TestOptIn_LegacyAllowsXMLWithoutOptIn(t *testing.T) {
	app := fiber.New()
	app.Get("/item", func(c fiber.Ctx) error {
		return CreateSuccessResponse(c, "ok", optInRow{ID: 1, Name: "Ada"})
	})
	req := httptest.NewRequest(http.MethodGet, "/item?format=xml", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), fiber.MIMEApplicationXML) {
		t.Errorf("content-type = %q, want xml (legacy allow-all)", resp.Header.Get(fiber.HeaderContentType))
	}
}

func TestOptIn_RestrictedToJSONRejectsExplicitXML(t *testing.T) {
	app := fiber.New()
	app.Get("/item", func(c fiber.Ctx) error {
		AllowResponseFormats(c, FormatJSON)
		return CreateSuccessResponse(c, "ok", optInRow{ID: 1, Name: "Ada"})
	})
	req := httptest.NewRequest(http.MethodGet, "/item?format=xml", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body := readOptBody(t, resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for disallowed explicit xml", resp.StatusCode)
	}
	if !strings.Contains(body, "not available") {
		t.Errorf("body = %q, want not-available message", body)
	}
}

func TestOptIn_RestrictedToJSONFallsBackOnAcceptHeader(t *testing.T) {
	app := fiber.New()
	app.Get("/item", func(c fiber.Ctx) error {
		AllowResponseFormats(c, FormatJSON)
		return CreateSuccessResponse(c, "ok", optInRow{ID: 1, Name: "Ada"})
	})
	req := httptest.NewRequest(http.MethodGet, "/item", nil)
	req.Header.Set("Accept", fiber.MIMEApplicationXML)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body := readOptBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 fallback", resp.StatusCode)
	}
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		t.Errorf("content-type = %q, want json fallback", resp.Header.Get(fiber.HeaderContentType))
	}
	if !strings.Contains(body, `"success":true`) {
		t.Errorf("body = %q, want json envelope", body)
	}
}

type optInListResponse struct {
	Rows []optInRow `json:"rows"`
}

func TestOptIn_WithFormatsMiddlewareAllowsXMLAndCSV(t *testing.T) {
	app := fiber.New()
	app.Get("/list", WithFormats(FormatJSON, FormatXML, FormatCSV), func(c fiber.Ctx) error {
		return CreateSuccessResponseWithMeta(c, "List.", optInListResponse{
			Rows: []optInRow{{ID: 1, Name: "Alice"}},
		}, Meta{Pagination: &PaginationMeta{Page: 1, PerPage: 20, Total: 1, TotalPages: 1}})
	})
	req := httptest.NewRequest(http.MethodGet, "/list?format=xml", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = readOptBody(t, resp)
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), fiber.MIMEApplicationXML) {
		t.Errorf("content-type = %q, want xml", resp.Header.Get(fiber.HeaderContentType))
	}

	req = httptest.NewRequest(http.MethodGet, "/list?format=csv", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body := readOptBody(t, resp)
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), "text/csv") {
		t.Errorf("content-type = %q, want csv", resp.Header.Get(fiber.HeaderContentType))
	}
	if !strings.Contains(body, "Alice") {
		t.Errorf("csv body = %q, want row", body)
	}
}

func TestOptIn_WriteListResponseCustomFilename(t *testing.T) {
	app := fiber.New()
	app.Get("/contacts", func(c fiber.Ctx) error {
		items := []optInRow{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		return WriteListResponse(c, "ok", map[string]any{"contacts": items}, items,
			BuildListMeta(&query.Query{Pagination: &query.Pagination{Page: 1, Size: 20}}, 2), "contacts.csv")
	})
	req := httptest.NewRequest(http.MethodGet, "/contacts?format=csv", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body := readOptBody(t, resp)
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), "text/csv") {
		t.Errorf("content-type = %q, want csv", resp.Header.Get(fiber.HeaderContentType))
	}
	if cd := resp.Header.Get("Content-Disposition"); !strings.Contains(cd, "contacts.csv") {
		t.Errorf("content-disposition = %q, want contacts.csv", cd)
	}
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if len(lines) != 3 || lines[0] != "id,name" {
		t.Errorf("csv body = %q, want header + 2 rows", body)
	}

	req = httptest.NewRequest(http.MethodGet, "/contacts", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body = readOptBody(t, resp)
	if !strings.Contains(body, `"success":true`) {
		t.Errorf("json body = %q, want envelope", body)
	}
}

func TestOptIn_RequestJSONDefault(t *testing.T) {
	app := fiber.New()
	app.Post("/create", func(c fiber.Ctx) error {
		var req optInCreateRequest
		if !BindAndValidate(c, &req) {
			return nil
		}
		return CreateCreatedResponse(c, "created", req)
	})
	req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(`{"name":"Ada","email":"ada@example.com"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = readOptBody(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
}

func TestOptIn_RequestXMLAndCSVByDefault(t *testing.T) {
	app := fiber.New()
	app.Post("/create", func(c fiber.Ctx) error {
		var req optInCreateRequest
		if !BindAndValidate(c, &req) {
			return nil
		}
		return CreateCreatedResponse(c, "created", req)
	})

	xmlBody := `<optInCreateRequest><name>Ada</name><email>ada@example.com</email></optInCreateRequest>`
	req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(xmlBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationXML)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("xml request failed: %v", err)
	}
	_ = readOptBody(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("xml status = %d, want 201 (legacy allow-all)", resp.StatusCode)
	}

	csvBody := "name,email\nAda,ada@example.com\n"
	req = httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(csvBody))
	req.Header.Set(fiber.HeaderContentType, "text/csv")
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("csv request failed: %v", err)
	}
	_ = readOptBody(t, resp)
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("csv status = %d, want 201", resp.StatusCode)
	}
}

func TestOptIn_RequestRestrictedToJSONRejectsXMLAndCSV(t *testing.T) {
	app := fiber.New()
	app.Post("/create", func(c fiber.Ctx) error {
		AllowRequestFormats(c, FormatJSON)
		var req optInCreateRequest
		if !BindAndValidate(c, &req) {
			return nil
		}
		return CreateCreatedResponse(c, "created", req)
	})

	req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader(`<optInCreateRequest><name>Ada</name></optInCreateRequest>`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationXML)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body := readOptBody(t, resp)
	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("xml status = %d, want 415", resp.StatusCode)
	}
	if !strings.Contains(body, ErrUnsupportedMediaTypeCode) {
		t.Errorf("body = %q, want 415 envelope", body)
	}

	req = httptest.NewRequest(http.MethodPost, "/create", strings.NewReader("name,email\nAda,ada@example.com\n"))
	req.Header.Set(fiber.HeaderContentType, "text/csv")
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = readOptBody(t, resp)
	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("csv status = %d, want 415", resp.StatusCode)
	}
}

func TestOptIn_RequestUnsupportedMediaType(t *testing.T) {
	app := fiber.New()
	app.Post("/create", func(c fiber.Ctx) error {
		var req optInCreateRequest
		if !BindAndValidate(c, &req) {
			return nil
		}
		return CreateCreatedResponse(c, "created", req)
	})
	req := httptest.NewRequest(http.MethodPost, "/create", strings.NewReader("raw bytes"))
	req.Header.Set(fiber.HeaderContentType, "application/octet-stream")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	_ = readOptBody(t, resp)
	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want 415", resp.StatusCode)
	}
}

func TestOptIn_DecodeCSVSlice(t *testing.T) {
	var rows []optInRow
	if err := decodeCSVBody([]byte("id,name\n1,Alice\n2,Bob\n"), &rows); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(rows) != 2 || rows[0].Name != "Alice" || rows[1].ID != 2 {
		t.Errorf("rows = %+v, want 2 parsed", rows)
	}
}

func TestOptIn_DecodeCSVErrors(t *testing.T) {
	var row optInRow
	if err := decodeCSVBody([]byte(""), &row); err == nil {
		t.Error("empty csv expected error")
	}
	if err := decodeCSVBody([]byte("id,name\n"), &row); err == nil {
		t.Error("header-only csv expected error for struct")
	}
	if err := decodeCSVBody([]byte("id,name\nnotanint,Alice\n"), &row); err == nil {
		t.Error("bad int expected error")
	}
	var notPtr optInRow
	_ = notPtr
	if err := decodeCSVBody([]byte("id,name\n1,Alice\n"), optInRow{}); err == nil {
		t.Error("non-pointer expected error")
	}
}
