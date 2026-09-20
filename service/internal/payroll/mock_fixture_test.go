package payroll

import (
	"context"
	"testing"

	"gorm.io/gorm"
)

func TestPayrollMock_Fallbacks(t *testing.T) {
	ctx := context.Background()

	t.Run("contract dao", func(t *testing.T) {
		mock := EmploymentContractDAOMock{}
		if _, err := mock.ListByEmployee(ctx, 7); err != nil {
			t.Errorf("ListByEmployee = %v", err)
		}
		if _, err := mock.FindActiveByEmployee(ctx, 7); err != nil {
			t.Errorf("FindActiveByEmployee = %v", err)
		}
	})

	t.Run("leave dao", func(t *testing.T) {
		mock := LeaveRequestDAOMock{}
		if _, err := mock.ListByEmployee(ctx, 7); err != nil {
			t.Errorf("ListByEmployee = %v", err)
		}
		if _, err := mock.ListApprovedByEmployeeAndType(ctx, 7, 3); err != nil {
			t.Errorf("ListApproved = %v", err)
		}
		if _, err := mock.ListApprovedByEmployeeBetweenDates(ctx, 7, "a", "b"); err != nil {
			t.Errorf("ListBetween = %v", err)
		}
	})

	t.Run("attendance dao", func(t *testing.T) {
		mock := AttendanceDAOMock{}
		if _, err := mock.ListByEmployee(ctx, 7); err != nil {
			t.Errorf("ListByEmployee = %v", err)
		}
		if _, err := mock.FindOpenByEmployeeAndDate(ctx, 7, "2026-08-01"); err != nil {
			t.Errorf("FindOpen = %v", err)
		}
		if _, err := mock.ListMissingByDate(ctx, 10, "2026-08-01"); err != nil {
			t.Errorf("ListMissing = %v", err)
		}
		if _, err := mock.ListByEmployeeAndDateRange(ctx, 7, "a", "b"); err != nil {
			t.Errorf("ListRange = %v", err)
		}
		if _, err := mock.ListEmployeeIDsByDate(ctx, 10, "2026-08-01"); err != nil {
			t.Errorf("ListIDs = %v", err)
		}
	})

	t.Run("timesheet dao", func(t *testing.T) {
		mock := TimesheetDAOMock{}
		if _, err := mock.ListByEmployee(ctx, 7); err != nil {
			t.Errorf("ListByEmployee = %v", err)
		}
		if _, err := mock.ListByProject(ctx, 9); err != nil {
			t.Errorf("ListByProject = %v", err)
		}
	})

	t.Run("run dao", func(t *testing.T) {
		mock := PayrollRunDAOMock{}
		run := &PayrollRun{}
		if _, err := mock.CreateWithPayslipsTx(ctx, nil, run, nil, nil); err != nil {
			t.Errorf("CreateWithPayslipsTx = %v", err)
		}
		if _, err := mock.UpdateTx(ctx, nil, &PayrollRun{}); err != nil {
			t.Errorf("UpdateTx = %v", err)
		}
	})

	t.Run("payslip dao", func(t *testing.T) {
		mock := PayslipDAOMock{}
		if _, err := mock.ListByRun(ctx, 1); err != nil {
			t.Errorf("ListByRun = %v", err)
		}
		if _, err := mock.ListByEmployee(ctx, 7); err != nil {
			t.Errorf("ListByEmployee = %v", err)
		}
		if _, err := mock.UpdateTx(ctx, nil, &Payslip{}); err != nil {
			t.Errorf("UpdateTx = %v", err)
		}
	})

	t.Run("line and assignment dao", func(t *testing.T) {
		if _, err := (PayslipLineDAOMock{}).ListByPayslip(ctx, 1); err != nil {
			t.Errorf("ListByPayslip = %v", err)
		}
		if _, err := (ShiftAssignmentDAOMock{}).FindByEmployeeAndDate(ctx, 7, "2026-08-01"); err != nil {
			t.Errorf("FindByEmployeeAndDate = %v", err)
		}
		if _, err := (AttendanceDAOWithOrgMock{}).ListOpenByOrganization(ctx, 10); err != nil {
			t.Errorf("ListOpenByOrganization = %v", err)
		}
	})

	t.Run("with func variants", func(t *testing.T) {
		contracts := EmploymentContractDAOMock{
			ListByEmployeeFunc: func(_ context.Context, _ uint64) ([]*EmploymentContract, error) {
				return []*EmploymentContract{{}}, nil
			},
			FindActiveByEmployeeFunc: func(_ context.Context, _ uint64) (*EmploymentContract, error) {
				return &EmploymentContract{}, nil
			},
		}
		if _, err := contracts.ListByEmployee(ctx, 7); err != nil {
			t.Errorf("ListByEmployee = %v", err)
		}
		if _, err := contracts.FindActiveByEmployee(ctx, 7); err != nil {
			t.Errorf("FindActiveByEmployee = %v", err)
		}

		leaves := LeaveRequestDAOMock{
			ListByEmployeeFunc: func(_ context.Context, _ uint64) ([]*LeaveRequest, error) {
				return []*LeaveRequest{{}}, nil
			},
			ListApprovedByEmployeeAndTypeFunc: func(_ context.Context, _, _ uint64) ([]*LeaveRequest, error) {
				return []*LeaveRequest{{}}, nil
			},
			ListApprovedByEmployeeBetweenDatesFunc: func(_ context.Context, _ uint64, _, _ string) ([]*LeaveRequest, error) {
				return []*LeaveRequest{{}}, nil
			},
		}
		if _, err := leaves.ListByEmployee(ctx, 7); err != nil {
			t.Errorf("ListByEmployee = %v", err)
		}
		if _, err := leaves.ListApprovedByEmployeeAndType(ctx, 7, 3); err != nil {
			t.Errorf("ListApproved = %v", err)
		}
		if _, err := leaves.ListApprovedByEmployeeBetweenDates(ctx, 7, "a", "b"); err != nil {
			t.Errorf("ListBetween = %v", err)
		}

		runs := PayrollRunDAOMock{
			CreateWithPayslipsTxFunc: func(_ context.Context, _ *gorm.DB, r *PayrollRun, _ []*Payslip, _ [][]*PayslipLine) (*PayrollRun, error) {
				return r, nil
			},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, r *PayrollRun) (*PayrollRun, error) { return r, nil },
		}
		if _, err := runs.CreateWithPayslipsTx(ctx, nil, &PayrollRun{}, nil, nil); err != nil {
			t.Errorf("CreateWithPayslipsTx = %v", err)
		}
		if _, err := runs.UpdateTx(ctx, nil, &PayrollRun{}); err != nil {
			t.Errorf("UpdateTx = %v", err)
		}
	})
}

func TestPayrollFixtures_Construct(t *testing.T) {
	if EmployeeFixture() == nil {
		t.Error("employee = nil")
	}
	if EmploymentContractFixture() == nil {
		t.Error("contract = nil")
	}
	if LeaveRequestFixture() == nil {
		t.Error("leave = nil")
	}
	if AttendanceFixture() == nil {
		t.Error("attendance = nil")
	}
	if TimesheetFixture() == nil {
		t.Error("timesheet = nil")
	}
	if PayrollRunFixture() == nil {
		t.Error("run = nil")
	}
	if PayslipFixture() == nil {
		t.Error("payslip = nil")
	}
	if PayslipLineFixture() == nil {
		t.Error("line = nil")
	}
	if ShiftFixture() == nil {
		t.Error("shift = nil")
	}
	if ShiftAssignmentFixture() == nil {
		t.Error("assignment = nil")
	}
	if EmployeeFixture(func(e *Employee) *Employee { e.Active = true; return e }) == nil {
		t.Error("employee opt = nil")
	}
}
