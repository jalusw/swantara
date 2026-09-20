package httpx

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

type csvRow struct {
	Name    string    `json:"name"`
	Active  bool      `json:"active"`
	Count   int       `json:"count"`
	Total   float64   `json:"total"`
	Note    *string   `json:"note"`
	When    time.Time `json:"when"`
	Skipped chan int  `json:"skipped"`
	NoTag   string
	Nested  csvNested `json:"nested"`
}

type csvNested struct {
	Inner string `json:"inner"`
}

func TestExportCSV_AllKinds(t *testing.T) {
	note := "hi"
	rows := []csvRow{
		{Name: "a", Active: true, Count: 3, Total: 1.5, Note: &note, When: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "b"},
	}

	app := fiber.New()
	app.Get("/csv", func(c fiber.Ctx) error {
		return ExportCSV(c, http.StatusOK, "rows.csv", rows)
	})
	resp := doTestRequest(app, http.MethodGet, "/csv")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d", resp.StatusCode)
	}

	ptrs := []*csvRow{&rows[0], nil}
	app.Get("/csvp", func(c fiber.Ctx) error {
		return ExportCSV(c, http.StatusOK, "rows.csv", ptrs)
	})
	resp = doTestRequest(app, http.MethodGet, "/csvp")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d", resp.StatusCode)
	}
}
