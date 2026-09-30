package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadDeviceConfig() (*DeviceConfig, error) {
	p, err := DeviceConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var cfg DeviceConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid device config: %w", err)
	}
	return &cfg, nil
}

func SaveDeviceConfig(cfg DeviceConfig) error {
	p, err := EnsureDeviceDirs()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o600)
}
