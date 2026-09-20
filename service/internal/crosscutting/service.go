package crosscutting

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/storage"
	"gorm.io/gorm"
)

type ApprovalService struct {
	requests ApprovalRequestDAO
	steps    ApprovalStepDAO
	tx       db.Transactioner
	now      func() time.Time
}

func NewApprovalService(requests ApprovalRequestDAO, steps ApprovalStepDAO, tx db.Transactioner) ApprovalService {
	return ApprovalService{requests: requests, steps: steps, tx: tx, now: time.Now}
}

func (s ApprovalService) List(ctx context.Context, q *query.Query) (*query.Page[ApprovalRequest], error) {
	return s.requests.List(ctx, q)
}

func (s ApprovalService) Find(ctx context.Context, id uint64) (*ApprovalRequest, error) {
	return s.requests.Find(ctx, id)
}

func (s ApprovalService) ListSteps(ctx context.Context, requestID uint64) ([]*ApprovalStep, error) {
	return s.steps.ListByRequest(ctx, requestID)
}

func (s ApprovalService) Create(ctx context.Context, organizationID uint64, ownerType string, ownerID, requestedBy uint64, approverIDs []uint64) (*ApprovalRequest, error) {
	if organizationID == 0 {
		return nil, ErrApprovalOrganization
	}
	if len(approverIDs) == 0 {
		return nil, ErrApprovalNoApprovers
	}
	if ownerID == 0 {
		return nil, ErrApprovalOwner
	}
	request := &ApprovalRequest{
		OrganizationID: organizationID,
		OwnerType:      ownerType,
		OwnerID:        ownerID,
		RequestedBy:    requestedBy,
		State:          ApprovalStatePending,
	}
	steps := make([]*ApprovalStep, 0, len(approverIDs))
	for i, approverID := range approverIDs {
		steps = append(steps, &ApprovalStep{
			OrganizationID: organizationID,
			ApproverID:     approverID,
			Sequence:       (i + 1) * 10,
			Decision:       ApprovalStepPending,
		})
	}
	return s.requests.CreateWithSteps(ctx, request, steps)
}

func (s ApprovalService) Decide(ctx context.Context, organizationID, requestID, stepID, approverID uint64, approve bool, comment string) (*ApprovalRequest, error) {
	request, err := s.requests.Find(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if request == nil || request.OrganizationID != organizationID {
		return nil, ErrApprovalNotFound
	}
	if request.State != ApprovalStatePending {
		return nil, ErrApprovalState
	}

	steps, err := s.steps.ListByRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if len(steps) == 0 {
		return nil, ErrApprovalNotFound
	}

	var target *ApprovalStep
	for _, step := range steps {
		if step.ID == stepID {
			target = step
			break
		}
	}
	if target == nil {
		return nil, ErrApprovalStepNotFound
	}
	if target.ApproverID != approverID {
		return nil, ErrApprovalNotApprover
	}
	if target.Decision != ApprovalStepPending {
		return nil, ErrApprovalState
	}
	for _, step := range steps {
		if step.Sequence < target.Sequence && step.Decision != ApprovalStepApproved {
			return nil, ErrApprovalOrder
		}
	}

	now := s.now().UTC()
	decision := ApprovalStepApproved
	if !approve {
		decision = ApprovalStepRefused
	}
	target.Decision = decision
	target.DecidedAt = &now
	if comment != "" {
		target.Comment = &comment
	}

	var updated *ApprovalRequest
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		if _, err := s.steps.UpdateTx(ctx, tx, target); err != nil {
			return err
		}
		nextState := ApprovalStatePending
		if decision == ApprovalStepRefused {
			nextState = ApprovalStateRefused
		} else {
			allApproved := true
			for _, step := range steps {
				if step.ID != target.ID && step.Decision != ApprovalStepApproved {
					allApproved = false
					break
				}
			}
			if allApproved {
				nextState = ApprovalStateApproved
			}
		}
		if nextState == ApprovalStatePending {
			updated = request
			return nil
		}
		request.State = nextState
		updated, err = s.requests.UpdateTx(ctx, tx, request)
		return err
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s ApprovalService) State(ctx context.Context, ownerType string, ownerID uint64) (*ApprovalRequest, []*ApprovalStep, error) {
	request, err := s.requests.FindByOwner(ctx, ownerType, ownerID)
	if err != nil {
		return nil, nil, err
	}
	if request == nil {
		return nil, nil, nil
	}
	steps, err := s.steps.ListByRequest(ctx, request.ID)
	if err != nil {
		return nil, nil, err
	}
	return request, steps, nil
}

func (s ApprovalService) IsApproved(ctx context.Context, ownerType string, ownerID uint64) (bool, error) {
	request, err := s.requests.FindByOwner(ctx, ownerType, ownerID)
	if err != nil {
		return false, err
	}
	return request != nil && request.State == ApprovalStateApproved, nil
}

type AttachmentService struct {
	attachments AttachmentDAO
	store       storage.Store
	now         func() time.Time
}

func NewAttachmentService(attachments AttachmentDAO, store storage.Store) AttachmentService {
	return AttachmentService{attachments: attachments, store: store, now: time.Now}
}

func (s AttachmentService) List(ctx context.Context, q *query.Query) (*query.Page[Attachment], error) {
	return s.attachments.List(ctx, q)
}

func (s AttachmentService) Find(ctx context.Context, id uint64) (*Attachment, error) {
	return s.attachments.Find(ctx, id)
}

func (s AttachmentService) Upload(ctx context.Context, organizationID uint64, ownerType string, ownerID, uploadedBy uint64, filename, mimeType string, reader io.Reader) (*Attachment, error) {
	if organizationID == 0 {
		return nil, ErrAttachmentOrganization
	}
	if ownerID == 0 {
		return nil, ErrAttachmentOwner
	}
	if filename == "" {
		return nil, ErrAttachmentFilename
	}
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return nil, ErrAttachmentEmpty
	}
	checksum := sha256.Sum256(content)
	objectID := fmt.Sprintf("attachments/%d/%s", ownerID, hex.EncodeToString(checksum[:]))
	if err := s.store.Save(ctx, objectID, bytes.NewReader(content)); err != nil {
		return nil, err
	}
	now := s.now().UTC()
	attachment := &Attachment{
		OrganizationID: organizationID,
		OwnerType:      ownerType,
		OwnerID:        ownerID,
		Filename:       filename,
		MimeType:       mimeType,
		ByteSize:       int64(len(content)),
		StorageURL:     objectID,
		Checksum:       hex.EncodeToString(checksum[:]),
		UploadedBy:     uploadedBy,
		UploadedAt:     &now,
	}
	created, err := s.attachments.Create(ctx, attachment)
	if err != nil {
		if delErr := s.store.Delete(ctx, objectID); delErr != nil {
			slog.Error("failed to cleanup orphaned attachment", "object_id", objectID, "error", delErr)
		}
		return nil, err
	}
	return created, nil
}

