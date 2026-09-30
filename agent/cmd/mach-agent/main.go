package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/SAR0406/MACH/internal/api"
	"github.com/SAR0406/MACH/internal/config"
)

const agentVersion = "0.1.0"

func main() {
	cfg, err := config.LoadDeviceConfig()
	if err != nil {
		log.Fatalf("failed to load device config: %v", err)
	}
	client := api.NewClient(cfg.ControllerURL, cfg.Token)

	reachableHost := envOr("MACH_REACHABLE_HOST", "127.0.0.1")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Printf("mach-agent running for device %s", cfg.DeviceID)
	for {
		if err := client.Heartbeat(api.HeartbeatRequest{
			ReachableHost: reachableHost,
			AgentVersion:  agentVersion,
			Capabilities: api.CapabilitySnapshot{
				CPUCores:   runtime.NumCPU(),
				OS:         runtime.GOOS,
				Arch:       runtime.GOARCH,
				MemoryMiB:  approximateMemMiB(),
				StorageGiB: 0,
			},
		}); err != nil {
			log.Printf("heartbeat failed: %v", err)
		} else {
			log.Printf("heartbeat ok")
		}

		status, err := client.Status()
		if err == nil {
			for _, ex := range status.Exposures {
				if !localPortOpen(ex.Port) {
					log.Printf("service check failed: exposure=%s port=%d", ex.ID, ex.Port)
				}
			}
		}

		<-ticker.C
	}
}

func approximateMemMiB() int {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if m.Sys == 0 {
		return 0
	}
	return int(m.Sys / 1024 / 1024)
}

func localPortOpen(port int) bool {
	c, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 800*time.Millisecond)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
