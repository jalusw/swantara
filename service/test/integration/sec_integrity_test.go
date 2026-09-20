//go:build integration

package integration

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/crosscutting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/storage"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func newTestAuthService(user *iam.User) iam.AuthService {
	return iam.NewAuthService(
		iam.NewUserDAO(testDB),
		iam.NewUserService(iam.NewUserDAO(testDB), iam.NewPasswordService(testCfg)),
		iam.NewUserSessionDAO(testDB),
		iam.NewPasswordService(testCfg),
		iam.NewTokenService(testCfg),
		iam.NewUserEmailVerificationDAO(testDB),
		iam.NewUserPasswordResetDAO(testDB),
		nil,
		time.Hour,
		time.Hour,
	)
}

func TestPasswordStoredOnlyAsSaltedHash(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	password := "correct horse battery staple"
	userSvc := iam.NewUserService(iam.NewUserDAO(testDB), iam.NewPasswordService(testCfg))
	created, err := userSvc.Create(ctx, &iam.User{
		Username:  gofakeit.Username(),
		FirstName: gofakeit.FirstName(),
		Email:     gofakeit.Email(),
		Password:  password,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	stored, err := iam.NewUserDAO(testDB).Find(ctx, created.ID)
	if err != nil {
		t.Fatalf("find stored user failed: %v", err)
	}
	if stored.Password == password {
		t.Fatal("password stored in plaintext")
	}
	if !strings.HasPrefix(stored.Password, "$2") {
		t.Fatalf("expected bcrypt hash prefix, got %q", stored.Password)
	}
	if strings.Contains(stored.Password, testCfg.Pepper) {
		t.Fatal("pepper must not be stored alongside the hash")
	}

	passwordSvc := iam.NewPasswordService(testCfg)
	if err := passwordSvc.Verify(stored.Password, password); err != nil {
		t.Fatalf("valid password must verify: %v", err)
	}
	if err := passwordSvc.Verify(stored.Password, "wrong-password"); err == nil {
		t.Fatal("wrong password must not verify")
	}
}

func TestRefreshTokenRotationAndReuseDetection(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	user, err := iam.NewUserService(iam.NewUserDAO(testDB), iam.NewPasswordService(testCfg)).Create(ctx, &iam.User{
		Username:  gofakeit.Username(),
		FirstName: gofakeit.FirstName(),
		Email:     gofakeit.Email(),
		Password:  "test-password",
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	authSvc := newTestAuthService(user)
	login, err := authSvc.Login(ctx, user.Email, "test-password", iam.SessionClientInfo{})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if login.AccessTokenExpiresAt.Before(time.Now()) {
		t.Fatal("access token must not already be expired")
	}

	if _, err := authSvc.Refresh(ctx, login.RefreshToken, iam.SessionClientInfo{}); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	if _, err := authSvc.Refresh(ctx, login.RefreshToken, iam.SessionClientInfo{}); err != nil {
		t.Fatalf("concurrent reuse within grace must succeed, got %v", err)
	}

	aged := time.Now().Add(-2 * time.Minute)
	if err := testDB.WithContext(ctx).Table("user_sessions").
		Where("refresh_token = ?", helper.HashToken(login.RefreshToken)).
		Update("revoked_at", aged).Error; err != nil {
		t.Fatalf("age revoked session failed: %v", err)
	}

	if _, err := authSvc.Refresh(ctx, login.RefreshToken, iam.SessionClientInfo{}); !errors.Is(err, iam.ErrTokenReused) {
		t.Fatalf("expected ErrTokenReused for rotated token, got %v", err)
	}

	if _, err := authSvc.Refresh(ctx, "definitely-not-a-token", iam.SessionClientInfo{}); !errors.Is(err, iam.ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestCrossOrganizationIsolation(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	orgA, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create org A failed: %v", err)
	}
	orgB, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create org B failed: %v", err)
	}

	accountA, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: orgA.ID, Code: "1000", Name: "Org A Cash", Type: "cash", Active: true,
	})
	if err != nil {
		t.Fatalf("create account A failed: %v", err)
	}
	accountB, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: orgB.ID, Code: "1000", Name: "Org B Cash", Type: "cash", Active: true,
	})
	if err != nil {
		t.Fatalf("create account B failed: %v", err)
	}

	accountsOfA, err := dao.NewBase[reference.Account](testDB).List(ctx, &query.Query{
		Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: orgA.ID}},
	})
	if err != nil {
		t.Fatalf("list accounts of org A failed: %v", err)
	}
	for _, account := range accountsOfA.Items {
		if account.ID == accountB.ID {
			t.Fatalf("cross-tenant leak: org A saw org B account %d", accountB.ID)
		}
	}
	if len(accountsOfA.Items) != 1 || accountsOfA.Items[0].ID != accountA.ID {
		t.Fatalf("expected only org A account, got %+v", accountsOfA.Items)
	}

	found, err := dao.NewBase[reference.Account](testDB).Find(ctx, accountB.ID)
	if err != nil {
		t.Fatalf("find account B failed: %v", err)
	}
	if found == nil || found.OrganizationID != orgB.ID {
		t.Fatalf("org B account must remain owned by org B, got %+v", found)
	}
}

