package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/agent"
	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/collector"
	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/config"
	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/exporter"
)

func main() {
	configPath := flag.String("config", "/config/edge-sensor/config.json", "path to configuration")
	once := flag.Bool("once", false, "collect and export one snapshot")
	dryRun := flag.Bool("dry-run", false, "write JSON batches to stdout")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	collectors := make([]collector.Collector, 0, 4)
	if cfg.Collectors.Conntrack {
		collectors = append(collectors, collector.ConntrackCollector{DeviceID: cfg.DeviceID})
	}
	if cfg.Collectors.Interfaces {
		collectors = append(collectors, collector.InterfaceCollector{DeviceID: cfg.DeviceID})
	}
	if cfg.Collectors.ARP {
		collectors = append(collectors, collector.ARPCollector{DeviceID: cfg.DeviceID})
	}
	if cfg.Collectors.Syslog {
		collectors = append(collectors, &collector.SyslogCollector{DeviceID: cfg.DeviceID, Path: cfg.SyslogPath})
	}

	var sink exporter.Exporter
	if *dryRun {
		sink = exporter.JSONExporter{Writer: os.Stdout}
	} else {
		if cfg.Export.URL == "" {
			log.Fatal("export.url is required unless -dry-run is used")
		}
		sink = exporter.NewHTTP(cfg.Export.URL, cfg.Export.BearerTokenFile, cfg.Export.Timeout.Duration)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	runner := agent.Agent{
		Collectors: collectors,
		Exporter:   sink,
		Interval:   cfg.Interval.Duration,
		BatchSize:  cfg.Export.BatchSize,
		MaxPending: cfg.Export.MaxPending,
		Logger:     log.New(os.Stderr, "edge-sensor: ", log.LstdFlags),
	}
	if err := runner.Run(ctx, *once); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
