#!/bin/bash
set -e

# preserve PATH for sudo
if [[ -n "$SUDO_COMMAND" ]]; then
    export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"
fi

# deploy params
BASIC_DEPLOY=true
DEPLOY_ROOT="$HOME/w2g"

# app params
APP_NAME="w2g"

PORT=8080
STORAGE_DIR=""
BIN_DIR=""
CONF_DIR=""
LOG_FILE=""

LOG_LEVEL=info

# logging
mkdir -p "$DEPLOY_ROOT/logs"
DEPLOY_LOG="$DEPLOY_ROOT/logs/deploy.log"

log() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] $1"
    echo "$msg" | tee -a "$DEPLOY_LOG"
}

while [[ $# -gt 0 ]]; do
    case $1 in
        --deploy-root) DEPLOY_ROOT="$2"; shift 2 ;;
        --default) BASIC_DEPLOY="$2"; shift 2 ;;
        --port) PORT="$2"; shift 2 ;;
        --storage) STORAGE="$2"; shift 2 ;;
        --log-level) LOG_LEVEL="$2"; shift 2 ;;
        --log-file) LOG_FILE="$2"; shift 2 ;;
        *) echo "Unknown: $1" >&2; exit 1 ;;
    esac
done

if [[ "$BASIC_DEPLOY" == "true" ]]; then
    PORT=8080
    LOG_LEVEL="info"

    STORAGE_DIR="/var/lib/$APP_NAME/storage"
    LOG_FILE="/var/log/$APP_NAME/w2g.log"
    CONF_DIR="/etc/$APP_NAME"
    BIN_DIR="/usr/bin"
fi

STORAGE_DIR="${STORAGE_DIR:-$DEPLOY_ROOT/storage}"
LOG_FILE="${LOG_FILE:-$DEPLOY_ROOT/logs/w2g.log}"
BIN_DIR="${BIN_DIR:-$DEPLOY_ROOT/bin}"
CONF_DIR="${CONF_DIR:-$DEPLOY_ROOT/config}"

# validation: check required paths and create if needed
for dir in "$(dirname "$LOG_FILE")" "$STORAGE_DIR" "$CONF_DIR" "$BIN_DIR"; do
    if [[ ! -d "$dir" ]]; then
        if mkdir -p "$dir" 2>/dev/null; then
            log "Created directory: $dir"
        else
            log "ERROR: cannot create directory: $dir (need sudo?)"
            exit 1
        fi
    fi
done

# validation: port not in use (basic check)
if command -v ss &> /dev/null; then
    if ss -tuln 2>/dev/null | grep -q ":$PORT "; then
        log "WARNING: port $PORT may be in use"
    fi
fi

# validation: check go version >= 1.21
GO_MIN_VERSION="1.21"
GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
if [[ $(echo "$GO_VERSION < $GO_MIN_VERSION" | bc -l) -eq 1 ]]; then
    log "ERROR: go version $GO_VERSION is too old, need >= $GO_MIN_VERSION"
    exit 1
fi
log "Go version $GO_VERSION OK"

if ! command -v go &> /dev/null; then
    log "ERROR: go not installed"
    exit 1
fi
go version | tee -a "$DEPLOY_LOG"

# check source exists
if [[ ! -d "web" ]]; then
    log "ERROR: web/ directory not found"
    exit 1
fi
if [[ ! -f "cmd/w2g/main.go" ]]; then
    log "ERROR: cmd/w2g/main.go not found"
    exit 1
fi

# build bin
mkdir -p "$DEPLOY_ROOT/bin"
log "Building w2g..."
go build -o "$DEPLOY_ROOT/bin/$APP_NAME" ./cmd/w2g

# make config
mkdir -p "$DEPLOY_ROOT/config"
cat > "$DEPLOY_ROOT/config/.env" <<EOF
PORT=$PORT
STORAGE_DIR=$STORAGE_DIR/
LOG_LEVEL=$LOG_LEVEL
LOG_FILE=$LOG_FILE
SESSIONS_CLEANUP_INTERVAL=300
MAX_ROOMS_PER_USER=10
EOF
log "Config written to $DEPLOY_ROOT/config/.env"

# systemd (system-wide)
RUN_USER="${RUN_USER:-$(whoami)}"
cat > "$DEPLOY_ROOT/config/$APP_NAME.service" <<EOF
[Unit]
Description=$APP_NAME watch together
After=network.target

[Service]
Type=simple
User=$RUN_USER
WorkingDirectory=$BIN_DIR
EnvironmentFile=$CONF_DIR/.env
ExecStart=$BIN_DIR/$APP_NAME
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
log "Systemd unit created"

# finalize
log "Copying artifacts to system directories..."

if systemctl is-active --quiet $APP_NAME.service 2>/dev/null; then
    log "Stopping service before update..."
    systemctl stop $APP_NAME.service
    sleep 1
fi

cp -f "$DEPLOY_ROOT/bin/$APP_NAME" "$BIN_DIR"
cp -f "$DEPLOY_ROOT/config/$APP_NAME.service" "/etc/systemd/system/$APP_NAME.service"
cp -f "$DEPLOY_ROOT/config/.env" "$CONF_DIR"
mkdir -p "$STORAGE_DIR"

log "Reloading systemd and restarting service..."
systemctl daemon-reload
systemctl restart $APP_NAME.service

# healthz
sleep 2
log "Checking health..."
curl -sf "http://localhost:$PORT/healthz" || { log "ERROR: health check failed"; exit 1; }
log "Health check OK"

# logs
log "=== Last 10 lines of app log ==="
tail -n 10 "$LOG_FILE" 2>/dev/null || log "Log file not found"

log "=== Deploy complete ==="
