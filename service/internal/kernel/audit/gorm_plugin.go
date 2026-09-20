package audit

import (
	"context"
	"log/slog"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	beforeStateKey = "swantara:audit:before"
	auditSavepoint = "swantara_audit"
)

var defaultSkipTables = map[string]struct{}{
	"audit_logs":            {},
	"goose_migrations":      {},
	"idempotency_keys":      {},
	"user_sessions":         {},
	"pos_sessions":          {},
	"integration_events":    {},
	"webhook_deliveries":    {},
	"webhook_subscriptions": {},
	"job_runs":              {},
	"kpi_account_summaries": {},
	"doc_sequences":         {},
	"consolidation_runs":    {},
}

type GormPlugin struct {
	logger *slog.Logger
	skip   map[string]struct{}
}

func NewGormPlugin() *GormPlugin {
	return &GormPlugin{logger: slog.Default(), skip: defaultSkipTables}
}

func (p *GormPlugin) Name() string {
	return "swantara:audit"
}

func (p *GormPlugin) Initialize(db *gorm.DB) error {
	if err := db.Callback().Create().After("gorm:create").Register("swantara:audit:create", p.afterCreate); err != nil {
		return err
	}
	if err := db.Callback().Update().Before("gorm:update").Register("swantara:audit:update:before", p.beforeUpdate); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register("swantara:audit:update:after", p.afterUpdate); err != nil {
		return err
	}
	if err := db.Callback().Delete().Before("gorm:delete").Register("swantara:audit:delete:before", p.beforeDelete); err != nil {
		return err
	}
	return db.Callback().Delete().After("gorm:delete").Register("swantara:audit:delete:after", p.afterDelete)
}

func (p *GormPlugin) afterCreate(db *gorm.DB) {
	stmt := db.Statement
	if stmt.Schema == nil || p.skipped(stmt.Schema.Table) {
		return
	}
	for _, entity := range rowsOf(stmt.ReflectValue) {
		id, ok := p.idOf(entity)
		if !ok {
			continue
		}
		p.record(stmt.Context, db, stmt.Schema.Table, id, ActionInsert, Diff(nil, entity))
	}
}

func (p *GormPlugin) beforeUpdate(db *gorm.DB) {
	stmt := db.Statement
	if stmt.Schema == nil || p.skipped(stmt.Schema.Table) {
		return
	}
	rows, ok := p.snapshot(db)
	if !ok {
		return
	}
	stmt.Settings.Store(beforeStateKey, rows)
}

func (p *GormPlugin) afterUpdate(db *gorm.DB) {
	stmt := db.Statement
	if stmt.Schema == nil || p.skipped(stmt.Schema.Table) {
		return
	}
	value, ok := stmt.Settings.Load(beforeStateKey)
	if !ok {
		return
	}
	before, _ := value.([]any)
	after, ok := p.snapshot(db)
	if !ok {
		return
	}
	afterByID := map[uint64]any{}
	for _, row := range after {
		if id, ok := p.idOf(row); ok {
			afterByID[id] = row
		}
	}
	for _, row := range before {
		id, ok := p.idOf(row)
		if !ok {
			continue
		}
		next, ok := afterByID[id]
		if !ok {
			continue
		}
		diff := Diff(row, next)
		if len(diff) == 0 {
			continue
		}
		p.record(stmt.Context, db, stmt.Schema.Table, id, ActionUpdate, diff)
	}
}

func (p *GormPlugin) beforeDelete(db *gorm.DB) {
	stmt := db.Statement
	if stmt.Schema == nil || p.skipped(stmt.Schema.Table) {
		return
	}
	rows, ok := p.snapshot(db)
	if !ok {
		return
	}
	stmt.Settings.Store(beforeStateKey, rows)
}

func (p *GormPlugin) afterDelete(db *gorm.DB) {
	stmt := db.Statement
	if stmt.Schema == nil || p.skipped(stmt.Schema.Table) {
		return
	}
	value, ok := stmt.Settings.Load(beforeStateKey)
	if !ok {
		return
	}
	before, _ := value.([]any)
	for _, row := range before {
		id, ok := p.idOf(row)
		if !ok {
			continue
		}
		p.record(stmt.Context, db, stmt.Schema.Table, id, ActionDelete, Diff(row, nil))
	}
}

