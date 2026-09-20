package seeders

type inventorySeed struct {
	UnitCategories    []string            `json:"unit_groups"`
	Units             []unitSeed          `json:"units"`
	ProductCategories []itemCategorySeed  `json:"item_categories"`
	Items             []itemSeed          `json:"items"`
	ItemAttributes    []itemAttributeSeed `json:"item_attributes"`
	Carriers          []string            `json:"carriers"`
	Warehouses        []warehouseSeed     `json:"warehouses"`
	StockLocations    []stockLocationSeed `json:"stock_locations"`
}

type unitSeed struct {
	Name     string  `json:"name"`
	Factor   float64 `json:"factor"`
	Type     string  `json:"type"`
	Rounding float64 `json:"rounding"`
	Category string  `json:"category"`
}

type itemCategorySeed struct {
	Name                  string `json:"name"`
	IncomeAccount         string `json:"income_account"`
	ExpenseAccount        string `json:"expense_account"`
	CogsAccount           string `json:"cogs_account"`
	StockValuationAccount string `json:"stock_cost_account"`
	StockInputAccount     string `json:"stock_input_account"`
	StockOutputAccount    string `json:"stock_output_account"`
	CostMethod            string `json:"cost_method"`
	Valuation             string `json:"valuation"`
}

type itemAttributeSeed struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type itemSeed struct {
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	Type          string  `json:"type"`
	ListPrice     float64 `json:"list_price"`
	StandardCost  float64 `json:"standard_cost"`
	IsPurchasable *bool   `json:"is_purchasable"`
	IsSellable    *bool   `json:"is_sellable"`
	SKU           string  `json:"sku"`
}

type warehouseSeed struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type stockLocationSeed struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Usage    string `json:"usage"`
	Parent   string `json:"parent"`
	Physical bool   `json:"physical"`
}

func loadInventorySeed() (*inventorySeed, error) {
	return loadSeedData[inventorySeed]("data/inventory.json")
}
