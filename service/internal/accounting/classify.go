package accounting

type Section string

const (
	SectionCurrentAsset        Section = "current_asset"
	SectionNonCurrentAsset     Section = "non_current_asset"
	SectionCurrentLiability    Section = "current_liability"
	SectionNonCurrentLiability Section = "non_current_liability"
	SectionEquity              Section = "equity"
	SectionRevenue             Section = "revenue"
	SectionCOGS                Section = "cogs"
	SectionOperatingExpense    Section = "operating_expense"
	SectionOtherIncome         Section = "other_income"
	SectionOtherExpense        Section = "other_expense"
	SectionTax                 Section = "tax"
	SectionDepreciation        Section = "depreciation"
)

type StatementLine struct {
	Name            string
	Code            string
	Amount          float64
	Section         Section
	CashFlowSection string
}

type StatementBalanceSheet struct {
	CurrentAssets         []StatementLine
	NonCurrentAssets      []StatementLine
	CurrentLiabilities    []StatementLine
	NonCurrentLiabilities []StatementLine
	Equity                []StatementLine
	TotalAssets           float64
	TotalLiabilities      float64
	TotalEquity           float64
}

type StatementProfitLoss struct {
	Revenue           []StatementLine
	COGS              []StatementLine
	GrossProfit       float64
	OperatingExpenses []StatementLine
	OperatingIncome   float64
	OtherIncome       []StatementLine
	OtherExpenses     []StatementLine
	ProfitBeforeTax   float64
	TaxExpense        []StatementLine
	NetIncome         float64
}

type StatementCashFlow struct {
	Operating   []StatementLine
	Investing   []StatementLine
	Financing   []StatementLine
	NetChange   float64
	OpeningCash float64
	ClosingCash float64
}

type AccountClassifier interface {
	BalanceSheetSection(accountType string) Section
	ProfitLossSection(accountType string) Section
	CashFlowSection(accountType string) string
	IsCashEquivalent(accountType string) bool
	IsCurrentAsset(accountType string) bool
	IsCurrentLiability(accountType string) bool
	IsIncome(accountType string) bool
	IsCOGS(accountType string) bool
	IsExpense(accountType string) bool
	IsDepreciation(accountType string) bool
}

type StatementFormatter interface {
	FormatBalanceSheet(lines []StatementLine) StatementBalanceSheet
	FormatProfitLoss(lines []StatementLine) StatementProfitLoss
	FormatCashFlow(lines []StatementLine, openingCash float64) StatementCashFlow
}

type GenericClassifier struct{}

func (GenericClassifier) BalanceSheetSection(accountType string) Section {
	switch accountType {
	case "cash", "bank":
		return SectionCurrentAsset
	case "receivable":
		return SectionCurrentAsset
	case "current_asset":
		return SectionCurrentAsset
	case "fixed_asset":
		return SectionNonCurrentAsset
	case "depreciation":
		return SectionNonCurrentAsset
	case "payable":
		return SectionCurrentLiability
	case "tax":
		return SectionCurrentLiability
	case "liability":
		return SectionCurrentLiability
	case "equity":
		return SectionEquity
	default:
		return SectionEquity
	}
}

func (GenericClassifier) ProfitLossSection(accountType string) Section {
	switch accountType {
	case "income":
		return SectionRevenue
	case "cogs":
		return SectionCOGS
	case "expense":
		return SectionOperatingExpense
	default:
		return SectionOperatingExpense
	}
}

func (GenericClassifier) CashFlowSection(accountType string) string {
	switch accountType {
	case "fixed_asset", "depreciation":
		return CashFlowInvesting
	case "equity":
		return CashFlowFinancing
	default:
		return CashFlowOperating
	}
}

func (GenericClassifier) IsCashEquivalent(accountType string) bool {
	return accountType == "cash" || accountType == "bank"
}

func (GenericClassifier) IsCurrentAsset(accountType string) bool {
	switch accountType {
	case "cash", "bank", "receivable", "current_asset":
		return true
	default:
		return false
	}
}

