package payroll

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestHRService_UpdateEmployee(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)

	tests := []struct {
		name       string
		employees  EmployeeDAOMock
		input      *Employee
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, employee *Employee, err error)
	}{
		{
			name: "updates_existing",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return &Employee{Base: model.Base{ID: 7}, OrganizationID: &organizationID, EmployeeNumber: "EMP-001"}, nil
					},
					SearchFunc: func(_ context.Context, _ string, _ any) (*Employee, error) {
						return nil, nil
					},
					UpdateFunc: func(_ context.Context, employee *Employee) (*Employee, error) {
						return employee, nil
					},
				},
			},
			input: &Employee{Base: model.Base{ID: 7}, OrganizationID: &organizationID, EmployeeNumber: "EMP-002", EmploymentType: EmploymentTypeFullTime},
			check: func(t *testing.T, employee *Employee, _ error) {
				if employee.EmployeeNumber != "EMP-002" {
					t.Fatalf("expected updated number, got %+v", employee)
				}
			},
		},
		{
			name: "rejects_missing",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return nil, nil
					},
				},
			},
			input:      &Employee{Base: model.Base{ID: 7}},
			wantErr:    true,
			wantErrVal: ErrEmployeeNotFound,
		},
		{
			name: "rejects_find_error",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return nil, errors.New("db down")
					},
				},
			},
			input:   &Employee{Base: model.Base{ID: 7}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newHRTestService(tt.employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{})
			employee, err := svc.UpdateEmployee(ctx, tt.input)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrVal) {
				return
			}
			if tt.check != nil {
				tt.check(t, employee, err)
			}
		})
	}
}

func TestHRService_CreateContract(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)

	tests := []struct {
		name       string
		employees  EmployeeDAOMock
		contracts  EmploymentContractDAOMock
		orgID      uint64
		input      *EmploymentContract
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, contract *EmploymentContract, err error)
	}{
		{
			name: "with_defaults",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return &Employee{Base: model.Base{ID: 7}, OrganizationID: &organizationID}, nil
					},
				},
			},
			contracts: EmploymentContractDAOMock{
				CRUDMock: dao.CRUDMock[EmploymentContract]{
					CreateFunc: func(_ context.Context, contract *EmploymentContract) (*EmploymentContract, error) {
						contract.ID = 4
						return contract, nil
					},
				},
			},
			orgID: organizationID,
			input: &EmploymentContract{EmployeeID: 7, Wage: 5000000},
			check: func(t *testing.T, contract *EmploymentContract, _ error) {
				if contract.ID != 4 || contract.State != ContractStateActive || contract.CurrencyCode != "IDR" || contract.WageType != WageTypeMonthly || contract.DateStart.IsZero() {
					t.Fatalf("unexpected contract: %+v", contract)
				}
			},
		},
		{
			name: "rejects_foreign_employee",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return &Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(99))}, nil
					},
				},
			},
			orgID:      1,
			input:      &EmploymentContract{EmployeeID: 7, Wage: 5000000},
			wantErr:    true,
			wantErrVal: ErrContractEmployee,
		},
		{
			name: "rejects_missing_employee",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return nil, nil
					},
				},
			},
			orgID:      1,
			input:      &EmploymentContract{EmployeeID: 7, Wage: 5000000},
			wantErr:    true,
			wantErrVal: ErrContractEmployee,
		},
		{
			name:       "rejects_zero_wage",
			employees:  EmployeeDAOMock{},
			orgID:      1,
			input:      &EmploymentContract{EmployeeID: 7, Wage: 0},
			wantErr:    true,
			wantErrVal: ErrContractWage,
		},
		{
			name:       "rejects_zero_employee_id",
			employees:  EmployeeDAOMock{},
			orgID:      1,
			input:      &EmploymentContract{EmployeeID: 0, Wage: 100},
			wantErr:    true,
			wantErrVal: ErrContractEmployee,
		},
		{
			name:       "rejects_nil",
			employees:  EmployeeDAOMock{},
			orgID:      1,
			input:      nil,
			wantErr:    true,
			wantErrVal: ErrContractNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newHRTestService(tt.employees, tt.contracts, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{})
			contract, err := svc.CreateContract(ctx, tt.orgID, tt.input)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrVal) {
				return
			}
			if tt.check != nil {
				tt.check(t, contract, err)
			}
		})
	}
}

