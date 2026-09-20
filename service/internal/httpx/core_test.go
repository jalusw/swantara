package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func doTestRequest(app *fiber.App, method, url string) *http.Response {
	req := httptest.NewRequest(method, url, nil)
	resp, err := app.Test(req)
	if err != nil {
		panic(err)
	}
	return resp
}

func TestCallerIDAndRoles(t *testing.T) {
	app := fiber.New()
	app.Get("/caller", func(c fiber.Ctx) error {
		id, idOk := CallerID(c)
		orgID, orgOk := CallerOrganizationID(c)
		return c.JSON(fiber.Map{"id": id, "id_ok": idOk, "org": orgID, "org_ok": orgOk})
	})
	app.Get("/with-locals", func(c fiber.Ctx) error {
		c.Locals(LocalUserID, uint64(1))
		c.Locals(LocalOrganizationID, uint64(3))
		id, idOk := CallerID(c)
		orgID, orgOk := CallerOrganizationID(c)
		return c.JSON(fiber.Map{"id": id, "id_ok": idOk, "org": orgID, "org_ok": orgOk})
	})

	resp := doTestRequest(app, http.MethodGet, "/caller")
	var out map[string]any
	if err := jsonDecode(resp, &out); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if out["id_ok"] != false || out["org_ok"] != false {
		t.Errorf("expected no locals, got %v", out)
	}

	resp = doTestRequest(app, http.MethodGet, "/with-locals")
	if err := jsonDecode(resp, &out); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if out["id"] != float64(1) || out["org"] != float64(3) {
		t.Errorf("expected locals 1/3, got %v", out)
	}
}

func TestResolveOrganizationID(t *testing.T) {
	app := fiber.New()
	app.Get("/orgs/:organization_id", func(c fiber.Ctx) error {
		id, ok := ResolveOrganizationID(c)
		return c.JSON(fiber.Map{"id": id, "ok": ok})
	})
	app.Get("/legacy/:id", func(c fiber.Ctx) error {
		id, ok := ResolveOrganizationID(c)
		return c.JSON(fiber.Map{"id": id, "ok": ok})
	})
	app.Get("/legacy/:id/modules", func(c fiber.Ctx) error {
		id, ok := ResolveOrganizationID(c)
		return c.JSON(fiber.Map{"id": id, "ok": ok})
	})
	app.Get("/by-header", func(c fiber.Ctx) error {
		id, ok := ResolveOrganizationID(c)
		return c.JSON(fiber.Map{"id": id, "ok": ok})
	})
	app.Get("/none", func(c fiber.Ctx) error {
		id, ok := ResolveOrganizationID(c)
		return c.JSON(fiber.Map{"id": id, "ok": ok})
	})

	check := func(url string, setHeader bool, wantID float64, wantOK bool) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, url, nil)
		if setHeader {
			req.Header.Set(OrganizationHeader, "7")
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error = %v", err)
		}
		var out map[string]any
		if err := jsonDecode(resp, &out); err != nil {
			t.Fatalf("decode error = %v", err)
		}
		if out["id"] != wantID || out["ok"] != wantOK {
			t.Errorf("%s => id=%v ok=%v, want id=%v ok=%v", url, out["id"], out["ok"], wantID, wantOK)
		}
	}

	check("/orgs/5", false, 5, true)
	check("/orgs/not-a-number", false, 0, false)
	check("/legacy/2", false, 2, true)
	check("/legacy/2/modules", false, 2, true)
	check("/legacy/not-a-number", false, 0, false)
	check("/by-header", true, 7, true)
	check("/none", true, 7, true)
	check("/none", false, 0, false)
}

func TestTenantOrganizationID(t *testing.T) {
	app := fiber.New()
	app.Get("/tenant", func(c fiber.Ctx) error {
		fallback := uint64(99)
		got := TenantOrganizationID(c, &fallback)
		c.Locals(LocalOrganizationID, uint64(42))
		gotWithTenant := TenantOrganizationID(c, &fallback)
		return c.JSON(fiber.Map{"fallback": got, "tenant": gotWithTenant})
	})

	resp := doTestRequest(app, http.MethodGet, "/tenant")
	var out map[string]any
	if err := jsonDecode(resp, &out); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if out["fallback"] != float64(99) {
		t.Errorf("fallback = %v, want 99", out["fallback"])
	}
	if out["tenant"] != float64(42) {
		t.Errorf("tenant = %v, want 42", out["tenant"])
	}
}

