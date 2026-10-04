package payroll

import (
	"context"
	"math"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type CreateEmployeeRequest struct {
	Employee *Employee
	Contract *EmploymentContract
	Contact  *contacts.Contact
}

type HRService struct {
	employees        EmployeeDAO
	contracts        EmploymentContractDAO
	leaves           LeaveRequestDAO
	leaveTypes       dao.CRUD[reference.LeaveType]
	departments      dao.CRUD[reference.Department]
	positions        dao.CRUD[reference.JobPosition]
	organizations    dao.CRUD[reference.Organization]
	contacts         contacts.ContactDAO
	dimensions       dao.CRUD[reference.Dimension]
	users            iam.UserDAO
	attendances      AttendanceDAO
	timesheets       TimesheetDAO
	shifts           ShiftDAO
	shiftAssignments ShiftAssignmentDAO
	tx               db.Transactioner
}

func NewHRService(
	employees EmployeeDAO,
	contracts EmploymentContractDAO,
	leaves LeaveRequestDAO,
	leaveTypes dao.CRUD[reference.LeaveType],
	departments dao.CRUD[reference.Department],
	positions dao.CRUD[reference.JobPosition],
	organizations dao.CRUD[reference.Organization],
	contacts contacts.ContactDAO,
	dimensions dao.CRUD[reference.Dimension],
	users iam.UserDAO,
	attendances AttendanceDAO,
	timesheets TimesheetDAO,
	shifts ShiftDAO,
	shiftAssignments ShiftAssignmentDAO,
	tx db.Transactioner,
) HRService {
	return HRService{
		employees:        employees,
		contracts:        contracts,
		leaves:           leaves,
		leaveTypes:       leaveTypes,
		departments:      departments,
		positions:        positions,
		organizations:    organizations,
		contacts:         contacts,
		dimensions:       dimensions,
		users:            users,
		attendances:      attendances,
		timesheets:       timesheets,
		shifts:           shifts,
		shiftAssignments: shiftAssignments,
		tx:               tx,
	}
}

func (s HRService) ListEmployees(ctx context.Context, q *query.Query) (*query.Page[Employee], error) {
	return s.employees.List(ctx, q)
}

func (s HRService) FindEmployee(ctx context.Context, id uint64) (*Employee, error) {
	return s.employees.Find(ctx, id)
}

func (s HRService) DeleteEmployee(ctx context.Context, id uint64) error {
	return s.employees.Delete(ctx, id)
}

func (s HRService) ListContractsInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[EmploymentContract], error) {
	return s.contracts.ListInOrg(ctx, q, organizationID)
}

func (s HRService) FindContractInOrg(ctx context.Context, id, organizationID uint64) (*EmploymentContract, error) {
	return s.contracts.FindInOrg(ctx, id, organizationID)
}

func (s HRService) ListLeaves(ctx context.Context, q *query.Query) (*query.Page[LeaveRequest], error) {
	return s.leaves.List(ctx, q)
}

func (s HRService) FindLeave(ctx context.Context, id uint64) (*LeaveRequest, error) {
	return s.leaves.Find(ctx, id)
}

func (s HRService) ListAttendances(ctx context.Context, q *query.Query) (*query.Page[Attendance], error) {
	return s.attendances.List(ctx, q)
}

func (s HRService) FindAttendance(ctx context.Context, id uint64) (*Attendance, error) {
	return s.attendances.Find(ctx, id)
}

func (s HRService) ListTimesheets(ctx context.Context, q *query.Query) (*query.Page[Timesheet], error) {
	return s.timesheets.List(ctx, q)
}

func (s HRService) FindTimesheet(ctx context.Context, id uint64) (*Timesheet, error) {
	return s.timesheets.Find(ctx, id)
}

func (s HRService) DeleteTimesheet(ctx context.Context, id uint64) error {
	return s.timesheets.Delete(ctx, id)
}

