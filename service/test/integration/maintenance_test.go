//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/service"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestMaintenancePlans_GenerateDueOrders(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	equipmentDAO := service.NewEquipmentDAO(testDB)
	equipment, err := equipmentDAO.Create(ctx, &service.Equipment{
		OrganizationID: helper.Ptr(fx.orgID),
		Name:           "Chiller",
		Category:       "HVAC",
	})
	if err != nil {
		t.Fatalf("create equipment failed: %v", err)
	}

	svc := service.NewMaintenanceService(
		service.NewMaintenancePlanDAO(testDB),
		equipmentDAO,
		service.NewServiceOrderDAO(testDB),
		db.NewDBTransactioner(testDB),
	)

	asOf := time.Now().UTC()
	pastDue := asOf.AddDate(0, 0, -10)
	plan, err := svc.CreatePlan(ctx, service.CreatePlanRequest{
		EquipmentID:  equipment.ID,
		Name:         "Quarterly service",
		IntervalDays: 30,
		NextDue:      &pastDue,
	})
	if err != nil {
		t.Fatalf("create plan failed: %v", err)
	}

	generated, err := svc.GenerateDueOrders(ctx, fx.orgID, asOf)
	if err != nil {
		t.Fatalf("generate due orders failed: %v", err)
	}
	if len(generated) != 1 {
		t.Fatalf("generated = %d, want 1", len(generated))
	}
	order := generated[0]
	if order.Type != service.OrderTypeMaintenance || order.State != service.OrderStateScheduled {
		t.Errorf("order = %+v, want maintenance / scheduled", order)
	}
	if order.EquipmentID == nil || *order.EquipmentID != equipment.ID {
		t.Errorf("order equipment = %v, want %d", order.EquipmentID, equipment.ID)
	}
	if order.ScheduledDate == nil {
		t.Errorf("order scheduled date missing")
	}

	planDAO := service.NewMaintenancePlanDAO(testDB)
	refreshed, err := planDAO.Find(ctx, plan.ID)
	if err != nil {
		t.Fatalf("find plan failed: %v", err)
	}
	if refreshed.NextDue == nil || !refreshed.NextDue.After(asOf) {
		t.Errorf("plan next due = %v, want advanced past %v", refreshed.NextDue, asOf)
	}

	again, err := svc.GenerateDueOrders(ctx, fx.orgID, asOf)
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if len(again) != 0 {
		t.Fatalf("second run generated = %d, want 0 (idempotent)", len(again))
	}

	orderDAO := service.NewServiceOrderDAO(testDB)
	all, err := orderDAO.List(ctx, &query.Query{
		Filters: []query.Filter{{Field: "equipment_id", Operator: query.Equal, Value: equipment.ID}},
	})
	if err != nil {
		t.Fatalf("list orders failed: %v", err)
	}
	if len(all.Items) != 1 {
		t.Fatalf("orders = %d, want exactly 1 for the equipment", len(all.Items))
	}
}
