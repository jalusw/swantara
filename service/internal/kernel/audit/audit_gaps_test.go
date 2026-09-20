package audit

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type errRecorder struct{ err error }

func (r errRecorder) Append(_ context.Context, _ *Log) error { return r.err }

func TestLogDAOMock(t *testing.T) {
	ctx := context.Background()
	def := LogDAOMock{}
	if _, err := def.Search(ctx, "code", "x"); err != nil {
		t.Errorf("search default = %v", err)
	}
	page, err := def.List(ctx, &query.Query{})
	if err != nil || len(page.Items) != 0 {
		t.Errorf("list default = %v %+v", page, err)
	}
	if _, err := def.Find(ctx, 1); err != nil {
		t.Errorf("find default = %v", err)
	}
	entry := &Log{EntityTable: "t"}
	if _, err := def.Create(ctx, entry); err != nil {
		t.Errorf("create default = %v", err)
	}
	if _, err := def.Update(ctx, entry); err != nil {
		t.Errorf("update default = %v", err)
	}
	if err := def.Delete(ctx, 1); err != nil {
		t.Errorf("delete default = %v", err)
	}
	if err := def.HardDelete(ctx, 1); err != nil {
		t.Errorf("hard delete default = %v", err)
	}

	boom := errors.New("boom")
	custom := LogDAOMock{
		SearchFunc:     func(_ context.Context, _ string, _ any) (*Log, error) { return entry, nil },
		ListFunc:       func(_ context.Context, _ *query.Query) (*query.Page[Log], error) { return nil, boom },
		FindFunc:       func(_ context.Context, _ uint64) (*Log, error) { return entry, nil },
		CreateFunc:     func(_ context.Context, e *Log) (*Log, error) { return e, boom },
		UpdateFunc:     func(_ context.Context, e *Log) (*Log, error) { return e, nil },
		DeleteFunc:     func(_ context.Context, _ uint64) error { return boom },
		HardDeleteFunc: func(_ context.Context, _ uint64) error { return nil },
	}
	if _, err := custom.Search(ctx, "code", "x"); err != nil {
		t.Errorf("search = %v", err)
	}
	helper.AssertError(t, func() error { _, err := custom.List(ctx, &query.Query{}); return err }(), true, boom)
	if _, err := custom.Find(ctx, 1); err != nil {
		t.Errorf("find = %v", err)
	}
	helper.AssertError(t, func() error { _, err := custom.Create(ctx, entry); return err }(), true, boom)
	if _, err := custom.Update(ctx, entry); err != nil {
		t.Errorf("update = %v", err)
	}
	helper.AssertError(t, custom.Delete(ctx, 1), true, boom)
	if err := custom.HardDelete(ctx, 1); err != nil {
		t.Errorf("hard delete = %v", err)
	}
}

