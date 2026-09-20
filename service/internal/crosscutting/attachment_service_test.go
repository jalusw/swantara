package crosscutting

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type storeMock struct {
	saved   map[string][]byte
	opened  map[string][]byte
	deleted []string
}

func newStoreMock() *storeMock {
	return &storeMock{saved: map[string][]byte{}, opened: map[string][]byte{}}
}

func (m *storeMock) Save(_ context.Context, id string, reader io.Reader) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	m.saved[id] = content
	m.opened[id] = content
	return nil
}

func (m *storeMock) Open(_ context.Context, id string) (io.ReadCloser, error) {
	content, ok := m.opened[id]
	if !ok {
		return nil, io.EOF
	}
	return io.NopCloser(bytes.NewReader(content)), nil
}

func (m *storeMock) Delete(_ context.Context, id string) error {
	m.deleted = append(m.deleted, id)
	delete(m.saved, id)
	delete(m.opened, id)
	return nil
}

func TestAttachmentServiceUploadStoresAndRecordsChecksum(t *testing.T) {
	store := newStoreMock()
	attachments := &AttachmentDAOMock{}
	svc := NewAttachmentService(attachments, store)
	content := []byte("invoice scan")
	attachments.CreateFunc = func(_ context.Context, attachment *Attachment) (*Attachment, error) {
		attachment.ID = 5
		return attachment, nil
	}

	created, err := svc.Upload(context.Background(), 7, "purchase_order", 10, 2, "scan.pdf", "application/pdf", bytes.NewReader(content))

	helper.AssertError(t, err, false, nil)
	if created.ID != 5 {
		t.Fatalf("expected attachment id 5, got %d", created.ID)
	}
	if created.OrganizationID != 7 {
		t.Errorf("organization not recorded: %+v", created)
	}
	if created.OwnerType != "purchase_order" || created.OwnerID != 10 || created.UploadedBy != 2 {
		t.Errorf("owner/uploader not recorded: %+v", created)
	}
	if created.Checksum == "" {
		t.Error("checksum was not recorded")
	}
	if created.StorageURL == "" {
		t.Error("storage url was not recorded")
	}
	if created.ByteSize != int64(len(content)) {
		t.Errorf("byte_size = %d, want %d", created.ByteSize, len(content))
	}
	if len(store.saved) != 1 {
		t.Errorf("stored objects = %d, want 1", len(store.saved))
	}
}

func TestAttachmentServiceUploadRequiresOrganization(t *testing.T) {
	svc := NewAttachmentService(&AttachmentDAOMock{}, newStoreMock())

	_, err := svc.Upload(context.Background(), 0, "purchase_order", 10, 2, "scan.pdf", "application/pdf", bytes.NewReader([]byte("x")))

	helper.AssertError(t, err, true, ErrAttachmentOrganization)
}

func TestAttachmentServiceUploadRequiresOwner(t *testing.T) {
	svc := NewAttachmentService(&AttachmentDAOMock{}, newStoreMock())

	_, err := svc.Upload(context.Background(), 7, "purchase_order", 0, 2, "scan.pdf", "application/pdf", bytes.NewReader([]byte("x")))

	helper.AssertError(t, err, true, ErrAttachmentOwner)
}

func TestAttachmentServiceUploadRequiresContent(t *testing.T) {
	svc := NewAttachmentService(&AttachmentDAOMock{}, newStoreMock())

	_, err := svc.Upload(context.Background(), 7, "purchase_order", 10, 2, "empty.pdf", "application/pdf", bytes.NewReader(nil))

	helper.AssertError(t, err, true, ErrAttachmentEmpty)
}

func TestAttachmentServiceDownloadReturnsNotFound(t *testing.T) {
	attachments := &AttachmentDAOMock{}
	attachments.FindFunc = func(_ context.Context, _ uint64) (*Attachment, error) {
		return nil, nil
	}
	svc := NewAttachmentService(attachments, newStoreMock())

	_, _, err := svc.Download(context.Background(), 7, 99)

	helper.AssertError(t, err, true, ErrAttachmentNotFound)
}

