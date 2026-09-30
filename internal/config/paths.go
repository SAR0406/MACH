package config

import (
	"errors"
	"os"
	"path/filepath"
)

const (
	deviceSchemaVersion = "v1"
)

type DeviceConfig struct {
	Version       string `json:"version"`
	DeviceID      string `json:"device_id"`
	DeviceName    string `json:"device_name"`
	PublicKey     string `json:"public_key"`
	PrivateKey    string `json:"private_key"`
	ControllerURL string `json:"controller_url"`
	Token         string `json:"token"`
}

func DeviceHome() (string, error) {
	if v := os.Getenv("MACH_HOME"); v != "" {
		return v, nil
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".mach"), nil
}

func DeviceConfigPath() (string, error) {
	h, err := DeviceHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, deviceSchemaVersion, "device.json"), nil
}

func EnsureDeviceDirs() (string, error) {
	p, err := DeviceConfigPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return "", err
	}
	return p, nil
}

func ControllerHome() (string, error) {
	if v := os.Getenv("MACH_CONTROLLER_HOME"); v != "" {
		return v, nil
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".mach-controller"), nil
}

func ControllerStateDir() (string, error) {
	h, err := ControllerHome()
	if err != nil {
		return "", err
	}
	p := filepath.Join(h, deviceSchemaVersion, "state")
	if err := os.MkdirAll(p, 0o700); err != nil {
		return "", err
	}
	return p, nil
}

func ControllerAuditLogPath() (string, error) {
	d, err := ControllerStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "audit.log"), nil
}

func ControllerStorePath() (string, error) {
	d, err := ControllerStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "store.json"), nil
}

func (c DeviceConfig) Validate() error {
	if c.DeviceID == "" || c.PublicKey == "" || c.PrivateKey == "" {
		return errors.New("device identity is incomplete")
	}
	if c.ControllerURL == "" {
		return errors.New("controller url is missing")
	}
	return nil
}
