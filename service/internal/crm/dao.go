package crm

import (
	"context"
	"errors"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type ProspectDAO interface {
	dao.CRUD[Prospect]
	ListOpen(ctx context.Context, organizationID *uint64) ([]*Prospect, error)
	CountWon(ctx context.Context, organizationID *uint64) (int64, error)
	CountLost(ctx context.Context, organizationID *uint64) (int64, error)
}

type crmLeadDAO struct {
	dao.Base[Prospect]
	db *gorm.DB
}

func NewProspectDAO(db *gorm.DB) ProspectDAO {
	return crmLeadDAO{Base: dao.NewBase[Prospect](db), db: db}
}

func (d crmLeadDAO) ListOpen(ctx context.Context, organizationID *uint64) ([]*Prospect, error) {
	var leads []Prospect
	tx := d.db.WithContext(ctx).
		Where("type = ?", ProspectKindOpportunity).
		Where("closed_at IS NULL")
	if organizationID != nil {
		tx = tx.Where("organization_id = ?", *organizationID)
	}
	err := tx.Find(&leads).Error
	if err != nil {
		return nil, err
	}
	items := make([]*Prospect, len(leads))
	for i := range leads {
		items[i] = &leads[i]
	}
	return items, nil
}

func (d crmLeadDAO) CountWon(ctx context.Context, organizationID *uint64) (int64, error) {
	var count int64
	tx := d.db.WithContext(ctx).Model(&Prospect{}).
		Joins("JOIN pipeline_stages ON pipeline_stages.id = prospects.stage_id").
		Where("prospects.type = ?", ProspectKindOpportunity).
		Where("pipeline_stages.is_won = true")
	if organizationID != nil {
		tx = tx.Where("prospects.organization_id = ?", *organizationID)
	}
	err := tx.Count(&count).Error
	return count, err
}

func (d crmLeadDAO) CountLost(ctx context.Context, organizationID *uint64) (int64, error) {
	var count int64
	tx := d.db.WithContext(ctx).Model(&Prospect{}).
		Where("type = ?", ProspectKindOpportunity).
		Where("lost_reason IS NOT NULL")
	if organizationID != nil {
		tx = tx.Where("organization_id = ?", *organizationID)
	}
	err := tx.Count(&count).Error
	return count, err
}

type ProspectActivityDAO interface {
	dao.CRUD[ProspectActivity]
	ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[ProspectActivity], error)
	FindInOrg(ctx context.Context, id, organizationID uint64) (*ProspectActivity, error)
}

type crmActivityDAO struct {
	dao.Base[ProspectActivity]
	db *gorm.DB
}

func NewProspectActivityDAO(db *gorm.DB) ProspectActivityDAO {
	return crmActivityDAO{Base: dao.NewBase[ProspectActivity](db), db: db}
}

func (d crmActivityDAO) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[ProspectActivity], error) {
	var count int64
	var entities []ProspectActivity

	join := "LEFT JOIN prospects ON prospects.id = prospect_activities.lead_id " +
		"LEFT JOIN contacts ON contacts.id = prospect_activities.contact_id " +
		"WHERE (prospect_activities.lead_id IS NOT NULL AND prospects.organization_id = ?) " +
		"OR (prospect_activities.lead_id IS NULL AND prospect_activities.contact_id IS NOT NULL AND contacts.organization_id = ?)"

	countTx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID, organizationID)
	if q != nil {
		countTx = query.ApplyFilters(countTx, q.Filters)
	}
	if err := countTx.Count(&count).Error; err != nil {
		return nil, err
	}

	tx := d.db.WithContext(ctx).Model(&entities).Joins(join, organizationID, organizationID)
	if q != nil {
		tx = query.ApplyQuery(tx, q)
	}
	if err := tx.Find(&entities).Error; err != nil {
		return nil, err
	}

	items := make([]*ProspectActivity, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}

	return &query.Page[ProspectActivity]{Items: items, Count: count}, nil
}

func (d crmActivityDAO) FindInOrg(ctx context.Context, id, organizationID uint64) (*ProspectActivity, error) {
	var activity ProspectActivity
	err := d.db.WithContext(ctx).Model(&ProspectActivity{}).
		Joins("LEFT JOIN prospects ON prospects.id = prospect_activities.lead_id "+
			"LEFT JOIN contacts ON contacts.id = prospect_activities.contact_id").
		Where("prospect_activities.id = ?", id).
		Where("(prospect_activities.lead_id IS NOT NULL AND prospects.organization_id = ?) "+
			"OR (prospect_activities.lead_id IS NULL AND prospect_activities.contact_id IS NOT NULL AND contacts.organization_id = ?)", organizationID, organizationID).
		Take(&activity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &activity, nil
}
