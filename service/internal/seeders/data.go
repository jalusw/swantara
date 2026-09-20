package seeders

import (
	"embed"
	"encoding/json"
)

//go:embed data/*.json
var seedDataFS embed.FS

func loadSeedData[T any](path string) (*T, error) {
	raw, err := seedDataFS.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