func TestForceTenantFilter(t *testing.T) {
	app := fiber.New()
	app.Get("/force", func(c fiber.Ctx) error {
		q := &query.Query{}
		if err := ForceTenantFilter(c, q); err != nil {
			return CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
		}
		return c.JSON(fiber.Map{"filters": len(q.Filters)})
	})
	app.Get("/force-with-tenant", func(c fiber.Ctx) error {
		c.Locals(LocalOrganizationID, uint64(9))
		q := &query.Query{}
		if err := ForceTenantFilter(c, q); err != nil {
			return CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
		}
		if len(q.Filters) != 1 || q.Filters[0].Value != uint64(9) {
			t.Errorf("filters = %+v, want organization_id = 9", q.Filters)
		}
		return c.JSON(fiber.Map{"filters": len(q.Filters)})
	})

	if resp := doTestRequest(app, http.MethodGet, "/force"); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
	if resp := doTestRequest(app, http.MethodGet, "/force-with-tenant"); resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOwnsTenant(t *testing.T) {
	app := fiber.New()
	app.Get("/owns", func(c fiber.Ctx) error {
		orgID := uint64(42)
		noTenant := OwnsTenant(c, &orgID)

		c.Locals(LocalOrganizationID, uint64(42))
		withOwner := OwnsTenant(c, &orgID)
		mismatch := OwnsTenant(c, helper.Ptr(uint64(43)))
		nilOrg := OwnsTenant(c, nil)

		return c.JSON(fiber.Map{
			"no_tenant":  noTenant,
			"with_owner": withOwner,
			"mismatch":   mismatch,
			"nil_org":    nilOrg,
		})
	})

	resp := doTestRequest(app, http.MethodGet, "/owns")
	var out map[string]any
	if err := jsonDecode(resp, &out); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if out["no_tenant"] != false || out["with_owner"] != true || out["mismatch"] != false || out["nil_org"] != false {
		t.Errorf("OwnsTenant results = %v", out)
	}
}

func TestForceTenantFilter_FailsClosedWithoutTenant(t *testing.T) {
	app := fiber.New()
	app.Get("/force", func(c fiber.Ctx) error {
		q := &query.Query{}
		if err := ForceTenantFilter(c, q); err != nil {
			return CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
		}
		return c.JSON(fiber.Map{"filters": len(q.Filters)})
	})

	resp := doTestRequest(app, http.MethodGet, "/force")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 (must not leak rows across tenants)", resp.StatusCode)
	}
}

func TestOwnsTenant_FailsClosedWithoutTenant(t *testing.T) {
	app := fiber.New()
	app.Get("/owns", func(c fiber.Ctx) error {
		orgID := uint64(42)
		return c.JSON(fiber.Map{"owns": OwnsTenant(c, &orgID)})
	})

	resp := doTestRequest(app, http.MethodGet, "/owns")
	var out map[string]any
	if err := jsonDecode(resp, &out); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if out["owns"] != false {
		t.Errorf("OwnsTenant without caller tenant = %v, want false", out["owns"])
	}
}

func TestParseQueryParams(t *testing.T) {
	allow := map[string]struct{}{"name": {}, "status": {}, "created_at": {}}

	app := fiber.New()
	app.Get("/q", func(c fiber.Ctx) error {
		parsed, err := ParseQueryParams(c, allow)
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString(err.Error())
		}
		return c.JSON(fiber.Map{"page": parsed.Pagination.Page, "size": parsed.Pagination.Size})
	})
	app.Get("/bad-page", func(c fiber.Ctx) error {
		_, err := ParseQueryParams(c, allow)
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString(err.Error())
		}
		return c.SendStatus(http.StatusOK)
	})
	app.Get("/full", func(c fiber.Ctx) error {
		parsed, err := ParseQueryParams(c, allow)
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString(err.Error())
		}
		return c.JSON(fiber.Map{
			"page":    parsed.Pagination.Page,
			"size":    parsed.Pagination.Size,
			"sorts":   len(parsed.Sorts),
			"filters": len(parsed.Filters),
			"sort0":   parsed.Sorts[0].Field + ":" + string(parsed.Sorts[0].Direction),
			"filter0": parsed.Filters[0].Field + ":" + string(parsed.Filters[0].Operator),
			"filter1": parsed.Filters[1].Field + ":" + string(parsed.Filters[1].Operator),
		})
	})

	resp := doTestRequest(app, http.MethodGet, "/q?page=3&size=200")
	var out map[string]any
	if err := jsonDecode(resp, &out); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if out["page"] != float64(3) || out["size"] != float64(maxQuerySize) {
		t.Errorf("pagination = %v, want page 3 size 100", out)
	}

	if resp := doTestRequest(app, http.MethodGet, "/q?page=0"); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("page=0 status = %d, want 400", resp.StatusCode)
	}
	if resp := doTestRequest(app, http.MethodGet, "/q?size=abc"); resp.StatusCode != http.StatusBadRequest {
		t.Errorf("size=abc status = %d, want 400", resp.StatusCode)
	}

	resp = doTestRequest(app, http.MethodGet, "/full?sort=name:desc,status&filter=name:eq:alice&filter=status:in:active,blocked")
	if err := jsonDecode(resp, &out); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if out["sorts"] != float64(2) || out["filters"] != float64(2) {
		t.Errorf("sorts/filters = %v/%v, want 2/2", out["sorts"], out["filters"])
	}
	if out["sort0"] != "name:DESC" || out["filter0"] != "name:=" || out["filter1"] != "status:IN" {
		t.Errorf("sort/filter details = %v", out)
	}
}

