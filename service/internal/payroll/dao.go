package payroll

import (
	"context"
	"errors"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type EmployeeDAO interface {
	dao.CRUD[Employee]
	CreateTx(ctx context.Context, tx *gorm.DB, entity *Employee) (*Employee, error)
}

type employeeDAO struct {
	dao.Base[Employee]
}

func NewEmployeeDAO(db *gorm.DB) EmployeeDAO {
	return employeeDAO{Base: dao.NewBase[Employee](db)}
}

func (d employeeDAO) CreateTx(ctx context.Context, tx *gorm.DB, entity *Employee) (*Employee, error) {
	if err := tx.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

type EmploymentContractDAO interface {
	dao.CRUD[EmploymentContract]
	CreateTx(ctx context.Context, tx *gorm.DB, entity *EmploymentContract) (*EmploymentContract, error)
	ListByEmployee(ctx context.Context, employeeID uint64) ([]*EmploymentContract, error)
	FindActiveByEmployee(ctx context.Context, employeeID uint64) (*EmploymentContract, error)
	ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[EmploymentContract], error)
	FindInOrg(ctx context.Context, id, organizationID uint64) (*EmploymentContract, error)
}

type employmentContractDAO struct {
	dao.Base[EmploymentContract]
	db *gorm.DB
}

func NewEmploymentContractDAO(db *gorm.DB) EmploymentContractDAO {
	return employmentContractDAO{Base: dao.NewBase[EmploymentContract](db), db: db}
}

func (d employmentContractDAO) CreateTx(ctx context.Context, tx *gorm.DB, entity *EmploymentContract) (*EmploymentContract, error) {
	if err := tx.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (d employmentContractDAO) ListByEmployee(ctx context.Context, employeeID uint64) ([]*EmploymentContract, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "employee_id", Operator: query.Equal, Value: employeeID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d employmentContractDAO) FindActiveByEmployee(ctx context.Context, employeeID uint64) (*EmploymentContract, error) {
	var contract EmploymentContract
	err := d.db.WithContext(ctx).
		Where("employee_id = ? AND state = ?", employeeID, ContractStateActive).
		First(&contract).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

func (d employmentContractDAO) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[EmploymentContract], error) {
	var count int64
	var entities []EmploymentContract

	join := "JOIN employees ON employees.id = employment_contracts.employee_id " +
		"AND employees.organization_id = ?"

	countTx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID)
	if q != nil {
		countTx = query.ApplyFilters(countTx, q.Filters)
	}
	if err := countTx.Count(&count).Error; err != nil {
		return nil, err
	}

	tx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID)
	if q != nil {
		tx = query.ApplyQuery(tx, q)
	}
	if err := tx.Find(&entities).Error; err != nil {
		return nil, err
	}

	items := make([]*EmploymentContract, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}

	return &query.Page[EmploymentContract]{Items: items, Count: count}, nil
}

func (d employmentContractDAO) FindInOrg(ctx context.Context, id, organizationID uint64) (*EmploymentContract, error) {
	var contract EmploymentContract
	err := d.db.WithContext(ctx).Model(&EmploymentContract{}).
		Joins("JOIN employees ON employees.id = employment_contracts.employee_id "+
			"AND employees.organization_id = ?", organizationID).
		Where("employment_contracts.id = ?", id).
		Take(&contract).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

type LeaveRequestDAO interface {
	dao.CRUD[LeaveRequest]
	ListByEmployee(ctx context.Context, employeeID uint64) ([]*LeaveRequest, error)
	ListApprovedByEmployeeAndType(ctx context.Context, employeeID, leaveTypeID uint64) ([]*LeaveRequest, error)
	ListApprovedByEmployeeBetweenDates(ctx context.Context, employeeID uint64, from, to string) ([]*LeaveRequest, error)
}

type leaveRequestDAO struct {
	dao.Base[LeaveRequest]
	db *gorm.DB
}

func NewLeaveRequestDAO(db *gorm.DB) LeaveRequestDAO {
	return leaveRequestDAO{Base: dao.NewBase[LeaveRequest](db), db: db}
}

func (d leaveRequestDAO) ListByEmployee(ctx context.Context, employeeID uint64) ([]*LeaveRequest, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "employee_id", Operator: query.Equal, Value: employeeID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d leaveRequestDAO) ListApprovedByEmployeeAndType(ctx context.Context, employeeID, leaveTypeID uint64) ([]*LeaveRequest, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "employee_id", Operator: query.Equal, Value: employeeID},
		{Field: "leave_type_id", Operator: query.Equal, Value: leaveTypeID},
		{Field: "state", Operator: query.Equal, Value: LeaveStateApproved},
	}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d leaveRequestDAO) ListApprovedByEmployeeBetweenDates(ctx context.Context, employeeID uint64, from, to string) ([]*LeaveRequest, error) {
	var results []LeaveRequest
	err := d.db.WithContext(ctx).
		Where("employee_id = ? AND state = ? AND date_from <= ? AND date_to >= ?",
			employeeID, LeaveStateApproved, to, from).
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	items := make([]*LeaveRequest, len(results))
	for i := range results {
		items[i] = &results[i]
	}
	return items, nil
}

