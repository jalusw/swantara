//go:build integration

package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

type integrityFixture struct {
	orgID         uint64
	journalID     uint64
	periodID      uint64
	date          time.Time
	actorID       uint64
	receivableID  uint64
	incomeID      uint64
	inventoryID   uint64
	variantID     uint64
	supplierLocID uint64
	internalLocID uint64
	customerLocID uint64
	bankID        uint64
	cogsID        uint64
	liabilityID   uint64
	otherIncomeID uint64

	postingSvc       accounting.PostingService
	taxPeriodSvc     accounting.TaxPeriodService
	moveDAO          accounting.JournalEntryDAO
	moveLineDAO      accounting.JournalLineDAO
	reconcileSvc     accounting.ReconcileService
	reportDAO        reporting.ReportDAO
	stockMovementDAO inventory.StockMovementDAO
	stockQuantDAO    inventory.StockBalanceDAO
	ledgerSvc        inventory.LedgerService
	valuationSvc     inventory.ValuationService
	userSvc          iam.UserService
	userDAO          iam.UserDAO
}

func seedIntegrityFixture(t *testing.T, costMethods ...string) integrityFixture {
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
		Username:  gofakeit.Username(),
		FirstName: gofakeit.FirstName(),
		Email:     gofakeit.Email(),
		Password:  "test-password",
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	if _, err := iam.NewMemberDAO(testDB).Create(ctx, &iam.Member{
		UserID: user.ID, OrganizationID: org.ID,
	}); err != nil {
		t.Fatalf("create member failed: %v", err)
	}

	accounts := map[string]*reference.Account{}
	for _, spec := range []struct {
		key  string
		code string
		name string
		typ  string
	}{
		{"receivable", "1200", "Accounts Receivable", "receivable"},
		{"income", "4000", "Sales Revenue", "income"},
		{"inventory", "1500", "Inventory", "asset"},
		{"stock_in", "1501", "Stock Input", "asset"},
		{"stock_out", "1502", "Stock Output", "asset"},
		{"cogs", "5000", "Cost of Goods Sold", "cogs"},
		{"bank", "1001", "Bank", "bank"},
		{"liability", "2400", "Gift Card Liability", "liability"},
		{"other_income", "4900", "Other Income", "income"},
	} {
		account, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
			OrganizationID: org.ID, Code: spec.code, Name: spec.name, Type: spec.typ, Active: true,
		})
		if err != nil {
			t.Fatalf("create account %s failed: %v", spec.key, err)
		}
		accounts[spec.key] = account
	}

	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID:   org.ID,
		Name:             "General Journal",
		Code:             helper.Ptr("GJ"),
		Type:             "general",
		DefaultAccountID: helper.Ptr(accounts["bank"].ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}

	date := time.Now().UTC()
	yearStart := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(date.Year(), 12, 31, 23, 59, 59, 0, time.UTC)
	year, err := dao.NewBase[reference.TaxYear](testDB).Create(ctx, &reference.TaxYear{
		OrganizationID: helper.Ptr(org.ID),
		Name:           "FY",
		DateStart:      helper.Ptr(yearStart),
		DateEnd:        helper.Ptr(yearEnd),
		State:          helper.Ptr("open"),
	})
	if err != nil {
		t.Fatalf("create tax year failed: %v", err)
	}

	moveDAO := accounting.NewJournalEntryDAO(testDB)
	moveLineDAO := accounting.NewJournalLineDAO(testDB)
	taxPeriodDAO := accounting.NewTaxPeriodDAO(testDB)
	taxPeriodSvc := accounting.NewTaxPeriodService(taxPeriodDAO, dao.NewBase[reference.TaxYear](testDB))
	period, err := taxPeriodSvc.Create(ctx, &accounting.TaxPeriod{
		OrganizationID: org.ID,
		TaxYearID:      year.ID,
		Name:           "Open",
		DateStart:      helper.Ptr(yearStart),
		DateEnd:        helper.Ptr(yearEnd),
		State:          accounting.TaxPeriodStateOpen,
	})
	if err != nil {
		t.Fatalf("create tax period failed: %v", err)
	}

	reversalEngine := accounting.NewReversalEngine(moveDAO, moveLineDAO)
	postingSvc := accounting.NewPostingService(moveDAO).SetPeriods(taxPeriodSvc).SetReverser(reversalEngine)

	costMethod := "fifo"
	if len(costMethods) > 0 {
		costMethod = costMethods[0]
	}
	category, err := dao.NewBase[reference.ItemCategory](testDB).Create(ctx, &reference.ItemCategory{
		Name:                    "General",
		CostMethod:              helper.Ptr(costMethod),
		IncomeAccountID:         helper.Ptr(accounts["income"].ID),
		StockValuationAccountID: helper.Ptr(accounts["inventory"].ID),
		StockInputAccountID:     helper.Ptr(accounts["stock_in"].ID),
		StockOutputAccountID:    helper.Ptr(accounts["stock_out"].ID),
		CogsAccountID:           helper.Ptr(accounts["cogs"].ID),
	})
	if err != nil {
		t.Fatalf("create item category failed: %v", err)
	}

	itemDAO := products.NewItemDAO(testDB)
	itemVariantDAO := products.NewItemVariantDAO(testDB)
	template, err := itemDAO.CreateWithVariants(ctx, &products.Item{
		OrganizationID: helper.Ptr(org.ID),
		Name:           "Integrity Item",
		CategoryID:     helper.Ptr(category.ID),
		Type:           "stockable",
		ListPrice:      100,
		StandardCost:   50,
		Tracking:       "none",
		IsSellable:     true,
		Active:         true,
	}, []*products.ItemVariant{{Active: true}})
	if err != nil {
		t.Fatalf("create item template failed: %v", err)
	}
	variants, err := itemVariantDAO.ListByTemplate(ctx, template.ID)
	if err != nil {
		t.Fatalf("list variants failed: %v", err)
	}

	warehouse, err := inventory.NewWarehouseDAO(testDB).Create(ctx, &reference.Warehouse{
		OrganizationID: helper.Ptr(org.ID), Name: "Main", Code: helper.Ptr("WH"),
	})
	if err != nil {
		t.Fatalf("create warehouse failed: %v", err)
	}
	stockLocationDAO := inventory.NewStockLocationDAO(testDB)
	internalLoc, err := stockLocationDAO.Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), WarehouseID: helper.Ptr(warehouse.ID), Name: "Stock", Code: helper.Ptr("STK"), Usage: "internal",
	})
	if err != nil {
		t.Fatalf("create internal location failed: %v", err)
	}
	supplierLoc, err := stockLocationDAO.Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), Name: "Suppliers", Usage: "supplier",
	})
	if err != nil {
		t.Fatalf("create supplier location failed: %v", err)
	}
	customerLoc, err := stockLocationDAO.Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), Name: "Customers", Usage: "customer",
	})
	if err != nil {
		t.Fatalf("create customer location failed: %v", err)
	}

	stockMovementDAO := inventory.NewStockMovementDAO(testDB)
	stockQuantDAO := inventory.NewStockBalanceDAO(testDB)
	productResolver := inventory.NewProductResolver(itemVariantDAO, itemDAO, dao.NewBase[reference.ItemCategory](testDB))
	valuationSvc := inventory.NewValuationService(
		stockMovementDAO,
		inventory.NewCostLayerDAO(testDB),
		stockLocationDAO,
		productResolver,
		postingSvc,
		db.NewDBTransactioner(testDB),
	)

	return integrityFixture{
		orgID:            org.ID,
		journalID:        journal.ID,
		periodID:         period.ID,
		date:             date,
		actorID:          user.ID,
		receivableID:     accounts["receivable"].ID,
		incomeID:         accounts["income"].ID,
		inventoryID:      accounts["inventory"].ID,
		variantID:        variants[0].ID,
		supplierLocID:    supplierLoc.ID,
		internalLocID:    internalLoc.ID,
		customerLocID:    customerLoc.ID,
		bankID:           accounts["bank"].ID,
		cogsID:           accounts["cogs"].ID,
		liabilityID:      accounts["liability"].ID,
		otherIncomeID:    accounts["other_income"].ID,
		postingSvc:       postingSvc,
		taxPeriodSvc:     taxPeriodSvc,
		moveDAO:          moveDAO,
		moveLineDAO:      moveLineDAO,
		reconcileSvc:     accounting.NewReconcileService(moveLineDAO, accounting.NewAccountPartialReconcileDAO(testDB), accounting.NewAccountFullReconcileDAO(testDB), moveDAO, db.NewDBTransactioner(testDB)),
		reportDAO:        reporting.NewReportDAO(testDB),
		stockMovementDAO: stockMovementDAO,
		stockQuantDAO:    stockQuantDAO,
		ledgerSvc:        inventory.NewLedgerService(stockMovementDAO, stockQuantDAO, stockLocationDAO, db.NewDBTransactioner(testDB)),
		valuationSvc:     valuationSvc,
		userSvc:          userSvc,
		userDAO:          iam.NewUserDAO(testDB),
	}
}

