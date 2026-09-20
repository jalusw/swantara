package crosscutting

import "errors"

var (
	ErrApprovalOrganization = errors.New("approval request organization is required")
	ErrApprovalNotFound     = errors.New("approval request not found")
	ErrApprovalStepNotFound = errors.New("approval step not found")
	ErrApprovalNoApprovers  = errors.New("approval request must have at least one approver")
	ErrApprovalOwner        = errors.New("approval request owner is required")
	ErrApprovalState        = errors.New("approval request cannot be decided in its current state")
	ErrApprovalNotApprover  = errors.New("approval step does not belong to the actor")
	ErrApprovalOrder        = errors.New("approval step must be decided after its predecessors")

	ErrAttachmentOrganization = errors.New("attachment organization is required")
	ErrAttachmentNotFound     = errors.New("attachment not found")
	ErrAttachmentOwner        = errors.New("attachment owner is required")
	ErrAttachmentFilename     = errors.New("attachment filename is required")
	ErrAttachmentEmpty        = errors.New("attachment content cannot be empty")

	ErrMessageOrganization = errors.New("message organization is required")
	ErrMessageNotFound     = errors.New("message not found")
	ErrMessageOwner        = errors.New("message owner is required")
	ErrMessageAuthor       = errors.New("message author is required")
	ErrMessageBody         = errors.New("message body is required")
)
