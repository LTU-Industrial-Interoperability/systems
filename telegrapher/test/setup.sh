#!/bin/bash
# Generates all certificates and password files needed for manual MQTT security testing.
# Run once from this directory: cd telegrapher/test && bash setup.sh
#
# Produces:
#   ca.key / ca.crt         — self-signed CA (signs the broker cert)
#   broker.key / broker.crt — broker certificate (presented to clients)
#   passwords               — mosquitto password file (user: testuser / pass: testpass)

set -e
cd "$(dirname "$0")"

echo "=== Generating CA ==="
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt \
  -subj "/CN=TestMQTT-CA/O=TestOrg"

echo "=== Generating broker key and certificate ==="
openssl genrsa -out broker.key 2048
# SAN must include the hostname the client will connect to (127.0.0.1 for local tests)
openssl req -new -key broker.key -out broker.csr \
  -subj "/CN=localhost/O=TestOrg"
openssl x509 -req -days 365 -in broker.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out broker.crt \
  -extfile <(printf "subjectAltName=IP:127.0.0.1,DNS:localhost")

echo "=== Creating password file (testuser / testpass) ==="
# -c creates a new file; -b reads password from command line (non-interactive)
mosquitto_passwd -c -b passwords testuser testpass

echo ""
echo "Done. Files created:"
ls -1 ca.crt ca.key broker.crt broker.key passwords broker.csr 2>/dev/null
echo ""
echo "To add more users:  mosquitto_passwd -b passwords anotheruser theirpass"
echo "To delete a user:   mosquitto_passwd -D passwords username"
