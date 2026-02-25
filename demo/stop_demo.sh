#!/bin/bash

# mbaigo Demo - Stop Script
# Usage: ./stop_demo.sh

DEMO_DIR="$(cd "$(dirname "$0")" && pwd)"
PIDS_FILE="$DEMO_DIR/demo.pids"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Stopping mbaigo demo...${NC}"

if [ -f "$PIDS_FILE" ]; then
  while read -r pid; do
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null
    fi
  done < "$PIDS_FILE"
  rm -f "$PIDS_FILE"
  echo -e "${GREEN}All processes stopped.${NC}"
else
  echo -e "${YELLOW}No PID file - stopping by name...${NC}"
fi

# Fallback: kill by service name in case anything escaped
pkill -f "start_modbus_simulator" 2>/dev/null || true
pkill -f "mosquitto.*mosquitto.conf" 2>/dev/null || true
for svc in esr orchestrator modboss telegrapher uaclient; do
  pkill -f "demo/bin/$svc-bin" 2>/dev/null || true
done

echo -e "${GREEN}Done.${NC}"
