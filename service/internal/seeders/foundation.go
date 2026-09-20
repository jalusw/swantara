package seeders

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func seedMissing[E any](ctx context.Context, repo dao.Base[E], keyField string, organizationID *uint64, items []*E, key func(*E) any) (bool, error) {
	created := false
	for _, item := range items {
		var existing *E
		var err error
		if organizationID != nil {
			page, err := repo.List(ctx, &query.Query{Filters: []query.Filter{
				{Field: "organization_id", Operator: query.Equal, Value: *organizationID},
				{Field: keyField, Operator: query.Equal, Value: key(item)},
			}})
			if err != nil {
				return created, err
			}
			if len(page.Items) > 0 {
				existing = page.Items[0]
			}
		} else {
			existing, err = repo.Search(ctx, keyField, key(item))
			if err != nil {
				return created, err
			}
		}
		if existing != nil {
			continue
		}
		if _, err := repo.Create(ctx, item); err != nil {
			return created, err
		}
		created = true
	}
	return created, nil
}

func (s *Seeder) SeedFoundation() error {
	ctx := s.ctx()

	orgID, err := s.defaultOrg(ctx)
	if err != nil {
		return fmt.Errorf("seed default org: %w", err)
	}
	slog.Info("Default organization ready", "organization_id", orgID)

	if err := s.seedCurrencies(ctx); err != nil {
		return fmt.Errorf("seed currencies: %w", err)
	}
	if err := s.seedFxRates(ctx, orgID); err != nil {
		return fmt.Errorf("seed fx rates: %w", err)
	}
	if err := s.seedUnits(ctx); err != nil {
		return fmt.Errorf("seed units: %w", err)
	}
	if err := s.seedPaymentTerms(ctx); err != nil {
		return fmt.Errorf("seed payment terms: %w", err)
	}
	if err := s.seedAccounts(ctx, orgID); err != nil {
		return fmt.Errorf("seed accounts: %w", err)
	}
	if err := s.seedDimensions(ctx, orgID); err != nil {
		return fmt.Errorf("seed dimension accounts: %w", err)
	}
	if err := s.seedJournals(ctx, orgID); err != nil {
		return fmt.Errorf("seed journals: %w", err)
	}
	if err := s.seedTaxYear(ctx, orgID); err != nil {
		return fmt.Errorf("seed tax year: %w", err)
	}
	if err := s.seedProductCatalog(ctx, orgID); err != nil {
		return fmt.Errorf("seed item catalog: %w", err)
	}
	if err := s.seedCRM(ctx, orgID); err != nil {
		return fmt.Errorf("seed crm: %w", err)
	}
	if err := s.seedCarriers(ctx); err != nil {
		return fmt.Errorf("seed carriers: %w", err)
	}
	if err := s.seedReminder(ctx); err != nil {
		return fmt.Errorf("seed reminder: %w", err)
	}
	if err := s.seedWarehouses(ctx, orgID); err != nil {
		return fmt.Errorf("seed warehouses: %w", err)
	}
	if err := s.seedDepartments(ctx, orgID); err != nil {
		return fmt.Errorf("seed departments: %w", err)
	}
	if err := s.seedWorkCenters(ctx, orgID); err != nil {
		return fmt.Errorf("seed work centers: %w", err)
	}
	if err := s.seedLeaveTypes(ctx); err != nil {
		return fmt.Errorf("seed leave types: %w", err)
	}
	if err := s.seedSalaryRules(ctx); err != nil {
		return fmt.Errorf("seed salary rules: %w", err)
	}
	if err := s.seedAssetCategories(ctx, orgID); err != nil {
		return fmt.Errorf("seed asset categories: %w", err)
	}
	if err := s.seedQualityPoints(ctx, orgID); err != nil {
		return fmt.Errorf("seed quality points: %w", err)
	}
	if err := s.seedPOSConfigs(ctx, orgID); err != nil {
		return fmt.Errorf("seed pos configs: %w", err)
	}
	if err := s.seedExpenseCategories(ctx, orgID); err != nil {
		return fmt.Errorf("seed expense categories: %w", err)
	}
	if err := s.seedSubscriptionPlans(ctx, orgID); err != nil {
		return fmt.Errorf("seed subscription plans: %w", err)
	}
	if err := s.seedSystemConfigs(ctx, orgID); err != nil {
		return fmt.Errorf("seed system configs: %w", err)
	}
	if err := s.seedDocSequences(ctx, orgID); err != nil {
		return fmt.Errorf("seed doc sequences: %w", err)
	}

	slog.Info("Foundation reference data seeded.")
	return nil
}

