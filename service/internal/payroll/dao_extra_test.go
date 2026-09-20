package payroll

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestPayrollDAO_Constructors(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewEmployeeDAO(db) == nil {
		t.Error("employees = nil")
	}
	if NewEmploymentContractDAO(db) == nil {
		t.Error("contracts = nil")
	}
	if NewLeaveRequestDAO(db) == nil {
		t.Error("leaves = nil")
	}
	if NewAttendanceDAO(db) == nil {
		t.Error("attendances = nil")
	}
	if NewTimesheetDAO(db) == nil {
		t.Error("timesheets = nil")
	}
	if NewPayrollRunDAO(db) == nil {
		t.Error("runs = nil")
	}
	if NewPayslipDAO(db) == nil {
		t.Error("payslips = nil")
	}
	if NewPayslipLineDAO(db) == nil {
		t.Error("lines = nil")
	}
	if NewShiftDAO(db) == nil {
		t.Error("shifts = nil")
	}
	if NewShiftAssignmentDAO(db) == nil {
		t.Error("assignments = nil")
	}
	if NewAttendanceDAOWithOrg(db) == nil {
		t.Error("attendances with org = nil")
	}
}

func countSelectPair(mock sqlmock.Sqlmock, table string, count int64, columns []string, rows ...[]driver.Value) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "` + table + `"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
	r := sqlmock.NewRows(columns)
	for _, row := range rows {
		r.AddRow(row...)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "` + table + `"`)).
		WillReturnRows(r)
}

func TestEmploymentContractDAO_Queries(t *testing.T) {
	ctx := context.Background()

	t.Run("lists by employee", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		countSelectPair(mock, "employment_contracts", 1, []string{"id", "employee_id"}, []driver.Value{1, 7})

		items, err := NewEmploymentContractDAO(db).ListByEmployee(ctx, 7)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("finds active", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employment_contracts"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		got, err := NewEmploymentContractDAO(db).FindActiveByEmployee(ctx, 7)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 {
			t.Errorf("id = %d", got.ID)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("returns nil when no active", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employment_contracts"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		got, err := NewEmploymentContractDAO(db).FindActiveByEmployee(ctx, 7)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != nil {
			t.Errorf("contract = %+v, want nil", got)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates active error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employment_contracts"`)).WillReturnError(dbErr)

		_, err := NewEmploymentContractDAO(db).FindActiveByEmployee(ctx, 7)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists in org", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "employment_contracts" JOIN employees`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "employment_contracts" JOIN employees`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		page, err := NewEmploymentContractDAO(db).ListInOrg(ctx, nil, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if page.Count != 1 {
			t.Errorf("count = %d", page.Count)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates org list error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "employment_contracts" JOIN employees`)).WillReturnError(dbErr)

		_, err := NewEmploymentContractDAO(db).ListInOrg(ctx, nil, 10)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("finds in org", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "employment_contracts" JOIN employees`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		got, err := NewEmploymentContractDAO(db).FindInOrg(ctx, 1, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 {
			t.Errorf("id = %d", got.ID)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("returns nil when not in org", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "employment_contracts" JOIN employees`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		got, err := NewEmploymentContractDAO(db).FindInOrg(ctx, 1, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != nil {
			t.Errorf("contract = %+v, want nil", got)
		}
		query.AssertDBMockDone(t, mock)
	})
}

func TestLeaveRequestDAO_Queries(t *testing.T) {
	ctx := context.Background()

	t.Run("lists by employee", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		countSelectPair(mock, "leave_requests", 1, []string{"id"}, []driver.Value{1})

		items, err := NewLeaveRequestDAO(db).ListByEmployee(ctx, 7)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists approved by type", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		countSelectPair(mock, "leave_requests", 2, []string{"id"}, []driver.Value{1}, []driver.Value{2})

		items, err := NewLeaveRequestDAO(db).ListApprovedByEmployeeAndType(ctx, 7, 3)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 2 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists approved between dates", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "leave_requests"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		items, err := NewLeaveRequestDAO(db).ListApprovedByEmployeeBetweenDates(ctx, 7, "2026-08-01", "2026-08-31")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates list error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "leave_requests"`)).WillReturnError(dbErr)

		_, err := NewLeaveRequestDAO(db).ListApprovedByEmployeeBetweenDates(ctx, 7, "2026-08-01", "2026-08-31")
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}

