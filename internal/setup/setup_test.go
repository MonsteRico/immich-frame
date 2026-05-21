package setup

import (
	"path/filepath"
	"testing"

	"github.com/MonsteRico/immich-frame/internal/config"
)

func TestEnsureGeneratesFixedSetupCodeUntilComplete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	manager := NewManager(path)
	manager.NewCode = func() (string, error) { return "123456", nil }

	first, err := manager.Ensure()
	if err != nil {
		t.Fatal(err)
	}
	if first.SetupCode != "123456" || first.SetupStatus != string(StatusSetupCodeRequired) {
		t.Fatalf("unexpected initial state: %+v", first)
	}

	manager.NewCode = func() (string, error) { return "999999", nil }
	second, err := manager.Ensure()
	if err != nil {
		t.Fatal(err)
	}
	if second.SetupCode != "123456" {
		t.Fatalf("setup code changed before completion: %q", second.SetupCode)
	}

	done, err := manager.Complete()
	if err != nil {
		t.Fatal(err)
	}
	if !done.SetupComplete || done.SetupCode != "" || done.SetupStatus != string(StatusConfigured) {
		t.Fatalf("unexpected completed state: %+v", done)
	}
}

func TestClaimRequiresMatchingActiveCode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := config.SaveState(path, config.State{SetupCode: "123456", SetupStatus: string(StatusSetupCodeRequired)}); err != nil {
		t.Fatal(err)
	}
	manager := NewManager(path)
	if _, ok, err := manager.Claim("000000"); err != nil || ok {
		t.Fatalf("wrong code ok=%t err=%v", ok, err)
	}
	if _, ok, err := manager.Claim("123456"); err != nil || !ok {
		t.Fatalf("right code ok=%t err=%v", ok, err)
	}
}
