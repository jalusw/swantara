package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/expense"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/project"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

var expenseCategoryColumns = []string{"id", "created_at", "updated_at", "deleted_at", "organization_id", "name", "expense_account_id", "default_tax_ids"}

func timeNow() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}

func expenseCategoryTestApp(t *testing.T, db *gorm.DB) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	base := dao.NewBase[reference.ExpenseCategory](db)
	h := NewExpenseCategoryHandler(expense.NewExpenseCategoryService(base))
	h.Register(app, passthroughGuards())
	return app
}

func expenseCategoryTestAppNoTenant(t *testing.T, db *gorm.DB) *fiber.App {
	t.Helper()
	app := fiber.New()
	base := dao.NewBase[reference.ExpenseCategory](db)
	h := NewExpenseCategoryHandler(expense.NewExpenseCategoryService(base))
	h.Register(app, passthroughGuards())
	return app
}

func expenseReportTestApp(t *testing.T, svc expense.ExpenseService) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(model.ActorKey, uint64(5))
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewExpenseReportHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func expenseReportTestAppNoTenant(t *testing.T, svc expense.ExpenseService) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewExpenseReportHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func expenseTestService(
	reports expense.ExpenseReportDAOMock,
	lines expense.ExpenseLineDAOMock,
) expense.ExpenseService {
	return expense.NewExpenseService(
		reports,
		lines,
		dao.CRUDMock[reference.ExpenseCategory]{},
		dao.CRUDMock[reference.Tax]{},
		expense.PosterMock{},
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	).SetBilling(expense.InvoiceEngineMock{}, expense.IncomeAccountResolverMock{}, expense.ProjectLookupMock{})
}

func expenseReportWithState(state string) *expense.ExpenseReport {
	return &expense.ExpenseReport{
		Base:           model.Base{ID: 1},
		OrganizationID: helper.Ptr(uint64(10)),
		Name:           "Travel 2026",
		EmployeeID:     2,
		State:          state,
		PaymentMode:    expense.ExpensePaymentOwnAccount,
	}
}

