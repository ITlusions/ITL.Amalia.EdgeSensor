package event

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewCreatesStableEnvelope(t *testing.T) {
	observedAt := time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)
	got, err := New("edge.conntrack", "er-x-01", "info", observedAt, map[string]string{"protocol": "tcp"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got.SchemaVersion != "1.0" || got.SourceType != "edge.conntrack" || got.DeviceID != "er-x-01" {
		t.Fatalf("unexpected envelope: %+v", got)
	}

	var data map[string]string
	if err := json.Unmarshal(got.Data, &data); err != nil {
		t.Fatalf("data is not valid JSON: %v", err)
	}
	if data["protocol"] != "tcp" {
		t.Fatalf("protocol = %q, want tcp", data["protocol"])
	}
}
