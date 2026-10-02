package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	DeviceID   string          `json:"device_id"`
	Interval   Duration        `json:"collection_interval"`
	Collectors CollectorConfig `json:"collectors"`
	SyslogPath string          `json:"syslog_path"`
	Export     ExportConfig    `json:"export"`
}

type CollectorConfig struct {
	Conntrack  bool `json:"conntrack"`
	Interfaces bool `json:"interfaces"`
	ARP        bool `json:"arp"`
	Syslog     bool `json:"syslog"`
}

type ExportConfig struct {
	URL             string   `json:"url"`
	BearerTokenFile string   `json:"bearer_token_file"`
	Timeout         Duration `json:"timeout"`
	BatchSize       int      `json:"batch_size"`
	MaxPending      int      `json:"max_pending_events"`
}

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("duration must be a string: %w", err)
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return err
	}
	d.Duration = parsed
	return nil
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.DeviceID == "" {
		return Config{}, fmt.Errorf("device_id is required")
	}
	if cfg.Interval.Duration <= 0 {
		cfg.Interval.Duration = 30 * time.Second
	}
	if cfg.Export.Timeout.Duration <= 0 {
		cfg.Export.Timeout.Duration = 10 * time.Second
	}
	if cfg.Export.BatchSize <= 0 {
		cfg.Export.BatchSize = 100
	}
	if cfg.Export.MaxPending <= 0 {
		cfg.Export.MaxPending = 1000
	}
	if cfg.SyslogPath == "" {
		cfg.SyslogPath = "/var/log/messages"
	}
	return cfg, nil
}
