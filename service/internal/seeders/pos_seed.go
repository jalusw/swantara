package seeders

type posSeed struct {
	POSConfigs []posConfigSeed `json:"pos_configs"`
}

type posConfigSeed struct {
	Name      string `json:"name"`
	Warehouse string `json:"warehouse"`
	Journal   string `json:"journal"`
}

func loadPOSSeed() (*posSeed, error) {
	return loadSeedData[posSeed]("data/pos.json")
}
