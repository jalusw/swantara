package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func activityApp(activities crm.ProspectActivityDAOMock, svc crm.ProspectActivityService) *fiber.App {
	return crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectActivityHandler(svc).Register(api, guards)
	})
}

func TestActivityHandler_List_ReturnsActivities(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[crm.ProspectActivity], error) {
				return &query.Page[crm.ProspectActivity]{Items: []*crm.ProspectActivity{sampleProspectActivity()}, Count: 1}, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodGet, "/crm/activities/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestActivityHandler_List_ExportsCSV(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[crm.ProspectActivity], error) {
				return &query.Page[crm.ProspectActivity]{Items: []*crm.ProspectActivity{sampleProspectActivity()}, Count: 1}, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodGet, "/crm/activities/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestActivityHandler_List_RejectsInvalidQuery(t *testing.T) {
	app := activityApp(crm.ProspectActivityDAOMock{}, crmActivityTestSvc(crm.ProspectActivityDAOMock{}, crm.ProspectDAOMock{}, contacts.ContactDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/crm/activities/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := crmTestApp(0, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectActivityHandler(svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/activities/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_List_ReturnsServerError(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[crm.ProspectActivity], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodGet, "/crm/activities/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestActivityHandler_Get_ReturnsActivity(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return sampleProspectActivity(), nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodGet, "/crm/activities/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestActivityHandler_Get_RejectsInvalidID(t *testing.T) {
	app := activityApp(crm.ProspectActivityDAOMock{}, crmActivityTestSvc(crm.ProspectActivityDAOMock{}, crm.ProspectDAOMock{}, contacts.ContactDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/crm/activities/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_Get_ReturnsNotFound(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return nil, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodGet, "/crm/activities/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestActivityHandler_Get_ReturnsServerError(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodGet, "/crm/activities/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestActivityHandler_Create_CreatesActivity(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			CreateFunc: func(_ context.Context, activity *crm.ProspectActivity) (*crm.ProspectActivity, error) {
				activity.ID = 1
				return activity, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	body := `{"type":"call","summary":"Intro call"}`
	resp, err := doRequest(app, http.MethodPost, "/crm/activities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestActivityHandler_Create_RejectsValidation(t *testing.T) {
	app := activityApp(crm.ProspectActivityDAOMock{}, crmActivityTestSvc(crm.ProspectActivityDAOMock{}, crm.ProspectDAOMock{}, contacts.ContactDAOMock{}))

	body := `{"summary":""}`
	resp, err := doRequest(app, http.MethodPost, "/crm/activities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_Create_RejectsBadBody(t *testing.T) {
	app := activityApp(crm.ProspectActivityDAOMock{}, crmActivityTestSvc(crm.ProspectActivityDAOMock{}, crm.ProspectDAOMock{}, contacts.ContactDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/crm/activities/", `{"summary":`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestActivityHandler_Create_RejectsInvalidDueDate(t *testing.T) {
	app := activityApp(crm.ProspectActivityDAOMock{}, crmActivityTestSvc(crm.ProspectActivityDAOMock{}, crm.ProspectDAOMock{}, contacts.ContactDAOMock{}))

	body := `{"summary":"Intro call","due_date":"tomorrow"}`
	resp, err := doRequest(app, http.MethodPost, "/crm/activities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_Create_RejectsUnknownLead(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{}
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return nil, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, leads, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	body := `{"summary":"Intro call","lead_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/activities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_Create_RejectsLeadInOtherOrg(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{}
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return leadInOtherOrg(), nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, leads, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	body := `{"summary":"Intro call","lead_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/activities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_Create_RejectsUnknownContact(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{}
	contactsMock := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return nil, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contactsMock)
	app := activityApp(activities, svc)

	body := `{"summary":"Intro call","contact_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/activities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_Create_MapsServiceErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"activity not found", crm.ErrActivityNotFound, 404},
		{"summary required", crm.ErrActivitySummaryRequired, 422},
		{"activity done", crm.ErrActivityDone, 409},
		{"prospect not found", crm.ErrLeadNotFound, 422},
		{"contact not found", crm.ErrContactNotFound, 422},
		{"internal error", errors.New("db down"), 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activities := crm.ProspectActivityDAOMock{
				CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
					CreateFunc: func(_ context.Context, _ *crm.ProspectActivity) (*crm.ProspectActivity, error) {
						return nil, tt.err
					},
				},
			}
			svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
			app := activityApp(activities, svc)

			body := `{"summary":"Intro call"}`
			resp, err := doRequest(app, http.MethodPost, "/crm/activities/", body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestActivityHandler_Update_UpdatesActivity(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return sampleProspectActivity(), nil
			},
			UpdateFunc: func(_ context.Context, activity *crm.ProspectActivity) (*crm.ProspectActivity, error) {
				return activity, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	body := `{"type":"meeting","summary":"Follow up"}`
	resp, err := doRequest(app, http.MethodPut, "/crm/activities/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestActivityHandler_Update_RejectsInvalidID(t *testing.T) {
	app := activityApp(crm.ProspectActivityDAOMock{}, crmActivityTestSvc(crm.ProspectActivityDAOMock{}, crm.ProspectDAOMock{}, contacts.ContactDAOMock{}))

	body := `{"summary":"Follow up"}`
	resp, err := doRequest(app, http.MethodPut, "/crm/activities/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_Update_RejectsValidation(t *testing.T) {
	app := activityApp(crm.ProspectActivityDAOMock{}, crmActivityTestSvc(crm.ProspectActivityDAOMock{}, crm.ProspectDAOMock{}, contacts.ContactDAOMock{}))

	body := `{"summary":""}`
	resp, err := doRequest(app, http.MethodPut, "/crm/activities/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_Update_ReturnsNotFound(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return nil, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	body := `{"summary":"Follow up"}`
	resp, err := doRequest(app, http.MethodPut, "/crm/activities/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestActivityHandler_Update_RejectsInvalidDueDate(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return sampleProspectActivity(), nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	body := `{"summary":"Follow up","due_date":"tomorrow"}`
	resp, err := doRequest(app, http.MethodPut, "/crm/activities/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	body := `{"summary":"Follow up"}`
	resp, err := doRequest(app, http.MethodPut, "/crm/activities/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestActivityHandler_Delete_DeletesActivity(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return sampleProspectActivity(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodDelete, "/crm/activities/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestActivityHandler_Delete_RejectsInvalidID(t *testing.T) {
	app := activityApp(crm.ProspectActivityDAOMock{}, crmActivityTestSvc(crm.ProspectActivityDAOMock{}, crm.ProspectDAOMock{}, contacts.ContactDAOMock{}))

	resp, err := doRequest(app, http.MethodDelete, "/crm/activities/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_Delete_ReturnsNotFound(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return nil, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodDelete, "/crm/activities/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestActivityHandler_Delete_ReturnsServerError(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return sampleProspectActivity(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return errors.New("db down")
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodDelete, "/crm/activities/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestActivityHandler_MarkDone_MarksActivityDone(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return sampleProspectActivity(), nil
			},
			UpdateFunc: func(_ context.Context, activity *crm.ProspectActivity) (*crm.ProspectActivity, error) {
				return activity, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodPost, "/crm/activities/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestActivityHandler_MarkDone_RejectsInvalidID(t *testing.T) {
	app := activityApp(crm.ProspectActivityDAOMock{}, crmActivityTestSvc(crm.ProspectActivityDAOMock{}, crm.ProspectDAOMock{}, contacts.ContactDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/crm/activities/abc/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestActivityHandler_MarkDone_ReturnsNotFound(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return nil, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodPost, "/crm/activities/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestActivityHandler_MarkDone_ReturnsConflictWhenAlreadyDone(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				activity := sampleProspectActivity()
				activity.Done = true
				return activity, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodPost, "/crm/activities/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestActivityHandler_MarkDone_ReturnsServerErrorOnFind(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodPost, "/crm/activities/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestActivityHandler_MarkDone_ReturnsServerErrorOnUpdate(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return sampleProspectActivity(), nil
			},
			UpdateFunc: func(_ context.Context, _ *crm.ProspectActivity) (*crm.ProspectActivity, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	resp, err := doRequest(app, http.MethodPost, "/crm/activities/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestActivityHandler_Create_SetsUserFromActor(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			CreateFunc: func(_ context.Context, activity *crm.ProspectActivity) (*crm.ProspectActivity, error) {
				activity.ID = 1
				if activity.UserID == nil || *activity.UserID != 5 {
					t.Errorf("user id = %v, want 5", activity.UserID)
				}
				return activity, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	body := `{"summary":"Intro call"}`
	resp, err := doRequest(app, http.MethodPost, "/crm/activities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestActivityHandler_Update_MarksDoneAtWhenDoneFirstTime(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				return sampleProspectActivity(), nil
			},
			UpdateFunc: func(_ context.Context, activity *crm.ProspectActivity) (*crm.ProspectActivity, error) {
				if activity.DoneAt == nil {
					t.Error("done_at should be set")
				}
				return activity, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	body := `{"summary":"Follow up","done":true}`
	resp, err := doRequest(app, http.MethodPut, "/crm/activities/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestActivityHandler_Update_SetsDone(t *testing.T) {
	activities := crm.ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[crm.ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.ProspectActivity, error) {
				activity := sampleProspectActivity()
				now := time.Now()
				activity.DoneAt = &now
				return activity, nil
			},
			UpdateFunc: func(_ context.Context, activity *crm.ProspectActivity) (*crm.ProspectActivity, error) {
				if !activity.Done {
					t.Error("done should be true")
				}
				return activity, nil
			},
		},
	}
	svc := crmActivityTestSvc(activities, crm.ProspectDAOMock{}, contacts.ContactDAOMock{})
	app := activityApp(activities, svc)

	body := `{"summary":"Follow up","done":true}`
	resp, err := doRequest(app, http.MethodPut, "/crm/activities/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}
