//go:build integration

package integration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/expense"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

type expenseFixture struct {
	orgID       uint64
	employeeID  uint64
	approverID  uint64
	expenseAcc  uint64
	payableAcc  uint64
	clearingAcc uint64
	bankAcc     uint64
	taxAcc      uint64
	journalID   uint64
	taxID       uint64
	categoryID  uint64

	expenseSvc      expense.ExpenseService
	reportDAO       expense.ExpenseReportDAO
	journalEntryDAO accounting.JournalEntryDAO
	moveLineDAO     accounting.JournalLineDAO
	systemConfigDAO dao.Base[reference.SystemConfig]
}

func TestExpenseReportOwnAccountLifecycleIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedExpenseFixture(t)
	ctx := testutil.SystemContext()

	report, err := fx.expenseSvc.Create(ctx, expense.CreateExpenseReportRequest{
		OrganizationID: fx.orgID,
		Name:           "Travel June",
		EmployeeID:     fx.employeeID,
		PaymentMode:    expense.ExpensePaymentOwnAccount,
		Lines: []expense.CreateExpenseLineRequest{
			{
				CategoryID:   helper.Ptr(fx.categoryID),
				Description:  "Taxi",
				ExpenseDate:  time.Now().UTC(),
				Quantity:     1,
				UnitPrice:    100,
				CurrencyCode: "USD",
				TaxIDs:       helper.Int64Array{int64(fx.taxID)},
			},
		},
	})
	if err != nil {
		t.Fatalf("create report failed: %v", err)
	}
	if report.State != expense.ExpenseStateDraft || report.TotalAmount != 100 {
		t.Errorf("report = %+v, want draft total 100", report)
	}

	submitted, err := fx.expenseSvc.Submit(ctx, fx.orgID, report.ID)
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	if submitted.State != expense.ExpenseStateSubmitted {
		t.Errorf("state = %q, want submitted", submitted.State)
	}

	approved, err := fx.expenseSvc.Approve(ctx, fx.orgID, report.ID, fx.approverID)
	if err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	if approved.State != expense.ExpenseStateApproved || approved.ApprovedBy == nil || *approved.ApprovedBy != fx.approverID {
		t.Errorf("approved = %+v, want approved by %d", approved, fx.approverID)
	}

	posted, err := fx.expenseSvc.Post(ctx, fx.orgID, report.ID)
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	if posted.State != expense.ExpenseStatePosted || posted.MovementID == nil {
		t.Fatalf("posted = %+v, want posted with movement", posted)
	}

	postLines, err := fx.moveLineDAO.ListByMovement(ctx, *posted.MovementID)
	if err != nil {
		t.Fatalf("list post lines failed: %v", err)
	}
	if len(postLines) != 3 {
		t.Fatalf("expected 3 posting lines, got %d", len(postLines))
	}
	assertJournalLine(t, postLines, fx.expenseAcc, 100, 0)
	assertJournalLine(t, postLines, fx.taxAcc, 10, 0)
	assertJournalLine(t, postLines, fx.payableAcc, 0, 110)

	reimbursed, err := fx.expenseSvc.Reimburse(ctx, fx.orgID, report.ID)
	if err != nil {
		t.Fatalf("reimburse failed: %v", err)
	}
	if reimbursed.State != expense.ExpenseStateReimbursed || reimbursed.ReimbursementEntryID == nil {
		t.Fatalf("reimbursed = %+v, want reimbursed with movement", reimbursed)
	}

	reimburseLines, err := fx.moveLineDAO.ListByMovement(ctx, *reimbursed.ReimbursementEntryID)
	if err != nil {
		t.Fatalf("list reimburse lines failed: %v", err)
	}
	if len(reimburseLines) != 2 {
		t.Fatalf("expected 2 reimbursement lines, got %d", len(reimburseLines))
	}
	assertJournalLine(t, reimburseLines, fx.payableAcc, 110, 0)
	assertJournalLine(t, reimburseLines, fx.bankAcc, 0, 110)
}

