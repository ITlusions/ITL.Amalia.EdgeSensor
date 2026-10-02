package collector

import (
	"bufio"
	"context"
	"os"
	"strings"
	"time"

	"github.com/ITlusions/ITL.Amalia.EdgeSensor/internal/event"
)

type ARPCollector struct {
	DeviceID string
	Path     string
}

type ARPEntry struct {
	IPAddress  string `json:"ip_address"`
	HWType     string `json:"hardware_type"`
	Flags      string `json:"flags"`
	MACAddress string `json:"mac_address"`
	Device     string `json:"device"`
}

func (c ARPCollector) Name() string { return "arp" }

func (c ARPCollector) Collect(_ context.Context) ([]event.Event, error) {
	path := c.Path
	if path == "" {
		path = "/proc/net/arp"
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []event.Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		entry, ok := parseARPLine(scanner.Text())
		if !ok {
			continue
		}
		e, err := event.New("edge.arp", c.DeviceID, "info", time.Now(), entry)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, scanner.Err()
}

func parseARPLine(line string) (ARPEntry, bool) {
	fields := strings.Fields(line)
	if len(fields) != 6 || fields[0] == "IP" {
		return ARPEntry{}, false
	}
	return ARPEntry{
		IPAddress:  fields[0],
		HWType:     fields[1],
		Flags:      fields[2],
		MACAddress: fields[3],
		Device:     fields[5],
	}, true
}
