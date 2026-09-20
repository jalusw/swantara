package iam

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/storage"
)

type AvatarService struct {
	userDAO UserDAO
	store   storage.Store
}

func NewAvatarService(userDAO UserDAO, store storage.Store) AvatarService {
	return AvatarService{userDAO: userDAO, store: store}
}

func (s AvatarService) Upload(ctx context.Context, userID uint64, reader io.Reader) (string, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	if len(content) == 0 {
		return "", ErrAvatarEmpty
	}
	if !helper.IsImage(content) {
		return "", ErrAvatarInvalidType
	}

	user, err := s.userDAO.Find(ctx, userID)
	if err != nil {
		return "", err
	}
	if user == nil || user.ID == 0 {
		return "", ErrUserNotFound
	}

	objectID := s.objectID(userID, content)
	if user.Avatar != nil && *user.Avatar == objectID {
		return objectID, nil
	}

	if err := s.store.Save(ctx, objectID, bytes.NewReader(content)); err != nil {
		return "", err
	}

	oldAvatar := user.Avatar
	user.Avatar = &objectID
	if _, err := s.userDAO.Update(ctx, user); err != nil {
		_ = s.store.Delete(ctx, objectID)
		return "", err
	}

	if oldAvatar != nil && *oldAvatar != "" {
		_ = s.store.Delete(ctx, *oldAvatar)
	}
	return objectID, nil
}

func (s AvatarService) Download(ctx context.Context, userID uint64) (io.ReadCloser, error) {
	user, err := s.userDAO.Find(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.ID == 0 || user.Avatar == nil || *user.Avatar == "" {
		return nil, ErrAvatarNotFound
	}
	reader, err := s.store.Open(ctx, *user.Avatar)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrAvatarNotFound
		}
		return nil, err
	}
	return reader, nil
}

func (s AvatarService) objectID(userID uint64, content []byte) string {
	checksum := sha256.Sum256(content)
	return fmt.Sprintf("avatars/%d/%s", userID, hex.EncodeToString(checksum[:]))
}
