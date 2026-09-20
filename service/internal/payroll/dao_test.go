package payroll

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

var contractColumns = []string{"id", "employee_id", "date_start", "date_end", "wage", "wage_type", "currency_code", "state", "created_at", "updated_at", "deleted_at"}

var leaveColumns = []string{"id", "employee_id", "leave_type_id", "date_from", "date_to", "days", "state", "created_at", "updated_at", "deleted_at"}

var attendanceColumns = []string{"id", "employee_id", "check_in", "check_out", "worked_hours", "created_at", "updated_at", "deleted_at"}

var timesheetColumns = []string{"id", "employee_id", "date", "project_id", "task_id", "dimension_id", "hours", "description", "created_at", "updated_at", "deleted_at"}

var payslipColumns = []string{"id", "run_id", "employee_id", "contract_id", "gross", "net", "entry_id", "state", "created_at", "updated_at", "deleted_at"}

var payslipLineColumns = []string{"id", "payslip_id", "rule_id", "code", "name", "category", "amount", "created_at", "updated_at", "deleted_at"}

func TestEmploymentContractDAO_ListByEmployee(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewEmploymentContractDAO(db)
				now := mockTime()

				mock.ExpectQuery(`SELECT count\(\*\) FROM "employment_contracts" .*`).
					WithArgs(uint64(2)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT \* FROM "employment_contracts" .*`).
					WithArgs(uint64(2)).
					WillReturnRows(sqlmock.NewRows(contractColumns).AddRow(1, 2, now, nil, 5000000, WageTypeMonthly, "IDR", ContractStateActive, now, now, nil))

				contracts, err := dao.ListByEmployee(context.Background(), 2)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(contracts) != 1 || contracts[0].EmployeeID != 2 {
					t.Fatalf("unexpected contracts: %+v", contracts)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
		{
			name: "returns error",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewEmploymentContractDAO(db)

				mock.ExpectQuery(`SELECT count\(\*\) FROM "employment_contracts" .*`).
					WithArgs(uint64(2)).
					WillReturnError(errors.New("db down"))

				if _, err := dao.ListByEmployee(context.Background(), 2); err == nil {
					t.Fatal("expected error")
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestEmploymentContractDAO_FindActiveByEmployee(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewEmploymentContractDAO(db)
				now := mockTime()

				mock.ExpectQuery(`SELECT \* FROM "employment_contracts" WHERE employee_id = .* AND state = .*`).
					WithArgs(uint64(2), ContractStateActive, 1).
					WillReturnRows(sqlmock.NewRows(contractColumns).
						AddRow(2, 2, now, nil, 5000000, WageTypeMonthly, "IDR", ContractStateActive, now, now, nil))

				contract, err := dao.FindActiveByEmployee(context.Background(), 2)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if contract == nil || contract.ID != 2 {
					t.Fatalf("unexpected active contract: %+v", contract)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
		{
			name: "none active",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewEmploymentContractDAO(db)

				mock.ExpectQuery(`SELECT \* FROM "employment_contracts" WHERE employee_id = .* AND state = .*`).
					WithArgs(uint64(2), ContractStateActive, 1).
					WillReturnError(gorm.ErrRecordNotFound)

				contract, err := dao.FindActiveByEmployee(context.Background(), 2)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if contract != nil {
					t.Fatalf("expected nil contract, got %+v", contract)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestEmploymentContractDAO_ListInOrg(t *testing.T) {
	join := `JOIN employees ON employees\.id = employment_contracts\.employee_id AND employees\.organization_id`

	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewEmploymentContractDAO(db)
				now := mockTime()

				mock.ExpectQuery(`SELECT count\(\*\) FROM "employment_contracts" ` + join + ` .*`).
					WithArgs(uint64(10)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT "employment_contracts"\."id".* FROM "employment_contracts" ` + join + ` .*`).
					WithArgs(uint64(10)).
					WillReturnRows(sqlmock.NewRows(contractColumns).AddRow(1, 2, now, nil, 5000000, WageTypeMonthly, "IDR", ContractStateActive, now, now, nil))

				page, err := dao.ListInOrg(context.Background(), nil, 10)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if page.Count != 1 || len(page.Items) != 1 {
					t.Fatalf("unexpected page: %+v", page)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
		{
			name: "with query",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewEmploymentContractDAO(db)
				now := mockTime()
				q := &query.Query{Filters: []query.Filter{{Field: "state", Operator: query.Equal, Value: ContractStateActive}}}

				mock.ExpectQuery(`SELECT count\(\*\) FROM "employment_contracts" `+join+` .*`).
					WithArgs(uint64(10), "active").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT "employment_contracts"\."id".* FROM "employment_contracts" `+join+` .*`).
					WithArgs(uint64(10), "active").
					WillReturnRows(sqlmock.NewRows(contractColumns).AddRow(1, 2, now, nil, 5000000, WageTypeMonthly, "IDR", ContractStateActive, now, now, nil))

				page, err := dao.ListInOrg(context.Background(), q, 10)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(page.Items) != 1 {
					t.Fatalf("unexpected page: %+v", page)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
		{
			name: "count error",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewEmploymentContractDAO(db)

				mock.ExpectQuery(`SELECT count\(\*\) FROM "employment_contracts" ` + join + ` .*`).
					WithArgs(uint64(10)).
					WillReturnError(errors.New("db down"))

				if _, err := dao.ListInOrg(context.Background(), nil, 10); err == nil {
					t.Fatal("expected error")
				}
				query.AssertDBMockDone(t, mock)
			},
		},
		{
			name: "find error",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewEmploymentContractDAO(db)

				mock.ExpectQuery(`SELECT count\(\*\) FROM "employment_contracts" ` + join + ` .*`).
					WithArgs(uint64(10)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT "employment_contracts"\."id".* FROM "employment_contracts" ` + join + ` .*`).
					WithArgs(uint64(10)).
					WillReturnError(errors.New("db down"))

				if _, err := dao.ListInOrg(context.Background(), nil, 10); err == nil {
					t.Fatal("expected error")
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestEmploymentContractDAO_FindInOrg(t *testing.T) {
	join := `JOIN employees ON employees\.id = employment_contracts\.employee_id AND employees\.organization_id`

	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewEmploymentContractDAO(db)
				now := mockTime()

				mock.ExpectQuery(`SELECT "employment_contracts"\."id".* FROM "employment_contracts" `+join+` .*`).
					WithArgs(uint64(10), uint64(7), uint64(1)).
					WillReturnRows(sqlmock.NewRows(contractColumns).AddRow(7, 2, now, nil, 5000000, WageTypeMonthly, "IDR", ContractStateActive, now, now, nil))

				contract, err := dao.FindInOrg(context.Background(), 7, 10)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if contract == nil || contract.ID != 7 {
					t.Fatalf("unexpected contract: %+v", contract)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
		{
			name: "not found",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewEmploymentContractDAO(db)

				mock.ExpectQuery(`SELECT "employment_contracts"\."id".* FROM "employment_contracts" `+join+` .*`).
					WithArgs(uint64(10), uint64(7), uint64(1)).
					WillReturnError(gorm.ErrRecordNotFound)

				contract, err := dao.FindInOrg(context.Background(), 7, 10)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if contract != nil {
					t.Fatalf("expected nil contract, got %+v", contract)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
		{
			name: "error",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewEmploymentContractDAO(db)

				mock.ExpectQuery(`SELECT "employment_contracts"\."id".* FROM "employment_contracts" `+join+` .*`).
					WithArgs(uint64(10), uint64(7), uint64(1)).
					WillReturnError(errors.New("db down"))

				if _, err := dao.FindInOrg(context.Background(), 7, 10); err == nil {
					t.Fatal("expected error")
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestLeaveRequestDAO_ListApprovedByEmployeeAndType(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewLeaveRequestDAO(db)
				now := mockTime()

				mock.ExpectQuery(`SELECT count\(\*\) FROM "leave_requests" .*`).
					WithArgs(uint64(2), uint64(3), LeaveStateApproved).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT \* FROM "leave_requests" .*`).
					WithArgs(uint64(2), uint64(3), LeaveStateApproved).
					WillReturnRows(sqlmock.NewRows(leaveColumns).
						AddRow(1, 2, 3, now, now, 2, LeaveStateApproved, now, now, nil))

				requests, err := dao.ListApprovedByEmployeeAndType(context.Background(), 2, 3)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(requests) != 1 || requests[0].ID != 1 {
					t.Fatalf("unexpected requests: %+v", requests)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestLeaveRequestDAO_ListByEmployee(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "returns error",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewLeaveRequestDAO(db)

				mock.ExpectQuery(`SELECT count\(\*\) FROM "leave_requests" .*`).
					WithArgs(uint64(2)).
					WillReturnError(errors.New("db down"))

				if _, err := dao.ListByEmployee(context.Background(), 2); err == nil {
					t.Fatal("expected error")
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestAttendanceDAO_ListByEmployee(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewAttendanceDAO(db)
				now := mockTime()

				mock.ExpectQuery(`SELECT count\(\*\) FROM "attendances" .*`).
					WithArgs(uint64(2)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT \* FROM "attendances" .*`).
					WithArgs(uint64(2)).
					WillReturnRows(sqlmock.NewRows(attendanceColumns).AddRow(1, 2, &now, &now, 8.5, now, now, nil))

				records, err := dao.ListByEmployee(context.Background(), 2)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(records) != 1 || records[0].EmployeeID != 2 {
					t.Fatalf("unexpected records: %+v", records)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestTimesheetDAO_ListByProject(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewTimesheetDAO(db)
				now := mockTime()

				mock.ExpectQuery(`SELECT count\(\*\) FROM "timesheets" .*`).
					WithArgs(uint64(5)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT \* FROM "timesheets" .*`).
					WithArgs(uint64(5)).
					WillReturnRows(sqlmock.NewRows(timesheetColumns).AddRow(1, 2, now, uint64(5), nil, nil, 8, nil, now, now, nil))

				records, err := dao.ListByProject(context.Background(), 5)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(records) != 1 || records[0].ProjectID == nil || *records[0].ProjectID != 5 {
					t.Fatalf("unexpected records: %+v", records)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestTimesheetDAO_ListByEmployee(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewTimesheetDAO(db)
				now := mockTime()

				mock.ExpectQuery(`SELECT count\(\*\) FROM "timesheets" .*`).
					WithArgs(uint64(2)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT \* FROM "timesheets" .*`).
					WithArgs(uint64(2)).
					WillReturnRows(sqlmock.NewRows(timesheetColumns).AddRow(1, 2, now, nil, nil, nil, 8, nil, now, now, nil))

				records, err := dao.ListByEmployee(context.Background(), 2)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(records) != 1 {
					t.Fatalf("unexpected records: %+v", records)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestPayrollRunDAO_CreateWithPayslipsTx(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewPayrollRunDAO(db)

				run := &PayrollRun{Base: model.Base{ID: 1}, OrganizationID: 10}
				payslips := []*Payslip{{Base: model.Base{ID: 1}, EmployeeID: 2}}
				lines := [][]*PayslipLine{{&PayslipLine{Base: model.Base{ID: 3}, RuleID: 4}}}

				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payroll_runs"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payslips"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payslip_lines"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
				mock.ExpectCommit()

				created, err := dao.CreateWithPayslipsTx(context.Background(), db, run, payslips, lines)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if created.ID != 1 || payslips[0].RunID != 1 || lines[0][0].PayslipID != 1 {
					t.Fatalf("unexpected result: %+v %+v %+v", created, payslips, lines)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
		{
			name: "run error",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewPayrollRunDAO(db)

				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payroll_runs"`).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()

				_, err := dao.CreateWithPayslipsTx(context.Background(), db, &PayrollRun{}, []*Payslip{}, [][]*PayslipLine{})
				if err == nil {
					t.Fatal("expected error")
				}
				query.AssertDBMockDone(t, mock)
			},
		},
		{
			name: "payslip error",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewPayrollRunDAO(db)

				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payroll_runs"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payslips"`).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()

				_, err := dao.CreateWithPayslipsTx(context.Background(), db, &PayrollRun{Base: model.Base{ID: 1}}, []*Payslip{{}}, [][]*PayslipLine{})
				if err == nil {
					t.Fatal("expected error")
				}
				query.AssertDBMockDone(t, mock)
			},
		},
		{
			name: "line error",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewPayrollRunDAO(db)

				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payroll_runs"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payslips"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "payslip_lines"`).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()

				_, err := dao.CreateWithPayslipsTx(context.Background(), db, &PayrollRun{Base: model.Base{ID: 1}}, []*Payslip{{}}, [][]*PayslipLine{{&PayslipLine{}}})
				if err == nil {
					t.Fatal("expected error")
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestPayrollRunDAO_UpdateTx(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewPayrollRunDAO(db)

				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "payroll_runs" SET`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				updated, err := dao.UpdateTx(context.Background(), db, &PayrollRun{Base: model.Base{ID: 1}, State: RunStateConfirmed})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if updated.State != RunStateConfirmed {
					t.Fatalf("unexpected updated run: %+v", updated)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
		{
			name: "error",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewPayrollRunDAO(db)

				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "payroll_runs" SET`).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()

				if _, err := dao.UpdateTx(context.Background(), db, &PayrollRun{Base: model.Base{ID: 1}}); err == nil {
					t.Fatal("expected error")
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestPayslipDAO_ListByRun(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewPayslipDAO(db)
				now := mockTime()

				mock.ExpectQuery(`SELECT count\(\*\) FROM "payslips" .*`).
					WithArgs(uint64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT \* FROM "payslips" .*`).
					WithArgs(uint64(1)).
					WillReturnRows(sqlmock.NewRows(payslipColumns).AddRow(1, 1, 2, 3, 5000000, 4000000, nil, PayslipStateDraft, now, now, nil))

				payslips, err := dao.ListByRun(context.Background(), 1)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(payslips) != 1 || payslips[0].RunID != 1 {
					t.Fatalf("unexpected payslips: %+v", payslips)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestPayslipDAO_ListByEmployee(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewPayslipDAO(db)
				now := mockTime()

				mock.ExpectQuery(`SELECT count\(\*\) FROM "payslips" .*`).
					WithArgs(uint64(2)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT \* FROM "payslips" .*`).
					WithArgs(uint64(2)).
					WillReturnRows(sqlmock.NewRows(payslipColumns).AddRow(1, 1, 2, 3, 5000000, 4000000, nil, PayslipStateDraft, now, now, nil))

				payslips, err := dao.ListByEmployee(context.Background(), 2)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(payslips) != 1 || payslips[0].EmployeeID != 2 {
					t.Fatalf("unexpected payslips: %+v", payslips)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestPayslipDAO_UpdateTx(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewPayslipDAO(db)

				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "payslips" SET`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				updated, err := dao.UpdateTx(context.Background(), db, &Payslip{Base: model.Base{ID: 1}, State: PayslipStatePosted})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if updated.State != PayslipStatePosted {
					t.Fatalf("unexpected updated payslip: %+v", updated)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
		{
			name: "error",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewPayslipDAO(db)

				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "payslips" SET`).
					WillReturnError(errors.New("db down"))
				mock.ExpectRollback()

				if _, err := dao.UpdateTx(context.Background(), db, &Payslip{Base: model.Base{ID: 1}}); err == nil {
					t.Fatal("expected error")
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func TestPayslipLineDAO_ListByPayslip(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "success",
			fn: func(t *testing.T) {
				db, mock := query.NewMockDB(t)
				dao := NewPayslipLineDAO(db)
				now := mockTime()

				mock.ExpectQuery(`SELECT count\(\*\) FROM "payslip_lines" .*`).
					WithArgs(uint64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT \* FROM "payslip_lines" .*`).
					WithArgs(uint64(1)).
					WillReturnRows(sqlmock.NewRows(payslipLineColumns).AddRow(3, 1, 4, "BAS", "Basic", "earning", 5000000, now, now, nil))

				lines, err := dao.ListByPayslip(context.Background(), 1)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(lines) != 1 || lines[0].PayslipID != 1 {
					t.Fatalf("unexpected lines: %+v", lines)
				}
				query.AssertDBMockDone(t, mock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func mockTime() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}