func TestAttachmentServiceDownloadRejectsForeignOrganization(t *testing.T) {
	attachments := &AttachmentDAOMock{}
	attachments.FindFunc = func(_ context.Context, _ uint64) (*Attachment, error) {
		return &Attachment{Base: model.Base{ID: 5}, OrganizationID: 8, StorageURL: "attachments/10/abc"}, nil
	}
	svc := NewAttachmentService(attachments, newStoreMock())

	_, _, err := svc.Download(context.Background(), 7, 5)

	helper.AssertError(t, err, true, ErrAttachmentNotFound)
}

func TestAttachmentServiceDeleteRemovesObject(t *testing.T) {
	store := newStoreMock()
	attachments := &AttachmentDAOMock{}
	attachments.FindFunc = func(_ context.Context, _ uint64) (*Attachment, error) {
		return &Attachment{Base: model.Base{ID: 5}, OrganizationID: 7, StorageURL: "attachments/10/abc"}, nil
	}
	svc := NewAttachmentService(attachments, store)

	err := svc.Delete(context.Background(), 7, 5)

	helper.AssertError(t, err, false, nil)
	if len(store.deleted) != 1 || store.deleted[0] != "attachments/10/abc" {
		t.Errorf("deleted = %v, want the stored object", store.deleted)
	}
}

func TestAttachmentServiceDeleteRejectsForeignOrganization(t *testing.T) {
	store := newStoreMock()
	attachments := &AttachmentDAOMock{}
	attachments.FindFunc = func(_ context.Context, _ uint64) (*Attachment, error) {
		return &Attachment{Base: model.Base{ID: 5}, OrganizationID: 8, StorageURL: "attachments/10/abc"}, nil
	}
	svc := NewAttachmentService(attachments, store)

	err := svc.Delete(context.Background(), 7, 5)

	helper.AssertError(t, err, true, ErrAttachmentNotFound)
	if len(store.deleted) != 0 {
		t.Errorf("deleted = %v, want none", store.deleted)
	}
}

func TestMessageServicePostCreatesMessage(t *testing.T) {
	messages := &MessageDAOMock{}
	svc := NewMessageService(messages)
	messages.CreateFunc = func(_ context.Context, message *Message) (*Message, error) {
		message.ID = 9
		return message, nil
	}

	created, err := svc.Post(context.Background(), 7, "purchase_order", 10, 2, "please expedite", "")

	helper.AssertError(t, err, false, nil)
	if created.ID != 9 {
		t.Fatalf("expected message id 9, got %d", created.ID)
	}
	if created.OrganizationID != 7 {
		t.Errorf("organization not recorded: %+v", created)
	}
	if created.OwnerType != "purchase_order" || created.OwnerID != 10 || created.AuthorID != 2 {
		t.Errorf("owner/author not recorded: %+v", created)
	}
	if created.MessageType != MessageTypeNote {
		t.Errorf("message_type = %s, want note", created.MessageType)
	}
}

func TestMessageServicePostRequiresOrganization(t *testing.T) {
	svc := NewMessageService(&MessageDAOMock{})

	_, err := svc.Post(context.Background(), 0, "purchase_order", 10, 2, "body", "")

	helper.AssertError(t, err, true, ErrMessageOrganization)
}

func TestMessageServicePostRequiresOwner(t *testing.T) {
	svc := NewMessageService(&MessageDAOMock{})

	_, err := svc.Post(context.Background(), 7, "purchase_order", 0, 2, "body", "")

	helper.AssertError(t, err, true, ErrMessageOwner)
}

func TestMessageServicePostRequiresAuthor(t *testing.T) {
	svc := NewMessageService(&MessageDAOMock{})

	_, err := svc.Post(context.Background(), 7, "purchase_order", 10, 0, "body", "")

	helper.AssertError(t, err, true, ErrMessageAuthor)
}

func TestMessageServicePostRequiresBody(t *testing.T) {
	svc := NewMessageService(&MessageDAOMock{})

	_, err := svc.Post(context.Background(), 7, "purchase_order", 10, 2, "", "")

	helper.AssertError(t, err, true, ErrMessageBody)
}
