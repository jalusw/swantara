package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type formatRow struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type formatList struct {
	Rows []formatRow `json:"rows"`
}

func formatTestApp() *fiber.App {
	app := fiber.New()
	app.Get("/list", func(c fiber.Ctx) error {
		return CreateSuccessResponseWithMeta(c, "List.", formatList{
			Rows: []formatRow{
				{ID: 1, Name: "Alice"},
				{ID: 2, Name: "Bob"},
			},
		}, Meta{
			Pagination: &PaginationMeta{Page: 1, PerPage: 2, Total: 5, TotalPages: 3},
		})
	})
	app.Get("/single", func(c fiber.Ctx) error {
		return CreateSuccessResponse(c, "Single.", formatRow{ID: 7, Name: "Carol"})
	})
	app.Get("/error", func(c fiber.Ctx) error {
		return CreateNotFoundResponse(c, "Missing.")
	})
	return app
}

func requestFormatTest(t *testing.T, url, accept string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, url, nil)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	resp, err := formatTestApp().Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func TestResponseNegotiation_JSONIsDefault(t *testing.T) {
	resp := requestFormatTest(t, "/list", "")
	defer func() { _ = resp.Body.Close() }()
	contentType := resp.Header.Get(fiber.HeaderContentType)
	if !strings.HasPrefix(contentType, fiber.MIMEApplicationJSON) {
		t.Errorf("content type = %q, want json", contentType)
	}
}

func TestResponseNegotiation_XMLViaAccept(t *testing.T) {
	resp := requestFormatTest(t, "/list", fiber.MIMEApplicationXML)
	defer func() { _ = resp.Body.Close() }()
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), fiber.MIMEApplicationXML) {
		t.Errorf("content type = %q, want xml", resp.Header.Get(fiber.HeaderContentType))
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "<success>true</success>") || !strings.Contains(body, "<message>List.</message>") {
		t.Errorf("xml body = %q, want success envelope", body)
	}
}

func TestResponseNegotiation_XMLViaQueryParam(t *testing.T) {
	resp := requestFormatTest(t, "/single?format=xml", "")
	defer func() { _ = resp.Body.Close() }()
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), fiber.MIMEApplicationXML) {
		t.Errorf("content type = %q, want xml", resp.Header.Get(fiber.HeaderContentType))
	}
}

func TestResponseNegotiation_CSVListDumpsRowsWithPaginationHeaders(t *testing.T) {
	resp := requestFormatTest(t, "/list?format=csv", "")
	defer func() { _ = resp.Body.Close() }()
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), "text/csv") {
		t.Errorf("content type = %q, want csv", resp.Header.Get(fiber.HeaderContentType))
	}
	body := readBody(t, resp)
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if len(lines) != 3 {
		t.Fatalf("csv lines = %d, want header + 2 rows", len(lines))
	}
	if lines[0] != "id,name" || lines[1] != "1,Alice" || lines[2] != "2,Bob" {
		t.Errorf("csv body = %q, want id,name / 1,Alice / 2,Bob", body)
	}
	if resp.Header.Get("X-Pagination-Page") != "1" || resp.Header.Get("X-Pagination-Total") != "5" || resp.Header.Get("X-Pagination-Total-Pages") != "3" {
		t.Errorf("pagination headers = %q/%q/%q", resp.Header.Get("X-Pagination-Page"), resp.Header.Get("X-Pagination-Total"), resp.Header.Get("X-Pagination-Total-Pages"))
	}
}

func TestResponseNegotiation_CSVListViaAccept(t *testing.T) {
	resp := requestFormatTest(t, "/list", "text/csv")
	defer func() { _ = resp.Body.Close() }()
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), "text/csv") {
		t.Errorf("content type = %q, want csv", resp.Header.Get(fiber.HeaderContentType))
	}
}

func TestResponseNegotiation_CSVSingleResourceReturnsBadRequest(t *testing.T) {
	resp := requestFormatTest(t, "/single?format=csv", "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for csv on singular endpoint", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `"error_code":"ERR_BAD_REQUEST"`) {
		t.Errorf("body = %q, want bad request envelope", body)
	}
}

func TestResponseNegotiation_CSVSingleResourceViaAcceptFallsBackToJSON(t *testing.T) {
	resp := requestFormatTest(t, "/single", "text/csv")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		t.Errorf("content type = %q, want json fallback", resp.Header.Get(fiber.HeaderContentType))
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `"success":true`) {
		t.Errorf("body = %q, want json envelope", body)
	}
}

func TestResponseNegotiation_CSVErrorFallsBackToJSON(t *testing.T) {
	resp := requestFormatTest(t, "/error?format=csv", "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want original 404 preserved", resp.StatusCode)
	}
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		t.Errorf("content type = %q, want json fallback", resp.Header.Get(fiber.HeaderContentType))
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `"success":false`) || !strings.Contains(body, `"error_code":"ERR_NOT_FOUND"`) {
		t.Errorf("body = %q, want json error envelope", body)
	}
}

func TestResponseNegotiation_XMLError(t *testing.T) {
	resp := requestFormatTest(t, "/error", fiber.MIMEApplicationXML)
	defer func() { _ = resp.Body.Close() }()
	if !strings.HasPrefix(resp.Header.Get(fiber.HeaderContentType), fiber.MIMEApplicationXML) {
		t.Errorf("content type = %q, want xml", resp.Header.Get(fiber.HeaderContentType))
	}
	body := readBody(t, resp)
	if !strings.Contains(body, "<Success>false</Success>") || !strings.Contains(body, "<ErrorCode>ERR_NOT_FOUND</ErrorCode>") {
		t.Errorf("xml body = %q, want error envelope", body)
	}
}