func (s HRService) CreateEmployee(ctx context.Context, request CreateEmployeeRequest) (*Employee, error) {
	employee := request.Employee
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}
	if err := s.validateEmployee(ctx, employee); err != nil {
		return nil, err
	}
	if request.Contract != nil && !(request.Contract.Wage > 0) {
		return nil, ErrContractWage
	}
	if request.Contact == nil {
		return nil, ErrEmployeeContact
	}
	var created *Employee
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		contact, err := s.contacts.CreateWithDetailsTx(ctx, tx, request.Contact, nil, nil, nil, nil)
		if err != nil {
			return err
		}
		employee.ContactID = contact.ID
		created, err = s.employees.CreateTx(ctx, tx, employee)
		if err != nil {
			return err
		}
		if request.Contract == nil {
			return nil
		}
		contract := request.Contract
		contract.EmployeeID = created.ID
		contract.State = ContractStateActive
		if contract.DateStart.IsZero() {
			contract.DateStart = time.Now().UTC()
		}
		if contract.CurrencyCode == "" {
			contract.CurrencyCode = "IDR"
		}
		if contract.WageType == "" {
			contract.WageType = WageTypeMonthly
		}
		_, err = s.contracts.CreateTx(ctx, tx, contract)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s HRService) UpdateEmployee(ctx context.Context, employee *Employee) (*Employee, error) {
	existing, err := s.employees.Find(ctx, employee.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrEmployeeNotFound
	}
	if err := s.validateEmployee(ctx, employee); err != nil {
		return nil, err
	}
	existing.UserID = employee.UserID
	existing.EmployeeNumber = employee.EmployeeNumber
	existing.DepartmentID = employee.DepartmentID
	existing.JobPositionID = employee.JobPositionID
	existing.ManagerID = employee.ManagerID
	existing.HireDate = employee.HireDate
	existing.TerminationDate = employee.TerminationDate
	existing.EmploymentType = employee.EmploymentType
	existing.WorkLocation = employee.WorkLocation
	existing.Active = employee.Active
	existing.RequiresApproval = employee.RequiresApproval
	return s.employees.Update(ctx, existing)
}

func (s HRService) CreateContract(ctx context.Context, organizationID uint64, contract *EmploymentContract) (*EmploymentContract, error) {
	if err := s.validateContract(ctx, contract); err != nil {
		return nil, err
	}
	employee, err := s.employees.Find(ctx, contract.EmployeeID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrContractEmployee
	}
	if employee.OrganizationID == nil || *employee.OrganizationID != organizationID {
		return nil, ErrContractEmployee
	}
	contract.State = ContractStateActive
	if contract.DateStart.IsZero() {
		contract.DateStart = time.Now().UTC()
	}
	if contract.CurrencyCode == "" {
		contract.CurrencyCode = "IDR"
	}
	if contract.WageType == "" {
		contract.WageType = WageTypeMonthly
	}
	return s.contracts.Create(ctx, contract)
}

func (s HRService) UpdateContract(ctx context.Context, contract *EmploymentContract) (*EmploymentContract, error) {
	existing, err := s.contracts.Find(ctx, contract.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrContractNotFound
	}
	if err := s.validateContract(ctx, contract); err != nil {
		return nil, err
	}
	existing.DateStart = contract.DateStart
	existing.DateEnd = contract.DateEnd
	existing.Wage = contract.Wage
	existing.WageType = contract.WageType
	existing.CurrencyCode = contract.CurrencyCode
	existing.State = contract.State
	return s.contracts.Update(ctx, existing)
}

func (s HRService) TerminateContract(ctx context.Context, contractID uint64) (*EmploymentContract, error) {
	contract, err := s.contracts.Find(ctx, contractID)
	if err != nil {
		return nil, err
	}
	if contract == nil {
		return nil, ErrContractNotFound
	}
	if contract.State != ContractStateActive {
		return nil, ErrContractState
	}
	contract.State = ContractStateClosed
	return s.contracts.Update(ctx, contract)
}

func (s HRService) SubmitLeave(ctx context.Context, leaveID uint64) (*LeaveRequest, error) {
	leave, err := s.leaves.Find(ctx, leaveID)
	if err != nil {
		return nil, err
	}
	if leave == nil {
		return nil, ErrLeaveRequestNotFound
	}
	if leave.State != LeaveStateDraft {
		return nil, ErrLeaveState
	}
	leave.State = LeaveStateSubmitted
	return s.leaves.Update(ctx, leave)
}

