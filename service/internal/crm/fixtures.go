package crm

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func ProspectFixture(opts ...func(*Prospect) *Prospect) *Prospect {
	now := time.Now()
	prospect := &Prospect{
		Base:            model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Name:            gofakeit.AppName(),
		Type:            ProspectKindLead,
		ExpectedRevenue: gofakeit.Float64Range(1, 10000),
		Probability:     gofakeit.Float64Range(1, 100),
		Priority:        int16(gofakeit.Number(0, 3)),
	}
	for _, opt := range opts {
		opt(prospect)
	}
	return prospect
}

func ProspectActivityFixture(opts ...func(*ProspectActivity) *ProspectActivity) *ProspectActivity {
	now := time.Now()
	activity := &ProspectActivity{
		Base:    model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		Type:    ProspectActivityTypeNote,
		Summary: gofakeit.Sentence(5),
	}
	for _, opt := range opts {
		opt(activity)
	}
	return activity
}
