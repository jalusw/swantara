package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestEquipmentDAOMock_DefaultsAndOverrides(t *testing.T) {
	ctx := context.Background()
	entity := &Equipment{Base: model.Base{ID: 1}, Name: "Pump"}
	m := EquipmentDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Equipment], error) {
			return &query.Page[Equipment]{Items: []*Equipment{entity}, Count: 1}, nil
		},
		CreateFunc: func(_ context.Context, _ *Equipment) (*Equipment, error) { return entity, nil },
		UpdateFunc: func(_ context.Context, _ *Equipment) (*Equipment, error) { return entity, nil },
		DeleteFunc: func(_ context.Context, _ uint64) error { return errors.New("boom") },
		HardDeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("boom")
		},
	}

	list, _ := m.List(ctx, nil)
	if len(list.Items) != 1 || list.Count != 1 {
		t.Errorf("List override = %+v, want the custom page", list)
	}
	created, _ := m.Create(ctx, entity)
	if created != entity {
		t.Error("Create override not used")
	}
	updated, _ := m.Update(ctx, entity)
	if updated != entity {
		t.Error("Update override not used")
	}
	if err := m.Delete(ctx, 1); err == nil {
		t.Error("Delete override not used")
	}
	if err := m.HardDelete(ctx, 1); err == nil {
		t.Error("HardDelete override not used")
	}
	if _, err := m.Search(ctx, "name", "Pump"); err != nil {
		t.Errorf("Search = %v, want nil result", err)
	}

	defaults := EquipmentDAOMock{}
	emptyList, _ := defaults.List(ctx, nil)
	if emptyList.Count != 0 {
		t.Errorf("List default = %+v, want empty page", emptyList)
	}
	if found, _ := defaults.Find(ctx, 1); found != nil {
		t.Errorf("Find default = %+v, want nil", found)
	}
	if created, _ := defaults.Create(ctx, entity); created != entity {
		t.Error("Create default should echo the entity")
	}
	if updated, _ := defaults.Update(ctx, entity); updated != entity {
		t.Error("Update default should echo the entity")
	}
	if err := defaults.Delete(ctx, 1); err != nil {
		t.Errorf("Delete default = %v, want nil", err)
	}
	if err := defaults.HardDelete(ctx, 1); err != nil {
		t.Errorf("HardDelete default = %v, want nil", err)
	}
	if _, err := defaults.Search(ctx, "name", "Pump"); err != nil {
		t.Errorf("Search default = %v, want nil", err)
	}
}

func TestServiceContractDAOMock_DefaultsAndOverrides(t *testing.T) {
	ctx := context.Background()
	entity := &ServiceContract{Base: model.Base{ID: 1}, Name: "Gold SLA"}
	m := ServiceContractDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[ServiceContract], error) {
			return &query.Page[ServiceContract]{Items: []*ServiceContract{entity}, Count: 1}, nil
		},
		FindFunc: func(_ context.Context, _ uint64) (*ServiceContract, error) { return entity, nil },
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *ServiceContract) (*ServiceContract, error) {
			return entity, nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error { return errors.New("boom") },
		HardDeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("boom")
		},
	}

	list, _ := m.List(ctx, nil)
	if len(list.Items) != 1 {
		t.Errorf("List override = %+v, want the custom page", list)
	}
	if found, _ := m.Find(ctx, 1); found != entity {
		t.Error("Find override not used")
	}
	if updated, _ := m.UpdateTx(ctx, nil, entity); updated != entity {
		t.Error("UpdateTx override not used")
	}
	if err := m.Delete(ctx, 1); err == nil {
		t.Error("Delete override not used")
	}
	if err := m.HardDelete(ctx, 1); err == nil {
		t.Error("HardDelete override not used")
	}

	defaults := ServiceContractDAOMock{}
	emptyList, _ := defaults.List(ctx, nil)
	if emptyList.Count != 0 {
		t.Errorf("List default = %+v, want empty page", emptyList)
	}
	if found, _ := defaults.Find(ctx, 1); found != nil {
		t.Errorf("Find default = %+v, want nil", found)
	}
	if created, _ := defaults.Create(ctx, entity); created != entity {
		t.Error("Create default should echo the entity")
	}
	if updated, _ := defaults.Update(ctx, entity); updated != entity {
		t.Error("Update default should echo the entity")
	}
	if err := defaults.Delete(ctx, 1); err != nil {
		t.Errorf("Delete default = %v, want nil", err)
	}
	if err := defaults.HardDelete(ctx, 1); err != nil {
		t.Errorf("HardDelete default = %v, want nil", err)
	}
	if _, err := defaults.Search(ctx, "name", "Gold"); err != nil {
		t.Errorf("Search default = %v, want nil", err)
	}
	if updated, _ := defaults.UpdateTx(ctx, nil, entity); updated != entity {
		t.Error("UpdateTx default should echo the entity")
	}
}

