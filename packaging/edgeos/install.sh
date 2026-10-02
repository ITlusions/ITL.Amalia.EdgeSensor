#!/bin/sh
set -eu

BINARY=${1:-./edge-sensor-linux-mipsle}
CONFIG=${2:-./config.json}
INSTALL_DIR=/config/edge-sensor
HOOK=/config/scripts/post-config.d/50-edge-sensor

if [ "$(id -u)" -ne 0 ]; then
  echo 'Run this installer as root.' >&2
  exit 1
fi
if [ ! -f "$BINARY" ] || [ ! -f "$CONFIG" ]; then
  echo 'Usage: install.sh [edge-sensor-binary] [config.json]' >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR" /config/scripts/post-config.d
cp "$BINARY" "$INSTALL_DIR/edge-sensor"
cp "$CONFIG" "$INSTALL_DIR/config.json"
chmod 0755 "$INSTALL_DIR/edge-sensor"
chmod 0600 "$INSTALL_DIR/config.json"

cat > "$HOOK" <<'EOF'
#!/bin/sh
INSTALL_DIR=/config/edge-sensor
PID_FILE=/var/run/edge-sensor.pid

if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
  exit 0
fi

nohup "$INSTALL_DIR/edge-sensor" -config "$INSTALL_DIR/config.json" \
  >>/var/log/edge-sensor.log 2>&1 &
echo $! > "$PID_FILE"
EOF
chmod 0755 "$HOOK"

"$HOOK"
echo "Installed EdgeSensor in $INSTALL_DIR"