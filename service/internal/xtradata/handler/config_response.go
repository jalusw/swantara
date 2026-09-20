package handler

import (
	"encoding/json"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type SystemConfigResponse struct {
	ID             uint64          `json:"id"`
	OrganizationID *uint64         `json:"organization_id"`
	Key            string          `json:"key"`
	Value          json.RawMessage `json:"value" swaggertype:"object"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func newSystemConfigResponse(config *reference.SystemConfig) SystemConfigResponse {
	return SystemConfigResponse{
		ID:             config.ID,
		OrganizationID: config.OrganizationID,
		Key:            config.Key,
		Value:          config.Value,
		CreatedAt:      config.CreatedAt,
		UpdatedAt:      config.UpdatedAt,
	}
}

var systemConfigQueryAllowlist = map[string]struct{}{
	"key":             {},
	"organization_id": {},
	"created_at":      {},
	"updated_at":      {},
}
