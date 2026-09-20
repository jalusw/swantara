package quality

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func QualityCheckFixture(opts ...func(*QualityCheck) *QualityCheck) *QualityCheck {
	now := time.Now()
	checkedAt := now.Add(-time.Duration(gofakeit.Number(1, 48)) * time.Hour)
	check := &QualityCheck{
		Base:              model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		PointID:           helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		ItemID:            helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		BatchID:           helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		ShipmentID:        helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		ProductionOrderID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		MeasuredValue:     helper.Ptr(gofakeit.Float64Range(1, 10000)),
		Result:            CheckResultPass,
		CheckedBy:         helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		CheckedAt:         &checkedAt,
	}
	for _, opt := range opts {
		opt(check)
	}
	return check
}

func QualityAlertFixture(opts ...func(*QualityAlert) *QualityAlert) *QualityAlert {
	now := time.Now()
	alert := &QualityAlert{
		Base:        model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ItemID:      helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		BatchID:     helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		CheckID:     helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Title:       helper.Ptr(gofakeit.Sentence(3)),
		Description: helper.Ptr(gofakeit.Paragraph(2, 3, 8, " ")),
		Severity:    helper.Ptr("high"),
		State:       AlertStateOpen,
		AssignedTo:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
	}
	for _, opt := range opts {
		opt(alert)
	}
	return alert
}
