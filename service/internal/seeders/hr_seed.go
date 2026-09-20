package seeders

type hrSeed struct {
	Departments  []string          `json:"departments"`
	JobPositions []jobPositionSeed `json:"job_positions"`
	LeaveTypes   []leaveTypeSeed   `json:"leave_types"`
	SalaryRules  []salaryRuleSeed  `json:"salary_rules"`
}

type jobPositionSeed struct {
	Name       string `json:"name"`
	Department string `json:"department"`
}

type leaveTypeSeed struct {
	Name           string  `json:"name"`
	Paid           bool    `json:"paid"`
	AllocationDays float64 `json:"allocation_days"`
}

type salaryRuleSeed struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	ComputeType   string  `json:"compute_type"`
	Amount        float64 `json:"amount"`
	DebitAccount  string  `json:"debit_account"`
	CreditAccount string  `json:"credit_account"`
}

func loadHRSeed() (*hrSeed, error) {
	return loadSeedData[hrSeed]("data/hr.json")
}
