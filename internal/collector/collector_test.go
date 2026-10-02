package collector

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseConntrackLine(t *testing.T) {
	line := "ipv4 2 tcp 6 431999 ESTABLISHED src=192.0.2.10 dst=198.51.100.20 sport=54321 dport=443 packets=4 bytes=320 src=198.51.100.20 dst=192.0.2.10 sport=443 dport=54321 packets=3 bytes=510 [ASSURED] mark=0 use=1"
	got, ok := parseConntrackLine(line)
	if !ok {
		t.Fatal("parseConntrackLine() rejected valid input")
	}
	if got.Protocol != "tcp" || got.State != "ESTABLISHED" || got.SourcePort != 54321 || got.DestinationPort != 443 || got.Bytes != 320 {
		t.Fatalf("unexpected flow: %+v", got)
	}
}

func TestParseInterfaceLine(t *testing.T) {
	line := " eth0: 1000 10 1 0 0 0 0 0 2000 20 2 0 0 0 0 0"
	got, ok := parseInterfaceLine(line)
	if !ok {
		t.Fatal("parseInterfaceLine() rejected valid input")
	}
	if got.Name != "eth0" || got.RXBytes != 1000 || got.RXErrors != 1 || got.TXBytes != 2000 || got.TXErrors != 2 {
		t.Fatalf("unexpected counters: %+v", got)
	}
}

func TestParseARPLine(t *testing.T) {
	line := "192.0.2.50 0x1 0x2 00:11:22:33:44:55 * switch0"
	got, ok := parseARPLine(line)
	if !ok {
		t.Fatal("parseARPLine() rejected valid input")
	}
	if got.IPAddress != "192.0.2.50" || got.MACAddress != "00:11:22:33:44:55" || got.Device != "switch0" {
		t.Fatalf("unexpected ARP entry: %+v", got)
	}
}

func TestSyslogCollectorOnlyReturnsNewLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "messages")
	if err := os.WriteFile(path, []byte("first message\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	source := &SyslogCollector{DeviceID: "er-x-01", Path: path}
	first, err := source.Collect(context.Background())
	if err != nil || len(first) != 1 {
		t.Fatalf("first Collect() events=%d err=%v", len(first), err)
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("second message\n"); err != nil {
		t.Fatal(err)
	}
	file.Close()
	second, err := source.Collect(context.Background())
	if err != nil || len(second) != 1 {
		t.Fatalf("second Collect() events=%d err=%v", len(second), err)
	}
}