func TestHRService_UpdateContract(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		contracts  EmploymentContractDAOMock
		input      *EmploymentContract
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, contract *EmploymentContract, err error)
	}{
		{
			name: "updates_existing",
			contracts: EmploymentContractDAOMock{
				CRUDMock: dao.CRUDMock[EmploymentContract]{
					FindFunc: func(_ context.Context, _ uint64) (*EmploymentContract, error) {
						return &EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, Wage: 4000000}, nil
					},
					UpdateFunc: func(_ context.Context, contract *EmploymentContract) (*EmploymentContract, error) {
						return contract, nil
					},
				},
			},
			input: &EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, Wage: 6000000, State: ContractStateActive},
			check: func(t *testing.T, contract *EmploymentContract, _ error) {
				if contract.Wage != 6000000 {
					t.Fatalf("expected updated wage, got %+v", contract)
				}
			},
		},
		{
			name: "rejects_missing",
			contracts: EmploymentContractDAOMock{
				CRUDMock: dao.CRUDMock[EmploymentContract]{
					FindFunc: func(_ context.Context, _ uint64) (*EmploymentContract, error) {
						return nil, nil
					},
				},
			},
			input:      &EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, Wage: 5000000},
			wantErr:    true,
			wantErrVal: ErrContractNotFound,
		},
		{
			name: "rejects_invalid_wage",
			contracts: EmploymentContractDAOMock{
				CRUDMock: dao.CRUDMock[EmploymentContract]{
					FindFunc: func(_ context.Context, _ uint64) (*EmploymentContract, error) {
						return &EmploymentContract{Base: model.Base{ID: 4}}, nil
					},
				},
			},
			input:      &EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, Wage: 0},
			wantErr:    true,
			wantErrVal: ErrContractWage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newHRTestService(EmployeeDAOMock{}, tt.contracts, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{})
			contract, err := svc.UpdateContract(ctx, tt.input)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrVal) {
				return
			}
			if tt.check != nil {
				tt.check(t, contract, err)
			}
		})
	}
}

func TestHRService_RefuseLeave(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		leaves     LeaveRequestDAOMock
		id         uint64
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, leave *LeaveRequest, err error)
	}{
		{
			name: "refuses_submitted",
			leaves: LeaveRequestDAOMock{
				CRUDMock: dao.CRUDMock[LeaveRequest]{
					FindFunc: func(_ context.Context, id uint64) (*LeaveRequest, error) {
						return &LeaveRequest{Base: model.Base{ID: id}, EmployeeID: 7, LeaveTypeID: 3, Days: 2, State: LeaveStateSubmitted}, nil
					},
					UpdateFunc: func(_ context.Context, request *LeaveRequest) (*LeaveRequest, error) {
						return request, nil
					},
				},
			},
			id: 1,
			check: func(t *testing.T, leave *LeaveRequest, _ error) {
				if leave.State != LeaveStateRefused {
					t.Fatalf("expected refused leave, got %+v", leave)
				}
			},
		},
		{
			name: "rejects_wrong_state",
			leaves: LeaveRequestDAOMock{
				CRUDMock: dao.CRUDMock[LeaveRequest]{
					FindFunc: func(_ context.Context, _ uint64) (*LeaveRequest, error) {
						return &LeaveRequest{Base: model.Base{ID: 1}, State: LeaveStateDraft}, nil
					},
				},
			},
			id:         1,
			wantErr:    true,
			wantErrVal: ErrLeaveState,
		},
		{
			name: "rejects_missing",
			leaves: LeaveRequestDAOMock{
				CRUDMock: dao.CRUDMock[LeaveRequest]{
					FindFunc: func(_ context.Context, _ uint64) (*LeaveRequest, error) {
						return nil, nil
					},
				},
			},
			id:         1,
			wantErr:    true,
			wantErrVal: ErrLeaveRequestNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newHRTestService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, tt.leaves, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{})
			leave, err := svc.RefuseLeave(ctx, tt.id)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrVal) {
				return
			}
			if tt.check != nil {
				tt.check(t, leave, err)
			}
		})
	}
}