type AttendanceDAO interface {
	dao.CRUD[Attendance]
	ListByEmployee(ctx context.Context, employeeID uint64) ([]*Attendance, error)
	FindOpenByEmployeeAndDate(ctx context.Context, employeeID uint64, date string) (*Attendance, error)
	ListMissingByDate(ctx context.Context, organizationID uint64, date string) ([]*Employee, error)
	ListByEmployeeAndDateRange(ctx context.Context, employeeID uint64, from, to string) ([]*Attendance, error)
	ListEmployeeIDsByDate(ctx context.Context, organizationID uint64, date string) ([]uint64, error)
}

type attendanceDAO struct {
	dao.Base[Attendance]
	db *gorm.DB
}

func NewAttendanceDAO(db *gorm.DB) AttendanceDAO {
	return attendanceDAO{Base: dao.NewBase[Attendance](db), db: db}
}

func (d attendanceDAO) ListByEmployee(ctx context.Context, employeeID uint64) ([]*Attendance, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "employee_id", Operator: query.Equal, Value: employeeID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d attendanceDAO) FindOpenByEmployeeAndDate(ctx context.Context, employeeID uint64, date string) (*Attendance, error) {
	var attendance Attendance
	err := d.db.WithContext(ctx).
		Where("employee_id = ? AND DATE(check_in) = ? AND check_out IS NULL AND deleted_at IS NULL", employeeID, date).
		First(&attendance).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &attendance, nil
}

func (d attendanceDAO) ListMissingByDate(ctx context.Context, organizationID uint64, date string) ([]*Employee, error) {
	var employees []Employee
	err := d.db.WithContext(ctx).
		Table("employees").
		Where("employees.organization_id = ? AND employees.active = true AND employees.deleted_at IS NULL", organizationID).
		Where("employees.id NOT IN (?)",
			d.db.WithContext(ctx).Table("attendances").
				Select("employee_id").
				Where("DATE(check_in) = ? AND deleted_at IS NULL", date),
		).
		Find(&employees).Error
	if err != nil {
		return nil, err
	}
	result := make([]*Employee, len(employees))
	for i := range employees {
		result[i] = &employees[i]
	}
	return result, nil
}

