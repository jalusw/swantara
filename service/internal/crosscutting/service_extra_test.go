package crosscutting

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestCrosscuttingService_Upload(t *testing.T) {
	tests := []struct {
		name       string
		svc        func() AttachmentService
		filename   string
		reader     io.Reader
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "requires filename",
			svc: func() AttachmentService {
				return NewAttachmentService(&AttachmentDAOMock{}, StoreMock{})
			},
			filename:   "",
			reader:     strings.NewReader("x"),
			wantErr:    true,
			wantErrVal: ErrAttachmentFilename,
		},
		{
			name: "fails when reader fails",
			svc: func() AttachmentService {
				return NewAttachmentService(&AttachmentDAOMock{}, StoreMock{})
			},
			filename: "scan.pdf",
			reader:   errReader{},
			wantErr:  true,
		},
		{
			name: "fails when store save fails",
			svc: func() AttachmentService {
				return NewAttachmentService(&AttachmentDAOMock{}, StoreMock{SaveFunc: func(_ context.Context, _ string, _ io.Reader) error {
					return errors.New("storage down")
				}})
			},
			filename: "scan.pdf",
			reader:   strings.NewReader("x"),
			wantErr:  true,
		},
		{
			name: "removes object when create fails",
			svc: func() AttachmentService {
				attachments := &AttachmentDAOMock{}
				attachments.CreateFunc = func(_ context.Context, _ *Attachment) (*Attachment, error) {
					return nil, errors.New("insert failed")
				}
				return NewAttachmentService(attachments, StoreMock{DeleteFunc: func(_ context.Context, _ string) error {
					return nil
				}})
			},
			filename: "scan.pdf",
			reader:   strings.NewReader("x"),
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc()
			_, err := svc.Upload(context.Background(), 7, "purchase_order", 10, 2, tt.filename, "application/pdf", tt.reader)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
		})
	}
}

func TestCrosscuttingService_Download(t *testing.T) {
	tests := []struct {
		name       string
		svc        func() AttachmentService
		wantErr    bool
		wantErrVal error
		assert     func(t *testing.T, attachment *Attachment, reader io.ReadCloser)
	}{
		{
			name: "fails when lookup fails",
			svc: func() AttachmentService {
				attachments := &AttachmentDAOMock{}
				attachments.FindFunc = func(_ context.Context, _ uint64) (*Attachment, error) {
					return nil, errors.New("db down")
				}
				return NewAttachmentService(attachments, StoreMock{})
			},
			wantErr: true,
		},
		{
			name: "returns object",
			svc: func() AttachmentService {
				attachments := &AttachmentDAOMock{}
				attachments.FindFunc = func(_ context.Context, _ uint64) (*Attachment, error) {
					return &Attachment{Base: model.Base{ID: 5}, OrganizationID: 7, StorageURL: "attachments/10/abc"}, nil
				}
				return NewAttachmentService(attachments, StoreMock{})
			},
			assert: func(t *testing.T, attachment *Attachment, reader io.ReadCloser) {
				if attachment == nil || reader == nil {
					t.Fatal("expected attachment and reader")
				}
				_ = reader.Close()
			},
		},
		{
			name: "fails when store open fails",
			svc: func() AttachmentService {
				attachments := &AttachmentDAOMock{}
				attachments.FindFunc = func(_ context.Context, _ uint64) (*Attachment, error) {
					return &Attachment{Base: model.Base{ID: 5}, OrganizationID: 7, StorageURL: "attachments/10/abc"}, nil
				}
				return NewAttachmentService(attachments, StoreMock{OpenFunc: func(_ context.Context, _ string) (io.ReadCloser, error) {
					return nil, errors.New("storage down")
				}})
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc()
			attachment, reader, err := svc.Download(context.Background(), 7, 5)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.assert != nil {
				tt.assert(t, attachment, reader)
			}
		})
	}
}

