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
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func hrTestService(attendances AttendanceDAOMock, employees EmployeeDAOMock, shifts ShiftDAOMock, assignments ShiftAssignmentDAOMock, orgs dao.CRUDMock[reference.Organization]) HRService {
	return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, orgs, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, TimesheetDAOMock{}, shifts, assignments, inventory.TransactionerMock{})
}

func openAttendance() *Attendance {
	checkIn := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	return &Attendance{Base: model.Base{ID: 1}, EmployeeID: 7, OrganizationID: 10, CheckIn: &checkIn, Status: AttendanceStatusDraft}
}

func TestHRService_UpdateDeleteAttendance(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(att *Attendance, findErr error, updateErr error, deleteErr error) HRService {
		return hrTestService(AttendanceDAOMock{
			CRUDMock: dao.CRUDMock[Attendance]{
				FindFunc:   func(_ context.Context, _ uint64) (*Attendance, error) { return att, findErr },
				UpdateFunc: func(_ context.Context, a *Attendance) (*Attendance, error) { return a, updateErr },
				DeleteFunc: func(_ context.Context, _ uint64) error { return deleteErr },
			},
		}, EmployeeDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, dao.CRUDMock[reference.Organization]{})
	}

	t.Run("updates attendance", func(t *testing.T) {
		svc := newSvc(openAttendance(), nil, nil, nil)
		got, err := svc.UpdateAttendance(ctx, &Attendance{Base: model.Base{ID: 1}, BreakMinutes: 30})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.BreakMinutes != 30 {
			t.Errorf("break = %d", got.BreakMinutes)
		}
	})

	t.Run("update failures", func(t *testing.T) {
		cancelled := openAttendance()
		cancelled.Status = AttendanceStatusCancelled
		cases := []struct {
			name      string
			att       *Attendance
			findErr   error
			updateErr error
			wantErr   error
		}{
			{name: "lookup error", findErr: dbErr, wantErr: dbErr},
			{name: "missing", wantErr: ErrAttendanceNotFound},
			{name: "cancelled", att: cancelled, wantErr: ErrAttendanceState},
			{name: "update error", att: openAttendance(), updateErr: dbErr, wantErr: dbErr},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc := newSvc(tt.att, tt.findErr, tt.updateErr, nil)
				_, err := svc.UpdateAttendance(ctx, &Attendance{Base: model.Base{ID: 1}})
				helper.AssertError(t, err, true, tt.wantErr)
			})
		}
	})

	t.Run("deletes attendance", func(t *testing.T) {
		svc := newSvc(openAttendance(), nil, nil, nil)
		if err := svc.DeleteAttendance(ctx, 1); helper.AssertError(t, err, false, nil) {
			return
		}
	})

	t.Run("delete failures", func(t *testing.T) {
		cases := []struct {
			name      string
			att       *Attendance
			findErr   error
			deleteErr error
			wantErr   error
		}{
			{name: "lookup error", findErr: dbErr, wantErr: dbErr},
			{name: "missing", wantErr: ErrAttendanceNotFound},
			{name: "delete error", att: openAttendance(), deleteErr: dbErr, wantErr: dbErr},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc := newSvc(tt.att, tt.findErr, nil, tt.deleteErr)
				err := svc.DeleteAttendance(ctx, 1)
				helper.AssertError(t, err, true, tt.wantErr)
			})
		}
	})
}

