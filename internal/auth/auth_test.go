package auth

import (
	"testing"
	"time"
)

func TestPasswordHashVerifiesWithoutStoringPlainText(t *testing.T) {
	hash, err := HashPassword("correct horse")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "correct horse" {
		t.Fatal("hash should not store the plain password")
	}
	if !VerifyPassword(hash, "correct horse") {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword(hash, "wrong horse") {
		t.Fatal("wrong password should not verify")
	}
}

func TestSessionsRenewAndExpire(t *testing.T) {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	manager := NewManager(30 * time.Minute)
	manager.now = func() time.Time { return now }
	manager.entropy = func(buf []byte) (int, error) {
		for i := range buf {
			buf[i] = byte(i + 1)
		}
		return len(buf), nil
	}

	session, err := manager.Create(ScopeAdmin)
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(20 * time.Minute)
	renewed, ok := manager.Validate(session.Token, ScopeAdmin)
	if !ok {
		t.Fatal("session should validate before ttl")
	}
	if !renewed.ExpiresAt.Equal(now.Add(30 * time.Minute)) {
		t.Fatalf("session expiry = %s", renewed.ExpiresAt)
	}
	if _, ok := manager.Validate(session.Token, ScopeSetup); ok {
		t.Fatal("admin session should not validate for setup-only scope")
	}
	now = now.Add(31 * time.Minute)
	if _, ok := manager.Validate(session.Token, ScopeAdmin); ok {
		t.Fatal("session should expire")
	}
}