func (s HRService) ApproveLeave(ctx context.Context, leaveID uint64) (*LeaveRequest, error) {
	leave, err := s.leaves.Find(ctx, leaveID)
	if err != nil {
		return nil, err
	}
	if leave == nil {
		return nil, ErrLeaveRequestNotFound
	}
	if leave.State != LeaveStateSubmitted {
		return nil, ErrLeaveState
	}
	if err := s.assertLeaveBalance(ctx, leave.EmployeeID, leave.LeaveTypeID, leave.Days); err != nil {
		return nil, err
	}
	leave.State = LeaveStateApproved
	return s.leaves.Update(ctx, leave)
}

func (s HRService) RefuseLeave(ctx context.Context, leaveID uint64) (*LeaveRequest, error) {
	leave, err := s.leaves.Find(ctx, leaveID)
	if err != nil {
		return nil, err
	}
	if leave == nil {
		return nil, ErrLeaveRequestNotFound
	}
	if leave.State != LeaveStateSubmitted {
		return nil, ErrLeaveState
	}
	leave.State = LeaveStateRefused
	return s.leaves.Update(ctx, leave)
}

func (s HRService) CreateLeaveRequest(ctx context.Context, request *LeaveRequest) (*LeaveRequest, error) {
	if request == nil {
		return nil, ErrLeaveRequestNotFound
	}
	if request.Days <= 0 {
		return nil, ErrLeaveDays
	}
	if !request.DateTo.Equal(request.DateFrom) && request.DateTo.Before(request.DateFrom) {
		return nil, ErrLeaveDateRange
	}
	employee, err := s.employees.Find(ctx, request.EmployeeID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrLeaveEmployee
	}
	leaveType, err := s.leaveTypes.Find(ctx, request.LeaveTypeID)
	if err != nil {
		return nil, err
	}
	if leaveType == nil {
		return nil, ErrLeaveTypeNotFound
	}
	request.State = LeaveStateDraft
	return s.leaves.Create(ctx, request)
}

func (s HRService) LeaveBalance(ctx context.Context, employeeID, leaveTypeID uint64) (float64, error) {
	leaveType, err := s.leaveTypes.Find(ctx, leaveTypeID)
	if err != nil {
		return 0, err
	}
	if leaveType == nil {
		return 0, ErrLeaveTypeNotFound
	}
	approved, err := s.leaves.ListApprovedByEmployeeAndType(ctx, employeeID, leaveTypeID)
	if err != nil {
		return 0, err
	}
	allocation := 0.0
	if leaveType.AllocationDays != nil {
		allocation = *leaveType.AllocationDays
	}
	used := 0.0
	for _, request := range approved {
		used += request.Days
	}
	balance := allocation - used
	if balance < 0 {
		balance = 0
	}
	return balance, nil
}

func (s HRService) CheckIn(ctx context.Context, attendance *Attendance) (*Attendance, error) {
	if attendance == nil || attendance.CheckIn == nil {
		return nil, ErrAttendanceCheckIn
	}
	employee, err := s.employees.Find(ctx, attendance.EmployeeID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrAttendanceEmployee
	}
	if !employee.Active {
		return nil, ErrAttendanceInactive
	}
	date := attendance.CheckIn.Format("2006-01-02")
	existing, err := s.attendances.FindOpenByEmployeeAndDate(ctx, attendance.EmployeeID, date)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAttendanceOpenExists
	}
	onLeave, err := s.leaves.ListApprovedByEmployeeBetweenDates(ctx, attendance.EmployeeID, date, date)
	if err != nil {
		return nil, err
	}
	if len(onLeave) > 0 {
		return nil, ErrAttendanceOnLeave
	}
	if employee.OrganizationID != nil {
		attendance.OrganizationID = *employee.OrganizationID
	}
	attendance.WorkedHours = 0
	attendance.Status = AttendanceStatusDraft
	if attendance.AttendanceType == "" {
		attendance.AttendanceType = AttendanceTypeNormal
	}
	if attendance.Source == "" {
		attendance.Source = AttendanceSourceManual
	}
	assignment, err := s.shiftAssignments.FindByEmployeeAndDate(ctx, attendance.EmployeeID, date)
	if err != nil {
		return nil, err
	}
	if assignment != nil {
		attendance.ShiftID = &assignment.ShiftID
	}
	return s.attendances.Create(ctx, attendance)
}

