package crm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestProspectActivityService_CreateActivity_RejectsInvalidInput(t *testing.T) {
	ctx := context.Background()
	svc := NewProspectActivityService(ProspectActivityDAOMock{}, ProspectDAOMock{}, contacts.ContactDAOMock{})

	tests := []struct {
		name     string
		activity *ProspectActivity
		wantErr  error
	}{
		{name: "empty summary", activity: &ProspectActivity{Summary: "  "}, wantErr: ErrActivitySummaryRequired},
		{name: "unknown prospect", activity: &ProspectActivity{Summary: "Follow up", ProspectID: helper.Ptr(uint64(9))}, wantErr: ErrLeadNotFound},
		{name: "unknown contact", activity: &ProspectActivity{Summary: "Follow up", ContactID: helper.Ptr(uint64(5))}, wantErr: ErrContactNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateActivity(ctx, 1, tt.activity)

			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestProspectActivityService_CreateActivity_DefaultsTypeAndSetsActor(t *testing.T) {
	ctx := model.ContextWithActor(context.Background(), 42)
	activities := ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[ProspectActivity]{
			CreateFunc: func(_ context.Context, activity *ProspectActivity) (*ProspectActivity, error) {
				activity.ID = 7
				return activity, nil
			},
		},
	}
	svc := NewProspectActivityService(activities, ProspectDAOMock{}, contacts.ContactDAOMock{})

	created, err := svc.CreateActivity(ctx, 1, &ProspectActivity{Summary: "Follow up"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Type != ProspectActivityTypeNote {
		t.Errorf("type = %q, want %q", created.Type, ProspectActivityTypeNote)
	}
	if created.UserID == nil || *created.UserID != 42 {
		t.Errorf("user_id = %v, want 42", created.UserID)
	}
}

func TestProspectActivityService_MarkDone_SetsDoneAndDoneAt(t *testing.T) {
	ctx := context.Background()
	activities := ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*ProspectActivity, error) {
				return &ProspectActivity{Base: model.Base{ID: 3}, Summary: "Follow up"}, nil
			},
		},
	}
	svc := NewProspectActivityService(activities, ProspectDAOMock{}, contacts.ContactDAOMock{})

	done, err := svc.MarkDone(ctx, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !done.Done {
		t.Error("done = false, want true")
	}
	if done.DoneAt == nil {
		t.Error("done_at = nil, want set")
	}
}

func TestProspectActivityService_MarkDone_RejectsInvalidActivity(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name       string
		activities ProspectActivityDAOMock
		wantErr    error
	}{
		{name: "missing activity", activities: ProspectActivityDAOMock{}, wantErr: ErrActivityNotFound},
		{
			name: "already done",
			activities: ProspectActivityDAOMock{
				CRUDMock: dao.CRUDMock[ProspectActivity]{
					FindFunc: func(_ context.Context, _ uint64) (*ProspectActivity, error) {
						return &ProspectActivity{Base: model.Base{ID: 3}, Summary: "Follow up", Done: true, DoneAt: &now}, nil
					},
				},
			},
			wantErr: ErrActivityDone,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewProspectActivityService(tt.activities, ProspectDAOMock{}, contacts.ContactDAOMock{})

			_, err := svc.MarkDone(ctx, 3)

			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestProspectActivityService_UpdateActivity_UpdatesSummary(t *testing.T) {
	ctx := context.Background()
	activities := ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[ProspectActivity]{
			UpdateFunc: func(_ context.Context, activity *ProspectActivity) (*ProspectActivity, error) {
				return activity, nil
			},
		},
	}
	svc := NewProspectActivityService(activities, ProspectDAOMock{}, contacts.ContactDAOMock{})

	updated, err := svc.UpdateActivity(ctx, 1, &ProspectActivity{Base: model.Base{ID: 3}, Summary: "Follow up"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Summary != "Follow up" {
		t.Errorf("summary = %q, want Follow up", updated.Summary)
	}
}

func TestProspectActivityService_UpdateActivity_RejectsEmptySummary(t *testing.T) {
	ctx := context.Background()
	svc := NewProspectActivityService(ProspectActivityDAOMock{}, ProspectDAOMock{}, contacts.ContactDAOMock{})

	_, err := svc.UpdateActivity(ctx, 1, &ProspectActivity{Summary: "  "})

	helper.AssertError(t, err, true, ErrActivitySummaryRequired)
}

func TestProspectActivityService_CreateActivity_PropagatesLeadLookupError(t *testing.T) {
	ctx := context.Background()
	leads := ProspectDAOMock{
		CRUDMock: dao.CRUDMock[Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*Prospect, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewProspectActivityService(ProspectActivityDAOMock{}, leads, contacts.ContactDAOMock{})

	_, err := svc.CreateActivity(ctx, 1, &ProspectActivity{Summary: "Follow up", ProspectID: helper.Ptr(uint64(1))})

	helper.AssertError(t, err, true, nil)
}

func TestProspectActivityService_CreateActivity_PropagatesContactLookupError(t *testing.T) {
	ctx := context.Background()
	contactsMock := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewProspectActivityService(ProspectActivityDAOMock{}, ProspectDAOMock{}, contactsMock)

	_, err := svc.CreateActivity(ctx, 1, &ProspectActivity{Summary: "Follow up", ContactID: helper.Ptr(uint64(5))})

	helper.AssertError(t, err, true, nil)
}

func TestProspectActivityService_MarkDone_PropagatesFindError(t *testing.T) {
	ctx := context.Background()
	activities := ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[ProspectActivity]{
			FindFunc: func(_ context.Context, _ uint64) (*ProspectActivity, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewProspectActivityService(activities, ProspectDAOMock{}, contacts.ContactDAOMock{})

	_, err := svc.MarkDone(ctx, 3)

	helper.AssertError(t, err, true, nil)
}

func TestProspectActivityService_CreateActivity_ScopesLeadToOrganization(t *testing.T) {
	ctx := context.Background()
	activities := ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[ProspectActivity]{
			CreateFunc: func(_ context.Context, activity *ProspectActivity) (*ProspectActivity, error) {
				return activity, nil
			},
		},
	}
	leads := foundLead(&Prospect{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(2)), Name: "Acme"})
	svc := NewProspectActivityService(activities, leads, contacts.ContactDAOMock{})

	_, err := svc.CreateActivity(ctx, 1, &ProspectActivity{Summary: "Follow up", ProspectID: helper.Ptr(uint64(1))})

	helper.AssertError(t, err, true, ErrLeadNotFound)
}

func TestProspectActivityService_CreateActivity_AcceptsLeadFromSameOrganization(t *testing.T) {
	ctx := context.Background()
	activities := ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[ProspectActivity]{
			CreateFunc: func(_ context.Context, activity *ProspectActivity) (*ProspectActivity, error) {
				return activity, nil
			},
		},
	}
	leads := foundLead(&Prospect{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Name: "Acme"})
	svc := NewProspectActivityService(activities, leads, contacts.ContactDAOMock{})

	created, err := svc.CreateActivity(ctx, 1, &ProspectActivity{Summary: "Follow up", ProspectID: helper.Ptr(uint64(1))})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Summary != "Follow up" {
		t.Errorf("summary = %q, want Follow up", created.Summary)
	}
}

func TestProspectActivityService_CreateActivity_ScopesContactToOrganization(t *testing.T) {
	ctx := context.Background()
	activities := ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[ProspectActivity]{
			CreateFunc: func(_ context.Context, activity *ProspectActivity) (*ProspectActivity, error) {
				return activity, nil
			},
		},
	}
	contactsMock := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return &contacts.Contact{Base: model.Base{ID: 5}, OrganizationID: helper.Ptr(uint64(2))}, nil
			},
		},
	}
	svc := NewProspectActivityService(activities, ProspectDAOMock{}, contactsMock)

	_, err := svc.CreateActivity(ctx, 1, &ProspectActivity{Summary: "Follow up", ContactID: helper.Ptr(uint64(5))})

	helper.AssertError(t, err, true, ErrContactNotFound)
}

func TestProspectActivityService_CreateActivity_AcceptsContactFromSameOrganization(t *testing.T) {
	ctx := context.Background()
	activities := ProspectActivityDAOMock{
		CRUDMock: dao.CRUDMock[ProspectActivity]{
			CreateFunc: func(_ context.Context, activity *ProspectActivity) (*ProspectActivity, error) {
				return activity, nil
			},
		},
	}
	contactsMock := contacts.ContactDAOMock{
		CRUDMock: dao.CRUDMock[contacts.Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
				return &contacts.Contact{Base: model.Base{ID: 5}, OrganizationID: helper.Ptr(uint64(1))}, nil
			},
		},
	}
	svc := NewProspectActivityService(activities, ProspectDAOMock{}, contactsMock)

	created, err := svc.CreateActivity(ctx, 1, &ProspectActivity{Summary: "Follow up", ContactID: helper.Ptr(uint64(5))})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ContactID == nil || *created.ContactID != 5 {
		t.Errorf("contact_id = %v, want 5", created.ContactID)
	}
}

func TestProspectActivityDAOMock_ListInOrg_DefaultsToEmptyPage(t *testing.T) {
	ctx := context.Background()

	page, err := (ProspectActivityDAOMock{}).ListInOrg(ctx, nil, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Count != 0 || len(page.Items) != 0 {
		t.Errorf("page = %+v, want empty page", page)
	}
}

func TestProspectActivityDAOMock_FindInOrg_DefaultsToNil(t *testing.T) {
	ctx := context.Background()

	activity, err := (ProspectActivityDAOMock{}).FindInOrg(ctx, 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity != nil {
		t.Errorf("activity = %+v, want nil", activity)
	}
}