func TestAttendanceDAO_Queries(t *testing.T) {
	ctx := context.Background()

	t.Run("lists by employee", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		countSelectPair(mock, "attendances", 1, []string{"id"}, []driver.Value{1})

		items, err := NewAttendanceDAO(db).ListByEmployee(ctx, 7)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("finds open by date", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		got, err := NewAttendanceDAO(db).FindOpenByEmployeeAndDate(ctx, 7, "2026-08-01")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 {
			t.Errorf("id = %d", got.ID)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("returns nil when none open", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		got, err := NewAttendanceDAO(db).FindOpenByEmployeeAndDate(ctx, 7, "2026-08-01")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != nil {
			t.Errorf("attendance = %+v, want nil", got)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists missing by date", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employees"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

		items, err := NewAttendanceDAO(db).ListMissingByDate(ctx, 10, "2026-08-01")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates missing error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "employees"`)).WillReturnError(dbErr)

		_, err := NewAttendanceDAO(db).ListMissingByDate(ctx, 10, "2026-08-01")
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists by date range", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))

		items, err := NewAttendanceDAO(db).ListByEmployeeAndDateRange(ctx, 7, "2026-08-01", "2026-08-31")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 2 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates range error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attendances"`)).WillReturnError(dbErr)

		_, err := NewAttendanceDAO(db).ListByEmployeeAndDateRange(ctx, 7, "2026-08-01", "2026-08-31")
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists employee ids by date", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT employee_id FROM "attendances"`)).
			WillReturnRows(sqlmock.NewRows([]string{"employee_id"}).AddRow(7))

		ids, err := NewAttendanceDAO(db).ListEmployeeIDsByDate(ctx, 10, "2026-08-01")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(ids) != 1 || ids[0] != 7 {
			t.Errorf("ids = %v", ids)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates ids error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "attendances"`)).WillReturnError(dbErr)

		_, err := NewAttendanceDAO(db).ListEmployeeIDsByDate(ctx, 10, "2026-08-01")
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}

func TestTimesheetDAO_Queries(t *testing.T) {
	ctx := context.Background()

	t.Run("lists by employee", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		countSelectPair(mock, "timesheets", 1, []string{"id"}, []driver.Value{1})

		items, err := NewTimesheetDAO(db).ListByEmployee(ctx, 7)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists by project", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		countSelectPair(mock, "timesheets", 1, []string{"id"}, []driver.Value{1})

		items, err := NewTimesheetDAO(db).ListByProject(ctx, 9)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "timesheets"`)).WillReturnError(dbErr)

		_, err := NewTimesheetDAO(db).ListByEmployee(ctx, 7)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}

func TestPayrollRunDAO_Tx(t *testing.T) {
	ctx := context.Background()

	t.Run("creates with payslips", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payroll_runs"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payslips"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payslip_lines"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(21))
		mock.ExpectCommit()

		tx := db.Begin()
		run := &PayrollRun{OrganizationID: 10}
		payslips := []*Payslip{{EmployeeID: 7}}
		lines := [][]*PayslipLine{{{RuleID: 1}}}
		got, err := NewPayrollRunDAO(db).CreateWithPayslipsTx(ctx, tx, run, payslips, lines)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 || payslips[0].RunID != 1 || lines[0][0].PayslipID != 11 {
			t.Errorf("run = %+v payslip = %+v line = %+v", got, payslips[0], lines[0][0])
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("commit = %v", err)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates create error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payroll_runs"`)).WillReturnError(dbErr)

		tx := db.Begin()
		_, err := NewPayrollRunDAO(db).CreateWithPayslipsTx(ctx, tx, &PayrollRun{}, nil, nil)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("updates in tx", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payroll_runs"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		tx := db.Begin()
		got, err := NewPayrollRunDAO(db).UpdateTx(ctx, tx, &PayrollRun{Base: model.Base{ID: 5}})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got == nil {
			t.Error("run = nil")
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("commit = %v", err)
		}
		query.AssertDBMockDone(t, mock)
	})
}

func TestPayslipDAO_Queries(t *testing.T) {
	ctx := context.Background()

	t.Run("lists by run", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		countSelectPair(mock, "payslips", 1, []string{"id"}, []driver.Value{1})

		items, err := NewPayslipDAO(db).ListByRun(ctx, 5)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists by employee", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		countSelectPair(mock, "payslips", 1, []string{"id"}, []driver.Value{1})

		items, err := NewPayslipDAO(db).ListByEmployee(ctx, 7)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("updates in tx", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payslips"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		tx := db.Begin()
		got, err := NewPayslipDAO(db).UpdateTx(ctx, tx, &Payslip{Base: model.Base{ID: 11}})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got == nil {
			t.Error("payslip = nil")
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("commit = %v", err)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payslips"`)).WillReturnError(dbErr)

		_, err := NewPayslipDAO(db).ListByRun(ctx, 5)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}

func TestPayslipLineDAO_List_Error(t *testing.T) {
	ctx := context.Background()

	db, mock := query.NewMockDB(t)
	dbErr := errors.New("db down")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payslip_lines"`)).WillReturnError(dbErr)

	_, err := NewPayslipLineDAO(db).ListByPayslip(ctx, 11)
	helper.AssertError(t, err, true, dbErr)
	query.AssertDBMockDone(t, mock)
}

func TestShiftAssignmentDAO_FindByDate(t *testing.T) {
	ctx := context.Background()

	t.Run("finds assignment", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shift_assignments"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "employee_id"}).AddRow(1, 7))

		got, err := NewShiftAssignmentDAO(db).FindByEmployeeAndDate(ctx, 7, "2026-08-01")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.EmployeeID != 7 {
			t.Errorf("employee = %d", got.EmployeeID)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("returns nil when missing", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shift_assignments"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		got, err := NewShiftAssignmentDAO(db).FindByEmployeeAndDate(ctx, 7, "2026-08-01")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != nil {
			t.Errorf("assignment = %+v, want nil", got)
		}
		query.AssertDBMockDone(t, mock)
	})
}

func TestAttendanceDAOWithOrg_ListOpen(t *testing.T) {
	ctx := context.Background()

	t.Run("lists open", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "attendances"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		items, err := NewAttendanceDAOWithOrg(db).ListOpenByOrganization(ctx, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		dbErr := errors.New("db down")
		mock.ExpectQuery(regexp.QuoteMeta(`FROM "attendances"`)).WillReturnError(dbErr)

		_, err := NewAttendanceDAOWithOrg(db).ListOpenByOrganization(ctx, 10)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}

var _ = time.Now