func glBalance(t *testing.T, orgID, accountID uint64) float64 {
	t.Helper()
	var balance float64
	err := testDB.WithContext(testutil.SystemContext()).Raw(`
		SELECT COALESCE(SUM(l.debit - l.credit), 0)
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.movement_id
		WHERE m.organization_id = ? AND m.state = 'posted'
		  AND l.account_id = ? AND m.deleted_at IS NULL AND l.deleted_at IS NULL`,
		orgID, accountID).Scan(&balance).Error
	if err != nil {
		t.Fatalf("compute gl balance failed: %v", err)
	}
	return balance
}

func TestPostingRejectsInvalidLinesAndUnbalancedMoves(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	base := accounting.PostRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		Description:    "DAT-001",
	}

	unbalanced := base
	unbalanced.Lines = []accounting.PostingLine{
		{AccountID: fx.receivableID, Name: "AR", Debit: amount.FromFloat64(100)},
		{AccountID: fx.incomeID, Name: "Income", Credit: amount.FromFloat64(50)},
	}
	if _, err := fx.postingSvc.Post(ctx, unbalanced); !errors.Is(err, accounting.ErrUnbalanced) {
		t.Fatalf("expected ErrUnbalanced, got %v", err)
	}

	negativeLine := base
	negativeLine.Lines = []accounting.PostingLine{
		{AccountID: fx.receivableID, Name: "AR", Debit: amount.FromFloat64(-5)},
		{AccountID: fx.incomeID, Name: "Income", Credit: amount.FromFloat64(-5)},
	}
	if _, err := fx.postingSvc.Post(ctx, negativeLine); !errors.Is(err, accounting.ErrInvalidLine) {
		t.Fatalf("expected ErrInvalidLine, got %v", err)
	}

	bothSides := base
	bothSides.Lines = []accounting.PostingLine{
		{AccountID: fx.receivableID, Name: "AR", Debit: amount.FromFloat64(100), Credit: amount.FromFloat64(100)},
	}
	if _, err := fx.postingSvc.Post(ctx, bothSides); !errors.Is(err, accounting.ErrInvalidLine) {
		t.Fatalf("expected ErrInvalidLine, got %v", err)
	}

	balanced := base
	balanced.Lines = []accounting.PostingLine{
		{AccountID: fx.receivableID, Name: "AR", Debit: amount.FromFloat64(100)},
		{AccountID: fx.incomeID, Name: "Income", Credit: amount.FromFloat64(100)},
	}
	posted, err := fx.postingSvc.Post(ctx, balanced)
	if err != nil {
		t.Fatalf("balanced post failed: %v", err)
	}
	if posted.State != accounting.MovementStatePosted {
		t.Errorf("expected posted state, got %q", posted.State)
	}

	if err := testDB.WithContext(ctx).Exec(`
		INSERT INTO journal_lines (movement_id, date, account_id, debit, credit)
		VALUES (?, ?, ?, -1, 0)`, posted.ID, fx.date, fx.receivableID).Error; err == nil {
		t.Fatal("expected negative-debit line to be rejected by database constraint")
	}
	if err := testDB.WithContext(ctx).Exec(`
		INSERT INTO journal_lines (movement_id, date, account_id, debit, credit)
		VALUES (?, ?, ?, 10, 10)`, posted.ID, fx.date, fx.receivableID).Error; err == nil {
		t.Fatal("expected double-sided line to be rejected by database constraint")
	}
}

