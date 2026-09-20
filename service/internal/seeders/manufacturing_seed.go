package seeders

type manufacturingSeed struct {
	WorkCenters   []workCenterSeed   `json:"work_centers"`
	QualityPoints []qualityPointSeed `json:"quality_points"`
}

type workCenterSeed struct {
	Name            string  `json:"name"`
	Code            string  `json:"code"`
	CostPerHour     float64 `json:"cost_per_hour"`
	CapacityPerHour float64 `json:"capacity_per_hour"`
	EfficiencyPct   float64 `json:"efficiency_pct"`
	OeeTarget       float64 `json:"oee_target"`
	Dimension       string  `json:"dimension"`
}

type qualityPointSeed struct {
	Operation string `json:"operation"`
	TestType  string `json:"test_type"`
	Unit      string `json:"unit"`
}

func loadManufacturingSeed() (*manufacturingSeed, error) {
	return loadSeedData[manufacturingSeed]("data/manufacturing.json")
}