func (s *Seeder) defaultOrg(ctx context.Context) (uint64, error) {
	seed, err := loadSystemSeed()
	if err != nil {
		return 0, err
	}
	organizationDAO := dao.NewBase[reference.Organization](s.db)

	org, err := organizationDAO.Search(ctx, "name", seed.DefaultOrg.Name)
	if err != nil {
		return 0, err
	}
	if org != nil {
		return org.ID, nil
	}

	created, err := organizationDAO.Create(ctx, &reference.Organization{
		Name:              seed.DefaultOrg.Name,
		LegalName:         helper.Ptr(seed.DefaultOrg.LegalName),
		BaseCurrency:      seed.DefaultOrg.BaseCurrency,
		CountryCode:       helper.Ptr(seed.DefaultOrg.CountryCode),
		Timezone:          seed.DefaultOrg.Timezone,
		TaxYearStartMonth: int16(seed.DefaultOrg.TaxYearStartMonth),
	})
	if err != nil {
		return 0, err
	}
	return created.ID, nil
}

func (s *Seeder) seedCurrencies(ctx context.Context) error {
	seed, err := LoadCurrencySeed()
	if err != nil {
		return err
	}
	currencyDAO := dao.NewBase[reference.Currency](s.db)
	currencies := make([]*reference.Currency, 0, len(seed.Currencies))
	for _, item := range seed.Currencies {
		symbol := item.Symbol
		currencies = append(currencies, &reference.Currency{
			Code:          item.Code,
			Name:          item.Name,
			Symbol:        &symbol,
			DecimalPlaces: item.DecimalPlaces,
			Rounding:      math.Pow(10, -float64(item.DecimalPlaces)),
		})
	}
	_, err = seedMissing(ctx, currencyDAO, "code", nil, currencies, func(c *reference.Currency) any { return c.Code })
	return err
}

func (s *Seeder) seedFxRates(ctx context.Context, orgID uint64) error {
	seed, err := loadAccountingSeed()
	if err != nil {
		return err
	}
	fxRateDAO := dao.NewBase[reference.FxRate](s.db)

	today := time.Now().Truncate(24 * time.Hour)
	rates := make([]*reference.FxRate, 0, len(seed.FxRates))
	for _, def := range seed.FxRates {
		rates = append(rates, &reference.FxRate{
			CurrencyCode:   def.CurrencyCode,
			OrganizationID: helper.Ptr(orgID),
			Rate:           def.Rate,
			RateType:       def.RateType,
			ValidFrom:      today,
		})
	}
	_, err = seedMissing(ctx, fxRateDAO, "currency_code", helper.Ptr(orgID), rates, func(r *reference.FxRate) any { return r.CurrencyCode })
	return err
}

func (s *Seeder) seedUnits(ctx context.Context) error {
	seed, err := loadInventorySeed()
	if err != nil {
		return err
	}
	unitGroupDAO := dao.NewBase[reference.UnitGroup](s.db)
	uomDAO := dao.NewBase[reference.Unit](s.db)

	categories := make([]*reference.UnitGroup, 0, len(seed.UnitCategories))
	for _, name := range seed.UnitCategories {
		categories = append(categories, &reference.UnitGroup{Name: name})
	}
	if _, err := seedMissing(ctx, unitGroupDAO, "name", nil, categories, func(c *reference.UnitGroup) any { return c.Name }); err != nil {
		return err
	}

	for _, def := range seed.Units {
		existing, err := uomDAO.Search(ctx, "name", def.Name)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}
		categoryID, err := refID(ctx, unitGroupDAO, "name", def.Category)
		if err != nil {
			return err
		}
		unit := &reference.Unit{
			Name:     def.Name,
			Factor:   def.Factor,
			UnitType: def.Type,
			Rounding: def.Rounding,
		}
		unit.CategoryID = categoryID
		if _, err := uomDAO.Create(ctx, unit); err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) seedPaymentTerms(ctx context.Context) error {
	seed, err := loadAccountingSeed()
	if err != nil {
		return err
	}
	orgID, err := s.defaultOrg(ctx)
	if err != nil {
		return err
	}
	return s.seedPaymentTermsForOrg(ctx, orgID, seed)
}

