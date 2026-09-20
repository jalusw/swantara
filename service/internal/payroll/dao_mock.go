package payroll

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type EmployeeDAOMock struct {
	dao.CRUDMock[Employee]
}

type EmploymentContractDAOMock struct {
	dao.CRUDMock[EmploymentContract]
	ListByEmployeeFunc       func(ctx context.Context, employeeID uint64) ([]*EmploymentContract, error)
	FindActiveByEmployeeFunc func(ctx context.Context, employeeID uint64) (*EmploymentContract, error)
}

func (m EmploymentContractDAOMock) ListByEmployee(ctx context.Context, employeeID uint64) ([]*EmploymentContract, error) {
	if m.ListByEmployeeFunc != nil {
		return m.ListByEmployeeFunc(ctx, employeeID)
	}
	return []*EmploymentContract{}, nil
}

func (m EmploymentContractDAOMock) FindActiveByEmployee(ctx context.Context, employeeID uint64) (*EmploymentContract, error) {
	if m.FindActiveByEmployeeFunc != nil {
		return m.FindActiveByEmployeeFunc(ctx, employeeID)
	}
	return nil, nil
}

func (m EmploymentContractDAOMock) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[EmploymentContract], error) {
	return m.List(ctx, q)
}

func (m EmploymentContractDAOMock) FindInOrg(ctx context.Context, id, organizationID uint64) (*EmploymentContract, error) {
	return m.Find(ctx, id)
}

type LeaveRequestDAOMock struct {
	dao.CRUDMock[LeaveRequest]
	ListByEmployeeFunc                     func(ctx context.Context, employeeID uint64) ([]*LeaveRequest, error)
	ListApprovedByEmployeeAndTypeFunc      func(ctx context.Context, employeeID, leaveTypeID uint64) ([]*LeaveRequest, error)
	ListApprovedByEmployeeBetweenDatesFunc func(ctx context.Context, employeeID uint64, from, to string) ([]*LeaveRequest, error)
}

func (m LeaveRequestDAOMock) ListByEmployee(ctx context.Context, employeeID uint64) ([]*LeaveRequest, error) {
	if m.ListByEmployeeFunc != nil {
		return m.ListByEmployeeFunc(ctx, employeeID)
	}
	return []*LeaveRequest{}, nil
}

func (m LeaveRequestDAOMock) ListApprovedByEmployeeAndType(ctx context.Context, employeeID, leaveTypeID uint64) ([]*LeaveRequest, error) {
	if m.ListApprovedByEmployeeAndTypeFunc != nil {
		return m.ListApprovedByEmployeeAndTypeFunc(ctx, employeeID, leaveTypeID)
	}
	return []*LeaveRequest{}, nil
}

func (m LeaveRequestDAOMock) ListApprovedByEmployeeBetweenDates(ctx context.Context, employeeID uint64, from, to string) ([]*LeaveRequest, error) {
	if m.ListApprovedByEmployeeBetweenDatesFunc != nil {
		return m.ListApprovedByEmployeeBetweenDatesFunc(ctx, employeeID, from, to)
	}
	return []*LeaveRequest{}, nil
}

type AttendanceDAOMock struct {
	dao.CRUDMock[Attendance]
	ListByEmployeeFunc             func(ctx context.Context, employeeID uint64) ([]*Attendance, error)
	FindOpenByEmployeeAndDateFunc  func(ctx context.Context, employeeID uint64, date string) (*Attendance, error)
	ListMissingByDateFunc          func(ctx context.Context, organizationID uint64, date string) ([]*Employee, error)
	ListByEmployeeAndDateRangeFunc func(ctx context.Context, employeeID uint64, from, to string) ([]*Attendance, error)
	ListEmployeeIDsByDateFunc      func(ctx context.Context, organizationID uint64, date string) ([]uint64, error)
}

func (m AttendanceDAOMock) ListByEmployee(ctx context.Context, employeeID uint64) ([]*Attendance, error) {
	if m.ListByEmployeeFunc != nil {
		return m.ListByEmployeeFunc(ctx, employeeID)
	}
	return []*Attendance{}, nil
}

func (m AttendanceDAOMock) FindOpenByEmployeeAndDate(ctx context.Context, employeeID uint64, date string) (*Attendance, error) {
	if m.FindOpenByEmployeeAndDateFunc != nil {
		return m.FindOpenByEmployeeAndDateFunc(ctx, employeeID, date)
	}
	return nil, nil
}

func (m AttendanceDAOMock) ListMissingByDate(ctx context.Context, organizationID uint64, date string) ([]*Employee, error) {
	if m.ListMissingByDateFunc != nil {
		return m.ListMissingByDateFunc(ctx, organizationID, date)
	}
	return []*Employee{}, nil
}

func (m AttendanceDAOMock) ListByEmployeeAndDateRange(ctx context.Context, employeeID uint64, from, to string) ([]*Attendance, error) {
	if m.ListByEmployeeAndDateRangeFunc != nil {
		return m.ListByEmployeeAndDateRangeFunc(ctx, employeeID, from, to)
	}
	return []*Attendance{}, nil
}

