package agent

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/collector"
	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/event"
)

type testCollector struct {
	events []event.Event
}

func (source *testCollector) Name() string { return "test" }

func (source *testCollector) Collect(context.Context) ([]event.Event, error) {
	events := source.events
	source.events = nil
	return events, nil
}

type testExporter struct {
	failOnCall int
	calls      int
	batches    [][]event.Event
	cancel     context.CancelFunc
}

func (sink *testExporter) Export(_ context.Context, events []event.Event) error {
	sink.calls++
	if sink.calls == sink.failOnCall {
		return errors.New("offline")
	}
	sink.batches = append(sink.batches, append([]event.Event(nil), events...))
	if sink.cancel != nil {
		sink.cancel()
	}
	return nil
}

func TestRetriesFailedBatchWithoutResendingSuccessfulBatch(t *testing.T) {
	source := &testCollector{events: []event.Event{{DeviceID: "first"}, {DeviceID: "second"}, {DeviceID: "third"}}}
	sink := &testExporter{failOnCall: 2}
	runner := Agent{
		Collectors: []collector.Collector{source}, Exporter: sink,
		BatchSize: 1, Logger: log.New(io.Discard, "", 0),
	}
	if err := runner.collectAndExport(context.Background()); err == nil {
		t.Fatal("expected export failure")
	}
	if len(runner.pending) != 2 || runner.pending[0].DeviceID != "second" {
		t.Fatalf("unexpected pending events: %+v", runner.pending)
	}
	source.events = []event.Event{{DeviceID: "fourth"}}
	if err := runner.collectAndExport(context.Background()); err != nil {
		t.Fatal(err)
	}
	for index, want := range []string{"first", "second", "third", "fourth"} {
		if len(sink.batches) != 4 || sink.batches[index][0].DeviceID != want {
			t.Fatalf("unexpected delivered batches: %+v", sink.batches)
		}
	}
	if len(runner.pending) != 0 {
		t.Fatal("queue not drained after recovery")
	}
}

func TestQueueOverflowPreservesOlderEvents(t *testing.T) {
	source := &testCollector{events: []event.Event{{DeviceID: "first"}, {DeviceID: "second"}}}
	sink := &testExporter{failOnCall: 1}
	runner := Agent{
		Collectors: []collector.Collector{source}, Exporter: sink,
		MaxPending: 2, Logger: log.New(io.Discard, "", 0),
	}
	if err := runner.collectAndExport(context.Background()); err == nil {
		t.Fatal("expected export failure")
	}
	source.events = []event.Event{{DeviceID: "dropped"}}
	if err := runner.collectAndExport(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(sink.batches) != 1 || len(sink.batches[0]) != 2 || sink.batches[0][1].DeviceID != "second" {
		t.Fatalf("unexpected batches: %+v", sink.batches)
	}
}

func TestDaemonSurvivesInitialExportFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	source := &testCollector{events: []event.Event{{DeviceID: "first"}}}
	sink := &testExporter{failOnCall: 1, cancel: cancel}
	runner := Agent{
		Collectors: []collector.Collector{source}, Exporter: sink,
		Interval: time.Millisecond, Logger: log.New(io.Discard, "", 0),
	}
	if err := runner.Run(ctx, false); err != nil {
		t.Fatal(err)
	}
	if sink.calls != 2 || len(sink.batches) != 1 {
		t.Fatalf("daemon failed to recover: calls=%d batches=%d", sink.calls, len(sink.batches))
	}
}

func TestOnceReturnsExportFailure(t *testing.T) {
	source := &testCollector{events: []event.Event{{DeviceID: "first"}}}
	runner := Agent{
		Collectors: []collector.Collector{source}, Exporter: &testExporter{failOnCall: 1},
		Logger: log.New(io.Discard, "", 0),
	}
	if err := runner.Run(context.Background(), true); err == nil {
		t.Fatal("expected one-shot failure")
	}
}