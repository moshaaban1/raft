// Package config
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NodeID             string            `yaml:"node_id"`
	Port               string            `yaml:"port"`
	Peers              map[string]string `yaml:"peers"`
	ElectionTimeoutMin int               `yaml:"election_timeout_min"`
	ElectionTimeoutMax int               `yaml:"election_timeout_max"`
	HeartbeatInterval  time.Duration     `yaml:"heartbeat_interval"`
}

func Load(filepath string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.NodeID == "" {
		return fmt.Errorf("node_id is required")
	}
	if len(c.Peers) == 0 {
		return fmt.Errorf("peers list cannot be empty")
	}
	if _, exists := c.Peers[c.NodeID]; !exists {
		return fmt.Errorf("node_id must exist in peers list")
	}
	return nil
}

// Get peers excluding self
func (c *Config) GetOtherPeers() map[string]string {
	others := make(map[string]string)
	for id, addr := range c.Peers {
		if id != c.NodeID {
			others[id] = addr
		}
	}
	return others
}
