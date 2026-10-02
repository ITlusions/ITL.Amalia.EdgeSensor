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

type ConntrackCollector struct {
	DeviceID string
	Path     string
}

type ConntrackFlow struct {
	Protocol        string `json:"protocol"`
	State           string `json:"state,omitempty"`
	SourceAddress   string `json:"source_address"`
	SourcePort      int    `json:"source_port,omitempty"`
	Destination     string `json:"destination_address"`
	DestinationPort int    `json:"destination_port,omitempty"`
	Packets         uint64 `json:"packets,omitempty"`
	Bytes           uint64 `json:"bytes,omitempty"`
}

func (c ConntrackCollector) Name() string { return "conntrack" }

func (c ConntrackCollector) Collect(_ context.Context) ([]event.Event, error) {
	path := c.Path
	if path == "" {
		path = "/proc/net/nf_conntrack"
	}

	file, err := os.Open(path)
	if err != nil && c.Path == "" {
		file, err = os.Open("/proc/net/ip_conntrack")
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []event.Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		flow, ok := parseConntrackLine(scanner.Text())
		if !ok {
			continue
		}
		e, err := event.New("edge.conntrack", c.DeviceID, "info", time.Now(), flow)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, scanner.Err()
}

func parseConntrackLine(line string) (ConntrackFlow, bool) {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return ConntrackFlow{}, false
	}

	flow := ConntrackFlow{Protocol: fields[2]}
	if fields[2] == "tcp" && len(fields) > 5 && !strings.Contains(fields[5], "=") {
		flow.State = fields[5]
	}

	values := make(map[string][]string)
	for _, field := range fields {
		key, value, found := strings.Cut(field, "=")
		if found {
			values[key] = append(values[key], value)
		}
	}
	if len(values["src"]) == 0 || len(values["dst"]) == 0 {
		return ConntrackFlow{}, false
	}

	flow.SourceAddress = values["src"][0]
	flow.Destination = values["dst"][0]
	flow.SourcePort = firstInt(values["sport"])
	flow.DestinationPort = firstInt(values["dport"])
	flow.Packets = firstUint(values["packets"])
	flow.Bytes = firstUint(values["bytes"])
	return flow, true
}

func firstInt(values []string) int {
	if len(values) == 0 {
		return 0
	}
	value, _ := strconv.Atoi(values[0])
	return value
}

func firstUint(values []string) uint64 {
	if len(values) == 0 {
		return 0
	}
	value, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return 0
	}
	return value
}
