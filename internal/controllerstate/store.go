package controllerstate

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/SAR0406/MACH/internal/api"
	"github.com/SAR0406/MACH/internal/config"
	"github.com/SAR0406/MACH/internal/identity"
)

type Device struct {
	ID            string
	Name          string
	PublicKey     string
	Token         string
	Challenge     string
	LastHeartbeat time.Time
	ReachableHost string
	Capabilities  api.CapabilitySnapshot
	Exposures     map[string]api.Exposure
	NetworkPath   string
}

type Store struct {
mu       sync.RWMutex
devices  map[string]*Device
tokenIdx map[string]string
auditLog string
state    string
}

func NewStore() (*Store, error) {
audit, err := config.ControllerAuditLogPath()
if err != nil {
	return nil, err
}
statePath, err := config.ControllerStorePath()
if err != nil {
	return nil, err
}
s := &Store{
	devices:  map[string]*Device{},
	tokenIdx: map[string]string{},
	auditLog: audit,
	state:    statePath,
}
if err := s.load(); err != nil {
	return nil, err
}
return s, nil

func (s *Store) RegisterChallenge(req api.RegisterChallengeRequest) (string, error) {
	if req.DeviceID == "" || req.PublicKey == "" {
		return "", fmt.Errorf("device_id and public_key are required")
	}
	challenge, err := randomHex(32)
	if err != nil {
		return "", err
	}
	name := req.DeviceName
	if name == "" {
		name = req.DeviceID
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	dev, ok := s.devices[req.DeviceID]
	if !ok {
		dev = &Device{ID: req.DeviceID, Name: name, PublicKey: req.PublicKey, Exposures: map[string]api.Exposure{}}
		s.devices[req.DeviceID] = dev
	} else {
		dev.Name = name
		dev.PublicKey = req.PublicKey
	}
	dev.Challenge = challenge
	s.audit("claim_challenge", req.DeviceID, "claim challenge issued")
	_ = s.saveLocked()
	return challenge, nil
}

func (s *Store) RegisterComplete(req api.RegisterCompleteRequest) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	dev, ok := s.devices[req.DeviceID]
	if !ok {
		return "", fmt.Errorf("device not found")
	}
	valid, err := identity.Verify(dev.PublicKey, []byte(dev.Challenge), req.Signature)
	if err != nil {
		return "", err
	}
	if !valid {
		return "", fmt.Errorf("invalid signature")
	}
	token, err := randomHex(32)
	if err != nil {
		return "", err
	}
	dev.Token = token
	s.tokenIdx[token] = dev.ID
	s.audit("claim_complete", req.DeviceID, "device claimed")
	_ = s.saveLocked()
	return token, nil
}

func (s *Store) Authenticate(token string) (*Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.tokenIdx[token]
	if !ok {
		return nil, false
	}
	dev, ok := s.devices[id]
	if !ok {
		return nil, false
	}
	copy := *dev
	copy.Exposures = map[string]api.Exposure{}
	for k, v := range dev.Exposures {
		copy.Exposures[k] = v
	}
	return &copy, true
}

func (s *Store) Heartbeat(deviceID string, hb api.HeartbeatRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	dev, ok := s.devices[deviceID]
	if !ok {
		return fmt.Errorf("device not found")
	}
	dev.LastHeartbeat = time.Now().UTC()
	dev.ReachableHost = hb.ReachableHost
	if dev.ReachableHost == "" {
		dev.ReachableHost = "127.0.0.1"
	}
	dev.Capabilities = hb.Capabilities
	s.audit("heartbeat", deviceID, "device heartbeat received")
	_ = s.saveLocked()
	return nil
}

func (s *Store) Expose(deviceID string, port int, publicBase string) (api.Exposure, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	dev, ok := s.devices[deviceID]
	if !ok {
		return api.Exposure{}, fmt.Errorf("device not found")
	}
	id, err := randomHex(3)
	if err != nil {
		return api.Exposure{}, err
	}
	path := api.NetworkPathRelay
	if directReachable(dev.ReachableHost, port) {
		path = api.NetworkPathDirect
	}
	ex := api.Exposure{
		ID:        id,
		Port:      port,
		Path:      path,
		PublicURL: fmt.Sprintf("https://%s.mach.dev", id),
		RelayURL:  fmt.Sprintf("%s/p/%s", trimSlash(publicBase), id),
		Status:    api.ExposureStatusActive,
	}
	dev.Exposures[id] = ex
	dev.NetworkPath = path
	s.audit("expose", deviceID, fmt.Sprintf("port %d exposed as %s", port, id))
	_ = s.saveLocked()
	return ex, nil
}

func (s *Store) Stop(deviceID string, exposureID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	dev, ok := s.devices[deviceID]
	if !ok {
		return fmt.Errorf("device not found")
	}
	if exposureID == "" {
		dev.Exposures = map[string]api.Exposure{}
		s.audit("stop", deviceID, "all exposures stopped")
		_ = s.saveLocked()
		return nil
	}
	if _, ok := dev.Exposures[exposureID]; !ok {
		return fmt.Errorf("exposure not found")
	}
	delete(dev.Exposures, exposureID)
	s.audit("stop", deviceID, fmt.Sprintf("exposure %s stopped", exposureID))
	_ = s.saveLocked()
	return nil
}

func (s *Store) Status(deviceID string) (api.DeviceStatusResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	dev, ok := s.devices[deviceID]
	if !ok {
		return api.DeviceStatusResponse{}, fmt.Errorf("device not found")
	}
	online := time.Since(dev.LastHeartbeat) < 20*time.Second
	netPath := dev.NetworkPath
	if netPath == "" {
		netPath = api.NetworkPathUnknown
	}
	deviceState := api.DeviceStateClaimed
	switch {
	case online:
		deviceState = api.DeviceStateOnline
	case !dev.LastHeartbeat.IsZero():
		deviceState = api.DeviceStateRecovering
	default:
		deviceState = api.DeviceStateOffline
	}
	exposures := make([]api.Exposure, 0, len(dev.Exposures))
	for _, ex := range dev.Exposures {
		exposures = append(exposures, ex)
	}
	sort.Slice(exposures, func(i, j int) bool {
		return exposures[i].ID < exposures[j].ID
	})
	return api.DeviceStatusResponse{
		DeviceID:      dev.ID,
		DeviceName:    dev.Name,
		Online:        online,
		DeviceState:   deviceState,
		NetworkPath:   netPath,
		LastHeartbeat: dev.LastHeartbeat,
		Capabilities:  dev.Capabilities,
		Exposures:     exposures,
	}, nil
}

func (s *Store) ResolveExposure(exposureID string) (host string, port int, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, dev := range s.devices {
		ex, found := dev.Exposures[exposureID]
		if !found {
			continue
		}
		if time.Since(dev.LastHeartbeat) > 20*time.Second {
			return "", 0, false
		}
		host := dev.ReachableHost
		if host == "" {
			host = "127.0.0.1"
		}
		return host, ex.Port, true
	}
	return "", 0, false
}

func (s *Store) Logs(deviceID string, limit int) ([]api.LogEntry, error) {
	b, err := os.ReadFile(s.auditLog)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := bytesSplitLines(b)
	out := make([]api.LogEntry, 0, len(lines))
	for _, ln := range lines {
		if len(ln) == 0 {
			continue
		}
		var e api.LogEntry
		if err := json.Unmarshal(ln, &e); err != nil {
			continue
		}
		if e.DeviceID == deviceID {
			out = append(out, e)
		}
	}
	if limit <= 0 || len(out) <= limit {
		return out, nil
	}
	return out[len(out)-limit:], nil
}

func (s *Store) audit(action, deviceID, message string) {
	entry := api.LogEntry{Timestamp: time.Now().UTC(), Action: action, DeviceID: deviceID, Message: message}
	b, _ := json.Marshal(entry)
	f, err := os.OpenFile(s.auditLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(b, '\n'))
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

func directReachable(host string, port int) bool {
	if host == "" {
		host = "127.0.0.1"
	}
	c, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 1200*time.Millisecond)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}

func bytesSplitLines(b []byte) [][]byte {
	start := 0
	out := make([][]byte, 0, 16)
	for i, ch := range b {
		if ch == '\n' {
			out = append(out, b[start:i])
			start = i + 1
		}

		type persistentStore struct {
			Devices map[string]*Device `json:"devices"`
		}

		func (s *Store) load() error {
			b, err := os.ReadFile(s.state)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			var p persistentStore
			if err := json.Unmarshal(b, &p); err != nil {
				return err
			}
			if p.Devices == nil {
				p.Devices = map[string]*Device{}
			}
			s.devices = p.Devices
			s.tokenIdx = map[string]string{}
			for id, dev := range s.devices {
				if dev.Exposures == nil {
					dev.Exposures = map[string]api.Exposure{}
				}
				if dev.Token != "" {
					s.tokenIdx[dev.Token] = id
				}
			}
			return nil
		}

		func (s *Store) saveLocked() error {
			if err := os.MkdirAll(filepath.Dir(s.state), 0o700); err != nil {
				return err
			}
			p := persistentStore{Devices: s.devices}
			b, err := json.MarshalIndent(p, "", "  ")
			if err != nil {
				return err
			}
			return os.WriteFile(s.state, b, 0o600)
		}
	}
	if start < len(b) {
		out = append(out, b[start:])
	}
	return out
}
