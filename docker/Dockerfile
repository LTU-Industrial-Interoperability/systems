FROM golang:1.25.6 AS builder

WORKDIR /app

# Copy all source code
COPY . .

# Build all services
RUN cd esr && go build -o ../esr-bin . && cd .. && \
    cd orchestrator && go build -o ../orchestrator-bin . && cd .. && \
    cd modboss && go build -o ../modboss-bin . && cd .. && \
    cd telegrapher && go build -o ../telegrapher-bin . && cd .. && \
    cd uaclient && go build -o ../uaclient-bin .

# Runtime image
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Create service directories
RUN mkdir -p /app/services/esr \
             /app/services/orchestrator \
             /app/services/modboss \
             /app/services/telegrapher \
             /app/services/uaclient

# Copy built binaries to service directories
COPY --from=builder /app/esr-bin /app/services/esr/esr
COPY --from=builder /app/orchestrator-bin /app/services/orchestrator/orchestrator
COPY --from=builder /app/modboss-bin /app/services/modboss/modboss
COPY --from=builder /app/telegrapher-bin /app/services/telegrapher/telegrapher
COPY --from=builder /app/uaclient-bin /app/services/uaclient/uaclient

# Copy config files (renamed to systemconfig.json in their respective directories)
COPY docker/configs/esr.json /app/services/esr/systemconfig.json
COPY docker/configs/orchestrator.json /app/services/orchestrator/systemconfig.json
COPY docker/configs/modboss.json /app/services/modboss/systemconfig.json
COPY docker/configs/telegrapher.json /app/services/telegrapher/systemconfig.json
COPY docker/configs/uaclient.json /app/services/uaclient/systemconfig.json

# Copy startup script
COPY docker/start-all-services.sh ./
RUN chmod +x ./start-all-services.sh

CMD ["./start-all-services.sh"]