func (s HRService) CheckOut(ctx context.Context, attendanceID uint64, checkOut time.Time, breakMinutes int) (*Attendance, error) {
	attendance, err := s.attendances.Find(ctx, attendanceID)
	if err != nil {
		return nil, err
	}
	if attendance == nil {
		return nil, ErrAttendanceNotFound
	}
	if attendance.CheckOut != nil {
		return nil, ErrAttendanceAlreadyClosed
	}
	if attendance.CheckIn == nil || !checkOut.After(*attendance.CheckIn) {
		return nil, ErrAttendanceCheckOut
	}
	attendance.CheckOut = &checkOut
	if breakMinutes < 0 {
		breakMinutes = 0
	}
	attendance.BreakMinutes = breakMinutes
	elapsed := checkOut.Sub(*attendance.CheckIn).Hours()
	breakHours := float64(breakMinutes) / 60.0
	worked := elapsed - breakHours
	if worked < 0 {
		worked = 0
	}
	attendance.WorkedHours = math.Round(worked*100) / 100
	employee, err := s.employees.Find(ctx, attendance.EmployeeID)
	if err != nil {
		return nil, err
	}
	if employee != nil && employee.OrganizationID != nil {
		rounding := s.getOrganizationRounding(ctx, *employee.OrganizationID)
		if rounding > 0 {
			attendance.WorkedHours = roundToNearest(attendance.WorkedHours, rounding)
		}
	}
	worked = attendance.WorkedHours
	if worked > OvertimeThresholdHours {
		attendance.OvertimeHours = math.Round((worked-OvertimeThresholdHours)*100) / 100
	} else {
		attendance.OvertimeHours = 0
	}
	employee, err = s.employees.Find(ctx, attendance.EmployeeID)
	if err != nil {
		return nil, err
	}
	if employee != nil && employee.RequiresApproval {
		attendance.Status = AttendanceStatusPendingApproval
	} else {
		attendance.Status = AttendanceStatusConfirmed
		confirmedAt := time.Now().UTC()
		attendance.ConfirmedAt = &confirmedAt
	}
	updated, err := s.attendances.Update(ctx, attendance)
	if err != nil {
		return nil, err
	}
	s.createTimesheetFromAttendance(ctx, attendance)
	return updated, nil
}

func (s HRService) UpdateAttendance(ctx context.Context, attendance *Attendance) (*Attendance, error) {
	existing, err := s.attendances.Find(ctx, attendance.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrAttendanceNotFound
	}
	if existing.Status == AttendanceStatusCancelled {
		return nil, ErrAttendanceState
	}
	existing.AttendanceType = attendance.AttendanceType
	existing.Source = attendance.Source
	existing.BreakMinutes = attendance.BreakMinutes
	existing.Notes = attendance.Notes
	existing.LateMinutes = attendance.LateMinutes
	existing.EarlyDepartureMinutes = attendance.EarlyDepartureMinutes
	return s.attendances.Update(ctx, existing)
}

func (s HRService) DeleteAttendance(ctx context.Context, attendanceID uint64) error {
	existing, err := s.attendances.Find(ctx, attendanceID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrAttendanceNotFound
	}
	return s.attendances.Delete(ctx, attendanceID)
}

func (s HRService) BulkCheckOut(ctx context.Context, attendanceIDs []uint64, checkOut time.Time, breakMinutes int) ([]*Attendance, error) {
	results := make([]*Attendance, 0, len(attendanceIDs))
	for _, id := range attendanceIDs {
		record, err := s.CheckOut(ctx, id, checkOut, breakMinutes)
		if err != nil {
			return nil, err
		}
		results = append(results, record)
	}
	return results, nil
}

