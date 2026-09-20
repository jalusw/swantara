//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
	"github.com/jalusw/swantara/apps/service/test/testutil"
	"gorm.io/gorm"
)

const (
	payrollBatchTarget     = 5 * time.Minute
	payrollBatchPayslips   = 5000
	payrollLinesPerSheet   = 5
	deferralBatchTarget    = 10 * time.Minute
	deferralBatchSchedules = 300
	deferralLinesPerSheet  = 10
)

func TestPayrollRunBatchScale(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}

	run := &payroll.PayrollRun{
		OrganizationID: org.ID,
		Name:           helper.Ptr("Monthly payroll"),
		PeriodStart:    time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:      time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
		State:          payroll.RunStateDraft,
	}
	people := make([]*contacts.Contact, 0, payrollBatchPayslips)
	for i := 0; i < payrollBatchPayslips; i++ {
		people = append(people, &contacts.Contact{
			OrganizationID: helper.Ptr(org.ID),
			Name:           fmt.Sprintf("Employee %d", i+1),
			Lang:           "en_US",
			Active:         true,
		})
	}
	if err := testDB.WithContext(ctx).CreateInBatches(people, 500).Error; err != nil {
		t.Fatalf("create contacts failed: %v", err)
	}
	employees := make([]*payroll.Employee, 0, payrollBatchPayslips)
	for i := 0; i < payrollBatchPayslips; i++ {
		employees = append(employees, &payroll.Employee{
			OrganizationID: helper.Ptr(org.ID),
			ContactID:      people[i].ID,
			EmployeeNumber: fmt.Sprintf("EMP%05d", i+1),
			EmploymentType: payroll.EmploymentTypeFullTime,
			Active:         true,
		})
	}
	if err := testDB.WithContext(ctx).CreateInBatches(employees, 500).Error; err != nil {
		t.Fatalf("create employees failed: %v", err)
	}
	contracts := make([]*payroll.EmploymentContract, 0, payrollBatchPayslips)
	for _, employee := range employees {
		contracts = append(contracts, &payroll.EmploymentContract{
			EmployeeID:   employee.ID,
			DateStart:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Wage:         5000000,
			WageType:     payroll.WageTypeMonthly,
			CurrencyCode: "IDR",
			State:        payroll.ContractStateActive,
		})
	}
	if err := testDB.WithContext(ctx).CreateInBatches(contracts, 500).Error; err != nil {
		t.Fatalf("create contracts failed: %v", err)
	}
	rules := make([]*reference.SalaryRule, 0, payrollLinesPerSheet)
	for j := 0; j < payrollLinesPerSheet; j++ {
		rules = append(rules, &reference.SalaryRule{
			Code: fmt.Sprintf("RULE%d", j),
			Name: "Rule",
		})
	}
	if err := testDB.WithContext(ctx).CreateInBatches(rules, 100).Error; err != nil {
		t.Fatalf("create salary rules failed: %v", err)
	}
	payslips := make([]*payroll.Payslip, 0, payrollBatchPayslips)
	lines := make([][]*payroll.PayslipLine, 0, payrollBatchPayslips)
	for i := 0; i < payrollBatchPayslips; i++ {
		payslips = append(payslips, &payroll.Payslip{
			EmployeeID: employees[i].ID,
			ContractID: contracts[i].ID,
			Gross:      5000000,
			Net:        4200000,
			State:      payroll.PayslipStateDraft,
		})
		payslipLines := make([]*payroll.PayslipLine, 0, payrollLinesPerSheet)
		for j := 0; j < payrollLinesPerSheet; j++ {
			payslipLines = append(payslipLines, &payroll.PayslipLine{
				RuleID:   rules[j].ID,
				Code:     fmt.Sprintf("CODE%d", j),
				Name:     "Line",
				Category: payroll.RuleCategoryEarning,
				Amount:   1000000,
			})
		}
		lines = append(lines, payslipLines)
	}

	start := time.Now()
	err = testDB.Transaction(func(tx *gorm.DB) error {
		_, err := payroll.NewPayrollRunDAO(testDB).CreateWithPayslipsTx(ctx, tx, run, payslips, lines)
		return err
	})
	duration := time.Since(start)
	if err != nil {
		t.Fatalf("batch write failed: %v", err)
	}
	if run.ID == 0 {
		t.Fatal("run produced no id")
	}
	if duration > payrollBatchTarget {
		t.Errorf("payroll batch = %v, want < %v (NFR-PER-004)", duration, payrollBatchTarget)
	}

	stored, err := payroll.NewPayslipDAO(testDB).ListByRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("list payslips failed: %v", err)
	}
	if len(stored) != payrollBatchPayslips {
		t.Errorf("payslips = %d, want %d", len(stored), payrollBatchPayslips)
	}
	totalLines := 0
	for _, payslip := range stored {
		sheetLines, err := payroll.NewPayslipLineDAO(testDB).ListByPayslip(ctx, payslip.ID)
		if err != nil {
			t.Fatalf("list lines failed: %v", err)
		}
		totalLines += len(sheetLines)
	}
	if totalLines != payrollBatchPayslips*payrollLinesPerSheet {
		t.Errorf("payslip lines = %d, want %d", totalLines, payrollBatchPayslips*payrollLinesPerSheet)
	}
	t.Logf("payroll batch %d payslips + %d lines took %v", payrollBatchPayslips, totalLines, duration)
}

