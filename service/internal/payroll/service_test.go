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
	"gorm.io/gorm"
)

func TestHRService_CreateEmployee(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) HRService
		input   CreateEmployeeRequest
		wantErr bool
		wantVal error
		check   func(t *testing.T, emp *Employee)
	}{
		{
			name: "rejects duplicate employee number",
			setup: func(t *testing.T) HRService {
				organizationID := uint64(1)
				existing := &Employee{Base: model.Base{ID: 2}, EmployeeNumber: "EMP-001", OrganizationID: &organizationID}
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						CreateFunc: func(_ context.Context, employee *Employee) (*Employee, error) {
							employee.ID = 1
							return employee, nil
						},
						SearchFunc: func(_ context.Context, field string, value any) (*Employee, error) {
							if field == "employee_number" {
								return existing, nil
							}
							return nil, nil
						},
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input: func() CreateEmployeeRequest {
				organizationID := uint64(1)
				return CreateEmployeeRequest{
					Employee: &Employee{OrganizationID: &organizationID, EmployeeNumber: "EMP-001"},
					Contact:  &contacts.Contact{Name: "John Doe"},
				}
			}(),
			wantErr: true,
			wantVal: ErrEmployeeNumberTaken,
		},
		{
			name: "rejects taken user",
			setup: func(t *testing.T) HRService {
				organizationID := uint64(1)
				userID := uint64(9)
				linked := &Employee{Base: model.Base{ID: 2}, UserID: &userID, EmployeeNumber: "EMP-002", OrganizationID: &organizationID}
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						CreateFunc: func(_ context.Context, employee *Employee) (*Employee, error) {
							employee.ID = 1
							return employee, nil
						},
						SearchFunc: func(_ context.Context, field string, value any) (*Employee, error) {
							if field == "employee_number" {
								return nil, nil
							}
							return linked, nil
						},
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{DAOMock: iam.DAOMock[iam.User]{FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) { return &iam.User{}, nil }}}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input: func() CreateEmployeeRequest {
				organizationID := uint64(1)
				userID := uint64(9)
				return CreateEmployeeRequest{
					Employee: &Employee{OrganizationID: &organizationID, EmployeeNumber: "EMP-003", UserID: &userID},
					Contact:  &contacts.Contact{Name: "Jane Doe"},
				}
			}(),
			wantErr: true,
			wantVal: ErrEmployeeUserTaken,
		},
		{
			name: "rejects missing user",
			setup: func(t *testing.T) HRService {
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						CreateFunc: func(_ context.Context, employee *Employee) (*Employee, error) {
							employee.ID = 1
							return employee, nil
						},
						SearchFunc: func(_ context.Context, _ string, _ any) (*Employee, error) {
							return nil, nil
						},
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input: func() CreateEmployeeRequest {
				organizationID := uint64(1)
				userID := uint64(99)
				return CreateEmployeeRequest{
					Employee: &Employee{OrganizationID: &organizationID, EmployeeNumber: "EMP-012", UserID: &userID},
					Contact:  &contacts.Contact{Name: "John Doe"},
				}
			}(),
			wantErr: true,
			wantVal: ErrEmployeeUserNotFound,
		},
		{
			name: "creates contact and contract",
			setup: func(t *testing.T) HRService {
				var createdContact *contacts.Contact
				var createdContract *EmploymentContract
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						CreateFunc: func(_ context.Context, employee *Employee) (*Employee, error) {
							employee.ID = 7
							return employee, nil
						},
						SearchFunc: func(_ context.Context, _ string, _ any) (*Employee, error) {
							return nil, nil
						},
					},
				}
				contactDAO := contacts.ContactDAOMock{
					CreateWithDetailsFunc: func(_ context.Context, contact *contacts.Contact, _ []*contacts.ContactAddress, _ []*contacts.ContactBankAccount, _ *contacts.CustomerProfile, _ *contacts.SupplierProfile) (*contacts.Contact, error) {
						contact.ID = 3
						createdContact = contact
						return contact, nil
					},
				}
				contracts := EmploymentContractDAOMock{
					CRUDMock: dao.CRUDMock[EmploymentContract]{
						CreateFunc: func(_ context.Context, contract *EmploymentContract) (*EmploymentContract, error) {
							contract.ID = 4
							createdContract = contract
							return contract, nil
						},
					},
				}
				svc := NewHRService(employees, contracts, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contactDAO, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
				t.Cleanup(func() {
					if createdContact == nil || createdContact.Name != "John Smith" {
						t.Errorf("expected contact to be created, got %+v", createdContact)
					}
					if createdContract == nil || createdContract.EmployeeID != 7 || createdContract.State != ContractStateActive {
						t.Errorf("expected active contract for employee, got %+v", createdContract)
					}
				})
				return svc
			},
			input: func() CreateEmployeeRequest {
				organizationID := uint64(1)
				return CreateEmployeeRequest{
					Employee: &Employee{OrganizationID: &organizationID, EmployeeNumber: "EMP-010"},
					Contract: &EmploymentContract{Wage: 5000000},
					Contact:  &contacts.Contact{Name: "John Smith"},
				}
			}(),
			check: func(t *testing.T, emp *Employee) {
				if emp.ID != 7 || emp.ContactID != 3 {
					t.Fatalf("expected employee with contact link, got %+v", emp)
				}
			},
		},
		{
			name: "provisions user link",
			setup: func(t *testing.T) HRService {
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						CreateFunc: func(_ context.Context, employee *Employee) (*Employee, error) {
							employee.ID = 7
							return employee, nil
						},
						SearchFunc: func(_ context.Context, _ string, _ any) (*Employee, error) {
							return nil, nil
						},
					},
				}
				contactDAO := contacts.ContactDAOMock{
					CreateWithDetailsFunc: func(_ context.Context, contact *contacts.Contact, _ []*contacts.ContactAddress, _ []*contacts.ContactBankAccount, _ *contacts.CustomerProfile, _ *contacts.SupplierProfile) (*contacts.Contact, error) {
						contact.ID = 3
						return contact, nil
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contactDAO, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{DAOMock: iam.DAOMock[iam.User]{FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) { return &iam.User{}, nil }}}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input: func() CreateEmployeeRequest {
				organizationID := uint64(1)
				userID := uint64(9)
				return CreateEmployeeRequest{
					Employee: &Employee{OrganizationID: &organizationID, EmployeeNumber: "EMP-011", UserID: &userID},
					Contact:  &contacts.Contact{Name: "John Doe"},
				}
			}(),
			check: func(t *testing.T, emp *Employee) {
				userID := uint64(9)
				if emp.UserID == nil || *emp.UserID != userID {
					t.Fatalf("expected user link provisioned, got %+v", emp)
				}
				if emp.ID != 7 {
					t.Fatalf("expected created employee id 7, got %d", emp.ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			emp, err := svc.CreateEmployee(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantVal)
			if tt.check != nil {
				tt.check(t, emp)
			}
		})
	}
}

func TestHRService_LeaveWorkflow(t *testing.T) {
	ctx := context.Background()
	requests := map[uint64]*LeaveRequest{
		1: {Base: model.Base{ID: 1}, EmployeeID: 7, LeaveTypeID: 3, Days: 2, State: LeaveStateDraft},
	}
	leaves := LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[LeaveRequest]{
			FindFunc: func(_ context.Context, id uint64) (*LeaveRequest, error) {
				return requests[id], nil
			},
			UpdateFunc: func(_ context.Context, request *LeaveRequest) (*LeaveRequest, error) {
				requests[request.ID] = request
				return request, nil
			},
		},
		ListApprovedByEmployeeAndTypeFunc: func(_ context.Context, _, _ uint64) ([]*LeaveRequest, error) {
			return []*LeaveRequest{}, nil
		},
	}
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return &reference.LeaveType{Base: model.Base{ID: 3}, AllocationDays: helper.Ptr(12.0)}, nil
		},
	}
	svc := NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, leaves, leaveTypes, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})

	if _, err := svc.SubmitLeave(ctx, 1); helper.AssertError(t, err, false, nil) {
		return
	}
	if requests[1].State != LeaveStateSubmitted {
		t.Fatalf("expected submitted leave, got %s", requests[1].State)
	}

	if _, err := svc.ApproveLeave(ctx, 1); helper.AssertError(t, err, false, nil) {
		return
	}
	if requests[1].State != LeaveStateApproved {
		t.Fatalf("expected approved leave, got %s", requests[1].State)
	}

	if _, err := svc.SubmitLeave(ctx, 1); !helper.AssertError(t, err, true, ErrLeaveState) {
		return
	}
}

