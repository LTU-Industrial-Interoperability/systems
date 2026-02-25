#!/bin/bash
set -e

echo "Starting mbaigo services..."

# Give external dependencies more time to be ready
echo "⏳ Waiting for MQTT broker and Modbus simulator to be ready..."
sleep 5

# Start ESR
(cd /app/services/esr && ./esr) &
ESR_PID=$!
echo "✓ ESR started (PID: $ESR_PID)"

sleep 3

# Start Orchestrator
(cd /app/services/orchestrator && ./orchestrator) &
ORCHESTRATOR_PID=$!
echo "✓ Orchestrator started (PID: $ORCHESTRATOR_PID)"

sleep 3

# Start Modboss (connects to modbus-sim)
(cd /app/services/modboss && ./modboss) &
MODBOSS_PID=$!
echo "✓ Modboss started (PID: $MODBOSS_PID)"

sleep 3

# Start Telegrapher (connects to mosquitto)
(cd /app/services/telegrapher && ./telegrapher) &
TELEGRAPHER_PID=$!
echo "✓ Telegrapher started (PID: $TELEGRAPHER_PID)"

sleep 3

# Start OPC UA Client (connects to Prosys on host)
(cd /app/services/uaclient && ./uaclient) &
UACLIENT_PID=$!
echo "✓ OPC UA Client started (PID: $UACLIENT_PID)"

echo ""
echo "🎉 All services running:"
echo "   ESR:         http://localhost:20102"
echo "   Orchestrator: http://localhost:20103"
echo "   Modboss:     http://localhost:20171"
echo "   Telegrapher: http://localhost:20172"
echo "   UAClient:    http://localhost:20170"
echo ""

# Keep container running
wait
