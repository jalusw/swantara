package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func contactsHandlerTest(t *testing.T, contactDAO contacts.ContactDAOMock, _ contacts.ContactService) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		c.Locals(model.ActorKey, uint64(5))
		return c.Next()
	})
	svc := contactsTestSvc(contactDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	h := NewContactHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func contactsHandlerTestNoTenant(t *testing.T, contactDAO contacts.ContactDAOMock, _ contacts.ContactService) *fiber.App {
	t.Helper()
	app := fiber.New()
	svc := contactsTestSvc(contactDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	h := NewContactHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func contactRelationHandlerTest(
	t *testing.T,
	contactDAO contacts.ContactDAOMock,
	addresses contacts.ContactAddressDAOMock,
	banks contacts.ContactBankAccountDAOMock,
	customers contacts.CustomerProfileDAOMock,
	suppliers contacts.SupplierProfileDAOMock,
	_ contacts.ContactService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		c.Locals(model.ActorKey, uint64(5))
		return c.Next()
	})
	svc := contactsTestSvc(contactDAO, addresses, banks, customers, suppliers)
	h := NewContactRelationHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func contactsTestSvc(
	contactDAO contacts.ContactDAOMock,
	addresses contacts.ContactAddressDAOMock,
	banks contacts.ContactBankAccountDAOMock,
	customers contacts.CustomerProfileDAOMock,
	suppliers contacts.SupplierProfileDAOMock,
) contacts.ContactService {
	return contacts.NewContactService(contactDAO, addresses, banks, customers, suppliers)
}

func emptyContactSvc() contacts.ContactService {
	return contactsTestSvc(
		contacts.ContactDAOMock{},
		contacts.ContactAddressDAOMock{},
		contacts.ContactBankAccountDAOMock{},
		contacts.CustomerProfileDAOMock{},
		contacts.SupplierProfileDAOMock{},
	)
}

func contactCRUD(crud dao.CRUDMock[contacts.Contact]) contacts.ContactDAOMock {
	return contacts.ContactDAOMock{CRUDMock: crud}
}

func addressCRUD(crud dao.CRUDMock[contacts.ContactAddress]) contacts.ContactAddressDAOMock {
	return contacts.ContactAddressDAOMock{CRUDMock: crud}
}

func bankCRUD(crud dao.CRUDMock[contacts.ContactBankAccount]) contacts.ContactBankAccountDAOMock {
	return contacts.ContactBankAccountDAOMock{CRUDMock: crud}
}

func customerCRUD(crud dao.CRUDMock[contacts.CustomerProfile]) contacts.CustomerProfileDAOMock {
	return contacts.CustomerProfileDAOMock{CRUDMock: crud}
}

func supplierCRUD(crud dao.CRUDMock[contacts.SupplierProfile]) contacts.SupplierProfileDAOMock {
	return contacts.SupplierProfileDAOMock{CRUDMock: crud}
}

func sampleContact() *contacts.Contact {
	return &contacts.Contact{
		Base:           model.Base{ID: 1},
		OrganizationID: helper.Ptr(uint64(10)),
		Name:           "Acme",
		DisplayName:    helper.Ptr("Acme Corp"),
		IsOrganization: true,
		Email:          helper.Ptr("billing@acme.com"),
		Lang:           "en",
		Active:         true,
	}
}

func foreignContact() *contacts.Contact {
	contact := sampleContact()
	contact.OrganizationID = helper.Ptr(uint64(99))
	return contact
}

func sampleAddress() *contacts.ContactAddress {
	return &contacts.ContactAddress{
		Base:      model.Base{ID: 1},
		ContactID: 1,
		Type:      helper.Ptr(contacts.AddressTypeBilling),
		Line1:     helper.Ptr("123 Main St"),
		City:      helper.Ptr("Jakarta"),
		IsDefault: true,
	}
}

func sampleBankAccount() *contacts.ContactBankAccount {
	return &contacts.ContactBankAccount{
		Base:          model.Base{ID: 1},
		ContactID:     1,
		AccountHolder: helper.Ptr("Acme"),
		BankName:      helper.Ptr("Bank X"),
		IBAN:          helper.Ptr("ID0001"),
	}
}

func sampleCustomer() *contacts.CustomerProfile {
	return &contacts.CustomerProfile{
		Base:      model.Base{ID: 1},
		ContactID: 1,
		Active:    true,
	}
}

func sampleSupplier() *contacts.SupplierProfile {
	return &contacts.SupplierProfile{
		Base:      model.Base{ID: 1},
		ContactID: 1,
		Active:    true,
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

func TestContactHandler_List_ReturnsContacts(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[contacts.Contact], error) {
			return &query.Page[contacts.Contact]{Items: []*contacts.Contact{sampleContact()}, Count: 1}, nil
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactHandler_List_RejectsInvalidQuery(t *testing.T) {
	app := contactsHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactHandler_List_RequiresTenant(t *testing.T) {
	app := contactsHandlerTestNoTenant(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestContactHandler_List_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[contacts.Contact], error) {
			return nil, errors.New("db down")
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactHandler_List_ExportsCSV(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[contacts.Contact], error) {
			return &query.Page[contacts.Contact]{Items: []*contacts.Contact{sampleContact()}, Count: 1}, nil
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactHandler_Get_ReturnsContact(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactHandler_Get_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactHandler_Get_ReturnsNotFoundForForeignTenant(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return foreignContact(), nil
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactHandler_Get_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, errors.New("db down")
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactHandler_Get_RejectsInvalidID(t *testing.T) {
	app := contactsHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactHandler_Create_CreatesContact(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{})
	contactsDAO.CreateWithDetailsFunc = func(_ context.Context, contact *contacts.Contact, _ []*contacts.ContactAddress, _ []*contacts.ContactBankAccount, _ *contacts.CustomerProfile, _ *contacts.SupplierProfile) (*contacts.Contact, error) {
		contact.ID = 1
		return contact, nil
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactsHandlerTest(t, contactsDAO, svc)

	body := `{"name":"Acme","lang":"en","active":true}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestContactHandler_Create_CreatesContactWithDetails(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{})
	contactsDAO.CreateWithDetailsFunc = func(_ context.Context, contact *contacts.Contact, addresses []*contacts.ContactAddress, banks []*contacts.ContactBankAccount, customer *contacts.CustomerProfile, supplier *contacts.SupplierProfile) (*contacts.Contact, error) {
		if len(addresses) != 1 || len(banks) != 1 || customer == nil || supplier == nil {
			t.Error("expected addresses, banks, customer and supplier to be mapped")
		}
		contact.ID = 1
		return contact, nil
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactsHandlerTest(t, contactsDAO, svc)

	body := `{"name":"Acme","addresses":[{"type":"billing","is_default":true}],"bank_accounts":[{"bank_name":"Bank X"}],"customer":{"active":true},"supplier":{"active":true}}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestContactHandler_Create_RejectsValidation(t *testing.T) {
	app := contactsHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), emptyContactSvc())

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactHandler_Create_RejectsInvalidAddressType(t *testing.T) {
	app := contactsHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), emptyContactSvc())

	body := `{"name":"Acme","addresses":[{"type":"invalid"}]}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactHandler_Create_RejectsMultipleDefaultAddresses(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{})
	contactsDAO.CreateWithDetailsFunc = func(_ context.Context, _ *contacts.Contact, _ []*contacts.ContactAddress, _ []*contacts.ContactBankAccount, _ *contacts.CustomerProfile, _ *contacts.SupplierProfile) (*contacts.Contact, error) {
		return nil, contacts.ErrMultipleDefaultAddresses
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactsHandlerTest(t, contactsDAO, svc)

	body := `{"name":"Acme","addresses":[{"type":"billing","is_default":true},{"type":"billing","is_default":true}]}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactHandler_Create_RejectsBlankName(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{})
	contactsDAO.CreateWithDetailsFunc = func(_ context.Context, _ *contacts.Contact, _ []*contacts.ContactAddress, _ []*contacts.ContactBankAccount, _ *contacts.CustomerProfile, _ *contacts.SupplierProfile) (*contacts.Contact, error) {
		return nil, contacts.ErrNameRequired
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactsHandlerTest(t, contactsDAO, svc)

	body := `{"name":"   "}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactHandler_Create_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{})
	contactsDAO.CreateWithDetailsFunc = func(_ context.Context, _ *contacts.Contact, _ []*contacts.ContactAddress, _ []*contacts.ContactBankAccount, _ *contacts.CustomerProfile, _ *contacts.SupplierProfile) (*contacts.Contact, error) {
		return nil, errors.New("db down")
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactsHandlerTest(t, contactsDAO, svc)

	body := `{"name":"Acme"}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactHandler_Update_UpdatesContact(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
		UpdateFunc: func(_ context.Context, contact *contacts.Contact) (*contacts.Contact, error) {
			return contact, nil
		},
	})
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactsHandlerTest(t, contactsDAO, svc)

	body := `{"name":"Acme Corp","lang":"en"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactHandler_Update_RejectsInvalidID(t *testing.T) {
	app := contactsHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), emptyContactSvc())

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactHandler_Update_RejectsValidation(t *testing.T) {
	app := contactsHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), emptyContactSvc())

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactHandler_Update_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactHandler_Update_ReturnsNotFoundForForeignTenant(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return foreignContact(), nil
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactHandler_Update_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, errors.New("db down")
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactHandler_Update_RejectsBlankName(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactsHandlerTest(t, contactsDAO, svc)

	body := `{"name":"   "}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
		UpdateFunc: func(_ context.Context, _ *contacts.Contact) (*contacts.Contact, error) {
			return nil, errors.New("db down")
		},
	})
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactsHandlerTest(t, contactsDAO, svc)

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactHandler_Delete_DeletesContact(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return nil
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestContactHandler_Delete_RejectsInvalidID(t *testing.T) {
	app := contactsHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactHandler_Delete_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactHandler_Delete_ReturnsNotFoundForForeignTenant(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return foreignContact(), nil
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactHandler_Delete_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, errors.New("db down")
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("referenced")
		},
	})
	app := contactsHandlerTest(t, contactsDAO, emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactHandler_WriteContactError_MapsErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "contact not found", err: contacts.ErrContactNotFound, want: http.StatusNotFound},
		{name: "name required", err: contacts.ErrNameRequired, want: http.StatusUnprocessableEntity},
		{name: "invalid address type", err: contacts.ErrInvalidAddressType, want: http.StatusUnprocessableEntity},
		{name: "multiple default addresses", err: contacts.ErrMultipleDefaultAddresses, want: http.StatusUnprocessableEntity},
		{name: "unexpected error", err: errors.New("boom"), want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/err", func(c fiber.Ctx) error {
				return writeContactError(c, tt.err)
			})

			resp, err := doRequest(app, http.MethodPost, "/err", "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestContactRelationHandler_ListAddresses_ReturnsAddresses(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{})
	addresses.ListByContactFunc = func(_ context.Context, _ uint64) ([]*contacts.ContactAddress, error) {
		return []*contacts.ContactAddress{sampleAddress()}, nil
	}
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1/addresses/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactRelationHandler_ListAddresses_ExportsCSV(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{})
	addresses.ListByContactFunc = func(_ context.Context, _ uint64) ([]*contacts.ContactAddress, error) {
		return []*contacts.ContactAddress{sampleAddress()}, nil
	}
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1/addresses/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactRelationHandler_ListAddresses_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/abc/addresses/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_ListAddresses_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1/addresses/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_ListAddresses_ReturnsNotFoundForForeignTenant(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return foreignContact(), nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1/addresses/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_ListAddresses_ReturnsServerErrorOnContactLookup(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, errors.New("db down")
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1/addresses/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_ListAddresses_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{})
	addresses.ListByContactFunc = func(_ context.Context, _ uint64) ([]*contacts.ContactAddress, error) {
		return nil, errors.New("db down")
	}
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1/addresses/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_CreateAddress_CreatesAddress(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		CreateFunc: func(_ context.Context, address *contacts.ContactAddress) (*contacts.ContactAddress, error) {
			address.ID = 1
			return address, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"billing","line1":"123 Main St","is_default":true}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/1/addresses/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestContactRelationHandler_CreateAddress_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"billing"}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/abc/addresses/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_CreateAddress_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"billing"}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/1/addresses/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_CreateAddress_RejectsValidation(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"invalid"}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/1/addresses/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_CreateAddress_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		CreateFunc: func(_ context.Context, _ *contacts.ContactAddress) (*contacts.ContactAddress, error) {
			return nil, errors.New("db down")
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"billing"}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/1/addresses/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateAddress_UpdatesAddress(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			return sampleAddress(), nil
		},
		UpdateFunc: func(_ context.Context, address *contacts.ContactAddress) (*contacts.ContactAddress, error) {
			return address, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"shipping","line1":"456 Oak Ave"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/addresses/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateAddress_RejectsInvalidAddressID(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"billing"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/addresses/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateAddress_RejectsValidation(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"invalid"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/addresses/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateAddress_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			return nil, errors.New("db down")
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"billing"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/addresses/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateAddress_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"billing"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/addresses/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateAddress_ReturnsNotFoundForOtherContact(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			address := sampleAddress()
			address.ContactID = 99
			return address, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"billing"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/addresses/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateAddress_ReturnsServerErrorOnUpdate(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			return sampleAddress(), nil
		},
		UpdateFunc: func(_ context.Context, _ *contacts.ContactAddress) (*contacts.ContactAddress, error) {
			return nil, errors.New("db down")
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"billing"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/addresses/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteAddress_DeletesAddress(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			return sampleAddress(), nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/addresses/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteAddress_RejectsInvalidAddressID(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/addresses/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteAddress_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			return nil, errors.New("db down")
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/addresses/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteAddress_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/addresses/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteAddress_ReturnsServerErrorOnDelete(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			return sampleAddress(), nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/addresses/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_SetDefaultAddress_SetsDefault(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			return sampleAddress(), nil
		},
	})
	svc := contactsTestSvc(contactsDAO, addresses, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), svc)

	resp, err := doRequest(app, http.MethodPost, "/contacts/1/addresses/1/default", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactRelationHandler_SetDefaultAddress_RejectsInvalidAddressID(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodPost, "/contacts/1/addresses/abc/default", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_SetDefaultAddress_ReturnsNotFoundForAddress(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			return nil, nil
		},
	})
	svc := contactsTestSvc(contactsDAO, addresses, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), svc)

	resp, err := doRequest(app, http.MethodPost, "/contacts/1/addresses/1/default", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_SetDefaultAddress_ReturnsNotFoundForOtherContact(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			address := sampleAddress()
			address.ContactID = 99
			return address, nil
		},
	})
	svc := contactsTestSvc(contactsDAO, addresses, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), svc)

	resp, err := doRequest(app, http.MethodPost, "/contacts/1/addresses/1/default", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_SetDefaultAddress_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			return nil, errors.New("db down")
		},
	})
	svc := contactsTestSvc(contactsDAO, addresses, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), svc)

	resp, err := doRequest(app, http.MethodPost, "/contacts/1/addresses/1/default", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_SetDefaultAddress_ReturnsServerErrorOnReload(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	calls := 0
	addresses := addressCRUD(dao.CRUDMock[contacts.ContactAddress]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactAddress, error) {
			calls++
			if calls == 1 {
				return sampleAddress(), nil
			}
			return nil, errors.New("db down")
		},
	})
	svc := contactsTestSvc(contactsDAO, addresses, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, contacts.SupplierProfileDAOMock{})
	app := contactRelationHandlerTest(t, contactsDAO, addresses, bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), svc)

	resp, err := doRequest(app, http.MethodPost, "/contacts/1/addresses/1/default", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_ListBankAccounts_ReturnsAccounts(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{})
	banks.ListByContactFunc = func(_ context.Context, _ uint64) ([]*contacts.ContactBankAccount, error) {
		return []*contacts.ContactBankAccount{sampleBankAccount()}, nil
	}
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1/bank-accounts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactRelationHandler_ListBankAccounts_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/abc/bank-accounts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_ListBankAccounts_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1/bank-accounts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_ListBankAccounts_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{})
	banks.ListByContactFunc = func(_ context.Context, _ uint64) ([]*contacts.ContactBankAccount, error) {
		return nil, errors.New("db down")
	}
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodGet, "/contacts/1/bank-accounts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_CreateBankAccount_CreatesAccount(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{
		CreateFunc: func(_ context.Context, account *contacts.ContactBankAccount) (*contacts.ContactBankAccount, error) {
			account.ID = 1
			return account, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":"Bank X","account_holder":"Acme"}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/1/bank-accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestContactRelationHandler_CreateBankAccount_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":"Bank X"}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/abc/bank-accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_CreateBankAccount_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":"Bank X"}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/1/bank-accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_CreateBankAccount_RejectsMalformedBody(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":`
	resp, err := doRequest(app, http.MethodPost, "/contacts/1/bank-accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestContactRelationHandler_CreateBankAccount_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{
		CreateFunc: func(_ context.Context, _ *contacts.ContactBankAccount) (*contacts.ContactBankAccount, error) {
			return nil, errors.New("db down")
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":"Bank X"}`
	resp, err := doRequest(app, http.MethodPost, "/contacts/1/bank-accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateBankAccount_UpdatesAccount(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactBankAccount, error) {
			return sampleBankAccount(), nil
		},
		UpdateFunc: func(_ context.Context, account *contacts.ContactBankAccount) (*contacts.ContactBankAccount, error) {
			return account, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":"Bank Y"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/bank-accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateBankAccount_RejectsInvalidAccountID(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":"Bank Y"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/bank-accounts/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateBankAccount_RejectsMalformedBody(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/bank-accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateBankAccount_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactBankAccount, error) {
			return nil, errors.New("db down")
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":"Bank Y"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/bank-accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateBankAccount_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactBankAccount, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":"Bank Y"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/bank-accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateBankAccount_ReturnsNotFoundForOtherContact(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactBankAccount, error) {
			account := sampleBankAccount()
			account.ContactID = 99
			return account, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":"Bank Y"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/bank-accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateBankAccount_ReturnsServerErrorOnUpdate(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactBankAccount, error) {
			return sampleBankAccount(), nil
		},
		UpdateFunc: func(_ context.Context, _ *contacts.ContactBankAccount) (*contacts.ContactBankAccount, error) {
			return nil, errors.New("db down")
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":"Bank Y"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/bank-accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteBankAccount_DeletesAccount(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactBankAccount, error) {
			return sampleBankAccount(), nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/bank-accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteBankAccount_RejectsInvalidAccountID(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/bank-accounts/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteBankAccount_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactBankAccount, error) {
			return nil, errors.New("db down")
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/bank-accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteBankAccount_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactBankAccount, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/bank-accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteBankAccount_ReturnsServerErrorOnDelete(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	banks := bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.ContactBankAccount, error) {
			return sampleBankAccount(), nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), banks, customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/bank-accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateAddress_ReturnsNotFoundForContact(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"billing"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/addresses/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteAddress_ReturnsNotFoundForContact(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/addresses/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_SetDefaultAddress_ReturnsNotFoundForContact(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodPost, "/contacts/1/addresses/1/default", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateBankAccount_ReturnsNotFoundForContact(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":"Bank Y"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/bank-accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteBankAccount_ReturnsNotFoundForContact(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/bank-accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateAddress_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"type":"billing"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/abc/addresses/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteAddress_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/abc/addresses/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_SetDefaultAddress_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodPost, "/contacts/abc/addresses/1/default", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_UpdateBankAccount_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"bank_name":"Bank Y"}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/abc/bank-accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_DeleteBankAccount_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/abc/bank-accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_EnableCustomer_EnablesRole(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	customers := customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{})
	customers.FindByContactFunc = func(_ context.Context, _ uint64) (*contacts.CustomerProfile, error) {
		return nil, nil
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, customers, contacts.SupplierProfileDAOMock{})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customers, supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), svc)

	body := `{"active":false}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/customer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactRelationHandler_EnableCustomer_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"active":true}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/abc/customer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_EnableCustomer_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"active":true}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/customer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_EnableCustomer_RejectsMalformedBody(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"active":`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/customer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestContactRelationHandler_EnableCustomer_ReturnsNotFoundFromService(t *testing.T) {
	calls := 0
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			calls++
			if calls > 1 {
				return nil, nil
			}
			return sampleContact(), nil
		},
	})
	customers := customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{})
	customers.FindByContactFunc = func(_ context.Context, _ uint64) (*contacts.CustomerProfile, error) {
		return nil, nil
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, customers, contacts.SupplierProfileDAOMock{})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customers, supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), svc)

	body := `{"active":true}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/customer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_EnableCustomer_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	customers := customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{})
	customers.FindByContactFunc = func(_ context.Context, _ uint64) (*contacts.CustomerProfile, error) {
		return nil, errors.New("db down")
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, customers, contacts.SupplierProfileDAOMock{})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customers, supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), svc)

	body := `{"active":true}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/customer", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_DisableCustomer_DisablesRole(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	customers := customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{})
	customers.FindByContactFunc = func(_ context.Context, _ uint64) (*contacts.CustomerProfile, error) {
		return sampleCustomer(), nil
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, customers, contacts.SupplierProfileDAOMock{})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customers, supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), svc)

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/customer", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactRelationHandler_DisableCustomer_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/abc/customer", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_DisableCustomer_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/customer", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_DisableCustomer_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	customers := customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{})
	customers.FindByContactFunc = func(_ context.Context, _ uint64) (*contacts.CustomerProfile, error) {
		return nil, errors.New("db down")
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, customers, contacts.SupplierProfileDAOMock{})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customers, supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), svc)

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/customer", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_EnableSupplier_EnablesRole(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	suppliers := supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{})
	suppliers.FindByContactFunc = func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
		return nil, nil
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, suppliers)
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), suppliers, svc)

	body := `{"active":false}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/supplier", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactRelationHandler_EnableSupplier_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"active":true}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/abc/supplier", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_EnableSupplier_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"active":true}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/supplier", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_EnableSupplier_RejectsMalformedBody(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	body := `{"active":`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/supplier", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestContactRelationHandler_EnableSupplier_ReturnsNotFoundFromService(t *testing.T) {
	calls := 0
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			calls++
			if calls > 1 {
				return nil, nil
			}
			return sampleContact(), nil
		},
	})
	suppliers := supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{})
	suppliers.FindByContactFunc = func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
		return nil, nil
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, suppliers)
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), suppliers, svc)

	body := `{"active":true}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/supplier", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_EnableSupplier_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	suppliers := supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{})
	suppliers.FindByContactFunc = func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
		return nil, errors.New("db down")
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, suppliers)
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), suppliers, svc)

	body := `{"active":true}`
	resp, err := doRequest(app, http.MethodPut, "/contacts/1/supplier", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContactRelationHandler_DisableSupplier_DisablesRole(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	suppliers := supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{})
	suppliers.FindByContactFunc = func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
		return sampleSupplier(), nil
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, suppliers)
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), suppliers, svc)

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/supplier", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContactRelationHandler_DisableSupplier_RejectsInvalidContactID(t *testing.T) {
	app := contactRelationHandlerTest(t, contactCRUD(dao.CRUDMock[contacts.Contact]{}), addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/abc/supplier", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContactRelationHandler_DisableSupplier_ReturnsNotFound(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return nil, nil
		},
	})
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{}), emptyContactSvc())

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/supplier", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContactRelationHandler_DisableSupplier_ReturnsServerError(t *testing.T) {
	contactsDAO := contactCRUD(dao.CRUDMock[contacts.Contact]{
		FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return sampleContact(), nil
		},
	})
	suppliers := supplierCRUD(dao.CRUDMock[contacts.SupplierProfile]{})
	suppliers.FindByContactFunc = func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
		return nil, errors.New("db down")
	}
	svc := contactsTestSvc(contactsDAO, contacts.ContactAddressDAOMock{}, contacts.ContactBankAccountDAOMock{}, contacts.CustomerProfileDAOMock{}, suppliers)
	app := contactRelationHandlerTest(t, contactsDAO, addressCRUD(dao.CRUDMock[contacts.ContactAddress]{}), bankCRUD(dao.CRUDMock[contacts.ContactBankAccount]{}), customerCRUD(dao.CRUDMock[contacts.CustomerProfile]{}), suppliers, svc)

	resp, err := doRequest(app, http.MethodDelete, "/contacts/1/supplier", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
