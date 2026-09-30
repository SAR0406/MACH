package proto

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type VPSManifest struct {
	Name string `yaml:"name"`

	Runtime struct {
		Type  string `yaml:"type"`
		Image string `yaml:"image,omitempty"`
		Exec  string `yaml:"exec,omitempty"`
	} `yaml:"runtime"`

	Network struct {
		HTTP struct {
			Port int `yaml:"port"`
		} `yaml:"http"`
	} `yaml:"network"`

	Restart struct {
		Policy string `yaml:"policy"`
	} `yaml:"restart"`

	Health struct {
		Command string `yaml:"command"`
	} `yaml:"health"`
}

func ParseManifest(content []byte) (*VPSManifest, error) {
	var m VPSManifest
	if err := yaml.Unmarshal(content, &m); err != nil {
		return nil, err
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

func (m VPSManifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("name is required")
	}
	if m.Runtime.Type == "" {
		return fmt.Errorf("runtime.type is required")
	}
	if m.Network.HTTP.Port <= 0 || m.Network.HTTP.Port > 65535 {
		return fmt.Errorf("network.http.port must be valid")
	}
	if m.Restart.Policy == "" {
		return fmt.Errorf("restart.policy is required")
	}
	return nil
}
