package crosscutting

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	ApprovalStatePending  = "pending"
	ApprovalStateApproved = "approved"
	ApprovalStateRefused  = "refused"

	ApprovalStepPending  = "pending"
	ApprovalStepApproved = "approved"
	ApprovalStepRefused  = "refused"
)

type ApprovalRequest struct {
	model.Base
	OrganizationID uint64 `gorm:"not null" json:"organization_id"`
	OwnerType      string `gorm:"type:text" json:"owner_type"`
	OwnerID        uint64 `json:"owner_id"`
	RequestedBy    uint64 `json:"requested_by"`
	State          string `gorm:"type:text" json:"state"`
}

func (ApprovalRequest) TableName() string {
	return "approval_requests"
}

type ApprovalStep struct {
	model.Base
	OrganizationID uint64     `gorm:"not null" json:"organization_id"`
	RequestID      uint64     `gorm:"not null" json:"request_id"`
	ApproverID     uint64     `gorm:"not null" json:"approver_id"`
	Sequence       int        `json:"sequence"`
	Decision       string     `gorm:"type:text" json:"decision"`
	DecidedAt      *time.Time `json:"decided_at"`
	Comment        *string    `json:"comment"`
}

func (ApprovalStep) TableName() string {
	return "approval_steps"
}

type Attachment struct {
	model.Base
	OrganizationID uint64     `gorm:"not null" json:"organization_id"`
	OwnerType      string     `gorm:"type:text" json:"owner_type"`
	OwnerID        uint64     `json:"owner_id"`
	Filename       string     `gorm:"type:text" json:"filename"`
	MimeType       string     `gorm:"type:text" json:"mime_type"`
	ByteSize       int64      `json:"byte_size"`
	StorageURL     string     `gorm:"type:text" json:"storage_url"`
	Checksum       string     `gorm:"type:text" json:"checksum"`
	UploadedBy     uint64     `json:"uploaded_by"`
	UploadedAt     *time.Time `json:"uploaded_at"`
}

func (Attachment) TableName() string {
	return "attachments"
}

const (
	MessageTypeNote   = "note"
	MessageTypeSystem = "system"
)

type Message struct {
	model.Base
	OrganizationID uint64 `gorm:"not null" json:"organization_id"`
	OwnerType      string `gorm:"type:text" json:"owner_type"`
	OwnerID        uint64 `json:"owner_id"`
	AuthorID       uint64 `json:"author_id"`
	Body           string `gorm:"type:text" json:"body"`
	MessageType    string `gorm:"type:text" json:"message_type"`
}

func (Message) TableName() string {
	return "messages"
}
