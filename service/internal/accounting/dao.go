package accounting

import (
	"context"
	"errors"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type JournalEntryDAO interface {
	dao.CRUD[JournalEntry]
	CreateWithLines(ctx context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error)
	CreateWithLinesTx(ctx context.Context, tx *gorm.DB, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error)
	FindByReversed(ctx context.Context, moveID uint64) (*JournalEntry, error)
}

type journalEntryDAO struct {
	dao.Base[JournalEntry]
	db *gorm.DB
}

func NewJournalEntryDAO(db *gorm.DB) JournalEntryDAO {
	return journalEntryDAO{Base: dao.NewBase[JournalEntry](db), db: db}
}

func (d journalEntryDAO) Update(ctx context.Context, entity *JournalEntry) (*JournalEntry, error) {
	if entity.IsPosted() {
		return nil, ErrEntryImmutable
	}
	return d.Base.Update(ctx, entity)
}

func (d journalEntryDAO) Delete(ctx context.Context, id uint64) error {
	entity, err := d.Find(ctx, id)
	if err != nil {
		return err
	}
	if entity != nil && entity.IsPosted() {
		return ErrEntryImmutable
	}
	return d.Base.Delete(ctx, id)
}

func (d journalEntryDAO) CreateWithLines(ctx context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		entry, err = d.CreateWithLinesTx(ctx, tx, entry, lines)
		return err
	})
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (d journalEntryDAO) CreateWithLinesTx(ctx context.Context, tx *gorm.DB, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
	if err := tx.WithContext(ctx).Create(entry).Error; err != nil {
		return nil, err
	}
	for _, line := range lines {
		line.EntryID = entry.ID
		line.Date = entry.Date
	}
	if err := tx.WithContext(ctx).CreateInBatches(lines, 500).Error; err != nil {
		return nil, err
	}
	return entry, nil
}

func (d journalEntryDAO) FindByReversed(ctx context.Context, moveID uint64) (*JournalEntry, error) {
	return d.Search(ctx, "reversed_entry_id", moveID)
}

type JournalLineDAO interface {
	dao.CRUD[JournalLine]
	ListByMovement(ctx context.Context, moveID uint64) ([]*JournalLine, error)
	ListUnreconciledByAccount(ctx context.Context, accountID uint64) ([]*JournalLine, error)
	BalanceByOriginAndAccount(ctx context.Context, originType string, originID, accountID uint64) (float64, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, line *JournalLine) (*JournalLine, error)
}

type journalEntryLineDAO struct {
	dao.Base[JournalLine]
	db *gorm.DB
}

func NewJournalLineDAO(db *gorm.DB) JournalLineDAO {
	return journalEntryLineDAO{Base: dao.NewBase[JournalLine](db), db: db}
}

func (d journalEntryLineDAO) ListByMovement(ctx context.Context, moveID uint64) ([]*JournalLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "entry_id", Operator: query.Equal, Value: moveID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d journalEntryLineDAO) ListUnreconciledByAccount(ctx context.Context, accountID uint64) ([]*JournalLine, error) {
	var entities []JournalLine
	if err := d.db.WithContext(ctx).Where("account_id = ? AND reconciled = ?", accountID, false).Find(&entities).Error; err != nil {
		return nil, err
	}
	items := make([]*JournalLine, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}

func (d journalEntryLineDAO) BalanceByOriginAndAccount(ctx context.Context, originType string, originID, accountID uint64) (float64, error) {
	var balance float64
	err := d.db.WithContext(ctx).
		Model(&JournalLine{}).
		Joins("JOIN journal_entrys ON journal_entrys.id = journal_lines.entry_id").
		Where("journal_entrys.origin_type = ? AND journal_entrys.origin_id = ? AND journal_lines.account_id = ?", originType, originID, accountID).
		Select("COALESCE(SUM(debit - credit), 0)").
		Scan(&balance).Error
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (d journalEntryLineDAO) Update(ctx context.Context, line *JournalLine) (*JournalLine, error) {
	if err := d.assertLineMutable(ctx, d.db, line); err != nil {
		return nil, err
	}
	return d.Base.Update(ctx, line)
}

func (d journalEntryLineDAO) UpdateTx(ctx context.Context, tx *gorm.DB, line *JournalLine) (*JournalLine, error) {
	if err := d.assertLineMutable(ctx, tx, line); err != nil {
		return nil, err
	}
	if err := tx.WithContext(ctx).Save(line).Error; err != nil {
		return nil, err
	}
	return line, nil
}

func (d journalEntryLineDAO) assertLineMutable(ctx context.Context, db *gorm.DB, line *JournalLine) error {
	var existing JournalLine
	if err := db.WithContext(ctx).Where("id = ?", line.ID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	var entry JournalEntry
	if err := db.WithContext(ctx).Where("id = ?", existing.EntryID).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if !entry.IsPosted() {
		return nil
	}
	if !existing.Debit.Equal(line.Debit) || !existing.Credit.Equal(line.Credit) || existing.AccountID != line.AccountID || existing.EntryID != line.EntryID {
		return ErrEntryImmutable
	}
	return nil
}
