package seeders

import "encoding/json"

type systemSeed struct {
	DefaultOrg    defaultOrgSeed     `json:"default_org"`
	AdminUser     adminUserSeed      `json:"admin_user"`
	SystemConfigs []systemConfigSeed `json:"system_configs"`
	DocSequences  []docSequenceSeed  `json:"doc_sequences"`
}

type defaultOrgSeed struct {
	Name              string `json:"name"`
	LegalName         string `json:"legal_name"`
	BaseCurrency      string `json:"base_currency"`
	CountryCode       string `json:"country_code"`
	Timezone          string `json:"timezone"`
	TaxYearStartMonth int    `json:"tax_year_start_month"`
}

type adminUserSeed struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Active    bool   `json:"active"`
}

type systemConfigSeed struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value" swaggertype:"object"`
}

type docSequenceSeed struct {
	Code        string `json:"code"`
	Prefix      string `json:"prefix"`
	Padding     int    `json:"padding"`
	ResetPeriod string `json:"reset_period"`
}

func loadSystemSeed() (*systemSeed, error) {
	return loadSeedData[systemSeed]("data/system.json")
}