func (s AttachmentService) Download(ctx context.Context, organizationID, id uint64) (*Attachment, io.ReadCloser, error) {
	attachment, err := s.ownedAttachment(ctx, organizationID, id)
	if err != nil {
		return nil, nil, err
	}
	reader, err := s.store.Open(ctx, attachment.StorageURL)
	if err != nil {
		return nil, nil, err
	}
	return attachment, reader, nil
}

func (s AttachmentService) Delete(ctx context.Context, organizationID, id uint64) error {
	attachment, err := s.ownedAttachment(ctx, organizationID, id)
	if err != nil {
		return err
	}
	if err := s.store.Delete(ctx, attachment.StorageURL); err != nil {
		return err
	}
	return s.attachments.Delete(ctx, id)
}

func (s AttachmentService) ownedAttachment(ctx context.Context, organizationID, id uint64) (*Attachment, error) {
	attachment, err := s.attachments.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if attachment == nil || attachment.OrganizationID != organizationID {
		return nil, ErrAttachmentNotFound
	}
	return attachment, nil
}

type MessageService struct {
	messages MessageDAO
	now      func() time.Time
}

func NewMessageService(messages MessageDAO) MessageService {
	return MessageService{messages: messages, now: time.Now}
}

func (s MessageService) List(ctx context.Context, q *query.Query) (*query.Page[Message], error) {
	return s.messages.List(ctx, q)
}

func (s MessageService) Find(ctx context.Context, id uint64) (*Message, error) {
	return s.messages.Find(ctx, id)
}

func (s MessageService) Post(ctx context.Context, organizationID uint64, ownerType string, ownerID, authorID uint64, body, messageType string) (*Message, error) {
	if organizationID == 0 {
		return nil, ErrMessageOrganization
	}
	if ownerID == 0 {
		return nil, ErrMessageOwner
	}
	if authorID == 0 {
		return nil, ErrMessageAuthor
	}
	if body == "" {
		return nil, ErrMessageBody
	}
	if messageType == "" {
		messageType = MessageTypeNote
	}
	message := &Message{
		OrganizationID: organizationID,
		OwnerType:      ownerType,
		OwnerID:        ownerID,
		AuthorID:       authorID,
		Body:           body,
		MessageType:    messageType,
	}
	return s.messages.Create(ctx, message)
}