func TestAttachmentChecksumVerifiedOnDownload(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create org failed: %v", err)
	}
	foreignOrg, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create foreign org failed: %v", err)
	}

	content := []byte("signed document payload")
	svc := crosscutting.NewAttachmentService(
		crosscutting.NewAttachmentDAO(testDB),
		storage.NewLocalStore(t.TempDir()),
	)

	uploaded, err := svc.Upload(ctx, org.ID, "invoice", 42, 7, "invoice.pdf", "application/pdf", bytes.NewReader(content))
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	expected := sha256.Sum256(content)
	if uploaded.Checksum != hex.EncodeToString(expected[:]) {
		t.Fatalf("expected checksum %s, got %s", hex.EncodeToString(expected[:]), uploaded.Checksum)
	}
	if uploaded.ByteSize != int64(len(content)) {
		t.Fatalf("expected byte size %d, got %d", len(content), uploaded.ByteSize)
	}

	attachment, reader, err := svc.Download(ctx, org.ID, uploaded.ID)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}
	defer func() { _ = reader.Close() }()
	downloaded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read downloaded payload failed: %v", err)
	}
	if !bytes.Equal(downloaded, content) {
		t.Fatal("downloaded content does not match uploaded content")
	}
	if attachment.Checksum != uploaded.Checksum {
		t.Fatal("downloaded attachment checksum must match stored checksum")
	}

	if _, _, err := svc.Download(ctx, foreignOrg.ID, uploaded.ID); !errors.Is(err, crosscutting.ErrAttachmentNotFound) {
		t.Fatalf("foreign organization download must be not found, got %v", err)
	}

	if _, err := svc.Upload(ctx, org.ID, "invoice", 42, 7, "empty.pdf", "application/pdf", bytes.NewReader(nil)); !errors.Is(err, crosscutting.ErrAttachmentEmpty) {
		t.Fatalf("expected ErrAttachmentEmpty, got %v", err)
	}
}

func TestSessionRecordsDeviceAndIPMetadata(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	user, err := iam.NewUserService(iam.NewUserDAO(testDB), iam.NewPasswordService(testCfg)).Create(ctx, &iam.User{
		Username:  gofakeit.Username(),
		FirstName: gofakeit.FirstName(),
		Email:     gofakeit.Email(),
		Password:  "test-password",
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	client := iam.SessionClientInfo{
		IPAddress:  "203.0.113.42",
		DeviceName: "Pixel 8",
		OS:         "Android 14",
		Browser:    "Chrome 120",
	}
	authSvc := newTestAuthService(user)
	login, err := authSvc.Login(ctx, user.Email, "test-password", client)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	sessions, err := iam.NewUserSessionDAO(testDB).ListUserSessions(ctx, user.ID)
	if err != nil {
		t.Fatalf("list sessions failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	session := sessions[0]
	if session.IPAddress == nil || *session.IPAddress != client.IPAddress {
		t.Fatalf("expected ip %s, got %+v", client.IPAddress, session.IPAddress)
	}
	if session.DeviceName == nil || *session.DeviceName != client.DeviceName {
		t.Fatalf("expected device %s, got %+v", client.DeviceName, session.DeviceName)
	}
	if session.OS == nil || *session.OS != client.OS {
		t.Fatalf("expected os %s, got %+v", client.OS, session.OS)
	}
	if session.Browser == nil || *session.Browser != client.Browser {
		t.Fatalf("expected browser %s, got %+v", client.Browser, session.Browser)
	}
	if session.ExpiresAt == nil || session.ExpiresAt.Before(time.Now()) {
		t.Fatalf("expected future session expiry, got %+v", session.ExpiresAt)
	}
	if session.RefreshToken == nil || *session.RefreshToken == login.RefreshToken {
		t.Fatal("refresh token must be stored hashed, not in plaintext")
	}
	if session.RefreshToken == nil || *session.RefreshToken != helper.HashToken(login.RefreshToken) {
		t.Fatal("stored refresh token must equal the expected digest")
	}
}