func TestHRService_CheckIn_Extra(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	checkIn := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		employees   EmployeeDAOMock
		attendances AttendanceDAOMock
		input       *Attendance
		wantErr     bool
		wantErrVal  error
		check       func(t *testing.T, attendance *Attendance, err error)
	}{
		{
			name: "records_attendance",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return &Employee{Base: model.Base{ID: 7}, OrganizationID: &organizationID, Active: true}, nil
					},
				},
			},
			attendances: AttendanceDAOMock{
				CRUDMock: dao.CRUDMock[Attendance]{
					CreateFunc: func(_ context.Context, attendance *Attendance) (*Attendance, error) {
						attendance.ID = 1
						return attendance, nil
					},
				},
			},
			input: &Attendance{EmployeeID: 7, CheckIn: &checkIn},
			check: func(t *testing.T, attendance *Attendance, _ error) {
				if attendance.ID != 1 || attendance.WorkedHours != 0 || attendance.Status != AttendanceStatusDraft {
					t.Fatalf("unexpected attendance: %+v", attendance)
				}
				if attendance.OrganizationID != 1 {
					t.Fatalf("expected organization_id 1, got %d", attendance.OrganizationID)
				}
			},
		},
		{
			name: "rejects_missing_employee",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return nil, nil
					},
				},
			},
			input:      &Attendance{EmployeeID: 7, CheckIn: &checkIn},
			wantErr:    true,
			wantErrVal: ErrAttendanceEmployee,
		},
		{
			name: "rejects_inactive_employee",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return &Employee{Base: model.Base{ID: 7}, Active: false}, nil
					},
				},
			},
			input:      &Attendance{EmployeeID: 7, CheckIn: &checkIn},
			wantErr:    true,
			wantErrVal: ErrAttendanceInactive,
		},
		{
			name: "rejects_open_attendance",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return &Employee{Base: model.Base{ID: 7}, OrganizationID: &organizationID, Active: true}, nil
					},
				},
			},
			attendances: AttendanceDAOMock{
				FindOpenByEmployeeAndDateFunc: func(_ context.Context, _ uint64, _ string) (*Attendance, error) {
					return &Attendance{Base: model.Base{ID: 99}}, nil
				},
			},
			input:      &Attendance{EmployeeID: 7, CheckIn: &checkIn},
			wantErr:    true,
			wantErrVal: ErrAttendanceOpenExists,
		},
		{
			name:       "rejects_nil_attendance",
			employees:  EmployeeDAOMock{},
			input:      nil,
			wantErr:    true,
			wantErrVal: ErrAttendanceCheckIn,
		},
		{
			name:       "rejects_missing_check_in",
			employees:  EmployeeDAOMock{},
			input:      &Attendance{EmployeeID: 7},
			wantErr:    true,
			wantErrVal: ErrAttendanceCheckIn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newHRTestService(tt.employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, tt.attendances, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{})
			attendance, err := svc.CheckIn(ctx, tt.input)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrVal) {
				return
			}
			if tt.check != nil {
				tt.check(t, attendance, err)
			}
		})
	}
}

func TestHRService_CreateTimesheet_Extra(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		employees  EmployeeDAOMock
		input      *Timesheet
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "rejects_missing_employee",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return nil, nil
					},
				},
			},
			input:      &Timesheet{EmployeeID: 7, Hours: 8},
			wantErr:    true,
			wantErrVal: ErrTimesheetEmployee,
		},
		{
			name: "rejects_zero_hours",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return &Employee{Base: model.Base{ID: 7}}, nil
					},
				},
			},
			input:      &Timesheet{EmployeeID: 7, Hours: 0},
			wantErr:    true,
			wantErrVal: ErrTimesheetHours,
		},
		{
			name:       "rejects_nil",
			employees:  EmployeeDAOMock{},
			input:      nil,
			wantErr:    true,
			wantErrVal: ErrTimesheetNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newHRTestService(tt.employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{})
			_, err := svc.CreateTimesheet(ctx, tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
		})
	}
}

