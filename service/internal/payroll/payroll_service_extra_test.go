package payroll

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func newPayrollTestService(runs PayrollRunDAO, payslips PayslipDAO, lines PayslipLineDAO, employees EmployeeDAO, contracts EmploymentContractDAO, rules dao.CRUD[reference.SalaryRule], journals dao.CRUD[reference.Journal], poster accounting.Poster, tx inventory.TransactionerMock) PayrollService {
	return NewPayrollService(runs, payslips, lines, employees, contracts, AttendanceDAOMock{}, rules, journals, poster, sequence.NewSequenceService(payrollSequenceMock()), tx)
}

func TestSalaryRuleService_Create(t *testing.T) {
	ctx := context.Background()
	svc := NewSalaryRuleService(dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Account]{})

	tests := []struct {
		name      string
		rule      *reference.SalaryRule
		wantErr   bool
		errTarget error
	}{
		{
			name:      "rejects missing code",
			rule:      &reference.SalaryRule{},
			wantErr:   true,
			errTarget: ErrRuleCode,
		},
		{
			name:      "rejects invalid category",
			rule:      &reference.SalaryRule{Code: "SAL", Category: helper.Ptr("bogus"), ComputeType: helper.Ptr("fixed")},
			wantErr:   true,
			errTarget: ErrRuleCategory,
		},
		{
			name:      "rejects missing accounts",
			rule:      &reference.SalaryRule{Code: "SAL", Category: helper.Ptr("earning"), ComputeType: helper.Ptr("fixed")},
			wantErr:   true,
			errTarget: ErrRuleAccounts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := svc.Create(ctx, tt.rule); helper.AssertError(t, err, tt.wantErr, tt.errTarget) {
				return
			}
		})
	}

	t.Run("rejects account error", func(t *testing.T) {
		accounts := dao.CRUDMock[reference.Account]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Account, error) {
				return nil, errors.New("db down")
			},
		}
		svc := NewSalaryRuleService(dao.CRUDMock[reference.SalaryRule]{}, accounts)
		debit := uint64(1)
		credit := uint64(2)
		_, err := svc.Create(ctx, &reference.SalaryRule{Code: "SAL", Category: helper.Ptr("earning"), ComputeType: helper.Ptr("fixed"), AccountDebitID: &debit, AccountCreditID: &credit})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestSalaryRuleService_Update(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		rules     dao.CRUDMock[reference.SalaryRule]
		rule      *reference.SalaryRule
		wantErr   bool
		errTarget error
		wantCode  string
	}{
		{
			name: "updates rule",
			rules: dao.CRUDMock[reference.SalaryRule]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
					return &reference.SalaryRule{Base: model.Base{ID: 1}, Code: "OLD"}, nil
				},
				UpdateFunc: func(_ context.Context, rule *reference.SalaryRule) (*reference.SalaryRule, error) {
					return rule, nil
				},
			},
			rule: func() *reference.SalaryRule {
				debit := uint64(1)
				credit := uint64(2)
				return &reference.SalaryRule{Base: model.Base{ID: 1}, Code: "NEW", Category: helper.Ptr("earning"), ComputeType: helper.Ptr("fixed"), AccountDebitID: &debit, AccountCreditID: &credit}
			}(),
			wantErr:  false,
			wantCode: "NEW",
		},
		{
			name: "rejects missing",
			rules: dao.CRUDMock[reference.SalaryRule]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
					return nil, nil
				},
			},
			rule: func() *reference.SalaryRule {
				debit := uint64(1)
				credit := uint64(2)
				return &reference.SalaryRule{Base: model.Base{ID: 1}, Code: "NEW", Category: helper.Ptr("earning"), ComputeType: helper.Ptr("fixed"), AccountDebitID: &debit, AccountCreditID: &credit}
			}(),
			wantErr:   true,
			errTarget: ErrRuleNotFound,
		},
		{
			name: "rejects invalid code",
			rules: dao.CRUDMock[reference.SalaryRule]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
					return &reference.SalaryRule{Base: model.Base{ID: 1}}, nil
				},
			},
			rule:      &reference.SalaryRule{Base: model.Base{ID: 1}},
			wantErr:   true,
			errTarget: ErrRuleCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts := dao.CRUDMock[reference.Account]{}
			if tt.wantCode != "" {
				accounts = dao.CRUDMock[reference.Account]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Account, error) {
						return &reference.Account{Base: model.Base{ID: 1}}, nil
					},
				}
			}
			svc := NewSalaryRuleService(tt.rules, accounts)
			rule, err := svc.Update(ctx, tt.rule)
			if helper.AssertError(t, err, tt.wantErr, tt.errTarget) {
				return
			}
			if tt.wantCode != "" && rule.Code != tt.wantCode {
				t.Fatalf("expected updated rule, got %+v", rule)
			}
		})
	}
}