func TestHRService_ApproveLeave(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) HRService
		input   uint64
		wantErr bool
		wantVal error
	}{
		{
			name: "rejects excess balance",
			setup: func(t *testing.T) HRService {
				requests := map[uint64]*LeaveRequest{
					1: {Base: model.Base{ID: 1}, EmployeeID: 7, LeaveTypeID: 3, Days: 15, State: LeaveStateSubmitted},
				}
				leaves := LeaveRequestDAOMock{
					CRUDMock: dao.CRUDMock[LeaveRequest]{
						FindFunc: func(_ context.Context, id uint64) (*LeaveRequest, error) {
							return requests[id], nil
						},
						UpdateFunc: func(_ context.Context, request *LeaveRequest) (*LeaveRequest, error) {
							requests[request.ID] = request
							return request, nil
						},
					},
					ListApprovedByEmployeeAndTypeFunc: func(_ context.Context, _, _ uint64) ([]*LeaveRequest, error) {
						return []*LeaveRequest{}, nil
					},
				}
				leaveTypes := dao.CRUDMock[reference.LeaveType]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
						return &reference.LeaveType{Base: model.Base{ID: 3}, AllocationDays: helper.Ptr(12.0)}, nil
					},
				}
				return NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, leaves, leaveTypes, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input:   1,
			wantErr: true,
			wantVal: ErrLeaveBalance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			_, err := svc.ApproveLeave(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantVal)
		})
	}
}