func TestDeferralRecognitionBatchScale(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}
	bsAccount, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "2400", Name: "Deferred Revenue", Type: "liability", Active: true,
	})
	if err != nil {
		t.Fatalf("create balance sheet account failed: %v", err)
	}
	plAccount, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "4100", Name: "Revenue", Type: "income", Active: true,
	})
	if err != nil {
		t.Fatalf("create P&L account failed: %v", err)
	}
	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: org.ID, Name: "Batch Journal", Code: helper.Ptr("BATCH"),
		Type: "general", DefaultAccountID: helper.Ptr(bsAccount.ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}
	configValue, err := json.Marshal(journal.ID)
	if err != nil {
		t.Fatalf("marshal config failed: %v", err)
	}
	if _, err := dao.NewBase[reference.SystemConfig](testDB).Create(ctx, &reference.SystemConfig{
		OrganizationID: helper.Ptr(org.ID), Key: "subscription.journal_id", Value: configValue,
	}); err != nil {
		t.Fatalf("create config failed: %v", err)
	}

	scheduleDAO := accounting.NewDeferredScheduleDAO(testDB)
	lineDAO := accounting.NewDeferredScheduleLineDAO(testDB)
	asOf := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	totalDue := 0
	err = testDB.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < deferralBatchSchedules; i++ {
			schedule, err := scheduleDAO.CreateTx(ctx, tx, &accounting.DeferredSchedule{
				OrganizationID:        helper.Ptr(org.ID),
				Type:                  accounting.DeferredTypeDeferredRevenue,
				SourceType:            "load_test",
				SourceID:              uint64(i + 1),
				TotalAmount:           deferralLinesPerSheet * 100,
				BalanceSheetAccountID: helper.Ptr(bsAccount.ID),
				PLAccountID:           helper.Ptr(plAccount.ID),
				Method:                accounting.DeferredMethodLinear,
				DateStart:             helper.Ptr(asOf.AddDate(0, -1, 0)),
				DateEnd:               helper.Ptr(asOf.AddDate(1, 0, 0)),
				Periods:               12,
				State:                 accounting.DeferredStateRunning,
			})
			if err != nil {
				return err
			}
			for j := 0; j < deferralLinesPerSheet; j++ {
				if _, err := lineDAO.CreateTx(ctx, tx, &accounting.DeferredScheduleLine{
					ScheduleID:      schedule.ID,
					Sequence:        j + 1,
					RecognitionDate: helper.Ptr(asOf.AddDate(0, -j, 0)),
					Amount:          100,
				}); err != nil {
					return err
				}
				totalDue++
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed schedules failed: %v", err)
	}

	deferralSvc := accounting.NewDeferralService(
		scheduleDAO,
		lineDAO,
		accounting.NewPostingService(accounting.NewJournalEntryDAO(testDB)),
		subscription.NewSubscriptionConfigSource(dao.NewBase[reference.SystemConfig](testDB)),
		db.NewDBTransactioner(testDB),
	)

	start := time.Now()
	recognized, err := deferralSvc.RecognizeDue(ctx, nil, asOf)
	duration := time.Since(start)
	if err != nil {
		t.Fatalf("recognize due failed: %v", err)
	}
	if recognized != totalDue {
		t.Errorf("recognized = %d, want %d", recognized, totalDue)
	}
	if duration > deferralBatchTarget {
		t.Errorf("deferral batch = %v, want < %v (NFR-PER-005)", duration, deferralBatchTarget)
	}
	t.Logf("deferral recognition %d lines took %v", totalDue, duration)
}
