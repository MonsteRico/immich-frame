package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"strings"
	"sync"
	"time"
)

const (
	hashPrefix       = "pbkdf2-sha256"
	defaultRounds    = 210000
	defaultSaltBytes = 16
	defaultKeyBytes  = 32
)

var ErrInvalidHash = errors.New("invalid password hash")

type Scope string

const (
	ScopeSetup Scope = "setup"
	ScopeAdmin Scope = "admin"
)

type Session struct {
	Token     string
	Scope     Scope
	ExpiresAt time.Time
}

type Manager struct {
	mu      sync.Mutex
	now     func() time.Time
	ttl     time.Duration
	entropy func([]byte) (int, error)
	byToken map[string]Session
}

func NewManager(ttl time.Duration) *Manager {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	return &Manager{
		now:     time.Now,
		ttl:     ttl,
		entropy: rand.Read,
		byToken: map[string]Session{},
	}
}

func (m *Manager) Create(scope Scope) (Session, error) {
	token, err := randomToken(m.entropy, 32)
	if err != nil {
		return Session{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	session := Session{Token: token, Scope: scope, ExpiresAt: m.now().Add(m.ttl)}
	m.byToken[token] = session
	return session, nil
}

func (m *Manager) Validate(token string, scopes ...Scope) (Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.byToken[token]
	if !ok || !m.now().Before(session.ExpiresAt) {
		delete(m.byToken, token)
		return Session{}, false
	}
	if len(scopes) > 0 && !scopeAllowed(session.Scope, scopes) {
		return Session{}, false
	}
	session.ExpiresAt = m.now().Add(m.ttl)
	m.byToken[token] = session
	return session, true
}

func (m *Manager) Destroy(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.byToken, token)
}

func scopeAllowed(scope Scope, allowed []Scope) bool {
	for _, candidate := range allowed {
		if scope == candidate {
			return true
		}
	}
	return false
}

func HashPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if len(password) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}
	salt := make([]byte, defaultSaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := pbkdf2Key([]byte(password), salt, defaultRounds, defaultKeyBytes, sha256.New)
	return fmt.Sprintf("%s$%d$%s$%s",
		hashPrefix,
		defaultRounds,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func VerifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != hashPrefix {
		return false
	}
	var rounds int
	if _, err := fmt.Sscanf(parts[1], "%d", &rounds); err != nil || rounds <= 0 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(want) == 0 {
		return false
	}
	got := pbkdf2Key([]byte(strings.TrimSpace(password)), salt, rounds, len(want), sha256.New)
	return subtle.ConstantTimeCompare(got, want) == 1
}

func randomToken(entropy func([]byte) (int, error), n int) (string, error) {
	buf := make([]byte, n)
	if _, err := entropy(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func pbkdf2Key(password, salt []byte, iter, keyLen int, h func() hash.Hash) []byte {
	prf := hmac.New(h, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen
	var out []byte
	var block [4]byte
	for i := 1; i <= numBlocks; i++ {
		block[0] = byte(i >> 24)
		block[1] = byte(i >> 16)
		block[2] = byte(i >> 8)
		block[3] = byte(i)
		prf.Reset()
		_, _ = prf.Write(salt)
		_, _ = prf.Write(block[:])
		u := prf.Sum(nil)
		t := append([]byte(nil), u...)
		for j := 1; j < iter; j++ {
			prf.Reset()
			_, _ = prf.Write(u)
			u = prf.Sum(nil)
			for k := range t {
				t[k] ^= u[k]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}