func TestHRService_LeaveBalance_Extra(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		leaves     LeaveRequestDAOMock
		leaveTypes dao.CRUDMock[reference.LeaveType]
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, balance float64, err error)
	}{
		{
			name: "floors_at_zero",
			leaves: LeaveRequestDAOMock{
				ListApprovedByEmployeeAndTypeFunc: func(_ context.Context, _, _ uint64) ([]*LeaveRequest, error) {
					return []*LeaveRequest{{Days: 4}, {Days: 2}}, nil
				},
			},
			leaveTypes: dao.CRUDMock[reference.LeaveType]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
					return &reference.LeaveType{Base: model.Base{ID: 3}, AllocationDays: helper.Ptr(5.0)}, nil
				},
			},
			check: func(t *testing.T, balance float64, _ error) {
				if balance != 0 {
					t.Fatalf("expected floored balance 0, got %v", balance)
				}
			},
		},
		{
			name: "rejects_missing_type",
			leaveTypes: dao.CRUDMock[reference.LeaveType]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
					return nil, nil
				},
			},
			wantErr:    true,
			wantErrVal: ErrLeaveTypeNotFound,
		},
		{
			name: "rejects_approved_error",
			leaveTypes: dao.CRUDMock[reference.LeaveType]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
					return &reference.LeaveType{Base: model.Base{ID: 3}}, nil
				},
			},
			leaves: LeaveRequestDAOMock{
				ListApprovedByEmployeeAndTypeFunc: func(_ context.Context, _, _ uint64) ([]*LeaveRequest, error) {
					return nil, errors.New("db down")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newHRTestService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, tt.leaves, tt.leaveTypes, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{})
			balance, err := svc.LeaveBalance(ctx, 7, 3)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrVal) {
				return
			}
			if tt.check != nil {
				tt.check(t, balance, err)
			}
		})
	}
}