func TestHRService_BulkCheckOut_Missing(t *testing.T) {
	ctx := context.Background()
	checkOut := time.Date(2026, 8, 1, 17, 30, 0, 0, time.UTC)

	t.Run("bulk checks out", func(t *testing.T) {
		svc := hrTestService(AttendanceDAOMock{
			CRUDMock: dao.CRUDMock[Attendance]{
				FindFunc:   func(_ context.Context, _ uint64) (*Attendance, error) { return openAttendance(), nil },
				UpdateFunc: func(_ context.Context, a *Attendance) (*Attendance, error) { return a, nil },
			},
		}, EmployeeDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, dao.CRUDMock[reference.Organization]{})
		got, err := svc.BulkCheckOut(ctx, []uint64{1, 2}, checkOut, 0)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(got) != 2 {
			t.Errorf("len = %d", len(got))
		}
	})

	t.Run("bulk propagates error", func(t *testing.T) {
		svc := hrTestService(AttendanceDAOMock{
			CRUDMock: dao.CRUDMock[Attendance]{
				FindFunc: func(_ context.Context, _ uint64) (*Attendance, error) { return nil, errors.New("db down") },
			},
		}, EmployeeDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, dao.CRUDMock[reference.Organization]{})
		_, err := svc.BulkCheckOut(ctx, []uint64{1}, checkOut, 0)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("lists missing", func(t *testing.T) {
		svc := hrTestService(AttendanceDAOMock{
			ListMissingByDateFunc: func(_ context.Context, _ uint64, _ string) ([]*Employee, error) {
				return []*Employee{{Base: model.Base{ID: 7}}}, nil
			},
		}, EmployeeDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, dao.CRUDMock[reference.Organization]{})
		got, err := svc.ListMissingAttendance(ctx, 10, "2026-08-01")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(got) != 1 {
			t.Errorf("len = %d", len(got))
		}

		failing := hrTestService(AttendanceDAOMock{
			ListMissingByDateFunc: func(_ context.Context, _ uint64, _ string) ([]*Employee, error) {
				return nil, errors.New("db down")
			},
		}, EmployeeDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, dao.CRUDMock[reference.Organization]{})
		_, err = failing.ListMissingAttendance(ctx, 10, "2026-08-01")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestHRService_AutoCheckout_MarkAbsent(t *testing.T) {
	ctx := context.Background()
	checkIn := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)

	newAutoSvc := func(attendances AttendanceDAOWithOrgMock, employees EmployeeDAOMock) HRService {
		return NewHRService(employees, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
	}

	t.Run("auto checks out open attendances", func(t *testing.T) {
		att := &Attendance{Base: model.Base{ID: 1}, EmployeeID: 7, CheckIn: &checkIn, Status: AttendanceStatusDraft}
		svc := newAutoSvc(AttendanceDAOWithOrgMock{
			AttendanceDAOMock: AttendanceDAOMock{
				CRUDMock: dao.CRUDMock[Attendance]{
					FindFunc:   func(_ context.Context, _ uint64) (*Attendance, error) { return att, nil },
					UpdateFunc: func(_ context.Context, a *Attendance) (*Attendance, error) { return a, nil },
				},
			},
			ListOpenByOrganizationFunc: func(_ context.Context, _ uint64) ([]*Attendance, error) {
				return []*Attendance{att, {Base: model.Base{ID: 2}, EmployeeID: 8}}, nil
			},
		}, EmployeeDAOMock{
			CRUDMock: dao.CRUDMock[Employee]{
				FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
					return &Employee{Base: model.Base{ID: 7}, Active: true}, nil
				},
			},
		})
		got, err := svc.AutoCheckout(ctx, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(got) != 1 {
			t.Errorf("len = %d, want 1 (no-checkin skipped)", len(got))
		}
	})

	t.Run("auto checkout skips inactive and propagates list error", func(t *testing.T) {
		att := &Attendance{Base: model.Base{ID: 1}, EmployeeID: 7, CheckIn: &checkIn, Status: AttendanceStatusDraft}
		svc := newAutoSvc(AttendanceDAOWithOrgMock{
			ListOpenByOrganizationFunc: func(_ context.Context, _ uint64) ([]*Attendance, error) {
				return []*Attendance{att}, nil
			},
		}, EmployeeDAOMock{
			CRUDMock: dao.CRUDMock[Employee]{
				FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
					return &Employee{Base: model.Base{ID: 7}, Active: false}, nil
				},
			},
		})
		got, err := svc.AutoCheckout(ctx, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(got) != 0 {
			t.Errorf("len = %d, want 0", len(got))
		}

		failing := newAutoSvc(AttendanceDAOWithOrgMock{
			ListOpenByOrganizationFunc: func(_ context.Context, _ uint64) ([]*Attendance, error) {
				return nil, errors.New("db down")
			},
		}, EmployeeDAOMock{})
		_, err = failing.AutoCheckout(ctx, 10)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("marks absent", func(t *testing.T) {
		svc := hrTestService(AttendanceDAOMock{
			ListMissingByDateFunc: func(_ context.Context, _ uint64, _ string) ([]*Employee, error) {
				return []*Employee{{Base: model.Base{ID: 7}}, {Base: model.Base{ID: 8}}}, nil
			},
			CRUDMock: dao.CRUDMock[Attendance]{
				CreateFunc: func(_ context.Context, a *Attendance) (*Attendance, error) { return a, nil },
			},
		}, EmployeeDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, dao.CRUDMock[reference.Organization]{})
		got, err := svc.MarkAbsent(ctx, 10, "2026-08-01")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(got) != 2 || got[0].Status != AttendanceStatusAbsent {
			t.Errorf("marked = %+v", got)
		}
	})

	t.Run("mark absent propagates errors", func(t *testing.T) {
		failing := hrTestService(AttendanceDAOMock{
			ListMissingByDateFunc: func(_ context.Context, _ uint64, _ string) ([]*Employee, error) {
				return nil, errors.New("db down")
			},
		}, EmployeeDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, dao.CRUDMock[reference.Organization]{})
		_, err := failing.MarkAbsent(ctx, 10, "2026-08-01")
		if err == nil {
			t.Error("expected error, got nil")
		}

		creating := hrTestService(AttendanceDAOMock{
			ListMissingByDateFunc: func(_ context.Context, _ uint64, _ string) ([]*Employee, error) {
				return []*Employee{{Base: model.Base{ID: 7}}}, nil
			},
			CRUDMock: dao.CRUDMock[Attendance]{
				CreateFunc: func(_ context.Context, _ *Attendance) (*Attendance, error) { return nil, errors.New("db down") },
			},
		}, EmployeeDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, dao.CRUDMock[reference.Organization]{})
		_, err = creating.MarkAbsent(ctx, 10, "2026-08-01")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestHRService_Shifts(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(shift *Shift, findErr error, writeErr error) HRService {
		return hrTestService(AttendanceDAOMock{}, EmployeeDAOMock{}, ShiftDAOMock{
			CRUDMock: dao.CRUDMock[Shift]{
				FindFunc:   func(_ context.Context, _ uint64) (*Shift, error) { return shift, findErr },
				CreateFunc: func(_ context.Context, s *Shift) (*Shift, error) { return s, writeErr },
				UpdateFunc: func(_ context.Context, s *Shift) (*Shift, error) { return s, writeErr },
				DeleteFunc: func(_ context.Context, _ uint64) error { return writeErr },
			},
		}, ShiftAssignmentDAOMock{}, dao.CRUDMock[reference.Organization]{})
	}
	validShift := func() *Shift {
		return &Shift{Base: model.Base{ID: 1}, OrganizationID: 10, Name: "Morning", StartTime: "08:00", EndTime: "17:00"}
	}

	t.Run("creates shift", func(t *testing.T) {
		svc := newSvc(nil, nil, nil)
		got, err := svc.CreateShift(ctx, validShift())
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.Name != "Morning" {
			t.Errorf("name = %q", got.Name)
		}
	})

	t.Run("create shift validations", func(t *testing.T) {
		svc := newSvc(nil, nil, nil)
		cases := []struct {
			name    string
			shift   *Shift
			wantErr error
		}{
			{name: "nil", shift: nil, wantErr: ErrShiftNotFound},
			{name: "empty name", shift: &Shift{OrganizationID: 10, StartTime: "08:00", EndTime: "17:00"}, wantErr: ErrShiftName},
			{name: "empty times", shift: &Shift{OrganizationID: 10, Name: "Morning"}, wantErr: ErrShiftTime},
			{name: "no org", shift: &Shift{Name: "Morning", StartTime: "08:00", EndTime: "17:00"}, wantErr: ErrShiftOrganization},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				_, err := svc.CreateShift(ctx, tt.shift)
				helper.AssertError(t, err, true, tt.wantErr)
			})
		}
	})

	t.Run("updates shift", func(t *testing.T) {
		svc := newSvc(validShift(), nil, nil)
		got, err := svc.UpdateShift(ctx, validShift())
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.Name != "Morning" {
			t.Errorf("name = %q", got.Name)
		}
	})

	t.Run("update shift failures", func(t *testing.T) {
		cases := []struct {
			name    string
			shift   *Shift
			findErr error
			stored  *Shift
			wantErr error
		}{
			{name: "lookup error", shift: validShift(), findErr: dbErr, wantErr: dbErr},
			{name: "missing", shift: validShift(), wantErr: ErrShiftNotFound},
			{name: "empty name", shift: &Shift{Base: model.Base{ID: 1}}, stored: validShift(), wantErr: ErrShiftName},
			{name: "empty times", shift: &Shift{Base: model.Base{ID: 1}, Name: "Morning"}, stored: validShift(), wantErr: ErrShiftTime},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc := newSvc(tt.stored, tt.findErr, nil)
				_, err := svc.UpdateShift(ctx, tt.shift)
				helper.AssertError(t, err, true, tt.wantErr)
			})
		}
	})

	t.Run("deletes shift", func(t *testing.T) {
		svc := newSvc(validShift(), nil, nil)
		if err := svc.DeleteShift(ctx, 1); helper.AssertError(t, err, false, nil) {
			return
		}
	})

	t.Run("delete shift failures", func(t *testing.T) {
		svc := newSvc(nil, dbErr, nil)
		helper.AssertError(t, svc.DeleteShift(ctx, 1), true, dbErr)

		svc = newSvc(nil, nil, nil)
		helper.AssertError(t, svc.DeleteShift(ctx, 1), true, ErrShiftNotFound)

		svc = newSvc(validShift(), nil, dbErr)
		helper.AssertError(t, svc.DeleteShift(ctx, 1), true, dbErr)
	})
}

func TestHRService_ShiftAssignments(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	workDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	newSvc := func(emp *Employee, empErr error, shift *Shift, shiftErr error, existing *ShiftAssignment, existErr error, writeErr error) HRService {
		return NewHRService(EmployeeDAOMock{
			CRUDMock: dao.CRUDMock[Employee]{
				FindFunc: func(_ context.Context, _ uint64) (*Employee, error) { return emp, empErr },
			},
		}, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{
			CRUDMock: dao.CRUDMock[Shift]{
				FindFunc: func(_ context.Context, _ uint64) (*Shift, error) { return shift, shiftErr },
			},
		}, ShiftAssignmentDAOMock{
			CRUDMock: dao.CRUDMock[ShiftAssignment]{
				FindFunc:   func(_ context.Context, _ uint64) (*ShiftAssignment, error) { return existing, existErr },
				CreateFunc: func(_ context.Context, a *ShiftAssignment) (*ShiftAssignment, error) { return a, writeErr },
				DeleteFunc: func(_ context.Context, _ uint64) error { return writeErr },
			},
			FindByEmployeeAndDateFunc: func(_ context.Context, _ uint64, _ string) (*ShiftAssignment, error) { return existing, existErr },
		}, inventory.TransactionerMock{})
	}
	emp := &Employee{Base: model.Base{ID: 7}, Active: true}
	shift := &Shift{Base: model.Base{ID: 3}, Name: "Morning", StartTime: "08:00", EndTime: "17:00"}
	newAssignment := func() *ShiftAssignment {
		return &ShiftAssignment{EmployeeID: 7, ShiftID: 3, Date: workDate}
	}

	t.Run("creates assignment", func(t *testing.T) {
		svc := newSvc(emp, nil, shift, nil, nil, nil, nil)
		got, err := svc.CreateShiftAssignment(ctx, newAssignment())
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.EmployeeID != 7 {
			t.Errorf("employee = %d", got.EmployeeID)
		}
	})

	t.Run("create assignment failures", func(t *testing.T) {
		cases := []struct {
			name     string
			assign   *ShiftAssignment
			emp      *Employee
			empErr   error
			shift    *Shift
			shiftErr error
			existing *ShiftAssignment
			existErr error
			wantErr  error
		}{
			{name: "nil", wantErr: ErrShiftAssignmentNotFound},
			{name: "employee error", assign: newAssignment(), empErr: dbErr, wantErr: dbErr},
			{name: "employee missing", assign: newAssignment(), wantErr: ErrShiftAssignmentEmployee},
			{name: "shift error", assign: newAssignment(), emp: emp, shiftErr: dbErr, wantErr: dbErr},
			{name: "shift missing", assign: newAssignment(), emp: emp, wantErr: ErrShiftAssignmentShift},
			{name: "lookup error", assign: newAssignment(), emp: emp, shift: shift, existErr: dbErr, wantErr: dbErr},
			{name: "already assigned", assign: newAssignment(), emp: emp, shift: shift, existing: &ShiftAssignment{}, wantErr: ErrShiftAssignmentExists},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc := newSvc(tt.emp, tt.empErr, tt.shift, tt.shiftErr, tt.existing, tt.existErr, nil)
				_, err := svc.CreateShiftAssignment(ctx, tt.assign)
				helper.AssertError(t, err, true, tt.wantErr)
			})
		}
	})

	t.Run("deletes assignment", func(t *testing.T) {
		svc := newSvc(nil, nil, nil, nil, &ShiftAssignment{}, nil, nil)
		if err := svc.DeleteShiftAssignment(ctx, 1); helper.AssertError(t, err, false, nil) {
			return
		}
	})

	t.Run("delete assignment failures", func(t *testing.T) {
		svc := newSvc(nil, nil, nil, nil, nil, dbErr, nil)
		helper.AssertError(t, svc.DeleteShiftAssignment(ctx, 1), true, dbErr)

		svc = newSvc(nil, nil, nil, nil, nil, nil, nil)
		helper.AssertError(t, svc.DeleteShiftAssignment(ctx, 1), true, ErrShiftAssignmentNotFound)

		svc = newSvc(nil, nil, nil, nil, &ShiftAssignment{}, nil, dbErr)
		helper.AssertError(t, svc.DeleteShiftAssignment(ctx, 1), true, dbErr)
	})
}

func TestHRService_Rounding(t *testing.T) {
	ctx := context.Background()

	t.Run("rounding minutes from organization", func(t *testing.T) {
		withOrg := hrTestService(AttendanceDAOMock{}, EmployeeDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
				return &reference.Organization{RoundingMinutes: 15}, nil
			},
		})
		if got := withOrg.getOrganizationRounding(ctx, 10); got != 15 {
			t.Errorf("rounding = %d, want 15", got)
		}

		missing := hrTestService(AttendanceDAOMock{}, EmployeeDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, dao.CRUDMock[reference.Organization]{})
		if got := missing.getOrganizationRounding(ctx, 10); got != 0 {
			t.Errorf("rounding = %d, want 0", got)
		}

		failing := hrTestService(AttendanceDAOMock{}, EmployeeDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, dao.CRUDMock[reference.Organization]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) { return nil, errors.New("db down") },
		})
		if got := failing.getOrganizationRounding(ctx, 10); got != 0 {
			t.Errorf("rounding = %d, want 0", got)
		}
	})

	t.Run("rounds to nearest", func(t *testing.T) {
		if got := roundToNearest(8.0, 0); got != 8.0 {
			t.Errorf("round = %v", got)
		}
		if got := roundToNearest(8.1, 15); got != 8.25 {
			t.Errorf("round = %v, want 8.25", got)
		}
		if got := roundToNearest(8.0, 15); got != 8.0 {
			t.Errorf("round = %v, want 8.0", got)
		}
	})
}