func TestPayrollService_CreateRun(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		req       CreateRunRequest
		svc       PayrollService
		wantErr   bool
		errTarget error
	}{
		{
			name: "rejects inverted period",
			req: CreateRunRequest{
				OrganizationID: 1,
				PeriodStart:    time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
				PeriodEnd:      time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC),
			},
			svc:       newPayrollTestService(PayrollRunDAOMock{}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			wantErr:   true,
			errTarget: ErrRunPeriod,
		},
		{
			name: "rejects sequence error",
			req: CreateRunRequest{
				OrganizationID: 1,
				PeriodStart:    time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC),
				PeriodEnd:      time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
			},
			svc: NewPayrollService(PayrollRunDAOMock{}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, AttendanceDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, sequence.NewSequenceService(sequence.DAOMock{
				ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
					return nil, errors.New("no sequence")
				},
			}), inventory.TransactionerMock{}),
			wantErr:   true,
			errTarget: ErrRunSequence,
		},
		{
			name: "rejects employee list error",
			req: CreateRunRequest{
				OrganizationID: 1,
				PeriodStart:    time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC),
				PeriodEnd:      time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
			},
			svc: newPayrollTestService(PayrollRunDAOMock{}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{
				CRUDMock: dao.CRUDMock[Employee]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Employee], error) {
						return nil, errors.New("db down")
					},
				},
			}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			wantErr: true,
		},
		{
			name: "rejects contract error",
			req: func() CreateRunRequest {
				id := uint64(1)
				return CreateRunRequest{
					OrganizationID: id,
					PeriodStart:    time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC),
					PeriodEnd:      time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
				}
			}(),
			svc: func() PayrollService {
				id := uint64(1)
				return newPayrollTestService(PayrollRunDAOMock{}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Employee], error) {
							return &query.Page[Employee]{Items: []*Employee{{Base: model.Base{ID: 1}, Active: true, OrganizationID: &id}}}, nil
						},
					},
				}, EmploymentContractDAOMock{
					FindActiveByEmployeeFunc: func(_ context.Context, _ uint64) (*EmploymentContract, error) {
						return nil, errors.New("db down")
					},
				}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{})
			}(),
			wantErr: true,
		},
		{
			name: "rejects rule list error",
			req: func() CreateRunRequest {
				id := uint64(1)
				return CreateRunRequest{
					OrganizationID: id,
					PeriodStart:    time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC),
					PeriodEnd:      time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
				}
			}(),
			svc: func() PayrollService {
				id := uint64(1)
				return newPayrollTestService(PayrollRunDAOMock{}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Employee], error) {
							return &query.Page[Employee]{Items: []*Employee{{Base: model.Base{ID: 1}, Active: true, OrganizationID: &id}}}, nil
						},
					},
				}, EmploymentContractDAOMock{
					FindActiveByEmployeeFunc: func(_ context.Context, _ uint64) (*EmploymentContract, error) {
						return &EmploymentContract{Base: model.Base{ID: 1}}, nil
					},
				}, dao.CRUDMock[reference.SalaryRule]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalaryRule], error) {
						return nil, errors.New("db down")
					},
				}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{})
			}(),
			wantErr: true,
		},
		{
			name: "rejects tx error",
			req: func() CreateRunRequest {
				id := uint64(1)
				return CreateRunRequest{
					OrganizationID: id,
					PeriodStart:    time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC),
					PeriodEnd:      time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
				}
			}(),
			svc: func() PayrollService {
				id := uint64(1)
				return newPayrollTestService(PayrollRunDAOMock{}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Employee], error) {
							return &query.Page[Employee]{Items: []*Employee{{Base: model.Base{ID: 1}, Active: true, OrganizationID: &id}}}, nil
						},
					},
				}, EmploymentContractDAOMock{
					FindActiveByEmployeeFunc: func(_ context.Context, _ uint64) (*EmploymentContract, error) {
						return &EmploymentContract{Base: model.Base{ID: 1}, Wage: 5000000}, nil
					},
				}, dao.CRUDMock[reference.SalaryRule]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalaryRule], error) {
						return &query.Page[reference.SalaryRule]{Items: []*reference.SalaryRule{
							{Base: model.Base{ID: 1}, Code: "BAS", Category: helper.Ptr("earning"), ComputeType: helper.Ptr("fixed"), Amount: helper.Ptr(5000000.0)},
						}}, nil
					},
				}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{
					RunFunc: func(_ context.Context, _ func(tx *gorm.DB) error) error {
						return errors.New("tx failed")
					},
				})
			}(),
			wantErr: true,
		},
		{
			name: "skips inactive employee",
			req: func() CreateRunRequest {
				id := uint64(1)
				return CreateRunRequest{
					OrganizationID: id,
					PeriodStart:    time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC),
					PeriodEnd:      time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
				}
			}(),
			svc: func() PayrollService {
				id := uint64(1)
				otherOrg := uint64(2)
				return newPayrollTestService(PayrollRunDAOMock{}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{
					CRUDMock: dao.CRUDMock[Employee]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Employee], error) {
							return &query.Page[Employee]{Items: []*Employee{
								{Base: model.Base{ID: 1}, Active: false, OrganizationID: &id},
								{Base: model.Base{ID: 2}, Active: true, OrganizationID: &otherOrg},
								{Base: model.Base{ID: 3}, Active: true, OrganizationID: &id},
							}}, nil
						},
					},
				}, EmploymentContractDAOMock{
					FindActiveByEmployeeFunc: func(_ context.Context, _ uint64) (*EmploymentContract, error) {
						return nil, nil
					},
				}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{})
			}(),
			wantErr:   true,
			errTarget: ErrRunNoPayslips,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.svc.CreateRun(ctx, tt.req)
			helper.AssertError(t, err, tt.wantErr, tt.errTarget)
		})
	}
}

