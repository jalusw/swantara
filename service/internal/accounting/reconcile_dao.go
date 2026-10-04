package accounting

import (
	"context"
	"errors"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type AccountFullReconcileDAO interface {
	dao.CRUD[AccountFullReconcile]
	CreateTx(ctx context.Context, tx *gorm.DB, reconcile *AccountFullReconcile) (*AccountFullReconcile, error)
}

type accountFullReconcileDAO struct {
	dao.Base[AccountFullReconcile]
	db *gorm.DB
}

func NewAccountFullReconcileDAO(db *gorm.DB) AccountFullReconcileDAO {
	return accountFullReconcileDAO{Base: dao.NewBase[AccountFullReconcile](db), db: db}
}

func (d accountFullReconcileDAO) CreateTx(ctx context.Context, tx *gorm.DB, reconcile *AccountFullReconcile) (*AccountFullReconcile, error) {
	if err := tx.WithContext(ctx).Create(reconcile).Error; err != nil {
		return nil, err
	}
	return reconcile, nil
}

type AccountPartialReconcileDAO interface {
	dao.CRUD[AccountPartialReconcile]
	CreateTx(ctx context.Context, tx *gorm.DB, reconcile *AccountPartialReconcile) (*AccountPartialReconcile, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, reconcile *AccountPartialReconcile) (*AccountPartialReconcile, error)
	TotalByLine(ctx context.Context, lineID uint64) (float64, error)
}

type accountPartialReconcileDAO struct {
	dao.Base[AccountPartialReconcile]
	db *gorm.DB
}

func NewAccountPartialReconcileDAO(db *gorm.DB) AccountPartialReconcileDAO {
	return accountPartialReconcileDAO{Base: dao.NewBase[AccountPartialReconcile](db), db: db}
}

func (d accountPartialReconcileDAO) CreateTx(ctx context.Context, tx *gorm.DB, reconcile *AccountPartialReconcile) (*AccountPartialReconcile, error) {
	if err := tx.WithContext(ctx).Create(reconcile).Error; err != nil {
		return nil, err
	}
	return reconcile, nil
}

func (d accountPartialReconcileDAO) UpdateTx(ctx context.Context, tx *gorm.DB, reconcile *AccountPartialReconcile) (*AccountPartialReconcile, error) {
	if err := tx.WithContext(ctx).Save(reconcile).Error; err != nil {
		return nil, err
	}
	return reconcile, nil
}

func (d accountPartialReconcileDAO) TotalByLine(ctx context.Context, lineID uint64) (float64, error) {
	var total float64
	if err := d.db.WithContext(ctx).
		Model(&AccountPartialReconcile{}).
		Where("debit_line_id = ? OR credit_line_id = ?", lineID, lineID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

type ReminderActionDAO interface {
	dao.CRUD[ReminderAction]
	CreateTx(ctx context.Context, tx *gorm.DB, action *ReminderAction) (*ReminderAction, error)
	FindByInvoiceLevel(ctx context.Context, invoiceID, levelID uint64) (*ReminderAction, error)
	ListByInvoice(ctx context.Context, invoiceID uint64) ([]*ReminderAction, error)
}

type reminderActionDAO struct {
	dao.Base[ReminderAction]
	db *gorm.DB
}

func NewReminderActionDAO(db *gorm.DB) ReminderActionDAO {
	return reminderActionDAO{Base: dao.NewBase[ReminderAction](db), db: db}
}

func (d reminderActionDAO) CreateTx(ctx context.Context, tx *gorm.DB, action *ReminderAction) (*ReminderAction, error) {
	if err := tx.WithContext(ctx).Create(action).Error; err != nil {
		return nil, err
	}
	return action, nil
}

func (d reminderActionDAO) FindByInvoiceLevel(ctx context.Context, invoiceID, levelID uint64) (*ReminderAction, error) {
	var entity ReminderAction
	err := d.db.WithContext(ctx).Where("invoice_id = ? AND level_id = ?", invoiceID, levelID).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (d reminderActionDAO) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*ReminderAction, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "invoice_id", Operator: query.Equal, Value: invoiceID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