func TestExpenseCategoryHandler_List_ReturnsCategories(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "expense_categories" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(expenseCategoryColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Travel", uint64(100), nil))
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/expense-categories/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/expense-categories/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := expenseCategoryTestAppNoTenant(t, db)

	resp, err := doRequest(app, http.MethodGet, "/expense-categories/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "expense_categories" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnError(errors.New("db down"))
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/expense-categories/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_List_ExportsCSV(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "expense_categories" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(expenseCategoryColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Travel", uint64(100), nil))
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/expense-categories/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Get_ReturnsCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(expenseCategoryColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Travel", uint64(100), nil))
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/expense-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/expense-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Get_ReturnsNotFoundForForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(expenseCategoryColumns).AddRow(1, timeNow(), timeNow(), nil, 99, "Travel", uint64(100), nil))
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/expense-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Get_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/expense-categories/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/expense-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Create_CreatesCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "expense_categories"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := expenseCategoryTestApp(t, db)

	body := `{"name":"Travel","expense_account_id":100,"default_tax_ids":[1,2]}`
	resp, err := doRequest(app, http.MethodPost, "/expense-categories/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Create_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := expenseCategoryTestApp(t, db)

	body := `{"name":"","expense_account_id":100}`
	resp, err := doRequest(app, http.MethodPost, "/expense-categories/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Create_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := expenseCategoryTestAppNoTenant(t, db)

	body := `{"name":"Travel","expense_account_id":100}`
	resp, err := doRequest(app, http.MethodPost, "/expense-categories/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Create_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "expense_categories"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := expenseCategoryTestApp(t, db)

	body := `{"name":"Travel","expense_account_id":100}`
	resp, err := doRequest(app, http.MethodPost, "/expense-categories/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Update_UpdatesCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(expenseCategoryColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Travel", uint64(100), nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "expense_categories" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := expenseCategoryTestApp(t, db)

	body := `{"name":"Travel Plus","expense_account_id":200}`
	resp, err := doRequest(app, http.MethodPut, "/expense-categories/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Update_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	app := expenseCategoryTestApp(t, db)

	body := `{"name":"Travel Plus","expense_account_id":200}`
	resp, err := doRequest(app, http.MethodPut, "/expense-categories/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Update_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := expenseCategoryTestApp(t, db)

	body := `{"name":"Travel Plus","expense_account_id":200}`
	resp, err := doRequest(app, http.MethodPut, "/expense-categories/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Update_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(expenseCategoryColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Travel", uint64(100), nil))
	app := expenseCategoryTestApp(t, db)

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/expense-categories/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Update_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(expenseCategoryColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Travel", uint64(100), nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "expense_categories" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := expenseCategoryTestApp(t, db)

	body := `{"name":"Travel Plus","expense_account_id":200}`
	resp, err := doRequest(app, http.MethodPut, "/expense-categories/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Update_ReturnsServerErrorOnLookup(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := expenseCategoryTestApp(t, db)

	body := `{"name":"Travel Plus","expense_account_id":200}`
	resp, err := doRequest(app, http.MethodPut, "/expense-categories/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Delete_DeletesCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(expenseCategoryColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Travel", uint64(100), nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "expense_categories" WHERE "expense_categories"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodDelete, "/expense-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Delete_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodDelete, "/expense-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Delete_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodDelete, "/expense-categories/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Delete_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(expenseCategoryColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Travel", uint64(100), nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "expense_categories" WHERE "expense_categories"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodDelete, "/expense-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseCategoryHandler_Delete_ReturnsServerErrorOnLookup(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_categories" WHERE id = $1 ORDER BY "expense_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := expenseCategoryTestApp(t, db)

	resp, err := doRequest(app, http.MethodDelete, "/expense-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseReportHandler_List_ReturnsReports(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		CRUDMock: dao.CRUDMock[expense.ExpenseReport]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[expense.ExpenseReport], error) {
				return &query.Page[expense.ExpenseReport]{Items: []*expense.ExpenseReport{expenseReportWithState(expense.ExpenseStateDraft)}, Count: 1}, nil
			},
		},
	}
	svc := expenseTestService(reports, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/expense-reports/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestExpenseReportHandler_List_ReturnsServerError(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		CRUDMock: dao.CRUDMock[expense.ExpenseReport]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[expense.ExpenseReport], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := expenseTestService(reports, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/expense-reports/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestExpenseReportHandler_List_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := expenseTestService(expense.ExpenseReportDAOMock{}, expense.ExpenseLineDAOMock{})
	app := expenseReportTestAppNoTenant(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/expense-reports/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Get_ReturnsReport(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateDraft), nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return []*expense.ExpenseLine{{Base: model.Base{ID: 1}, CategoryID: helper.Ptr(uint64(1)), Description: helper.Ptr("dinner"), Quantity: 1, UnitPrice: 100, Amount: 100}}, nil
		},
	}
	svc := expenseTestService(reports, lines)
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/expense-reports/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Get_ReturnsNotFound(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return nil, nil
		},
	}
	svc := expenseTestService(reports, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/expense-reports/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := expenseTestService(expense.ExpenseReportDAOMock{}, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/expense-reports/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Get_ReturnsServerError(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return nil, errors.New("db down")
		},
	}
	svc := expenseTestService(reports, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/expense-reports/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Get_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := expenseTestService(expense.ExpenseReportDAOMock{}, expense.ExpenseLineDAOMock{})
	app := expenseReportTestAppNoTenant(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/expense-reports/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Get_ReturnsServerErrorOnLines(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateDraft), nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return nil, errors.New("db down")
		},
	}
	svc := expenseTestService(reports, lines)
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/expense-reports/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Create_CreatesReport(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, report *expense.ExpenseReport) (*expense.ExpenseReport, error) {
			report.ID = 1
			return report, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, report *expense.ExpenseReport) (*expense.ExpenseReport, error) {
			return report, nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, line *expense.ExpenseLine) (*expense.ExpenseLine, error) {
			line.ID = 1
			return line, nil
		},
	}
	svc := expenseTestService(reports, lines)
	app := expenseReportTestApp(t, svc)

	body := `{"name":"Travel 2026","employee_id":2,"payment_mode":"own_account","lines":[{"category_id":1,"expense_date":"2026-01-15","quantity":1,"unit_price":100}]}`
	resp, err := doRequest(app, http.MethodPost, "/expense-reports/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Create_RejectsValidation(t *testing.T) {
	svc := expenseTestService(expense.ExpenseReportDAOMock{}, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	body := `{"name":"","employee_id":2}`
	resp, err := doRequest(app, http.MethodPost, "/expense-reports/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Create_RejectsInvalidDate(t *testing.T) {
	svc := expenseTestService(expense.ExpenseReportDAOMock{}, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	body := `{"name":"Travel 2026","employee_id":2,"payment_mode":"own_account","lines":[{"category_id":1,"expense_date":"not-a-date","unit_price":100}]}`
	resp, err := doRequest(app, http.MethodPost, "/expense-reports/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Create_MapsInvalidPaymentMode(t *testing.T) {
	svc := expenseTestService(expense.ExpenseReportDAOMock{}, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	body := `{"name":"Travel 2026","employee_id":2,"payment_mode":"bogus","lines":[{"category_id":1,"expense_date":"2026-01-15","unit_price":100}]}`
	resp, err := doRequest(app, http.MethodPost, "/expense-reports/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Create_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := expenseTestService(expense.ExpenseReportDAOMock{}, expense.ExpenseLineDAOMock{})
	app := expenseReportTestAppNoTenant(t, svc)

	body := `{"name":"Travel 2026","employee_id":2,"payment_mode":"own_account","lines":[{"category_id":1,"expense_date":"2026-01-15","unit_price":100}]}`
	resp, err := doRequest(app, http.MethodPost, "/expense-reports/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Create_ReturnsServerError(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, report *expense.ExpenseReport) (*expense.ExpenseReport, error) {
			return nil, errors.New("db down")
		},
	}
	svc := expenseTestService(reports, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	body := `{"name":"Travel 2026","employee_id":2,"payment_mode":"own_account","lines":[{"category_id":1,"expense_date":"2026-01-15","unit_price":100}]}`
	resp, err := doRequest(app, http.MethodPost, "/expense-reports/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Create_AllowsExplicitReimbursable(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, report *expense.ExpenseReport) (*expense.ExpenseReport, error) {
			report.ID = 1
			return report, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, report *expense.ExpenseReport) (*expense.ExpenseReport, error) {
			return report, nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, line *expense.ExpenseLine) (*expense.ExpenseLine, error) {
			line.ID = 1
			return line, nil
		},
	}
	svc := expenseTestService(reports, lines)
	app := expenseReportTestApp(t, svc)

	body := `{"name":"Travel 2026","employee_id":2,"payment_mode":"own_account","lines":[{"category_id":1,"expense_date":"2026-01-15","quantity":1,"unit_price":100,"reimbursable":false}]}`
	resp, err := doRequest(app, http.MethodPost, "/expense-reports/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Submit_SubmitsReport(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateDraft), nil
		},
		CRUDMock: dao.CRUDMock[expense.ExpenseReport]{
			UpdateFunc: func(_ context.Context, report *expense.ExpenseReport) (*expense.ExpenseReport, error) {
				return report, nil
			},
		},
	}
	svc := expenseTestService(reports, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/submit", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Approve_ApprovesReport(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateSubmitted), nil
		},
		CRUDMock: dao.CRUDMock[expense.ExpenseReport]{
			UpdateFunc: func(_ context.Context, report *expense.ExpenseReport) (*expense.ExpenseReport, error) {
				return report, nil
			},
		},
	}
	svc := expenseTestService(reports, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/approve", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Refuse_RefusesReport(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateSubmitted), nil
		},
		CRUDMock: dao.CRUDMock[expense.ExpenseReport]{
			UpdateFunc: func(_ context.Context, report *expense.ExpenseReport) (*expense.ExpenseReport, error) {
				return report, nil
			},
		},
	}
	svc := expenseTestService(reports, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/refuse", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Post_PostsReport(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateApproved), nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, report *expense.ExpenseReport) (*expense.ExpenseReport, error) {
			return report, nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return []*expense.ExpenseLine{{Base: model.Base{ID: 1}, CategoryID: helper.Ptr(uint64(1)), Description: helper.Ptr("dinner"), Quantity: 1, UnitPrice: 100, Amount: 100}}, nil
		},
	}
	categories := dao.CRUDMock[reference.ExpenseCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ExpenseCategory, error) {
			return &reference.ExpenseCategory{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), ExpenseAccountID: helper.Ptr(uint64(500))}, nil
		},
	}
	poster := expense.PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := expense.NewExpenseService(
		reports,
		lines,
		categories,
		dao.CRUDMock[reference.Tax]{},
		poster,
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	)
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/post", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Reimburse_ReimbursesReport(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStatePosted), nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, report *expense.ExpenseReport) (*expense.ExpenseReport, error) {
			return report, nil
		},
	}
	poster := expense.PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := expense.NewExpenseService(
		reports,
		expense.ExpenseLineDAOMock{},
		dao.CRUDMock[reference.ExpenseCategory]{},
		dao.CRUDMock[reference.Tax]{},
		poster,
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	)
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/reimburse", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Transition_ReturnsNotFound(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return nil, nil
		},
	}
	svc := expenseTestService(reports, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/submit", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Transition_ReturnsConflictOnInvalidState(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateSubmitted), nil
		},
	}
	svc := expenseTestService(reports, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/submit", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Transition_RejectsInvalidID(t *testing.T) {
	svc := expenseTestService(expense.ExpenseReportDAOMock{}, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/abc/submit", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Transition_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := expenseTestService(expense.ExpenseReportDAOMock{}, expense.ExpenseLineDAOMock{})
	app := expenseReportTestAppNoTenant(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/submit", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Transition_ReturnsServerError(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return nil, errors.New("db down")
		},
	}
	svc := expenseTestService(reports, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/submit", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Post_MapsNoLines(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateApproved), nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return []*expense.ExpenseLine{}, nil
		},
	}
	svc := expenseTestService(reports, lines)
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/post", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Post_MapsNoCategory(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateApproved), nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return []*expense.ExpenseLine{{Base: model.Base{ID: 1}, Amount: 100}}, nil
		},
	}
	svc := expenseTestService(reports, lines)
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/post", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Post_MapsNoAccount(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateApproved), nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return []*expense.ExpenseLine{{Base: model.Base{ID: 1}, CategoryID: helper.Ptr(uint64(1)), Amount: 100}}, nil
		},
	}
	categories := dao.CRUDMock[reference.ExpenseCategory]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.ExpenseCategory, error) {
			return &reference.ExpenseCategory{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10))}, nil
		},
	}
	svc := expense.NewExpenseService(
		reports,
		lines,
		categories,
		dao.CRUDMock[reference.Tax]{},
		expense.PosterMock{},
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	)
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/post", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Post_MapsConfigMissing(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateApproved), nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return []*expense.ExpenseLine{{Base: model.Base{ID: 1}, CategoryID: helper.Ptr(uint64(1)), Amount: 100}}, nil
		},
	}
	configs := expense.ExpenseConfigSourceMock{
		JournalIDFunc: func(_ context.Context, _ uint64) (uint64, error) {
			return 0, nil
		},
	}
	svc := expense.NewExpenseService(
		reports,
		lines,
		dao.CRUDMock[reference.ExpenseCategory]{},
		dao.CRUDMock[reference.Tax]{},
		expense.PosterMock{},
		configs,
		expense.TransactionerMock{},
	)
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/post", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Post_ReturnsServerError(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateApproved), nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return nil, errors.New("db down")
		},
	}
	svc := expenseTestService(reports, lines)
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/post", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Bill_BillsReport(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStatePosted), nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return []*expense.ExpenseLine{{Base: model.Base{ID: 1}, ProjectID: helper.Ptr(uint64(5)), ItemID: helper.Ptr(uint64(7)), Quantity: 1, UnitPrice: 100, Reimbursable: false}}, nil
		},
	}
	invoices := expense.InvoiceEngineMock{
		CreateFunc: func(_ context.Context, _ accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 1}}, nil
		},
	}
	projects := expense.ProjectLookupMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*project.Project, error) {
			return &project.Project{Base: model.Base{ID: 5}, ContactID: 7}, nil
		},
	}
	svc := expense.NewExpenseService(
		reports,
		lines,
		dao.CRUDMock[reference.ExpenseCategory]{},
		dao.CRUDMock[reference.Tax]{},
		expense.PosterMock{},
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	).SetBilling(invoices, expense.IncomeAccountResolverMock{}, projects)
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/bill", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Bill_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := expenseTestService(expense.ExpenseReportDAOMock{}, expense.ExpenseLineDAOMock{})
	app := expenseReportTestAppNoTenant(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/bill", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Bill_MapsConfigMissing(t *testing.T) {
	svc := expense.NewExpenseService(
		expense.ExpenseReportDAOMock{},
		expense.ExpenseLineDAOMock{},
		dao.CRUDMock[reference.ExpenseCategory]{},
		dao.CRUDMock[reference.Tax]{},
		expense.PosterMock{},
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	)
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/bill", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Bill_MapsNoBillableLines(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStatePosted), nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return []*expense.ExpenseLine{{Base: model.Base{ID: 1}, Reimbursable: true}}, nil
		},
	}
	svc := expense.NewExpenseService(
		reports,
		lines,
		dao.CRUDMock[reference.ExpenseCategory]{},
		dao.CRUDMock[reference.Tax]{},
		expense.PosterMock{},
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	).SetBilling(expense.InvoiceEngineMock{}, expense.IncomeAccountResolverMock{}, expense.ProjectLookupMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/bill", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Bill_MapsMissingProject(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStatePosted), nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return []*expense.ExpenseLine{{Base: model.Base{ID: 1}, Reimbursable: false}}, nil
		},
	}
	svc := expense.NewExpenseService(
		reports,
		lines,
		dao.CRUDMock[reference.ExpenseCategory]{},
		dao.CRUDMock[reference.Tax]{},
		expense.PosterMock{},
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	).SetBilling(expense.InvoiceEngineMock{}, expense.IncomeAccountResolverMock{}, expense.ProjectLookupMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/bill", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Bill_MapsProjectNotFound(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStatePosted), nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return []*expense.ExpenseLine{{Base: model.Base{ID: 1}, ProjectID: helper.Ptr(uint64(5)), Reimbursable: false}}, nil
		},
	}
	projects := expense.ProjectLookupMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*project.Project, error) {
			return nil, nil
		},
	}
	svc := expense.NewExpenseService(
		reports,
		lines,
		dao.CRUDMock[reference.ExpenseCategory]{},
		dao.CRUDMock[reference.Tax]{},
		expense.PosterMock{},
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	).SetBilling(expense.InvoiceEngineMock{}, expense.IncomeAccountResolverMock{}, projects)
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/bill", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Bill_MapsNoIncomeAccount(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStatePosted), nil
		},
	}
	lines := expense.ExpenseLineDAOMock{
		ListByReportFunc: func(_ context.Context, _ uint64) ([]*expense.ExpenseLine, error) {
			return []*expense.ExpenseLine{{Base: model.Base{ID: 1}, ProjectID: helper.Ptr(uint64(5)), Reimbursable: false}}, nil
		},
	}
	svc := expense.NewExpenseService(
		reports,
		lines,
		dao.CRUDMock[reference.ExpenseCategory]{},
		dao.CRUDMock[reference.Tax]{},
		expense.PosterMock{},
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	).SetBilling(expense.InvoiceEngineMock{}, expense.IncomeAccountResolverMock{}, expense.ProjectLookupMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/bill", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Bill_RejectsInvalidID(t *testing.T) {
	svc := expenseTestService(expense.ExpenseReportDAOMock{}, expense.ExpenseLineDAOMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/abc/bill", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Bill_ReturnsNotFound(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return nil, nil
		},
	}
	svc := expense.NewExpenseService(
		reports,
		expense.ExpenseLineDAOMock{},
		dao.CRUDMock[reference.ExpenseCategory]{},
		dao.CRUDMock[reference.Tax]{},
		expense.PosterMock{},
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	).SetBilling(expense.InvoiceEngineMock{}, expense.IncomeAccountResolverMock{}, expense.ProjectLookupMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/bill", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Bill_ReturnsConflictOnInvalidState(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return expenseReportWithState(expense.ExpenseStateDraft), nil
		},
	}
	svc := expense.NewExpenseService(
		reports,
		expense.ExpenseLineDAOMock{},
		dao.CRUDMock[reference.ExpenseCategory]{},
		dao.CRUDMock[reference.Tax]{},
		expense.PosterMock{},
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	).SetBilling(expense.InvoiceEngineMock{}, expense.IncomeAccountResolverMock{}, expense.ProjectLookupMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/bill", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestExpenseReportHandler_Bill_ReturnsServerError(t *testing.T) {
	reports := expense.ExpenseReportDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*expense.ExpenseReport, error) {
			return nil, errors.New("db down")
		},
	}
	svc := expense.NewExpenseService(
		reports,
		expense.ExpenseLineDAOMock{},
		dao.CRUDMock[reference.ExpenseCategory]{},
		dao.CRUDMock[reference.Tax]{},
		expense.PosterMock{},
		expense.ExpenseConfigSourceMock{},
		expense.TransactionerMock{},
	).SetBilling(expense.InvoiceEngineMock{}, expense.IncomeAccountResolverMock{}, expense.ProjectLookupMock{})
	app := expenseReportTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/expense-reports/1/bill", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func passthroughGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			return c.Next()
		},
		Guard: func(_, _ string) fiber.Handler {
			return func(c fiber.Ctx) error {
				return c.Next()
			}
		},
	}
}

func doRequest(app *fiber.App, method, path, body string) (*http.Response, error) {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	return app.Test(req)
}