func TestPayrollService_Confirm(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		svc       PayrollService
		wantErr   bool
		errTarget error
	}{
		{
			name: "rejects find error",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return nil, errors.New("db down")
					},
				},
			}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			wantErr: true,
		},
		{
			name: "rejects missing run",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return nil, nil
					},
				},
			}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			wantErr:   true,
			errTarget: ErrRunNotFound,
		},
		{
			name: "rejects no payslips",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateDraft}, nil
					},
				},
			}, PayslipDAOMock{
				ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
					return []*Payslip{}, nil
				},
			}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			wantErr:   true,
			errTarget: ErrRunNoPayslips,
		},
		{
			name: "rejects payslip list error",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateDraft}, nil
					},
				},
			}, PayslipDAOMock{
				ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
					return nil, errors.New("db down")
				},
			}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			wantErr: true,
		},
		{
			name: "rejects line error",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateDraft}, nil
					},
				},
			}, PayslipDAOMock{
				ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
					return []*Payslip{{Base: model.Base{ID: 1}}}, nil
				},
			}, PayslipLineDAOMock{
				ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*PayslipLine, error) {
					return nil, errors.New("db down")
				},
			}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			wantErr: true,
		},
		{
			name: "rejects missing rule",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateDraft}, nil
					},
				},
			}, PayslipDAOMock{
				ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
					return []*Payslip{{Base: model.Base{ID: 1}}}, nil
				},
			}, PayslipLineDAOMock{
				ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*PayslipLine, error) {
					return []*PayslipLine{{RuleID: 5}}, nil
				},
			}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
					return nil, nil
				},
			}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			wantErr:   true,
			errTarget: ErrRuleNotFound,
		},
		{
			name: "rejects missing rule accounts",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateDraft}, nil
					},
				},
			}, PayslipDAOMock{
				ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
					return []*Payslip{{Base: model.Base{ID: 1}}}, nil
				},
			}, PayslipLineDAOMock{
				ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*PayslipLine, error) {
					return []*PayslipLine{{RuleID: 5}}, nil
				},
			}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
					return &reference.SalaryRule{Base: model.Base{ID: 5}}, nil
				},
			}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			wantErr:   true,
			errTarget: ErrRuleAccounts,
		},
		{
			name: "rejects post error",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateDraft}, nil
					},
				},
			}, PayslipDAOMock{
				ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
					return []*Payslip{{Base: model.Base{ID: 1}}}, nil
				},
			}, PayslipLineDAOMock{
				ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*PayslipLine, error) {
					return []*PayslipLine{}, nil
				},
			}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{
				PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return nil, errors.New("post failed")
				},
			}, inventory.TransactionerMock{}),
			wantErr: true,
		},
		{
			name: "rejects run update error",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateDraft}, nil
					},
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *PayrollRun) (*PayrollRun, error) {
					return nil, errors.New("update failed")
				},
			}, PayslipDAOMock{
				ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
					return []*Payslip{{Base: model.Base{ID: 1}}}, nil
				},
			}, PayslipLineDAOMock{
				ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*PayslipLine, error) {
					return []*PayslipLine{}, nil
				},
			}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{
				PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return &accounting.JournalEntry{Base: model.Base{ID: 5}}, nil
				},
			}, inventory.TransactionerMock{}),
			wantErr: true,
		},
		{
			name: "rejects payslip update error",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateDraft}, nil
					},
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, run *PayrollRun) (*PayrollRun, error) {
					return run, nil
				},
			}, PayslipDAOMock{
				ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
					return []*Payslip{{Base: model.Base{ID: 1}}, {Base: model.Base{ID: 2}}}, nil
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *Payslip) (*Payslip, error) {
					return nil, errors.New("update failed")
				},
			}, PayslipLineDAOMock{
				ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*PayslipLine, error) {
					return []*PayslipLine{}, nil
				},
			}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{
				PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return &accounting.JournalEntry{Base: model.Base{ID: 5}}, nil
				},
			}, inventory.TransactionerMock{}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.svc.Confirm(ctx, 1, 9, time.Now())
			helper.AssertError(t, err, tt.wantErr, tt.errTarget)
		})
	}
}

