# ITL.Amalia.EdgeSensor

Lightweight defensive telemetry sensor for Ubiquiti EdgeRouter X devices. It runs entirely in userspace and does not require eBPF, kernel modules, or additional runtime dependencies.

## Capabilities

- Conntrack flow snapshots from `/proc/net/nf_conntrack` or `/proc/net/ip_conntrack`
- Interface byte, packet, and error counters from `/proc/net/dev`
- Neighbor observations from `/proc/net/arp`
- Incremental EdgeOS messages from `/var/log/messages`
- Bounded JSON batches exported over HTTPS with optional bearer authentication
- Static `linux/mipsle` build suitable for ER-X (`GOMIPS=softfloat`)

The sensor is transparent defensive monitoring software. It does not hide processes, suppress audit data, capture packet payloads, or provide remote command execution.

## Build

Requirements: Go 1.23 or newer on the build machine.

```powershell
./scripts/build-erx.ps1
```

Or on Linux:

```bash
./scripts/build-erx.sh
```

The binary is written to `dist/edge-sensor-linux-mipsle`.

## Test locally

```bash
go test ./...
go run ./cmd/edge-sensor -config config.example.json -once -dry-run
```

Collectors unavailable on the development host are logged and skipped.

## Install on EdgeRouter X

Before installation, copy and edit the example configuration. The export endpoint should terminate TLS and accept the batch contract described below.

```bash
scp dist/edge-sensor-linux-mipsle config.json \
  packaging/edgeos/install.sh ubnt@router:/tmp/
ssh ubnt@router
sudo -i
cd /tmp
echo 'YOUR_BEARER_TOKEN' > token
./install.sh ./edge-sensor-linux-mipsle ./config.json
mv token /config/edge-sensor/token
chmod 600 /config/edge-sensor/token
```

Restart once after installation to verify the `/config/scripts/post-config.d` persistence hook.

## Verify

```bash
cat /var/run/edge-sensor.pid
ps | grep edge-sensor
tail -f /var/log/edge-sensor.log
```

Run a one-time local diagnostic without exporting:

```bash
/config/edge-sensor/edge-sensor \
  -config /config/edge-sensor/config.json -once -dry-run
```

## Event contract

```json
{
  "schema_version": "1.0",
  "sent_at": "2026-10-02T12:00:00Z",
  "events": [
    {
      "schema_version": "1.0",
      "source_type": "edge.conntrack",
      "device_id": "er-x-01",
      "observed_at": "2026-10-02T12:00:00Z",
      "received_at": "2026-10-02T12:00:00Z",
      "severity": "info",
      "data": {}
    }
  ]
}
```

Source types are `edge.conntrack`, `edge.interface`, `edge.arp`, and `edge.syslog`.

## Resource profile

The agent uses the Go standard library only. Start with a 30-second interval and a batch size of 100. Monitor CPU and memory on the ER-X before reducing the interval.

## Delivery and outages

The daemon keeps failed batches in memory and retries them on the next collection cycle, including when the first export fails. Successfully acknowledged batches are removed from the queue.

`export.max_pending_events` limits retained events (default: 1000). When full, the queue preserves older events and drops new arrivals with a warning. This limit counts events, not bytes, and does not bound the collectors' temporary snapshot allocations.

Pending events do not survive process restarts or router reboots. A server that accepts a batch but loses its response can receive the same events again on retry. Durable spooling and receiver-side deduplication are not implemented. One-shot mode reports export failures instead of retrying.

## Remove

```bash
sudo packaging/edgeos/uninstall.sh
```

## Security

- Use an HTTPS endpoint with a valid certificate.
- Store the bearer token only in `/config/edge-sensor/token` with mode `0600`.
- Restrict outbound firewall access to the collector address and port.
- Treat router logs and flow metadata as sensitive operational data.