func TestParseQueryParamsErrors(t *testing.T) {
	allow := map[string]struct{}{"name": {}, "status": {}}

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		_, err := ParseQueryParams(c, allow)
		if err != nil {
			return c.Status(http.StatusBadRequest).SendString(err.Error())
		}
		return c.SendStatus(http.StatusOK)
	})

	cases := []string{
		"/?sort=unknown",
		"/?sort=name:bogus",
		"/?filter=name",
		"/?filter=unknown:eq:x",
		"/?filter=name:eq",
		"/?filter=name:bogus:x",
	}
	for _, u := range cases {
		if resp := doTestRequest(app, http.MethodGet, u); resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s status = %d, want 400", u, resp.StatusCode)
		}
	}
}

func TestBuildListMeta(t *testing.T) {
	q := &query.Query{
		Pagination: &query.Pagination{Page: 2, Size: 10},
		Sorts:      []query.Sort{{Field: "name", Direction: query.Descending}},
		Filters:    []query.Filter{{Field: "status", Operator: query.Equal, Value: "active"}, {Field: "tags", Operator: query.In, Value: []string{"a", "b"}}},
	}
	meta := BuildListMeta(q, 25)
	if meta.Pagination.Page != 2 || meta.Pagination.PerPage != 10 || meta.Pagination.Total != 25 || meta.Pagination.TotalPages != 3 {
		t.Errorf("pagination = %+v", meta.Pagination)
	}
	if meta.Sort.Field != "name" || meta.Sort.Direction != "desc" {
		t.Errorf("sort = %+v", meta.Sort)
	}
	if meta.Filter["status"] != "active" || meta.Filter["tags"] != "a,b" {
		t.Errorf("filters = %+v", meta.Filter)
	}
}

func TestSetLinkBaseURLAndLinkBaseFor(t *testing.T) {
	SetLinkBaseURL("")
	app := fiber.New()
	app.Get("/link", func(c fiber.Ctx) error {
		return c.SendString(LinkBaseFor(c))
	})
	req := httptest.NewRequest(http.MethodGet, "/link", nil)
	req.Host = "api.example.com"
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request error = %v", err)
	}
	_ = resp.Body.Close()

	SetLinkBaseURL("http://proxy.example.com/")
	if got := LinkBaseFor(nil); got != "http://proxy.example.com" {
		t.Errorf("LinkBaseFor() = %q, want trimmed base", got)
	}
}

func TestPaginationLinks(t *testing.T) {
	SetLinkBaseURL("")
	app := fiber.New()
	app.Get("/list", func(c fiber.Ctx) error {
		links := paginationLinks(c, &PaginationMeta{Page: 2, PerPage: 10, Total: 25, TotalPages: 3})
		return c.JSON(fiber.Map{"links": len(links)})
	})

	resp := doTestRequest(app, http.MethodGet, "/list?page=2&size=10")
	var out map[string]any
	if err := jsonDecode(resp, &out); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if out["links"] != float64(4) {
		t.Errorf("links = %v, want 4", out["links"])
	}

	if got := paginationLinks(nil, &PaginationMeta{Page: 1, TotalPages: 1}); got != nil {
		t.Errorf("paginationLinks() single page = %v, want nil", got)
	}
}

func jsonDecode(resp *http.Response, out any) error {
	defer func() { _ = resp.Body.Close() }()
	return json.NewDecoder(resp.Body).Decode(out)
}