func (GenericClassifier) IsCurrentLiability(accountType string) bool {
	switch accountType {
	case "payable", "tax", "liability":
		return true
	default:
		return false
	}
}

func (GenericClassifier) IsIncome(accountType string) bool {
	return accountType == "income"
}

func (GenericClassifier) IsCOGS(accountType string) bool {
	return accountType == "cogs"
}

func (GenericClassifier) IsExpense(accountType string) bool {
	return accountType == "expense"
}

func (GenericClassifier) IsDepreciation(accountType string) bool {
	return accountType == "depreciation"
}

type GenericFormatter struct{}

func (GenericFormatter) FormatBalanceSheet(lines []StatementLine) StatementBalanceSheet {
	bs := StatementBalanceSheet{}
	for _, line := range lines {
		switch line.Section {
		case SectionCurrentAsset:
			bs.CurrentAssets = append(bs.CurrentAssets, line)
			bs.TotalAssets += line.Amount
		case SectionNonCurrentAsset:
			bs.NonCurrentAssets = append(bs.NonCurrentAssets, line)
			bs.TotalAssets += line.Amount
		case SectionCurrentLiability:
			bs.CurrentLiabilities = append(bs.CurrentLiabilities, line)
			bs.TotalLiabilities += line.Amount
		case SectionNonCurrentLiability:
			bs.NonCurrentLiabilities = append(bs.NonCurrentLiabilities, line)
			bs.TotalLiabilities += line.Amount
		case SectionEquity:
			bs.Equity = append(bs.Equity, line)
			bs.TotalEquity += line.Amount
		}
	}
	return bs
}

func (GenericFormatter) FormatProfitLoss(lines []StatementLine) StatementProfitLoss {
	pl := StatementProfitLoss{}
	for _, line := range lines {
		switch line.Section {
		case SectionRevenue:
			pl.Revenue = append(pl.Revenue, line)
		case SectionCOGS:
			pl.COGS = append(pl.COGS, line)
		case SectionOperatingExpense:
			pl.OperatingExpenses = append(pl.OperatingExpenses, line)
		case SectionOtherIncome:
			pl.OtherIncome = append(pl.OtherIncome, line)
		case SectionOtherExpense:
			pl.OtherExpenses = append(pl.OtherExpenses, line)
		case SectionTax:
			pl.TaxExpense = append(pl.TaxExpense, line)
		}
	}
	for _, l := range pl.Revenue {
		pl.GrossProfit += l.Amount
	}
	for _, l := range pl.COGS {
		pl.GrossProfit -= l.Amount
	}
	pl.OperatingIncome = pl.GrossProfit
	for _, l := range pl.OperatingExpenses {
		pl.OperatingIncome -= l.Amount
	}
	pl.ProfitBeforeTax = pl.OperatingIncome
	for _, l := range pl.OtherIncome {
		pl.ProfitBeforeTax += l.Amount
	}
	for _, l := range pl.OtherExpenses {
		pl.ProfitBeforeTax -= l.Amount
	}
	pl.NetIncome = pl.ProfitBeforeTax
	for _, l := range pl.TaxExpense {
		pl.NetIncome -= l.Amount
	}
	return pl
}

func (GenericFormatter) FormatCashFlow(lines []StatementLine, openingCash float64) StatementCashFlow {
	cf := StatementCashFlow{OpeningCash: openingCash}
	for _, line := range lines {
		switch line.CashFlowSection {
		case CashFlowOperating:
			cf.Operating = append(cf.Operating, line)
			cf.NetChange += line.Amount
		case CashFlowInvesting:
			cf.Investing = append(cf.Investing, line)
			cf.NetChange += line.Amount
		case CashFlowFinancing:
			cf.Financing = append(cf.Financing, line)
			cf.NetChange += line.Amount
		}
	}
	cf.ClosingCash = cf.OpeningCash + cf.NetChange
	return cf
}
