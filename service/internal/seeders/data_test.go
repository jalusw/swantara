package seeders

import "testing"

func TestSeedDataLoaders(t *testing.T) {
	loaders := []struct {
		name string
		load func() error
	}{
		{"iam", func() error { _, err := loadIamSeed(); return err }},
		{"accounting", func() error { _, err := loadAccountingSeed(); return err }},
		{"inventory", func() error { _, err := loadInventorySeed(); return err }},
		{"manufacturing", func() error { _, err := loadManufacturingSeed(); return err }},
		{"hr", func() error { _, err := loadHRSeed(); return err }},
		{"crm", func() error { _, err := loadCRMSeed(); return err }},
		{"pos", func() error { _, err := loadPOSSeed(); return err }},
		{"subscription", func() error { _, err := loadSubscriptionSeed(); return err }},
		{"system", func() error { _, err := loadSystemSeed(); return err }},
	}

	for _, tt := range loaders {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.load(); err != nil {
				t.Fatalf("seed data failed to load: %v", err)
			}
		})
	}
}

func TestIamSeedData(t *testing.T) {
	seed, err := loadIamSeed()
	if err != nil {
		t.Fatal(err)
	}

	if len(seed.permissions) != 312 {
		t.Errorf("permissions: got %d, want 312", len(seed.permissions))
	}
}
