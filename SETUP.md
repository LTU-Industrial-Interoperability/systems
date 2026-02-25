# mbaigo Systems — Setup

Two ways to run the demo. Both require **Prosys OPC UA** running separately on the host at `localhost:53530` — there is no OPC UA simulator included.

---

## Option 1: Docker

```bash
docker-compose build
docker-compose up        # add -d to run in background
docker-compose down      # to stop
```

Services started: ESR, Orchestrator, Modboss, Telegrapher, UAClient, Mosquitto, Modbus TCP simulator.

> OPC UA (UAClient) will connect to Prosys on your host machine. Start Prosys before running.

---

## Option 2: Local script

### Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.25+ | `brew install go` |
| Python | 3.11+ | `brew install python@3.11` |
| pymodbus | **3.11.4** | `pip3 install pymodbus==3.11.4` |
| Mosquitto | 2.0+ | `brew install mosquitto` |

> pymodbus version matters — the Modbus simulator uses the 3.x API (`ModbusDeviceContext`). Earlier versions will fail silently.

### Run

```bash
cd demo
./start_demo.sh    # builds Go binaries, starts all services, shows live logs
./stop_demo.sh     # stop everything cleanly
```

Configs are in `demo/configs/`. Logs go to `demo/logs/`.

> UAClient connects to Prosys at `localhost:53530`. Start Prosys before running.

---

## Service ports

| Service | Port |
|---------|------|
| ESR (Service Registrar) | 20102 |
| Orchestrator | 20103 |
| UAClient (OPC UA) | 20170 |
| Modboss (Modbus TCP) | 20171 |
| Telegrapher (MQTT) | 20172 |
| Mosquitto (MQTT broker) | 1883 |
| Modbus TCP simulator | 5020 |