func TestPostingRequiresOriginTypeWhenOriginIDPresent(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	lines := []accounting.PostingLine{
		{AccountID: fx.receivableID, Name: "AR", Debit: amount.FromFloat64(100)},
		{AccountID: fx.incomeID, Name: "Income", Credit: amount.FromFloat64(100)},
	}

	orphan := accounting.PostRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		Description:    "DAT-005 orphan origin",
		OriginID:       999,
		Lines:          lines,
	}
	if _, err := fx.postingSvc.Post(ctx, orphan); !errors.Is(err, accounting.ErrOriginIncomplete) {
		t.Fatalf("expected ErrOriginIncomplete for origin_id without origin_type, got %v", err)
	}

	traced := orphan
	traced.Description = "DAT-005 traced origin"
	traced.OriginType = "sale_order"
	posted, err := fx.postingSvc.Post(ctx, traced)
	if err != nil {
		t.Fatalf("traced origin post failed: %v", err)
	}
	if posted.OriginType == nil || *posted.OriginType != "sale_order" {
		t.Fatalf("expected origin_type sale_order, got %+v", posted.OriginType)
	}
	if posted.OriginID == nil || *posted.OriginID != 999 {
		t.Fatalf("expected origin_id 999, got %+v", posted.OriginID)
	}

	if err := testDB.WithContext(ctx).Exec(`
		INSERT INTO journal_entrys (organization_id, journal_id, name, date, state, origin_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		fx.orgID, fx.journalID, "DAT-005 direct", fx.date, "posted", 999).Error; err == nil {
		t.Fatal("expected orphan origin_id to be rejected by database constraint")
	}
}

func TestInventoryValuationReconcilesToControlAccount(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	receive, err := fx.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(fx.orgID),
		ItemID:         fx.variantID,
		Qty:            10,
		SrcLocationID:  fx.supplierLocID,
		DstLocationID:  fx.internalLocID,
		State:          inventory.MovementStateConfirmed,
		ScheduledDate:  helper.Ptr(fx.date),
	})
	if err != nil {
		t.Fatalf("create receive movement failed: %v", err)
	}
	if _, err := fx.valuationSvc.Receive(ctx, receive.ID, amount.FromFloat64(50), fx.journalID, fx.date); err != nil {
		t.Fatalf("receive stock failed: %v", err)
	}
	if balance := glBalance(t, fx.orgID, fx.inventoryID); balance != 500 {
		t.Fatalf("expected inventory GL balance 500 after receive, got %v", balance)
	}

	rows, err := fx.reportDAO.InventoryValuation(ctx, fx.orgID)
	if err != nil {
		t.Fatalf("inventory valuation failed: %v", err)
	}
	if len(rows) != 1 || rows[0].Value != 500 || rows[0].Quantity != 10 {
		t.Fatalf("expected valuation 10 qty / 500 value, got %+v", rows)
	}

	ship, err := fx.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(fx.orgID),
		ItemID:         fx.variantID,
		Qty:            4,
		SrcLocationID:  fx.internalLocID,
		DstLocationID:  fx.customerLocID,
		State:          inventory.MovementStateConfirmed,
		ScheduledDate:  helper.Ptr(fx.date),
	})
	if err != nil {
		t.Fatalf("create ship movement failed: %v", err)
	}
	if _, err := fx.valuationSvc.Ship(ctx, ship.ID, fx.journalID, fx.date); err != nil {
		t.Fatalf("ship stock failed: %v", err)
	}

	if balance := glBalance(t, fx.orgID, fx.inventoryID); balance != 300 {
		t.Fatalf("expected inventory GL balance 300 after ship, got %v", balance)
	}
	rows, err = fx.reportDAO.InventoryValuation(ctx, fx.orgID)
	if err != nil {
		t.Fatalf("inventory valuation after ship failed: %v", err)
	}
	if len(rows) != 1 || rows[0].Value != 300 || rows[0].Quantity != 6 {
		t.Fatalf("expected valuation 6 qty / 300 value, got %+v", rows)
	}
}

func TestOpenReceivablesReconcileToControlAccount(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	post, err := fx.postingSvc.Post(ctx, accounting.PostRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		Description:    "DAT-002 invoice",
		Lines: []accounting.PostingLine{
			{AccountID: fx.receivableID, Name: "AR", Debit: amount.FromFloat64(200)},
			{AccountID: fx.incomeID, Name: "Income", Credit: amount.FromFloat64(200)},
		},
	})
	if err != nil {
		t.Fatalf("post receivable failed: %v", err)
	}

	open, err := fx.moveLineDAO.ListUnreconciledByAccount(ctx, fx.receivableID)
	if err != nil {
		t.Fatalf("list unreconciled failed: %v", err)
	}
	var openSum float64
	for _, line := range open {
		openSum += line.Debit.Float64() - line.Credit.Float64()
	}
	if openSum != 200 {
		t.Fatalf("expected open sub-ledger 200, got %v", openSum)
	}
	if balance := glBalance(t, fx.orgID, fx.receivableID); balance != openSum {
		t.Fatalf("expected GL AR balance %v to match open sub-ledger, got %v", openSum, balance)
	}

	payment, err := fx.postingSvc.Post(ctx, accounting.PostRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		Description:    "DAT-002 payment",
		Lines: []accounting.PostingLine{
			{AccountID: fx.bankID, Name: "Bank", Debit: amount.FromFloat64(200)},
			{AccountID: fx.receivableID, Name: "AR", Credit: amount.FromFloat64(200)},
		},
	})
	if err != nil {
		t.Fatalf("post payment failed: %v", err)
	}

	lines, err := fx.moveLineDAO.ListByMovement(ctx, post.ID)
	if err != nil {
		t.Fatalf("list invoice lines failed: %v", err)
	}
	paymentLines, err := fx.moveLineDAO.ListByMovement(ctx, payment.ID)
	if err != nil {
		t.Fatalf("list payment lines failed: %v", err)
	}
	var debitID, creditID uint64
	for _, line := range lines {
		if line.Debit.IsPositive() && line.AccountID == fx.receivableID {
			debitID = line.ID
		}
	}
	for _, line := range paymentLines {
		if line.Credit.IsPositive() && line.AccountID == fx.receivableID {
			creditID = line.ID
		}
	}
	if debitID == 0 || creditID == 0 {
		t.Fatal("failed to locate AR debit and credit lines")
	}

	result, err := fx.reconcileSvc.Reconcile(ctx, accounting.ReconcileRequest{
		OrganizationID: fx.orgID,
		DebitLineID:    debitID,
		CreditLineID:   creditID,
		Amount:         200,
	})
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}
	if !result.Reconciled || result.Full == nil {
		t.Fatalf("expected full reconciliation, got %+v", result)
	}

	open, err = fx.moveLineDAO.ListUnreconciledByAccount(ctx, fx.receivableID)
	if err != nil {
		t.Fatalf("list unreconciled after reconcile failed: %v", err)
	}
	if len(open) != 0 {
		t.Fatalf("expected no open sub-ledger after reconcile, got %d lines", len(open))
	}
	if balance := glBalance(t, fx.orgID, fx.receivableID); balance != 0 {
		t.Fatalf("expected zeroed GL AR balance after payment, got %v", balance)
	}
}

func TestRebuildQuantsReconstructsOnHand(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	receive, err := fx.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(fx.orgID),
		ItemID:         fx.variantID,
		Qty:            10,
		SrcLocationID:  fx.supplierLocID,
		DstLocationID:  fx.internalLocID,
		State:          inventory.MovementStateConfirmed,
		ScheduledDate:  helper.Ptr(fx.date),
	})
	if err != nil {
		t.Fatalf("create receive movement failed: %v", err)
	}
	if _, err := fx.valuationSvc.Receive(ctx, receive.ID, amount.FromFloat64(50), fx.journalID, fx.date); err != nil {
		t.Fatalf("receive stock failed: %v", err)
	}
	ship, err := fx.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(fx.orgID),
		ItemID:         fx.variantID,
		Qty:            4,
		SrcLocationID:  fx.internalLocID,
		DstLocationID:  fx.customerLocID,
		State:          inventory.MovementStateConfirmed,
		ScheduledDate:  helper.Ptr(fx.date),
	})
	if err != nil {
		t.Fatalf("create ship movement failed: %v", err)
	}
	if _, err := fx.valuationSvc.Ship(ctx, ship.ID, fx.journalID, fx.date); err != nil {
		t.Fatalf("ship stock failed: %v", err)
	}

	quant, err := fx.stockQuantDAO.FindByKey(ctx, fx.variantID, fx.internalLocID, nil)
	if err != nil {
		t.Fatalf("find quant failed: %v", err)
	}
	if quant == nil || quant.Quantity != 6 {
		t.Fatalf("expected on-hand 6, got %+v", quant)
	}

	quant.Quantity = 999
	if _, err := fx.stockQuantDAO.Update(ctx, quant); err != nil {
		t.Fatalf("corrupt quant failed: %v", err)
	}

	if err := fx.ledgerSvc.RebuildQuants(ctx, helper.Ptr(fx.orgID)); err != nil {
		t.Fatalf("rebuild quants failed: %v", err)
	}

	rebuilt, err := fx.stockQuantDAO.FindByKey(ctx, fx.variantID, fx.internalLocID, nil)
	if err != nil {
		t.Fatalf("find rebuilt quant failed: %v", err)
	}
	if rebuilt == nil || rebuilt.Quantity != 6 {
		t.Fatalf("expected rebuilt on-hand 6, got %+v", rebuilt)
	}
}

func TestPartialUniqueIndexesAllowReuseAfterSoftDelete(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	username := gofakeit.Username()
	email := gofakeit.Email()
	first, err := fx.userSvc.Create(ctx, &iam.User{
		Username:  username,
		FirstName: gofakeit.FirstName(),
		Email:     email,
		Password:  "test-password",
	})
	if err != nil {
		t.Fatalf("create first user failed: %v", err)
	}
	if err := fx.userDAO.Delete(ctx, first.ID); err != nil {
		t.Fatalf("soft-delete user failed: %v", err)
	}

	if _, err := fx.userSvc.Create(ctx, &iam.User{
		Username:  username,
		FirstName: gofakeit.FirstName(),
		Email:     email,
		Password:  "test-password",
	}); err != nil {
		t.Fatalf("recreating deleted username/email should be allowed, got %v", err)
	}

	if _, err := fx.userSvc.Create(ctx, &iam.User{
		Username:  username,
		FirstName: gofakeit.FirstName(),
		Email:     gofakeit.Email(),
		Password:  "test-password",
	}); !errors.Is(err, iam.ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken for live duplicate, got %v", err)
	}

	if _, err := fx.userSvc.Create(ctx, &iam.User{
		Username:  gofakeit.Username(),
		FirstName: gofakeit.FirstName(),
		Email:     email,
		Password:  "test-password",
	}); !errors.Is(err, iam.ErrEmailRegistered) {
		t.Fatalf("expected ErrEmailRegistered for live duplicate, got %v", err)
	}
}

func TestTaxPeriodLockBlocksPosting(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	request := accounting.PostRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		Description:    "DAT-006",
		Lines: []accounting.PostingLine{
			{AccountID: fx.receivableID, Name: "AR", Debit: amount.FromFloat64(100)},
			{AccountID: fx.incomeID, Name: "Income", Credit: amount.FromFloat64(100)},
		},
	}
	if _, err := fx.postingSvc.Post(ctx, request); err != nil {
		t.Fatalf("posting into open period failed: %v", err)
	}

	if _, err := fx.taxPeriodSvc.Close(ctx, fx.periodID); err != nil {
		t.Fatalf("close period failed: %v", err)
	}

	if _, err := fx.postingSvc.Post(ctx, request); !errors.Is(err, accounting.ErrPeriodLocked) {
		t.Fatalf("expected ErrPeriodLocked after close, got %v", err)
	}
}

func TestReversalFullyReversesAndCannotReopen(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	original, err := fx.postingSvc.Post(ctx, accounting.PostRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		Description:    "DAT-007",
		Lines: []accounting.PostingLine{
			{AccountID: fx.receivableID, Name: "AR", Debit: amount.FromFloat64(100)},
			{AccountID: fx.incomeID, Name: "Income", Credit: amount.FromFloat64(100)},
		},
	})
	if err != nil {
		t.Fatalf("post original failed: %v", err)
	}

	reversal, err := fx.postingSvc.Reverse(ctx, accounting.ReverseRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		Description:    "DAT-007 reversal",
		MovementID:     original.ID,
	})
	if err != nil {
		t.Fatalf("reverse failed: %v", err)
	}

	originalLines, err := fx.moveLineDAO.ListByMovement(ctx, original.ID)
	if err != nil {
		t.Fatalf("list original lines failed: %v", err)
	}
	reversalLines, err := fx.moveLineDAO.ListByMovement(ctx, reversal.ID)
	if err != nil {
		t.Fatalf("list reversal lines failed: %v", err)
	}
	if len(reversalLines) != len(originalLines) {
		t.Fatalf("expected %d reversal lines, got %d", len(originalLines), len(reversalLines))
	}
	for i, line := range reversalLines {
		if !line.Debit.Equal(originalLines[i].Credit) || !line.Credit.Equal(originalLines[i].Debit) {
			t.Fatalf("expected reversal line %d to mirror original, got %+v", i, line)
		}
	}

	if reversal.OriginType == nil || *reversal.OriginType != "reversal" {
		t.Fatalf("expected reversal origin_type, got %+v", reversal.OriginType)
	}
	if reversal.OriginID == nil || *reversal.OriginID != original.ID {
		t.Fatalf("expected reversal to reference original movement, got %+v", reversal.OriginID)
	}

	if _, err := fx.postingSvc.Reverse(ctx, accounting.ReverseRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		Date:           fx.date,
		Description:    "DAT-007 double reversal",
		MovementID:     original.ID,
	}); !errors.Is(err, accounting.ErrMoveReversed) {
		t.Fatalf("expected ErrMoveReversed for double reversal, got %v", err)
	}
}