func (s *Seeder) seedPaymentTermsForOrg(ctx context.Context, orgID uint64, seed *accountingSeed) error {
	paymentTermDAO := reference.NewPaymentTermDAO(s.db)

	terms := map[string]*reference.PaymentTerm{}
	for _, def := range seed.PaymentTerms {
		page, err := paymentTermDAO.List(ctx, &query.Query{Filters: []query.Filter{
			{Field: "organization_id", Operator: query.Equal, Value: orgID},
			{Field: "name", Operator: query.Equal, Value: def.Name},
		}})
		if err != nil {
			return err
		}
		var existing *reference.PaymentTerm
		if len(page.Items) > 0 {
			existing = page.Items[0]
		}
		code := paymentTermCodeForName(def.Name)
		templateKey := paymentTermTemplateKeyForName(def.Name)
		term := &reference.PaymentTerm{OrganizationID: orgID, Name: def.Name, Note: helper.Ptr(def.Note), Code: code, TemplateKey: templateKey, IsActive: true}
		if existing == nil {
			created, err := paymentTermDAO.Create(ctx, term)
			if err != nil {
				return err
			}
			term = created
		} else {
			term = existing
		}
		terms[def.Name] = term
	}

	for _, def := range seed.PaymentTerms {
		stored := terms[def.Name]
		existingLines, err := paymentTermDAO.ListLines(ctx, stored.ID)
		if err != nil {
			return err
		}
		if len(existingLines) > 0 {
			continue
		}
		lines := make([]*reference.PaymentTermLine, 0, len(def.Lines))
		for _, lineDef := range def.Lines {
			lines = append(lines, &reference.PaymentTermLine{
				Sequence:  lineDef.Sequence,
				ValueType: lineDef.ValueType,
				Value:     lineDef.Value,
				DaysAfter: lineDef.DaysAfter,
			})
		}
		if err := paymentTermDAO.ReplaceLines(ctx, stored.ID, lines); err != nil {
			return err
		}
	}
	return nil
}

func paymentTermCodeForName(name string) *string {
	mapping := map[string]string{
		"Immediate Payment":           "immediate",
		"Net 15":                      "net_15",
		"Net 30":                      "net_30",
		"Net 45":                      "net_45",
		"Net 60":                      "net_60",
		"2/10 Net 30":                 "2_10_net30",
		"30% Advance, Balance Net 30": "30_adv_balance",
		"50/50":                       "50_50",
	}
	if code, ok := mapping[name]; ok {
		return helper.Ptr(code)
	}
	return nil
}

func paymentTermTemplateKeyForName(name string) *string {
	if code := paymentTermCodeForName(name); code != nil {
		return code
	}
	return helper.Ptr(name)
}

func (s *Seeder) seedAccounts(ctx context.Context, orgID uint64) error {
	seed, err := loadAccountingSeed()
	if err != nil {
		return err
	}
	accountDAO := dao.NewBase[reference.Account](s.db)

	accounts := make([]*reference.Account, 0, len(seed.Accounts))
	for _, def := range seed.Accounts {
		accounts = append(accounts, &reference.Account{
			OrganizationID: orgID,
			Code:           def.Code,
			Name:           def.Name,
			Type:           def.Type,
			Reconcilable:   def.Reconcilable,
			Active:         true,
		})
	}
	_, err = seedMissing(ctx, accountDAO, "code", helper.Ptr(orgID), accounts, func(a *reference.Account) any { return a.Code })
	return err
}