func TestServiceOrderDAOMock_DefaultsAndOverrides(t *testing.T) {
	ctx := context.Background()
	entity := &ServiceOrder{Base: model.Base{ID: 1}, Name: "Fix printer"}
	m := ServiceOrderDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[ServiceOrder], error) {
			return &query.Page[ServiceOrder]{Items: []*ServiceOrder{entity}, Count: 1}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *ServiceOrder) (*ServiceOrder, error) {
			return entity, nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error { return errors.New("boom") },
		HardDeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("boom")
		},
	}

	list, _ := m.List(ctx, nil)
	if len(list.Items) != 1 {
		t.Errorf("List override = %+v, want the custom page", list)
	}
	if updated, _ := m.UpdateTx(ctx, nil, entity); updated != entity {
		t.Error("UpdateTx override not used")
	}
	if err := m.Delete(ctx, 1); err == nil {
		t.Error("Delete override not used")
	}
	if err := m.HardDelete(ctx, 1); err == nil {
		t.Error("HardDelete override not used")
	}

	defaults := ServiceOrderDAOMock{}
	emptyList, _ := defaults.List(ctx, nil)
	if emptyList.Count != 0 {
		t.Errorf("List default = %+v, want empty page", emptyList)
	}
	if found, _ := defaults.Find(ctx, 1); found != nil {
		t.Errorf("Find default = %+v, want nil", found)
	}
	if created, _ := defaults.Create(ctx, entity); created != entity {
		t.Error("Create default should echo the entity")
	}
	if updated, _ := defaults.Update(ctx, entity); updated != entity {
		t.Error("Update default should echo the entity")
	}
	if err := defaults.Delete(ctx, 1); err != nil {
		t.Errorf("Delete default = %v, want nil", err)
	}
	if err := defaults.HardDelete(ctx, 1); err != nil {
		t.Errorf("HardDelete default = %v, want nil", err)
	}
	if _, err := defaults.Search(ctx, "name", "Fix"); err != nil {
		t.Errorf("Search default = %v, want nil", err)
	}
	if updated, _ := defaults.UpdateTx(ctx, nil, entity); updated != entity {
		t.Error("UpdateTx default should echo the entity")
	}
}