func (s HRService) ListMissingAttendance(ctx context.Context, organizationID uint64, date string) ([]*Employee, error) {
	return s.attendances.ListMissingByDate(ctx, organizationID, date)
}

func (s HRService) CreateTimesheet(ctx context.Context, timesheet *Timesheet) (*Timesheet, error) {
	if timesheet == nil {
		return nil, ErrTimesheetNotFound
	}
	employee, err := s.employees.Find(ctx, timesheet.EmployeeID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrTimesheetEmployee
	}
	if timesheet.Hours <= 0 {
		return nil, ErrTimesheetHours
	}
	if timesheet.DimensionID != nil {
		account, err := s.dimensions.Find(ctx, *timesheet.DimensionID)
		if err != nil {
			return nil, err
		}
		if account == nil {
			return nil, ErrTimesheetDimension
		}
	}
	return s.timesheets.Create(ctx, timesheet)
}

func (s HRService) validateEmployee(ctx context.Context, employee *Employee) error {
	if employee.OrganizationID == nil {
		return ErrEmployeeOrganization
	}
	if employee.EmployeeNumber == "" {
		return ErrEmployeeNumber
	}
	existing, err := s.employees.Search(ctx, "employee_number", employee.EmployeeNumber)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != employee.ID {
		return ErrEmployeeNumberTaken
	}
	if employee.UserID != nil {
		user, err := s.users.Find(ctx, *employee.UserID)
		if err != nil {
			return err
		}
		if user == nil {
			return ErrEmployeeUserNotFound
		}
		linked, err := s.employees.Search(ctx, "user_id", *employee.UserID)
		if err != nil {
			return err
		}
		if linked != nil && linked.ID != employee.ID {
			return ErrEmployeeUserTaken
		}
	}
	if employee.DepartmentID != nil {
		department, err := s.departments.Find(ctx, *employee.DepartmentID)
		if err != nil {
			return err
		}
		if department == nil {
			return ErrEmployeeDepartment
		}
	}
	if employee.JobPositionID != nil {
		position, err := s.positions.Find(ctx, *employee.JobPositionID)
		if err != nil {
			return err
		}
		if position == nil {
			return ErrEmployeePosition
		}
	}
	if employee.ManagerID != nil {
		manager, err := s.employees.Find(ctx, *employee.ManagerID)
		if err != nil {
			return err
		}
		if manager == nil {
			return ErrEmployeeManager
		}
	}
	switch employee.EmploymentType {
	case "", EmploymentTypeFullTime, EmploymentTypePartTime, EmploymentTypeContract:
	default:
		return ErrEmployeeEmploymentType
	}
	return nil
}

func (s HRService) validateContract(ctx context.Context, contract *EmploymentContract) error {
	if contract == nil {
		return ErrContractNotFound
	}
	if contract.EmployeeID == 0 {
		return ErrContractEmployee
	}
	if contract.Wage <= 0 {
		return ErrContractWage
	}
	return nil
}

func (s HRService) assertLeaveBalance(ctx context.Context, employeeID, leaveTypeID uint64, days float64) error {
	if days <= 0 {
		return ErrLeaveDays
	}
	balance, err := s.LeaveBalance(ctx, employeeID, leaveTypeID)
	if err != nil {
		return err
	}
	if days > balance {
		return ErrLeaveBalance
	}
	return nil
}

func (s HRService) createTimesheetFromAttendance(ctx context.Context, attendance *Attendance) {
	if attendance.WorkedHours <= 0 || attendance.CheckIn == nil {
		return
	}
	date := attendance.CheckIn.Format("2006-01-02")
	existing, err := s.timesheets.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "employee_id", Operator: query.Equal, Value: attendance.EmployeeID},
	}})
	if err != nil {
		return
	}
	for _, t := range existing.Items {
		if t.Date.Format("2006-01-02") == date {
			return
		}
	}
	_, _ = s.timesheets.Create(ctx, &Timesheet{
		EmployeeID: attendance.EmployeeID,
		Date:       *attendance.CheckIn,
		Hours:      attendance.WorkedHours,
	})
}

