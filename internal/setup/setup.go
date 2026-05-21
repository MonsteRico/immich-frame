package setup

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/MonsteRico/immich-frame/internal/config"
)

type Status string

const (
	StatusSetupCodeRequired Status = "setup_code_required"
	StatusConfigured        Status = "configured"
	StatusError             Status = "error"
)

type Manager struct {
	StatePath string
	Now       func() time.Time
	NewCode   func() (string, error)
}

func NewManager(statePath string) *Manager {
	return &Manager{StatePath: statePath, Now: time.Now, NewCode: GenerateCode}
}

func (m *Manager) Ensure() (config.State, error) {
	state, err := config.LoadState(m.StatePath)
	if err != nil {
		return state, err
	}
	if state.SetupComplete {
		state.SetupCode = ""
		state.SetupStatus = string(StatusConfigured)
		return state, config.SaveState(m.StatePath, state)
	}
	if state.SetupCode == "" {
		code, err := m.NewCode()
		if err != nil {
			return state, err
		}
		state.SetupCode = code
	}
	state.SetupStatus = string(StatusSetupCodeRequired)
	if state.UpdatedAt.IsZero() && m.Now != nil {
		state.UpdatedAt = m.Now()
	}
	return state, config.SaveState(m.StatePath, state)
}

func (m *Manager) Claim(code string) (config.State, bool, error) {
	state, err := config.LoadState(m.StatePath)
	if err != nil {
		return state, false, err
	}
	if state.SetupComplete || state.SetupCode == "" || code != state.SetupCode {
		return state, false, nil
	}
	return state, true, nil
}

func (m *Manager) Complete() (config.State, error) {
	state, err := config.LoadState(m.StatePath)
	if err != nil {
		return state, err
	}
	state.SetupComplete = true
	state.SetupCode = ""
	state.SetupStatus = string(StatusConfigured)
	return state, config.SaveState(m.StatePath, state)
}

func GenerateCode() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	n := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if n < 0 {
		n = -n
	}
	return fmt.Sprintf("%06d", n%1000000), nil
}