func (m AttendanceDAOMock) ListEmployeeIDsByDate(ctx context.Context, organizationID uint64, date string) ([]uint64, error) {
	if m.ListEmployeeIDsByDateFunc != nil {
		return m.ListEmployeeIDsByDateFunc(ctx, organizationID, date)
	}
	return []uint64{}, nil
}

type TimesheetDAOMock struct {
	dao.CRUDMock[Timesheet]
	ListByEmployeeFunc func(ctx context.Context, employeeID uint64) ([]*Timesheet, error)
	ListByProjectFunc  func(ctx context.Context, projectID uint64) ([]*Timesheet, error)
}

func (m TimesheetDAOMock) ListByEmployee(ctx context.Context, employeeID uint64) ([]*Timesheet, error) {
	if m.ListByEmployeeFunc != nil {
		return m.ListByEmployeeFunc(ctx, employeeID)
	}
	return []*Timesheet{}, nil
}

func (m TimesheetDAOMock) ListByProject(ctx context.Context, projectID uint64) ([]*Timesheet, error) {
	if m.ListByProjectFunc != nil {
		return m.ListByProjectFunc(ctx, projectID)
	}
	return []*Timesheet{}, nil
}

type PayrollRunDAOMock struct {
	dao.CRUDMock[PayrollRun]
	CreateWithPayslipsTxFunc func(ctx context.Context, tx *gorm.DB, run *PayrollRun, payslips []*Payslip, lines [][]*PayslipLine) (*PayrollRun, error)
	UpdateTxFunc             func(ctx context.Context, tx *gorm.DB, run *PayrollRun) (*PayrollRun, error)
}

func (m PayrollRunDAOMock) CreateWithPayslipsTx(ctx context.Context, tx *gorm.DB, run *PayrollRun, payslips []*Payslip, lines [][]*PayslipLine) (*PayrollRun, error) {
	if m.CreateWithPayslipsTxFunc != nil {
		return m.CreateWithPayslipsTxFunc(ctx, tx, run, payslips, lines)
	}
	return run, nil
}

func (m PayrollRunDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, run *PayrollRun) (*PayrollRun, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, run)
	}
	return run, nil
}

type PayslipDAOMock struct {
	dao.CRUDMock[Payslip]
	ListByRunFunc      func(ctx context.Context, runID uint64) ([]*Payslip, error)
	ListByEmployeeFunc func(ctx context.Context, employeeID uint64) ([]*Payslip, error)
	UpdateTxFunc       func(ctx context.Context, tx *gorm.DB, payslip *Payslip) (*Payslip, error)
}

func (m PayslipDAOMock) ListByRun(ctx context.Context, runID uint64) ([]*Payslip, error) {
	if m.ListByRunFunc != nil {
		return m.ListByRunFunc(ctx, runID)
	}
	return []*Payslip{}, nil
}

func (m PayslipDAOMock) ListByEmployee(ctx context.Context, employeeID uint64) ([]*Payslip, error) {
	if m.ListByEmployeeFunc != nil {
		return m.ListByEmployeeFunc(ctx, employeeID)
	}
	return []*Payslip{}, nil
}

func (m PayslipDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, payslip *Payslip) (*Payslip, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, payslip)
	}
	return payslip, nil
}

type PayslipLineDAOMock struct {
	dao.CRUDMock[PayslipLine]
	ListByPayslipFunc func(ctx context.Context, payslipID uint64) ([]*PayslipLine, error)
}

func (m PayslipLineDAOMock) ListByPayslip(ctx context.Context, payslipID uint64) ([]*PayslipLine, error) {
	if m.ListByPayslipFunc != nil {
		return m.ListByPayslipFunc(ctx, payslipID)
	}
	return []*PayslipLine{}, nil
}

type ShiftDAOMock struct {
	dao.CRUDMock[Shift]
}

type ShiftAssignmentDAOMock struct {
	dao.CRUDMock[ShiftAssignment]
	FindByEmployeeAndDateFunc func(ctx context.Context, employeeID uint64, date string) (*ShiftAssignment, error)
}

func (m ShiftAssignmentDAOMock) FindByEmployeeAndDate(ctx context.Context, employeeID uint64, date string) (*ShiftAssignment, error) {
	if m.FindByEmployeeAndDateFunc != nil {
		return m.FindByEmployeeAndDateFunc(ctx, employeeID, date)
	}
	return nil, nil
}

type AttendanceDAOWithOrgMock struct {
	AttendanceDAOMock
	ListOpenByOrganizationFunc func(ctx context.Context, organizationID uint64) ([]*Attendance, error)
}

func (m AttendanceDAOWithOrgMock) ListOpenByOrganization(ctx context.Context, organizationID uint64) ([]*Attendance, error) {
	if m.ListOpenByOrganizationFunc != nil {
		return m.ListOpenByOrganizationFunc(ctx, organizationID)
	}
	return []*Attendance{}, nil
}
