package organization

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func moduleStoreWith(rows ...*reference.OrganizationModule) dao.CRUDMock[reference.OrganizationModule] {
	return dao.CRUDMock[reference.OrganizationModule]{
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.OrganizationModule], error) {
			items := make([]*reference.OrganizationModule, 0, len(rows))
			for _, row := range rows {
				match := true
				for _, filter := range q.Filters {
					switch filter.Field {
					case "organization_id":
						if row.OrganizationID != filter.Value.(uint64) {
							match = false
						}
					case "module_id":
						if row.ModuleID != filter.Value.(string) {
							match = false
						}
					}
				}
				if match {
					items = append(items, row)
				}
			}
			return &query.Page[reference.OrganizationModule]{Items: items, Count: int64(len(items))}, nil
		},
	}
}

func TestOrganizationService_ListModules(t *testing.T) {
	ctx := context.Background()

	t.Run("returns all modules active when store unset", func(t *testing.T) {
		orgs, currencies, members := quickCreateMocks("IDR", nil)
		svc := NewOrganizationService(orgs, currencies, members)

		states, err := svc.ListModules(ctx, 11)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(states) != len(OrgModules) {
			t.Fatalf("states = %d, want %d", len(states), len(OrgModules))
		}
		for _, state := range states {
			if !state.Active {
				t.Errorf("module %q inactive, want active by default", state.ModuleID)
			}
		}
	})

	t.Run("reflects stored active flags", func(t *testing.T) {
		orgs, currencies, members := quickCreateMocks("IDR", nil)
		svc := NewOrganizationService(orgs, currencies, members)
		svc.SetModuleStore(moduleStoreWith(
			&reference.OrganizationModule{OrganizationID: 11, ModuleID: "hr", Active: false},
		))

		states, err := svc.ListModules(ctx, 11)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		byModule := make(map[string]bool, len(states))
		for _, state := range states {
			byModule[state.ModuleID] = state.Active
		}
		if byModule["hr"] {
			t.Errorf("hr active, want inactive")
		}
		if !byModule["crm"] {
			t.Errorf("crm inactive, want active")
		}
	})
}

func TestOrganizationService_SetModuleActive(t *testing.T) {
	ctx := context.Background()

	t.Run("rejects unknown module", func(t *testing.T) {
		orgs, currencies, members := quickCreateMocks("IDR", nil)
		svc := NewOrganizationService(orgs, currencies, members)
		svc.SetModuleStore(moduleStoreWith())

		err := svc.SetModuleActive(ctx, 11, "other", false)
		helper.AssertError(t, err, true, ErrModuleUnknown)
	})

	t.Run("creates a row when none exists", func(t *testing.T) {
		orgs, currencies, members := quickCreateMocks("IDR", nil)
		svc := NewOrganizationService(orgs, currencies, members)
		var created *reference.OrganizationModule
		store := moduleStoreWith()
		store.CreateFunc = func(_ context.Context, entity *reference.OrganizationModule) (*reference.OrganizationModule, error) {
			created = entity
			return entity, nil
		}
		svc.SetModuleStore(store)

		if err := svc.SetModuleActive(ctx, 11, "hr", false); helper.AssertError(t, err, false, nil) {
			return
		}
		if created == nil || created.OrganizationID != 11 || created.ModuleID != "hr" || created.Active {
			t.Errorf("created = %+v, want org 11 hr inactive", created)
		}
	})

	t.Run("updates the existing row", func(t *testing.T) {
		orgs, currencies, members := quickCreateMocks("IDR", nil)
		svc := NewOrganizationService(orgs, currencies, members)
		existing := &reference.OrganizationModule{OrganizationID: 11, ModuleID: "hr", Active: true}
		store := moduleStoreWith(existing)
		var updated *reference.OrganizationModule
		store.UpdateFunc = func(_ context.Context, entity *reference.OrganizationModule) (*reference.OrganizationModule, error) {
			updated = entity
			return entity, nil
		}
		svc.SetModuleStore(store)

		if err := svc.SetModuleActive(ctx, 11, "hr", false); helper.AssertError(t, err, false, nil) {
			return
		}
		if updated == nil || updated.Active {
			t.Errorf("updated = %+v, want inactive", updated)
		}
	})
}

func TestOrganizationService_QuickCreate_ActivatesModules(t *testing.T) {
	ctx := context.Background()

	t.Run("activates every module for the new org", func(t *testing.T) {
		orgs, currencies, members := quickCreateMocks("IDR", nil)
		svc := NewOrganizationService(orgs, currencies, members)
		activated := make(map[string]bool)
		store := moduleStoreWith()
		store.CreateFunc = func(_ context.Context, entity *reference.OrganizationModule) (*reference.OrganizationModule, error) {
			activated[entity.ModuleID] = entity.Active
			return entity, nil
		}
		svc.SetModuleStore(store)

		created, err := svc.QuickCreate(ctx, "Toko Maju", "ID", 5)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(activated) != len(OrgModules) {
			t.Fatalf("activated = %d modules, want %d", len(activated), len(OrgModules))
		}
		for _, module := range OrgModules {
			if !activated[module] {
				t.Errorf("module %q not activated for org %d", module, created.ID)
			}
		}
	})
}