func TestHRService_LeaveBalance(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) HRService
		employeeID uint64
		leaveType  uint64
		want       float64
		wantErr    bool
		wantVal    error
	}{
		{
			name: "honors approved leave",
			setup: func(t *testing.T) HRService {
				leaves := LeaveRequestDAOMock{
					ListApprovedByEmployeeAndTypeFunc: func(_ context.Context, _, _ uint64) ([]*LeaveRequest, error) {
						return []*LeaveRequest{{Days: 4}, {Days: 2}}, nil
					},
				}
				leaveTypes := dao.CRUDMock[reference.LeaveType]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
						return &reference.LeaveType{Base: model.Base{ID: 3}, AllocationDays: helper.Ptr(10.0)}, nil
					},
				}
				return NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, leaves, leaveTypes, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			employeeID: 7,
			leaveType:  3,
			want:       4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			balance, err := svc.LeaveBalance(context.Background(), tt.employeeID, tt.leaveType)
			helper.AssertError(t, err, tt.wantErr, tt.wantVal)
			if balance != tt.want {
				t.Fatalf("expected balance %v, got %v", tt.want, balance)
			}
		})
	}
}

func TestHRService_CheckOut(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) HRService
		id       uint64
		checkOut time.Time
		breakMin int
		wantErr  bool
		wantVal  error
		check    func(t *testing.T, record *Attendance)
	}{
		{
			name: "computes worked hours",
			setup: func(t *testing.T) HRService {
				checkIn := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
				attendance := &Attendance{Base: model.Base{ID: 1}, EmployeeID: 7, CheckIn: &checkIn}
				attendances := AttendanceDAOMock{
					CRUDMock: dao.CRUDMock[Attendance]{
						FindFunc: func(_ context.Context, _ uint64) (*Attendance, error) {
							return attendance, nil
						},
						UpdateFunc: func(_ context.Context, record *Attendance) (*Attendance, error) {
							attendance = record
							return record, nil
						},
					},
				}
				return NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			id:       1,
			checkOut: time.Date(2026, 8, 1, 17, 30, 0, 0, time.UTC),
			breakMin: 0,
			check: func(t *testing.T, record *Attendance) {
				if record.WorkedHours != 8.5 {
					t.Fatalf("expected 8.5 worked hours, got %v", record.WorkedHours)
				}
				if record.Status != AttendanceStatusConfirmed {
					t.Fatalf("expected confirmed status, got %s", record.Status)
				}
				if record.OvertimeHours != 0.5 {
					t.Fatalf("expected 0.5 overtime hours, got %v", record.OvertimeHours)
				}
			},
		},
		{
			name: "computes worked hours with break",
			setup: func(t *testing.T) HRService {
				checkIn := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
				attendance := &Attendance{Base: model.Base{ID: 1}, EmployeeID: 7, CheckIn: &checkIn}
				attendances := AttendanceDAOMock{
					CRUDMock: dao.CRUDMock[Attendance]{
						FindFunc: func(_ context.Context, _ uint64) (*Attendance, error) {
							return attendance, nil
						},
						UpdateFunc: func(_ context.Context, record *Attendance) (*Attendance, error) {
							attendance = record
							return record, nil
						},
					},
				}
				return NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			id:       1,
			checkOut: time.Date(2026, 8, 1, 17, 30, 0, 0, time.UTC),
			breakMin: 60,
			check: func(t *testing.T, record *Attendance) {
				if record.WorkedHours != 7.5 {
					t.Fatalf("expected 7.5 worked hours with 60min break, got %v", record.WorkedHours)
				}
				if record.BreakMinutes != 60 {
					t.Fatalf("expected 60 break minutes, got %d", record.BreakMinutes)
				}
				if record.OvertimeHours != 0 {
					t.Fatalf("expected 0 overtime hours, got %v", record.OvertimeHours)
				}
			},
		},
		{
			name: "rejects already closed",
			setup: func(t *testing.T) HRService {
				checkIn := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
				checkOut := time.Date(2026, 8, 1, 17, 0, 0, 0, time.UTC)
				attendances := AttendanceDAOMock{
					CRUDMock: dao.CRUDMock[Attendance]{
						FindFunc: func(_ context.Context, _ uint64) (*Attendance, error) {
							return &Attendance{Base: model.Base{ID: 1}, EmployeeID: 7, CheckIn: &checkIn, CheckOut: &checkOut}, nil
						},
					},
				}
				return NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			id:       1,
			checkOut: time.Date(2026, 8, 1, 18, 0, 0, 0, time.UTC),
			breakMin: 0,
			wantErr:  true,
			wantVal:  ErrAttendanceAlreadyClosed,
		},
		{
			name: "rejects before check in",
			setup: func(t *testing.T) HRService {
				checkIn := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
				attendances := AttendanceDAOMock{
					CRUDMock: dao.CRUDMock[Attendance]{
						FindFunc: func(_ context.Context, _ uint64) (*Attendance, error) {
							return &Attendance{Base: model.Base{ID: 1}, EmployeeID: 7, CheckIn: &checkIn}, nil
						},
					},
				}
				return NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			id:       1,
			checkOut: time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC),
			breakMin: 0,
			wantErr:  true,
			wantVal:  ErrAttendanceCheckOut,
		},
		{
			name: "sets pending approval when required",
			setup: func(t *testing.T) HRService {
				checkIn := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
				attendance := &Attendance{Base: model.Base{ID: 1}, EmployeeID: 7, CheckIn: &checkIn}
				attendances := AttendanceDAOMock{
					CRUDMock: dao.CRUDMock[Attendance]{
						FindFunc: func(_ context.Context, _ uint64) (*Attendance, error) {
							return attendance, nil
						},
						UpdateFunc: func(_ context.Context, record *Attendance) (*Attendance, error) {
							attendance = record
							return record, nil
						},
					},
				}
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
							return &Employee{Base: model.Base{ID: 7}, RequiresApproval: true}, nil
						},
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			id:       1,
			checkOut: time.Date(2026, 8, 1, 17, 0, 0, 0, time.UTC),
			breakMin: 0,
			check: func(t *testing.T, record *Attendance) {
				if record.Status != AttendanceStatusPendingApproval {
					t.Fatalf("expected pending_approval status, got %s", record.Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			record, err := svc.CheckOut(context.Background(), tt.id, tt.checkOut, tt.breakMin)
			helper.AssertError(t, err, tt.wantErr, tt.wantVal)
			if tt.check != nil {
				tt.check(t, record)
			}
		})
	}
}

func TestHRService_CreateTimesheet(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) HRService
		input   *Timesheet
		wantErr bool
		wantVal error
		check   func(t *testing.T, ts *Timesheet)
	}{
		{
			name: "rejects unknown dimension account",
			setup: func(t *testing.T) HRService {
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
							return &Employee{Base: model.Base{ID: 7}}, nil
						},
					},
				}
				dimensions := dao.CRUDMock[reference.Dimension]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Dimension, error) {
						return nil, nil
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dimensions, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input: func() *Timesheet {
				dimensionID := uint64(99)
				return &Timesheet{EmployeeID: 7, Hours: 8, DimensionID: &dimensionID}
			}(),
			wantErr: true,
			wantVal: ErrTimesheetDimension,
		},
		{
			name: "tags dimension account",
			setup: func(t *testing.T) HRService {
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
							return &Employee{Base: model.Base{ID: 7}}, nil
						},
					},
				}
				dimensions := dao.CRUDMock[reference.Dimension]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Dimension, error) {
						return &reference.Dimension{Base: model.Base{ID: 5}}, nil
					},
				}
				timesheets := TimesheetDAOMock{
					CRUDMock: dao.CRUDMock[Timesheet]{
						CreateFunc: func(_ context.Context, timesheet *Timesheet) (*Timesheet, error) {
							timesheet.ID = 2
							return timesheet, nil
						},
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dimensions, iam.UserDAOMock{}, AttendanceDAOMock{}, timesheets, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input: func() *Timesheet {
				dimensionID := uint64(5)
				return &Timesheet{EmployeeID: 7, Hours: 8, DimensionID: &dimensionID}
			}(),
			check: func(t *testing.T, ts *Timesheet) {
				dimensionID := uint64(5)
				if ts.DimensionID == nil || *ts.DimensionID != dimensionID {
					t.Fatalf("expected dimension tagging, got %+v", ts)
				}
				if ts.ID != 2 {
					t.Fatalf("expected created timesheet id 2, got %d", ts.ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			ts, err := svc.CreateTimesheet(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantVal)
			if tt.check != nil {
				tt.check(t, ts)
			}
		})
	}
}

func TestHRService_CreateLeaveRequest(t *testing.T) {
	from := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		setup   func(t *testing.T) HRService
		input   *LeaveRequest
		wantErr bool
		wantVal error
		check   func(t *testing.T, lr *LeaveRequest)
	}{
		{
			name: "rejects invalid date range",
			setup: func(t *testing.T) HRService {
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
							return &Employee{Base: model.Base{ID: 7}}, nil
						},
					},
				}
				leaveTypes := dao.CRUDMock[reference.LeaveType]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
						return &reference.LeaveType{Base: model.Base{ID: 3}}, nil
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, leaveTypes, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input: &LeaveRequest{
				EmployeeID:  7,
				LeaveTypeID: 3,
				DateFrom:    time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
				DateTo:      time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC),
				Days:        2,
			},
			wantErr: true,
			wantVal: ErrLeaveDateRange,
		},
		{
			name: "rejects zero days",
			setup: func(t *testing.T) HRService {
				return NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input: &LeaveRequest{
				EmployeeID:  7,
				LeaveTypeID: 3,
				DateFrom:    from,
				DateTo:      to,
				Days:        0,
			},
			wantErr: true,
			wantVal: ErrLeaveDays,
		},
		{
			name: "rejects unknown employee",
			setup: func(t *testing.T) HRService {
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
							return nil, nil
						},
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input: &LeaveRequest{
				EmployeeID:  7,
				LeaveTypeID: 3,
				DateFrom:    from,
				DateTo:      to,
				Days:        2,
			},
			wantErr: true,
			wantVal: ErrLeaveEmployee,
		},
		{
			name: "rejects unknown leave type",
			setup: func(t *testing.T) HRService {
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
							return &Employee{Base: model.Base{ID: 7}}, nil
						},
					},
				}
				leaveTypes := dao.CRUDMock[reference.LeaveType]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
						return nil, nil
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, leaveTypes, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input: &LeaveRequest{
				EmployeeID:  7,
				LeaveTypeID: 3,
				DateFrom:    from,
				DateTo:      to,
				Days:        2,
			},
			wantErr: true,
			wantVal: ErrLeaveTypeNotFound,
		},
		{
			name: "creates draft leave request",
			setup: func(t *testing.T) HRService {
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
							return &Employee{Base: model.Base{ID: 7}}, nil
						},
					},
				}
				leaveTypes := dao.CRUDMock[reference.LeaveType]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
						return &reference.LeaveType{Base: model.Base{ID: 3}}, nil
					},
				}
				leaves := LeaveRequestDAOMock{
					CRUDMock: dao.CRUDMock[LeaveRequest]{
						CreateFunc: func(_ context.Context, request *LeaveRequest) (*LeaveRequest, error) {
							request.ID = 9
							return request, nil
						},
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, leaves, leaveTypes, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input: &LeaveRequest{
				EmployeeID:  7,
				LeaveTypeID: 3,
				DateFrom:    from,
				DateTo:      to,
				Days:        2,
			},
			check: func(t *testing.T, lr *LeaveRequest) {
				if lr.ID != 9 || lr.State != LeaveStateDraft {
					t.Fatalf("expected draft leave created, got %+v", lr)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			lr, err := svc.CreateLeaveRequest(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantVal)
			if tt.check != nil {
				tt.check(t, lr)
			}
		})
	}
}

