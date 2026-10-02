package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPendingLimit(t *testing.T) {
	for _, testCase := range []struct {
		name string
		json string
		want int
	}{
		{name: "default", json: `{"device_id":"er-x"}`, want: 1000},
		{name: "custom", json: `{"device_id":"er-x","export":{"max_pending_events":25}}`, want: 25},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(testCase.json), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Export.MaxPending != testCase.want {
				t.Fatalf("MaxPending=%d, want %d", cfg.Export.MaxPending, testCase.want)
			}
		})
	}
}