func TestServiceOrderLineDAOMock_DefaultsAndOverrides(t *testing.T) {
	ctx := context.Background()
	entity := &ServiceOrderLine{Base: model.Base{ID: 1}, Type: LineTypePart}
	m := ServiceOrderLineDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[ServiceOrderLine], error) {
			return &query.Page[ServiceOrderLine]{Items: []*ServiceOrderLine{entity}, Count: 1}, nil
		},
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrderLine, error) { return entity, nil },
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *ServiceOrderLine) (*ServiceOrderLine, error) {
			return entity, nil
		},
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*ServiceOrderLine, error) {
			return []*ServiceOrderLine{entity}, nil
		},
		CreateFunc: func(_ context.Context, _ *ServiceOrderLine) (*ServiceOrderLine, error) { return entity, nil },
		UpdateFunc: func(_ context.Context, _ *ServiceOrderLine) (*ServiceOrderLine, error) { return entity, nil },
		DeleteFunc: func(_ context.Context, _ uint64) error { return errors.New("boom") },
		HardDeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("boom")
		},
	}

	list, _ := m.List(ctx, nil)
	if len(list.Items) != 1 {
		t.Errorf("List override = %+v, want the custom page", list)
	}
	if found, _ := m.Find(ctx, 1); found != entity {
		t.Error("Find override not used")
	}
	if created, _ := m.Create(ctx, entity); created != entity {
		t.Error("Create override not used")
	}
	if created, _ := m.CreateTx(ctx, nil, entity); created != entity {
		t.Error("CreateTx override not used")
	}
	if lines, _ := m.ListByOrder(ctx, 1); len(lines) != 1 {
		t.Errorf("ListByOrder override = %+v, want one line", lines)
	}
	if updated, _ := m.Update(ctx, entity); updated != entity {
		t.Error("Update override not used")
	}
	if err := m.Delete(ctx, 1); err == nil {
		t.Error("Delete override not used")
	}
	if err := m.HardDelete(ctx, 1); err == nil {
		t.Error("HardDelete override not used")
	}

	defaults := ServiceOrderLineDAOMock{}
	emptyList, _ := defaults.List(ctx, nil)
	if emptyList.Count != 0 {
		t.Errorf("List default = %+v, want empty page", emptyList)
	}
	if found, _ := defaults.Find(ctx, 1); found != nil {
		t.Errorf("Find default = %+v, want nil", found)
	}
	if created, _ := defaults.Create(ctx, entity); created != entity {
		t.Error("Create default should echo the entity")
	}
	if updated, _ := defaults.Update(ctx, entity); updated != entity {
		t.Error("Update default should echo the entity")
	}
	if err := defaults.Delete(ctx, 1); err != nil {
		t.Errorf("Delete default = %v, want nil", err)
	}
	if err := defaults.HardDelete(ctx, 1); err != nil {
		t.Errorf("HardDelete default = %v, want nil", err)
	}
	if _, err := defaults.Search(ctx, "type", "part"); err != nil {
		t.Errorf("Search default = %v, want nil", err)
	}
	if created, _ := defaults.CreateTx(ctx, nil, entity); created != entity {
		t.Error("CreateTx default should echo the entity")
	}
	if lines, _ := defaults.ListByOrder(ctx, 1); len(lines) != 0 {
		t.Errorf("ListByOrder default = %+v, want empty", lines)
	}
}

func TestMaintenancePlanDAOMock_DefaultsAndOverrides(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	entity := &MaintenancePlan{Base: model.Base{ID: 1}, Name: "Quarterly"}
	m := MaintenancePlanDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[MaintenancePlan], error) {
			return &query.Page[MaintenancePlan]{Items: []*MaintenancePlan{entity}, Count: 1}, nil
		},
		FindFunc: func(_ context.Context, _ uint64) (*MaintenancePlan, error) { return entity, nil },
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *MaintenancePlan) (*MaintenancePlan, error) {
			return entity, nil
		},
		UpdateFunc: func(_ context.Context, _ *MaintenancePlan) (*MaintenancePlan, error) { return entity, nil },
		DeleteFunc: func(_ context.Context, _ uint64) error { return errors.New("boom") },
		HardDeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("boom")
		},
	}

	list, _ := m.List(ctx, nil)
	if len(list.Items) != 1 {
		t.Errorf("List override = %+v, want the custom page", list)
	}
	if found, _ := m.Find(ctx, 1); found != entity {
		t.Error("Find override not used")
	}
	if updated, _ := m.UpdateTx(ctx, nil, entity); updated != entity {
		t.Error("UpdateTx override not used")
	}
	if updated, _ := m.Update(ctx, entity); updated != entity {
		t.Error("Update override not used")
	}
	if err := m.Delete(ctx, 1); err == nil {
		t.Error("Delete override not used")
	}
	if err := m.HardDelete(ctx, 1); err == nil {
		t.Error("HardDelete override not used")
	}

	defaults := MaintenancePlanDAOMock{}
	emptyList, _ := defaults.List(ctx, nil)
	if emptyList.Count != 0 {
		t.Errorf("List default = %+v, want empty page", emptyList)
	}
	if found, _ := defaults.Find(ctx, 1); found != nil {
		t.Errorf("Find default = %+v, want nil", found)
	}
	if created, _ := defaults.Create(ctx, entity); created != entity {
		t.Error("Create default should echo the entity")
	}
	if updated, _ := defaults.Update(ctx, entity); updated != entity {
		t.Error("Update default should echo the entity")
	}
	if err := defaults.Delete(ctx, 1); err != nil {
		t.Errorf("Delete default = %v, want nil", err)
	}
	if err := defaults.HardDelete(ctx, 1); err != nil {
		t.Errorf("HardDelete default = %v, want nil", err)
	}
	if _, err := defaults.Search(ctx, "name", "Q"); err != nil {
		t.Errorf("Search default = %v, want nil", err)
	}
	if due, _ := defaults.ListDue(ctx, 1, now); len(due) != 0 {
		t.Errorf("ListDue default = %+v, want empty", due)
	}
	if updated, _ := defaults.UpdateTx(ctx, nil, entity); updated != entity {
		t.Error("UpdateTx default should echo the entity")
	}
}