func TestHRService_TerminateContract(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) HRService
		id      uint64
		wantErr bool
		wantVal error
		check   func(t *testing.T, c *EmploymentContract)
	}{
		{
			name: "closes active contract",
			setup: func(t *testing.T) HRService {
				contracts := EmploymentContractDAOMock{
					CRUDMock: dao.CRUDMock[EmploymentContract]{
						FindFunc: func(_ context.Context, _ uint64) (*EmploymentContract, error) {
							return &EmploymentContract{Base: model.Base{ID: 4}, State: ContractStateActive}, nil
						},
						UpdateFunc: func(_ context.Context, contract *EmploymentContract) (*EmploymentContract, error) {
							return contract, nil
						},
					},
				}
				return NewHRService(EmployeeDAOMock{}, contracts, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			id: 4,
			check: func(t *testing.T, c *EmploymentContract) {
				if c.State != ContractStateClosed {
					t.Fatalf("expected closed contract, got %s", c.State)
				}
			},
		},
		{
			name: "rejects non-active contract",
			setup: func(t *testing.T) HRService {
				contracts := EmploymentContractDAOMock{
					CRUDMock: dao.CRUDMock[EmploymentContract]{
						FindFunc: func(_ context.Context, _ uint64) (*EmploymentContract, error) {
							return &EmploymentContract{Base: model.Base{ID: 4}, State: ContractStateClosed}, nil
						},
					},
				}
				return NewHRService(EmployeeDAOMock{}, contracts, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			id:      4,
			wantErr: true,
			wantVal: ErrContractState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			c, err := svc.TerminateContract(context.Background(), tt.id)
			helper.AssertError(t, err, tt.wantErr, tt.wantVal)
			if tt.check != nil {
				tt.check(t, c)
			}
		})
	}
}

func TestHRService_ApproveAttendance(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) HRService
		id      uint64
		wantErr bool
		wantVal error
		check   func(t *testing.T, a *Attendance)
	}{
		{
			name: "confirms pending attendance",
			setup: func(t *testing.T) HRService {
				attendance := &Attendance{Base: model.Base{ID: 1}, EmployeeID: 7, Status: AttendanceStatusPendingApproval}
				attendances := AttendanceDAOMock{
					CRUDMock: dao.CRUDMock[Attendance]{
						FindFunc: func(_ context.Context, _ uint64) (*Attendance, error) {
							return attendance, nil
						},
						UpdateFunc: func(_ context.Context, record *Attendance) (*Attendance, error) {
							attendance = record
							return record, nil
						},
					},
				}
				return NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			id: 1,
			check: func(t *testing.T, a *Attendance) {
				if a.Status != AttendanceStatusConfirmed {
					t.Fatalf("expected confirmed status, got %s", a.Status)
				}
			},
		},
		{
			name: "rejects non-pending attendance",
			setup: func(t *testing.T) HRService {
				attendance := &Attendance{Base: model.Base{ID: 1}, EmployeeID: 7, Status: AttendanceStatusConfirmed}
				attendances := AttendanceDAOMock{
					CRUDMock: dao.CRUDMock[Attendance]{
						FindFunc: func(_ context.Context, _ uint64) (*Attendance, error) {
							return attendance, nil
						},
					},
				}
				return NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			id:      1,
			wantErr: true,
			wantVal: ErrAttendanceNotPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			a, err := svc.ApproveAttendance(context.Background(), tt.id)
			helper.AssertError(t, err, tt.wantErr, tt.wantVal)
			if tt.check != nil {
				tt.check(t, a)
			}
		})
	}
}

func TestHRService_RejectAttendance(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) HRService
		id      uint64
		wantErr bool
		wantVal error
		check   func(t *testing.T, a *Attendance)
	}{
		{
			name: "cancels pending attendance",
			setup: func(t *testing.T) HRService {
				attendance := &Attendance{Base: model.Base{ID: 1}, EmployeeID: 7, Status: AttendanceStatusPendingApproval}
				attendances := AttendanceDAOMock{
					CRUDMock: dao.CRUDMock[Attendance]{
						FindFunc: func(_ context.Context, _ uint64) (*Attendance, error) {
							return attendance, nil
						},
						UpdateFunc: func(_ context.Context, record *Attendance) (*Attendance, error) {
							attendance = record
							return record, nil
						},
					},
				}
				return NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			id: 1,
			check: func(t *testing.T, a *Attendance) {
				if a.Status != AttendanceStatusCancelled {
					t.Fatalf("expected cancelled status, got %s", a.Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			a, err := svc.RejectAttendance(context.Background(), tt.id)
			helper.AssertError(t, err, tt.wantErr, tt.wantVal)
			if tt.check != nil {
				tt.check(t, a)
			}
		})
	}
}

func TestHRService_CreateShift(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) HRService
		input   *Shift
		wantErr bool
		wantVal error
	}{
		{
			name: "rejects empty name",
			setup: func(t *testing.T) HRService {
				return NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
			},
			input:   &Shift{Name: ""},
			wantErr: true,
			wantVal: ErrShiftName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			_, err := svc.CreateShift(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantVal)
		})
	}
}