func TestPayrollService_Pay(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		svc               PayrollService
		netPayableAccount uint64
		wantErr           bool
		errTarget         error
	}{
		{
			name: "rejects find error",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return nil, errors.New("db down")
					},
				},
			}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			netPayableAccount: 2200,
			wantErr:           true,
		},
		{
			name: "rejects missing run",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return nil, nil
					},
				},
			}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			netPayableAccount: 2200,
			wantErr:           true,
			errTarget:         ErrRunNotFound,
		},
		{
			name: "rejects wrong state",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateDraft}, nil
					},
				},
			}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			netPayableAccount: 2200,
			wantErr:           true,
			errTarget:         ErrRunState,
		},
		{
			name: "rejects missing journal",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateConfirmed}, nil
					},
				},
			}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
					return nil, nil
				},
			}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			netPayableAccount: 2200,
			wantErr:           true,
			errTarget:         ErrRunNotFound,
		},
		{
			name:              "rejects missing net payable account",
			svc:               newPayrollTestService(PayrollRunDAOMock{}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			netPayableAccount: 0,
			wantErr:           true,
			errTarget:         ErrNoNetPayable,
		},
		{
			name: "rejects no net payable amount",
			svc: func() PayrollService {
				bankAccount := uint64(1110)
				return newPayrollTestService(PayrollRunDAOMock{
					CRUDMock: dao.CRUDMock[PayrollRun]{
						FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
							return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateConfirmed}, nil
						},
					},
				}, PayslipDAOMock{
					ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
						return []*Payslip{{Base: model.Base{ID: 1}, Net: 0}}, nil
					},
				}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
						return &reference.Journal{Base: model.Base{ID: 9}, DefaultAccountID: &bankAccount}, nil
					},
				}, inventory.PosterMock{}, inventory.TransactionerMock{})
			}(),
			netPayableAccount: 2200,
			wantErr:           true,
			errTarget:         ErrNoNetPayable,
		},
		{
			name: "rejects rules error",
			svc: func() PayrollService {
				bankAccount := uint64(1110)
				return newPayrollTestService(PayrollRunDAOMock{
					CRUDMock: dao.CRUDMock[PayrollRun]{
						FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
							return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateConfirmed}, nil
						},
					},
				}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalaryRule], error) {
						return nil, errors.New("db down")
					},
				}, dao.CRUDMock[reference.Journal]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
						return &reference.Journal{Base: model.Base{ID: 9}, DefaultAccountID: &bankAccount}, nil
					},
				}, inventory.PosterMock{}, inventory.TransactionerMock{})
			}(),
			netPayableAccount: 2200,
			wantErr:           true,
		},
		{
			name: "rejects payslip error",
			svc: func() PayrollService {
				bankAccount := uint64(1110)
				netPayable := uint64(2200)
				return newPayrollTestService(PayrollRunDAOMock{
					CRUDMock: dao.CRUDMock[PayrollRun]{
						FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
							return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateConfirmed}, nil
						},
					},
				}, PayslipDAOMock{
					ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
						return nil, errors.New("db down")
					},
				}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalaryRule], error) {
						return &query.Page[reference.SalaryRule]{Items: []*reference.SalaryRule{
							{Base: model.Base{ID: 1}, Category: helper.Ptr("earning"), AccountCreditID: &netPayable},
						}}, nil
					},
				}, dao.CRUDMock[reference.Journal]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
						return &reference.Journal{Base: model.Base{ID: 9}, DefaultAccountID: &bankAccount}, nil
					},
				}, inventory.PosterMock{}, inventory.TransactionerMock{})
			}(),
			netPayableAccount: 2200,
			wantErr:           true,
		},
		{
			name: "rejects post error",
			svc: func() PayrollService {
				bankAccount := uint64(1110)
				netPayable := uint64(2200)
				return newPayrollTestService(PayrollRunDAOMock{
					CRUDMock: dao.CRUDMock[PayrollRun]{
						FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
							return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateConfirmed}, nil
						},
					},
				}, PayslipDAOMock{
					ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
						return []*Payslip{{Base: model.Base{ID: 1}, Net: 1000000}}, nil
					},
				}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalaryRule], error) {
						return &query.Page[reference.SalaryRule]{Items: []*reference.SalaryRule{
							{Base: model.Base{ID: 1}, Category: helper.Ptr("earning"), AccountCreditID: &netPayable},
						}}, nil
					},
				}, dao.CRUDMock[reference.Journal]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
						return &reference.Journal{Base: model.Base{ID: 9}, DefaultAccountID: &bankAccount}, nil
					},
				}, inventory.PosterMock{
					PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
						return nil, errors.New("post failed")
					},
				}, inventory.TransactionerMock{})
			}(),
			netPayableAccount: 2200,
			wantErr:           true,
		},
		{
			name: "rejects run update error",
			svc: func() PayrollService {
				bankAccount := uint64(1110)
				netPayable := uint64(2200)
				return newPayrollTestService(PayrollRunDAOMock{
					CRUDMock: dao.CRUDMock[PayrollRun]{
						FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
							return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateConfirmed}, nil
						},
					},
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *PayrollRun) (*PayrollRun, error) {
						return nil, errors.New("update failed")
					},
				}, PayslipDAOMock{
					ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
						return []*Payslip{{Base: model.Base{ID: 1}, Net: 1000000}}, nil
					},
				}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalaryRule], error) {
						return &query.Page[reference.SalaryRule]{Items: []*reference.SalaryRule{
							{Base: model.Base{ID: 1}, Category: helper.Ptr("earning"), AccountCreditID: &netPayable},
						}}, nil
					},
				}, dao.CRUDMock[reference.Journal]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
						return &reference.Journal{Base: model.Base{ID: 9}, DefaultAccountID: &bankAccount}, nil
					},
				}, inventory.PosterMock{
					PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
						return &accounting.JournalEntry{Base: model.Base{ID: 6}}, nil
					},
				}, inventory.TransactionerMock{})
			}(),
			netPayableAccount: 2200,
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.svc.Pay(ctx, 1, 9, tt.netPayableAccount, time.Now())
			helper.AssertError(t, err, tt.wantErr, tt.errTarget)
		})
	}
}

func TestPayrollService_Close_RejectsMissingRunOrWrongState(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		svc       PayrollService
		wantErr   bool
		errTarget error
	}{
		{
			name: "rejects missing run",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return nil, nil
					},
				},
			}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			wantErr:   true,
			errTarget: ErrRunNotFound,
		},
		{
			name: "rejects wrong state",
			svc: newPayrollTestService(PayrollRunDAOMock{
				CRUDMock: dao.CRUDMock[PayrollRun]{
					FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
						return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateConfirmed}, nil
					},
				},
			}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, inventory.TransactionerMock{}),
			wantErr:   true,
			errTarget: ErrRunState,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.svc.Close(ctx, 1)
			helper.AssertError(t, err, tt.wantErr, tt.errTarget)
		})
	}
}
