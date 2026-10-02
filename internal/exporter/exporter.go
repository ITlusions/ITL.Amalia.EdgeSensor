package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/event"
)

type Exporter interface {
	Export(context.Context, []event.Event) error
}

type Batch struct {
	SchemaVersion string        `json:"schema_version"`
	SentAt        time.Time     `json:"sent_at"`
	Events        []event.Event `json:"events"`
}

type HTTPExporter struct {
	URL       string
	TokenFile string
	Client    *http.Client
}

func NewHTTP(url, tokenFile string, timeout time.Duration) *HTTPExporter {
	return &HTTPExporter{URL: url, TokenFile: tokenFile, Client: &http.Client{Timeout: timeout}}
}

func (e *HTTPExporter) Export(ctx context.Context, events []event.Event) error {
	if len(events) == 0 {
		return nil
	}
	payload, err := json.Marshal(Batch{SchemaVersion: "1.0", SentAt: time.Now().UTC(), Events: events})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.URL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if e.TokenFile != "" {
		token, err := os.ReadFile(e.TokenFile)
		if err != nil {
			return fmt.Errorf("read bearer token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
	}
	response, err := e.Client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return fmt.Errorf("export failed: %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

type JSONExporter struct{ Writer io.Writer }

func (e JSONExporter) Export(_ context.Context, events []event.Event) error {
	return json.NewEncoder(e.Writer).Encode(Batch{SchemaVersion: "1.0", SentAt: time.Now().UTC(), Events: events})
}
