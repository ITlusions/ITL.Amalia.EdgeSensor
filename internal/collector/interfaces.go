package collector

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/event"
)

type InterfaceCollector struct {
	DeviceID string
	Path     string
}

type InterfaceCounters struct {
	Name      string `json:"name"`
	RXBytes   uint64 `json:"rx_bytes"`
	RXPackets uint64 `json:"rx_packets"`
	RXErrors  uint64 `json:"rx_errors"`
	TXBytes   uint64 `json:"tx_bytes"`
	TXPackets uint64 `json:"tx_packets"`
	TXErrors  uint64 `json:"tx_errors"`
}

func (c InterfaceCollector) Name() string { return "interfaces" }

func (c InterfaceCollector) Collect(_ context.Context) ([]event.Event, error) {
	path := c.Path
	if path == "" {
		path = "/proc/net/dev"
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []event.Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		counters, ok := parseInterfaceLine(scanner.Text())
		if !ok {
			continue
		}
		e, err := event.New("edge.interface", c.DeviceID, "info", time.Now(), counters)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, scanner.Err()
}

func parseInterfaceLine(line string) (InterfaceCounters, bool) {
	name, raw, found := strings.Cut(line, ":")
	if !found {
		return InterfaceCounters{}, false
	}
	fields := strings.Fields(raw)
	if len(fields) < 16 {
		return InterfaceCounters{}, false
	}
	return InterfaceCounters{
		Name:      strings.TrimSpace(name),
		RXBytes:   parseUint(fields[0]),
		RXPackets: parseUint(fields[1]),
		RXErrors:  parseUint(fields[2]),
		TXBytes:   parseUint(fields[8]),
		TXPackets: parseUint(fields[9]),
		TXErrors:  parseUint(fields[10]),
	}, true
}

func parseUint(value string) uint64 {
	parsed, _ := strconv.ParseUint(value, 10, 64)
	return parsed
}