func (p *GormPlugin) snapshot(db *gorm.DB) ([]any, bool) {
	stmt := db.Statement
	if stmt.Schema == nil {
		return nil, false
	}
	table := stmt.Schema.Table
	session := db.Session(&gorm.Session{NewDB: true, SkipHooks: true})
	tx := session.Table(table).Unscoped()
	if where, ok := stmt.Clauses["WHERE"]; ok {
		if w, ok := where.Expression.(clause.Where); ok && len(w.Exprs) > 0 {
			tx = tx.Clauses(clause.Where{Exprs: w.Exprs})
		}
	}
	if pk := p.pkFromReflect(stmt); pk != nil {
		tx = tx.Where(clause.Eq{Column: stmt.Schema.PrioritizedPrimaryField.DBName, Value: pk})
	}
	rows := reflect.New(reflect.SliceOf(stmt.Schema.ModelType)).Interface()
	if err := tx.Find(rows).Error; err != nil {
		p.logger.Warn("audit snapshot failed", "table", table, "error", err)
		return nil, false
	}
	rv := reflect.ValueOf(rows).Elem()
	out := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out[i] = rv.Index(i).Addr().Interface()
	}
	return out, true
}

func (p *GormPlugin) record(ctx context.Context, db *gorm.DB, table string, recordID uint64, action Action, diff map[string]any) {
	if len(diff) == 0 {
		return
	}
	if _, ok := db.Statement.ConnPool.(gorm.TxCommitter); !ok {
		if err := p.writeAudit(ctx, db, table, recordID, action, diff); err != nil {
			p.logger.Warn("audit record failed", "table", table, "record_id", recordID, "error", err)
		}
		return
	}
	tx := db.Session(&gorm.Session{NewDB: true})
	if err := tx.SavePoint(auditSavepoint).Error; err != nil {
		p.logger.Warn("audit savepoint failed", "table", table, "record_id", recordID, "error", err)
		return
	}
	defer func() {
		if err := tx.Exec("RELEASE SAVEPOINT " + auditSavepoint).Error; err != nil {
			p.logger.Warn("audit release failed", "table", table, "record_id", recordID, "error", err)
		}
	}()
	if err := p.writeAudit(ctx, tx, table, recordID, action, diff); err != nil {
		if rollbackErr := tx.RollbackTo(auditSavepoint); rollbackErr != nil {
			p.logger.Warn("audit savepoint rollback failed", "table", table, "record_id", recordID, "error", rollbackErr)
		}
		p.logger.Warn("audit record failed", "table", table, "record_id", recordID, "error", err)
	}
}

func (p *GormPlugin) writeAudit(ctx context.Context, db *gorm.DB, table string, recordID uint64, action Action, diff map[string]any) error {
	session := db.Session(&gorm.Session{NewDB: true}).Model(&Log{})
	svc := NewAuditLogService(NewLogDAO(session))
	return Record(ctx, svc, table, recordID, action, diff)
}

func (p *GormPlugin) pkFromReflect(stmt *gorm.Statement) any {
	if stmt.Schema == nil || stmt.Schema.PrioritizedPrimaryField == nil {
		return nil
	}
	rv := stmt.ReflectValue
	if !rv.IsValid() || rv.Kind() == reflect.Slice {
		return nil
	}
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	field := stmt.Schema.PrioritizedPrimaryField
	value, zero := field.ValueOf(stmt.Context, rv)
	if zero {
		return nil
	}
	return value
}

func (p *GormPlugin) idOf(row any) (uint64, bool) {
	rv := reflect.ValueOf(row)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return 0, false
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return 0, false
	}
	id := rv.FieldByName("ID")
	if !id.IsValid() || id.Kind() != reflect.Uint64 {
		return 0, false
	}
	return id.Uint(), true
}

func (p *GormPlugin) skipped(table string) bool {
	_, ok := p.skip[table]
	return ok
}

func rowsOf(rv reflect.Value) []any {
	for rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Struct:
		return []any{rv.Addr().Interface()}
	case reflect.Slice:
		out := make([]any, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			if elem.Kind() == reflect.Pointer {
				out = append(out, elem.Interface())
			} else {
				out = append(out, elem.Addr().Interface())
			}
		}
		return out
	}
	return nil
}
