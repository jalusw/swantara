package crosscutting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type ApprovalRequestDAO interface {
	dao.CRUD[ApprovalRequest]
	FindByOwner(ctx context.Context, ownerType string, ownerID uint64) (*ApprovalRequest, error)
	CreateWithSteps(ctx context.Context, request *ApprovalRequest, steps []*ApprovalStep) (*ApprovalRequest, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, request *ApprovalRequest) (*ApprovalRequest, error)
}

type approvalRequestDAO struct {
	dao.Base[ApprovalRequest]
	db *gorm.DB
}

func NewApprovalRequestDAO(db *gorm.DB) ApprovalRequestDAO {
	return approvalRequestDAO{Base: dao.NewBase[ApprovalRequest](db), db: db}
}

func (d approvalRequestDAO) FindByOwner(ctx context.Context, ownerType string, ownerID uint64) (*ApprovalRequest, error) {
	var entity ApprovalRequest
	err := d.db.WithContext(ctx).Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).Take(&entity).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (d approvalRequestDAO) CreateWithSteps(ctx context.Context, request *ApprovalRequest, steps []*ApprovalStep) (*ApprovalRequest, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(request).Error; err != nil {
			return err
		}
		for _, step := range steps {
			step.RequestID = request.ID
			if err := tx.Create(step).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (d approvalRequestDAO) UpdateTx(ctx context.Context, tx *gorm.DB, request *ApprovalRequest) (*ApprovalRequest, error) {
	if err := tx.WithContext(ctx).Save(request).Error; err != nil {
		return nil, err
	}
	return request, nil
}

type ApprovalStepDAO interface {
	dao.CRUD[ApprovalStep]
	ListByRequest(ctx context.Context, requestID uint64) ([]*ApprovalStep, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, step *ApprovalStep) (*ApprovalStep, error)
}

type approvalStepDAO struct {
	dao.Base[ApprovalStep]
	db *gorm.DB
}

func NewApprovalStepDAO(db *gorm.DB) ApprovalStepDAO {
	return approvalStepDAO{Base: dao.NewBase[ApprovalStep](db), db: db}
}

func (d approvalStepDAO) ListByRequest(ctx context.Context, requestID uint64) ([]*ApprovalStep, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "request_id", Operator: query.Equal, Value: requestID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d approvalStepDAO) UpdateTx(ctx context.Context, tx *gorm.DB, step *ApprovalStep) (*ApprovalStep, error) {
	if err := tx.WithContext(ctx).Save(step).Error; err != nil {
		return nil, err
	}
	return step, nil
}

type AttachmentDAO interface {
	dao.CRUD[Attachment]
	ListByOwner(ctx context.Context, ownerType string, ownerID uint64) ([]*Attachment, error)
}

type attachmentDAO struct {
	dao.Base[Attachment]
	db *gorm.DB
}

func NewAttachmentDAO(db *gorm.DB) AttachmentDAO {
	return attachmentDAO{Base: dao.NewBase[Attachment](db), db: db}
}

func (d attachmentDAO) ListByOwner(ctx context.Context, ownerType string, ownerID uint64) ([]*Attachment, error) {
	var entities []Attachment
	err := d.db.WithContext(ctx).Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	items := make([]*Attachment, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}

type MessageDAO interface {
	dao.CRUD[Message]
	ListByOwner(ctx context.Context, ownerType string, ownerID uint64) ([]*Message, error)
}

type messageDAO struct {
	dao.Base[Message]
	db *gorm.DB
}

func NewMessageDAO(db *gorm.DB) MessageDAO {
	return messageDAO{Base: dao.NewBase[Message](db), db: db}
}

func (d messageDAO) ListByOwner(ctx context.Context, ownerType string, ownerID uint64) ([]*Message, error) {
	var entities []Message
	err := d.db.WithContext(ctx).Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	items := make([]*Message, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}
