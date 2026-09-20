package seeders

type CurrencySeed struct {
	Currencies []CurrencySeedItem `json:"currencies"`
}

type CurrencySeedItem struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	DecimalPlaces int16  `json:"decimal_places"`
}

func LoadCurrencySeed() (*CurrencySeed, error) {
	return loadSeedData[CurrencySeed]("data/currencies.json")
}
