package reporting

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestConfigSource_JournalID_ReturnsConfiguredValue(t *testing.T) {
	ctx := context.Background()
	source := NewConfigSource(ConfigLookupMock{
		ListFn: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configReportingJournalID, Value: json.RawMessage(`5`)},
			}}, nil
		},
	})

	journalID, err := source.JournalID(ctx, 1)
	if err != nil {
		t.Fatalf("journal id failed: %v", err)
	}
	if journalID != 5 {
		t.Errorf("journal id = %d, want 5", journalID)
	}
}

func TestConfigSource_FxGainLossAccountID_ReturnsConfiguredValue(t *testing.T) {
	ctx := context.Background()
	source := NewConfigSource(ConfigLookupMock{
		ListFn: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configReportingFxGainLossAcct, Value: json.RawMessage(`9`)},
			}}, nil
		},
	})

	accountID, err := source.FxGainLossAccountID(ctx, 1)
	if err != nil {
		t.Fatalf("fx gain loss account id failed: %v", err)
	}
	if accountID != 9 {
		t.Errorf("account id = %d, want 9", accountID)
	}
}

func TestConfigSource_Read_PropagatesLookupError(t *testing.T) {
	ctx := context.Background()
	source := NewConfigSource(ConfigLookupMock{
		ListFn: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return nil, errors.New("lookup failed")
		},
	})

	if helper.AssertError(t, mustConfigValue(t, source, ctx), true, nil) {
		return
	}
}

func TestConfigSource_Read_IgnoresOtherKeys(t *testing.T) {
	ctx := context.Background()
	source := NewConfigSource(ConfigLookupMock{
		ListFn: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: "reporting.other", Value: json.RawMessage(`5`)},
			}}, nil
		},
	})

	if err := mustConfigValue(t, source, ctx); !errors.Is(err, ErrConfigMissing) {
		t.Errorf("err = %v, want ErrConfigMissing", err)
	}
}

func TestConfigSource_Read_IgnoresOtherOrganizations(t *testing.T) {
	ctx := context.Background()
	source := NewConfigSource(ConfigLookupMock{
		ListFn: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configReportingJournalID, OrganizationID: helper.Ptr(uint64(2)), Value: json.RawMessage(`5`)},
			}}, nil
		},
	})

	if err := mustConfigValue(t, source, ctx); !errors.Is(err, ErrConfigMissing) {
		t.Errorf("err = %v, want ErrConfigMissing", err)
	}
}

func TestConfigSource_Read_MissingValue(t *testing.T) {
	ctx := context.Background()
	source := NewConfigSource(ConfigLookupMock{
		ListFn: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configReportingJournalID, Value: json.RawMessage(``)},
			}}, nil
		},
	})

	if err := mustConfigValue(t, source, ctx); !errors.Is(err, ErrConfigMissing) {
		t.Errorf("err = %v, want ErrConfigMissing", err)
	}
}

func TestConfigSource_Read_InvalidValue(t *testing.T) {
	ctx := context.Background()
	source := NewConfigSource(ConfigLookupMock{
		ListFn: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configReportingJournalID, Value: json.RawMessage(`not-a-number`)},
			}}, nil
		},
	})

	if err := mustConfigValue(t, source, ctx); err == nil {
		t.Error("err = nil, want unmarshal error")
	}
}

func TestConfigSource_Read_ZeroValue(t *testing.T) {
	ctx := context.Background()
	source := NewConfigSource(ConfigLookupMock{
		ListFn: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
				{Key: configReportingJournalID, Value: json.RawMessage(`0`)},
			}}, nil
		},
	})

	if err := mustConfigValue(t, source, ctx); !errors.Is(err, ErrConfigMissing) {
		t.Errorf("err = %v, want ErrConfigMissing", err)
	}
}

func TestConfigSource_Read_NoMatch(t *testing.T) {
	ctx := context.Background()
	source := NewConfigSource(ConfigLookupMock{
		ListFn: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{}, nil
		},
	})

	if err := mustConfigValue(t, source, ctx); !errors.Is(err, ErrConfigMissing) {
		t.Errorf("err = %v, want ErrConfigMissing", err)
	}
}

func mustConfigValue(t *testing.T, source ConfigSource, ctx context.Context) error {
	t.Helper()
	_, err := source.JournalID(ctx, 1)
	return err
}
