package payroll

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func EmployeeFixture(opts ...func(*Employee) *Employee) *Employee {
	now := time.Now()
	emp := &Employee{
		Base:             model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID:   helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		ContactID:        uint64(gofakeit.Number(1, 10000)),
		UserID:           helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		EmployeeNumber:   gofakeit.AppName(),
		DepartmentID:     helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		JobPositionID:    helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		ManagerID:        helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		HireDate:         helper.Ptr(time.Now()),
		TerminationDate:  nil,
		EmploymentType:   "full_time",
		WorkLocation:     helper.Ptr("office"),
		Active:           true,
		RequiresApproval: false,
	}
	for _, opt := range opts {
		opt(emp)
	}
	return emp
}

func EmploymentContractFixture(opts ...func(*EmploymentContract) *EmploymentContract) *EmploymentContract {
	now := time.Now()
	contract := &EmploymentContract{
		Base:         model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		EmployeeID:   uint64(gofakeit.Number(1, 10000)),
		DateStart:    now,
		DateEnd:      nil,
		Wage:         gofakeit.Float64Range(1, 10000),
		WageType:     "monthly",
		CurrencyCode: "IDR",
		State:        "open",
	}
	for _, opt := range opts {
		opt(contract)
	}
	return contract
}

func LeaveRequestFixture(opts ...func(*LeaveRequest) *LeaveRequest) *LeaveRequest {
	now := time.Now()
	req := &LeaveRequest{
		Base:        model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		EmployeeID:  uint64(gofakeit.Number(1, 10000)),
		LeaveTypeID: uint64(gofakeit.Number(1, 10000)),
		DateFrom:    now,
		DateTo:      now.AddDate(0, 0, 1),
		Days:        1,
		State:       "draft",
	}
	for _, opt := range opts {
		opt(req)
	}
	return req
}

func AttendanceFixture(opts ...func(*Attendance) *Attendance) *Attendance {
	now := time.Now()
	checkIn := now.Add(-8 * time.Hour)
	att := &Attendance{
		Base:                  model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID:        uint64(gofakeit.Number(1, 10000)),
		EmployeeID:            uint64(gofakeit.Number(1, 10000)),
		ShiftID:               helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		CheckIn:               helper.Ptr(checkIn),
		CheckOut:              helper.Ptr(checkIn.Add(8 * time.Hour)),
		WorkedHours:           8,
		Status:                "present",
		AttendanceType:        "regular",
		Source:                "manual",
		BreakMinutes:          60,
		OvertimeHours:         0,
		LateMinutes:           0,
		EarlyDepartureMinutes: 0,
		Notes:                 nil,
		ConfirmedAt:           helper.Ptr(now),
		Latitude:              helper.Ptr(gofakeit.Latitude()),
		Longitude:             helper.Ptr(gofakeit.Longitude()),
		DeviceID:              nil,
		IPAddress:             nil,
		UserAgent:             nil,
	}
	for _, opt := range opts {
		opt(att)
	}
	return att
}

func TimesheetFixture(opts ...func(*Timesheet) *Timesheet) *Timesheet {
	now := time.Now()
	ts := &Timesheet{
		Base:        model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		EmployeeID:  uint64(gofakeit.Number(1, 10000)),
		Date:        now,
		ProjectID:   helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		TaskID:      helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		DimensionID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Hours:       gofakeit.Float64Range(1, 12),
		Description: helper.Ptr(gofakeit.Sentence()),
	}
	for _, opt := range opts {
		opt(ts)
	}
	return ts
}

func PayrollRunFixture(opts ...func(*PayrollRun) *PayrollRun) *PayrollRun {
	now := time.Now()
	run := &PayrollRun{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: uint64(gofakeit.Number(1, 10000)),
		Name:           helper.Ptr(gofakeit.AppName()),
		PeriodStart:    now.AddDate(0, -1, 0),
		PeriodEnd:      now,
		State:          "draft",
	}
	for _, opt := range opts {
		opt(run)
	}
	return run
}

func PayslipFixture(opts ...func(*Payslip) *Payslip) *Payslip {
	now := time.Now()
	sl := &Payslip{
		Base:       model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		RunID:      uint64(gofakeit.Number(1, 10000)),
		EmployeeID: uint64(gofakeit.Number(1, 10000)),
		ContractID: uint64(gofakeit.Number(1, 10000)),
		Gross:      gofakeit.Float64Range(1, 10000),
		Net:        gofakeit.Float64Range(1, 10000),
		EntryID:    nil,
		State:      "draft",
	}
	for _, opt := range opts {
		opt(sl)
	}
	return sl
}

func PayslipLineFixture(opts ...func(*PayslipLine) *PayslipLine) *PayslipLine {
	now := time.Now()
	line := &PayslipLine{
		Base:      model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		PayslipID: uint64(gofakeit.Number(1, 10000)),
		RuleID:    uint64(gofakeit.Number(1, 10000)),
		Code:      gofakeit.AppName(),
		Name:      gofakeit.Word(),
		Category:  "earn",
		Amount:    gofakeit.Float64Range(1, 10000),
	}
	for _, opt := range opts {
		opt(line)
	}
	return line
}

func ShiftFixture(opts ...func(*Shift) *Shift) *Shift {
	now := time.Now()
	shift := &Shift{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: uint64(gofakeit.Number(1, 10000)),
		Name:           gofakeit.Word(),
		StartTime:      "08:00",
		EndTime:        "17:00",
	}
	for _, opt := range opts {
		opt(shift)
	}
	return shift
}

func ShiftAssignmentFixture(opts ...func(*ShiftAssignment) *ShiftAssignment) *ShiftAssignment {
	now := time.Now()
	sa := &ShiftAssignment{
		Base:       model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		EmployeeID: uint64(gofakeit.Number(1, 10000)),
		ShiftID:    uint64(gofakeit.Number(1, 10000)),
		Date:       now,
	}
	for _, opt := range opts {
		opt(sa)
	}
	return sa
}
