//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestProjectDeliveryLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	projectID := createProject(t, e, f)

	task := f.authed(t, http.MethodPost, "/projects/"+itoa(projectID)+"/tasks").
		WithJSON(map[string]any{"name": "E2E " + gofakeit.Word()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("task").
		Object()

	_ = uint64(task.Value("id").Number().Raw())

	milestone := f.authed(t, http.MethodPost, "/projects/"+itoa(projectID)+"/milestones").
		WithJSON(map[string]any{"name": "E2E " + gofakeit.Word()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("milestone").
		Object()

	milestoneID := uint64(milestone.Value("id").Number().Raw())

	f.authed(t, http.MethodPut, "/projects/"+itoa(projectID)+"/milestones/"+itoa(milestoneID)+"/reached").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/projects/"+itoa(projectID)+"/summary").
		Expect().
		Status(http.StatusOK)
}

func TestExpenseReportLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	employeeID := createEmployee(t, e, f, 0)
	categoryID := createExpenseCategory(t, e, f)

	report := f.authed(t, http.MethodPost, "/expense-reports").
		WithJSON(map[string]any{
			"name":         "E2E " + gofakeit.Word(),
			"employee_id":  employeeID,
			"payment_mode": "reimburse",
			"lines": []map[string]any{
				{
					"category_id":  categoryID,
					"expense_date": "2026-06-01",
					"unit_price":   500000,
					"quantity":     2,
				},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("report").
		Object()

	reportID := uint64(report.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/expense-reports/"+itoa(reportID)+"/submit").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/expense-reports/"+itoa(reportID)+"/approve").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/expense-reports/"+itoa(reportID)+"/post").
		WithJSON(map[string]any{"journal_id": f.journalGeneralID}).
		Expect().
		Status(http.StatusOK)
}
