package audit

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

var (
	ErrImmutable      = errors.New("record is immutable once posted")
	ErrRecordNotFound = errors.New("record not found")
	ErrNotAuditable   = errors.New("entity does not expose an id")
)

type AuditLogService struct {
	logs LogDAO
	now  func() time.Time
}

func NewAuditLogService(logs LogDAO) AuditLogService {
	return AuditLogService{logs: logs, now: time.Now}
}

func (s AuditLogService) Append(ctx context.Context, entry *Log) error {
	if entry.ChangedAt.IsZero() {
		entry.ChangedAt = s.now().UTC()
	}
	if entry.ChangedBy == 0 {
		entry.ChangedBy = model.ActorID(ctx)
	}
	_, err := s.logs.Create(ctx, entry)
	return err
}

func Record(ctx context.Context, recorder Recorder, tableName string, recordID uint64, action Action, diff map[string]any) error {
	data, err := json.Marshal(diff)
	if err != nil {
		return err
	}
	return recorder.Append(ctx, &Log{
		EntityTable: tableName,
		RecordID:    recordID,
		Action:      action,
		Diff:        data,
	})
}

type idProvider interface {
	GetID() uint64
}

type postedEntity interface {
	IsPosted() bool
}

type Audited[E any] struct {
	dao.CRUD[E]
	recorder Recorder
	table    string
}

func NewAudited[E any](inner dao.CRUD[E], recorder Recorder, table string) Audited[E] {
	return Audited[E]{CRUD: inner, recorder: recorder, table: table}
}

func (a Audited[E]) Create(ctx context.Context, entity *E) (*E, error) {
	created, err := a.CRUD.Create(ctx, entity)
	if err != nil {
		return nil, err
	}
	if id, ok := idOf(created); ok {
		err = a.append(ctx, id, ActionInsert, Diff(nil, created))
		if err != nil {
			return nil, err
		}
	}
	return created, nil
}

func (a Audited[E]) Update(ctx context.Context, entity *E) (*E, error) {
	id, ok := idOf(entity)
	if !ok {
		return nil, ErrNotAuditable
	}
	before, err := a.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if before == nil {
		return nil, ErrRecordNotFound
	}
	if posted, ok := any(before).(postedEntity); ok && posted.IsPosted() {
		return nil, ErrImmutable
	}

	updated, err := a.CRUD.Update(ctx, entity)
	if err != nil {
		return nil, err
	}
	if err := a.append(ctx, id, ActionUpdate, Diff(before, updated)); err != nil {
		return nil, err
	}
	return updated, nil
}

func (a Audited[E]) Delete(ctx context.Context, id uint64) error {
	before, err := a.Find(ctx, id)
	if err != nil {
		return err
	}
	if before == nil {
		return nil
	}
	if posted, ok := any(before).(postedEntity); ok && posted.IsPosted() {
		return ErrImmutable
	}

	if err := a.CRUD.Delete(ctx, id); err != nil {
		return err
	}
	return a.append(ctx, id, ActionDelete, Diff(before, nil))
}

func (a Audited[E]) append(ctx context.Context, recordID uint64, action Action, diff map[string]any) error {
	return Record(ctx, a.recorder, a.table, recordID, action, diff)
}

func idOf[E any](entity *E) (uint64, bool) {
	provider, ok := any(entity).(idProvider)
	if !ok {
		return 0, false
	}
	return provider.GetID(), true
}