func TestHRService_MasterData(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	newSvc := func() HRService {
		return NewHRService(EmployeeDAOMock{}, EmploymentContractDAOMock{}, LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, AttendanceDAOMock{}, TimesheetDAOMock{}, ShiftDAOMock{}, ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
	}

	t.Run("departments", func(t *testing.T) {
		svc := newSvc()
		if _, err := svc.ListDepartments(ctx, q); err != nil {
			t.Errorf("list = %v", err)
		}
		if _, err := svc.FindDepartment(ctx, 1); err != nil {
			t.Errorf("find = %v", err)
		}
		if _, err := svc.CreateDepartment(ctx, &reference.Department{}); err != nil {
			t.Errorf("create = %v", err)
		}
		if _, err := svc.UpdateDepartment(ctx, &reference.Department{}); err != nil {
			t.Errorf("update = %v", err)
		}
		if err := svc.DeleteDepartment(ctx, 1); err != nil {
			t.Errorf("delete = %v", err)
		}
	})

	t.Run("job positions", func(t *testing.T) {
		svc := newSvc()
		if _, err := svc.ListJobPositions(ctx, q); err != nil {
			t.Errorf("list = %v", err)
		}
		if _, err := svc.FindJobPosition(ctx, 1); err != nil {
			t.Errorf("find = %v", err)
		}
		if _, err := svc.CreateJobPosition(ctx, &reference.JobPosition{}); err != nil {
			t.Errorf("create = %v", err)
		}
		if _, err := svc.UpdateJobPosition(ctx, &reference.JobPosition{}); err != nil {
			t.Errorf("update = %v", err)
		}
		if err := svc.DeleteJobPosition(ctx, 1); err != nil {
			t.Errorf("delete = %v", err)
		}
	})

	t.Run("leave types", func(t *testing.T) {
		svc := newSvc()
		if _, err := svc.ListLeaveTypes(ctx, q); err != nil {
			t.Errorf("list = %v", err)
		}
		if _, err := svc.FindLeaveType(ctx, 1); err != nil {
			t.Errorf("find = %v", err)
		}
		if _, err := svc.CreateLeaveType(ctx, &reference.LeaveType{}); err != nil {
			t.Errorf("create = %v", err)
		}
		if _, err := svc.UpdateLeaveType(ctx, &reference.LeaveType{}); err != nil {
			t.Errorf("update = %v", err)
		}
		if err := svc.DeleteLeaveType(ctx, 1); err != nil {
			t.Errorf("delete = %v", err)
		}
	})
}
