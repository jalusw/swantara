package seeders

type crmSeed struct {
	PipelineStages []crmStageSeed `json:"pipeline_stages"`
	SalesGroups    []string       `json:"sales_groups"`
}

type crmStageSeed struct {
	Name        string  `json:"name"`
	Sequence    int     `json:"sequence"`
	IsWon       bool    `json:"is_won"`
	Probability float64 `json:"probability"`
}

func loadCRMSeed() (*crmSeed, error) {
	return loadSeedData[crmSeed]("data/crm.json")
}
