package accounting

import (
	"context"
	"sort"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type ReminderLevelDAO interface {
	dao.CRUD[reference.ReminderLevel]
	ListSorted(ctx context.Context) ([]*reference.ReminderLevel, error)
}

type reminderLevelDAO struct {
	dao.Base[reference.ReminderLevel]
	db *gorm.DB
}

func NewReminderLevelDAO(db *gorm.DB) ReminderLevelDAO {
	return reminderLevelDAO{Base: dao.NewBase[reference.ReminderLevel](db), db: db}
}

func (d reminderLevelDAO) ListSorted(ctx context.Context) ([]*reference.ReminderLevel, error) {
	var entities []reference.ReminderLevel
	if err := d.db.WithContext(ctx).Order("sequence ASC").Find(&entities).Error; err != nil {
		return nil, err
	}
	levels := make([]*reference.ReminderLevel, len(entities))
	for i := range entities {
		levels[i] = &entities[i]
	}
	return levels, nil
}

type ReminderService struct {
	invoices InvoiceDAO
	levels   ReminderLevelDAO
	actions  ReminderActionDAO
	tx       db.Transactioner
	now      func() time.Time
	sender   queue.TaskEnqueuer
}

func (s ReminderService) WithSender(sender queue.TaskEnqueuer) ReminderService {
	s.sender = sender
	return s
}

func NewReminderService(
	invoices InvoiceDAO,
	levels ReminderLevelDAO,
	actions ReminderActionDAO,
	tx db.Transactioner,
) ReminderService {
	return ReminderService{invoices: invoices, levels: levels, actions: actions, tx: tx, now: time.Now}
}

type GenerateReminderRequest struct {
	OrganizationID uint64
	AsOf           *time.Time
}

func (s ReminderService) List(ctx context.Context, q *query.Query) (*query.Page[ReminderAction], error) {
	return s.actions.List(ctx, q)
}

func (s ReminderService) Generate(ctx context.Context, request GenerateReminderRequest) ([]*ReminderAction, error) {
	levels, err := s.levels.ListSorted(ctx)
	if err != nil {
		return nil, err
	}
	if len(levels) == 0 {
		return nil, ErrReminderLevelNotFound
	}

	asOf := s.now().UTC()
	if request.AsOf != nil {
		asOf = *request.AsOf
	}
	invoices, err := s.invoices.ListOverdue(ctx, &asOf)
	if err != nil {
		return nil, err
	}

	type pendingAction struct {
		action *ReminderAction
		days   int
	}
	pending := make([]pendingAction, 0, len(invoices))
	for _, invoice := range invoices {
		if invoice.OrganizationID == nil || *invoice.OrganizationID != request.OrganizationID {
			continue
		}
		if invoice.DueDate == nil {
			continue
		}
		daysOverdue := int(asOf.Sub(*invoice.DueDate).Hours() / 24)
		if daysOverdue < 0 {
			continue
		}
		level := applicableLevel(levels, daysOverdue)
		if level == nil {
			continue
		}
		existing, err := s.actions.FindByInvoiceLevel(ctx, invoice.ID, level.ID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			continue
		}
		channel := "email"
		action := &ReminderAction{
			ContactID: invoice.ContactID,
			InvoiceID: invoice.ID,
			LevelID:   level.ID,
			Channel:   &channel,
		}
		err = s.tx.Run(ctx, func(tx *gorm.DB) error {
			var err error
			action, err = s.actions.Create(ctx, action)
			return err
		})
		if err != nil {
			return nil, err
		}
		if s.sender != nil {
			task, err := tasks.NewSendReminderTask(action.ID, action.ContactID, action.InvoiceID)
			if err != nil {
				return nil, err
			}
			if _, err := s.sender.Enqueue(task); err != nil {
				_ = s.actions.Delete(ctx, action.ID)
				return nil, err
			}
		}
		pending = append(pending, pendingAction{action: action, days: daysOverdue})
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].days > pending[j].days })
	created := make([]*ReminderAction, 0, len(pending))
	for _, item := range pending {
		created = append(created, item.action)
	}
	return created, nil
}

func applicableLevel(levels []*reference.ReminderLevel, daysOverdue int) *reference.ReminderLevel {
	sorted := make([]*reference.ReminderLevel, len(levels))
	copy(sorted, levels)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].DaysOverdue < sorted[j].DaysOverdue
	})
	var selected *reference.ReminderLevel
	for _, level := range sorted {
		if daysOverdue >= level.DaysOverdue {
			selected = level
		}
	}
	return selected
}