func (d attendanceDAO) ListByEmployeeAndDateRange(ctx context.Context, employeeID uint64, from, to string) ([]*Attendance, error) {
	var results []Attendance
	err := d.db.WithContext(ctx).
		Where("employee_id = ? AND DATE(check_in) >= ? AND DATE(check_in) <= ? AND deleted_at IS NULL",
			employeeID, from, to).
		Order("check_in ASC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	items := make([]*Attendance, len(results))
	for i := range results {
		items[i] = &results[i]
	}
	return items, nil
}

func (d attendanceDAO) ListEmployeeIDsByDate(ctx context.Context, organizationID uint64, date string) ([]uint64, error) {
	var ids []uint64
	err := d.db.WithContext(ctx).
		Table("attendances").
		Select("DISTINCT employee_id").
		Where("organization_id = ? AND DATE(check_in) = ? AND deleted_at IS NULL", organizationID, date).
		Pluck("employee_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

type TimesheetDAO interface {
	dao.CRUD[Timesheet]
	ListByEmployee(ctx context.Context, employeeID uint64) ([]*Timesheet, error)
	ListByProject(ctx context.Context, projectID uint64) ([]*Timesheet, error)
}

type timesheetDAO struct {
	dao.Base[Timesheet]
}

func NewTimesheetDAO(db *gorm.DB) TimesheetDAO {
	return timesheetDAO{Base: dao.NewBase[Timesheet](db)}
}

func (d timesheetDAO) ListByEmployee(ctx context.Context, employeeID uint64) ([]*Timesheet, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "employee_id", Operator: query.Equal, Value: employeeID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d timesheetDAO) ListByProject(ctx context.Context, projectID uint64) ([]*Timesheet, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "project_id", Operator: query.Equal, Value: projectID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type PayrollRunDAO interface {
	dao.CRUD[PayrollRun]
	CreateWithPayslipsTx(ctx context.Context, tx *gorm.DB, run *PayrollRun, payslips []*Payslip, lines [][]*PayslipLine) (*PayrollRun, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, run *PayrollRun) (*PayrollRun, error)
}

type payrollRunDAO struct {
	dao.Base[PayrollRun]
	db *gorm.DB
}

func NewPayrollRunDAO(db *gorm.DB) PayrollRunDAO {
	return payrollRunDAO{Base: dao.NewBase[PayrollRun](db), db: db}
}

func (d payrollRunDAO) CreateWithPayslipsTx(ctx context.Context, tx *gorm.DB, run *PayrollRun, payslips []*Payslip, lines [][]*PayslipLine) (*PayrollRun, error) {
	if err := tx.WithContext(ctx).Create(run).Error; err != nil {
		return nil, err
	}
	for _, payslip := range payslips {
		payslip.RunID = run.ID
	}
	if err := tx.WithContext(ctx).CreateInBatches(payslips, 500).Error; err != nil {
		return nil, err
	}
	allLines := make([]*PayslipLine, 0, len(lines))
	for i, payslip := range payslips {
		for _, line := range lines[i] {
			line.PayslipID = payslip.ID
			allLines = append(allLines, line)
		}
	}
	if err := tx.WithContext(ctx).CreateInBatches(allLines, 500).Error; err != nil {
		return nil, err
	}
	return run, nil
}

func (d payrollRunDAO) UpdateTx(ctx context.Context, tx *gorm.DB, run *PayrollRun) (*PayrollRun, error) {
	if err := tx.Save(run).Error; err != nil {
		return nil, err
	}
	return run, nil
}

type PayslipDAO interface {
	dao.CRUD[Payslip]
	ListByRun(ctx context.Context, runID uint64) ([]*Payslip, error)
	ListByEmployee(ctx context.Context, employeeID uint64) ([]*Payslip, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, payslip *Payslip) (*Payslip, error)
}

type payslipDAO struct {
	dao.Base[Payslip]
}

func NewPayslipDAO(db *gorm.DB) PayslipDAO {
	return payslipDAO{Base: dao.NewBase[Payslip](db)}
}

func (d payslipDAO) ListByRun(ctx context.Context, runID uint64) ([]*Payslip, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "run_id", Operator: query.Equal, Value: runID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d payslipDAO) ListByEmployee(ctx context.Context, employeeID uint64) ([]*Payslip, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "employee_id", Operator: query.Equal, Value: employeeID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d payslipDAO) UpdateTx(ctx context.Context, tx *gorm.DB, payslip *Payslip) (*Payslip, error) {
	if err := tx.Save(payslip).Error; err != nil {
		return nil, err
	}
	return payslip, nil
}

type PayslipLineDAO interface {
	dao.CRUD[PayslipLine]
	ListByPayslip(ctx context.Context, payslipID uint64) ([]*PayslipLine, error)
}

type payslipLineDAO struct {
	dao.Base[PayslipLine]
}

func NewPayslipLineDAO(db *gorm.DB) PayslipLineDAO {
	return payslipLineDAO{Base: dao.NewBase[PayslipLine](db)}
}

func (d payslipLineDAO) ListByPayslip(ctx context.Context, payslipID uint64) ([]*PayslipLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "payslip_id", Operator: query.Equal, Value: payslipID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type ShiftDAO interface {
	dao.CRUD[Shift]
}

type shiftDAO struct {
	dao.Base[Shift]
}

func NewShiftDAO(db *gorm.DB) ShiftDAO {
	return shiftDAO{Base: dao.NewBase[Shift](db)}
}

type ShiftAssignmentDAO interface {
	dao.CRUD[ShiftAssignment]
	FindByEmployeeAndDate(ctx context.Context, employeeID uint64, date string) (*ShiftAssignment, error)
}

type shiftAssignmentDAO struct {
	dao.Base[ShiftAssignment]
	db *gorm.DB
}

func NewShiftAssignmentDAO(db *gorm.DB) ShiftAssignmentDAO {
	return shiftAssignmentDAO{Base: dao.NewBase[ShiftAssignment](db), db: db}
}

func (d shiftAssignmentDAO) FindByEmployeeAndDate(ctx context.Context, employeeID uint64, date string) (*ShiftAssignment, error) {
	var assignment ShiftAssignment
	err := d.db.WithContext(ctx).
		Where("employee_id = ? AND date = ? AND deleted_at IS NULL", employeeID, date).
		First(&assignment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

type AttendanceDAOWithOrg interface {
	AttendanceDAO
	ListOpenByOrganization(ctx context.Context, organizationID uint64) ([]*Attendance, error)
}

type attendanceDAOWithOrg struct {
	attendanceDAO
}

func NewAttendanceDAOWithOrg(db *gorm.DB) AttendanceDAOWithOrg {
	return attendanceDAOWithOrg{attendanceDAO: attendanceDAO{Base: dao.NewBase[Attendance](db), db: db}}
}

func (d attendanceDAOWithOrg) ListOpenByOrganization(ctx context.Context, organizationID uint64) ([]*Attendance, error) {
	var results []Attendance
	err := d.db.WithContext(ctx).
		Table("attendances").
		Joins("JOIN employees ON employees.id = attendances.employee_id AND employees.organization_id = ?", organizationID).
		Where("attendances.check_out IS NULL AND attendances.deleted_at IS NULL AND employees.deleted_at IS NULL").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	items := make([]*Attendance, len(results))
	for i := range results {
		items[i] = &results[i]
	}
	return items, nil
}