func TestHRService_CreateEmployee_Extra(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)

	tests := []struct {
		name        string
		employees   EmployeeDAOMock
		departments dao.CRUDMock[reference.Department]
		positions   dao.CRUDMock[reference.JobPosition]
		input       CreateEmployeeRequest
		wantErr     bool
		wantErrVal  error
	}{
		{
			name: "rejects_missing_department",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*Employee, error) {
						return nil, nil
					},
				},
			},
			departments: dao.CRUDMock[reference.Department]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Department, error) {
					return nil, nil
				},
			},
			input: CreateEmployeeRequest{
				Employee: &Employee{OrganizationID: &organizationID, EmployeeNumber: "EMP-020", DepartmentID: helper.Ptr(uint64(50))},
				Contact:  &contacts.Contact{Name: "John Doe"},
			},
			wantErr:    true,
			wantErrVal: ErrEmployeeDepartment,
		},
		{
			name: "rejects_missing_position",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*Employee, error) {
						return nil, nil
					},
				},
			},
			positions: dao.CRUDMock[reference.JobPosition]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.JobPosition, error) {
					return nil, nil
				},
			},
			input: CreateEmployeeRequest{
				Employee: &Employee{OrganizationID: &organizationID, EmployeeNumber: "EMP-021", JobPositionID: helper.Ptr(uint64(60))},
				Contact:  &contacts.Contact{Name: "John Doe"},
			},
			wantErr:    true,
			wantErrVal: ErrEmployeePosition,
		},
		{
			name: "rejects_missing_manager",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*Employee, error) {
						return nil, nil
					},
					FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
						return nil, nil
					},
				},
			},
			input: CreateEmployeeRequest{
				Employee: &Employee{OrganizationID: &organizationID, EmployeeNumber: "EMP-022", ManagerID: helper.Ptr(uint64(70))},
				Contact:  &contacts.Contact{Name: "John Doe"},
			},
			wantErr:    true,
			wantErrVal: ErrEmployeeManager,
		},
		{
			name: "rejects_invalid_employment_type",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*Employee, error) {
						return nil, nil
					},
				},
			},
			input: CreateEmployeeRequest{
				Employee: &Employee{OrganizationID: &organizationID, EmployeeNumber: "EMP-023", EmploymentType: "freelance"},
				Contact:  &contacts.Contact{Name: "John Doe"},
			},
			wantErr:    true,
			wantErrVal: ErrEmployeeEmploymentType,
		},
		{
			name:      "rejects_missing_employee_number",
			employees: EmployeeDAOMock{},
			input: CreateEmployeeRequest{
				Employee: &Employee{OrganizationID: &organizationID},
				Contact:  &contacts.Contact{Name: "John Doe"},
			},
			wantErr:    true,
			wantErrVal: ErrEmployeeNumber,
		},
		{
			name:      "rejects_missing_organization",
			employees: EmployeeDAOMock{},
			input: CreateEmployeeRequest{
				Employee: &Employee{EmployeeNumber: "EMP-024"},
				Contact:  &contacts.Contact{Name: "John Doe"},
			},
			wantErr:    true,
			wantErrVal: ErrEmployeeOrganization,
		},
		{
			name: "rejects_search_error",
			employees: EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*Employee, error) {
						return nil, errors.New("db down")
					},
				},
			},
			input: CreateEmployeeRequest{
				Employee: &Employee{OrganizationID: &organizationID, EmployeeNumber: "EMP-025"},
				Contact:  &contacts.Contact{Name: "John Doe"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			departments := tt.departments
			if departments.FindFunc == nil {
				departments = dao.CRUDMock[reference.Department]{}
			}
			positions := tt.positions
			if positions.FindFunc == nil {
				positions = dao.CRUDMock[reference.JobPosition]{}
			}
			svc := newHRTestService(tt.employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, departments, positions, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{})
			_, err := svc.CreateEmployee(ctx, tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
		})
	}
}

func TestHRService_SubmitLeave(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		leaves     LeaveRequestDAOMock
		id         uint64
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "rejects_missing",
			leaves: LeaveRequestDAOMock{
				CRUDMock: dao.CRUDMock[LeaveRequest]{
					FindFunc: func(_ context.Context, _ uint64) (*LeaveRequest, error) {
						return nil, nil
					},
				},
			},
			id:         1,
			wantErr:    true,
			wantErrVal: ErrLeaveRequestNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newHRTestService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, tt.leaves, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{})
			_, err := svc.SubmitLeave(ctx, tt.id)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
		})
	}
}

func TestHRService_TerminateContract_Extra(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		contracts  EmploymentContractDAOMock
		id         uint64
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "rejects_missing",
			contracts: EmploymentContractDAOMock{
				CRUDMock: dao.CRUDMock[EmploymentContract]{
					FindFunc: func(_ context.Context, _ uint64) (*EmploymentContract, error) {
						return nil, nil
					},
				},
			},
			id:         4,
			wantErr:    true,
			wantErrVal: ErrContractNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newHRTestService(EmployeeDAOMock{}, tt.contracts, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{})
			_, err := svc.TerminateContract(ctx, tt.id)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
		})
	}
}

func newHRTestService(employees EmployeeDAO, contracts EmploymentContractDAO, leaves LeaveRequestDAO, leaveTypes dao.CRUD[reference.LeaveType], departments dao.CRUD[reference.Department], positions dao.CRUD[reference.JobPosition], organizations dao.CRUD[reference.Organization], contacts contacts.ContactDAO, dimensions dao.CRUD[reference.Dimension], users iam.UserDAO, attendances AttendanceDAO, timesheets TimesheetDAO, shifts ShiftDAO, shiftAssignments ShiftAssignmentDAO) HRService {
	return NewHRService(employees, contracts, leaves, leaveTypes, departments, positions, organizations, contacts, dimensions, users, attendances, timesheets, shifts, shiftAssignments, inventory.TransactionerMock{})
}
