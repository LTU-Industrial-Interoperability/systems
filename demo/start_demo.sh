#!/bin/bash

# mbaigo Demo - Startup Script
# Usage: cd demo && ./start_demo.sh
# Stop:  ./stop_demo.sh  (or Ctrl+C)

DEMO_DIR="$(cd "$(dirname "$0")" && pwd)"
PIDS_FILE="$DEMO_DIR/demo.pids"
SRC_DIR="$(dirname "$DEMO_DIR")"

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

cleanup() {
  echo ""
  echo -e "${YELLOW}Shutting down demo...${NC}"
  "$DEMO_DIR/stop_demo.sh"
  exit 0
}
trap cleanup SIGINT SIGTERM

save_pid() { echo "$1" >> "$PIDS_FILE"; }

wait_for_http() {
  local name=$1 url=$2 timeout=${3:-15} elapsed=0
  echo -ne "  Waiting for $name..."
  while ! curl -sf "$url" > /dev/null 2>&1; do
    sleep 1; elapsed=$((elapsed+1))
    if [ $elapsed -ge $timeout ]; then
      echo -e " ${RED}TIMEOUT${NC}"
      echo -e "  ${YELLOW}Check: tail -f $DEMO_DIR/logs/${name}.log${NC}"
      return 1
    fi
    echo -n "."
  done
  echo -e " ${GREEN}OK${NC}"; return 0
}

wait_for_port() {
  local name=$1 host=$2 port=$3 timeout=${4:-10} elapsed=0
  echo -ne "  Waiting for $name..."
  while ! (echo > /dev/tcp/$host/$port) 2>/dev/null; do
    sleep 1; elapsed=$((elapsed+1))
    if [ $elapsed -ge $timeout ]; then
      echo -e " ${RED}TIMEOUT${NC}"; return 1
    fi
    echo -n "."
  done
  echo -e " ${GREEN}OK${NC}"; return 0
}

# ── Preflight ─────────────────────────────────────────────────────────────────
echo -e "${BLUE}=== mbaigo Demo ===${NC}\n"
echo -e "${BLUE}Checking dependencies...${NC}"

for cmd in go python3 mosquitto curl; do
  if ! command -v $cmd &>/dev/null; then
    echo -e "${RED}x $cmd not found. Install: brew install $cmd${NC}"; exit 1
  fi
  echo -e "${GREEN}v $cmd${NC}"
done

if ! python3 -c "import pymodbus" 2>/dev/null; then
  echo -e "${RED}x pymodbus not installed. Run: pip3 install pymodbus==3.11.4${NC}"; exit 1
fi
echo -e "${GREEN}v pymodbus${NC}"

# ── Build ─────────────────────────────────────────────────────────────────────
echo ""
echo -e "${BLUE}Building Go services...${NC}"
mkdir -p "$DEMO_DIR/bin"

BUILD_FAILED=0
for svc in esr orchestrator modboss telegrapher uaclient; do
  echo -ne "  Building $svc..."
  if (cd "$SRC_DIR/$svc" && go build -o "$DEMO_DIR/bin/$svc-bin" . 2>"$DEMO_DIR/logs/build_$svc.err"); then
    echo -e " ${GREEN}OK${NC}"
  else
    echo -e " ${RED}FAILED - see logs/build_$svc.err${NC}"
    BUILD_FAILED=$((BUILD_FAILED+1))
  fi
done

if [ $BUILD_FAILED -gt 0 ]; then
  echo -e "${RED}$BUILD_FAILED build(s) failed. Fix errors before starting.${NC}"; exit 1
fi

# ── Prepare ───────────────────────────────────────────────────────────────────
mkdir -p "$DEMO_DIR/logs"
rm -f "$PIDS_FILE"
save_pid $$  # save this script's PID so stop_demo.sh can signal it

echo ""
echo -e "${BLUE}Copying demo configs...${NC}"
for svc in esr orchestrator modboss telegrapher uaclient; do
  mkdir -p "$DEMO_DIR/run/$svc"
  cp "$DEMO_DIR/configs/$svc.json" "$DEMO_DIR/run/$svc/systemconfig.json"
  echo -e "  ${GREEN}v $svc${NC}"
done

# ── Start infrastructure ──────────────────────────────────────────────────────
echo ""
echo -e "${BLUE}Starting infrastructure...${NC}"

echo -e "${BLUE}[1/2] Mosquitto (MQTT broker)${NC}"
mosquitto -c "$DEMO_DIR/mosquitto.conf" >> "$DEMO_DIR/logs/mosquitto.log" 2>&1 &
save_pid $!
wait_for_port "mosquitto" "127.0.0.1" "1883" 10

echo -e "${BLUE}[2/2] Modbus TCP simulator${NC}"
python3 "$DEMO_DIR/start_modbus_simulator.sh" >> "$DEMO_DIR/logs/modbus.log" 2>&1 &
save_pid $!
wait_for_port "modbus" "127.0.0.1" "5020" 10

# ── Start Go services ─────────────────────────────────────────────────────────
# Each service writes to its log file. We start a tail -f for each so output
# appears on screen too. Both the service PID and tail PID are saved so
# stop_demo.sh can cleanly kill everything.
echo ""
echo -e "${BLUE}Starting Go services...${NC}"

start_service() {
  local name=$1
  local log="$DEMO_DIR/logs/$name.log"
  echo -e "  ${BLUE}$name${NC}"
  > "$log"  # truncate log on fresh start
  (cd "$DEMO_DIR/run/$name" && exec "$DEMO_DIR/bin/$name-bin") >> "$log" 2>&1 &
  save_pid $!
  tail -f "$log" &   # track tail PID directly (no pipe) so stop_demo.sh can kill it
  save_pid $!
}

start_service esr
start_service orchestrator
sleep 2
start_service modboss
start_service telegrapher
start_service uaclient

# ── Health checks ─────────────────────────────────────────────────────────────
echo ""
echo -e "${BLUE}Waiting for services to be ready...${NC}"

# Use port checks (not HTTP) - curl -sf fails on any non-2xx, port check just
# confirms the service is listening, which is enough.
FAILED=0
wait_for_port "esr"          "127.0.0.1" "20102" 20 || FAILED=$((FAILED+1))
wait_for_port "orchestrator" "127.0.0.1" "20103" 20 || FAILED=$((FAILED+1))
wait_for_port "modboss"      "127.0.0.1" "20171" 20 || FAILED=$((FAILED+1))
wait_for_port "telegrapher"  "127.0.0.1" "20172" 20 || FAILED=$((FAILED+1))
wait_for_port "uaclient"     "127.0.0.1" "20170" 20 || FAILED=$((FAILED+1))

echo ""
if [ $FAILED -eq 0 ]; then
  echo -e "${GREEN}All services running!${NC}"
else
  echo -e "${YELLOW}$FAILED service(s) did not respond - check logs above${NC}"
fi

echo ""
echo "  ESR:          http://localhost:20102/serviceregistrar/registry"
echo "  Orchestrator: http://localhost:20103/orchestrator/orchestration"
echo "  Modboss:      http://localhost:20171/modboss/access"
echo "  Telegrapher:  http://localhost:20172/telegrapher/access"
echo "  UAClient:     http://localhost:20170/opcuac/access"
echo ""
echo -e "${YELLOW}Stop: ./stop_demo.sh  (or Ctrl+C)${NC}"
echo ""

wait