func TestServicePosterMock_DefaultsAndOverrides(t *testing.T) {
	ctx := context.Background()

	m := ServicePosterMock{
		PostFunc: func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{}, errors.New("post boom")
		},
		ReverseFunc: func(_ context.Context, _ accounting.ReverseRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{}, errors.New("reverse boom")
		},
		ReverseTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.ReverseRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{}, errors.New("reverse boom")
		},
	}
	if _, err := m.Post(ctx, accounting.PostRequest{}); err == nil {
		t.Error("Post override not used")
	}
	if _, err := m.Reverse(ctx, accounting.ReverseRequest{}); err == nil {
		t.Error("Reverse override not used")
	}
	if _, err := m.ReverseTx(ctx, nil, accounting.ReverseRequest{}); err == nil {
		t.Error("ReverseTx override not used")
	}

	defaults := ServicePosterMock{}
	if movement, _ := defaults.Post(ctx, accounting.PostRequest{}); movement == nil {
		t.Error("Post default should return an account movement")
	}
	if movement, _ := defaults.PostTx(ctx, nil, accounting.PostRequest{}); movement == nil {
		t.Error("PostTx default should delegate to Post")
	}
	if movement, _ := defaults.Reverse(ctx, accounting.ReverseRequest{}); movement == nil {
		t.Error("Reverse default should return an account movement")
	}
	if movement, _ := defaults.ReverseTx(ctx, nil, accounting.ReverseRequest{}); movement == nil {
		t.Error("ReverseTx default should delegate to Reverse")
	}
}

func TestServiceInvoiceBuilderAndTransactionerDefaultsAndOverrides(t *testing.T) {
	ctx := context.Background()

	builder := ServiceInvoiceBuilderMock{
		CreateFunc: func(_ context.Context, _ accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
			return &accounting.Invoice{}, errors.New("build boom")
		},
	}
	if _, err := builder.Create(ctx, accounting.CreateInvoiceRequest{}); err == nil {
		t.Error("Create override not used")
	}
	defaultInvoice, _ := ServiceInvoiceBuilderMock{}.Create(ctx, accounting.CreateInvoiceRequest{})
	if defaultInvoice == nil {
		t.Error("Create default should return an invoice")
	}

	ran := false
	withFn := ServiceTransactionerMock{
		RunFunc: func(_ context.Context, fn func(*gorm.DB) error) error {
			ran = true
			return fn(nil)
		},
	}
	if err := withFn.Run(ctx, func(_ *gorm.DB) error { return nil }); err != nil {
		t.Errorf("Run override = %v, want nil", err)
	}
	if !ran {
		t.Error("Run override not used")
	}
	defaultRun := ServiceTransactionerMock{}.Run(ctx, func(_ *gorm.DB) error { return nil })
	if defaultRun != nil {
		t.Errorf("Run default = %v, want nil", defaultRun)
	}
}