func TestExpenseReportOrgAccountPostsToClearingIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedExpenseFixture(t)
	ctx := testutil.SystemContext()

	report, err := fx.expenseSvc.Create(ctx, expense.CreateExpenseReportRequest{
		OrganizationID: fx.orgID,
		Name:           "Travel July",
		EmployeeID:     fx.employeeID,
		PaymentMode:    expense.ExpensePaymentOrgAccount,
		Lines: []expense.CreateExpenseLineRequest{
			{CategoryID: helper.Ptr(fx.categoryID), Description: "Hotel", ExpenseDate: time.Now().UTC(), Quantity: 1, UnitPrice: 80, CurrencyCode: "USD"},
		},
	})
	if err != nil {
		t.Fatalf("create report failed: %v", err)
	}
	if _, err := fx.expenseSvc.Submit(ctx, fx.orgID, report.ID); err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	if _, err := fx.expenseSvc.Approve(ctx, fx.orgID, report.ID, fx.approverID); err != nil {
		t.Fatalf("approve failed: %v", err)
	}
	posted, err := fx.expenseSvc.Post(ctx, fx.orgID, report.ID)
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	if posted.State != expense.ExpenseStatePosted {
		t.Fatalf("posted = %+v, want posted", posted)
	}

	postLines, err := fx.moveLineDAO.ListByMovement(ctx, *posted.MovementID)
	if err != nil {
		t.Fatalf("list post lines failed: %v", err)
	}
	assertJournalLine(t, postLines, fx.expenseAcc, 80, 0)
	assertJournalLine(t, postLines, fx.clearingAcc, 0, 80)
}

