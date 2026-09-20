package crm

import (
	"context"
	"strings"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type ProspectActivityService struct {
	activities ProspectActivityDAO
	leads      ProspectDAO
	contacts   contacts.ContactDAO
}

func NewProspectActivityService(activities ProspectActivityDAO, leads ProspectDAO, contacts contacts.ContactDAO) ProspectActivityService {
	return ProspectActivityService{activities: activities, leads: leads, contacts: contacts}
}

func (s ProspectActivityService) List(ctx context.Context, q *query.Query) (*query.Page[ProspectActivity], error) {
	return s.activities.List(ctx, q)
}

func (s ProspectActivityService) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[ProspectActivity], error) {
	return s.activities.ListInOrg(ctx, q, organizationID)
}

func (s ProspectActivityService) Find(ctx context.Context, id uint64) (*ProspectActivity, error) {
	return s.activities.Find(ctx, id)
}

func (s ProspectActivityService) FindInOrg(ctx context.Context, id, organizationID uint64) (*ProspectActivity, error) {
	return s.activities.FindInOrg(ctx, id, organizationID)
}

func (s ProspectActivityService) Delete(ctx context.Context, id uint64) error {
	return s.activities.Delete(ctx, id)
}

func (s ProspectActivityService) CreateActivity(ctx context.Context, organizationID uint64, activity *ProspectActivity) (*ProspectActivity, error) {
	if err := s.validateActivity(ctx, organizationID, activity); err != nil {
		return nil, err
	}
	if actorID := model.ActorID(ctx); actorID != 0 {
		activity.UserID = &actorID
	}
	return s.activities.Create(ctx, activity)
}

func (s ProspectActivityService) UpdateActivity(ctx context.Context, organizationID uint64, activity *ProspectActivity) (*ProspectActivity, error) {
	if err := s.validateActivity(ctx, organizationID, activity); err != nil {
		return nil, err
	}
	if activity.Done && activity.DoneAt == nil {
		now := time.Now()
		activity.DoneAt = &now
	}
	return s.activities.Update(ctx, activity)
}

func (s ProspectActivityService) MarkDone(ctx context.Context, activityID uint64) (*ProspectActivity, error) {
	activity, err := s.activities.Find(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if activity == nil {
		return nil, ErrActivityNotFound
	}
	if activity.Done {
		return nil, ErrActivityDone
	}

	now := time.Now()
	activity.Done = true
	activity.DoneAt = &now
	return s.activities.Update(ctx, activity)
}

func (s ProspectActivityService) validateActivity(ctx context.Context, organizationID uint64, activity *ProspectActivity) error {
	if strings.TrimSpace(activity.Summary) == "" {
		return ErrActivitySummaryRequired
	}
	if activity.Type == "" {
		activity.Type = ProspectActivityTypeNote
	}
	if activity.ProspectID != nil {
		prospect, err := s.leads.Find(ctx, *activity.ProspectID)
		if err != nil {
			return err
		}
		if prospect == nil || !leadBelongsToOrg(prospect, organizationID) {
			return ErrLeadNotFound
		}
	}
	if activity.ContactID != nil {
		contact, err := s.contacts.Find(ctx, *activity.ContactID)
		if err != nil {
			return err
		}
		if contact == nil || !contactBelongsToOrg(contact, organizationID) {
			return ErrContactNotFound
		}
	}
	return nil
}

func leadBelongsToOrg(prospect *Prospect, organizationID uint64) bool {
	return prospect.OrganizationID != nil && *prospect.OrganizationID == organizationID
}

func contactBelongsToOrg(contact *contacts.Contact, organizationID uint64) bool {
	return contact.OrganizationID != nil && *contact.OrganizationID == organizationID
}
