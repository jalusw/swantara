package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func paymentTermOf(id uint64, name string) *reference.PaymentTerm {
	return &reference.PaymentTerm{Base: model.Base{ID: id, CreatedAt: timeNow(), UpdatedAt: timeNow()}, OrganizationID: 10, Name: name, IsActive: true}
}

func paymentTermLineOf(id uint64, valueType string, value float64) *reference.PaymentTermLine {
	return &reference.PaymentTermLine{Base: model.Base{ID: id}, Sequence: 10, ValueType: valueType, Value: value}
}

func paymentTermCRUD() dao.CRUDMock[reference.PaymentTerm] {
	return dao.CRUDMock[reference.PaymentTerm]{}
}

func paymentTermHandlerTest(t *testing.T, terms reference.PaymentTermDAOMock, svc reference.PaymentTermService) *fiber.App {
	t.Helper()
	return referenceTestApp(t, true, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewPaymentTermHandler(svc)
		h.Register(api, guards)
	})
}

func paymentTermTestSvc(terms reference.PaymentTermDAOMock) reference.PaymentTermService {
	return reference.NewPaymentTermService(terms)
}

func TestPaymentTermHandler_List_ReturnsTerms(t *testing.T) {
	crud := paymentTermCRUD()
	crud.ListFunc = func(_ context.Context, q *query.Query) (*query.Page[reference.PaymentTerm], error) {
		return &query.Page[reference.PaymentTerm]{Items: []*reference.PaymentTerm{paymentTermOf(1, "Net 30")}, Count: 1}, nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ListLinesFunc: func(_ context.Context, termID uint64) ([]*reference.PaymentTermLine, error) {
			return []*reference.PaymentTermLine{paymentTermLineOf(1, "percent", 100)}, nil
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodGet, "/payment-terms/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPaymentTermHandler_List_RejectsInvalidQuery(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodGet, "/payment-terms/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_List_ReturnsServerError(t *testing.T) {
	crud := paymentTermCRUD()
	crud.ListFunc = func(_ context.Context, q *query.Query) (*query.Page[reference.PaymentTerm], error) {
		return nil, errors.New("db down")
	}
	terms := reference.PaymentTermDAOMock{CRUDMock: crud}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodGet, "/payment-terms/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_List_ReturnsServerErrorOnLines(t *testing.T) {
	crud := paymentTermCRUD()
	crud.ListFunc = func(_ context.Context, q *query.Query) (*query.Page[reference.PaymentTerm], error) {
		return &query.Page[reference.PaymentTerm]{Items: []*reference.PaymentTerm{paymentTermOf(1, "Net 30")}, Count: 1}, nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ListLinesFunc: func(_ context.Context, termID uint64) ([]*reference.PaymentTermLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodGet, "/payment-terms/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_List_ExportsCSV(t *testing.T) {
	crud := paymentTermCRUD()
	crud.ListFunc = func(_ context.Context, q *query.Query) (*query.Page[reference.PaymentTerm], error) {
		return &query.Page[reference.PaymentTerm]{Items: []*reference.PaymentTerm{paymentTermOf(1, "Net 30")}, Count: 1}, nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ListLinesFunc: func(_ context.Context, termID uint64) ([]*reference.PaymentTermLine, error) {
			return []*reference.PaymentTermLine{paymentTermLineOf(1, "percent", 100)}, nil
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodGet, "/payment-terms/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Get_ReturnsTerm(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, "Net 30"), nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ListLinesFunc: func(_ context.Context, termID uint64) ([]*reference.PaymentTermLine, error) {
			return []*reference.PaymentTermLine{paymentTermLineOf(1, "percent", 100)}, nil
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodGet, "/payment-terms/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Get_ReturnsNotFound(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return nil, nil
	}
	terms := reference.PaymentTermDAOMock{CRUDMock: crud}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodGet, "/payment-terms/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Get_RejectsInvalidID(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodGet, "/payment-terms/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Get_ReturnsServerErrorOnFind(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return nil, errors.New("db down")
	}
	terms := reference.PaymentTermDAOMock{CRUDMock: crud}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodGet, "/payment-terms/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Get_ReturnsServerErrorOnLines(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, "Net 30"), nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ListLinesFunc: func(_ context.Context, termID uint64) ([]*reference.PaymentTermLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodGet, "/payment-terms/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Create_CreatesTerm(t *testing.T) {
	crud := paymentTermCRUD()
	crud.CreateFunc = func(_ context.Context, term *reference.PaymentTerm) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, term.Name), nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ReplaceLinesFunc: func(_ context.Context, termID uint64, lines []*reference.PaymentTermLine) error {
			return nil
		},
		ListLinesFunc: func(_ context.Context, termID uint64) ([]*reference.PaymentTermLine, error) {
			return []*reference.PaymentTermLine{paymentTermLineOf(1, "percent", 100)}, nil
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 30","lines":[{"sequence":10,"value_type":"percent","value":100,"days_after":30}]}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Create_RejectsValidation(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"","lines":[]}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Create_MapsInvalidValueType(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 30","lines":[{"value_type":"bogus","value":100}]}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Create_MapsMultipleBalanceLines(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 30","lines":[{"value_type":"balance"},{"value_type":"balance"}]}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Create_MapsPercentSumMismatch(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 30","lines":[{"value_type":"percent","value":50}]}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Create_MapsPercentOverBalance(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 30","lines":[{"value_type":"percent","value":120},{"value_type":"balance"}]}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Create_ReturnsServerErrorOnCreate(t *testing.T) {
	crud := paymentTermCRUD()
	crud.CreateFunc = func(_ context.Context, term *reference.PaymentTerm) (*reference.PaymentTerm, error) {
		return nil, errors.New("db down")
	}
	terms := reference.PaymentTermDAOMock{CRUDMock: crud}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 30","lines":[{"value_type":"percent","value":100}]}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Create_ReturnsServerErrorOnReplaceLines(t *testing.T) {
	crud := paymentTermCRUD()
	crud.CreateFunc = func(_ context.Context, term *reference.PaymentTerm) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, term.Name), nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ReplaceLinesFunc: func(_ context.Context, termID uint64, lines []*reference.PaymentTermLine) error {
			return errors.New("db down")
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 30","lines":[{"value_type":"percent","value":100}]}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Create_ReturnsServerErrorOnLines(t *testing.T) {
	crud := paymentTermCRUD()
	crud.CreateFunc = func(_ context.Context, term *reference.PaymentTerm) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, term.Name), nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ReplaceLinesFunc: func(_ context.Context, termID uint64, lines []*reference.PaymentTermLine) error {
			return nil
		},
		ListLinesFunc: func(_ context.Context, termID uint64) ([]*reference.PaymentTermLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 30","lines":[{"value_type":"percent","value":100}]}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Splits_ComputesSplits(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, "Net 30"), nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ListLinesFunc: func(_ context.Context, termID uint64) ([]*reference.PaymentTermLine, error) {
			return []*reference.PaymentTermLine{paymentTermLineOf(1, "percent", 100)}, nil
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"total":"1000.00","date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/1/splits", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Splits_RejectsInvalidID(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"total":"1000.00","date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/abc/splits", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Splits_RejectsValidation(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"total":"","date":""}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/1/splits", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Splits_RejectsInvalidTotal(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"total":"abc","date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/1/splits", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Splits_RejectsInvalidDate(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"total":"1000.00","date":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/1/splits", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Splits_ReturnsNotFound(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return nil, nil
	}
	terms := reference.PaymentTermDAOMock{CRUDMock: crud}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"total":"1000.00","date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/1/splits", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Splits_ReturnsServerErrorOnFind(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return nil, errors.New("db down")
	}
	terms := reference.PaymentTermDAOMock{CRUDMock: crud}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"total":"1000.00","date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/1/splits", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Splits_ReturnsServerErrorOnLines(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, "Net 30"), nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ListLinesFunc: func(_ context.Context, termID uint64) ([]*reference.PaymentTermLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"total":"1000.00","date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/1/splits", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Splits_MapsInvalidLines(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, "Net 30"), nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ListLinesFunc: func(_ context.Context, termID uint64) ([]*reference.PaymentTermLine, error) {
			return []*reference.PaymentTermLine{paymentTermLineOf(1, "bogus", 100)}, nil
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"total":"1000.00","date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/payment-terms/1/splits", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Update_UpdatesTerm(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, "Net 30"), nil
	}
	crud.UpdateFunc = func(_ context.Context, term *reference.PaymentTerm) (*reference.PaymentTerm, error) {
		return term, nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ReplaceLinesFunc: func(_ context.Context, termID uint64, lines []*reference.PaymentTermLine) error {
			return nil
		},
		ListLinesFunc: func(_ context.Context, termID uint64) ([]*reference.PaymentTermLine, error) {
			return []*reference.PaymentTermLine{paymentTermLineOf(1, "percent", 100)}, nil
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 45","lines":[{"value_type":"percent","value":100}]}`
	resp, err := doRequest(app, http.MethodPut, "/payment-terms/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Update_RejectsInvalidID(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 45","lines":[{"value_type":"percent","value":100}]}`
	resp, err := doRequest(app, http.MethodPut, "/payment-terms/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Update_RejectsValidation(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"","lines":[]}`
	resp, err := doRequest(app, http.MethodPut, "/payment-terms/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Update_MapsInvalidValueType(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 45","lines":[{"value_type":"bogus","value":100}]}`
	resp, err := doRequest(app, http.MethodPut, "/payment-terms/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Update_ReturnsNotFound(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return nil, nil
	}
	terms := reference.PaymentTermDAOMock{CRUDMock: crud}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 45","lines":[{"value_type":"percent","value":100}]}`
	resp, err := doRequest(app, http.MethodPut, "/payment-terms/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return nil, errors.New("db down")
	}
	terms := reference.PaymentTermDAOMock{CRUDMock: crud}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 45","lines":[{"value_type":"percent","value":100}]}`
	resp, err := doRequest(app, http.MethodPut, "/payment-terms/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Update_ReturnsServerErrorOnUpdate(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, "Net 30"), nil
	}
	crud.UpdateFunc = func(_ context.Context, term *reference.PaymentTerm) (*reference.PaymentTerm, error) {
		return nil, errors.New("db down")
	}
	terms := reference.PaymentTermDAOMock{CRUDMock: crud}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 45","lines":[{"value_type":"percent","value":100}]}`
	resp, err := doRequest(app, http.MethodPut, "/payment-terms/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Update_ReturnsServerErrorOnReplaceLines(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, "Net 30"), nil
	}
	crud.UpdateFunc = func(_ context.Context, term *reference.PaymentTerm) (*reference.PaymentTerm, error) {
		return term, nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ReplaceLinesFunc: func(_ context.Context, termID uint64, lines []*reference.PaymentTermLine) error {
			return errors.New("db down")
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 45","lines":[{"value_type":"percent","value":100}]}`
	resp, err := doRequest(app, http.MethodPut, "/payment-terms/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Update_ReturnsServerErrorOnLines(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, "Net 30"), nil
	}
	crud.UpdateFunc = func(_ context.Context, term *reference.PaymentTerm) (*reference.PaymentTerm, error) {
		return term, nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		ReplaceLinesFunc: func(_ context.Context, termID uint64, lines []*reference.PaymentTermLine) error {
			return nil
		},
		ListLinesFunc: func(_ context.Context, termID uint64) ([]*reference.PaymentTermLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	body := `{"name":"Net 45","lines":[{"value_type":"percent","value":100}]}`
	resp, err := doRequest(app, http.MethodPut, "/payment-terms/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Delete_DeletesTerm(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, "Net 30"), nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		DeleteWithLinesFunc: func(_ context.Context, termID uint64) error {
			return nil
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodDelete, "/payment-terms/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Delete_RejectsInvalidID(t *testing.T) {
	terms := reference.PaymentTermDAOMock{CRUDMock: paymentTermCRUD()}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodDelete, "/payment-terms/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Delete_ReturnsNotFound(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return nil, nil
	}
	terms := reference.PaymentTermDAOMock{CRUDMock: crud}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodDelete, "/payment-terms/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return nil, errors.New("db down")
	}
	terms := reference.PaymentTermDAOMock{CRUDMock: crud}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodDelete, "/payment-terms/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	crud := paymentTermCRUD()
	crud.FindFunc = func(_ context.Context, id uint64) (*reference.PaymentTerm, error) {
		return paymentTermOf(1, "Net 30"), nil
	}
	terms := reference.PaymentTermDAOMock{
		CRUDMock: crud,
		DeleteWithLinesFunc: func(_ context.Context, termID uint64) error {
			return errors.New("db down")
		},
	}
	app := paymentTermHandlerTest(t, terms, paymentTermTestSvc(terms))

	resp, err := doRequest(app, http.MethodDelete, "/payment-terms/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPaymentTermHandler_WritePaymentTermError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "term not found", err: reference.ErrTermNotFound, want: http.StatusNotFound},
		{name: "invalid lines", err: reference.ErrInvalidTermLines, want: http.StatusUnprocessableEntity},
		{name: "multiple balance", err: reference.ErrMultipleBalanceLines, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writePaymentTermError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
