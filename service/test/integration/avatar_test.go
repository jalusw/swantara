//go:build integration

package integration

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/storage"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestAvatarUploadDownloadIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	userSvc := iam.NewUserService(iam.NewUserDAO(testDB), iam.NewPasswordService(testCfg))
	user, err := userSvc.Create(ctx, &iam.User{
		Username:  gofakeit.Username(),
		FirstName: gofakeit.FirstName(),
		Email:     gofakeit.Email(),
		Password:  "test-password",
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	root := t.TempDir()
	avatarSvc := iam.NewAvatarService(iam.NewUserDAO(testDB), storage.NewLocalStore(root))

	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	objectID, err := avatarSvc.Upload(ctx, user.ID, bytes.NewReader(png))
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	checksum := sha256.Sum256(png)
	expected := "avatars/" + strconv.FormatUint(user.ID, 10) + "/" + hex.EncodeToString(checksum[:])
	if objectID != expected {
		t.Fatalf("object id = %q, want %q", objectID, expected)
	}

	loaded, err := iam.NewUserDAO(testDB).Find(ctx, user.ID)
	if err != nil || loaded == nil {
		t.Fatalf("find user failed: %v", err)
	}
	if loaded.Avatar == nil || *loaded.Avatar != expected {
		t.Fatalf("avatar = %v, want %q", loaded.Avatar, expected)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(expected))); err != nil {
		t.Fatalf("avatar object missing from storage: %v", err)
	}

	reader, err := avatarSvc.Download(ctx, user.ID)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}
	content, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		t.Fatalf("read avatar failed: %v", err)
	}
	if !bytes.Equal(content, png) {
		t.Error("downloaded content differs from uploaded image")
	}

	replacement := append([]byte(nil), png...)
	replacement = append(replacement, 0x00)
	oldObjectID := objectID
	newObjectID, err := avatarSvc.Upload(ctx, user.ID, bytes.NewReader(replacement))
	if err != nil {
		t.Fatalf("second upload failed: %v", err)
	}
	if newObjectID == oldObjectID {
		t.Fatal("expected a different object id for different content")
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(oldObjectID))); !os.IsNotExist(err) {
		t.Error("old avatar object was not removed after replacement")
	}
}
