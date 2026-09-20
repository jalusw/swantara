package crosscutting

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestApprovalRequestDAO_FindByOwner_ReturnsRequest(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "approval_requests"`)).
		WithArgs("purchase_order", 10, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_type", "owner_id", "requested_by", "state"}).
			AddRow(1, "purchase_order", 10, 2, ApprovalStatePending))

	requests := NewApprovalRequestDAO(db)
	request, err := requests.FindByOwner(ctx, "purchase_order", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if request == nil || request.OwnerID != 10 {
		t.Fatalf("request = %+v, want the found approval request", request)
	}

	query.AssertDBMockDone(t, mock)
}

func TestApprovalRequestDAO_FindByOwner_ReturnsNilWhenMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "approval_requests"`)).
		WithArgs("purchase_order", 99, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	requests := NewApprovalRequestDAO(db)
	request, err := requests.FindByOwner(ctx, "purchase_order", 99)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if request != nil {
		t.Fatalf("request = %+v, want nil", request)
	}

	query.AssertDBMockDone(t, mock)
}

func TestApprovalRequestDAO_FindByOwner_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "approval_requests"`)).
		WithArgs("purchase_order", 10).
		WillReturnError(errors.New("db down"))

	requests := NewApprovalRequestDAO(db)
	_, err := requests.FindByOwner(ctx, "purchase_order", 10)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestApprovalRequestDAO_CreateWithSteps_CreatesRequestAndSteps(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "approval_requests"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "approval_steps"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()

	requests := NewApprovalRequestDAO(db)
	request := &ApprovalRequest{OwnerType: "purchase_order", OwnerID: 10, RequestedBy: 2, State: ApprovalStatePending}
	steps := []*ApprovalStep{{ApproverID: 3, Sequence: 10, Decision: ApprovalStepPending}}

	created, err := requests.CreateWithSteps(ctx, request, steps)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("request id = %d, want 1", created.ID)
	}
	if steps[0].RequestID != 1 {
		t.Errorf("step request_id = %d, want 1", steps[0].RequestID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestApprovalRequestDAO_CreateWithSteps_RollsBackOnStepError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "approval_requests"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "approval_steps"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	requests := NewApprovalRequestDAO(db)
	_, err := requests.CreateWithSteps(ctx, &ApprovalRequest{}, []*ApprovalStep{{ApproverID: 3}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestApprovalRequestDAO_CreateWithSteps_RollsBackOnRequestError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "approval_requests"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	requests := NewApprovalRequestDAO(db)
	_, err := requests.CreateWithSteps(ctx, &ApprovalRequest{}, []*ApprovalStep{{ApproverID: 3}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestApprovalRequestDAO_UpdateTx_SavesRequest(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "approval_requests" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	requests := NewApprovalRequestDAO(db)
	request := &ApprovalRequest{Base: model.Base{ID: 1}, OwnerType: "purchase_order", OwnerID: 10, State: ApprovalStateApproved}

	updated, err := requests.UpdateTx(ctx, db, request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != request {
		t.Errorf("UpdateTx returned a different request")
	}

	query.AssertDBMockDone(t, mock)
}

func TestApprovalRequestDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "approval_requests" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	requests := NewApprovalRequestDAO(db)
	_, err := requests.UpdateTx(ctx, db, &ApprovalRequest{Base: model.Base{ID: 1}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestApprovalStepDAO_ListByRequest_ReturnsSteps(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "approval_steps" WHERE request_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "approval_steps" WHERE request_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_id", "approver_id", "sequence", "decision"}).
			AddRow(11, 7, 3, 10, ApprovalStepPending).
			AddRow(12, 7, 4, 20, ApprovalStepApproved))

	steps := NewApprovalStepDAO(db)

	items, err := steps.ListByRequest(ctx, 7)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].ApproverID != 3 || items[1].Sequence != 20 {
		t.Errorf("items = %+v, want two steps", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestApprovalStepDAO_ListByRequest_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "approval_steps"`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	steps := NewApprovalStepDAO(db)

	_, err := steps.ListByRequest(ctx, 7)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestApprovalStepDAO_UpdateTx_SavesStep(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "approval_steps" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	steps := NewApprovalStepDAO(db)
	step := &ApprovalStep{Base: model.Base{ID: 11}, RequestID: 7, ApproverID: 3, Decision: ApprovalStepApproved}

	updated, err := steps.UpdateTx(ctx, db, step)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != step {
		t.Errorf("UpdateTx returned a different step")
	}

	query.AssertDBMockDone(t, mock)
}

func TestApprovalStepDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "approval_steps" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	steps := NewApprovalStepDAO(db)
	_, err := steps.UpdateTx(ctx, db, &ApprovalStep{Base: model.Base{ID: 11}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestAttachmentDAO_ListByOwner_ReturnsAttachments(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attachments"`)).
		WithArgs("purchase_order", 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_type", "owner_id", "filename"}).
			AddRow(5, "purchase_order", 10, "scan.pdf"))

	attachments := NewAttachmentDAO(db)

	items, err := attachments.ListByOwner(ctx, "purchase_order", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Filename != "scan.pdf" {
		t.Errorf("items = %+v, want one attachment", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAttachmentDAO_ListByOwner_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "attachments"`)).
		WithArgs("purchase_order", 10).
		WillReturnError(errors.New("db down"))

	attachments := NewAttachmentDAO(db)

	_, err := attachments.ListByOwner(ctx, "purchase_order", 10)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestMessageDAO_ListByOwner_ReturnsMessages(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "messages"`)).
		WithArgs("purchase_order", 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_type", "owner_id", "body"}).
			AddRow(9, "purchase_order", 10, "please expedite"))

	messages := NewMessageDAO(db)

	items, err := messages.ListByOwner(ctx, "purchase_order", 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Body != "please expedite" {
		t.Errorf("items = %+v, want one message", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestMessageDAO_ListByOwner_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "messages"`)).
		WithArgs("purchase_order", 10).
		WillReturnError(errors.New("db down"))

	messages := NewMessageDAO(db)

	_, err := messages.ListByOwner(ctx, "purchase_order", 10)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
