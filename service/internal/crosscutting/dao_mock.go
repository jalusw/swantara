package crosscutting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type ApprovalRequestDAOMock struct {
	dao.CRUDMock[ApprovalRequest]
	FindByOwnerFunc     func(ctx context.Context, ownerType string, ownerID uint64) (*ApprovalRequest, error)
	CreateWithStepsFunc func(ctx context.Context, request *ApprovalRequest, steps []*ApprovalStep) (*ApprovalRequest, error)
	UpdateTxFunc        func(ctx context.Context, tx *gorm.DB, request *ApprovalRequest) (*ApprovalRequest, error)
}

func (m ApprovalRequestDAOMock) FindByOwner(ctx context.Context, ownerType string, ownerID uint64) (*ApprovalRequest, error) {
	if m.FindByOwnerFunc != nil {
		return m.FindByOwnerFunc(ctx, ownerType, ownerID)
	}
	return nil, nil
}

func (m ApprovalRequestDAOMock) CreateWithSteps(ctx context.Context, request *ApprovalRequest, steps []*ApprovalStep) (*ApprovalRequest, error) {
	if m.CreateWithStepsFunc != nil {
		return m.CreateWithStepsFunc(ctx, request, steps)
	}
	return request, nil
}

func (m ApprovalRequestDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, request *ApprovalRequest) (*ApprovalRequest, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, request)
	}
	return request, nil
}

type ApprovalStepDAOMock struct {
	dao.CRUDMock[ApprovalStep]
	ListByRequestFunc func(ctx context.Context, requestID uint64) ([]*ApprovalStep, error)
	UpdateTxFunc      func(ctx context.Context, tx *gorm.DB, step *ApprovalStep) (*ApprovalStep, error)
}

func (m ApprovalStepDAOMock) ListByRequest(ctx context.Context, requestID uint64) ([]*ApprovalStep, error) {
	if m.ListByRequestFunc != nil {
		return m.ListByRequestFunc(ctx, requestID)
	}
	return []*ApprovalStep{}, nil
}

func (m ApprovalStepDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, step *ApprovalStep) (*ApprovalStep, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, step)
	}
	return step, nil
}

type AttachmentDAOMock struct {
	dao.CRUDMock[Attachment]
	ListByOwnerFunc func(ctx context.Context, ownerType string, ownerID uint64) ([]*Attachment, error)
}

func (m AttachmentDAOMock) ListByOwner(ctx context.Context, ownerType string, ownerID uint64) ([]*Attachment, error) {
	if m.ListByOwnerFunc != nil {
		return m.ListByOwnerFunc(ctx, ownerType, ownerID)
	}
	return []*Attachment{}, nil
}

type MessageDAOMock struct {
	dao.CRUDMock[Message]
	ListByOwnerFunc func(ctx context.Context, ownerType string, ownerID uint64) ([]*Message, error)
}

func (m MessageDAOMock) ListByOwner(ctx context.Context, ownerType string, ownerID uint64) ([]*Message, error) {
	if m.ListByOwnerFunc != nil {
		return m.ListByOwnerFunc(ctx, ownerType, ownerID)
	}
	return []*Message{}, nil
}