func TestHRService_CreateShiftAssignment(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) HRService
		input   *ShiftAssignment
		wantErr bool
		wantVal error
	}{
		{
			name: "rejects duplicate assignment",
			setup: func(t *testing.T) HRService {
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
							return &Employee{Base: model.Base{ID: 7}}, nil
						},
					},
				}
				shifts := ShiftDAOMock{
					CRUDMock: dao.CRUDMock[Shift]{
						FindFunc: func(_ context.Context, _ uint64) (*Shift, error) {
							return &Shift{Base: model.Base{ID: 1}}, nil
						},
					},
				}
				assignments := ShiftAssignmentDAOMock{
					CRUDMock: dao.CRUDMock[ShiftAssignment]{},
					FindByEmployeeAndDateFunc: func(_ context.Context, _ uint64, _ string) (*ShiftAssignment, error) {
						return &ShiftAssignment{Base: model.Base{ID: 1}}, nil
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, shifts, assignments, inventory.TransactionerMock{})
			},
			input: &ShiftAssignment{
				EmployeeID: 7,
				ShiftID:    1,
				Date:       time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
			wantVal: ErrShiftAssignmentExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			_, err := svc.CreateShiftAssignment(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantVal)
		})
	}
}

func TestHRService_CheckIn(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) HRService
		input   *Attendance
		wantErr bool
		wantVal error
		check   func(t *testing.T, a *Attendance)
	}{
		{
			name: "links shift assignment to attendance",
			setup: func(t *testing.T) HRService {
				attendances := AttendanceDAOMock{
					CRUDMock: dao.CRUDMock[Attendance]{
						CreateFunc: func(_ context.Context, record *Attendance) (*Attendance, error) {
							record.ID = 10
							return record, nil
						},
					},
				}
				employees := EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						FindFunc: func(_ context.Context, id uint64) (*Employee, error) {
							return &Employee{Base: model.Base{ID: id}, Active: true, OrganizationID: helper.Ptr(uint64(1))}, nil
						},
					},
				}
				shiftAssignment := &ShiftAssignment{Base: model.Base{ID: 1}, ShiftID: 5}
				assignments := ShiftAssignmentDAOMock{
					FindByEmployeeAndDateFunc: func(_ context.Context, _ uint64, _ string) (*ShiftAssignment, error) {
						return shiftAssignment, nil
					},
				}
				return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, TimesheetDAOMock{}, ShiftDAOMock{}, assignments, inventory.TransactionerMock{})
			},
			input: &Attendance{
				EmployeeID: 7,
				CheckIn:    helper.Ptr(time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)),
			},
			check: func(t *testing.T, a *Attendance) {
				if a.ShiftID == nil || *a.ShiftID != 5 {
					t.Fatalf("expected shift_id 5, got %v", a.ShiftID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup(t)
			a, err := svc.CheckIn(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantVal)
			if tt.check != nil {
				tt.check(t, a)
			}
		})
	}
}

