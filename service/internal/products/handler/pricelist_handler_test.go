package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestPriceBookHandler_List_ReturnsPriceBooks(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[products.PriceBook], error) {
				return &query.Page[products.PriceBook]{Items: []*products.PriceBook{samplePriceBook()}, Count: 1}, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPriceBookHandler_List_ExportsCSV(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[products.PriceBook], error) {
				return &query.Page[products.PriceBook]{Items: []*products.PriceBook{samplePriceBook()}, Count: 1}, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPriceBookHandler_List_RejectsInvalidQuery(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_List_RequiresTenant(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := price_bookHandlerTestNoTenant(t, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestPriceBookHandler_List_ReturnsServerError(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[products.PriceBook], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPriceBookHandler_Get_ReturnsPriceBook(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPriceBookHandler_Get_ReturnsNotFound(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPriceBookHandler_Get_ReturnsNotFoundForForeignTenant(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return foreignPriceBook(), nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPriceBookHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_Get_ReturnsServerError(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPriceBookHandler_Create_CreatesPriceBook(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			CreateFunc: func(_ context.Context, price_book *products.PriceBook) (*products.PriceBook, error) {
				price_book.ID = 1
				return price_book, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"name":"Standard","currency_code":"IDR"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPriceBookHandler_Create_CreatesInactivePriceBook(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			CreateFunc: func(_ context.Context, price_book *products.PriceBook) (*products.PriceBook, error) {
				price_book.ID = 1
				return price_book, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"name":"Standard","active":false}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPriceBookHandler_Create_RejectsValidation(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{}, svc)

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_Create_ReturnsServerError(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			CreateFunc: func(_ context.Context, _ *products.PriceBook) (*products.PriceBook, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"name":"Standard"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPriceBookHandler_Update_UpdatesPriceBook(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
			UpdateFunc: func(_ context.Context, price_book *products.PriceBook) (*products.PriceBook, error) {
				return price_book, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"name":"Standard Plus","active":true}`
	resp, err := doRequest(app, http.MethodPut, "/price_books/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPriceBookHandler_Update_RejectsInvalidID(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{}, svc)

	body := `{"name":"Standard Plus"}`
	resp, err := doRequest(app, http.MethodPut, "/price_books/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_Update_RejectsValidation(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{}, svc)

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/price_books/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_Update_ReturnsNotFound(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"name":"Standard Plus"}`
	resp, err := doRequest(app, http.MethodPut, "/price_books/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPriceBookHandler_Update_ReturnsNotFoundForForeignTenant(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return foreignPriceBook(), nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"name":"Standard Plus"}`
	resp, err := doRequest(app, http.MethodPut, "/price_books/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPriceBookHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"name":"Standard Plus"}`
	resp, err := doRequest(app, http.MethodPut, "/price_books/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPriceBookHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
			UpdateFunc: func(_ context.Context, _ *products.PriceBook) (*products.PriceBook, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"name":"Standard Plus"}`
	resp, err := doRequest(app, http.MethodPut, "/price_books/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPriceBookHandler_Delete_DeletesPriceBook(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodDelete, "/price_books/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestPriceBookHandler_Delete_RejectsInvalidID(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodDelete, "/price_books/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_Delete_ReturnsNotFound(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodDelete, "/price_books/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPriceBookHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodDelete, "/price_books/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPriceBookHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return errors.New("referenced")
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodDelete, "/price_books/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPriceBookHandler_ListRules_ReturnsRules(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	rules := products.PriceRuleDAOMock{
		ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*products.PriceRule, error) {
			return []*products.PriceRule{sampleRule()}, nil
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, rules)
	app := price_bookHandlerTest(t, price_books, rules, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/1/rules", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPriceBookHandler_ListRules_RejectsInvalidID(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/abc/rules", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_ListRules_ReturnsNotFound(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/1/rules", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPriceBookHandler_ListRules_ReturnsServerErrorOnFind(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/1/rules", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPriceBookHandler_ListRules_ReturnsServerError(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	rules := products.PriceRuleDAOMock{
		ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*products.PriceRule, error) {
			return nil, errors.New("db down")
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, rules)
	app := price_bookHandlerTest(t, price_books, rules, svc)

	resp, err := doRequest(app, http.MethodGet, "/price_books/1/rules", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPriceBookHandler_CreateRule_CreatesRule(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	rules := products.PriceRuleDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceRule]{
			CreateFunc: func(_ context.Context, rule *products.PriceRule) (*products.PriceRule, error) {
				rule.ID = 1
				return rule, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, rules)
	app := price_bookHandlerTest(t, price_books, rules, svc)

	body := `{"applies_to":"all","compute_type":"fixed","fixed_price":100}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPriceBookHandler_CreateRule_RejectsInvalidID(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{}, svc)

	body := `{"applies_to":"all","compute_type":"fixed"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/abc/rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_CreateRule_RejectsValidation(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_CreateRule_ReturnsNotFound(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"applies_to":"all","compute_type":"fixed"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPriceBookHandler_CreateRule_ReturnsServerErrorOnFind(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"applies_to":"all","compute_type":"fixed"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPriceBookHandler_CreateRule_RejectsInvalidAppliesTo(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"applies_to":"bogus","compute_type":"fixed"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_CreateRule_RejectsInvalidComputeType(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"applies_to":"all","compute_type":"bogus"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_CreateRule_RejectsInvalidScope(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"applies_to":"item","compute_type":"fixed"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_CreateRule_ReturnsServerErrorOnCreate(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	rules := products.PriceRuleDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceRule]{
			CreateFunc: func(_ context.Context, _ *products.PriceRule) (*products.PriceRule, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, rules)
	app := price_bookHandlerTest(t, price_books, rules, svc)

	body := `{"applies_to":"all","compute_type":"fixed"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPriceBookHandler_ResolvePrice_ResolvesBasePrice(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return sampleVariant(), nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	rules := products.PriceRuleDAOMock{
		ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*products.PriceRule, error) {
			return []*products.PriceRule{}, nil
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, price_books, rules)
	app := price_bookHandlerTest(t, price_books, rules, svc)

	body := `{"variant_id":1,"qty":"1.5"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPriceBookHandler_ResolvePrice_ResolvesWithRule(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return sampleVariant(), nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	rules := products.PriceRuleDAOMock{
		ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*products.PriceRule, error) {
			return []*products.PriceRule{sampleRule()}, nil
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, price_books, rules)
	app := price_bookHandlerTest(t, price_books, rules, svc)

	body := `{"variant_id":1,"qty":"1.5","date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPriceBookHandler_ResolvePrice_RejectsInvalidID(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{}, svc)

	body := `{"variant_id":1,"qty":"1"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/abc/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_ResolvePrice_RejectsValidation(t *testing.T) {
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, products.PriceBookDAOMock{}, products.PriceRuleDAOMock{}, svc)

	body := `{}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_ResolvePrice_RejectsInvalidQuantity(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"variant_id":1,"qty":"abc"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_ResolvePrice_RejectsInvalidDate(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"variant_id":1,"qty":"1","date":"bogus"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_ResolvePrice_ReturnsNotFoundForPriceBook(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"variant_id":1,"qty":"1"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPriceBookHandler_ResolvePrice_ReturnsNotFoundForVariant(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, variants, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"variant_id":1,"qty":"1"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPriceBookHandler_ResolvePrice_ReturnsNotFoundForTemplate(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return sampleVariant(), nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return nil, nil
			},
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"variant_id":1,"qty":"1"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPriceBookHandler_ResolvePrice_RejectsUnavailablePrice(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return samplePriceBook(), nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return sampleVariant(), nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return sampleProduct(), nil
			},
		},
	}
	rule := sampleRule()
	rule.ComputeType = products.ComputeFormula
	rules := products.PriceRuleDAOMock{
		ListByPriceBookFunc: func(_ context.Context, _ uint64) ([]*products.PriceRule, error) {
			return []*products.PriceRule{rule}, nil
		},
	}
	svc := productTestSvc(templates, variants, dao.CRUDMock[reference.ItemCategory]{}, price_books, rules)
	app := price_bookHandlerTest(t, price_books, rules, svc)

	body := `{"variant_id":1,"qty":"1"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPriceBookHandler_ResolvePrice_ReturnsServerError(t *testing.T) {
	price_books := products.PriceBookDAOMock{
		CRUDMock: dao.CRUDMock[products.PriceBook]{
			FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productTestSvc(products.ItemDAOMock{}, products.ItemVariantDAOMock{}, dao.CRUDMock[reference.ItemCategory]{}, price_books, products.PriceRuleDAOMock{})
	app := price_bookHandlerTest(t, price_books, products.PriceRuleDAOMock{}, svc)

	body := `{"variant_id":1,"qty":"1"}`
	resp, err := doRequest(app, http.MethodPost, "/price_books/1/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPriceBookHandler_WritePriceBookError_MapsErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "price_book not found", err: products.ErrPriceBookNotFound, want: http.StatusNotFound},
		{name: "variant not found", err: products.ErrVariantNotFound, want: http.StatusNotFound},
		{name: "template not found", err: products.ErrItemNotFound, want: http.StatusNotFound},
		{name: "invalid applies to", err: products.ErrInvalidAppliesTo, want: http.StatusUnprocessableEntity},
		{name: "invalid compute type", err: products.ErrInvalidComputeType, want: http.StatusUnprocessableEntity},
		{name: "invalid rule scope", err: products.ErrInvalidRuleScope, want: http.StatusUnprocessableEntity},
		{name: "price unavailable", err: products.ErrPriceUnavailable, want: http.StatusUnprocessableEntity},
		{name: "unexpected error", err: errors.New("boom"), want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/err", func(c fiber.Ctx) error {
				return writePriceBookError(c, tt.err)
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
