package products

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestItemDAOMock_DefaultCreateWithVariants(t *testing.T) {
	m := ItemDAOMock{}
	template := &Item{Name: "T-Shirt"}

	created, err := m.CreateWithVariants(context.Background(), template, []*ItemVariant{{}})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created != template {
		t.Errorf("created = %+v, want original template", created)
	}
}

func TestItemDAOMock_ConfiguredCreateWithVariants(t *testing.T) {
	called := false
	m := ItemDAOMock{
		CreateWithVariantsFunc: func(_ context.Context, _ *Item, _ []*ItemVariant) (*Item, error) {
			called = true
			return nil, nil
		},
	}

	created, err := m.CreateWithVariants(context.Background(), &Item{}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || created != nil {
		t.Errorf("created = %+v, want configured func result", created)
	}
}

func TestItemVariantDAOMock_DefaultListByTemplate(t *testing.T) {
	m := ItemVariantDAOMock{}

	items, err := m.ListByTemplate(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("items = %+v, want none", items)
	}
}

func TestItemVariantDAOMock_ConfiguredListByTemplate(t *testing.T) {
	called := false
	m := ItemVariantDAOMock{
		ListByTemplateFunc: func(_ context.Context, _ uint64) ([]*ItemVariant, error) {
			called = true
			return []*ItemVariant{{}}, nil
		},
	}

	items, err := m.ListByTemplate(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(items) != 1 {
		t.Errorf("items = %+v, want configured func result", items)
	}
}

func TestItemVariantDAOMock_DefaultVariantExists(t *testing.T) {
	m := ItemVariantDAOMock{}

	found, err := m.VariantExists(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("found = false, want true")
	}
}

func TestItemVariantDAOMock_ConfiguredVariantExists(t *testing.T) {
	called := false
	m := ItemVariantDAOMock{
		VariantExistsFunc: func(_ context.Context, _ uint64) (bool, error) {
			called = true
			return false, nil
		},
	}

	found, err := m.VariantExists(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || found {
		t.Errorf("found = %v, want configured func result", found)
	}
}

func TestSupplierProductDAOMock_DefaultListInOrg(t *testing.T) {
	m := SupplierProductDAOMock{}

	page, err := m.ListInOrg(context.Background(), &query.Query{}, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page == nil || len(page.Items) != 0 {
		t.Errorf("page = %+v, want empty page", page)
	}
}

func TestSupplierProductDAOMock_DefaultFindInOrg(t *testing.T) {
	m := SupplierProductDAOMock{}

	found, err := m.FindInOrg(context.Background(), 1, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Errorf("found = %+v, want nil", found)
	}
}
