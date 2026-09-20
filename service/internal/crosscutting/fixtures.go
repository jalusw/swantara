package crosscutting

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func ApprovalRequestFixture(opts ...func(*ApprovalRequest) *ApprovalRequest) *ApprovalRequest {
	a := &ApprovalRequest{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000))},
		OrganizationID: uint64(gofakeit.Number(1, 10000)),
		OwnerType:      gofakeit.Word(),
		OwnerID:        uint64(gofakeit.Number(1, 10000)),
		RequestedBy:    uint64(gofakeit.Number(1, 10000)),
		State:          ApprovalStatePending,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func AttachmentFixture(opts ...func(*Attachment) *Attachment) *Attachment {
	a := &Attachment{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000))},
		OrganizationID: uint64(gofakeit.Number(1, 10000)),
		OwnerType:      gofakeit.Word(),
		OwnerID:        uint64(gofakeit.Number(1, 10000)),
		Filename:       gofakeit.Word() + ".pdf",
		MimeType:       "application/octet-stream",
		UploadedBy:     uint64(gofakeit.Number(1, 10000)),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func MessageFixture(opts ...func(*Message) *Message) *Message {
	m := &Message{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000))},
		OrganizationID: uint64(gofakeit.Number(1, 10000)),
		OwnerType:      gofakeit.Word(),
		OwnerID:        uint64(gofakeit.Number(1, 10000)),
		AuthorID:       uint64(gofakeit.Number(1, 10000)),
		Body:           gofakeit.Sentence(5),
		MessageType:    MessageTypeNote,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}
