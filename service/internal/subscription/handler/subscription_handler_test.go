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
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
	"gorm.io/gorm"
)

func subscriptionTestService(
	subscriptions subscription.SubscriptionDAOMock,
	lines subscription.SubscriptionLineDAOMock,
	plans dao.CRUDMock[reference.SubscriptionPlan],
) subscription.SubscriptionService {
	return subscription.NewSubscriptionService(
		subscriptions,
		lines,
		plans,
		testContactDAO(),
		testPriceBookDAO(),
		subscription.IncomeAccountResolverMock{},
		subscription.InvoiceEngineMock{},
		accounting.DeferralService{},
		subscription.SubscriptionConfigSourceMock{},
		subscription.TransactionerMock{},
	)
}

func testContactDAO() contacts.ContactDAOMock {
	return contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return &contacts.Contact{Base: model.Base{ID: 7}, OrganizationID: helperUint64Ptr(10)}, nil
			},
		},
	}
}

func testPriceBookDAO() products.PriceBookDAOMock {
	return products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return &products.PriceBook{Base: model.Base{ID: 3}, OrganizationID: helperUint64Ptr(10)}, nil
			},
		},
	}
}

func helperUint64Ptr(value uint64) *uint64 {
	return &value
}

var planColumns = []string{"id", "created_at", "updated_at", "deleted_at", "organization_id", "name", "recurring_interval", "recurring_count"}

func timeNow() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}

func sampleSubscription() *subscription.Subscription {
	return &subscription.Subscription{
		Base:           model.Base{ID: 1},
		OrganizationID: helperUint64Ptr(10),
		Name:           "Pro",
		ContactID:      helperUint64Ptr(7),
		PlanID:         helperUint64Ptr(1),
		PriceBookID:    helperUint64Ptr(3),
		CurrencyCode:   helperStringPtr("USD"),
		State:          subscription.SubscriptionStateDraft,
		MRR:            120,
	}
}

func helperStringPtr(value string) *string {
	return &value
}

func subscriptionPlanTestApp(
	t *testing.T,
	db *gorm.DB,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewSubscriptionPlanHandler(subscription.NewSubscriptionPlanService(dao.NewBase[reference.SubscriptionPlan](db)))
	h.Register(app, passthroughGuards())
	return app
}

