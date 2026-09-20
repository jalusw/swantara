package organization

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

var OrgModules = []string{
	"crm",
	"products",
	"inventory",
	"procurement",
	"hr",
	"finance",
	"projects",
	"quality",
	"subscriptions",
	"pos",
	"service",
}

type ModuleState struct {
	ModuleID string
	Active   bool
}

func IsKnownModule(moduleID string) bool {
	for _, module := range OrgModules {
		if module == moduleID {
			return true
		}
	}
	return false
}

func (s Service) ListModules(ctx context.Context, orgID uint64) ([]ModuleState, error) {
	states := make([]ModuleState, 0, len(OrgModules))
	for _, module := range OrgModules {
		states = append(states, ModuleState{ModuleID: module, Active: true})
	}
	if s.modules == nil {
		return states, nil
	}
	page, err := s.modules.List(ctx, &query.Query{
		Filters: []query.Filter{
			{Field: "organization_id", Operator: query.Equal, Value: orgID},
		},
		Pagination: &query.Pagination{Page: 1, Size: 100},
	})
	if err != nil {
		return nil, err
	}
	byModule := make(map[string]bool, len(page.Items))
	for _, item := range page.Items {
		byModule[item.ModuleID] = item.Active
	}
	for i, state := range states {
		if active, ok := byModule[state.ModuleID]; ok {
			states[i].Active = active
		}
	}
	return states, nil
}

func (s Service) SetModuleActive(ctx context.Context, orgID uint64, moduleID string, active bool) error {
	if !IsKnownModule(moduleID) {
		return ErrModuleUnknown
	}
	if s.modules == nil {
		return nil
	}
	page, err := s.modules.List(ctx, &query.Query{
		Filters: []query.Filter{
			{Field: "organization_id", Operator: query.Equal, Value: orgID},
			{Field: "module_id", Operator: query.Equal, Value: moduleID},
		},
		Pagination: &query.Pagination{Page: 1, Size: 1},
	})
	if err != nil {
		return err
	}
	if len(page.Items) > 0 {
		existing := page.Items[0]
		existing.Active = active
		_, err := s.modules.Update(ctx, existing)
		return err
	}
	_, err = s.modules.Create(ctx, &reference.OrganizationModule{
		OrganizationID: orgID,
		ModuleID:       moduleID,
		Active:         active,
	})
	return err
}

func (s Service) activateAllModules(ctx context.Context, orgID uint64) error {
	for _, module := range OrgModules {
		if err := s.SetModuleActive(ctx, orgID, module, true); err != nil {
			return err
		}
	}
	return nil
}