func seedExpenseFixture(t *testing.T) expenseFixture {
	t.Helper()
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}

	userSvc := iam.NewUserService(iam.NewUserDAO(testDB), iam.NewPasswordService(testCfg))
	user, err := userSvc.Create(ctx, &iam.User{
		Username: gofakeit.Username(), FirstName: gofakeit.FirstName(), Email: gofakeit.Email(),
		Password: "test-password",
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	if _, err := iam.NewMemberDAO(testDB).Create(ctx, &iam.Member{
		UserID: user.ID, OrganizationID: org.ID,
	}); err != nil {
		t.Fatalf("create member failed: %v", err)
	}

	contact, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: helper.Ptr(org.ID), Name: gofakeit.Name(),
	})
	if err != nil {
		t.Fatalf("create contact failed: %v", err)
	}
	employee, err := payroll.NewEmployeeDAO(testDB).Create(ctx, &payroll.Employee{
		OrganizationID: helper.Ptr(org.ID), ContactID: contact.ID, UserID: helper.Ptr(user.ID),
		EmployeeNumber: gofakeit.UUID(),
		Active:         true,
	})
	if err != nil {
		t.Fatalf("create employee failed: %v", err)
	}

	expenseAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "6200", Name: "Travel Expense", Type: "expense", Active: true,
	})
	if err != nil {
		t.Fatalf("create expense account failed: %v", err)
	}
	payableAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "2100", Name: "Employee Payable", Type: "liability", Active: true,
	})
	if err != nil {
		t.Fatalf("create payable account failed: %v", err)
	}
	clearingAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1200", Name: "Card Clearing", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create clearing account failed: %v", err)
	}
	bankAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1001", Name: "Bank", Type: "bank", Active: true,
	})
	if err != nil {
		t.Fatalf("create bank account failed: %v", err)
	}
	taxOut, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "2200", Name: "Input Tax", Type: "tax", Active: true,
	})
	if err != nil {
		t.Fatalf("create tax account failed: %v", err)
	}

	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: org.ID, Name: "Expense Journal", Code: helper.Ptr("EXP"),
		Type: "general", DefaultAccountID: helper.Ptr(expenseAcc.ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}

	tax, err := dao.NewBase[reference.Tax](testDB).Create(ctx, &reference.Tax{
		OrganizationID: helper.Ptr(org.ID), Name: "Input VAT", Amount: helper.Ptr(10.0),
		Type: reference.TaxTypePercent, Scope: reference.TaxScopePurchase, TaxAccountID: helper.Ptr(taxOut.ID), Active: true,
	})
	if err != nil {
		t.Fatalf("create tax failed: %v", err)
	}

	category, err := dao.NewBase[reference.ExpenseCategory](testDB).Create(ctx, &reference.ExpenseCategory{
		OrganizationID:   helper.Ptr(org.ID),
		Name:             "Travel",
		ExpenseAccountID: helper.Ptr(expenseAcc.ID),
	})
	if err != nil {
		t.Fatalf("create expense category failed: %v", err)
	}

	systemConfigDAO := dao.NewBase[reference.SystemConfig](testDB)
	configs := map[string]uint64{
		"expense.journal_id":                    journal.ID,
		"expense.employee_payable_account_id":   payableAcc.ID,
		"expense.card_clearing_account_id":      clearingAcc.ID,
		"expense.reimbursement_bank_account_id": bankAcc.ID,
	}
	for key, value := range configs {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal config failed: %v", err)
		}
		if _, err := systemConfigDAO.Create(ctx, &reference.SystemConfig{
			OrganizationID: helper.Ptr(org.ID), Key: key, Value: raw,
		}); err != nil {
			t.Fatalf("create system config %q failed: %v", key, err)
		}
	}

	journalEntryDAO := accounting.NewJournalEntryDAO(testDB)
	journalEntryLineDAO := accounting.NewJournalLineDAO(testDB)
	taxPeriodSvc := accounting.NewTaxPeriodService(accounting.NewTaxPeriodDAO(testDB), dao.NewBase[reference.TaxYear](testDB))
	reversalEngine := accounting.NewReversalEngine(journalEntryDAO, journalEntryLineDAO)
	postingSvc := accounting.NewPostingService(journalEntryDAO).SetPeriods(taxPeriodSvc).SetReverser(reversalEngine)

	expenseSvc := expense.NewExpenseService(
		expense.NewExpenseReportDAO(testDB),
		expense.NewExpenseLineDAO(testDB),
		dao.NewBase[reference.ExpenseCategory](testDB),
		dao.NewBase[reference.Tax](testDB),
		postingSvc,
		expense.NewExpenseConfigSource(systemConfigDAO),
		db.NewDBTransactioner(testDB),
	)

	return expenseFixture{
		orgID:           org.ID,
		employeeID:      employee.ID,
		approverID:      user.ID,
		expenseAcc:      expenseAcc.ID,
		payableAcc:      payableAcc.ID,
		clearingAcc:     clearingAcc.ID,
		bankAcc:         bankAcc.ID,
		taxAcc:          taxOut.ID,
		journalID:       journal.ID,
		taxID:           tax.ID,
		categoryID:      category.ID,
		expenseSvc:      expenseSvc,
		reportDAO:       expense.NewExpenseReportDAO(testDB),
		journalEntryDAO: journalEntryDAO,
		moveLineDAO:     journalEntryLineDAO,
		systemConfigDAO: systemConfigDAO,
	}
}

func assertJournalLine(t *testing.T, lines []*accounting.JournalLine, accountID uint64, wantDebit, wantCredit float64) {
	t.Helper()
	for _, line := range lines {
		if line.AccountID == accountID {
			if line.Debit.Float64() != wantDebit || line.Credit.Float64() != wantCredit {
				t.Errorf("line for account %d = debit %v credit %v, want debit %v credit %v", accountID, line.Debit, line.Credit, wantDebit, wantCredit)
			}
			return
		}
	}
	t.Errorf("no movement line found for account %d in %+v", accountID, lines)
}