func (s *Seeder) seedDimensions(ctx context.Context, orgID uint64) error {
	seed, err := loadAccountingSeed()
	if err != nil {
		return err
	}
	dimensionDAO := dao.NewBase[reference.Dimension](s.db)

	for _, name := range seed.Dimensions {
		existing, err := dimensionDAO.Search(ctx, "name", name)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}
		if _, err := dimensionDAO.Create(ctx, &reference.Dimension{
			OrganizationID: helper.Ptr(orgID),
			Name:           name,
			Kind:           helper.Ptr("general"),
			Active:         helper.Ptr(true),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) seedJournals(ctx context.Context, orgID uint64) error {
	seed, err := loadAccountingSeed()
	if err != nil {
		return err
	}
	accountDAO := dao.NewBase[reference.Account](s.db)
	journalDAO := dao.NewBase[reference.Journal](s.db)

	journals := make([]*reference.Journal, 0, len(seed.Journals))
	for _, def := range seed.Journals {
		journal := &reference.Journal{
			OrganizationID: orgID,
			Name:           def.Name,
			Code:           helper.Ptr(def.Code),
			Type:           def.Type,
		}
		if def.DefaultAccount != "" {
			accountID, err := refID(ctx, accountDAO, "code", def.DefaultAccount)
			if err != nil {
				return err
			}
			journal.DefaultAccountID = helper.Ptr(accountID)
		}
		journals = append(journals, journal)
	}
	_, err = seedMissing(ctx, journalDAO, "code", helper.Ptr(orgID), journals, func(j *reference.Journal) any { return *j.Code })
	return err
}

func (s *Seeder) seedTaxYear(ctx context.Context, orgID uint64) error {
	seed, err := loadAccountingSeed()
	if err != nil {
		return err
	}
	taxYearDAO := dao.NewBase[reference.TaxYear](s.db)

	fy := seed.TaxYear
	start := time.Date(fy.Year, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(fy.Year, time.December, 31, 23, 59, 59, 0, time.UTC)
	taxYear := &reference.TaxYear{
		OrganizationID: helper.Ptr(orgID),
		Name:           fy.Name,
		DateStart:      helper.Ptr(start),
		DateEnd:        helper.Ptr(end),
		State:          helper.Ptr(fy.State),
	}
	_, err = seedMissing(ctx, taxYearDAO, "name", helper.Ptr(orgID), []*reference.TaxYear{taxYear}, func(f *reference.TaxYear) any { return f.Name })
	return err
}

func (s *Seeder) seedProductCatalog(ctx context.Context, orgID uint64) error {
	seed, err := loadInventorySeed()
	if err != nil {
		return err
	}
	accountDAO := dao.NewBase[reference.Account](s.db)
	itemCategoryDAO := dao.NewBase[reference.ItemCategory](s.db)
	itemAttributeDAO := dao.NewBase[reference.ItemAttribute](s.db)
	itemAttributeValueDAO := dao.NewBase[reference.ItemAttributeValue](s.db)

	categories := make([]*reference.ItemCategory, 0, len(seed.ProductCategories))
	for _, def := range seed.ProductCategories {
		category := &reference.ItemCategory{Name: def.Name, OrganizationID: helper.Ptr(orgID)}
		if def.IncomeAccount != "" {
			accountID, err := refID(ctx, accountDAO, "code", def.IncomeAccount)
			if err != nil {
				return err
			}
			category.IncomeAccountID = helper.Ptr(accountID)
		}
		if def.ExpenseAccount != "" {
			accountID, err := refID(ctx, accountDAO, "code", def.ExpenseAccount)
			if err != nil {
				return err
			}
			category.ExpenseAccountID = helper.Ptr(accountID)
		}
		if def.CogsAccount != "" {
			accountID, err := refID(ctx, accountDAO, "code", def.CogsAccount)
			if err != nil {
				return err
			}
			category.CogsAccountID = helper.Ptr(accountID)
		}
		if def.StockValuationAccount != "" {
			accountID, err := refID(ctx, accountDAO, "code", def.StockValuationAccount)
			if err != nil {
				return err
			}
			category.StockValuationAccountID = helper.Ptr(accountID)
		}
		if def.StockInputAccount != "" {
			accountID, err := refID(ctx, accountDAO, "code", def.StockInputAccount)
			if err != nil {
				return err
			}
			category.StockInputAccountID = helper.Ptr(accountID)
		}
		if def.StockOutputAccount != "" {
			accountID, err := refID(ctx, accountDAO, "code", def.StockOutputAccount)
			if err != nil {
				return err
			}
			category.StockOutputAccountID = helper.Ptr(accountID)
		}
		if def.CostMethod != "" {
			category.CostMethod = helper.Ptr(def.CostMethod)
		}
		if def.Valuation != "" {
			category.Valuation = helper.Ptr(def.Valuation)
		}
		categories = append(categories, category)
	}
	if _, err := seedMissing(ctx, itemCategoryDAO, "name", helper.Ptr(orgID), categories, func(c *reference.ItemCategory) any { return c.Name }); err != nil {
		return err
	}

	if err := s.seedItems(ctx, orgID, seed); err != nil {
		return err
	}

	attributes := make([]*reference.ItemAttribute, 0, len(seed.ItemAttributes))
	for _, def := range seed.ItemAttributes {
		attributes = append(attributes, &reference.ItemAttribute{Name: def.Name})
	}
	createdAttributes, err := seedMissing(ctx, itemAttributeDAO, "name", nil, attributes, func(a *reference.ItemAttribute) any { return a.Name })
	if err != nil {
		return err
	}
	if !createdAttributes {
		return nil
	}

	for _, def := range seed.ItemAttributes {
		var attribute *reference.ItemAttribute
		for _, candidate := range attributes {
			if candidate.Name == def.Name {
				attribute = candidate
				break
			}
		}
		for _, value := range def.Values {
			if _, err := itemAttributeValueDAO.Create(ctx, &reference.ItemAttributeValue{
				AttributeID: attribute.ID,
				Value:       value,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Seeder) seedItems(ctx context.Context, orgID uint64, seed *inventorySeed) error {
	uomDAO := dao.NewBase[reference.Unit](s.db)
	templateDAO := dao.NewBase[products.Item](s.db)
	variantDAO := dao.NewBase[products.ItemVariant](s.db)
	categoryDAO := dao.NewBase[reference.ItemCategory](s.db)

	for _, def := range seed.Items {
		existing, err := templateDAO.Search(ctx, "name", def.Name)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}

		categoryID, err := refID(ctx, categoryDAO, "name", def.Category)
		if err != nil {
			return err
		}
		uomID, err := refID(ctx, uomDAO, "name", "Unit")
		if err != nil {
			return err
		}
		template := &products.Item{
			OrganizationID: helper.Ptr(orgID),
			Name:           def.Name,
			CategoryID:     helper.Ptr(categoryID),
			Type:           def.Type,
			UnitID:         &uomID,
			PurchaseUnitID: &uomID,
			ListPrice:      def.ListPrice,
			StandardCost:   def.StandardCost,
			IsPurchasable:  true,
			IsSellable:     true,
			Tracking:       "none",
		}
		if def.IsPurchasable != nil {
			template.IsPurchasable = *def.IsPurchasable
		}
		if def.IsSellable != nil {
			template.IsSellable = *def.IsSellable
		}
		created, err := templateDAO.Create(ctx, template)
		if err != nil {
			return err
		}
		if _, err := variantDAO.Create(ctx, &products.ItemVariant{
			ItemID: created.ID,
			Sku:    helper.Ptr(def.SKU),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) seedCRM(ctx context.Context, orgID uint64) error {
	seed, err := loadCRMSeed()
	if err != nil {
		return err
	}
	crmStageDAO := dao.NewBase[reference.PipelineStage](s.db)
	salesTeamDAO := dao.NewBase[reference.SalesGroup](s.db)

	stages := make([]*reference.PipelineStage, 0, len(seed.PipelineStages))
	for _, def := range seed.PipelineStages {
		stages = append(stages, &reference.PipelineStage{
			Name:           def.Name,
			Sequence:       def.Sequence,
			IsWon:          def.IsWon,
			Probability:    def.Probability,
			OrganizationID: helper.Ptr(orgID),
		})
	}
	if _, err := seedMissing(ctx, crmStageDAO, "name", helper.Ptr(orgID), stages, func(c *reference.PipelineStage) any { return c.Name }); err != nil {
		return err
	}

	teams := make([]*reference.SalesGroup, 0, len(seed.SalesGroups))
	for _, name := range seed.SalesGroups {
		teams = append(teams, &reference.SalesGroup{Name: name, OrganizationID: helper.Ptr(orgID)})
	}
	_, err = seedMissing(ctx, salesTeamDAO, "name", helper.Ptr(orgID), teams, func(t *reference.SalesGroup) any { return t.Name })
	return err
}

func (s *Seeder) seedCarriers(ctx context.Context) error {
	seed, err := loadInventorySeed()
	if err != nil {
		return err
	}
	carrierDAO := dao.NewBase[reference.Carrier](s.db)

	carriers := make([]*reference.Carrier, 0, len(seed.Carriers))
	for _, name := range seed.Carriers {
		carriers = append(carriers, &reference.Carrier{Name: name})
	}
	_, err = seedMissing(ctx, carrierDAO, "name", nil, carriers, func(c *reference.Carrier) any { return c.Name })
	return err
}

func (s *Seeder) seedReminder(ctx context.Context) error {
	seed, err := loadAccountingSeed()
	if err != nil {
		return err
	}
	reminderLevelDAO := dao.NewBase[reference.ReminderLevel](s.db)

	levels := make([]*reference.ReminderLevel, 0, len(seed.ReminderLevels))
	for _, def := range seed.ReminderLevels {
		levels = append(levels, &reference.ReminderLevel{
			Name:        def.Name,
			DaysOverdue: def.DaysOverdue,
			Sequence:    def.Sequence,
		})
	}
	_, err = seedMissing(ctx, reminderLevelDAO, "name", nil, levels, func(d *reference.ReminderLevel) any { return d.Name })
	return err
}

func (s *Seeder) seedWarehouses(ctx context.Context, orgID uint64) error {
	seed, err := loadInventorySeed()
	if err != nil {
		return err
	}
	warehouseDAO := dao.NewBase[reference.Warehouse](s.db)
	stockLocationDAO := dao.NewBase[reference.StockLocation](s.db)

	warehouseDef := seed.Warehouses[0]
	warehouse := &reference.Warehouse{OrganizationID: helper.Ptr(orgID), Name: warehouseDef.Name, Code: helper.Ptr(warehouseDef.Code)}
	existingWarehouse, err := warehouseDAO.Search(ctx, "name", warehouse.Name)
	if err != nil {
		return err
	}
	warehouseID := uint64(0)
	if existingWarehouse != nil {
		warehouseID = existingWarehouse.ID
	} else {
		created, err := warehouseDAO.Create(ctx, warehouse)
		if err != nil {
			return err
		}
		warehouseID = created.ID
	}

	locationIDs := map[string]uint64{}
	for _, loc := range seed.StockLocations {
		existing, err := stockLocationDAO.Search(ctx, "name", loc.Name)
		if err != nil {
			return err
		}
		if existing != nil {
			if loc.Physical && existing.OrganizationID == nil {
				existing.OrganizationID = helper.Ptr(orgID)
				if _, err := stockLocationDAO.Update(ctx, existing); err != nil {
					return err
				}
			}
			locationIDs[loc.Name] = existing.ID
			continue
		}
		item := &reference.StockLocation{Name: loc.Name, Code: helper.Ptr(loc.Code), Usage: loc.Usage}
		if loc.Physical {
			item.WarehouseID = helper.Ptr(warehouseID)
			item.OrganizationID = helper.Ptr(orgID)
		}
		if loc.Parent != "" {
			item.ParentID = helper.Ptr(locationIDs[loc.Parent])
		}
		created, err := stockLocationDAO.Create(ctx, item)
		if err != nil {
			return err
		}
		locationIDs[loc.Name] = created.ID
	}
	return nil
}

func (s *Seeder) seedDepartments(ctx context.Context, orgID uint64) error {
	seed, err := loadHRSeed()
	if err != nil {
		return err
	}
	departmentDAO := dao.NewBase[reference.Department](s.db)
	jobPositionDAO := dao.NewBase[reference.JobPosition](s.db)

	for _, name := range seed.Departments {
		existing, err := departmentDAO.Search(ctx, "name", name)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}
		if _, err := departmentDAO.Create(ctx, &reference.Department{OrganizationID: helper.Ptr(orgID), Name: name}); err != nil {
			return err
		}
	}

	for _, position := range seed.JobPositions {
		existing, err := jobPositionDAO.Search(ctx, "name", position.Name)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}
		departmentID, err := refID(ctx, departmentDAO, "name", position.Department)
		if err != nil {
			return err
		}
		if _, err := jobPositionDAO.Create(ctx, &reference.JobPosition{Name: position.Name, DepartmentID: helper.Ptr(departmentID)}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) seedWorkCenters(ctx context.Context, orgID uint64) error {
	seed, err := loadManufacturingSeed()
	if err != nil {
		return err
	}
	dimensionDAO := dao.NewBase[reference.Dimension](s.db)
	workCenterDAO := dao.NewBase[reference.WorkCenter](s.db)

	centers := make([]*reference.WorkCenter, 0, len(seed.WorkCenters))
	for _, def := range seed.WorkCenters {
		dimensionID, err := refID(ctx, dimensionDAO, "name", def.Dimension)
		if err != nil {
			return err
		}
		centers = append(centers, &reference.WorkCenter{
			OrganizationID:  helper.Ptr(orgID),
			Name:            def.Name,
			Code:            helper.Ptr(def.Code),
			CostPerHour:     helper.Ptr(def.CostPerHour),
			CapacityPerHour: helper.Ptr(def.CapacityPerHour),
			EfficiencyPct:   def.EfficiencyPct,
			OeeTarget:       helper.Ptr(def.OeeTarget),
			DimensionID:     helper.Ptr(dimensionID),
		})
	}
	_, err = seedMissing(ctx, workCenterDAO, "name", helper.Ptr(orgID), centers, func(w *reference.WorkCenter) any { return w.Name })
	return err
}

func (s *Seeder) seedLeaveTypes(ctx context.Context) error {
	seed, err := loadHRSeed()
	if err != nil {
		return err
	}
	leaveTypeDAO := dao.NewBase[reference.LeaveType](s.db)

	leaveTypes := make([]*reference.LeaveType, 0, len(seed.LeaveTypes))
	for _, def := range seed.LeaveTypes {
		leaveTypes = append(leaveTypes, &reference.LeaveType{
			Name:           def.Name,
			Paid:           def.Paid,
			AllocationDays: helper.Ptr(def.AllocationDays),
		})
	}
	_, err = seedMissing(ctx, leaveTypeDAO, "name", nil, leaveTypes, func(l *reference.LeaveType) any { return l.Name })
	return err
}

func (s *Seeder) seedSalaryRules(ctx context.Context) error {
	seed, err := loadHRSeed()
	if err != nil {
		return err
	}
	accountDAO := dao.NewBase[reference.Account](s.db)
	salaryRuleDAO := dao.NewBase[reference.SalaryRule](s.db)

	rules := make([]*reference.SalaryRule, 0, len(seed.SalaryRules))
	for _, def := range seed.SalaryRules {
		debitID, err := refID(ctx, accountDAO, "code", def.DebitAccount)
		if err != nil {
			return err
		}
		creditID, err := refID(ctx, accountDAO, "code", def.CreditAccount)
		if err != nil {
			return err
		}
		rule := &reference.SalaryRule{
			Code:            def.Code,
			Name:            def.Name,
			Category:        helper.Ptr(def.Category),
			ComputeType:     helper.Ptr(def.ComputeType),
			AccountDebitID:  helper.Ptr(debitID),
			AccountCreditID: helper.Ptr(creditID),
		}
		if def.Amount != 0 {
			rule.Amount = helper.Ptr(def.Amount)
		}
		rules = append(rules, rule)
	}
	_, err = seedMissing(ctx, salaryRuleDAO, "code", nil, rules, func(r *reference.SalaryRule) any { return r.Code })
	return err
}

func (s *Seeder) seedAssetCategories(ctx context.Context, orgID uint64) error {
	seed, err := loadAccountingSeed()
	if err != nil {
		return err
	}
	accountDAO := dao.NewBase[reference.Account](s.db)
	assetCategoryDAO := dao.NewBase[reference.AssetCategory](s.db)

	accounts := seed.AssetCategories.Accounts
	assetAcc, err := refID(ctx, accountDAO, "code", accounts.Asset)
	if err != nil {
		return err
	}
	depreciationAcc, err := refID(ctx, accountDAO, "code", accounts.Depreciation)
	if err != nil {
		return err
	}
	expenseAcc, err := refID(ctx, accountDAO, "code", accounts.Expense)
	if err != nil {
		return err
	}
	gainAcc, err := refID(ctx, accountDAO, "code", accounts.Gain)
	if err != nil {
		return err
	}
	lossAcc, err := refID(ctx, accountDAO, "code", accounts.Loss)
	if err != nil {
		return err
	}

	categories := make([]*reference.AssetCategory, 0, len(seed.AssetCategories.Categories))
	for _, def := range seed.AssetCategories.Categories {
		categories = append(categories, &reference.AssetCategory{
			Name:                  def.Name,
			OrganizationID:        helper.Ptr(orgID),
			AssetAccountID:        helper.Ptr(assetAcc),
			DepreciationAccountID: helper.Ptr(depreciationAcc),
			ExpenseAccountID:      helper.Ptr(expenseAcc),
			GainAccountID:         helper.Ptr(gainAcc),
			LossAccountID:         helper.Ptr(lossAcc),
			Method:                helper.Ptr(def.Method),
			MethodNumber:          helper.Ptr(def.MethodNumber),
			MethodPeriod:          helper.Ptr(def.MethodPeriod),
		})
	}
	_, err = seedMissing(ctx, assetCategoryDAO, "name", helper.Ptr(orgID), categories, func(c *reference.AssetCategory) any { return c.Name })
	return err
}

func (s *Seeder) seedQualityPoints(ctx context.Context, orgID uint64) error {
	seed, err := loadManufacturingSeed()
	if err != nil {
		return err
	}
	uomDAO := dao.NewBase[reference.Unit](s.db)
	qualityPointDAO := dao.NewBase[reference.QualityPoint](s.db)

	uomID, err := refID(ctx, uomDAO, "name", seed.QualityPoints[0].Unit)
	if err != nil {
		return err
	}
	points := make([]*reference.QualityPoint, 0, len(seed.QualityPoints))
	for _, def := range seed.QualityPoints {
		points = append(points, &reference.QualityPoint{
			OrganizationID: helper.Ptr(orgID),
			Operation:      helper.Ptr(def.Operation),
			TestType:       def.TestType,
			UnitID:         helper.Ptr(uomID),
		})
	}
	_, err = seedMissing(ctx, qualityPointDAO, "operation", helper.Ptr(orgID), points, func(q *reference.QualityPoint) any { return *q.Operation })
	return err
}

func (s *Seeder) seedPOSConfigs(ctx context.Context, orgID uint64) error {
	seed, err := loadPOSSeed()
	if err != nil {
		return err
	}
	warehouseDAO := dao.NewBase[reference.Warehouse](s.db)
	journalDAO := dao.NewBase[reference.Journal](s.db)
	posConfigDAO := dao.NewBase[reference.POSConfig](s.db)

	configs := make([]*reference.POSConfig, 0, len(seed.POSConfigs))
	for _, def := range seed.POSConfigs {
		warehouseID, err := refID(ctx, warehouseDAO, "name", def.Warehouse)
		if err != nil {
			return err
		}
		journalID, err := refID(ctx, journalDAO, "code", def.Journal)
		if err != nil {
			return err
		}
		configs = append(configs, &reference.POSConfig{
			Name:           def.Name,
			OrganizationID: helper.Ptr(orgID),
			WarehouseID:    helper.Ptr(warehouseID),
			JournalID:      helper.Ptr(journalID),
		})
	}
	_, err = seedMissing(ctx, posConfigDAO, "name", helper.Ptr(orgID), configs, func(p *reference.POSConfig) any { return p.Name })
	return err
}

func (s *Seeder) seedExpenseCategories(ctx context.Context, orgID uint64) error {
	seed, err := loadAccountingSeed()
	if err != nil {
		return err
	}
	accountDAO := dao.NewBase[reference.Account](s.db)
	expenseCategoryDAO := dao.NewBase[reference.ExpenseCategory](s.db)

	categories := make([]*reference.ExpenseCategory, 0, len(seed.ExpenseCategories))
	for _, def := range seed.ExpenseCategories {
		accountID, err := refID(ctx, accountDAO, "code", def.ExpenseAccount)
		if err != nil {
			return err
		}
		categories = append(categories, &reference.ExpenseCategory{
			OrganizationID:   helper.Ptr(orgID),
			Name:             def.Name,
			ExpenseAccountID: helper.Ptr(accountID),
		})
	}
	_, err = seedMissing(ctx, expenseCategoryDAO, "name", helper.Ptr(orgID), categories, func(c *reference.ExpenseCategory) any { return c.Name })
	return err
}

func (s *Seeder) seedSubscriptionPlans(ctx context.Context, orgID uint64) error {
	seed, err := loadSubscriptionSeed()
	if err != nil {
		return err
	}
	subscriptionPlanDAO := dao.NewBase[reference.SubscriptionPlan](s.db)

	plans := make([]*reference.SubscriptionPlan, 0, len(seed.SubscriptionPlans))
	for _, def := range seed.SubscriptionPlans {
		plans = append(plans, &reference.SubscriptionPlan{
			Name:              def.Name,
			RecurringInterval: def.Interval,
			RecurringCount:    def.Count,
			OrganizationID:    helper.Ptr(orgID),
		})
	}
	_, err = seedMissing(ctx, subscriptionPlanDAO, "name", helper.Ptr(orgID), plans, func(p *reference.SubscriptionPlan) any { return p.Name })
	return err
}

func (s *Seeder) seedSystemConfigs(ctx context.Context, orgID uint64) error {
	seed, err := loadSystemSeed()
	if err != nil {
		return err
	}
	systemConfigDAO := dao.NewBase[reference.SystemConfig](s.db)

	configs := make([]*reference.SystemConfig, 0, len(seed.SystemConfigs))
	for _, def := range seed.SystemConfigs {
		configs = append(configs, &reference.SystemConfig{
			OrganizationID: helper.Ptr(orgID),
			Key:            def.Key,
			Value:          def.Value,
		})
	}
	_, err = seedMissing(ctx, systemConfigDAO, "key", helper.Ptr(orgID), configs, func(c *reference.SystemConfig) any { return c.Key })
	return err
}

func (s *Seeder) seedApproverIDs(ctx context.Context, orgID uint64, approverIDs []uint64) error {
	value, err := json.Marshal(approverIDs)
	if err != nil {
		return err
	}
	systemConfigDAO := dao.NewBase[reference.SystemConfig](s.db)
	configs := []*reference.SystemConfig{
		{OrganizationID: helper.Ptr(orgID), Key: "purchase.approver_ids", Value: value},
	}
	_, err = seedMissing(ctx, systemConfigDAO, "key", helper.Ptr(orgID), configs, func(c *reference.SystemConfig) any { return c.Key })
	return err
}

func (s *Seeder) seedDocSequences(ctx context.Context, orgID uint64) error {
	seed, err := loadSystemSeed()
	if err != nil {
		return err
	}

	seqDao := dao.NewBase[sequence.DocumentSequence](s.db)
	sequences := make([]*sequence.DocumentSequence, 0, len(seed.DocSequences))
	for _, def := range seed.DocSequences {
		sequences = append(sequences, &sequence.DocumentSequence{
			OrganizationID: orgID,
			Code:           def.Code,
			Prefix:         def.Prefix,
			NextNumber:     1,
			Padding:        def.Padding,
			ResetPeriod:    def.ResetPeriod,
		})
	}
	_, err = seedMissing(ctx, seqDao, "code", helper.Ptr(orgID), sequences, func(d *sequence.DocumentSequence) any { return d.Code })
	return err
}

func refID[E interface{ GetID() uint64 }](ctx context.Context, dao dao.Base[E], field string, value any) (uint64, error) {
	entity, err := dao.Search(ctx, field, value)
	if err != nil {
		return 0, fmt.Errorf("lookup %s=%v: %w", field, value, err)
	}
	if entity == nil {
		return 0, fmt.Errorf("reference %s=%v not found: %w", field, value, ErrReferenceMissing)
	}
	return (*entity).GetID(), nil
}
