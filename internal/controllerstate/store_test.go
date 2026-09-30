package controllerstate

import (
	"net"
	"os"
	"testing"

	"github.com/SAR0406/MACH/internal/api"
	"github.com/SAR0406/MACH/internal/identity"
)

func TestStorePersistsState(t *testing.T) {
	t.Setenv("MACH_CONTROLLER_HOME", t.TempDir())

	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	pub, priv, err := identity.GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair: %v", err)
	}
	id, err := identity.DeviceIDFromPublicKey(pub)
	if err != nil {
		t.Fatalf("DeviceIDFromPublicKey: %v", err)
	}

	challenge, err := store.RegisterChallenge(api.RegisterChallengeRequest{
		DeviceID:   id,
		DeviceName: "dev",
		PublicKey:  pub,
	})
	if err != nil {
		t.Fatalf("RegisterChallenge: %v", err)
	}
	sig, err := identity.Sign(priv, []byte(challenge))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	token, err := store.RegisterComplete(api.RegisterCompleteRequest{DeviceID: id, Signature: sig})
	if err != nil {
		t.Fatalf("RegisterComplete: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	if err := store.Heartbeat(id, api.HeartbeatRequest{ReachableHost: "127.0.0.1"}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	ex, err := store.Expose(id, port, "http://127.0.0.1:8080")
	if err != nil {
		t.Fatalf("Expose: %v", err)
	}
	if ex.Path != api.NetworkPathDirect {
		t.Fatalf("expected direct path, got %s", ex.Path)
	}

	store2, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore reload: %v", err)
	}
	if _, ok := store2.Authenticate(token); !ok {
		t.Fatal("expected persisted token to authenticate")
	}
	st, err := store2.Status(id)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(st.Exposures) != 1 {
		t.Fatalf("expected one exposure, got %d", len(st.Exposures))
	}
}

func TestStoreStatusBeforeHeartbeat(t *testing.T) {
	t.Setenv("MACH_CONTROLLER_HOME", t.TempDir())
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	pub, priv, err := identity.GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair: %v", err)
	}
	id, err := identity.DeviceIDFromPublicKey(pub)
	if err != nil {
		t.Fatalf("DeviceIDFromPublicKey: %v", err)
	}
	ch, err := store.RegisterChallenge(api.RegisterChallengeRequest{
		DeviceID:  id,
		PublicKey: pub,
	})
	if err != nil {
		t.Fatalf("RegisterChallenge: %v", err)
	}
	sig, err := identity.Sign(priv, []byte(ch))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if _, err := store.RegisterComplete(api.RegisterCompleteRequest{DeviceID: id, Signature: sig}); err != nil {
		t.Fatalf("RegisterComplete: %v", err)
	}

	st, err := store.Status(id)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.DeviceState != api.DeviceStateClaimed {
		t.Fatalf("expected claimed state, got %s", st.DeviceState)
	}
	if st.NetworkPath != api.NetworkPathUnknown {
		t.Fatalf("expected unknown path, got %s", st.NetworkPath)
	}

	if _, err := os.Stat(os.Getenv("MACH_CONTROLLER_HOME")); err != nil {
		t.Fatalf("controller home missing: %v", err)
	}
}