func TestAudited_ErrorPaths(t *testing.T) {
	ctx := context.Background()
	boom := errors.New("boom")
	before := &saleOrder{Base: model.Base{ID: 3}, Code: "SO/3", State: "draft"}

	t.Run("create inner error", func(t *testing.T) {
		audited := NewAudited(dao.CRUDMock[saleOrder]{
			CreateFunc: func(_ context.Context, _ *saleOrder) (*saleOrder, error) { return nil, boom },
		}, &recorderMock{}, "sale_orders")
		_, err := audited.Create(ctx, &saleOrder{})
		helper.AssertError(t, err, true, boom)
	})

	t.Run("create recorder error", func(t *testing.T) {
		audited := NewAudited(dao.CRUDMock[saleOrder]{
			CreateFunc: func(_ context.Context, e *saleOrder) (*saleOrder, error) { e.ID = 5; return e, nil },
		}, errRecorder{boom}, "sale_orders")
		_, err := audited.Create(ctx, &saleOrder{})
		helper.AssertError(t, err, true, boom)
	})

	t.Run("create non auditable skips record", func(t *testing.T) {
		recorder := &recorderMock{}
		audited := NewAudited(dao.CRUDMock[plainRecord]{
			CreateFunc: func(_ context.Context, e *plainRecord) (*plainRecord, error) { return e, nil },
		}, recorder, "plains")
		if _, err := audited.Create(ctx, &plainRecord{Name: "x"}); err != nil {
			t.Fatal(err)
		}
		if len(recorder.entries) != 0 {
			t.Errorf("entries = %d", len(recorder.entries))
		}
	})

	t.Run("update find error", func(t *testing.T) {
		audited := NewAudited(dao.CRUDMock[saleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*saleOrder, error) { return nil, boom },
		}, &recorderMock{}, "sale_orders")
		_, err := audited.Update(ctx, &saleOrder{Base: model.Base{ID: 3}})
		helper.AssertError(t, err, true, boom)
	})

	t.Run("update inner error", func(t *testing.T) {
		audited := NewAudited(dao.CRUDMock[saleOrder]{
			FindFunc:   func(_ context.Context, _ uint64) (*saleOrder, error) { return before, nil },
			UpdateFunc: func(_ context.Context, _ *saleOrder) (*saleOrder, error) { return nil, boom },
		}, &recorderMock{}, "sale_orders")
		_, err := audited.Update(ctx, &saleOrder{Base: model.Base{ID: 3}})
		helper.AssertError(t, err, true, boom)
	})

	t.Run("update recorder error", func(t *testing.T) {
		audited := NewAudited(dao.CRUDMock[saleOrder]{
			FindFunc:   func(_ context.Context, _ uint64) (*saleOrder, error) { return before, nil },
			UpdateFunc: func(_ context.Context, e *saleOrder) (*saleOrder, error) { return e, nil },
		}, errRecorder{boom}, "sale_orders")
		_, err := audited.Update(ctx, &saleOrder{Base: model.Base{ID: 3}, State: "confirmed"})
		helper.AssertError(t, err, true, boom)
	})

	t.Run("delete missing is no-op", func(t *testing.T) {
		recorder := &recorderMock{}
		audited := NewAudited(dao.CRUDMock[saleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*saleOrder, error) { return nil, nil },
		}, recorder, "sale_orders")
		if err := audited.Delete(ctx, 9); err != nil {
			t.Fatal(err)
		}
		if len(recorder.entries) != 0 {
			t.Errorf("entries = %d", len(recorder.entries))
		}
	})

	t.Run("delete find error", func(t *testing.T) {
		audited := NewAudited(dao.CRUDMock[saleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*saleOrder, error) { return nil, boom },
		}, &recorderMock{}, "sale_orders")
		helper.AssertError(t, audited.Delete(ctx, 9), true, boom)
	})

	t.Run("delete inner error", func(t *testing.T) {
		audited := NewAudited(dao.CRUDMock[saleOrder]{
			FindFunc:   func(_ context.Context, _ uint64) (*saleOrder, error) { return before, nil },
			DeleteFunc: func(_ context.Context, _ uint64) error { return boom },
		}, &recorderMock{}, "sale_orders")
		helper.AssertError(t, audited.Delete(ctx, 3), true, boom)
	})

	t.Run("delete recorder error", func(t *testing.T) {
		audited := NewAudited(dao.CRUDMock[saleOrder]{
			FindFunc:   func(_ context.Context, _ uint64) (*saleOrder, error) { return before, nil },
			DeleteFunc: func(_ context.Context, _ uint64) error { return nil },
		}, errRecorder{boom}, "sale_orders")
		helper.AssertError(t, audited.Delete(ctx, 3), true, boom)
	})

	t.Run("record marshal error", func(t *testing.T) {
		err := Record(ctx, &recorderMock{}, "t", 1, ActionInsert, map[string]any{"fn": func() {}})
		if err == nil {
			t.Error("expected marshal error")
		}
	})
}