func TestCrosscuttingService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		svc        func() AttachmentService
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "fails when lookup fails",
			svc: func() AttachmentService {
				attachments := &AttachmentDAOMock{}
				attachments.FindFunc = func(_ context.Context, _ uint64) (*Attachment, error) {
					return nil, errors.New("db down")
				}
				return NewAttachmentService(attachments, StoreMock{})
			},
			wantErr: true,
		},
		{
			name: "returns not found when missing",
			svc: func() AttachmentService {
				attachments := &AttachmentDAOMock{}
				attachments.FindFunc = func(_ context.Context, _ uint64) (*Attachment, error) {
					return nil, nil
				}
				return NewAttachmentService(attachments, StoreMock{})
			},
			wantErr:    true,
			wantErrVal: ErrAttachmentNotFound,
		},
		{
			name: "fails when store delete fails",
			svc: func() AttachmentService {
				attachments := &AttachmentDAOMock{}
				attachments.FindFunc = func(_ context.Context, _ uint64) (*Attachment, error) {
					return &Attachment{Base: model.Base{ID: 5}, OrganizationID: 7, StorageURL: "attachments/10/abc"}, nil
				}
				return NewAttachmentService(attachments, StoreMock{DeleteFunc: func(_ context.Context, _ string) error {
					return errors.New("storage down")
				}})
			},
			wantErr: true,
		},
		{
			name: "fails when record delete fails",
			svc: func() AttachmentService {
				attachments := &AttachmentDAOMock{}
				attachments.FindFunc = func(_ context.Context, _ uint64) (*Attachment, error) {
					return &Attachment{Base: model.Base{ID: 5}, OrganizationID: 7, StorageURL: "attachments/10/abc"}, nil
				}
				attachments.DeleteFunc = func(_ context.Context, _ uint64) error {
					return errors.New("db down")
				}
				return NewAttachmentService(attachments, StoreMock{})
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc()
			err := svc.Delete(context.Background(), 7, 5)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
		})
	}
}

func TestMessageDAOMock_ListByOwner(t *testing.T) {
	messages := &MessageDAOMock{ListByOwnerFunc: func(_ context.Context, ownerType string, ownerID uint64) ([]*Message, error) {
		return []*Message{{OwnerType: ownerType, OwnerID: ownerID}}, nil
	}}

	items, err := messages.ListByOwner(context.Background(), "purchase_order", 10)

	helper.AssertError(t, err, false, nil)
	if len(items) != 1 || items[0].OwnerID != 10 {
		t.Fatalf("items = %+v, want the configured message", items)
	}
}

func TestStoreMock_DeleteDefaults(t *testing.T) {
	store := StoreMock{}

	err := store.Delete(context.Background(), "attachments/10/abc")

	helper.AssertError(t, err, false, nil)
}

func TestApprovalRequestDAOMock_Defaults(t *testing.T) {
	ctx := context.Background()
	requests := &ApprovalRequestDAOMock{}

	request, err := requests.FindByOwner(ctx, "purchase_order", 10)
	if err != nil || request != nil {
		t.Fatalf("FindByOwner = %+v, %v, want nil", request, err)
	}

	created, err := requests.CreateWithSteps(ctx, &ApprovalRequest{}, []*ApprovalStep{})
	if err != nil || created == nil {
		t.Fatalf("CreateWithSteps = %+v, %v, want request back", created, err)
	}

	updated, err := requests.UpdateTx(ctx, nil, &ApprovalRequest{})
	if err != nil || updated == nil {
		t.Fatalf("UpdateTx = %+v, %v, want request back", updated, err)
	}
}

func TestAttachmentDAOMock_ListByOwner(t *testing.T) {
	ctx := context.Background()
	attachments := &AttachmentDAOMock{ListByOwnerFunc: func(_ context.Context, ownerType string, ownerID uint64) ([]*Attachment, error) {
		return []*Attachment{{OwnerType: ownerType, OwnerID: ownerID}}, nil
	}}

	items, err := attachments.ListByOwner(ctx, "purchase_order", 10)

	helper.AssertError(t, err, false, nil)
	if len(items) != 1 || items[0].OwnerID != 10 {
		t.Fatalf("items = %+v, want the configured attachment", items)
	}
}

func TestAttachmentDAOMock_ListByOwnerDefaults(t *testing.T) {
	attachments := &AttachmentDAOMock{}
	messages := &MessageDAOMock{}

	attachmentItems, err := attachments.ListByOwner(context.Background(), "purchase_order", 10)
	if err != nil || len(attachmentItems) != 0 {
		t.Fatalf("attachment items = %+v, %v, want empty", attachmentItems, err)
	}

	messageItems, err := messages.ListByOwner(context.Background(), "purchase_order", 10)
	if err != nil || len(messageItems) != 0 {
		t.Fatalf("message items = %+v, %v, want empty", messageItems, err)
	}
}

type errReader struct{}

func (errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read failed")
}
