package api

import "time"

const (
	NetworkPathUnknown = "UNKNOWN"
	NetworkPathDirect  = "DIRECT"
	NetworkPathRelay   = "RELAY"
)

const (
	DeviceStateClaimed    = "CLAIMED"
	DeviceStateOnline     = "ONLINE"
	DeviceStateOffline    = "OFFLINE"
	DeviceStateRecovering = "RECOVERING"
)

const (
	ExposureStatusPending = "PENDING"
	ExposureStatusActive  = "ACTIVE"
	ExposureStatusStopped = "STOPPED"
)

type CapabilitySnapshot struct {
	CPUCores   int    `json:"cpu_cores"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	MemoryMiB  int    `json:"memory_mib"`
	StorageGiB int    `json:"storage_gib"`
}

type RegisterChallengeRequest struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	PublicKey  string `json:"public_key"`
}

type RegisterChallengeResponse struct {
	Challenge string `json:"challenge"`
}

type RegisterCompleteRequest struct {
	DeviceID  string `json:"device_id"`
	Signature string `json:"signature"`
}

type RegisterCompleteResponse struct {
	Token string `json:"token"`
}

type HeartbeatRequest struct {
	ReachableHost string             `json:"reachable_host"`
	Capabilities  CapabilitySnapshot `json:"capabilities"`
	AgentVersion  string             `json:"agent_version"`
}

type Exposure struct {
	ID        string `json:"id"`
	Port      int    `json:"port"`
	Path      string `json:"path"`
	PublicURL string `json:"public_url"`
	RelayURL  string `json:"relay_url"`
	Status    string `json:"status"`
}

type ExposeRequest struct {
	Port int `json:"port"`
}

type StopRequest struct {
	ExposureID string `json:"exposure_id"`
}

type DeviceStatusResponse struct {
	DeviceID      string             `json:"device_id"`
	DeviceName    string             `json:"device_name"`
	Online        bool               `json:"online"`
	DeviceState   string             `json:"device_state"`
	NetworkPath   string             `json:"network_path"`
	LastHeartbeat time.Time          `json:"last_heartbeat"`
	Capabilities  CapabilitySnapshot `json:"capabilities"`
	Exposures     []Exposure         `json:"exposures"`
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	DeviceID  string    `json:"device_id"`
	Message   string    `json:"message"`
}

type LogsResponse struct {
	Entries []LogEntry `json:"entries"`
}
