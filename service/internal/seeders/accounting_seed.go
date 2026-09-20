package seeders

type accountingSeed struct {
	FxRates           []fxRateSeed          `json:"fx_rates"`
	PaymentTerms      []paymentTermSeed     `json:"payment_terms"`
	Accounts          []accountSeed         `json:"accounts"`
	Dimensions        []string              `json:"dimensions"`
	Journals          []journalSeed         `json:"journals"`
	TaxYear           taxYearSeed           `json:"tax_year"`
	ReminderLevels    []reminderLevelSeed   `json:"reminder_levels"`
	ExpenseCategories []expenseCategorySeed `json:"expense_categories"`
	AssetCategories   assetCategorySeed     `json:"asset_categories"`
}

type fxRateSeed struct {
	CurrencyCode string  `json:"currency_code"`
	Rate         float64 `json:"rate"`
	RateType     string  `json:"rate_type"`
}

type paymentTermSeed struct {
	Name  string                `json:"name"`
	Note  string                `json:"note"`
	Lines []paymentTermLineSeed `json:"lines"`
}

type paymentTermLineSeed struct {
	Sequence  int     `json:"sequence"`
	ValueType string  `json:"value_type"`
	Value     float64 `json:"value"`
	DaysAfter int     `json:"days_after"`
}

type accountSeed struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Reconcilable bool   `json:"reconcilable"`
}

type journalSeed struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	Type           string `json:"type"`
	DefaultAccount string `json:"default_account"`
}

type taxYearSeed struct {
	Year  int    `json:"year"`
	Name  string `json:"name"`
	State string `json:"state"`
}

type reminderLevelSeed struct {
	Name        string `json:"name"`
	DaysOverdue int    `json:"days_overdue"`
	Sequence    int    `json:"sequence"`
}

type expenseCategorySeed struct {
	Name           string `json:"name"`
	ExpenseAccount string `json:"expense_account"`
}

type assetCategorySeed struct {
	Accounts   assetCategoryAccountSeed `json:"accounts"`
	Categories []assetCategoryItemSeed  `json:"categories"`
}

type assetCategoryAccountSeed struct {
	Asset        string `json:"asset"`
	Depreciation string `json:"depreciation"`
	Expense      string `json:"expense"`
	Gain         string `json:"gain"`
	Loss         string `json:"loss"`
}

type assetCategoryItemSeed struct {
	Name         string `json:"name"`
	Method       string `json:"method"`
	MethodNumber int    `json:"method_number"`
	MethodPeriod string `json:"method_period"`
}

func loadAccountingSeed() (*accountingSeed, error) {
	return loadSeedData[accountingSeed]("data/accounting.json")
}
