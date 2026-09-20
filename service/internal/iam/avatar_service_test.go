package iam

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type avatarStoreMock struct {
	saved   map[string][]byte
	opened  map[string][]byte
	deleted []string
}

func newAvatarStoreMock() *avatarStoreMock {
	return &avatarStoreMock{saved: map[string][]byte{}, opened: map[string][]byte{}}
}

func (m *avatarStoreMock) Save(_ context.Context, id string, reader io.Reader) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	m.saved[id] = content
	m.opened[id] = content
	return nil
}

func (m *avatarStoreMock) Open(_ context.Context, id string) (io.ReadCloser, error) {
	content, ok := m.opened[id]
	if !ok {
		return nil, errors.New("object not found")
	}
	return io.NopCloser(bytes.NewReader(content)), nil
}

func (m *avatarStoreMock) Delete(_ context.Context, id string) error {
	m.deleted = append(m.deleted, id)
	delete(m.saved, id)
	delete(m.opened, id)
	return nil
}

func testPng() []byte {
	return []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
}

func avatarObjectID(userID uint64, content []byte) string {
	checksum := sha256.Sum256(content)
	return "avatars/" + strconv.FormatUint(userID, 10) + "/" + hex.EncodeToString(checksum[:])
}

func TestAvatarUploadStoresObjectAndUpdatesUser(t *testing.T) {
	store := newAvatarStoreMock()
	var updated *User
	users := UserDAOMock{
		DAOMock: DAOMock[User]{
			FindFunc: func(_ context.Context, _ uint64) (*User, error) {
				return &User{Base: model.Base{ID: 7}, Avatar: helper.Ptr("avatars/7/old")}, nil
			},
			UpdateFunc: func(_ context.Context, entity *User) (*User, error) {
				clone := *entity
				updated = &clone
				return entity, nil
			},
		},
	}
	svc := NewAvatarService(users, store)

	expected := avatarObjectID(7, testPng())
	objectID, err := svc.Upload(context.Background(), 7, bytes.NewReader(testPng()))

	helper.AssertError(t, err, false, nil)
	if objectID != expected {
		t.Errorf("object id = %q, want %q", objectID, expected)
	}
	if updated == nil || updated.Avatar == nil || *updated.Avatar != expected {
		t.Errorf("user avatar = %v, want %q", updated, expected)
	}
	if _, ok := store.saved[expected]; !ok {
		t.Errorf("final object %q was not stored", expected)
	}
	if len(store.deleted) != 1 || store.deleted[0] != "avatars/7/old" {
		t.Errorf("deleted = %v, want old avatar removed", store.deleted)
	}
}

func TestAvatarUploadRejectsEmptyContent(t *testing.T) {
	svc := NewAvatarService(UserDAOMock{}, newAvatarStoreMock())

	_, err := svc.Upload(context.Background(), 7, bytes.NewReader(nil))

	helper.AssertError(t, err, true, ErrAvatarEmpty)
}

func TestAvatarUploadRejectsNonImageContent(t *testing.T) {
	svc := NewAvatarService(UserDAOMock{}, newAvatarStoreMock())

	_, err := svc.Upload(context.Background(), 7, bytes.NewReader([]byte("just plain text")))

	helper.AssertError(t, err, true, ErrAvatarInvalidType)
}

func TestAvatarUploadRejectsMissingUser(t *testing.T) {
	users := UserDAOMock{DAOMock: DAOMock[User]{FindFunc: func(_ context.Context, _ uint64) (*User, error) {
		return nil, nil
	}}}
	svc := NewAvatarService(users, newAvatarStoreMock())

	_, err := svc.Upload(context.Background(), 7, bytes.NewReader(testPng()))

	helper.AssertError(t, err, true, ErrUserNotFound)
}

func TestAvatarUploadRemovesStoredObjectWhenUpdateFails(t *testing.T) {
	store := newAvatarStoreMock()
	users := UserDAOMock{
		DAOMock: DAOMock[User]{
			FindFunc: func(_ context.Context, _ uint64) (*User, error) {
				return &User{Base: model.Base{ID: 7}}, nil
			},
			UpdateFunc: func(_ context.Context, _ *User) (*User, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewAvatarService(users, store)

	_, err := svc.Upload(context.Background(), 7, bytes.NewReader(testPng()))

	helper.AssertError(t, err, true, nil)
	if len(store.saved) != 0 {
		t.Errorf("stored objects = %d, want 0 after rollback", len(store.saved))
	}
}

func TestAvatarUploadSkipsWorkWhenChecksumUnchanged(t *testing.T) {
	existing := avatarObjectID(7, testPng())
	updateCalled := false
	users := UserDAOMock{
		DAOMock: DAOMock[User]{
			FindFunc: func(_ context.Context, _ uint64) (*User, error) {
				return &User{Base: model.Base{ID: 7}, Avatar: helper.Ptr(existing)}, nil
			},
			UpdateFunc: func(_ context.Context, _ *User) (*User, error) {
				updateCalled = true
				return nil, nil
			},
		},
	}
	svc := NewAvatarService(users, newAvatarStoreMock())

	objectID, err := svc.Upload(context.Background(), 7, bytes.NewReader(testPng()))

	helper.AssertError(t, err, false, nil)
	if objectID != existing {
		t.Errorf("object id = %q, want %q", objectID, existing)
	}
	if updateCalled {
		t.Error("expected no user update when checksum is unchanged")
	}
}

func TestAvatarDownloadReturnsStoredObject(t *testing.T) {
	users := UserDAOMock{DAOMock: DAOMock[User]{FindFunc: func(_ context.Context, _ uint64) (*User, error) {
		return &User{Base: model.Base{ID: 7}, Avatar: helper.Ptr("avatars/7/abc")}, nil
	}}}
	store := newAvatarStoreMock()
	store.opened["avatars/7/abc"] = testPng()
	svc := NewAvatarService(users, store)

	reader, err := svc.Download(context.Background(), 7)

	helper.AssertError(t, err, false, nil)
	if reader == nil {
		t.Fatal("expected a reader")
	}
	content, _ := io.ReadAll(reader)
	if !bytes.Equal(content, testPng()) {
		t.Errorf("content = %q, want png bytes", content)
	}
	_ = reader.Close()
}

func TestAvatarDownloadRejectsMissingAvatar(t *testing.T) {
	users := UserDAOMock{DAOMock: DAOMock[User]{FindFunc: func(_ context.Context, _ uint64) (*User, error) {
		return &User{Base: model.Base{ID: 7}}, nil
	}}}
	svc := NewAvatarService(users, newAvatarStoreMock())

	_, err := svc.Download(context.Background(), 7)

	helper.AssertError(t, err, true, ErrAvatarNotFound)
}