func subscriptionTestApp(
	t *testing.T,
	svc subscription.SubscriptionService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewSubscriptionHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func subscriptionTestAppNoTenant(
	t *testing.T,
	svc subscription.SubscriptionService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewSubscriptionHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func TestSubscriptionPlanHandler_List_ReturnsPlans(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "subscription_plans" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_plans" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(planColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Pro", "month", 1))
	app := subscriptionPlanTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/subscription-plans/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := subscriptionPlanTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/subscription-plans/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "subscription_plans" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnError(errors.New("db down"))
	app := subscriptionPlanTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/subscription-plans/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Get_ReturnsPlan(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_plans" WHERE id = $1 ORDER BY "subscription_plans"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(planColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Pro", "month", 1))
	app := subscriptionPlanTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/subscription-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_plans" WHERE id = $1 ORDER BY "subscription_plans"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	app := subscriptionPlanTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/subscription-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Get_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_plans" WHERE id = $1 ORDER BY "subscription_plans"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(planColumns).AddRow(1, timeNow(), timeNow(), nil, 99, "Pro", "month", 1))
	app := subscriptionPlanTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/subscription-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Get_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := subscriptionPlanTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/subscription-plans/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_plans" WHERE id = $1 ORDER BY "subscription_plans"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := subscriptionPlanTestApp(t, db)

	resp, err := doRequest(app, http.MethodGet, "/subscription-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Create_CreatesPlan(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "subscription_plans"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := subscriptionPlanTestApp(t, db)

	body := `{"name":"Pro","recurring_interval":"month"}`
	resp, err := doRequest(app, http.MethodPost, "/subscription-plans/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Create_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := subscriptionPlanTestApp(t, db)

	body := `{"name":"","recurring_interval":""}`
	resp, err := doRequest(app, http.MethodPost, "/subscription-plans/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Create_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "subscription_plans"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := subscriptionPlanTestApp(t, db)

	body := `{"name":"Pro","recurring_interval":"month"}`
	resp, err := doRequest(app, http.MethodPost, "/subscription-plans/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Update_UpdatesPlan(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_plans" WHERE id = $1 ORDER BY "subscription_plans"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(planColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Pro", "month", 1))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "subscription_plans" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := subscriptionPlanTestApp(t, db)

	body := `{"name":"Pro Max","recurring_interval":"year","recurring_count":1}`
	resp, err := doRequest(app, http.MethodPut, "/subscription-plans/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Update_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_plans" WHERE id = $1 ORDER BY "subscription_plans"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	app := subscriptionPlanTestApp(t, db)

	body := `{"name":"Pro","recurring_interval":"month"}`
	resp, err := doRequest(app, http.MethodPut, "/subscription-plans/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Update_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := subscriptionPlanTestApp(t, db)

	body := `{"name":"Pro","recurring_interval":"month"}`
	resp, err := doRequest(app, http.MethodPut, "/subscription-plans/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Update_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_plans" WHERE id = $1 ORDER BY "subscription_plans"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(planColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Pro", "month", 1))
	app := subscriptionPlanTestApp(t, db)

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/subscription-plans/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Update_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_plans" WHERE id = $1 ORDER BY "subscription_plans"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(planColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Pro", "month", 1))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "subscription_plans" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := subscriptionPlanTestApp(t, db)

	body := `{"name":"Pro","recurring_interval":"month"}`
	resp, err := doRequest(app, http.MethodPut, "/subscription-plans/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Delete_DeletesPlan(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_plans" WHERE id = $1 ORDER BY "subscription_plans"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(planColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Pro", "month", 1))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "subscription_plans" WHERE "subscription_plans"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := subscriptionPlanTestApp(t, db)

	resp, err := doRequest(app, http.MethodDelete, "/subscription-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Delete_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_plans" WHERE id = $1 ORDER BY "subscription_plans"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	app := subscriptionPlanTestApp(t, db)

	resp, err := doRequest(app, http.MethodDelete, "/subscription-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Delete_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := subscriptionPlanTestApp(t, db)

	resp, err := doRequest(app, http.MethodDelete, "/subscription-plans/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionPlanHandler_Delete_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_plans" WHERE id = $1 ORDER BY "subscription_plans"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(planColumns).AddRow(1, timeNow(), timeNow(), nil, 10, "Pro", "month", 1))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "subscription_plans" WHERE "subscription_plans"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := subscriptionPlanTestApp(t, db)

	resp, err := doRequest(app, http.MethodDelete, "/subscription-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionHandler_List_ReturnsSubscriptions(t *testing.T) {
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[subscription.Subscription], error) {
				return &query.Page[subscription.Subscription]{Items: []*subscription.Subscription{sampleSubscription()}, Count: 1}, nil
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/subscriptions/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSubscriptionHandler_List_ReturnsServerError(t *testing.T) {
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[subscription.Subscription], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/subscriptions/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSubscriptionHandler_List_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := subscriptionTestService(subscription.SubscriptionDAOMock{}, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestAppNoTenant(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/subscriptions/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Get_ReturnsSubscription(t *testing.T) {
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*subscription.Subscription, error) {
				return sampleSubscription(), nil
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/subscriptions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Get_ReturnsNotFound(t *testing.T) {
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*subscription.Subscription, error) {
				return nil, nil
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/subscriptions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := subscriptionTestService(subscription.SubscriptionDAOMock{}, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/subscriptions/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Get_ReturnsServerError(t *testing.T) {
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*subscription.Subscription, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/subscriptions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_CreatesSubscription(t *testing.T) {
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, OrganizationID: helperUint64Ptr(10), RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			CreateFunc: func(_ context.Context, entity *subscription.Subscription) (*subscription.Subscription, error) {
				entity.ID = 1
				return entity, nil
			},
		},
	}
	lines := subscription.SubscriptionLineDAOMock{
		CRUDMock: dao.CRUDMock[subscription.SubscriptionLine]{
			CreateFunc: func(_ context.Context, entity *subscription.SubscriptionLine) (*subscription.SubscriptionLine, error) {
				entity.ID = 2
				return entity, nil
			},
		},
	}
	svc := subscriptionTestService(subscriptions, lines, plans)
	app := subscriptionTestApp(t, svc)

	body := `{"name":"Pro","contact_id":7,"plan_id":1,"price_book_id":3,"currency_code":"USD","lines":[{"item_id":100,"qty":1,"unit_price":120}]}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_RejectsValidation(t *testing.T) {
	svc := subscriptionTestService(subscription.SubscriptionDAOMock{}, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := subscriptionTestService(subscription.SubscriptionDAOMock{}, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestAppNoTenant(t, svc)

	body := `{"name":"Pro","contact_id":7,"plan_id":1,"price_book_id":3,"currency_code":"USD","lines":[{"item_id":100,"qty":1}]}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_MapsPlanNotFound(t *testing.T) {
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return nil, nil
		},
	}
	svc := subscriptionTestService(subscription.SubscriptionDAOMock{}, subscription.SubscriptionLineDAOMock{}, plans)
	app := subscriptionTestApp(t, svc)

	body := `{"name":"Pro","contact_id":7,"plan_id":1,"price_book_id":3,"currency_code":"USD","lines":[{"item_id":100,"qty":1}]}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_MapsNoLines(t *testing.T) {
	svc := subscriptionTestService(subscription.SubscriptionDAOMock{}, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	body := `{"name":"Pro","contact_id":7,"plan_id":1,"price_book_id":3,"currency_code":"USD"}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_ReturnsServerError(t *testing.T) {
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, OrganizationID: helperUint64Ptr(10), RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			CreateFunc: func(_ context.Context, entity *subscription.Subscription) (*subscription.Subscription, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, plans)
	app := subscriptionTestApp(t, svc)

	body := `{"name":"Pro","contact_id":7,"plan_id":1,"price_book_id":3,"currency_code":"USD","lines":[{"item_id":100,"qty":1,"unit_price":120}]}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Metrics_ReturnsMetrics(t *testing.T) {
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[subscription.Subscription], error) {
				return &query.Page[subscription.Subscription]{Items: []*subscription.Subscription{{State: subscription.SubscriptionStateActive, MRR: 100}}}, nil
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/subscriptions/metrics", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Metrics_ReturnsServerError(t *testing.T) {
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[subscription.Subscription], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/subscriptions/metrics", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Activate_ActivatesSubscription(t *testing.T) {
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*subscription.Subscription, error) {
				return sampleSubscription(), nil
			},
			UpdateFunc: func(_ context.Context, entity *subscription.Subscription) (*subscription.Subscription, error) {
				return entity, nil
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/subscriptions/1/activate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Pause_ReturnsConflictOnInvalidState(t *testing.T) {
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*subscription.Subscription, error) {
				return sampleSubscription(), nil
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/subscriptions/1/pause", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Resume_ReturnsNotFound(t *testing.T) {
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*subscription.Subscription, error) {
				return nil, nil
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/subscriptions/1/resume", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Churn_RejectsInvalidID(t *testing.T) {
	svc := subscriptionTestService(subscription.SubscriptionDAOMock{}, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/subscriptions/abc/churn", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Close_ReturnsServerError(t *testing.T) {
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*subscription.Subscription, error) {
				return &subscription.Subscription{Base: model.Base{ID: 1}, OrganizationID: helperUint64Ptr(10), State: subscription.SubscriptionStateActive}, nil
			},
			UpdateFunc: func(_ context.Context, entity *subscription.Subscription) (*subscription.Subscription, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/subscriptions/1/close", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_MapsZeroQty(t *testing.T) {
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, OrganizationID: helperUint64Ptr(10), RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	svc := subscriptionTestService(subscription.SubscriptionDAOMock{}, subscription.SubscriptionLineDAOMock{}, plans)
	app := subscriptionTestApp(t, svc)

	body := `{"name":"Pro","contact_id":7,"plan_id":1,"price_book_id":3,"currency_code":"USD","lines":[{"item_id":100,"qty":0}]}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_MapsNegativePrice(t *testing.T) {
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, OrganizationID: helperUint64Ptr(10), RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	svc := subscriptionTestService(subscription.SubscriptionDAOMock{}, subscription.SubscriptionLineDAOMock{}, plans)
	app := subscriptionTestApp(t, svc)

	body := `{"name":"Pro","contact_id":7,"plan_id":1,"price_book_id":3,"currency_code":"USD","lines":[{"item_id":100,"qty":1,"unit_price":-5}]}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_MapsInvalidDiscount(t *testing.T) {
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, OrganizationID: helperUint64Ptr(10), RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	svc := subscriptionTestService(subscription.SubscriptionDAOMock{}, subscription.SubscriptionLineDAOMock{}, plans)
	app := subscriptionTestApp(t, svc)

	body := `{"name":"Pro","contact_id":7,"plan_id":1,"price_book_id":3,"currency_code":"USD","lines":[{"item_id":100,"qty":1,"unit_price":10,"discount_pct":150}]}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_MapsForeignPlan(t *testing.T) {
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, OrganizationID: helperUint64Ptr(99), RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	svc := subscriptionTestService(subscription.SubscriptionDAOMock{}, subscription.SubscriptionLineDAOMock{}, plans)
	app := subscriptionTestApp(t, svc)

	body := `{"name":"Pro","contact_id":7,"plan_id":1,"price_book_id":3,"currency_code":"USD","lines":[{"item_id":100,"qty":1}]}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_MapsForeignContact(t *testing.T) {
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, OrganizationID: helperUint64Ptr(10), RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	contactsDAO := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return &contacts.Contact{Base: model.Base{ID: 7}, OrganizationID: helperUint64Ptr(99)}, nil
			},
		},
	}
	svc := subscription.NewSubscriptionService(
		subscription.SubscriptionDAOMock{},
		subscription.SubscriptionLineDAOMock{},
		plans,
		contactsDAO,
		testPriceBookDAO(),
		subscription.IncomeAccountResolverMock{},
		subscription.InvoiceEngineMock{},
		accounting.DeferralService{},
		subscription.SubscriptionConfigSourceMock{},
		subscription.TransactionerMock{},
	)
	app := subscriptionTestApp(t, svc)

	body := `{"name":"Pro","contact_id":7,"plan_id":1,"price_book_id":3,"currency_code":"USD","lines":[{"item_id":100,"qty":1}]}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_MapsForeignPriceBook(t *testing.T) {
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, OrganizationID: helperUint64Ptr(10), RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	price_booksDAO := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return &products.PriceBook{Base: model.Base{ID: 3}, OrganizationID: helperUint64Ptr(99)}, nil
			},
		},
	}
	svc := subscription.NewSubscriptionService(
		subscription.SubscriptionDAOMock{},
		subscription.SubscriptionLineDAOMock{},
		plans,
		testContactDAO(),
		price_booksDAO,
		subscription.IncomeAccountResolverMock{},
		subscription.InvoiceEngineMock{},
		accounting.DeferralService{},
		subscription.SubscriptionConfigSourceMock{},
		subscription.TransactionerMock{},
	)
	app := subscriptionTestApp(t, svc)

	body := `{"name":"Pro","contact_id":7,"plan_id":1,"price_book_id":3,"currency_code":"USD","lines":[{"item_id":100,"qty":1}]}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Create_MapsForeignProduct(t *testing.T) {
	plans := dao.CRUDMock[reference.SubscriptionPlan]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SubscriptionPlan, error) {
			return &reference.SubscriptionPlan{Base: model.Base{ID: 1}, OrganizationID: helperUint64Ptr(10), RecurringInterval: "month", RecurringCount: 1}, nil
		},
	}
	resolver := subscription.IncomeAccountResolverMock{
		ResolveVariantOrganizationFunc: func(_ context.Context, _ uint64) (*uint64, error) {
			return helperUint64Ptr(99), nil
		},
	}
	svc := subscription.NewSubscriptionService(
		subscription.SubscriptionDAOMock{},
		subscription.SubscriptionLineDAOMock{},
		plans,
		testContactDAO(),
		testPriceBookDAO(),
		resolver,
		subscription.InvoiceEngineMock{},
		accounting.DeferralService{},
		subscription.SubscriptionConfigSourceMock{},
		subscription.TransactionerMock{},
	)
	app := subscriptionTestApp(t, svc)

	body := `{"name":"Pro","contact_id":7,"plan_id":1,"price_book_id":3,"currency_code":"USD","lines":[{"item_id":100,"qty":1}]}`
	resp, err := doRequest(app, http.MethodPost, "/subscriptions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSubscriptionHandler_Churn_ChurnsSubscription(t *testing.T) {
	subscriptions := subscription.SubscriptionDAOMock{
		CRUDMock: dao.CRUDMock[subscription.Subscription]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*subscription.Subscription, error) {
				return &subscription.Subscription{Base: model.Base{ID: 1}, OrganizationID: helperUint64Ptr(10), State: subscription.SubscriptionStateActive}, nil
			},
			UpdateFunc: func(_ context.Context, entity *subscription.Subscription) (*subscription.Subscription, error) {
				return entity, nil
			},
		},
	}
	svc := subscriptionTestService(subscriptions, subscription.SubscriptionLineDAOMock{}, dao.CRUDMock[reference.SubscriptionPlan]{})
	app := subscriptionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/subscriptions/1/churn", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
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