func TestHRService_CreateEmployee_WritesInsideTransaction(t *testing.T) {
	var runs int
	txer := inventory.TransactionerMock{
		RunFunc: func(_ context.Context, fn func(tx *gorm.DB) error) error {
			runs++
			return fn(nil)
		},
	}
	outsideTx := errors.New("write outside transaction")
	contactDAO := contacts.ContactDAOMock{
		CreateWithDetailsFunc: func(context.Context, *contacts.Contact, []*contacts.ContactAddress, []*contacts.ContactBankAccount, *contacts.CustomerProfile, *contacts.SupplierProfile) (*contacts.Contact, error) {
			t.Error("contact created outside the transaction")
			return nil, outsideTx
		},
		CreateWithDetailsTxFunc: func(_ context.Context, _ *gorm.DB, contact *contacts.Contact, _ []*contacts.ContactAddress, _ []*contacts.ContactBankAccount, _ *contacts.CustomerProfile, _ *contacts.SupplierProfile) (*contacts.Contact, error) {
			contact.ID = 9
			return contact, nil
		},
	}
	employees := EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[Employee]{
			CreateFunc: func(context.Context, *Employee) (*Employee, error) {
				t.Error("employee created outside the transaction")
				return nil, outsideTx
			},
		},
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, employee *Employee) (*Employee, error) {
			employee.ID = 3
			return employee, nil
		},
	}
	contracts := EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[EmploymentContract]{
			CreateFunc: func(context.Context, *EmploymentContract) (*EmploymentContract, error) {
				t.Error("contract created outside the transaction")
				return nil, outsideTx
			},
		},
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, contract *EmploymentContract) (*EmploymentContract, error) {
			contract.ID = 4
			return contract, nil
		},
	}
	svc := NewHRService(employees, contracts, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contactDAO, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, txer)

	organizationID := uint64(1)
	employee, err := svc.CreateEmployee(context.Background(), CreateEmployeeRequest{
		Employee: &Employee{OrganizationID: &organizationID, EmployeeNumber: "EMP-001"},
		Contact:  &contacts.Contact{},
		Contract: &EmploymentContract{Wage: 5000000},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runs != 1 {
		t.Errorf("transaction runs = %d, want 1", runs)
	}
	if employee.ID != 3 || employee.ContactID != 9 {
		t.Errorf("employee = id %d contact %d, want id 3 contact 9", employee.ID, employee.ContactID)
	}
}
