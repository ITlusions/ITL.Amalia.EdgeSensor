#!/bin/sh
set -eu

PID_FILE=/var/run/edge-sensor.pid
if [ -f "$PID_FILE" ]; then
  kill "$(cat "$PID_FILE")" 2>/dev/null || true
  rm -f "$PID_FILE"
fi
rm -f /config/scripts/post-config.d/50-edge-sensor
rm -rf /config/edge-sensor
rm -f /var/log/edge-sensor.log
echo 'EdgeSensor removed.'