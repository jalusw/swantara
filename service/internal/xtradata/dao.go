package xtradata

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IntegrationEventDAO interface {
	dao.CRUD[IntegrationEvent]
}

type integrationEventDAO struct {
	dao.Base[IntegrationEvent]
}

func NewIntegrationEventDAO(db *gorm.DB) IntegrationEventDAO {
	return integrationEventDAO{Base: dao.NewBase[IntegrationEvent](db)}
}

type WebhookSubscriptionDAO interface {
	dao.CRUD[WebhookSubscription]
	ListEnabledForOrg(ctx context.Context, organizationID *uint64) ([]*WebhookSubscription, error)
}

type webhookSubscriptionDAO struct {
	dao.Base[WebhookSubscription]
	db *gorm.DB
}

func NewWebhookSubscriptionDAO(db *gorm.DB) WebhookSubscriptionDAO {
	return webhookSubscriptionDAO{Base: dao.NewBase[WebhookSubscription](db), db: db}
}

func (d webhookSubscriptionDAO) ListEnabledForOrg(ctx context.Context, organizationID *uint64) ([]*WebhookSubscription, error) {
	var subscriptions []WebhookSubscription
	q := d.db.WithContext(ctx).Where("enabled = ?", true)
	if organizationID == nil {
		q = q.Where("organization_id IS NULL")
	} else {
		q = q.Where("organization_id = ? OR organization_id IS NULL", *organizationID)
	}
	if err := q.Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	items := make([]*WebhookSubscription, len(subscriptions))
	for i := range subscriptions {
		items[i] = &subscriptions[i]
	}
	return items, nil
}

type WebhookDeliveryDAO interface {
	dao.CRUD[WebhookDelivery]
	DeliveredForEvent(ctx context.Context, eventID uint64) ([]*WebhookDelivery, error)
}

type webhookDeliveryDAO struct {
	dao.Base[WebhookDelivery]
	db *gorm.DB
}

func NewWebhookDeliveryDAO(db *gorm.DB) WebhookDeliveryDAO {
	return webhookDeliveryDAO{Base: dao.NewBase[WebhookDelivery](db), db: db}
}

func (d webhookDeliveryDAO) DeliveredForEvent(ctx context.Context, eventID uint64) ([]*WebhookDelivery, error) {
	var deliveries []WebhookDelivery
	if err := d.db.WithContext(ctx).Where("event_id = ? AND status = ?", eventID, WebhookDeliveryStatusDelivered).Find(&deliveries).Error; err != nil {
		return nil, err
	}
	items := make([]*WebhookDelivery, len(deliveries))
	for i := range deliveries {
		items[i] = &deliveries[i]
	}
	return items, nil
}

type SystemConfigDAO interface {
	dao.CRUD[reference.SystemConfig]
}

type systemConfigDAO struct {
	dao.Base[reference.SystemConfig]
}

func NewSystemConfigDAO(db *gorm.DB) SystemConfigDAO {
	return systemConfigDAO{Base: dao.NewBase[reference.SystemConfig](db)}
}

type IdempotencyKeyDAO interface {
	dao.CRUD[IdempotencyKey]
	Claim(ctx context.Context, organizationID *uint64, key, resource string) (bool, *IdempotencyKey, error)
	Complete(ctx context.Context, organizationID *uint64, key string, statusCode int, response []byte) error
	Release(ctx context.Context, organizationID *uint64, key string) error
}

type idempotencyKeyDAO struct {
	dao.Base[IdempotencyKey]
	db *gorm.DB
}

func NewIdempotencyKeyDAO(db *gorm.DB) IdempotencyKeyDAO {
	return idempotencyKeyDAO{Base: dao.NewBase[IdempotencyKey](db), db: db}
}

func (d idempotencyKeyDAO) Claim(ctx context.Context, organizationID *uint64, key, resource string) (bool, *IdempotencyKey, error) {
	var claimed bool
	var existing *IdempotencyKey
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := &IdempotencyKey{OrganizationID: organizationID, Key: key, Resource: resource}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(row)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 1 {
			claimed = true
			return nil
		}
		var found IdempotencyKey
		q := tx.Where("key = ?", key)
		if organizationID == nil {
			q = q.Where("organization_id IS NULL")
		} else {
			q = q.Where("organization_id = ?", *organizationID)
		}
		if err := q.Take(&found).Error; err != nil {
			return err
		}
		existing = &found
		return nil
	})
	if err != nil {
		return false, nil, err
	}
	return claimed, existing, nil
}

func (d idempotencyKeyDAO) Complete(ctx context.Context, organizationID *uint64, key string, statusCode int, response []byte) error {
	q := d.db.WithContext(ctx).Model(&IdempotencyKey{}).
		Where("key = ? AND status_code IS NULL AND response IS NULL", key)
	if organizationID == nil {
		q = q.Where("organization_id IS NULL")
	} else {
		q = q.Where("organization_id = ?", *organizationID)
	}
	return q.Updates(map[string]any{"status_code": statusCode, "response": response}).Error
}

func (d idempotencyKeyDAO) Release(ctx context.Context, organizationID *uint64, key string) error {
	q := d.db.WithContext(ctx).Unscoped().Where("key = ?", key)
	if organizationID == nil {
		q = q.Where("organization_id IS NULL")
	} else {
		q = q.Where("organization_id = ?", *organizationID)
	}
	return q.Delete(&IdempotencyKey{}).Error
}
