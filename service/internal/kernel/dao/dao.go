package dao

import (
	"context"
	"errors"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type CRUD[E any] interface {
	List(ctx context.Context, query *query.Query) (*query.Page[E], error)
	Search(ctx context.Context, field string, value any) (*E, error)
	Find(ctx context.Context, id uint64) (*E, error)
	Create(ctx context.Context, entity *E) (*E, error)
	Update(ctx context.Context, entity *E) (*E, error)
	Delete(ctx context.Context, id uint64) error
	HardDelete(ctx context.Context, id uint64) error
}

type Base[E any] struct {
	db *gorm.DB
}

func NewBase[E any](db *gorm.DB) Base[E] {
	return Base[E]{db: db}
}

func (r Base[E]) List(ctx context.Context, q *query.Query) (*query.Page[E], error) {
	var count int64
	var entities []E

	countTx := r.db.WithContext(ctx).Model(&entities)
	if q != nil {
		countTx = query.ApplyFilters(countTx, q.Filters)
	}
	if err := countTx.Count(&count).Error; err != nil {
		return nil, err
	}

	tx := r.db.WithContext(ctx).Model(&entities)
	if q != nil {
		tx = query.ApplyQuery(tx, q)
	}
	if err := tx.Find(&entities).Error; err != nil {
		return nil, err
	}

	items := make([]*E, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}

	return &query.Page[E]{Items: items, Count: count}, nil
}

func (r Base[E]) Search(ctx context.Context, field string, value any) (*E, error) {
	var entity E
	tx := query.ApplyFilters(r.db.WithContext(ctx).Model(&entity), []query.Filter{
		{Field: field, Operator: query.Equal, Value: value},
	})
	if err := tx.Take(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (r Base[E]) Find(ctx context.Context, id uint64) (*E, error) {
	var entity E
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (r Base[E]) Create(ctx context.Context, entity *E) (*E, error) {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (r Base[E]) Update(ctx context.Context, entity *E) (*E, error) {
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (r Base[E]) Delete(ctx context.Context, id uint64) error {
	var entity E
	return r.db.WithContext(ctx).Delete(&entity, id).Error
}

func (r Base[E]) HardDelete(ctx context.Context, id uint64) error {
	var entity E
	return r.db.WithContext(ctx).Unscoped().Delete(&entity, id).Error
}
