package exporter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/event"
)

func TestHTTPExporterSendsAuthenticatedBatch(t *testing.T) {
	tokenFile := t.TempDir() + "/token"
	if err := os.WriteFile(tokenFile, []byte("secret-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var got Batch
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer secret-token" {
			t.Errorf("unexpected request: method=%s auth=%q", r.Method, r.Header.Get("Authorization"))
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	e, err := event.New("edge.arp", "er-x-01", "info", time.Now(), map[string]string{"ip": "192.0.2.1"})
	if err != nil {
		t.Fatal(err)
	}
	exporter := NewHTTP(server.URL, tokenFile, time.Second)
	if err := exporter.Export(context.Background(), []event.Event{e}); err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if len(got.Events) != 1 || got.Events[0].DeviceID != "er-x-01" {
		t.Fatalf("unexpected batch: %+v", got)
	}
}