func (s HRService) ApproveAttendance(ctx context.Context, attendanceID uint64) (*Attendance, error) {
	attendance, err := s.attendances.Find(ctx, attendanceID)
	if err != nil {
		return nil, err
	}
	if attendance == nil {
		return nil, ErrAttendanceNotFound
	}
	if attendance.Status != AttendanceStatusPendingApproval {
		return nil, ErrAttendanceNotPending
	}
	attendance.Status = AttendanceStatusConfirmed
	confirmedAt := time.Now().UTC()
	attendance.ConfirmedAt = &confirmedAt
	updated, err := s.attendances.Update(ctx, attendance)
	if err != nil {
		return nil, err
	}
	s.createTimesheetFromAttendance(ctx, updated)
	return updated, nil
}

func (s HRService) RejectAttendance(ctx context.Context, attendanceID uint64) (*Attendance, error) {
	attendance, err := s.attendances.Find(ctx, attendanceID)
	if err != nil {
		return nil, err
	}
	if attendance == nil {
		return nil, ErrAttendanceNotFound
	}
	if attendance.Status != AttendanceStatusPendingApproval {
		return nil, ErrAttendanceNotPending
	}
	attendance.Status = AttendanceStatusCancelled
	return s.attendances.Update(ctx, attendance)
}

func (s HRService) AutoCheckout(ctx context.Context, organizationID uint64) ([]*Attendance, error) {
	attendances, err := s.attendances.(AttendanceDAOWithOrg).ListOpenByOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	results := make([]*Attendance, 0, len(attendances))
	for _, att := range attendances {
		if att.CheckIn == nil {
			continue
		}
		employee, err := s.employees.Find(ctx, att.EmployeeID)
		if err != nil {
			return nil, err
		}
		if employee == nil || !employee.Active {
			continue
		}
		checkoutTime := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, now.Location())
		updated, err := s.CheckOut(ctx, att.ID, checkoutTime, 0)
		if err != nil {
			continue
		}
		results = append(results, updated)
	}
	return results, nil
}

func (s HRService) CreateShift(ctx context.Context, shift *Shift) (*Shift, error) {
	if shift == nil {
		return nil, ErrShiftNotFound
	}
	if shift.Name == "" {
		return nil, ErrShiftName
	}
	if shift.StartTime == "" || shift.EndTime == "" {
		return nil, ErrShiftTime
	}
	if shift.OrganizationID == 0 {
		return nil, ErrShiftOrganization
	}
	return s.shifts.Create(ctx, shift)
}

func (s HRService) UpdateShift(ctx context.Context, shift *Shift) (*Shift, error) {
	existing, err := s.shifts.Find(ctx, shift.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrShiftNotFound
	}
	if shift.Name == "" {
		return nil, ErrShiftName
	}
	if shift.StartTime == "" || shift.EndTime == "" {
		return nil, ErrShiftTime
	}
	existing.Name = shift.Name
	existing.StartTime = shift.StartTime
	existing.EndTime = shift.EndTime
	return s.shifts.Update(ctx, existing)
}

func (s HRService) DeleteShift(ctx context.Context, shiftID uint64) error {
	existing, err := s.shifts.Find(ctx, shiftID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrShiftNotFound
	}
	return s.shifts.Delete(ctx, shiftID)
}

func (s HRService) CreateShiftAssignment(ctx context.Context, assignment *ShiftAssignment) (*ShiftAssignment, error) {
	if assignment == nil {
		return nil, ErrShiftAssignmentNotFound
	}
	employee, err := s.employees.Find(ctx, assignment.EmployeeID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrShiftAssignmentEmployee
	}
	shift, err := s.shifts.Find(ctx, assignment.ShiftID)
	if err != nil {
		return nil, err
	}
	if shift == nil {
		return nil, ErrShiftAssignmentShift
	}
	date := assignment.Date.Format("2006-01-02")
	existing, err := s.shiftAssignments.FindByEmployeeAndDate(ctx, assignment.EmployeeID, date)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrShiftAssignmentExists
	}
	return s.shiftAssignments.Create(ctx, assignment)
}

