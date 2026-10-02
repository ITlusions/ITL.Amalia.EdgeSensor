package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/collector"
	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/event"
	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/exporter"
)

type Agent struct {
	Collectors []collector.Collector
	Exporter   exporter.Exporter
	Interval   time.Duration
	BatchSize  int
	Logger     *log.Logger
	MaxPending int
	pending    []event.Event
}

func (a *Agent) Run(ctx context.Context, once bool) error {
	if err := a.collectAndExport(ctx); err != nil {
		if once {
			return err
		}
		a.Logger.Printf("collection cycle failed: %v", err)
	}
	if once {
		return nil
	}

	ticker := time.NewTicker(a.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := a.collectAndExport(ctx); err != nil {
				a.Logger.Printf("collection cycle failed: %v", err)
			}
		}
	}
}

func (a *Agent) collectAndExport(ctx context.Context) error {
	maxPending := a.MaxPending
	if maxPending <= 0 {
		maxPending = 1000
	}
	for _, source := range a.Collectors {
		collected, err := source.Collect(ctx)
		if err != nil {
			a.Logger.Printf("collector %s unavailable: %v", source.Name(), err)
			continue
		}
		available := maxPending - len(a.pending)
		if len(collected) > available {
			a.Logger.Printf("pending queue full: dropping %d new events from %s", len(collected)-available, source.Name())
			collected = collected[:available]
		}
		a.pending = append(a.pending, collected...)
	}

	batchSize := a.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	for len(a.pending) > 0 {
		end := batchSize
		if end > len(a.pending) {
			end = len(a.pending)
		}
		if err := a.Exporter.Export(ctx, a.pending[:end]); err != nil {
			return fmt.Errorf("export events: %w", err)
		}
		clear(a.pending[:end])
		a.pending = a.pending[end:]
	}
	a.pending = nil
	return nil
}
