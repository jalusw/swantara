package seeders

type subscriptionSeed struct {
	SubscriptionPlans []subscriptionPlanSeed `json:"subscription_plans"`
}

type subscriptionPlanSeed struct {
	Name     string `json:"name"`
	Interval string `json:"interval"`
	Count    int    `json:"count"`
}

func loadSubscriptionSeed() (*subscriptionSeed, error) {
	return loadSeedData[subscriptionSeed]("data/subscription.json")
}