func (s HRService) DeleteShiftAssignment(ctx context.Context, assignmentID uint64) error {
	existing, err := s.shiftAssignments.Find(ctx, assignmentID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrShiftAssignmentNotFound
	}
	return s.shiftAssignments.Delete(ctx, assignmentID)
}

func (s HRService) MarkAbsent(ctx context.Context, organizationID uint64, date string) ([]*Attendance, error) {
	missing, err := s.attendances.ListMissingByDate(ctx, organizationID, date)
	if err != nil {
		return nil, err
	}
	var created []*Attendance
	for _, emp := range missing {
		att := &Attendance{
			EmployeeID:     emp.ID,
			OrganizationID: organizationID,
			Status:         AttendanceStatusAbsent,
			AttendanceType: AttendanceTypeNormal,
			Source:         AttendanceSourceManual,
		}
		result, err := s.attendances.Create(ctx, att)
		if err != nil {
			return nil, err
		}
		created = append(created, result)
	}
	return created, nil
}

func (s HRService) getOrganizationRounding(ctx context.Context, organizationID uint64) int {
	org, err := s.organizations.Find(ctx, organizationID)
	if err != nil || org == nil {
		return 0
	}
	return org.RoundingMinutes
}

func roundToNearest(hours float64, roundingMinutes int) float64 {
	if roundingMinutes <= 0 {
		return hours
	}
	totalMinutes := hours * 60.0
	rounded := math.Ceil(totalMinutes/float64(roundingMinutes)) * float64(roundingMinutes)
	return math.Round(rounded/60.0*100) / 100
}

func (s HRService) ListDepartments(ctx context.Context, q *query.Query) (*query.Page[reference.Department], error) {
	return s.departments.List(ctx, q)
}

func (s HRService) FindDepartment(ctx context.Context, id uint64) (*reference.Department, error) {
	return s.departments.Find(ctx, id)
}

func (s HRService) CreateDepartment(ctx context.Context, dept *reference.Department) (*reference.Department, error) {
	return s.departments.Create(ctx, dept)
}

func (s HRService) UpdateDepartment(ctx context.Context, dept *reference.Department) (*reference.Department, error) {
	return s.departments.Update(ctx, dept)
}

func (s HRService) DeleteDepartment(ctx context.Context, id uint64) error {
	return s.departments.Delete(ctx, id)
}

func (s HRService) ListJobPositions(ctx context.Context, q *query.Query) (*query.Page[reference.JobPosition], error) {
	return s.positions.List(ctx, q)
}

func (s HRService) FindJobPosition(ctx context.Context, id uint64) (*reference.JobPosition, error) {
	return s.positions.Find(ctx, id)
}

func (s HRService) CreateJobPosition(ctx context.Context, pos *reference.JobPosition) (*reference.JobPosition, error) {
	return s.positions.Create(ctx, pos)
}

func (s HRService) UpdateJobPosition(ctx context.Context, pos *reference.JobPosition) (*reference.JobPosition, error) {
	return s.positions.Update(ctx, pos)
}

func (s HRService) DeleteJobPosition(ctx context.Context, id uint64) error {
	return s.positions.Delete(ctx, id)
}

func (s HRService) ListLeaveTypes(ctx context.Context, q *query.Query) (*query.Page[reference.LeaveType], error) {
	return s.leaveTypes.List(ctx, q)
}

func (s HRService) FindLeaveType(ctx context.Context, id uint64) (*reference.LeaveType, error) {
	return s.leaveTypes.Find(ctx, id)
}

func (s HRService) CreateLeaveType(ctx context.Context, lt *reference.LeaveType) (*reference.LeaveType, error) {
	return s.leaveTypes.Create(ctx, lt)
}

func (s HRService) UpdateLeaveType(ctx context.Context, lt *reference.LeaveType) (*reference.LeaveType, error) {
	return s.leaveTypes.Update(ctx, lt)
}

func (s HRService) DeleteLeaveType(ctx context.Context, id uint64) error {
	return s.leaveTypes.Delete(ctx, id)
}
